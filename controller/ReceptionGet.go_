package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func ReceptionGet(w http.ResponseWriter, r *http.Request) {
  _, err := common.Session(w,r)
  if err != nil {
    http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
    return
  }
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  coll := db1.Collection("reception")

  receptionID, err := primitive.ObjectIDFromHex(r.FormValue("receptionID"))
  if err != nil {
    http.Error(w, "Invalid receptionID format", http.StatusBadRequest)
    return
  }

  var reception collection.ReceptionStruct
  filter := bson.M{"_id": receptionID}
  err = coll.FindOne(ctx, filter).Decode(&reception)
  if err != nil {
    log.Print(err, " reception ", receptionID)
    http.Error(w, "reception not found", http.StatusNotFound)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(reception); err != nil {
    http.Error(w, "Failed to encode response to JSON", http.StatusInternalServerError)
  }
}
