package controller

import (
  "context"
  // "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"
)

func WindowGet(w http.ResponseWriter, r *http.Request) {
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  var window collection.WindowStruct
  coll := db1.Collection("window")
  filter := bson.M{"_id": r.FormValue("windowID")}
  err = coll.FindOne(context.TODO(), filter).Decode(&window)
  if err != nil {
    log.Print(err, "window", r.FormValue("windowID"))
  }
  fmt.Fprint(w, window.Contents)
}

