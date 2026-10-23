package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"chat/common"
)

func SignInGoogleMobile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userDB := common.DB.UserDB
	sessionDB := common.DB.SessionDB

	// Mobile Csrf form value.
	formCsrf := r.FormValue("csrf")
	if formCsrf == "" {
		http.Error(w, "missing csrf", http.StatusBadRequest)
		return
	}

	// Device identifier is used only for session counting.
	// It is not an authentication credential and does not cause
	// any other sessions to be deleted.
	deviceID := strings.TrimSpace(r.FormValue("deviceID"))
	if len(deviceID) > 255 {
		http.Error(w, "invalid deviceID", http.StatusBadRequest)
		return
	}

	credential := r.FormValue("credential")
	if credential == "" {
		http.Error(w, "missing credential", http.StatusBadRequest)
		return
	}

	// Validate using the MOBILE Google OAuth client ID.
	expectedClientID := strings.TrimSpace(
		os.Getenv("GOOGLE_CLIENT_MOBILE_ID"),
	)
	if expectedClientID == "" {
		log.Printf("GOOGLE_CLIENT_MOBILE_ID is not configured")
		http.Error(
			w,
			"Google mobile OAuth is not configured",
			http.StatusInternalServerError,
		)
		return
	}

	// Verify Google ID token signature.
	token, err := jwt.Parse(
		credential,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method: %v",
					token.Header["alg"],
				)
			}

			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, fmt.Errorf("missing key ID")
			}

			return getGooglePublicKey(kid)
		},
	)
	if err != nil || token == nil || !token.Valid {
		log.Printf("Google mobile token validation failed: %v", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	// Require expiration.
	if _, exists := claims["exp"]; !exists {
		http.Error(
			w,
			"missing token expiration",
			http.StatusUnauthorized,
		)
		return
	}

	// Validate Google issuer.
	issuer, _ := claims["iss"].(string)
	if issuer != "https://accounts.google.com" &&
		issuer != "accounts.google.com" {
		http.Error(w, "invalid token issuer", http.StatusUnauthorized)
		return
	}

	// Validate the mobile OAuth audience.
	if !googleMobileAudienceMatches(claims, expectedClientID) {
		http.Error(
			w,
			"invalid Google mobile client ID",
			http.StatusUnauthorized,
		)
		return
	}

	googleSub, ok := claims["sub"].(string)
	if !ok || googleSub == "" {
		http.Error(
			w,
			"missing Google subject",
			http.StatusUnauthorized,
		)
		return
	}

	now := time.Now()

	// Create a user or update the existing user's signed_at.
	// PostgreSQL generates user_id through its identity column.
	var userID int64
	var mail, telephone string

	err = userDB.QueryRowContext(ctx, `
		INSERT INTO "user" (
			google_jwt_sub,
			signed_at
		)
		VALUES ($1, $2)
		ON CONFLICT (google_jwt_sub)
		DO UPDATE SET
			signed_at = EXCLUDED.signed_at
		RETURNING
			user_id,
			COALESCE(mail, ''),
			COALESCE(telephone, '')
	`,
		googleSub,
		now,
	).Scan(&userID, &mail, &telephone)
	if err != nil {
		log.Printf("Google mobile user upsert failed: %v", err)
		http.Error(w, "user update fail", http.StatusInternalServerError)
		return
	}

	// Create a new server-generated session ID and Csrf token.
	sessionID, err := randomToken()
	if err != nil {
		log.Printf("session ID generation failed: %v", err)
		http.Error(
			w,
			"cannot create session",
			http.StatusInternalServerError,
		)
		return
	}

	csrf, err := randomToken()
	if err != nil {
		log.Printf("session Csrf generation failed: %v", err)
		http.Error(
			w,
			"cannot create session",
			http.StatusInternalServerError,
		)
		return
	}

	// Native app login.
	// PushSubscribeMobile registers the FCM token separately after login.
	// Do not copy a push token from another session or device.
	//
	// device_id is stored only as a session-counting identifier.
	// An empty deviceID is stored as NULL.
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
			ARRAY[]::TEXT[],
			$6,
			$7
		)
	`,
		sessionID,
		userID,
		deviceID,
		now,
		csrf,
		mail,
		telephone,
	)
	if err != nil {
		log.Printf("Google mobile session insert failed: %v", err)
		http.Error(
			w,
			"cannot create session",
			http.StatusInternalServerError,
		)
		return
	}

	// Keep cookie authentication for clients that still use it.
	// This cookie is separate from deviceID.
	http.SetCookie(w, &http.Cookie{
		Name:     "ss",
		Value:    sessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Remove sessions older than one month.
	// Cleanup errors are logged but do not invalidate the new session.
	_, err = sessionDB.ExecContext(ctx, `
		DELETE FROM "session"
		WHERE user_id = $1
		  AND updated_at < $2
	`,
		userID,
		now.AddDate(0, -1, 0),
	)
	if err != nil {
		log.Printf("expired session cleanup failed: %v", err)
	}

	// Keep at most five sessions per user.
	// Always retain the newly created session.
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
		log.Printf("session retention cleanup failed: %v", err)
	}

	log.Printf(
		"Google mobile session created: user_id=%d",
		userID,
	)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Session-Token", sessionID)

	// Keep userId as a string and nickname as an empty string for
	// compatibility with clients expecting the previous JSON fields.
	// The database user_id itself is numeric.
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
		log.Printf("Google mobile response encoding failed: %v", err)
	}
}

// Validate the mobile Google OAuth audience.
// Google JWT aud can be either a string or an array.
func googleMobileAudienceMatches(
	claims jwt.MapClaims,
	expectedClientID string,
) bool {
	if expectedClientID == "" {
		return false
	}

	switch aud := claims["aud"].(type) {
	case string:
		return aud == expectedClientID

	case []interface{}:
		for _, item := range aud {
			if value, ok := item.(string); ok &&
				value == expectedClientID {
				return true
			}
		}

	case []string:
		for _, value := range aud {
			if value == expectedClientID {
				return true
			}
		}
	}

	return false
}