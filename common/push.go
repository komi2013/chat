package common

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "runtime"
  // "time"
  "unicode/utf8"

  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"

  "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
)

func SendWebPushNotification(arr []interface{}, pushID string, session collection.SessionStruct) (*http.Response, error) {
	if _, token := ResolvePushTarget(session); token == "" {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendWebPushNotification no push target registered: %s", functionName, session.SessionID)
		return nil, nil
	}
	jsonData, err := json.Marshal(arr)
	if err != nil {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendWebPushNotification arr: %v", functionName, err)
	}

	// coll := db1.Collection("session")
	coll := DB.SessionDB.Collection("session")
	filter := bson.D{{"_id", session.SessionID}}
	update := bson.M{
		"$push": bson.M{
			"pushContents": string(jsonData),
		},
	}
	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendWebPushNotification session UpdateOne: %v", functionName, err)
	}

  return dispatchPush(arr, pushID, string(jsonData), session)
}

func PushNotification(data string, subscription string) (*http.Response, error) {
  cfg := LoadConfig()
	webpushSub := &webpush.Subscription{}
	json.Unmarshal([]byte(subscription), webpushSub)
	resp, err := webpush.SendNotification([]byte(data), webpushSub, &webpush.Options{
		Subscriber:      "example@example.com",
    VAPIDPublicKey:  cfg.VAPIDPublicKey,
    VAPIDPrivateKey: cfg.VAPIDPrivateKey,
		TTL:             30,
	})
  if err != nil {
  	// log.Printf("push fail: %v; ", err)
    return nil, fmt.Errorf("failed to send notification: %w", err)
  }
  return resp, nil
}

// ResolvePushTarget returns the push transport and token registered for a session.
// Only the merged pushToken/deviceType fields are used. Sessions without a
// registered token resolve to (0, "").
//
//	deviceType: 0 = none, 1 = VAPID (web), 2 = FCM (Android), 3 = APNs (iOS).
func ResolvePushTarget(session collection.SessionStruct) (int, string) {
	if session.PushToken != "" {
		switch session.DeviceType {
		case 1, 2, 3:
			return session.DeviceType, session.PushToken
		default:
			// PushToken inherits the deviceType written at registration time,
			// so an unknown value means corrupt data — never guess a transport.
			return 0, ""
		}
	}
	return 0, ""
}

// deviceTypeLabel is used for logs and console output.
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

// sendFCMPush forwards the payload array as an FCM data-only message.
// The client dispatches on pd[1] like vue pushReceive.js and decides
// importance/display — the server never maps event types.
func sendFCMPush(arr []interface{}, pushID string, session collection.SessionStruct) error {
	_, token := ResolvePushTarget(session)
	if token == "" {
		return nil
	}
	jsonData, err := json.Marshal(arr)
	if err != nil {
		return fmt.Errorf("sendFCMPush marshal: %w", err)
	}
	return NewFCMManager(LoadConfig()).SendData(pushID, string(jsonData), token)
}

// dispatchPush routes the payload to the transport registered for this session.
// The returned response is nil for non-web transports, so callers must nil-check it.
func dispatchPush(arr []interface{}, pushID string, jsonData string, session collection.SessionStruct) (*http.Response, error) {
	// VAPID gets the raw JSON array; FCM gets the same JSON as a data-only
	// message. The client dispatches on pd[1] (like vue pushReceive.js) and
	// decides importance/display — the server never maps event types.
	deviceType, token := ResolvePushTarget(session)
	// 0 = none, 1 = VAPID (web), 2 = FCM (Android), 3 = APNs (iOS).
	switch deviceType {
	case 1:
		return PushNotification(jsonData, token)
	case 2:
		if err := sendFCMPush(arr, pushID, session); err != nil {
			return nil, err
		}
		return nil, nil
	case 3:
		// TODO: APNs sender is not implemented yet.
		log.Printf("[dispatchPush] APNs is not implemented yet: session=%s", session.SessionID)
		return nil, nil
	default:
		return nil, nil
	}
}

