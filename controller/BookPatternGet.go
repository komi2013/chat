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

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")
	bookPatternID, err := primitive.ObjectIDFromHex(r.FormValue("bookPatternID"))
	if err != nil {
		http.Error(w, "Invalid bookPatternID format", http.StatusBadRequest)
		return
	}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheck: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  trueAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    log.Printf("ChannelAliases !trueAccess: %v; Req: ", session.ChannelAliases, r.URL.Path, r.Form)
    return
  }

	coll := db1.Collection("book_pattern")

	var bookPattern collection.BookPatternStruct
	filter := bson.M{"_id": bookPatternID}
	err = coll.FindOne(ctx, filter).Decode(&bookPattern)
	if err != nil {
		log.Print(err, " bookPattern ", bookPatternID)
		http.Error(w, "Book pattern not found", http.StatusNotFound)
		return
	}

  responseData := struct {
    Csrf         string        `json:"csrf"`
    PushContents []string `json:"pushContents"`
    BookPattern collection.BookPatternStruct `json:"bookPattern"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    BookPattern: bookPattern,
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)


}
