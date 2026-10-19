package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"

	"chat/collection"
	"chat/common"
)

func ChannelDelete(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	channelID, err := strconv.ParseInt(r.FormValue("channelID"), 10, 64)
	if err != nil || channelID <= 0 {
		http.Error(w, "invalid channelID", http.StatusBadRequest)
		return
	}

	updatedBy := r.FormValue("updatedBy")

	// 1. Session check.
	session, err := common.SessionCheckTake(
		w,
		r,
		r.FormValue("csrf"),
	)
	if err != nil {
		common.WriteResponseWithoutSession(
			w,
			r.FormValue("csrf"),
			"Session Error",
			http.StatusOK,
		)
		return
	}

	// 2. Check database connections.
	if common.DB == nil ||
		common.DB.ChannelDB == nil ||
		common.DB.AliasDB == nil ||
		common.DB.SessionDB == nil {
		common.WriteResponseWithSession(
			w,
			session,
			"database is not initialized",
			http.StatusOK,
		)
		return
	}

	// 3. Confirm that the channel exists.
	var channelExists bool

	err = common.DB.ChannelDB.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1 FROM channel WHERE channel_id = $1
		)`,
		channelID,
	).Scan(&channelExists)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" channel lookup error",
			http.StatusOK,
		)
		return
	}

	if !channelExists {
		common.WriteResponseWithSession(
			w,
			session,
			"Channel not found",
			http.StatusOK,
		)
		return
	}

	// 4. Verify that the requesting user owns this admin alias.
	var isAdmin bool

	err = common.DB.AliasDB.QueryRowContext(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM alias
			WHERE channel_id = $1
			  AND user_id = $2
			  AND alias_name = $3
			  AND access_right = 'admin'
		)
		`,
		channelID,
		session.UserID,
		updatedBy,
	).Scan(&isAdmin)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" admin check error",
			http.StatusOK,
		)
		return
	}

	if !isAdmin {
		common.WriteResponseWithSession(
			w,
			session,
			"No admin access right",
			http.StatusOK,
		)
		return
	}

	// 5. Get the distinct users who have aliases in this channel.
	// These users' sessions need the deletion notification.
	userRows, err := common.DB.AliasDB.QueryContext(
		ctx,
		`
		SELECT DISTINCT user_id
		FROM alias
		WHERE channel_id = $1
		  AND user_id IS NOT NULL
		`,
		channelID,
	)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" alias user lookup error",
			http.StatusOK,
		)
		return
	}

	var userIDs []int64

	for userRows.Next() {
		var userID int64

		if err := userRows.Scan(&userID); err != nil {
			_ = userRows.Close()
			common.WriteResponseWithSession(
				w,
				session,
				err.Error()+" alias user scan error",
				http.StatusOK,
			)
			return
		}

		userIDs = append(userIDs, userID)
	}

	rowsErr := userRows.Err()
	_ = userRows.Close()

	if rowsErr != nil {
		common.WriteResponseWithSession(
			w,
			session,
			rowsErr.Error()+" alias user rows error",
			http.StatusOK,
		)
		return
	}

	// 6. Fetch the sessions belonging to those users.
	var sessions []collection.SessionStruct

	if len(userIDs) > 0 {
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
			common.WriteResponseWithSession(
				w,
				session,
				err.Error()+" session query error",
				http.StatusOK,
			)
			return
		}

		for rows.Next() {
			var s collection.SessionStruct
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
				_ = rows.Close()
				common.WriteResponseWithSession(
					w,
					session,
					err.Error()+" session scan error",
					http.StatusOK,
				)
				return
			}

			s.PushContents = []string(pushContents)
			sessions = append(sessions, s)
		}

		rowsErr := rows.Err()
		_ = rows.Close()

		if rowsErr != nil {
			common.WriteResponseWithSession(
				w,
				session,
				rowsErr.Error()+" session rows error",
				http.StatusOK,
			)
			return
		}
	}

	// 7. Delete the channel.
	// Cross-database operations cannot share one PostgreSQL transaction.
	result, err := common.DB.ChannelDB.ExecContext(
		ctx,
		`DELETE FROM channel WHERE channel_id = $1`,
		channelID,
	)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" channel delete error",
			http.StatusOK,
		)
		return
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" channel delete result error",
			http.StatusOK,
		)
		return
	}

	if deleted == 0 {
		common.WriteResponseWithSession(
			w,
			session,
			"Channel not found",
			http.StatusOK,
		)
		return
	}

	// 8. Delete the channel's aliases explicitly because there is no FK cascade.
	_, err = common.DB.AliasDB.ExecContext(
		ctx,
		`DELETE FROM alias WHERE channel_id = $1`,
		channelID,
	)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" alias delete error",
			http.StatusOK,
		)
		return
	}

	// 10. Queue the channel deletion notification.
	// ChunkPush writes notification content first. If the DB write fails,
	// it returns an error and does not send the pushes.
	channelPushArray := []interface{}{
		"channelEdit",
		channelID,
		updatedBy,
		"delete",
	}

	if err := common.ChunkPush(sessions, channelPushArray); err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" ChunkPush error",
			http.StatusOK,
		)
		return
	}

	// 11. Return the response.
	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}