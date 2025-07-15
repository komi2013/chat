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

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"

  "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
)

func SendWebPushNotification(db1 *mongo.Database, arr []interface{}, pushID string, session collection.SessionStruct) (*http.Response, error) {
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

	coll := db1.Collection("session")
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
	webpushSub := &webpush.Subscription{}
	json.Unmarshal([]byte(subscription), webpushSub)
	resp, err := webpush.SendNotification([]byte(data), webpushSub, &webpush.Options{
		Subscriber:      "example@example.com",
    VAPIDPublicKey:  VAPIDPublicKey,
    VAPIDPrivateKey: VAPIDPrivateKey,
		TTL:             30,
	})
  if err != nil {
  	// log.Printf("push fail: %v; ", err)
    return nil, fmt.Errorf("failed to send notification: %w", err)
  }
  return resp, nil
}

func ChunkPush(
  sessions []collection.SessionStruct,
  db1 *mongo.Database,
  arr []interface{},
) {
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
      resp, err := SendWebPushNotification(db1, arrForJson, pushID, session)
      if err != nil {
        LogError("SendWebPushNotification:", err)
      } else {
        defer resp.Body.Close()
      }
    }
  }
}

func ChunkJustPush(
  sessions []collection.SessionStruct,
  db1 *mongo.Database,
  arr []interface{},
) {
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
      resp, err := SendWebJustPush(db1, arrForJson, pushID, session)
      if err != nil {
        LogError("SendWebJustPush:", err)
      } else {
        defer resp.Body.Close()
      }
    }
  }
}

func SendWebJustPush(db1 *mongo.Database, arr []interface{}, pushID string, session collection.SessionStruct) (*http.Response, error) {
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
