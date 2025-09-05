package controller

import (
  "context"
  "encoding/json"
  "log"
  "net/http"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"
)

func ChannelDelete(w http.ResponseWriter, r *http.Request) {
	var userIDs []string
  if err := json.Unmarshal([]byte(r.FormValue("userIDs")), &userIDs); err != nil {
  	log.Printf("userIDs: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON userIDs", http.StatusBadRequest)
    return
  }

  var deleteAliases []collection.Alias
  if err := json.Unmarshal([]byte(r.FormValue("deleteAliases")), &deleteAliases); err != nil {
    log.Printf("deleteAliases: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON deleteAliases", http.StatusBadRequest)
    return
  }

  updatedBy := r.FormValue("updatedBy")
  channelID := r.FormValue("channelID")

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
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

  trueAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == updatedBy && d.ChannelID == channelID {
      trueAccess = true
    }
  }

  if !trueAccess {
    log.Printf("ChannelAliases !trueAccess: %v; Req: ", session.ChannelAliases, updatedBy, channelID, r.URL.Path, r.Form)
    return
  }

  // coll := db1.Collection("session")
  coll := common.DB.SessionDB.Collection("session")
  filter := bson.D{{"userID", bson.D{{"$in", userIDs}}}}
  cursor, err := coll.Find(context.TODO(), filter)
  if err != nil {
    log.Printf("coll.Find: %v; Req: ", err, userIDs, r.URL.Path, r.Form)
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
    log.Printf("cursor.All: %v; Req: ", err, userIDs, r.URL.Path, r.Form)
  }

	sessionCol := common.DB.SessionDB.Collection("session")
	userCol := common.DB.UserDB.Collection("user")

	for _, del := range deleteAliases {
		update := bson.M{
			"$pull": bson.M{"channelAliases": bson.M{"channelID": channelID, "alias": del.AliasName}},
		}
		sessionFilter := bson.M{"userID": del.UserID}
		_, err = sessionCol.UpdateMany(ctx, sessionFilter, update)
		if err != nil {
			log.Printf("Failed to update session: %v", err, del.UserID, r.URL.Path, r.Form)
		}
		userFilter := bson.M{"_id": del.UserID}
		_, err = userCol.UpdateMany(ctx, userFilter, update)
		if err != nil {
			log.Printf("Failed to update user: %v", err, del.UserID, r.URL.Path, r.Form)
		}

		aliasData := []string{del.UserID, del.AliasName, "", "delete"}
	  var arr []interface{}
		arr = append(arr, "alias")
		arr = append(arr, channelID)
		arr = append(arr, updatedBy)
		arr = append(arr, aliasData)
		arr = append(arr, del.AliasImg)
		common.ChunkPush(sessions, arr)

	}

  // session, err = common.ReGenerateData(db1, session)
  // if err != nil {
  //   log.Printf("ReGenerateData: %v; Req:", err, r.URL.Path, r.Form)
  // }

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

