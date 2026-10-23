package controller

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

	// Validate Csrf.
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

	// This endpoint uses the WEB Google OAuth client ID.
	// The mobile endpoint should use GOOGLE_CLIENT_MOBILE_ID instead.
	expectedClientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	if expectedClientID == "" {
		log.Printf("GOOGLE_CLIENT_ID is not configured")
		http.Error(w, "Google OAuth is not configured", http.StatusInternalServerError)
		return
	}

	// Verify Google ID token signature.
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

	// A valid Google ID token must have an expiration claim.
	// jwt.Parse above checks that the token has not expired when exp is present.
	if _, exists := claims["exp"]; !exists {
		http.Error(w, "missing token expiration", http.StatusUnauthorized)
		return
	}

	// Validate issuer.
	issuer, _ := claims["iss"].(string)
	if issuer != "https://accounts.google.com" &&
		issuer != "accounts.google.com" {
		http.Error(w, "invalid token issuer", http.StatusUnauthorized)
		return
	}

	// Validate audience against the web OAuth client ID.
	if !googleAudienceMatches(claims, expectedClientID) {
		http.Error(w, "invalid Google client ID", http.StatusUnauthorized)
		return
	}

	googleSub, ok := claims["sub"].(string)
	if !ok || googleSub == "" {
		http.Error(w, "missing Google subject", http.StatusUnauthorized)
		return
	}

	now := time.Now()

	// Create or reuse the browser's device ID.
	// This cookie is only for counting sessions by browser/device.
	deviceID, err := getOrCreateWebDeviceID(w, r)
	if err != nil {
		log.Printf("device ID creation failed: %v", err)
		http.Error(w, "cannot create device ID", http.StatusInternalServerError)
		return
	}

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

	// Look up the current session using the authentication cookie.
	// deviceID is NOT used for authentication or session deletion.
	currentSessionID := ""
	if c, err := r.Cookie("ss"); err == nil {
		currentSessionID = c.Value
	}

	type previousSession struct {
		deviceType   int
		pushContents []string
		pushToken    string
	}

	var previous previousSession
	previous.pushContents = []string{}
	sameBrowser := false

	if currentSessionID != "" {
		var pushContents pq.StringArray

		err := sessionDB.QueryRowContext(ctx, `
			SELECT
				COALESCE(device_type, 0),
				COALESCE(push_contents, ARRAY[]::TEXT[]),
				COALESCE(push_token, '')
			FROM "session"
			WHERE user_id = $1
			  AND session_id = $2
		`, userID, currentSessionID).Scan(
			&previous.deviceType,
			&pushContents,
			&previous.pushToken,
		)

		switch {
		case err == nil:
			previous.pushContents = []string(pushContents)
			sameBrowser = true

		case err == sql.ErrNoRows:
			// The cookie may refer to a deleted or expired session.
			// Continue with a new session.

		default:
			log.Printf("current session lookup failed: %v", err)
			http.Error(w, "session lookup fail", http.StatusInternalServerError)
			return
		}
	}

	// Generate a secure session ID and Csrf token.
	sessionID, err := randomToken()
	if err != nil {
		log.Printf("session ID creation failed: %v", err)
		http.Error(w, "cannot create session", http.StatusInternalServerError)
		return
	}

	csrf, err := randomToken()
	if err != nil {
		log.Printf("Csrf token creation failed: %v", err)
		http.Error(w, "cannot create session", http.StatusInternalServerError)
		return
	}

	// Preserve push registration and display settings when replacing
	// the session associated with this browser.
	if previous.pushToken != "" {
		deviceType = previous.deviceType
	}

	pushContents := previous.pushContents
	if pushContents == nil {
		pushContents = []string{}
	}

	// Store device_id only as a session-counting identifier.
	// The SQL has 11 columns and 10 parameters because created_at
	// and updated_at both use $4.
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
		VALUES ($1, $2, $3, $4, $4, $5, $6, $7, $8, $9, $10)
	`,
		sessionID,
		userID,
		deviceID,
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

	// Authentication cookie. Keep its path at "/" because other
	// endpoints need this cookie to authenticate the user.
	http.SetCookie(w, &http.Cookie{
		Name:     "ss",
		Value:    sessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Replace the previous session for this authenticated browser.
	// This uses the session cookie, NOT deviceID.
	if sameBrowser && currentSessionID != "" {
		_, err := sessionDB.ExecContext(ctx, `
			DELETE FROM "session"
			WHERE session_id = $1
			  AND user_id = $2
		`, currentSessionID, userID)
		if err != nil {
			log.Printf("old session deletion failed: %v", err)
		}
	}

	// Delete sessions older than one month.
	_, err = sessionDB.ExecContext(ctx, `
		DELETE FROM "session"
		WHERE user_id = $1
		  AND updated_at < $2
	`, userID, now.AddDate(0, -1, 0))
	if err != nil {
		log.Printf("expired session cleanup failed: %v", err)
	}

	// Keep at most the five most recently updated sessions per user.
	// This limit is per user, not per device.
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

	if sameBrowser {
		http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
		return
	}

	if deviceType == 2 {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(struct {
			Success bool   `json:"success"`
			Csrf    string `json:"csrf"`
			UserID  int64  `json:"userId"`
			Message string `json:"message"`
		}{
			Success: true,
			Csrf:    csrf,
			UserID:  userID,
			Message: "Success",
		})
		return
	}

	http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
}

// googleAudienceMatches verifies the token's aud claim against the expected
// OAuth client ID. Google tokens can represent aud as a string or an array.
func googleAudienceMatches(claims jwt.MapClaims, expectedClientID string) bool {
	if expectedClientID == "" {
		return false
	}

	switch aud := claims["aud"].(type) {
	case string:
		return aud == expectedClientID

	case []interface{}:
		for _, item := range aud {
			if value, ok := item.(string); ok && value == expectedClientID {
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

// getOrCreateWebDeviceID reads or generates a UUID for the browser.
//
// The device_id cookie is scoped to the current sign-in endpoint path.
// It is NOT an authentication credential and is NOT used to delete sessions.
func getOrCreateWebDeviceID(w http.ResponseWriter, r *http.Request) (string, error) {
	const cookieName = "device_id"

	var deviceID string

	if c, err := r.Cookie(cookieName); err == nil && isUUIDString(c.Value) {
		deviceID = c.Value
	} else {
		var err error
		deviceID, err = newWebDeviceID()
		if err != nil {
			return "", err
		}
	}

	cookiePath := r.URL.Path
	if cookiePath == "" {
		cookiePath = "/"
	}

	secureCookie := r.TLS != nil ||
		strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")

	// Remove a previously created root-scoped device_id cookie, if present.
	// The replacement below is scoped to this sign-in endpoint.
	if cookiePath != "/" {
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   secureCookie,
			SameSite: http.SameSiteLaxMode,
		})
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    deviceID,
		Path:     cookiePath,
		MaxAge:   365 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	return deviceID, nil
}

// newWebDeviceID generates a UUID v4 using a cryptographically secure RNG.
func newWebDeviceID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// UUID v4 and RFC 4122 variant bits.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	hexID := hex.EncodeToString(b)

	return hexID[0:8] + "-" +
		hexID[8:12] + "-" +
		hexID[12:16] + "-" +
		hexID[16:20] + "-" +
		hexID[20:32], nil
}

func isUUIDString(value string) bool {
	if len(value) != 36 {
		return false
	}

	if value[8] != '-' ||
		value[13] != '-' ||
		value[18] != '-' ||
		value[23] != '-' {
		return false
	}

	_, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil
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