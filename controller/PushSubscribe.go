package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"chat/collection"
	"chat/common"

	"github.com/lib/pq"
)

// Temporary response structures for assembling channel data from PostgreSQL.
// Myname is intentionally omitted because it is no longer stored in channel.
type pushSubscribeChannel struct {
	ChannelID            string                  `json:"channelID"`
	ChannelName          string                  `json:"channelName"`
	ChannelDescription   string                  `json:"channelDescription"`
	CreatedAt            time.Time               `json:"createdAt"`
	UpdatedAt            time.Time               `json:"updatedAt"`
	UpdatedBy            string                  `json:"updatedBy"`
	InvitationCode       string                  `json:"invitationCode"`
	InvitationGuestCode  string                  `json:"invitationGuestCode"`
	InvitedAt            time.Time               `json:"invitedAt"`
	Aliases              []pushSubscribeAlias    `json:"aliases"`
	Groups               []pushSubscribeGroup    `json:"groups"`
}

type pushSubscribeAlias struct {
	AliasID     string `json:"aliasID"`
	ChannelID   string `json:"channelID"`
	AliasName   string `json:"aliasName"`
	AliasImg    string `json:"aliasImg"`
	UserID      string `json:"userID"`
	AliasBio    string `json:"aliasBio"`
	AccessRight int16  `json:"accessRight"`
}

type pushSubscribeGroup struct {
	GroupID    string   `json:"groupID"`
	ChannelID  string   `json:"channelID"`
	GroupName  string   `json:"groupName"`
	GroupImg   string   `json:"groupImg"`
	GroupBio   string   `json:"groupBio"`
	AliasNames []string `json:"aliasNames"`
}

