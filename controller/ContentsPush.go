package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"
	// "unicode/utf8"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  // webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func ContentsPush(w http.ResponseWriter, r *http.Request) {
	var userIDs []string
  if err := json.Unmarshal([]byte(r.FormValue("userIDs")), &userIDs); err != nil {
  	log.Printf("userIDs: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON userIDs", http.StatusBadRequest)
    return
  }

  updatedBy := r.FormValue("updatedBy")
  channelID := r.FormValue("channelID")

  var contents interface{}
  if err := json.Unmarshal([]byte(r.FormValue("contents")), &contents); err != nil {
  	log.Printf("contents: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON contents", http.StatusBadRequest)
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
  for _, d := range session.AliasChannels {
    if d.Alias == updatedBy && d.ChannelID == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    log.Printf("AliasChannels !trueAccess: %v; Req: ", session.AliasChannels, updatedBy, channelID, r.URL.Path, r.Form)
    return
  }

  coll := db1.Collection("session")
  filter := bson.D{{
    "user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"subscription", 1}}
  opts4 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts4)
  if err != nil {
    log.Printf("Find session : %v; Req: ", err, userIDs, r.URL.Path, r.Form)
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
    log.Printf("All session : %v; Req: ", err, userIDs, r.URL.Path, r.Form)
  }

  var arr []interface{}
  arr = append(arr, r.FormValue("pushTitle"))
  arr = append(arr, channelID)
  arr = append(arr, updatedBy)
  arr = append(arr, contents)
	common.ChunkPush(sessions, db1, r, arr)
  fmt.Fprint(w, `{"Status":"1"}`)
}

