package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  "log"
  // "math"
  "net/http"
  // "time"

  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/mongo/options"
)

func UserGet(w http.ResponseWriter, r *http.Request) {

  // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  // defer cancel()
  // c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  // if err != nil {
  //   log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  // }
  // defer c.Disconnect(ctx)
  // db1 := c.Database(common.MongoDb1)

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheck: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  // collUser := db1.Collection("user")
  collUser := common.DB.UserDB.Collection("user")
  filterUser := bson.M{"_id": session.UserID}
  var user collection.UserStruct
  err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
  if err != nil {
    log.Printf("user FindOne: %v; Req:", err, r.URL.Path, r.Form)
  }

  collNickname := common.DB.NicknameDB.Collection("nickname")
  filterNickname := bson.M{"userID": session.UserID}
  cursor, err := collNickname.Find(context.TODO(), filterNickname)
  if err != nil {
    log.Printf("coll.Find: %v; Req: ", err, session.UserID, r.URL.Path, r.Form)
  }
  var nicknames []collection.NicknameStruct
  if err = cursor.All(context.TODO(), &nicknames); err != nil {
    log.Printf("cursor.All: %v; Req: ", err, session.UserID, r.URL.Path, r.Form)
  }

  responseData := struct {
    Csrf         string       `json:"csrf"`
    PushContents []string     `json:"pushContents"`
    User       collection.UserStruct  `json:"user"`
    Nicknames   []collection.NicknameStruct `json:"nicknames"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    User         : user,
    Nicknames: nicknames,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

