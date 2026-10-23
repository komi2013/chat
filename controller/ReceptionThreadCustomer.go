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

	"chat/common"
	"chat/table"
)

func ReceptionThreadCustomer(w http.ResponseWriter, r *http.Request) {
	csrf := r.FormValue("csrf")
	receptionID, err := strconv.ParseInt(r.FormValue("receptionID"), 10, 64)
	if err != nil || receptionID <= 0 {
		common.WriteResponseWithoutSession(w, csrf, "Invalid receptionID format", http.StatusOK)
		return
	}
	messageID := r.FormValue("messageID")
	parentID := r.FormValue("parentID")
	inquiryName := r.FormValue("inquiryName")
	inquiryImg := r.FormValue("inquiryImg")
	messageTxt := common.SanitizeRichText(r.FormValue("messageTxt"))

	threadHeadNew := false
	var threadHead map[string]interface{}
	if raw := r.FormValue("threadHead"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &threadHead); err != nil {
			common.WriteResponseWithoutSession(w, csrf, "threadHead Invalid JSON", http.StatusOK)
			return
		}
		threadHeadNew = true
		if threadHead == nil {
			threadHead = make(map[string]interface{})
		}
		if value, ok := threadHead["messageTxt"].(string); ok {
			threadHead["messageTxt"] = common.SanitizeRichText(value)
		}
		if value, ok := threadHead["description"].(string); ok {
			threadHead["description"] = common.SanitizeRichText(value)
		}
	}

	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, "。サインインし直してください", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// receptionID is the request identifier. channelID and inquiry_names are
	// always read from the PostgreSQL reception master row, not from POST data.
	var channelID int64
	var adminNames, joinNames, inquiryNames []string
	err = common.DB.ReceptionDB.QueryRowContext(ctx, `
		SELECT channel_id, admin_names, join_names, inquiry_names
		FROM reception
		WHERE reception_id = $1
	`, receptionID).Scan(
		&channelID,
		pq.Array(&adminNames),
		pq.Array(&joinNames),
		pq.Array(&inquiryNames),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "Failed to find reception", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to get reception data: "+err.Error(), http.StatusOK)
		}
		return
	}

	aliasNames := make([]string, 0, len(joinNames))
	if existing, ok := threadHead["aliasNames"].([]interface{}); ok {
		for _, value := range existing {
			if name, ok := value.(string); ok {
				aliasNames = append(aliasNames, name)
			}
		}
	}
	aliasNames = append(aliasNames, joinNames...)
	threadHead["aliasNames"] = aliasNames
	threadHead["adminNames"] = adminNames
	threadHead["aliasImg"] = inquiryImg

	threadHeadArray := []interface{}{"threadHead", channelID, inquiryName, threadHead}

	// Access-right 3 represents an inquirer in the PostgreSQL schema.
	// Avoid creating duplicate aliases for the same user/channel/name tuple.
	_, err = common.DB.AliasDB.ExecContext(ctx, `
		INSERT INTO alias (channel_id, alias_name, alias_img, user_id, alias_bio, access_right)
		SELECT $1, $2, '', $3, '', 3
		WHERE NOT EXISTS (
			SELECT 1 FROM alias
			WHERE channel_id = $1 AND user_id = $3 AND alias_name = $2
		)
	`, channelID, inquiryName, session.UserID)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("alias登録失敗: %v", err), http.StatusInternalServerError)
		return
	}

	// Build the notification recipient set from reception.inquiry_names and
	// aliases belonging to the reception's channel.
	userIDs := make([]int64, 0)
	seenUserIDs := make(map[int64]struct{})
	if len(inquiryNames) > 0 {
		rows, queryErr := common.DB.AliasDB.QueryContext(ctx, `
			SELECT DISTINCT user_id
			FROM alias
			WHERE channel_id = $1
			  AND alias_name = ANY($2)
			  AND user_id IS NOT NULL
			ORDER BY user_id
		`, channelID, pq.Array(inquiryNames))
		if queryErr != nil {
			common.WriteResponseWithSession(w, session, queryErr.Error(), http.StatusOK)
			return
		}
		for rows.Next() {
			var userID int64
			if scanErr := rows.Scan(&userID); scanErr != nil {
				rows.Close()
				common.WriteResponseWithSession(w, session, scanErr.Error(), http.StatusOK)
				return
			}
			if _, exists := seenUserIDs[userID]; !exists {
				seenUserIDs[userID] = struct{}{}
				userIDs = append(userIDs, userID)
			}
		}
		if rowsErr := rows.Err(); rowsErr != nil {
			rows.Close()
			common.WriteResponseWithSession(w, session, rowsErr.Error(), http.StatusOK)
			return
		}
		rows.Close()
	}
	if _, exists := seenUserIDs[session.UserID]; !exists {
		userIDs = append(userIDs, session.UserID)
	}


	threadData := []interface{}{
		parentID, messageID, messageTxt, inquiryImg,
		[]string{inquiryName}, "", []interface{}{}, inquiryName,
	}
	pushArray := []interface{}{"thread", channelID, inquiryName, threadData, []string{}}

	filteredSessions := make([]table.SessionStruct, 0)
	if len(userIDs) > 0 {
		rows, queryErr := common.DB.SessionDB.QueryContext(ctx, `
			SELECT session_id, device_id, user_id, created_at, updated_at,
			       push_token, device_type, csrf, push_contents, mail, telephone
			FROM session
			WHERE user_id = ANY($1)
		`, pq.Array(userIDs))
		if queryErr != nil {
			common.WriteResponseWithSession(w, session, queryErr.Error(), http.StatusOK)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var item table.SessionStruct
			if scanErr := rows.Scan(
				&item.SessionID, &item.DeviceID, &item.UserID, &item.CreatedAt,
				&item.UpdatedAt, &item.PushToken, &item.DeviceType, &item.Csrf,
				pq.Array(&item.PushContents), &item.Mail, &item.Telephone,
			); scanErr != nil {
				common.WriteResponseWithSession(w, session, scanErr.Error(), http.StatusOK)
				return
			}
			filteredSessions = append(filteredSessions, item)
		}
		if rowsErr := rows.Err(); rowsErr != nil {
			common.WriteResponseWithSession(w, session, rowsErr.Error(), http.StatusOK)
			return
		}
	}

	if threadHeadNew {
		common.ChunkPush(filteredSessions, threadHeadArray)
	} else {
		common.ChunkPush(filteredSessions, pushArray)
	}

	responseData := common.BaseResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}
