package controller

import (
	"context"
	"encoding/json"
  // "fmt"
  "log"
  "net/http"
  "time"

	// "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/common"
	"chat/collection"
)

func PushSubscribe(w http.ResponseWriter, r *http.Request) {
	if !json.Valid([]byte(r.FormValue("subscription"))) {
		log.Printf("Invalid JSON subscription: %s; Req:", r.URL.Path, r.Form)
		http.Error(w, "Invalid JSON subscription", http.StatusBadRequest)
		return
	}

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
  var nickname collection.NicknameStruct
  if session.Nickname == "" {
    err = common.DB.NicknameDB.Collection("nickname").
        FindOne(ctx, bson.M{"userID": session.UserID}).Decode(&nickname)
    if err != nil {
      log.Printf("Nickname not found for userID=%s: %v", session.UserID, err)
    }
    session.Nickname = nickname.Nickname
    session.NickImg = nickname.NickImg
  }

	coll := common.DB.SessionDB.Collection("session")
	filter := bson.D{{"_id", session.SessionID}}
	update := bson.D{{"$set", bson.D{
		{"subscription", r.FormValue("subscription")},
		{"nickname", session.Nickname},
		{"nickImg", session.NickImg},
		{"updatedAt", time.Now()}}}}
	opts := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts)

	session.Subscription = r.FormValue("subscription")
  var arr []interface{}
  arr = append(arr, "pushCheck")
  arr = append(arr, "push登録完了")
  var sessions []collection.SessionStruct
  sessions = append(sessions, session)
	common.ChunkPush(sessions, arr)

  responseData := struct {
    Csrf         string        `json:"csrf"`
    PushContents []string `json:"pushContents"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)

}
