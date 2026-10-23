package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"chat/table"
	"chat/common"

	"github.com/lib/pq"
)

func ContentsPush(w http.ResponseWriter, r *http.Request) {
	csrf := r.FormValue("csrf")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		common.WriteResponseWithoutSession(
			w, csrf, err.Error()+";files more than 10MB", http.StatusOK,
		)
		return
	}

	var pushNames []string
	if err := json.Unmarshal([]byte(r.FormValue("pushNames")), &pushNames); err != nil {
		common.WriteResponseWithoutSession(
			w, csrf, err.Error()+";Invalid JSON pushNames", http.StatusOK,
		)
		return
	}

	updatedBy := r.FormValue("updatedBy")
	channelID := r.FormValue("channelID")
	pushTitle := r.FormValue("pushTitle")
	notifyURL := r.FormValue("notifyUrl")
	notifyMessage := r.FormValue("notifyMessage")
	notifyIcon := r.FormValue("notifyIcon")

	var contents interface{}
	if err := json.Unmarshal([]byte(r.FormValue("contents")), &contents); err != nil {
		common.WriteResponseWithoutSession(
			w, csrf, err.Error()+";Invalid JSON contents", http.StatusOK,
		)
		return
	}

	// PostgreSQL channel.channel_id is BIGINT.
	var channelIDInt int64
	if _, err := json.Number(channelID).Int64(); err != nil {
		common.WriteResponseWithoutSession(
			w, csrf, "Invalid channelID", http.StatusOK,
		)
		return
	} else {
		channelIDInt, _ = json.Number(channelID).Int64()
	}

	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(
			w, csrf, err.Error()+";SessionCheckTake", http.StatusOK,
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// 1. Validate updatedBy against the alias table.
	// The alias must belong to this user and channel.
	// Inquirers (access_right = 3) are not allowed to post.
	var hasAccess bool
	err = common.DB.AliasDB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM alias
			WHERE channel_id = $1
			  AND alias_name = $2
			  AND user_id = $3
			  AND COALESCE(access_right, 0) <> 3
		)
	`, channelIDInt, updatedBy, session.UserID).Scan(&hasAccess)
	if err != nil {
		log.Printf(
			"ContentsPush: check alias access: %v; channelID=%s; updatedBy=%s",
			err, channelID, updatedBy,
		)
		common.WriteResponseWithSession(
			w, session, err.Error()+";check alias access", http.StatusOK,
		)
		return
	}

	if !hasAccess {
		common.WriteResponseWithSession(
			w, session, "no true access right", http.StatusOK,
		)
		return
	}

	// 2. Find the recipient user IDs from the selected aliases.
	// Exclude inquirers and deduplicate users who own multiple aliases.
	rows, err := common.DB.AliasDB.QueryContext(ctx, `
		SELECT DISTINCT user_id
		FROM alias
		WHERE channel_id = $1
		  AND alias_name = ANY($2::TEXT[])
		  AND user_id IS NOT NULL
		  AND COALESCE(access_right, 0) <> 3
	`, channelIDInt, pq.Array(pushNames))
	if err != nil {
		log.Printf(
			"ContentsPush: query recipient aliases: %v; channelID=%s",
			err, channelID,
		)
		common.WriteResponseWithSession(
			w, session, err.Error()+";find recipient aliases", http.StatusOK,
		)
		return
	}

	userIDs := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			rows.Close()
			log.Printf("ContentsPush: scan recipient user ID: %v", err)
			common.WriteResponseWithSession(
				w, session, err.Error()+";scan recipient user ID", http.StatusOK,
			)
			return
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		log.Printf("ContentsPush: iterate recipient aliases: %v", err)
		common.WriteResponseWithSession(
			w, session, err.Error()+";iterate recipient aliases", http.StatusOK,
		)
		return
	}
	rows.Close()

	// Do not save attachments or send notifications if no aliases matched.
	if len(userIDs) == 0 {
		common.WriteResponseWithSession(
			w, session, "該当するネームは存在しません", http.StatusOK,
		)
		return
	}

	// Keep the existing daily log.
	contentsPush := common.NewDailyLogger("contents_push_channel_id")
	contentsPush.Printf(channelID)

	// 3. Find recipient sessions.
	// Sessions no longer store channel aliases, so no FilterSessionsByChannelID.
	sessionRows, err := common.DB.SessionDB.QueryContext(ctx, `
		SELECT
			session_id,
			user_id,
			COALESCE(push_token, ''),
			COALESCE(device_type, 0),
			COALESCE(csrf, ''),
			COALESCE(push_contents, ARRAY[]::TEXT[])
		FROM "session"
		WHERE user_id = ANY($1::BIGINT[])
	`, pq.Array(userIDs))
	if err != nil {
		log.Printf("ContentsPush: query sessions: %v; userIDs=%v", err, userIDs)
		common.WriteResponseWithSession(
			w, session, err.Error()+";find recipient sessions", http.StatusOK,
		)
		return
	}

	sessions := make([]table.SessionStruct, 0)
	for sessionRows.Next() {
		var (
			sessionID    string
			userID       int64
			pushToken    string
			deviceType   int
			sessionCsrf  string
			pushContents []string
		)

		if err := sessionRows.Scan(
			&sessionID,
			&userID,
			&pushToken,
			&deviceType,
			&sessionCsrf,
			pq.Array(&pushContents),
		); err != nil {
			sessionRows.Close()
			log.Printf("ContentsPush: scan session: %v", err)
			common.WriteResponseWithSession(
				w, session, err.Error()+";scan recipient session", http.StatusOK,
			)
			return
		}

		sessions = append(sessions, table.SessionStruct{
			SessionID:    sessionID,
			UserID:       int(userID),
			PushToken:    pushToken,
			DeviceType:   deviceType,
			Csrf:         sessionCsrf,
			PushContents: pushContents,
		})
	}

	if err := sessionRows.Err(); err != nil {
		sessionRows.Close()
		log.Printf("ContentsPush: iterate sessions: %v", err)
		common.WriteResponseWithSession(
			w, session, err.Error()+";iterate recipient sessions", http.StatusOK,
		)
		return
	}
	sessionRows.Close()

	// 4. Save only one attachment.
	// If imgPath is provided, use it; otherwise save uploaded file links.
	imgPath, err := common.ImgSave(
		r.FormValue("imgPath"),
		session.UserID,
		updatedBy,
		channelID,
		0,
		0,
	)
	if err != nil {
		log.Printf("ContentsPush: ImgSave: %v; channelID=%s", err, channelID)
		common.WriteResponseWithSession(
			w, session, err.Error()+";imgPath", http.StatusOK,
		)
		return
	}

	var attachment interface{}
	if imgPath != "" {
		attachment = imgPath
	} else {
		fileLinks, err := common.FileSave(r, channelID, updatedBy, userIDs, 0)
		if err != nil {
			log.Printf("ContentsPush: FileSave: %v; channelID=%s", err, channelID)
			common.WriteResponseWithSession(
				w, session, err.Error()+";FileSave", http.StatusOK,
			)
			return
		}
		attachment = fileLinks
	}

	// 5. Construct the push payload.
	arr := []interface{}{
		pushTitle,
		channelID,
		updatedBy,
		contents,
		attachment,
	}

	// Preserve the existing service-worker notification behavior.
	sw := map[string]interface{}{
		"url":     notifyURL,
		"message": notifyMessage,
		"icon":    notifyIcon,
	}
	if notifyURL != "" && notifyMessage != "" {
		arr = append(arr, sw)
	}

	// 6. Send push notifications.
	// ChunkPush handles/logs individual push errors internally.
	common.ChunkPush(sessions, arr)

	// 7. Return the current session's response data.
	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		log.Printf("ContentsPush: encode response: %v", err)
	}
}