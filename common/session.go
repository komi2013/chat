package common

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"

	"chat/collection"
)

func secureRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func loadSession(ctx context.Context, sessionID string) (collection.SessionStruct, error) {
	var session collection.SessionStruct
	var pushContents pq.StringArray

	err := DB.SessionDB.QueryRowContext(
		ctx,
		`
		SELECT
			session_id,
			user_id,
			COALESCE(device_id, ''),
			COALESCE(created_at, NOW()),
			COALESCE(updated_at, NOW()),
			COALESCE(push_token, ''),
			COALESCE(device_type, 0),
			COALESCE(csrf, ''),
			COALESCE(push_contents, ARRAY[]::TEXT[]),
			COALESCE(mail, ''),
			COALESCE(telephone, '')
		FROM "session"
		WHERE session_id = $1
		`,
		sessionID,
	).Scan(
		&session.SessionID,
		&session.UserID,
		&session.DeviceID,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.PushToken,
		&session.DeviceType,
		&session.Csrf,
		&pushContents,
		&session.Mail,
		&session.Telephone,
	)
	if err != nil {
		return session, err
	}

	session.PushContents = []string(pushContents)
	return session, nil
}

func SessionGet(w http.ResponseWriter, r *http.Request) (collection.SessionStruct, error) {
	sessionID, _, err := sessionIDFromRequest(r)
	if err != nil {
		return collection.SessionStruct{}, err
	}

	if DB == nil || DB.SessionDB == nil {
		return collection.SessionStruct{}, errors.New("session database unavailable")
	}

	return loadSession(r.Context(), sessionID)
}

func SessionCheckTake(
	w http.ResponseWriter,
	r *http.Request,
	token string,
) (collection.SessionStruct, error) {
	if DB == nil || DB.SessionDB == nil {
		return collection.SessionStruct{}, errors.New("session database unavailable")
	}

	sessionID, bearerToken, err := sessionIDFromRequest(r)
	if err != nil {
		return collection.SessionStruct{}, err
	}

	session, err := loadSession(r.Context(), sessionID)
	if err != nil {
		return session, err
	}

	if session.Csrf != token {
		LogError("SessionCheckTake token error", nil)
		return session, errors.New("SessionCheckTake token error")
	}

	age := time.Since(session.UpdatedAt)

	// Expired sessions are removed from PostgreSQL.
	if age > 20*24*time.Hour {
		_, deleteErr := DB.SessionDB.ExecContext(
			r.Context(),
			`DELETE FROM "session" WHERE session_id = $1`,
			session.SessionID,
		)
		if deleteErr != nil {
			LogError("delete expired session", deleteErr)
		}

		if !bearerToken {
			http.SetCookie(w, &http.Cookie{
				Name:     "ss",
				Value:    "",
				MaxAge:   -1,
				Secure:   true,
				HttpOnly: true,
				Path:     "/",
			})
		}

		return session, errors.New("session expired")
	}

	// Regenerate old sessions after ten days.
	if age > 10*24*time.Hour {
		session, err = SessionRegenerate(session)
		if err != nil {
			return session, err
		}

		session.SessionIDRotated = true

		if bearerToken {
			// Android and other Bearer clients receive the rotated ID in a header.
			w.Header().Set("X-Session-Token", session.SessionID)
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:     "ss",
				Value:    session.SessionID,
				MaxAge:   2592000,
				Secure:   true,
				HttpOnly: true,
				Path:     "/",
			})
		}

		return session, nil
	}

	return CSRFcheckMake(session, token)
}

func sessionIDFromRequest(r *http.Request) (string, bool, error) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))

	if authorization != "" {
		parts := strings.Fields(authorization)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return "", false, errors.New("invalid Authorization header")
		}

		return parts[1], true, nil
	}

	cookie, err := r.Cookie("ss")
	if err != nil {
		return "", false, err
	}

	return cookie.Value, false, nil
}

func CSRFcheckMake(
	session collection.SessionStruct,
	token string,
) (collection.SessionStruct, error) {
	if session.Csrf != token {
		return session, errors.New("CSRFcheckMake token error")
	}

	session, err := CSRFReGenerate(session)
	if err != nil {
		return session, err
	}

	return PushReGenerate(session)
}

func CSRFReGenerate(
	session collection.SessionStruct,
) (collection.SessionStruct, error) {
	token, err := secureRandomToken()
	if err != nil {
		return session, err
	}

	session.Csrf = token
	return session, nil
}

func PushReGenerate(
	session collection.SessionStruct,
) (collection.SessionStruct, error) {
	if DB == nil || DB.SessionDB == nil {
		return session, errors.New("session database unavailable")
	}

	contents := session.PushContents

	var returnContents []string
	var remaining []string

	if len(contents) > 100 {
		returnContents = append([]string{}, contents[:100]...)
		remaining = append([]string{}, contents[100:]...)
	} else {
		returnContents = append([]string{}, contents...)
		remaining = []string{}
	}

	now := time.Now()

	result, err := DB.SessionDB.ExecContext(
		context.Background(),
		`
		UPDATE "session"
		SET csrf = $1,
		    updated_at = $2,
		    push_contents = $3
		WHERE session_id = $4
		`,
		session.Csrf,
		now,
		pq.Array(remaining),
		session.SessionID,
	)
	if err != nil {
		return session, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return session, err
	}
	if count == 0 {
		return session, errors.New("session no longer exists")
	}

	session.UpdatedAt = now
	session.PushContents = returnContents

	return session, nil
}

// SessionRegenerate replaces the session in a single transaction.
// Cookie/header updates are performed by SessionCheckTake after this succeeds.
func SessionRegenerate(
	session collection.SessionStruct,
) (collection.SessionStruct, error) {
	if DB == nil || DB.SessionDB == nil {
		return session, errors.New("session database unavailable")
	}

	sessionID, err := secureRandomToken()
	if err != nil {
		return session, err
	}

	csrf, err := secureRandomToken()
	if err != nil {
		return session, err
	}

	now := time.Now()
	newSession := session
	newSession.SessionID = sessionID
	newSession.Csrf = csrf
	newSession.CreatedAt = now
	newSession.UpdatedAt = now
	newSession.PushContents = []string{}
	newSession.SessionIDRotated = false

	tx, err := DB.SessionDB.BeginTx(context.Background(), nil)
	if err != nil {
		return session, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
		newSession.SessionID,
		newSession.UserID,
		newSession.DeviceID,
		newSession.CreatedAt,
		newSession.UpdatedAt,
		newSession.PushToken,
		newSession.DeviceType,
		newSession.Csrf,
		pq.Array(newSession.PushContents),
		newSession.Mail,
		newSession.Telephone,
	)
	if err != nil {
		return session, err
	}

	result, err := tx.Exec(
		`
		DELETE FROM "session"
		WHERE session_id = $1 AND user_id = $2
		`,
		session.SessionID,
		session.UserID,
	)
	if err != nil {
		return session, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return session, err
	}
	if count == 0 {
		return session, sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return session, err
	}

	return newSession, nil
}