package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"chat/table"
	"chat/common"

	"github.com/lib/pq"
)

// PushSubscribeMobile handles mobile push-token registration.
func PushSubscribeMobile(w http.ResponseWriter, r *http.Request) {
	pushToken := r.FormValue("pushToken")
	if pushToken == "" {
		http.Error(w, "push token is required", http.StatusBadRequest)
		return
	}

	// Default to FCM. Preserve support for deviceType 1, 2, and 3.
	deviceType := 2
	if value := r.FormValue("deviceType"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("PushSubscribeMobile: invalid deviceType %q: %v; defaulting to FCM", value, err)
		} else {
			switch parsed {
			case 1, 2, 3:
				deviceType = parsed
			default:
				log.Printf("PushSubscribeMobile: unknown deviceType %q; defaulting to FCM", value)
			}
		}
	}

	csrf := r.FormValue("csrf")
	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(
			w, csrf, err.Error()+";SessionCheckTake", http.StatusOK,
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// 1. Register the push token in PostgreSQL.
	result, err := common.DB.SessionDB.ExecContext(ctx, `
		UPDATE "session"
		SET
			push_token = $1,
			device_type = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $3
	`, pushToken, deviceType, session.SessionID)
	if err != nil {
		log.Printf("PushSubscribeMobile: update session push token: %v", err)
		http.Error(w, "Failed to register push token", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("PushSubscribeMobile: check session update: %v", err)
		http.Error(w, "Failed to register push token", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("PushSubscribeMobile: session not found: %s", session.SessionID)
		http.Error(w, "Session not found", http.StatusInternalServerError)
		return
	}

	// 2. Remove the same push token from other sessions belonging to
	// this user, so that the same device does not receive duplicate pushes.
	//
	// This cleanup is best-effort, matching the previous behavior:
	// log failures, but do not fail the registration.
	clearResult, err := common.DB.SessionDB.ExecContext(ctx, `
		UPDATE "session"
		SET
			push_token = NULL,
			device_type = 0
		WHERE user_id = $1
		  AND session_id <> $2
		  AND push_token = $3
	`, session.UserID, session.SessionID, pushToken)
	if err != nil {
		log.Printf(
			"PushSubscribeMobile: clear duplicate push tokens for user %d: %v",
			session.UserID, err,
		)
	} else {
		cleared, countErr := clearResult.RowsAffected()
		if countErr != nil {
			log.Printf("PushSubscribeMobile: read duplicate cleanup count: %v", countErr)
		} else if cleared > 0 {
			log.Printf(
				"PushSubscribeMobile: cleared duplicate push token from %d session(s) of user %d",
				cleared, session.UserID,
			)
		}
	}

	session.PushToken = pushToken
	session.DeviceType = deviceType

	// 3. Load all channels the user belongs to, including each channel's
	// aliases and groups. This data is assembled from PostgreSQL tables.
	channels, err := loadPushSubscribeMobileChannels(ctx, session.UserID)
	if err != nil {
		log.Printf(
			"PushSubscribeMobile: load channels for user %d: %v",
			session.UserID, err,
		)
		common.WriteResponseWithSession(
			w, session, err.Error()+";load channels", http.StatusOK,
		)
		return
	}

	// 4. Send the registration confirmation push after the database writes.
	// ChunkPush handles and logs push-delivery errors internally.
	pushPayload := []interface{}{
		"pushCheck",
		"FCM token registration completed",
	}
	common.ChunkPush([]table.SessionStruct{session}, pushPayload)

	// 5. Return channel, alias, and group data.
	//
	// TODO: This response may become very large because each channel includes
	// every alias and group. Consider separate endpoints and client-side
	// loading when revisiting the mobile synchronization flow.
	responseData := struct {
		Csrf         string                       `json:"csrf"`
		PushContents []string                     `json:"pushContents"`
		Channels     []pushSubscribeMobileChannel `json:"channels"`
		Success      bool                         `json:"success"`
		Message      string                       `json:"message"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Channels:     channels,
		Success:      true,
		Message:      "FCM token registered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		log.Printf("PushSubscribeMobile: encode response: %v", err)
	}
}

// Separate response structures keep myname out of channel data.
type pushSubscribeMobileChannel struct {
	ChannelID           string                       `json:"channelID"`
	ChannelName         string                       `json:"channelName"`
	ChannelDescription  string                       `json:"channelDescription"`
	CreatedAt           time.Time                    `json:"createdAt"`
	UpdatedAt           time.Time                    `json:"updatedAt"`
	UpdatedBy           string                       `json:"updatedBy"`
	InvitationCode      string                       `json:"invitationCode"`
	InvitationGuestCode string                       `json:"invitationGuestCode"`
	InvitedAt           time.Time                    `json:"invitedAt"`
	Aliases             []pushSubscribeMobileAlias   `json:"aliases"`
	Groups              []pushSubscribeMobileGroup   `json:"groups"`
}

type pushSubscribeMobileAlias struct {
	AliasID     string `json:"aliasID"`
	ChannelID   string `json:"channelID"`
	AliasName   string `json:"aliasName"`
	AliasImg    string `json:"aliasImg"`
	UserID      string `json:"userID"`
	AliasBio    string `json:"aliasBio"`
	AccessRight int16  `json:"accessRight"`
}

type pushSubscribeMobileGroup struct {
	GroupID    string   `json:"groupID"`
	ChannelID  string   `json:"channelID"`
	GroupName  string   `json:"groupName"`
	GroupImg   string   `json:"groupImg"`
	GroupBio   string   `json:"groupBio"`
	AliasNames []string `json:"aliasNames"`
}

// loadPushSubscribeMobileChannels gets the user's channels and assembles
// each channel's aliases and groups from their respective PostgreSQL databases.
func loadPushSubscribeMobileChannels(
	ctx context.Context,
	userID int,
) ([]pushSubscribeMobileChannel, error) {
	channelIDs := make([]int64, 0)

	// Only actual members are considered. Inquirers (access_right = 3)
	// cannot join as members and are excluded.
	rows, err := common.DB.AliasDB.QueryContext(ctx, `
		SELECT DISTINCT channel_id
		FROM alias
		WHERE user_id = $1
		  AND COALESCE(access_right, 0) <> 3
		ORDER BY channel_id
	`, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		channelIDs = append(channelIDs, id)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	channels := make([]pushSubscribeMobileChannel, 0, len(channelIDs))
	channelIndex := make(map[int64]int, len(channelIDs))

	if len(channelIDs) == 0 {
		return channels, nil
	}

	// Load channel metadata.
	channelRows, err := common.DB.ChannelDB.QueryContext(ctx, `
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
		return nil, err
	}

	for channelRows.Next() {
		var (
			id                  int64
			name                string
			description         sql.NullString
			createdAt           sql.NullTime
			updatedAt           sql.NullTime
			updatedBy            sql.NullString
			invitationCode      sql.NullString
			invitationGuestCode sql.NullString
			invitedAt           sql.NullTime
		)

		if err := channelRows.Scan(
			&id,
			&name,
			&description,
			&createdAt,
			&updatedAt,
			&updatedBy,
			&invitationCode,
			&invitationGuestCode,
			&invitedAt,
		); err != nil {
			channelRows.Close()
			return nil, err
		}

		ch := pushSubscribeMobileChannel{
			ChannelID:           strconv.FormatInt(id, 10),
			ChannelName:         name,
			ChannelDescription:  pushSubscribeMobileNullString(description),
			UpdatedBy:           pushSubscribeMobileNullString(updatedBy),
			InvitationCode:      pushSubscribeMobileNullString(invitationCode),
			InvitationGuestCode: pushSubscribeMobileNullString(invitationGuestCode),
			Aliases:             make([]pushSubscribeMobileAlias, 0),
			Groups:              make([]pushSubscribeMobileGroup, 0),
		}

		if createdAt.Valid {
			ch.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			ch.UpdatedAt = updatedAt.Time
		}
		if invitedAt.Valid {
			ch.InvitedAt = invitedAt.Time
		}

		channelIndex[id] = len(channels)
		channels = append(channels, ch)
	}

	if err := channelRows.Err(); err != nil {
		channelRows.Close()
		return nil, err
	}
	channelRows.Close()

	// Load all aliases for the user's channels.
	// Unlike web PushSubscribe, mobile retains alias owner IDs, preserving
	// the previous sanitizeMyChannel(..., true) behavior.
	aliasRows, err := common.DB.AliasDB.QueryContext(ctx, `
		SELECT
			alias_id,
			channel_id,
			alias_name,
			alias_img,
			user_id,
			alias_bio,
			access_right
		FROM alias
		WHERE channel_id = ANY($1::BIGINT[])
		ORDER BY channel_id, alias_name
	`, pq.Array(channelIDs))
	if err != nil {
		return nil, err
	}

	for aliasRows.Next() {
		var (
			aliasID     string
			channelID   int64
			aliasName   string
			aliasImg    sql.NullString
			aliasUserID sql.NullInt64
			aliasBio    sql.NullString
			accessRight sql.NullInt64
		)

		if err := aliasRows.Scan(
			&aliasID,
			&channelID,
			&aliasName,
			&aliasImg,
			&aliasUserID,
			&aliasBio,
			&accessRight,
		); err != nil {
			aliasRows.Close()
			return nil, err
		}

		index, ok := channelIndex[channelID]
		if !ok {
			// The channel may have been removed between queries.
			continue
		}

		var ownerID string
		if aliasUserID.Valid {
			ownerID = strconv.FormatInt(aliasUserID.Int64, 10)
		}

		var right int16
		if accessRight.Valid {
			right = int16(accessRight.Int64)
		}

		channels[index].Aliases = append(
			channels[index].Aliases,
			pushSubscribeMobileAlias{
				AliasID:     aliasID,
				ChannelID:   strconv.FormatInt(channelID, 10),
				AliasName:   aliasName,
				AliasImg:    pushSubscribeMobileNullString(aliasImg),
				UserID:      ownerID,
				AliasBio:    pushSubscribeMobileNullString(aliasBio),
				AccessRight: right,
			},
		)
	}

	if err := aliasRows.Err(); err != nil {
		aliasRows.Close()
		return nil, err
	}
	aliasRows.Close()

	// Load all groups for these channels.
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
		return nil, err
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
			return nil, err
		}

		index, ok := channelIndex[channelID]
		if !ok {
			continue
		}

		names := []string(aliasNames)
		if names == nil {
			names = make([]string, 0)
		}

		channels[index].Groups = append(
			channels[index].Groups,
			pushSubscribeMobileGroup{
				GroupID:    groupID,
				ChannelID:  strconv.FormatInt(channelID, 10),
				GroupName:  groupName,
				GroupImg:   pushSubscribeMobileNullString(groupImg),
				GroupBio:   pushSubscribeMobileNullString(groupBio),
				AliasNames: names,
			},
		)
	}

	if err := groupRows.Err(); err != nil {
		groupRows.Close()
		return nil, err
	}
	groupRows.Close()

	return channels, nil
}

func pushSubscribeMobileNullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}