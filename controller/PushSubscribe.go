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

	// channelIDs 取得
	channelIDSet := make(map[string]struct{})
	var channelIDs []string
	for _, ca := range session.ChannelAliases {
		if _, ok := channelIDSet[ca.ChannelID]; !ok {
			channelIDSet[ca.ChannelID] = struct{}{}
			channelIDs = append(channelIDs, ca.ChannelID)
		}
	}

	// channel 取得
	collChannel := common.DB.ChannelDB.Collection("channel")
	var channels []collection.ChannelStruct
	if len(channelIDs) > 0 {
		cursor, err := collChannel.Find(ctx, bson.M{"_id": bson.M{"$in": channelIDs}})
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
			return
		}
		if err := cursor.All(ctx, &channels); err != nil {
			common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
			return
		}
	}

	// sanitize
	safeChannels := make([]collection.ChannelStruct, 0, len(channels))
	for _, ch := range channels {
		safeChannels = append(safeChannels, sanitizeChannel(ch))
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

	// response
	responseData := struct {
		Csrf         string                     `json:"csrf"`
		PushContents []string                   `json:"pushContents"`
		Channels     []collection.ChannelStruct `json:"channels"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Channels:     safeChannels,
	}
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)

}

// func sanitizeChannel(channel collection.ChannelStruct) collection.ChannelStruct {
// 	safe := channel
// 	safeAliases := make([]collection.Alias, 0, len(channel.Aliases))
// 	for _, a := range channel.Aliases {
// 		a.UserID = "" // ★ UserIDはレスポンスでは返さない
// 		safeAliases = append(safeAliases, a)
// 	}
// 	safe.Aliases = safeAliases
// 	return safe
// }

