package common

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/SherClockHolmes/webpush-go"

	"chat/table"
)

// pushContentExecutor is implemented by both *sql.DB and *sql.Tx.
type pushContentExecutor interface {
	ExecContext(
		context.Context,
		string,
		...interface{},
	) (sql.Result, error)
}

// savePushContent stores a notification in the session's PostgreSQL TEXT[].
func savePushContent(
	ctx context.Context,
	db pushContentExecutor,
	sessionID string,
	content string,
) error {
	result, err := db.ExecContext(
		ctx,
		`
		UPDATE "session"
		SET push_contents = array_append(
			COALESCE(push_contents, ARRAY[]::TEXT[]),
			$1
		)
		WHERE session_id = $2
		`,
		content,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("save push content: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check push content update: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// closePushResponse closes an HTTP response returned by Web Push.
func closePushResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

// dispatchPushAndLog attempts delivery. Delivery errors are logged rather
// than returned to the controller. A returned response must still be closed.
func dispatchPushAndLog(
	arr []interface{},
	pushID string,
	jsonData string,
	session table.SessionStruct,
) *http.Response {
	if _, token := ResolvePushTarget(session); token == "" {
		log.Printf(
			"push: no registered push target for session %s",
			session.SessionID,
		)
		return nil
	}

	resp, err := dispatchPush(arr, pushID, jsonData, session)
	if err != nil {
		log.Printf(
			"push delivery failed for session %s: %v",
			session.SessionID,
			err,
		)
		closePushResponse(resp)
		return nil
	}

	if resp != nil && resp.StatusCode >= 300 {
		log.Printf(
			"push endpoint returned HTTP %s for session %s",
			resp.Status,
			session.SessionID,
		)
	}

	return resp
}

// SendWebPushNotification saves the notification first, then attempts delivery.
// Only JSON/database errors are returned to the caller.
func SendWebPushNotification(
	arr []interface{},
	pushID string,
	session table.SessionStruct,
) (*http.Response, error) {
	jsonData, err := json.Marshal(arr)
	if err != nil {
		return nil, fmt.Errorf("marshal push notification: %w", err)
	}

	if DB == nil || DB.SessionDB == nil {
		return nil, fmt.Errorf("session database is not initialized")
	}

	if err := savePushContent(
		context.Background(),
		DB.SessionDB,
		session.SessionID,
		string(jsonData),
	); err != nil {
		return nil, err
	}

	// Delivery errors are logged, not propagated.
	resp := dispatchPushAndLog(arr, pushID, string(jsonData), session)
	return resp, nil
}

func PushNotification(
	data string,
	subscription string,
) (*http.Response, error) {
	cfg := LoadConfig()

	webpushSub := &webpush.Subscription{}
	if err := json.Unmarshal([]byte(subscription), webpushSub); err != nil {
		return nil, fmt.Errorf("invalid Web Push subscription: %w", err)
	}

	resp, err := webpush.SendNotification(
		[]byte(data),
		webpushSub,
		&webpush.Options{
			Subscriber:      "example@example.com",
			VAPIDPublicKey:  cfg.VAPIDPublicKey,
			VAPIDPrivateKey: cfg.VAPIDPrivateKey,
			TTL:             30,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	return resp, nil
}

// ResolvePushTarget returns the registered transport and token.
//
//	deviceType: 0 = none, 1 = VAPID (web), 2 = FCM (Android), 3 = APNs (iOS).
func ResolvePushTarget(session table.SessionStruct) (int, string) {
	if session.PushToken == "" {
		return 0, ""
	}

	switch session.DeviceType {
	case 1, 2, 3:
		return session.DeviceType, session.PushToken
	default:
		return 0, ""
	}
}

func deviceTypeLabel(deviceType int) string {
	switch deviceType {
	case 1:
		return "vapid"
	case 2:
		return "fcm"
	case 3:
		return "apns"
	default:
		return "none"
	}
}

// sendFCMPush sends an FCM data-only message.
func sendFCMPush(
	arr []interface{},
	pushID string,
	session table.SessionStruct,
) error {
	_, token := ResolvePushTarget(session)
	if token == "" {
		return nil
	}

	jsonData, err := json.Marshal(arr)
	if err != nil {
		return fmt.Errorf("marshal FCM push: %w", err)
	}

	return NewFCMManager(LoadConfig()).SendData(
		pushID,
		string(jsonData),
		token,
	)
}

// dispatchPush sends through the transport registered for this session.
// Push delivery errors are returned here so dispatchPushAndLog can log them.
func dispatchPush(
	arr []interface{},
	pushID string,
	jsonData string,
	session table.SessionStruct,
) (*http.Response, error) {
	deviceType, _ := ResolvePushTarget(session)

	switch deviceType {
	case 1:
		return PushNotification(jsonData, session.PushToken)

	case 2:
		if err := sendFCMPush(arr, pushID, session); err != nil {
			return nil, err
		}
		return nil, nil

	case 3:
		// TODO: APNs sender is not implemented yet.
		log.Printf(
			"APNs is not implemented: session=%s",
			session.SessionID,
		)
		return nil, nil

	default:
		return nil, nil
	}
}

// ChunkPush queues every notification before attempting any delivery.
// If one database write fails, the transaction rolls back and no pushes
// are sent. Push delivery errors are logged only.
func ChunkPush(
	sessions []table.SessionStruct,
	arr []interface{},
) error {
	jsonData, err := json.Marshal(arr)
	if err != nil {
		return fmt.Errorf("ChunkPush marshal: %w", err)
	}

	if len(sessions) == 0 {
		return nil
	}

	if DB == nil || DB.SessionDB == nil {
		return fmt.Errorf("session database is not initialized")
	}

	const maxChunkSize = 2000 / utf8.UTFMax

	chunk := len(jsonData) > 2000
	chunks := []string{string(jsonData)}
	if chunk {
		chunks = SplitIntoByteChunks(string(jsonData), maxChunkSize)
	}

	type pendingPush struct {
		session table.SessionStruct
		pushID  string
		payload []interface{}
		json    string
	}

	chunkLength := len(chunks)
	chunkPass := StringRand(2)
	pending := make([]pendingPush, 0, len(chunks)*len(sessions))

	// Prepare all payloads before writing them.
	for chunkIndex, chunkData := range chunks {
		for _, session := range sessions {
			pushID := StringRand(1)

			var payload []interface{}
			if chunk {
				payload = []interface{}{
					pushID,
					"chunk",
					chunkData,
					chunkPass,
					chunkIndex,
					chunkLength,
				}
			} else {
				payload = append([]interface{}{pushID}, arr...)
			}

			encoded, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("marshal chunk payload: %w", err)
			}

			pending = append(pending, pendingPush{
				session: session,
				pushID:  pushID,
				payload: payload,
				json:    string(encoded),
			})
		}
	}

	// First: store every notification in one transaction.
	ctx := context.Background()
	tx, err := DB.SessionDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin ChunkPush transaction: %w", err)
	}
	defer tx.Rollback()

	for _, item := range pending {
		if err := savePushContent(
			ctx,
			tx,
			item.session.SessionID,
			item.json,
		); err != nil {
			return fmt.Errorf(
				"ChunkPush: save content for session %s: %w",
				item.session.SessionID,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit ChunkPush transaction: %w", err)
	}

	// Second: delivery happens only after the database transaction commits.
	for _, item := range pending {
		resp := dispatchPushAndLog(
			item.payload,
			item.pushID,
			item.json,
			item.session,
		)
		closePushResponse(resp)
	}

	return nil
}

// ChunkJustPush sends chunked payloads without saving push_contents.
func ChunkJustPush(
	sessions []table.SessionStruct,
	arr []interface{},
) {
	jsonData, err := json.Marshal(arr)
	if err != nil {
		LogError("ChunkJustPush marshal", err)
		return
	}

	const maxChunkSize = 2000 / utf8.UTFMax

	chunk := len(jsonData) > 2000
	chunks := []string{string(jsonData)}
	if chunk {
		chunks = SplitIntoByteChunks(string(jsonData), maxChunkSize)
	}

	chunkLength := len(chunks)
	chunkPass := StringRand(2)

	for chunkIndex, chunkData := range chunks {
		for _, session := range sessions {
			pushID := StringRand(1)

			var payload []interface{}
			if chunk {
				payload = []interface{}{
					pushID,
					"chunk",
					chunkData,
					chunkPass,
					chunkIndex,
					chunkLength,
				}
			} else {
				payload = append([]interface{}{pushID}, arr...)
			}

			resp, err := SendWebJustPush(payload, pushID, session)
			if err != nil {
				LogError("ChunkJustPush", err)
			}
			closePushResponse(resp)
		}
	}
}

// SendWebJustPush dispatches without saving to push_contents.
func SendWebJustPush(
	arr []interface{},
	pushID string,
	session table.SessionStruct,
) (*http.Response, error) {
	jsonData, err := json.Marshal(arr)
	if err != nil {
		return nil, fmt.Errorf("marshal push notification: %w", err)
	}

	resp := dispatchPushAndLog(arr, pushID, string(jsonData), session)
	return resp, nil
}

// SendPushNotification saves content first, then attempts delivery.
// Database failures are returned; push-delivery failures are logged only.
func SendPushNotification(
	arr []interface{},
	pushID string,
	session table.SessionStruct,
) error {
	jsonData, err := json.Marshal(arr)
	if err != nil {
		return fmt.Errorf("marshal push notification: %w", err)
	}

	if DB == nil || DB.SessionDB == nil {
		return fmt.Errorf("session database is not initialized")
	}

	if err := savePushContent(
		context.Background(),
		DB.SessionDB,
		session.SessionID,
		string(jsonData),
	); err != nil {
		return err
	}

	// Delivery errors are logged, not propagated.
	resp := dispatchPushAndLog(arr, pushID, string(jsonData), session)
	closePushResponse(resp)

	return nil
}

// SendFCMPushNotification sends an FCM notification directly.
func SendFCMPushNotification(
	title string,
	body string,
	channelID string,
	fcmToken string,
) error {
	if fcmToken == "" {
		return fmt.Errorf("no FCM token provided")
	}

	fcmManager := NewFCMManager(LoadConfig())
	if err := fcmManager.SendNotification(
		fcmToken,
		title,
		body,
		channelID,
	); err != nil {
		return fmt.Errorf("FCM push failed: %w", err)
	}

	return nil
}