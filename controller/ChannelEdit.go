package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"

	"chat/table"
	"chat/common"
)

func ChannelEdit(w http.ResponseWriter, r *http.Request) {
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
	channelName := r.FormValue("channelName")
	channelDescription := common.SanitizeRichText(
		r.FormValue("channelDescription"),
	)
	generateInvitation := r.FormValue("generateInvitation") != ""
	guest := r.FormValue("guest") != ""

	if common.DB == nil ||
		common.DB.ChannelDB == nil ||
		common.DB.AliasDB == nil ||
		common.DB.SessionDB == nil {
		common.WriteResponseWithSession(
			w, session, "database is not initialized", http.StatusOK,
		)
		return
	}

	// Channel metadata and invitation codes require an admin alias.
	accessRight, err := channelActorAccess(
		ctx, channelID, session.UserID, updatedBy,
	)
	if err != nil {
		common.WriteResponseWithSession(
			w, session, err.Error()+"; alias authorization error", http.StatusOK,
		)
		return
	}
	if accessRight != 1 {
		common.WriteResponseWithSession(
			w, session, "Admin access required", http.StatusOK,
		)
		return
	}

	pushNames, err := parsePushNames(r)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	now := time.Now()
	invitationCode := ""
	invitationGuestCode := ""

	if generateInvitation {
		if guest {
			invitationGuestCode = common.StringRand(20)
		} else {
			invitationCode = common.StringRand(16)
		}
	}

	// channel.updated_by stores the alias name. invited_by stores user_id.
	result, err := common.DB.ChannelDB.ExecContext(
		ctx,
		`
		UPDATE channel
		SET
			channel_name = CASE
				WHEN $2 <> '' THEN $2 ELSE channel_name
			END,
			channel_description = CASE
				WHEN $3 <> '' THEN $3 ELSE channel_description
			END,
			updated_by = $4,
			updated_at = $5,
			invitation_code = CASE
				WHEN $6 AND NOT $7 THEN $8 ELSE invitation_code
			END,
			invitation_guest_code = CASE
				WHEN $6 AND $7 THEN $9 ELSE invitation_guest_code
			END,
			invited_at = CASE
				WHEN $6 THEN $5 ELSE invited_at
			END,
			invited_by = CASE
				WHEN $6 THEN $10 ELSE invited_by
			END
		WHERE channel_id = $1
		`,
		channelID,
		channelName,
		channelDescription,
		updatedBy,
		now,
		generateInvitation,
		guest,
		invitationCode,
		invitationGuestCode,
		session.UserID,
	)
	if err != nil {
		common.WriteResponseWithSession(
			w, session, err.Error()+"; channel update error", http.StatusOK,
		)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		common.WriteResponseWithSession(
			w, session, "Channel not found or not updated", http.StatusOK,
		)
		return
	}

	// Notify clients only when channel metadata changed.
	if channelName != "" || channelDescription != "" {
		sessions, err := loadSessionsForAliasNames(ctx, channelID, pushNames)
		if err != nil {
			common.WriteResponseWithSession(
				w, session, err.Error()+"; session lookup error", http.StatusOK,
			)
			return
		}

		payload := []interface{}{
			"channelEdit",
			channelID,
			updatedBy,
			[]string{channelName, channelDescription},
		}

		if err := common.ChunkPush(sessions, payload); err != nil {
			common.WriteResponseWithSession(
				w, session, err.Error()+"; push content save error", http.StatusOK,
			)
			return
		}
	}

	responseData := struct {
		Csrf               string `json:"csrf"`
		ChannelID          int64  `json:"channelID"`
		ChannelName        string `json:"channelName,omitempty"`
		ChannelDescription string `json:"channelDescription,omitempty"`
		InvitationCode     string `json:"invitationCode,omitempty"`
		InvitationGuestCode string `json:"invitationGuestCode,omitempty"`
	}{
		Csrf:                session.Csrf,
		ChannelID:           channelID,
		ChannelName:         channelName,
		ChannelDescription:  channelDescription,
		InvitationCode:      invitationCode,
		InvitationGuestCode: invitationGuestCode,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}

func parseChannelID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.FormValue("channelID"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid channelID")
	}
	return id, nil
}

