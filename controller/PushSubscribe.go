package controller

import (
	"context"
	"encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	 "chat/common"
)

func PushSubscribe(w http.ResponseWriter, r *http.Request) {
	if !json.Valid([]byte(r.FormValue("subscription"))) {
		log.Printf("Invalid JSON subscription: %s; Req:", r.URL.Path, r.Form)
		http.Error(w, "Invalid JSON subscription", http.StatusBadRequest)
		return
	}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect: %v; Req:", err, r.URL.Path, r.Form)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck: %v; Req:", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}

	coll := db1.Collection("session")
	filter := bson.D{{"_id", session.SessionID}}
	update := bson.D{{"$set", bson.D{
		{"subscription", r.FormValue("subscription")},
		{"updated_at", time.Now()}}}}
	opts := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts)

  fmt.Fprint(w, `{"Status":"1"}`)
}
