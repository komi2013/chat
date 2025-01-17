package common

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "runtime"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"

  "github.com/SherClockHolmes/webpush-go"
)

func SendWebPushNotification(db1 *mongo.Database, arr []interface{}, pushID string, subscription string) (*http.Response, error) {
	if subscription == "" {
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
	coll := db1.Collection("push")
	document := bson.M{
    "_id": pushID,
    "contents": string(jsonData),
    "created_at": time.Now().Format("2006-01-02 15:04:05"),
	}
	_, err = coll.InsertOne(context.TODO(), document)
	if err != nil {
		pc, _, _, _ := runtime.Caller(1)
		functionName := runtime.FuncForPC(pc).Name()
		log.Printf("[%s] SendWebPushNotification push InsertOne: %v", functionName, err)
	}
  resp, err := PushNotification(string(jsonData), subscription)
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

