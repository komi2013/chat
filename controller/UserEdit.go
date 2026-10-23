package controller

import (
	"context"
	// "database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"chat/common"
)

func UserEdit(w http.ResponseWriter, r *http.Request) {
	lat, err := strconv.ParseFloat(r.FormValue("latitude"), 64)
	if err != nil || lat < -90 || lat > 90 {
		common.WriteResponseWithoutSession(
			w, r.FormValue("csrf"), "緯度 (latitude) が不正です", http.StatusOK,
		)
		return
	}

	lng, err := strconv.ParseFloat(r.FormValue("longitude"), 64)
	if err != nil || lng < -180 || lng > 180 {
		common.WriteResponseWithoutSession(
			w, r.FormValue("csrf"), "経度 (longitude) が不正です", http.StatusOK,
		)
		return
	}

	mail := r.FormValue("mail")
	telephone := r.FormValue("telephone")
	walletAddress := r.FormValue("walletAddress")

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(
			w, r.FormValue("csrf"), err.Error(), http.StatusOK,
		)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// UserDB 更新
	tx, err := common.DB.UserDB.BeginTx(ctx, nil)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE "user"
		SET latitude = $1,
		    longitude = $2,
		    mail = $3,
		    telephone = $4,
		    wallet_address = $5
		WHERE user_id = $6
	`, lat, lng, mail, telephone, walletAddress, session.UserID)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}
	if rowsAffected == 0 {
		common.WriteResponseWithSession(
			w, session, "ユーザー情報が見つかりません", http.StatusOK,
		)
		return
	}

	if err := tx.Commit(); err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	// SessionDB の全セッションを更新
	_, err = common.DB.SessionDB.ExecContext(ctx, `
		UPDATE "session"
		SET mail = $1,
		    telephone = $2,
		    updated_at = $3
		WHERE user_id = $4
	`, mail, telephone, time.Now(), session.UserID)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	// 念のため、更新処理で未使用の sql.ErrNoRows などを
	// この関数で参照する必要はありません。

	sessionID := ""
	if session.SessionIDRotated {
		sessionID = session.SessionID
	}

	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
		SessionID    string   `json:"sessionId,omitempty"`
		Message      string   `json:"message"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		SessionID:    sessionID,
		Message:      "ユーザー情報は更新されました",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		log.Printf("UserEdit: response encode failed: %v", err)
	}
}