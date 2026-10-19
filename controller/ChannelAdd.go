package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lib/pq"

	"chat/collection"
	"chat/common"
)

func ChannelAdd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	myname := r.FormValue("myname")
	myimg := r.FormValue("myimg")
	channelName := r.FormValue("channelName")
	channelDescription := common.SanitizeRichText(
		r.FormValue("channelDescription"),
	)

	// 1. Check session.
	session, err := common.SessionCheckTake(
		w,
		r,
		r.FormValue("csrf"),
	)
	if err != nil {
		common.WriteResponseWithoutSession(
			w,
			r.FormValue("csrf"),
			err.Error()+" session check some error",
			http.StatusOK,
		)
		return
	}

	// 2. Ensure required database connections are initialized.
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

	// 3. Check channel creation limit using the alias table.
	var channelCount int64

	err = common.DB.AliasDB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(DISTINCT channel_id)
		FROM alias
		WHERE user_id = $1
		`,
		session.UserID,
	).Scan(&channelCount)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" alias channel count error",
			http.StatusOK,
		)
		return
	}

	if channelCount >= 10 {
		common.WriteResponseWithSession(
			w,
			session,
			"すでにチャネル作成の上限です",
			http.StatusOK,
		)
		return
	}

	// 4. Fetch all of this user's sessions.
	// Used to queue notifications and send pushes to the user's devices.
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
		WHERE user_id = $1
		ORDER BY updated_at ASC
		`,
		session.UserID,
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

	var mySessions []collection.SessionStruct

	for rows.Next() {
		var s collection.SessionStruct
		var pushContents pq.StringArray

		err := rows.Scan(
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
		)
		if err != nil {
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
		mySessions = append(mySessions, s)
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

	// 5. Insert the channel.
	// PostgreSQL generates channel_id using the identity column.
	now := time.Now()

	invitationCode := common.StringRand(6)
	invitationGuestCode := common.StringRand(8)

	var channelID int64

	err = common.DB.ChannelDB.QueryRowContext(
		ctx,
		`
		INSERT INTO channel (
			channel_name,
			channel_description,
			created_at,
			updated_at,
			updated_by,
			invitation_code,
			invitation_guest_code,
			invited_at
		)
		VALUES ($1, $2, $3, $3, $4, $5, $6, $3)
		RETURNING channel_id
		`,
		channelName,
		channelDescription,
		now,
		myname,
		invitationCode,
		invitationGuestCode,
	).Scan(&channelID)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" channel insert error",
			http.StatusOK,
		)
		return
	}

	// 6. Save the alias image.
	// channelID is now available for the image path and file metadata.
	aliasImg, err := common.ImgSave(
		myimg,
		session.UserID,
		myname,
		channelID,
		0,
		1,
	)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" image save error",
			http.StatusOK,
		)
		return
	}

	// 7. Insert the initial alias.
	// alias_id is generated automatically by PostgreSQL.
	_, err = common.DB.AliasDB.ExecContext(
		ctx,
		`
		INSERT INTO alias (
			channel_id,
			user_id,
			alias_name,
			alias_img,
			access_right
		)
		VALUES ($1, $2, $3, $4, $5)
		`,
		channelID,
		session.UserID,
		myname,
		aliasImg,
		"admin",
	)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" alias insert error",
			http.StatusOK,
		)
		return
	}

	// 8. Queue the channel update for all sessions.
	// ChunkPush saves all notification content before dispatching pushes.
	contents := []string{channelName, channelDescription}

	arr := []interface{}{
		"channelEdit",
		channelID,
		myname,
		contents,
	}

	if err := common.ChunkPush(mySessions, arr); err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+" ChunkPush error",
			http.StatusOK,
		)
		return
	}

	// 9. Queue and send the new-alias notification.
	for _, d := range mySessions {
		pushID := common.StringRand(1)

		pushArr := []interface{}{
			pushID,
			"alias",
			channelID,
			myname,
			[]interface{}{
				session.UserID,
				myname,
				"",
				"admin",
			},
			aliasImg,
		}

		resp, err := common.SendWebPushNotification(
			pushArr,
			pushID,
			d,
		)
		if err != nil {
			common.WriteResponseWithSession(
				w,
				session,
				err.Error()+" alias push content save error",
				http.StatusOK,
			)
			return
		}

		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}

	// 10. Return the newly created channel.
	responseData := struct {
		ChannelID    int64    `json:"channelID"`
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		ChannelID:    channelID,
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}