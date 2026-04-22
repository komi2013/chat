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
	if session.Subscription == "" {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendWebPushNotification no subscription data: %v", functionName)
		return nil, fmt.Errorf("no subscription data")
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

  resp, err := PushNotification(string(jsonData), session.Subscription)
  return resp, err
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
      } else {
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
      } else {
        defer resp.Body.Close()
      }
    }
  }
}

func SendWebJustPush(arr []interface{}, pushID string, session collection.SessionStruct) (*http.Response, error) {
	if session.Subscription == "" {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendWebJustPush no subscription data: %v", functionName)
		return nil, fmt.Errorf("no subscription data")
	}
	jsonData, err := json.Marshal(arr)
	if err != nil {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendWebJustPush arr: %v", functionName, err)
	}
  resp, err := PushNotification(string(jsonData), session.Subscription)
  return resp, err
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

	// Send Web Push if subscription exists
	if session.Subscription != "" {
		resp, err := PushNotification(string(jsonData), session.Subscription)
		if err != nil {
			log.Printf("Web Push failed: %v", err)
		} else {
			defer resp.Body.Close()
		}
	}

	// Send FCM if FCM token exists and it's a mobile session
	if session.FcmToken != "" && session.IsMobile {
		cfg := LoadConfig()
		fcmManager := NewFCMManager(cfg)
		
		// Extract title and body from the notification data
		title := "New Message"
		body := "You have a new message"
		channelId := ""
		
		if len(arr) > 1 {
			if titleStr, ok := arr[1].(string); ok {
				title = titleStr
			}
		}
		if len(arr) > 2 {
			if bodyStr, ok := arr[2].(string); ok {
				body = bodyStr
			}
		}
		if len(arr) > 3 {
			if channelStr, ok := arr[3].(string); ok {
				channelId = channelStr
			}
		}
		
		err = fcmManager.SendNotification(session.FcmToken, title, body, channelId)
		if err != nil {
			log.Printf("FCM Push failed: %v", err)
		}
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
