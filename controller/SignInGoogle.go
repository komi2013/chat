package controller

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lib/pq"

	"chat/common"
)

func SignInGoogle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userDB := common.DB.UserDB
	sessionDB := common.DB.SessionDB

	// Validate CSRF.
	cookie, err := r.Cookie("g_csrf_token")
	if err != nil || cookie.Value != r.FormValue("g_csrf_token") {
		http.Error(w, "invalid csrf token", http.StatusBadRequest)
		return
	}

	credential := r.FormValue("credential")
	if credential == "" {
		http.Error(w, "missing credential", http.StatusBadRequest)
		return
	}

	// Verify Google ID token.
	token, err := jwt.Parse(credential, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("missing key ID")
		}

		return getGooglePublicKey(kid)
	})
	if err != nil || token == nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	googleSub, ok := claims["sub"].(string)
	if !ok || googleSub == "" {
		http.Error(w, "missing Google subject", http.StatusUnauthorized)
		return
	}

	now := time.Now()

	// 0 = web, 1 = web push registered, 2 = native/API
	isAPI := strings.Contains(r.Header.Get("Accept"), "application/json") ||
		r.Header.Get("X-Requested-With") == "XMLHttpRequest"

	deviceType := 0
	if common.IsMobile(r.Header.Get("User-Agent")) || isAPI {
		deviceType = 2
	}

	// PostgreSQL generates user_id through its identity column.
	var userID int64
	var mail, telephone string

	err = userDB.QueryRowContext(ctx, `
		INSERT INTO "user" (google_jwt_sub, signed_at)
		VALUES ($1, $2)
		ON CONFLICT (google_jwt_sub)
		DO UPDATE SET signed_at = EXCLUDED.signed_at
		RETURNING user_id, COALESCE(mail, ''), COALESCE(telephone, '')
	`, googleSub, now).Scan(&userID, &mail, &telephone)
	if err != nil {
		log.Printf("user upsert failed: %v", err)
		http.Error(w, "user update fail", http.StatusInternalServerError)
		return
	}

	// Find the current browser's session, if any.
	currentSessionID := ""
	if c, err := r.Cookie("ss"); err == nil {
		currentSessionID = c.Value
	}

	type session struct {
		id           string
		deviceType   int
		pushContents []string
		pushToken    string
		createdAt    time.Time
		updatedAt    time.Time
	}

	var sessions []session
	var previous session
	sameBrowser := false

	rows, err := sessionDB.QueryContext(ctx, `
		SELECT
			session_id,
			COALESCE(device_type, 0),
			COALESCE(push_contents, ARRAY[]::TEXT[]),
			COALESCE(push_token, ''),
			COALESCE(created_at, NOW()),
			COALESCE(updated_at, NOW())
		FROM "session"
		WHERE user_id = $1
		ORDER BY updated_at ASC NULLS FIRST
	`, userID)
	if err != nil {
		log.Printf("session lookup failed: %v", err)
		http.Error(w, "session lookup fail", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var s session
		var pushContents pq.StringArray

		if err := rows.Scan(
			&s.id,
			&s.deviceType,
			&pushContents,
			&s.pushToken,
			&s.createdAt,
			&s.updatedAt,
		); err != nil {
			rows.Close()
			log.Printf("session scan failed: %v", err)
			http.Error(w, "session lookup fail", http.StatusInternalServerError)
			return
		}

		s.pushContents = []string(pushContents)

		if currentSessionID != "" && s.id == currentSessionID {
			previous = s
			sameBrowser = true
		}

		sessions = append(sessions, s)
	}

	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		log.Printf("session rows failed: %v", rowsErr)
		http.Error(w, "session lookup fail", http.StatusInternalServerError)
		return
	}

	// Generate a secure session ID and CSRF token.
	sessionID, err := randomToken()
	if err != nil {
		http.Error(w, "cannot create session", http.StatusInternalServerError)
		return
	}

	csrf, err := randomToken()
	if err != nil {
		http.Error(w, "cannot create session", http.StatusInternalServerError)
		return
	}

	// Keep push registration and display settings for the same browser.
	if previous.pushToken != "" {
		deviceType = previous.deviceType
	}

	pushContents := previous.pushContents
	if pushContents == nil {
		pushContents = []string{}
	}

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
		VALUES ($1, $2, $3, $4, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		sessionID,
		userID,
		"",
		now,
		previous.pushToken,
		deviceType,
		csrf,
		pq.Array(pushContents),
		mail,
		telephone,
	)
	if err != nil {
		log.Printf("session insert failed: %v", err)
		http.Error(w, "cannot create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "ss",
		Value:    sessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})

	// Same browser: replace the previous session and redirect.
	if sameBrowser && currentSessionID != "" {
		_, err := sessionDB.ExecContext(ctx, `
			DELETE FROM "session"
			WHERE session_id = $1 AND user_id = $2
		`, currentSessionID, userID)
		if err != nil {
			log.Printf("old session deletion failed: %v", err)
		}

		http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
		return
	}

	// Delete sessions older than one month.
	_, err = sessionDB.ExecContext(ctx, `
		DELETE FROM "session"
		WHERE user_id = $1 AND updated_at < $2
	`, userID, now.AddDate(0, -1, 0))
	if err != nil {
		log.Printf("expired session cleanup failed: %v", err)
	}

	// Keep at most the five most recently updated sessions.
	_, err = sessionDB.ExecContext(ctx, `
		DELETE FROM "session"
		WHERE user_id = $1
		  AND session_id NOT IN (
			  SELECT session_id
			  FROM "session"
			  WHERE user_id = $1
			  ORDER BY updated_at DESC NULLS LAST,
			           created_at DESC NULLS LAST
			  LIMIT 5
		  )
	`, userID)
	if err != nil {
		log.Printf("session retention cleanup failed: %v", err)
	}

	if deviceType == 2 {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(struct {
			Success  bool   `json:"success"`
			CSRF     string `json:"csrf"`
			UserID   int64  `json:"userId"`
			Message  string `json:"message"`
		}{
			Success:  true,
			CSRF:     csrf,
			UserID:   userID,
			Message:  "Success",
		})
		return
	}

	http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
}

func getGooglePublicKey(kid string) (*rsa.PublicKey, error) {
	set, err := jwk.Fetch(
		context.Background(),
		"https://www.googleapis.com/oauth2/v3/certs",
	)
	if err != nil {
		return nil, fmt.Errorf("fetch Google certificates: %w", err)
	}

	key, found := set.LookupKeyID(kid)
	if !found {
		return nil, fmt.Errorf("Google key not found: %s", kid)
	}

	var publicKey rsa.PublicKey
	if err := key.Raw(&publicKey); err != nil {
		return nil, fmt.Errorf("read Google public key: %w", err)
	}

	return &publicKey, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
