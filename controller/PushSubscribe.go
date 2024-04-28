package controller

import (
	"context"
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
	cookie, _ := r.Cookie("ss")
  log.Println(r.URL)
  log.Println("hihii")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	if err != nil {
		log.Print(err)
	}
	defer c.Disconnect(ctx)
	db1 := c.Database(common.MongoDb1)
	
	coll := db1.Collection("session")
	filter := bson.D{{"_id", cookie.Value}}
	update := bson.D{{"$set", bson.D{
		{"subscription", r.FormValue("subscription")},
		{"updated_at", time.Now()}}}}
	opts := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts)

  fmt.Fprint(w, `{"Status":"1"}`)
}
