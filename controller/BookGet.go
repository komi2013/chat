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

  "chat/collection"
  "chat/common"
)

func BookGet(w http.ResponseWriter, r *http.Request) {
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  coll := db1.Collection("book")
  filter := bson.M{
    "window_id": r.FormValue("windowID"),
  }
  project := bson.D{
    {"_id", 0},
    {"window_id", 0},
  }
  opts := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts)
  if err != nil {
    fmt.Printf(" book find err %s\n", err)
  }
  var books []collection.BookStruct
  if err = cursor.All(context.TODO(), &books); err != nil {
    fmt.Printf(" book all err %s\n", err)
    return
  }
  fmt.Printf(" no record %s\n", books)
  if len(books) == 0 {
    w.WriteHeader(http.StatusOK)
    return
  }

  w.Header().Set("Content-Type", "application/json")

  if err := json.NewEncoder(w).Encode(books); err != nil {
    http.Error(w, "Failed to encode books to JSON", http.StatusInternalServerError)
  }

  // fmt.Fprint(w, books)
}

