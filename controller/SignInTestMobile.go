package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"

	"chat/common"
)

func SignInTestMobile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userDB := common.DB.UserDB
	sessionDB := common.DB.SessionDB

	// Test-only authentication.
	// Do not expose this endpoint publicly.
	formCsrf := r.FormValue("csrf")
	if formCsrf == "" {
		http.Error(w, "missing csrf", http.StatusBadRequest)
		return
	}

	// Device ID is only for counting sessions by device.
	// It is not an authentication credential.
	deviceID := strings.TrimSpace(r.FormValue("deviceID"))
	if len(deviceID) > 255 {
		http.Error(w, "invalid deviceID", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	if username == "" {
		http.Error(w, "missing username", http.StatusBadRequest)
		return
	}

	now := time.Now()

	// Create or update the test user.
	// PostgreSQL generates user_id through its identity column.
	var userID int64
	var mail, telephone string

	err := userDB.QueryRowContext(ctx, `
		INSERT INTO "user" (
			test_username,
			signed_at
		)
		VALUES ($1, $2)
		ON CONFLICT (test_username)
		DO UPDATE SET
			signed_at = EXCLUDED.signed_at
		RETURNING
			user_id,
			COALESCE(mail, ''),
			COALESCE(telephone, '')
	`,
		username,
		now,
	).Scan(&userID, &mail, &telephone)

	if err != nil {
		log.Printf("test user upsert failed: %v", err)
		http.Error(
			w,
			"user update fail",
			http.StatusInternalServerError,
		)
		return
	}

	// Generate a new session ID and CSRF token.
	sessionID, err := randomToken()
	if err != nil {
		log.Printf("test session ID generation failed: %v", err)
		http.Error(
			w,
			"cannot create session",
			http.StatusInternalServerError,
		)
		return
	}

	csrf, err := randomToken()
	if err != nil {
		log.Printf("test CSRF generation failed: %v", err)
		http.Error(
			w,
			"cannot create session",
			http.StatusInternalServerError,
		)
		return
	}

	// Create the session.
	// device_id is stored only for session counting.
	// The mobile app can register its push token later via
	// PushSubscribeMobile.
	_, err = sessionDB.ExecContext(ctx, `
		INSERT INTO "session" (
			session_id,
			user_id,
			device_id,
			created_at,
			updated_at,
			push_token,
			device_type,
			csrf,
			push_contents,
			mail,
			telephone
		)
		VALUES (
			$1,
			$2,
			NULLIF($3::VARCHAR, ''),
			$4,
			$4,
			NULL,
			2,
			$5,
			$6,
			$7,
			$8
		)
	`,
		sessionID,
		userID,
		deviceID,
		now,
		csrf,
		pq.Array([]string{}),
		mail,
		telephone,
	)
	if err != nil {
		log.Printf("test session insert failed: %v", err)
		http.Error(
			w,
			"cannot create session",
			http.StatusInternalServerError,
		)
		return
	}

	// Set the authentication cookie after the session is inserted.
	http.SetCookie(w, &http.Cookie{
		Name:     "ss",
		Value:    sessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	log.Printf("test session created: user_id=%d", userID)

	// Delete sessions older than one month.
	// Cleanup failure does not invalidate the newly created session.
	_, err = sessionDB.ExecContext(ctx, `
		DELETE FROM "session"
		WHERE user_id = $1
		  AND updated_at < $2
	`,
		userID,
		now.AddDate(0, -1, 0),
	)
	if err != nil {
		log.Printf("expired test session cleanup failed: %v", err)
	}

	// Keep at most five sessions per user, including the newly created one.
	// This limit is per user, not per device.
	_, err = sessionDB.ExecContext(ctx, `
		DELETE FROM "session"
		WHERE user_id = $1
		  AND session_id NOT IN (
			  SELECT session_id
			  FROM "session"
			  WHERE user_id = $1
			  ORDER BY
				  (session_id = $2) DESC,
				  updated_at DESC NULLS LAST,
				  created_at DESC NULLS LAST
			  LIMIT 5
		  )
	`,
		userID,
		sessionID,
	)
	if err != nil {
		log.Printf("test session retention cleanup failed: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Session-Token", sessionID)

	response := struct {
		Success   bool   `json:"success"`
		Csrf      string `json:"csrf"`
		SessionID string `json:"sessionId"`
		UserID    string `json:"userId"`
		Nickname  string `json:"nickname"`
		Message   string `json:"message"`
	}{
		Success:   true,
		Csrf:      csrf,
		SessionID: sessionID,
		UserID:    fmt.Sprintf("%d", userID),
		Nickname:  "",
		Message:   "Success",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("test login response encoding failed: %v", err)
	}
}