func parsePushNames(r *http.Request) ([]string, error) {
	value := r.FormValue("pushNames")
	if value == "" {
		return nil, nil
	}

	var names []string
	if err := json.Unmarshal([]byte(value), &names); err != nil {
		return nil, fmt.Errorf("pushNames JSON unmarshal: %w", err)
	}
	return names, nil
}

// channelActorAccess verifies that updatedBy is an alias owned by the
// authenticated user in this channel.
func channelActorAccess(
	ctx context.Context,
	channelID int64,
	userID int64,
	updatedBy string,
) (int16, error) {
	var accessRight int16

	err := common.DB.AliasDB.QueryRowContext(
		ctx,
		`
		SELECT access_right
		FROM alias
		WHERE channel_id = $1
		  AND user_id = $2
		  AND alias_name = $3
		`,
		channelID,
		userID,
		updatedBy,
	).Scan(&accessRight)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("alias not found or does not belong to current user")
		}
		return 0, err
	}

	if accessRight < 0 || accessRight > 2 {
		return 0, fmt.Errorf("invalid alias access_right")
	}

	return accessRight, nil
}

// loadSessionsForUserIDs fetches sessions through SessionDB.
// No JOIN to UserDB or AliasDB is needed.
func loadSessionsForUserIDs(
	ctx context.Context,
	userIDs []int64,
) ([]table.SessionStruct, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	rows, err := common.DB.SessionDB.QueryContext(
		ctx,
		`
		SELECT
			session_id,
			user_id,
			COALESCE(device_id, ''),
			COALESCE(created_at, NOW()),
			COALESCE(updated_at, NOW()),
			COALESCE(push_token, ''),
			COALESCE(device_type, 0),
			COALESCE(csrf, ''),
			COALESCE(push_contents, ARRAY[]::TEXT[]),
			COALESCE(mail, ''),
			COALESCE(telephone, '')
		FROM "session"
		WHERE user_id = ANY($1::BIGINT[])
		ORDER BY updated_at ASC
		`,
		pq.Array(userIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []table.SessionStruct

	for rows.Next() {
		var s table.SessionStruct
		var pushContents pq.StringArray

		if err := rows.Scan(
			&s.SessionID,
			&s.UserID,
			&s.DeviceID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.PushToken,
			&s.DeviceType,
			&s.Csrf,
			&pushContents,
			&s.Mail,
			&s.Telephone,
		); err != nil {
			return nil, err
		}

		s.PushContents = []string(pushContents)
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

// loadSessionsForAliasNames finds distinct users through AliasDB, then loads
// their sessions through SessionDB. An empty names list means all channel users.
func loadSessionsForAliasNames(
	ctx context.Context,
	channelID int64,
	names []string,
) ([]table.SessionStruct, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if len(names) == 0 {
		rows, err = common.DB.AliasDB.QueryContext(
			ctx,
			`
			SELECT DISTINCT user_id
			FROM alias
			WHERE channel_id = $1 AND user_id IS NOT NULL
			`,
			channelID,
		)
	} else {
		rows, err = common.DB.AliasDB.QueryContext(
			ctx,
			`
			SELECT DISTINCT user_id
			FROM alias
			WHERE channel_id = $1
			  AND user_id IS NOT NULL
			  AND alias_name::text = ANY($2::text[])
			`,
			channelID,
			pq.Array(names),
		)
	}
	if err != nil {
		return nil, err
	}

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			_ = rows.Close()
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	rowsErr := rows.Err()
	_ = rows.Close()
	if rowsErr != nil {
		return nil, rowsErr
	}

	return loadSessionsForUserIDs(ctx, userIDs)
}
