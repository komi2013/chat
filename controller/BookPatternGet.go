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

func BookPatternGet(w http.ResponseWriter, r *http.Request) {
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

	coll := db1.Collection("book_pattern")

	bookPatternID, err := primitive.ObjectIDFromHex(r.FormValue("bookPatternID"))
	if err != nil {
		http.Error(w, "Invalid bookPatternID format", http.StatusBadRequest)
		return
	}

	// Find the document by `_id`
	var bookPattern collection.BookPatternStruct
	filter := bson.M{"_id": bookPatternID}
	err = coll.FindOne(ctx, filter).Decode(&bookPattern)
	if err != nil {
		log.Print(err, " bookPattern ", bookPatternID)
		http.Error(w, "Book pattern not found", http.StatusNotFound)
		return
	}

	// Respond with the bookPattern in JSON format
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bookPattern); err != nil {
		http.Error(w, "Failed to encode response to JSON", http.StatusInternalServerError)
	}
}
