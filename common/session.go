package common

import (
  "context"
  // "encoding/json"
  // "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  // webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
)

func Session(w http.ResponseWriter, r *http.Request) (collection.SessionStruct, error) {
  var session collection.SessionStruct
  cookie, err := r.Cookie("ss")
  if err != nil {
    return session, err
  }
  // if len(cookie.Value) != 16 {
 //    http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
 //    return
  // }
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(MongoDb1)


  coll := db1.Collection("session")
  filter := bson.D{{"_id", cookie.Value}}
  opts := options.FindOne().SetProjection(bson.D{})
  err = coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  if err != nil {
    return session, err
  }
  return session, nil
}
