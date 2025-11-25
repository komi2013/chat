package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  // "log"
  // "math"
  "net/http"
  // "time"

  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/mongo/options"
)

func UserGet(w http.ResponseWriter, r *http.Request) {
  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
  	common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";Session Check", http.StatusOK)
    return
  }

  collUser := common.DB.UserDB.Collection("user")
  filterUser := bson.M{"_id": session.UserID}
  var user collection.UserResponse
  err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";collUser.Find", http.StatusOK)
    return
  }
  collNickname := common.DB.NicknameDB.Collection("nickname")
  filterNickname := bson.M{"userID": session.UserID}
  cursor, err := collNickname.Find(context.TODO(), filterNickname)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";collNickname.Find", http.StatusOK)
    return
  }
  var nicknames []collection.NicknameResponse
  if err = cursor.All(context.TODO(), &nicknames); err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";cursor.All", http.StatusOK)
    return
  }

  responseData := struct {
    Csrf         string       `json:"csrf"`
    PushContents []string     `json:"pushContents"`
    User       collection.UserResponse  `json:"user"`
    Nicknames   []collection.NicknameResponse `json:"nicknames"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    User         : user,
    Nicknames: nicknames,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

