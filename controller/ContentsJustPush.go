package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"chat/common"
	"chat/table"

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
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	// updatedBy must be an alias belonging to the current user.
	// Inquirers (access_right = 3) cannot post.
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
		log.Printf("ContentsJustPush: check alias access: %v; channelID=%d; updatedBy=%s", err, channelID, updatedBy)
		http.Error(w, "Failed to check alias access", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		common.WriteResponseWithSession(w, session, "no true access right", http.StatusOK)
		return
	}

	rows, err := common.DB.AliasDB.Query(`
		SELECT DISTINCT user_id
		FROM alias
		WHERE channel_id = $1
		  AND alias_name = ANY($2::TEXT[])
		  AND user_id IS NOT NULL
		  AND COALESCE(access_right, 0) <> 3
	`, channelID, pq.Array(pushNames))
	if err != nil {
		log.Printf("ContentsJustPush: query recipient aliases: %v; channelID=%d", err, channelID)
		http.Error(w, "Failed to find recipients", http.StatusInternalServerError)
		return
	}

	userIDs := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			_ = rows.Close()
			log.Printf("ContentsJustPush: scan recipient user ID: %v", err)
			http.Error(w, "Failed to find recipients", http.StatusInternalServerError)
			return
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		log.Printf("ContentsJustPush: iterate recipient aliases: %v", err)
		http.Error(w, "Failed to find recipients", http.StatusInternalServerError)
		return
	}
	_ = rows.Close()

	if len(userIDs) == 0 {
		common.WriteResponseWithSession(w, session, "該当するネームは存在しません", http.StatusOK)
		return
	}

	imgPath, err := common.ImgSave(
		r.FormValue("imgPath"),
		session.UserID,
		updatedBy,
		channelID,
		3,
		0,
	)
	if err != nil {
		log.Printf("ContentsJustPush: ImgSave: %v; channelID=%d", err, channelID)
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	var fileLinks interface{}
	if imgPath == "" {
		savedLinks, err := common.FileSave(r, channelID, session.UserID, userIDs, 0)
		if err != nil {
			log.Printf("ContentsJustPush: FileSave: %v; channelID=%d", err, channelID)
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		fileLinks = savedLinks
	}

	sessionRows, err := common.DB.SessionDB.Query(`
		SELECT session_id, user_id, COALESCE(push_token, ''),
		       COALESCE(device_type, 0), COALESCE(csrf, ''),
		       COALESCE(push_contents, ARRAY[]::TEXT[])
		FROM "session"
		WHERE user_id = ANY($1::BIGINT[])
	`, pq.Array(userIDs))
	if err != nil {
		log.Printf("ContentsJustPush: query sessions: %v; userIDs=%v", err, userIDs)
		http.Error(w, "Failed to find recipient sessions", http.StatusInternalServerError)
		return
	}

	sessions := make([]table.SessionStruct, 0)
	for sessionRows.Next() {
		var (
			sessionID    string
			userID       int64
			pushToken    string
			deviceType   int
			csrf         string
			pushContents []string
		)
		if err := sessionRows.Scan(&sessionID, &userID, &pushToken, &deviceType, &csrf, pq.Array(&pushContents)); err != nil {
			_ = sessionRows.Close()
			log.Printf("ContentsJustPush: scan session: %v", err)
			http.Error(w, "Failed to load recipient sessions", http.StatusInternalServerError)
			return
		}
		sessions = append(sessions, table.SessionStruct{
			SessionID: sessionID, UserID: userID, PushToken: pushToken,
			DeviceType: deviceType, Csrf: csrf, PushContents: pushContents,
		})
	}
	if err := sessionRows.Err(); err != nil {
		_ = sessionRows.Close()
		log.Printf("ContentsJustPush: iterate sessions: %v", err)
		http.Error(w, "Failed to load recipient sessions", http.StatusInternalServerError)
		return
	}
	_ = sessionRows.Close()

	var attachment interface{}
	if imgPath != "" {
		attachment = imgPath
	} else {
		attachment = fileLinks
	}
	arr := []interface{}{pushTitle, channelIDText, updatedBy, contents, attachment}
	common.ChunkJustPush(sessions, arr)

	session, err = common.PushReGenerate(session)
	if err != nil {
		log.Printf("ContentsJustPush: PushReGenerate: %v; path=%s", err, r.URL.Path)
		common.WriteResponseWithSession(w, session, "Failed to regenerate push data", http.StatusInternalServerError)
		return
	}

	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{Csrf: session.Csrf, PushContents: session.PushContents}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		log.Printf("ContentsJustPush: encode response: %v", err)
	}
}