func PushSubscribe(w http.ResponseWriter, r *http.Request) {
	csrf := r.FormValue("csrf")

	// Accept both current and legacy field names.
	pushToken := r.FormValue("pushToken")
	if pushToken == "" {
		pushToken = r.FormValue("subscription")
	}

	if !json.Valid([]byte(pushToken)) {
		log.Printf("PushSubscribe: invalid JSON push token; path=%s", r.URL.Path)
		http.Error(w, "Invalid JSON push token", http.StatusBadRequest)
		return
	}

	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(
			w,
			csrf,
			err.Error()+";SessionCheckTake",
			http.StatusOK,
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// 1. Find the channels this user belongs to.
	// Inquirers (access_right = 3) are not considered members.
	channelRows, err := common.DB.AliasDB.QueryContext(ctx, `
		SELECT DISTINCT channel_id
		FROM alias
		WHERE user_id = $1
		  AND COALESCE(access_right, 0) <> 3
		ORDER BY channel_id
	`, session.UserID)
	if err != nil {
		log.Printf("PushSubscribe: query user's channels: %v", err)
		common.WriteResponseWithSession(
			w, session, err.Error()+";query user's channels", http.StatusOK,
		)
		return
	}

	channelIDs := make([]int64, 0)
	for channelRows.Next() {
		var channelID int64
		if err := channelRows.Scan(&channelID); err != nil {
			channelRows.Close()
			log.Printf("PushSubscribe: scan channel ID: %v", err)
			common.WriteResponseWithSession(
				w, session, err.Error()+";scan channel ID", http.StatusOK,
			)
			return
		}
		channelIDs = append(channelIDs, channelID)
	}

	if err := channelRows.Err(); err != nil {
		channelRows.Close()
		log.Printf("PushSubscribe: iterate channel IDs: %v", err)
		common.WriteResponseWithSession(
			w, session, err.Error()+";iterate channel IDs", http.StatusOK,
		)
		return
	}
	channelRows.Close()

	// Use non-nil slices so the JSON response contains [] rather than null.
	channels := make([]pushSubscribeChannel, 0)
	channelIndex := make(map[int64]int)

	// The user's channel list can be empty, so only query channel data
	// when we have IDs.
	if len(channelIDs) > 0 {
		// 2. Load channel metadata.
		rows, err := common.DB.ChannelDB.QueryContext(ctx, `
			SELECT
				channel_id,
				channel_name,
				channel_description,
				created_at,
				updated_at,
				updated_by,
				invitation_code,
				invitation_guest_code,
				invited_at
			FROM channel
			WHERE channel_id = ANY($1::BIGINT[])
			ORDER BY channel_id
		`, pq.Array(channelIDs))
		if err != nil {
			log.Printf("PushSubscribe: query channel data: %v", err)
			common.WriteResponseWithSession(
				w, session, err.Error()+";query channel data", http.StatusOK,
			)
			return
		}

		for rows.Next() {
			var (
				channelID           int64
				channelName         string
				channelDescription  sql.NullString
				createdAt           sql.NullTime
				updatedAt           sql.NullTime
				updatedBy            sql.NullString
				invitationCode      sql.NullString
				invitationGuestCode sql.NullString
				invitedAt           sql.NullTime
			)

			if err := rows.Scan(
				&channelID,
				&channelName,
				&channelDescription,
				&createdAt,
				&updatedAt,
				&updatedBy,
				&invitationCode,
				&invitationGuestCode,
				&invitedAt,
			); err != nil {
				rows.Close()
				log.Printf("PushSubscribe: scan channel: %v", err)
				common.WriteResponseWithSession(
					w, session, err.Error()+";scan channel", http.StatusOK,
				)
				return
			}

			channel := pushSubscribeChannel{
				ChannelID:           strconv.FormatInt(channelID, 10),
				ChannelName:         channelName,
				ChannelDescription:  nullStringToString(channelDescription),
				UpdatedBy:           nullStringToString(updatedBy),
				InvitationCode:      nullStringToString(invitationCode),
				InvitationGuestCode: nullStringToString(invitationGuestCode),
				Aliases:             make([]pushSubscribeAlias, 0),
				Groups:              make([]pushSubscribeGroup, 0),
			}

			if createdAt.Valid {
				channel.CreatedAt = createdAt.Time
			}
			if updatedAt.Valid {
				channel.UpdatedAt = updatedAt.Time
			}
			if invitedAt.Valid {
				channel.InvitedAt = invitedAt.Time
			}

			channelIndex[channelID] = len(channels)
			channels = append(channels, channel)
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			log.Printf("PushSubscribe: iterate channels: %v", err)
			common.WriteResponseWithSession(
				w, session, err.Error()+";iterate channels", http.StatusOK,
			)
			return
		}
		rows.Close()

		// 3. Load aliases for these channels.
		// UserID is deliberately returned as an empty string to preserve
		// the existing behavior of hiding alias ownership identifiers.
		aliasRows, err := common.DB.AliasDB.QueryContext(ctx, `
			SELECT
				alias_id,
				channel_id,
				alias_name,
				alias_img,
				alias_bio,
				access_right
			FROM alias
			WHERE channel_id = ANY($1::BIGINT[])
			ORDER BY channel_id, alias_name
		`, pq.Array(channelIDs))
		if err != nil {
			log.Printf("PushSubscribe: query aliases: %v", err)
			common.WriteResponseWithSession(
				w, session, err.Error()+";query aliases", http.StatusOK,
			)
			return
		}

		for aliasRows.Next() {
			var (
				aliasID     string
				channelID   int64
				aliasName   string
				aliasImg    sql.NullString
				aliasBio    sql.NullString
				accessRight sql.NullInt64
			)

			if err := aliasRows.Scan(
				&aliasID,
				&channelID,
				&aliasName,
				&aliasImg,
				&aliasBio,
				&accessRight,
			); err != nil {
				aliasRows.Close()
				log.Printf("PushSubscribe: scan alias: %v", err)
				common.WriteResponseWithSession(
					w, session, err.Error()+";scan alias", http.StatusOK,
				)
				return
			}

			index, ok := channelIndex[channelID]
			if !ok {
				// The channel may have been removed between queries.
				continue
			}

			var right int16
			if accessRight.Valid {
				right = int16(accessRight.Int64)
			}

			channels[index].Aliases = append(
				channels[index].Aliases,
				pushSubscribeAlias{
					AliasID:     aliasID,
					ChannelID:   strconv.FormatInt(channelID, 10),
					AliasName:   aliasName,
					AliasImg:    nullStringToString(aliasImg),
					UserID:      "",
					AliasBio:    nullStringToString(aliasBio),
					AccessRight: right,
				},
			)
		}

		if err := aliasRows.Err(); err != nil {
			aliasRows.Close()
			log.Printf("PushSubscribe: iterate aliases: %v", err)
			common.WriteResponseWithSession(
				w, session, err.Error()+";iterate aliases", http.StatusOK,
			)
			return
		}
		aliasRows.Close()

		// 4. Load groups for these channels.
		groupRows, err := common.DB.GroupDB.QueryContext(ctx, `
			SELECT
				group_id::TEXT,
				channel_id,
				group_name,
				group_img,
				group_bio,
				alias_names
			FROM "group"
			WHERE channel_id = ANY($1::BIGINT[])
			ORDER BY channel_id, group_id
		`, pq.Array(channelIDs))
		if err != nil {
			log.Printf("PushSubscribe: query groups: %v", err)
			common.WriteResponseWithSession(
				w, session, err.Error()+";query groups", http.StatusOK,
			)
			return
		}

		for groupRows.Next() {
			var (
				groupID    string
				channelID  int64
				groupName  string
				groupImg   sql.NullString
				groupBio   sql.NullString
				aliasNames pq.StringArray
			)

			if err := groupRows.Scan(
				&groupID,
				&channelID,
				&groupName,
				&groupImg,
				&groupBio,
				&aliasNames,
			); err != nil {
				groupRows.Close()
				log.Printf("PushSubscribe: scan group: %v", err)
				common.WriteResponseWithSession(
					w, session, err.Error()+";scan group", http.StatusOK,
				)
				return
			}

			index, ok := channelIndex[channelID]
			if !ok {
				continue
			}

			channels[index].Groups = append(
				channels[index].Groups,
				pushSubscribeGroup{
					GroupID:    groupID,
					ChannelID:  strconv.FormatInt(channelID, 10),
					GroupName:  groupName,
					GroupImg:   nullStringToString(groupImg),
					GroupBio:   nullStringToString(groupBio),
					AliasNames: []string(aliasNames),
				},
			)
		}

		if err := groupRows.Err(); err != nil {
			groupRows.Close()
			log.Printf("PushSubscribe: iterate groups: %v", err)
			common.WriteResponseWithSession(
				w, session, err.Error()+";iterate groups", http.StatusOK,
			)
			return
		}
		groupRows.Close()
	}

	// 5. Register the Web Push subscription in PostgreSQL.
	result, err := common.DB.SessionDB.ExecContext(ctx, `
		UPDATE "session"
		SET
			push_token = $1,
			device_type = 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $2
	`, pushToken, session.SessionID)
	if err != nil {
		log.Printf("PushSubscribe: update session push token: %v", err)
		common.WriteResponseWithSession(
			w, session, err.Error()+";update session push token", http.StatusOK,
		)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("PushSubscribe: check session update result: %v", err)
		common.WriteResponseWithSession(
			w, session, err.Error()+";check session update", http.StatusOK,
		)
		return
	}
	if rowsAffected == 0 {
		log.Printf("PushSubscribe: session not found: %s", session.SessionID)
		common.WriteResponseWithSession(
			w, session, "Session not found", http.StatusOK,
		)
		return
	}

	// Update the in-memory copy only after the database update succeeds.
	session.PushToken = pushToken
	session.DeviceType = 1

	// 6. Send the registration test notification after the DB write.
	// ChunkPush handles/logs push delivery errors internally.
	pushPayload := []interface{}{
		"pushCheck",
		"push登録完了",
	}
	common.ChunkPush([]collection.SessionStruct{session}, pushPayload)

	// 7. Respond with channel, alias and group data.
	//
	// TODO: This response may become very large because each channel
	// includes all of its aliases and groups. Keep it for compatibility
	// now; consider splitting channel, alias and group data into separate
	// endpoints when updating the client.
	responseData := struct {
		Csrf         string                 `json:"csrf"`
		PushContents []string               `json:"pushContents"`
		Channels     []pushSubscribeChannel `json:"channels"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Channels:     channels,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		log.Printf("PushSubscribe: encode response: %v", err)
	}
}

func nullStringToString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}