package common

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "runtime"
  // "time"

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
			"push_contents": string(jsonData),
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

