package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chat/common"
)

func ReceptionQueueEdit(w http.ResponseWriter, r *http.Request) {
	aliasName := r.FormValue("aliasName")
	channelIDStr := r.FormValue("channelID")
	receptionID, err := strconv.ParseInt(r.FormValue("receptionID"), 10, 64)
	if err != nil || receptionID <= 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid receptionID", http.StatusOK)
		return
	}

	editType := r.FormValue("editType") // "1"=追加, "2"=削除
	if editType != "1" && editType != "2" {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid editType", http.StatusOK)
		return
	}

	var guestCount int
	if editType == "1" {
		guestCount, err = strconv.Atoi(r.FormValue("guestCount"))
		if err != nil || guestCount < 0 {
			common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid guestCount", http.StatusOK)
			return
		}
	}

	code := r.FormValue("code")
	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Staff validation: the submitted user, alias, and channel must match a row in alias.
	staffAccess := false
	var submittedChannelID int64
	channelIDValid := false
	if channelIDStr != "" {
		submittedChannelID, err = strconv.ParseInt(channelIDStr, 10, 64)
		channelIDValid = err == nil && submittedChannelID > 0
	}
	if aliasName != "" && channelIDValid {
		err = common.DB.AliasDB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM alias
				WHERE user_id = $1
				  AND alias_name = $2
				  AND channel_id = $3
				  AND COALESCE(access_right, 0) <> 3
			)
		`, session.UserID, aliasName, submittedChannelID).Scan(&staffAccess)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to validate alias: "+err.Error(), http.StatusOK)
			return
		}
	}

	// Queue deletion is staff-only.
	if editType == "2" && !staffAccess {
		common.WriteResponseWithSession(w, session, "スタッフのみ削除できます", http.StatusOK)
		return
	}
	if !staffAccess && code == "" {
		common.WriteResponseWithSession(w, session, "コードがない", http.StatusOK)
		return
	}

	// Confirm the reception exists and that staff are acting on the channel they submitted.
	var receptionChannelID int64
	err = common.DB.ReceptionDB.QueryRowContext(ctx, `
		SELECT channel_id
		FROM reception
		WHERE reception_id = $1
	`, receptionID).Scan(&receptionChannelID)
	if err != nil {
		if err == sql.ErrNoRows {
			common.WriteResponseWithSession(w, session, "reception not found", http.StatusOK)
		} else {
			common.WriteResponseWithSession(w, session, "Failed to get reception: "+err.Error(), http.StatusOK)
		}
		return
	}
	if staffAccess && submittedChannelID != receptionChannelID {
		common.WriteResponseWithSession(w, session, "channelID does not match reception", http.StatusOK)
		return
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("JST", 9*60*60)
	}
	now := time.Now().In(jst)

	tx, err := common.DB.ReceptionDB.BeginTx(ctx, nil)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to begin transaction: "+err.Error(), http.StatusOK)
		return
	}
	defer tx.Rollback()

	// Non-staff users must provide an active reception passcode.
	// Expired matching codes are removed; malformed dates are treated as invalid,
	// but are not deleted automatically.
	if !staffAccess {
		rows, queryErr := tx.QueryContext(ctx, `
			SELECT id, COALESCE(pass_start, ''), COALESCE(pass_end, '')
			FROM reception_passcode
			WHERE reception_id = $1
			  AND passkey = $2
		`, receptionID, code)
		if queryErr != nil {
			common.WriteResponseWithSession(w, session, "Failed to query passcodes: "+queryErr.Error(), http.StatusOK)
			return
		}

		validCode := false
		var expiredPasscodeIDs []int64
		for rows.Next() {
			var passcodeID int64
			var passStart, passEnd string
			if scanErr := rows.Scan(&passcodeID, &passStart, &passEnd); scanErr != nil {
				rows.Close()
				common.WriteResponseWithSession(w, session, "Failed to read passcodes: "+scanErr.Error(), http.StatusOK)
				return
			}

			startTime, startErr := time.ParseInLocation("2006-01-02T15:04", passStart, jst)
			endTime, endErr := time.ParseInLocation("2006-01-02T15:04", passEnd, jst)
			if startErr == nil && endErr == nil {
				if now.After(startTime) && now.Before(endTime) {
					validCode = true
				} else if now.After(endTime) {
					expiredPasscodeIDs = append(expiredPasscodeIDs, passcodeID)
				}
			}
		}
		rowsErr := rows.Err()
		_ = rows.Close()
		if rowsErr != nil {
			common.WriteResponseWithSession(w, session, "Failed to iterate passcodes: "+rowsErr.Error(), http.StatusOK)
			return
		}

		for _, passcodeID := range expiredPasscodeIDs {
			if _, err = tx.ExecContext(ctx, `
				DELETE FROM reception_passcode
				WHERE id = $1 AND reception_id = $2
			`, passcodeID, receptionID); err != nil {
				common.WriteResponseWithSession(w, session, "Failed to remove expired passcode: "+err.Error(), http.StatusOK)
				return
			}
		}

		if !validCode {
			// Preserve expired-code cleanup even though this request is denied.
			if len(expiredPasscodeIDs) > 0 {
				if err = tx.Commit(); err != nil {
					common.WriteResponseWithSession(w, session, "Failed to commit passcode cleanup: "+err.Error(), http.StatusOK)
					return
				}
			}
			common.WriteResponseWithSession(w, session, "コードが一致していません", http.StatusOK)
			return
		}
	}

	var queueID int64
	switch editType {
	case "1": // Add a queue entry.
		// queue_name replaces the former session nickname with the numeric user ID string.
		queueName := strconv.FormatInt(session.UserID, 10)
		queuedAt := now.Format("2006-01-02T15:04")
		err = tx.QueryRowContext(ctx, `
			INSERT INTO reception_queue (
				reception_id,
				sort_order,
				waiting_guest,
				queue_name,
				queued_at,
				user_id
			)
			VALUES (
				$1,
				COALESCE((SELECT MAX(sort_order) + 1 FROM reception_queue WHERE reception_id = $1), 0),
				$2, $3, $4, $5
			)
			RETURNING queue_id
		`, receptionID, guestCount, queueName, queuedAt, session.UserID).Scan(&queueID)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to add queue: "+err.Error(), http.StatusOK)
			return
		}

	case "2": // Delete one queue entry by queue_id. Staff only.
		queueID, err = strconv.ParseInt(r.FormValue("queueID"), 10, 64)
		if err != nil || queueID <= 0 {
			common.WriteResponseWithSession(w, session, "Invalid queueID", http.StatusOK)
			return
		}

		result, execErr := tx.ExecContext(ctx, `
			DELETE FROM reception_queue
			WHERE queue_id = $1 AND reception_id = $2
		`, queueID, receptionID)
		if execErr != nil {
			common.WriteResponseWithSession(w, session, "Failed to delete queue: "+execErr.Error(), http.StatusOK)
			return
		}
		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			common.WriteResponseWithSession(w, session, "Failed to verify deleted queue: "+rowsErr.Error(), http.StatusOK)
			return
		}
		if rowsAffected == 0 {
			common.WriteResponseWithSession(w, session, "queue not found", http.StatusOK)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to commit queue update: "+err.Error(), http.StatusOK)
		return
	}

	// queueID is returned on add and delete so callers can identify the affected entry.
	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
		QueueID      int64    `json:"queueID,omitempty"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		QueueID:      queueID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		// The database transaction has already committed; do not write a second response.
		return
	}
}
