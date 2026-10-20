package controller

import (
  "context"
  "database/sql"
  "encoding/json"
  "fmt"
  "net/http"
  "time"

  "chat/common"
)

type aliasEditInput struct {
  AliasName   string  `json:"aliasName"`
  AliasImg    *string `json:"aliasImg"`
  AliasBio    *string `json:"aliasBio"`
  AccessRight *int16  `json:"accessRight"`
}

type aliasRecord struct {
  AliasID     int64
  UserID      int64
  AccessRight int16
  AliasImg    string
  AliasBio    string
}

func getAliasRecord(
  ctx context.Context,
  channelID int64,
  aliasName string,
) (aliasRecord, error) {
  var a aliasRecord

  err := common.DB.AliasDB.QueryRowContext(
    ctx,
    `
    SELECT
      alias_id,
      COALESCE(user_id, 0),
      access_right,
      COALESCE(alias_img, ''),
      COALESCE(alias_bio, '')
    FROM alias
    WHERE channel_id = $1
      AND alias_name = $2
    `,
    channelID,
    aliasName,
  ).Scan(
    &a.AliasID,
    &a.UserID,
    &a.AccessRight,
    &a.AliasImg,
    &a.AliasBio,
  )

  return a, err
}

func AliasEdit(w http.ResponseWriter, r *http.Request) {
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
    common.DB.AliasDB == nil ||
    common.DB.GroupDB == nil ||
    common.DB.SessionDB == nil {
    common.WriteResponseWithSession(
      w, session, "database is not initialized", http.StatusOK,
    )
    return
  }

  actorRole, err := channelActorAccess(
    ctx, channelID, session.UserID, updatedBy,
  )
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
  }
  if actorRole == 2 {
    common.WriteResponseWithSession(
      w, session, "Guest access cannot edit aliases", http.StatusOK,
    )
    return
  }

  // Read the alias edit payload. "admin" is temporarily accepted for
  // compatibility with the old combined endpoint.
  aliasJSON := r.FormValue("alias")
  if aliasJSON == "" {
    aliasJSON = r.FormValue("admin")
  }

  var input aliasEditInput
  hasAliasEdit := aliasJSON != ""
  if hasAliasEdit {
    if err := json.Unmarshal([]byte(aliasJSON), &input); err != nil {
      common.WriteResponseWithSession(
        w, session, "alias JSON unmarshal error", http.StatusOK,
      )
      return
    }
    if input.AliasName == "" {
      common.WriteResponseWithSession(
        w, session, "aliasName is required", http.StatusOK,
      )
      return
    }
  }

  var deleteAliases []string
  if value := r.FormValue("deleteAliases"); value != "" {
    if err := json.Unmarshal([]byte(value), &deleteAliases); err != nil {
      common.WriteResponseWithSession(
        w, session, "deleteAliases JSON unmarshal error", http.StatusOK,
      )
      return
    }
  }

  if !hasAliasEdit && len(deleteAliases) == 0 {
    common.WriteResponseWithSession(
      w, session, "no alias changes supplied", http.StatusOK,
    )
    return
  }

  pushNames, err := parsePushNames(r)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
  }

  // Capture recipients before deleting any aliases.
  sessions, err := loadSessionsForAliasNames(ctx, channelID, pushNames)
  if err != nil {
    common.WriteResponseWithSession(
      w, session, err.Error()+"; session lookup error", http.StatusOK,
    )
    return
  }

  // Edit an existing alias. The alias_name itself is never updated.
  var edited *aliasRecord

  if hasAliasEdit {
    target, err := getAliasRecord(ctx, channelID, input.AliasName)
    if err != nil {
      message := "alias not found"
      if err != sql.ErrNoRows {
        message = err.Error()
      }
      common.WriteResponseWithSession(w, session, message, http.StatusOK)
      return
    }

    // Normal users may edit their own alias. Admins may edit others.
    if target.UserID != session.UserID && actorRole != 1 {
      common.WriteResponseWithSession(
        w, session, "Admin access required to edit another user's alias",
        http.StatusOK,
      )
      return
    }

    newAccessRight := target.AccessRight
    if input.AccessRight != nil {
      if *input.AccessRight < 0 || *input.AccessRight > 2 {
        common.WriteResponseWithSession(
          w, session, "invalid accessRight", http.StatusOK,
        )
        return
      }
      if actorRole != 1 && *input.AccessRight != target.AccessRight {
        common.WriteResponseWithSession(
          w, session, "Admin access required to change accessRight",
          http.StatusOK,
        )
        return
      }
      newAccessRight = *input.AccessRight
    }

    newImg := target.AliasImg
    if input.AliasImg != nil {
      newImg = *input.AliasImg
      if newImg != "" {
        newImg, err = common.ImgSave(
          newImg,
          session.UserID,
          input.AliasName,
          channelID,
          3,
          1,
        )
        if err != nil {
          common.WriteResponseWithSession(
            w, session, err.Error()+"; alias image save error",
            http.StatusOK,
          )
          return
        }
      }
    }

    newBio := target.AliasBio
    if input.AliasBio != nil {
      newBio = *input.AliasBio
    }

    _, err = common.DB.AliasDB.ExecContext(
      ctx,
      `
      UPDATE alias
      SET alias_img = $1,
          alias_bio = $2,
          access_right = $3
      WHERE alias_id = $4
      `,
      newImg,
      newBio,
      newAccessRight,
      target.AliasID,
    )
    if err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; alias update error", http.StatusOK,
      )
      return
    }

    target.AliasImg = newImg
    target.AliasBio = newBio
    target.AccessRight = newAccessRight
    edited = &target
  }

  // Delete aliases. Owners can remove their own aliases; admins can
  // remove other users' aliases and other admins' aliases.
  deletedNames := make([]string, 0, len(deleteAliases))

  for _, name := range deleteAliases {
    if name == "" {
      continue
    }

    target, err := getAliasRecord(ctx, channelID, name)
    if err != nil {
      message := "alias not found: " + name
      if err != sql.ErrNoRows {
        message = err.Error()
      }
      common.WriteResponseWithSession(w, session, message, http.StatusOK)
      return
    }

    if actorRole != 1 {
      if target.UserID != session.UserID {
        common.WriteResponseWithSession(
          w, session, "Admin access required to delete another user's alias",
          http.StatusOK,
        )
        return
      }
      if target.AccessRight == 1 {
        common.WriteResponseWithSession(
          w, session, "Admin access required to delete an admin alias",
          http.StatusOK,
        )
        return
      }
    }

    _, err = common.DB.AliasDB.ExecContext(
      ctx,
      `DELETE FROM alias WHERE alias_id = $1`,
      target.AliasID,
    )
    if err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; alias delete error", http.StatusOK,
      )
      return
    }

    // Aliases and groups may be hosted separately, so this cleanup is
    // a separate operation, not part of the alias deletion transaction.
    _, err = common.DB.GroupDB.ExecContext(
      ctx,
      `
      UPDATE "group"
      SET alias_names = array_remove(
        COALESCE(alias_names, ARRAY[]::TEXT[]),
        $2
      )
      WHERE channel_id = $1
      `,
      channelID,
      name,
    )
    if err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; group member cleanup error", http.StatusOK,
      )
      return
    }

    _, err = common.DB.GroupDB.ExecContext(
      ctx,
      `
      DELETE FROM "group"
      WHERE channel_id = $1
        AND cardinality(COALESCE(alias_names, ARRAY[]::TEXT[])) = 0
      `,
      channelID,
    )
    if err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; empty group cleanup error", http.StatusOK,
      )
      return
    }

    deletedNames = append(deletedNames, name)

    // If this user no longer has any alias in the channel, notify their
    // sessions to remove the channel from their local view.
    var hasAnotherAlias bool
    err = common.DB.AliasDB.QueryRowContext(
      ctx,
      `
      SELECT EXISTS (
        SELECT 1 FROM alias
        WHERE channel_id = $1 AND user_id = $2
      )
      `,
      channelID,
      target.UserID,
    ).Scan(&hasAnotherAlias)
    if err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; remaining alias check error",
        http.StatusOK,
      )
      return
    }

    if !hasAnotherAlias {
      userSessions, err := loadSessionsForUserIDs(
        ctx, []int64{target.UserID},
      )
      if err != nil {
        common.WriteResponseWithSession(
          w, session, err.Error()+"; user session lookup error",
          http.StatusOK,
        )
        return
      }

      if err := common.ChunkPush(userSessions, []interface{}{
        "channelEdit", channelID, updatedBy, "delete",
      }); err != nil {
        common.WriteResponseWithSession(
          w, session, err.Error()+"; channel delete notification queue error",
          http.StatusOK,
        )
        return
      }
    }
  }

  // Queue update/delete notifications after the database operations.
  if edited != nil {
    payload := []interface{}{
      "alias",
      channelID,
      updatedBy,
      []interface{}{
        edited.UserID,
        input.AliasName,
        edited.AliasBio,
        edited.AccessRight,
      },
      edited.AliasImg,
    }

    if err := common.ChunkPush(sessions, payload); err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; alias notification queue error",
        http.StatusOK,
      )
      return
    }
  }

  for _, name := range deletedNames {
    payload := []interface{}{
      "alias",
      channelID,
      updatedBy,
      []interface{}{"", name, "", "delete"},
      "",
    }

    if err := common.ChunkPush(sessions, payload); err != nil {
      common.WriteResponseWithSession(
        w, session, err.Error()+"; alias delete notification queue error",
        http.StatusOK,
      )
      return
    }
  }

  responseData := struct {
    Csrf          string   `json:"csrf"`
    ChannelID     int64    `json:"channelID"`
    AliasName     string   `json:"aliasName,omitempty"`
    AliasImg      string   `json:"aliasImg,omitempty"`
    AliasBio      string   `json:"aliasBio,omitempty"`
    AccessRight   *int16   `json:"accessRight,omitempty"`
    DeletedAliases []string `json:"deletedAliases,omitempty"`
  }{
    Csrf:           session.Csrf,
    ChannelID:      channelID,
    DeletedAliases: deletedNames,
  }

  if edited != nil {
    responseData.AliasName = input.AliasName
    responseData.AliasImg = edited.AliasImg
    responseData.AliasBio = edited.AliasBio
    responseData.AccessRight = &edited.AccessRight
  }

  w.Header().Set("Content-Type", "application/json")
  _ = json.NewEncoder(w).Encode(responseData)
}