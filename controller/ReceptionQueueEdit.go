package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chat/common"
	table "chat/table"
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

	if editType == "2" && !staffAccess {
		common.WriteResponseWithSession(w, session, "スタッフのみ削除できます", http.StatusOK)
		return
	}
	if !staffAccess && code == "" {
		common.WriteResponseWithSession(w, session, "コードがない", http.StatusOK)
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

	// Lock and load the reception row. Passcodes are stored in reception.passcodes JSONB.
	var receptionChannelID int64
	var passcodesJSON []byte
	err = tx.QueryRowContext(ctx, `
		SELECT channel_id, passcodes
		FROM reception
		WHERE reception_id = $1
		FOR UPDATE
	`, receptionID).Scan(&receptionChannelID, &passcodesJSON)
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

	if !staffAccess {
		var passcodes []table.Passcode
		if err := json.Unmarshal(passcodesJSON, &passcodes); err != nil {
			common.WriteResponseWithSession(w, session, "Failed to decode passcodes: "+err.Error(), http.StatusOK)
			return
		}

		validCode := false
		expiredPasscodes := make([]table.Passcode, 0)
		remainingPasscodes := make([]table.Passcode, 0, len(passcodes))
		for _, passcode := range passcodes {
			if passcode.Passkey != code {
				remainingPasscodes = append(remainingPasscodes, passcode)
				continue
			}

			startTime, startErr := time.ParseInLocation("2006-01-02T15:04", passcode.PassStart, jst)
			endTime, endErr := time.ParseInLocation("2006-01-02T15:04", passcode.PassEnd, jst)
			if startErr == nil && endErr == nil {
				if now.After(startTime) && now.Before(endTime) {
					validCode = true
					remainingPasscodes = append(remainingPasscodes, passcode)
				} else if now.After(endTime) {
					expiredPasscodes = append(expiredPasscodes, passcode)
				} else {
					remainingPasscodes = append(remainingPasscodes, passcode)
				}
			} else {
				// Malformed/missing dates remain stored but do not authorize the request.
				remainingPasscodes = append(remainingPasscodes, passcode)
			}
		}

		if len(expiredPasscodes) > 0 {
			updatedJSON, marshalErr := json.Marshal(remainingPasscodes)
			if marshalErr != nil {
				common.WriteResponseWithSession(w, session, "Failed to encode passcodes: "+marshalErr.Error(), http.StatusOK)
				return
			}
			if _, err = tx.ExecContext(ctx, `
				UPDATE reception
				SET passcodes = $1::JSONB, updated_at = CURRENT_TIMESTAMP
				WHERE reception_id = $2
			`, string(updatedJSON), receptionID); err != nil {
				common.WriteResponseWithSession(w, session, "Failed to remove expired passcodes: "+err.Error(), http.StatusOK)
				return
			}
		}

		if !validCode {
			if len(expiredPasscodes) > 0 {
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
	case "1":
		queueName := strconv.FormatInt(session.UserID, 10)
		queuedAt := now.Format("2006-01-02T15:04")
		err = tx.QueryRowContext(ctx, `
			INSERT INTO reception_queue (
				reception_id, sort_order, waiting_guest, queue_name, queued_at, user_id
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

	case "2":
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
		return
	}
}
