package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "net/http"
  "strings"
  "time"

  "github.com/lib/pq"

  "chat/common"
)

type groupEditInput struct {
  GroupName  string   `json:"groupName"`
  GroupImg   string   `json:"groupImg"`
  Bio        string   `json:"bio"`
  GroupBio   string   `json:"groupBio"`
  AliasNames []string `json:"aliasNames"`
}

func uniqueGroupAliasNames(names []string) []string {
  seen := make(map[string]struct{}, len(names))
  result := make([]string, 0, len(names))

  for _, name := range names {
    if _, found := seen[name]; found {
      continue
    }
    seen[name] = struct{}{}
    result = append(result, name)
  }

  return result
}

func GroupEdit(w http.ResponseWriter, r *http.Request) {
  ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
  defer cancel()

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    common.WriteResponseWithoutSession(
      w, r.FormValue("csrf"), "Session Error", http.StatusOK,
    )
    return
  }

  channelID, err := parseChannelID(r)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
  }

  updatedBy := r.FormValue("updatedBy")

  if common.DB == nil ||
    common.DB.GroupDB == nil ||
    common.DB.AliasDB == nil ||
    common.DB.SessionDB == nil {
    common.WriteResponseWithSession(
      w, session, "database is not initialized", http.StatusOK,
    )
    return
  }

  // Normal/admin aliases can edit groups; guests cannot.
  accessRight, err := channelActorAccess(
    ctx, channelID, session.UserID, updatedBy,
  )
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
  }
  if accessRight == 2 {
    common.WriteResponseWithSession(
      w, session, "Guest access cannot edit groups", http.StatusOK,
    )
    return
  }

  groupsJSON := r.FormValue("groups")
  if groupsJSON == "" {
    common.WriteResponseWithSession(
      w, session, "groups is required", http.StatusOK,
    )
    return
  }

  var groups []groupEditInput
  if err := json.Unmarshal([]byte(groupsJSON), &groups); err != nil {
    common.WriteResponseWithSession(
      w, session, "groups JSON unmarshal error", http.StatusOK,
    )
    return
  }

  pushNames, err := parsePushNames(r)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
  }

  sessions, err := loadSessionsForAliasNames(ctx, channelID, pushNames)
  if err != nil {
    common.WriteResponseWithSession(
      w, session, err.Error()+"; session lookup error", http.StatusOK,
    )
    return
  }

  // Process each group independently.
  for i := range groups {
    g := &groups[i]

    if strings.TrimSpace(g.GroupName) == "" {
      common.WriteResponseWithSession(
        w, session, "groupName is required", http.StatusOK,
      )
      return
    }

    // Accept either JSON name while the frontend is being migrated.
    if g.Bio == "" && g.GroupBio != "" {
      g.Bio = g.GroupBio
    }

    // Empty membership means delete the group.
    if len(g.AliasNames) == 0 {
      _, err := common.DB.GroupDB.ExecContext(
        ctx,
        `
        DELETE FROM "group"
        WHERE channel_id = $1
          AND group_name = $2
        `,
        channelID,
        g.GroupName,
      )
      if err != nil {
        common.WriteResponseWithSession(
          w, session, err.Error()+"; group delete error", http.StatusOK,
        )
        return
      }

      g.AliasNames = []string{}
      g.GroupImg = ""
      continue
    }

    // Deduplicate alias names and ensure every name belongs to this channel.
    g.AliasNames = uniqueGroupAliasNames(g.AliasNames)

    var matchedCount int
    err := common.DB.AliasDB.QueryRowContext(
      ctx,
      `
      SELECT COUNT(DISTINCT alias_name)
      FROM alias
      WHERE channel_id = $1
        AND alias_name::text = ANY($2::text[])
      `,
      channelID,
      pq.Array(g.AliasNames),
    ).Scan(&matchedCount)
    if err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; group alias validation error",
        http.StatusOK,
      )
      return
    }

    if matchedCount != len(g.AliasNames) {
      common.WriteResponseWithSession(
        w,
        session,
        fmt.Sprintf("group %q contains an alias not in this channel", g.GroupName),
        http.StatusOK,
      )
      return
    }

    // Save a new image if one was supplied. An existing stored path is
    // passed through by ImgSave; an empty image preserves the DB image.
    if g.GroupImg != "" {
      g.GroupImg, err = common.ImgSave(
        g.GroupImg,
        session.UserID,
        g.GroupName,
        channelID,
        4,
        2,
      )
      if err != nil {
        common.WriteResponseWithSession(
          w, session, err.Error()+"; group image save error",
          http.StatusOK,
        )
        return
      }
    }

    // Requires a UNIQUE constraint/index on (channel_id, group_name).
    _, err = common.DB.GroupDB.ExecContext(
      ctx,
      `
      INSERT INTO "group" AS existing (
        channel_id,
        group_name,
        group_img,
        bio,
        alias_names
      )
      VALUES ($1, $2, $3, $4, $5)
      ON CONFLICT (channel_id, group_name)
      DO UPDATE SET
        group_img = CASE
          WHEN EXCLUDED.group_img = ''
          THEN existing.group_img
          ELSE EXCLUDED.group_img
        END,
        bio = EXCLUDED.bio,
        alias_names = EXCLUDED.alias_names
      `,
      channelID,
      g.GroupName,
      g.GroupImg,
      g.Bio,
      pq.Array(g.AliasNames),
    )
    if err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; group upsert error", http.StatusOK,
      )
      return
    }
  }

  // The group changes are stored. Queue notifications afterward.
  for _, g := range groups {
    payload := []interface{}{
      "group",
      channelID,
      updatedBy,
      []interface{}{
        g.GroupName,
        g.AliasNames,
      },
      g.GroupImg,
    }

    if err := common.ChunkPush(sessions, payload); err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; group notification queue error",
        http.StatusOK,
      )
      return
    }
  }

  responseData := struct {
    Csrf      string          `json:"csrf"`
    ChannelID int64           `json:"channelID"`
    Groups    []groupEditInput `json:"groups"`
  }{
    Csrf:      session.Csrf,
    ChannelID: channelID,
    Groups:    groups,
  }

  w.Header().Set("Content-Type", "application/json")
  _ = json.NewEncoder(w).Encode(responseData)
}