func ChunkPush(sessions []collection.SessionStruct, arr []interface{}) {
	jsonData, _ := json.Marshal(arr)
  const maxChunkSize = 2000 / utf8.UTFMax

  chunk := false
  chunks := []string{string(jsonData)}
  if len(jsonData) > 2000 {
    chunk = true
    chunks = SplitIntoByteChunks(string(jsonData), maxChunkSize)
  }

  chunkLength := len(chunks)
  chunkPass := StringRand(2)

  for chIndex, ch := range chunks {
    for _, session := range sessions {
      pushID := StringRand(1)
      var arrForJson []interface{}

      if chunk {
        arrForJson = []interface{}{
          pushID, "chunk", ch, chunkPass, chIndex, chunkLength,
        }
      } else {
        arrForJson = append([]interface{}{pushID}, arr...)
      }
      resp, err := SendWebPushNotification(arrForJson, pushID, session)
      if err != nil {
        LogError("SendWebPushNotification:", err)
      } else if resp != nil {
        defer resp.Body.Close()
      }
    }
  }
}

func ChunkJustPush(sessions []collection.SessionStruct, arr []interface{}) {
	jsonData, _ := json.Marshal(arr)
  const maxChunkSize = 2000 / utf8.UTFMax

  chunk := false
  chunks := []string{string(jsonData)}
  if len(jsonData) > 2000 {
    chunk = true
    chunks = SplitIntoByteChunks(string(jsonData), maxChunkSize)
  }

  chunkLength := len(chunks)
  chunkPass := StringRand(2)

  for chIndex, ch := range chunks {
    for _, session := range sessions {
      pushID := StringRand(1)
      var arrForJson []interface{}

      if chunk {
        arrForJson = []interface{}{
          pushID, "chunk", ch, chunkPass, chIndex, chunkLength,
        }
      } else {
        arrForJson = append([]interface{}{pushID}, arr...)
      }
      resp, err := SendWebJustPush(arrForJson, pushID, session)
      if err != nil {
        LogError("SendWebJustPush:", err)
      } else if resp != nil {
        defer resp.Body.Close()
      }
    }
  }
}

func SendWebJustPush(arr []interface{}, pushID string, session collection.SessionStruct) (*http.Response, error) {
	if _, token := ResolvePushTarget(session); token == "" {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendWebJustPush no push target registered: %s", functionName, session.SessionID)
		return nil, nil
	}
	jsonData, err := json.Marshal(arr)
	if err != nil {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendWebJustPush arr: %v", functionName, err)
	}
  return dispatchPush(arr, pushID, string(jsonData), session)
}

// Unified push notification function that supports both Web Push and FCM
func SendPushNotification(arr []interface{}, pushID string, session collection.SessionStruct) error {
	jsonData, err := json.Marshal(arr)
	if err != nil {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendPushNotification arr: %v", functionName, err)
		return err
	}

	// Store push content in session
	coll := DB.SessionDB.Collection("session")
	filter := bson.D{{"_id", session.SessionID}}
	update := bson.M{
		"$push": bson.M{
			"pushContents": string(jsonData),
		},
	}
	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendPushNotification session UpdateOne: %v", functionName, err)
	}

	// Route to the transport registered for this session (Web Push / FCM / APNs).
	// dispatchPush expects arr = [pushID, event, ...payload].
	if _, err := dispatchPush(arr, pushID, string(jsonData), session); err != nil {
		log.Printf("SendPushNotification dispatch failed: %v", err)
	}

	return nil
}

// FCM-only push notification function
func SendFCMPushNotification(title, body, channelId, fcmToken string) error {
	if fcmToken == "" {
		return fmt.Errorf("no FCM token provided")
	}
	
	cfg := LoadConfig()
	fcmManager := NewFCMManager(cfg)
	
	err := fcmManager.SendNotification(fcmToken, title, body, channelId)
	if err != nil {
		return fmt.Errorf("FCM push failed: %w", err)
	}
	
	return nil
}
