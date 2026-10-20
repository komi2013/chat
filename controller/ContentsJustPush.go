package controller

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"chat/collection"
	"chat/common"

	"github.com/lib/pq"
)

func ContentsJustPush(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("ContentsJustPush: ParseMultipartForm: %v; path=%s", err, r.URL.Path)
		http.Error(w, "Failed to parse form because more than 10MB", http.StatusBadRequest)
		return
	}

	var pushNames []string
	if err := json.Unmarshal([]byte(r.FormValue("pushNames")), &pushNames); err != nil {
		log.Printf("ContentsJustPush: pushNames: %v; path=%s", err, r.URL.Path)
		http.Error(w, "Invalid JSON pushNames", http.StatusBadRequest)
		return
	}

	updatedBy := r.FormValue("updatedBy")
	channelIDText := r.FormValue("channelID")
	pushTitle := r.FormValue("pushTitle")

	var contents interface{}
	if err := json.Unmarshal([]byte(r.FormValue("contents")), &contents); err != nil {
		log.Printf("ContentsJustPush: contents: %v; path=%s", err, r.URL.Path)
		http.Error(w, "Invalid JSON contents", http.StatusBadRequest)
		return
	}

	channelID, err := strconv.ParseInt(channelIDText, 10, 64)
	if err != nil || channelID <= 0 {
		http.Error(w, "Invalid channelID", http.StatusBadRequest)
		return
	}

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(
			w,
			r.FormValue("csrf"),
			err.Error()+";SessionCheckTake",
			http.StatusOK,
		)
		return
	}

	// 1. Verify that updatedBy is an alias belonging to this user
	// in this channel. Inquirers (access_right = 3) cannot post.
	//
	// Access rights 0 (normal), 1 (admin), and 2 (guest) are allowed.
	var hasAccess bool
	err = common.DB.AliasDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM alias
			WHERE channel_id = $1
			  AND alias_name = $2
			  AND user_id = $3
			  AND COALESCE(access_right, 0) <> 3
		)
	`, channelID, updatedBy, session.UserID).Scan(&hasAccess)
	if err != nil {
		log.Printf("ContentsJustPush: check alias access: %v; channelID=%d; updatedBy=%s",
			err, channelID, updatedBy)
		http.Error(w, "Failed to check alias access", http.StatusInternalServerError)
		return
	}

	if !hasAccess {
		common.WriteResponseWithSession(w, session, "no true access right", http.StatusOK)
		return
	}

	// 2. Find users belonging to the selected aliases in this channel.
	// Inquirers are excluded from recipients. DISTINCT avoids sending
	// duplicate notifications when one user owns multiple selected aliases.
	rows, err := common.DB.AliasDB.Query(`
		SELECT DISTINCT user_id
		FROM alias
		WHERE channel_id = $1
		  AND alias_name = ANY($2)
		  AND user_id IS NOT NULL
		  AND COALESCE(access_right, 0) <> 3
	`, channelID, pq.Array(pushNames))
	if err != nil {
		log.Printf("ContentsJustPush: query recipient aliases: %v; channelID=%d",
			err, channelID)
		http.Error(w, "Failed to find recipients", http.StatusInternalServerError)
		return
	}

	userIDs := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			rows.Close()
			log.Printf("ContentsJustPush: scan recipient user ID: %v", err)
			http.Error(w, "Failed to find recipients", http.StatusInternalServerError)
			return
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		log.Printf("ContentsJustPush: iterate recipient aliases: %v", err)
		http.Error(w, "Failed to find recipients", http.StatusInternalServerError)
		return
	}
	rows.Close()

	// Do not save attachments or send anything when no aliases matched.
	if len(userIDs) == 0 {
		common.WriteResponseWithSession(
			w,
			session,
			"該当するネームは存在しません",
			http.StatusOK,
		)
		return
	}

	// 3. Save the image path, if supplied.
	imgPath, err := common.ImgSave(
		r.FormValue("imgPath"),
		session.UserID,
		updatedBy,
		channelIDText,
		3,
		0,
	)
	if err != nil {
		log.Printf("ContentsJustPush: ImgSave: %v; channelID=%d", err, channelID)
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	// 4. Only one attachment is used:
	// use imgPath when present; otherwise save uploaded file attachments.
	var fileLinks interface{}
	if imgPath == "" {
		savedLinks, err := common.FileSave(r, channelIDText, updatedBy, userIDs, 0)
		if err != nil {
			log.Printf("ContentsJustPush: FileSave: %v; channelID=%d", err, channelID)
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		fileLinks = savedLinks
	}

	// 5. Load all sessions belonging to recipient users.
	// Session records no longer contain channel aliases, so there is no
	// FilterSessionsByChannelID call here.
	sessionRows, err := common.DB.SessionDB.Query(`
		SELECT
			session_id,
			user_id,
			COALESCE(push_token, ''),
			COALESCE(device_type, 0),
			COALESCE(csrf, ''),
			COALESCE(push_contents, ARRAY[]::TEXT[])
		FROM "session"
		WHERE user_id = ANY($1)
	`, pq.Array(userIDs))
	if err != nil {
		log.Printf("ContentsJustPush: query sessions: %v; userIDs=%v", err, userIDs)
		http.Error(w, "Failed to find recipient sessions", http.StatusInternalServerError)
		return
	}

	sessions := make([]collection.SessionStruct, 0)
	for sessionRows.Next() {
		var (
			sessionID   string
			userID      int64
			pushToken   string
			deviceType  int
			csrf        string
			pushContents []string
		)

		if err := sessionRows.Scan(
			&sessionID,
			&userID,
			&pushToken,
			&deviceType,
			&csrf,
			pq.Array(&pushContents),
		); err != nil {
			sessionRows.Close()
			log.Printf("ContentsJustPush: scan session: %v", err)
			http.Error(w, "Failed to load recipient sessions", http.StatusInternalServerError)
			return
		}

		sessions = append(sessions, collection.SessionStruct{
			SessionID:    sessionID,
			UserID:       int(userID),
			PushToken:    pushToken,
			DeviceType:   deviceType,
			Csrf:         csrf,
			PushContents: pushContents,
		})
	}

	if err := sessionRows.Err(); err != nil {
		sessionRows.Close()
		log.Printf("ContentsJustPush: iterate sessions: %v", err)
		http.Error(w, "Failed to load recipient sessions", http.StatusInternalServerError)
		return
	}
	sessionRows.Close()

	// 6. Build the push payload. Preserve the existing payload structure:
	// [pushTitle, channelID, updatedBy, contents, attachment]
	var attachment interface{}
	if imgPath != "" {
		attachment = imgPath
	} else {
		attachment = fileLinks
	}

	arr := []interface{}{
		pushTitle,
		channelIDText,
		updatedBy,
		contents,
		attachment,
	}

	// 7. Deliver notifications. ChunkJustPush handles/logs delivery failures.
	common.ChunkJustPush(sessions, arr)

	// 8. Refresh the response data.
	session, err = common.PushReGenerate(session)
	if err != nil {
		log.Printf("ContentsJustPush: PushReGenerate: %v; path=%s", err, r.URL.Path)
		common.WriteResponseWithSession(
			w,
			session,
			"Failed to regenerate push data",
			http.StatusInternalServerError,
		)
		return
	}

	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		log.Printf("ContentsJustPush: encode response: %v", err)
	}
}