package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  // "io"
  "log"
  "net/http"
  // "os"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  // webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func ChannelAdd(w http.ResponseWriter, r *http.Request) {

	myname := r.FormValue("myname")
	myimg := r.FormValue("myimg")
	channelName := r.FormValue("channelName")
	channelDescription := r.FormValue("channelDescription")

  // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  // defer cancel()
  // c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  // if err != nil {
  //   log.Printf("mongo.Connect: %v; Req:", err, r.URL.Path, r.Form)
  // }
  // defer c.Disconnect(ctx)
  // db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheckTake: %v; Req:", err, r.URL.Path, r.Form)
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
    return
  }

	// --- チャンネル上限チェック ---
	uniqueChannels := make(map[string]struct{})
	for _, alias := range user.ChannelAliases {
	  uniqueChannels[alias.ChannelID] = struct{}{}
	}

	if len(uniqueChannels) >= 3 {
		log.Printf("len(uniqueChannels) >= 3: %v; Req:", err, r.URL.Path, r.Form)
    common.WriteResponseWithSession(w, session, "すでにチャネル作成の上限です", http.StatusOK)
    return
	}

  channelID, err := common.CountUpID("channelID")
  if err != nil {
    log.Printf("CountUpID error: %v", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
	aliasImg, err := common.ImgSave(myimg, session.UserID, myname, channelID, 3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// coll := db1.Collection("session")
	coll := common.DB.SessionDB.Collection("session")
  filter := bson.M{"userID": session.UserID}
	project := bson.D{{"updatedAt", 0}}
	opts4 := options.Find().SetProjection(project)
	cursor, err := coll.Find(context.TODO(), filter, opts4)
	if err != nil {
	  log.Printf("coll.Find: %v; Req:", err, r.URL.Path, r.Form)
	}
	var mySessions []collection.SessionStruct
	if err = cursor.All(context.TODO(), &mySessions); err != nil {
	  log.Printf("cursor.All: %v; Req:", err, r.URL.Path, r.Form)
	}
	contents := []string{channelName, channelDescription}
  var arr []interface{}
  arr = append(arr, "channelEdit")
  arr = append(arr, channelID)
  arr = append(arr, myname)
  arr = append(arr, contents)
	common.ChunkPush(mySessions, arr)

  contents = []string{session.UserID, myname, "", "admin"}
  newAliasChannel := collection.ChannelAlias{
		ChannelID: channelID,
		Alias:     myname,
	}
  for _, d := range mySessions { // go to alias
	  pushID := common.StringRand(1)
		var arr []interface{}
		arr = append(arr, pushID)
		arr = append(arr, "alias")
		arr = append(arr, channelID)
		arr = append(arr, myname)
		arr = append(arr, contents)
		arr = append(arr, aliasImg)
		// cursor.Decode(&d)
    resp, err := common.SendWebPushNotification(arr, pushID, d)
		if err != nil {
	    log.Printf("resp SendWebPushNotification: %v; Req:", err, r.URL.Path, r.Form)
		}
		defer resp.Body.Close()
		// d.ChannelAliases = append(d.ChannelAliases, newAliasChannel)
		// d.UpdatedAt = time.Now()
		// coll := db1.Collection("session")
		coll := common.DB.SessionDB.Collection("session")
		filter := bson.D{{"_id", d.SessionID}}
		// update := bson.D{{"$set", d}}
		update := bson.M{
	    "$set": bson.M{
        "channelAliases": append(d.ChannelAliases, newAliasChannel),
        "updatedAt":      time.Now(),
	    },
		}
		_, err = coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
		  log.Printf("coll.UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
		}
  }

 //  user.ChannelAliases = append(user.ChannelAliases, newAliasChannel)
 //  user.UpdatedAt = time.Now()
	// userUpdate := bson.D{{"$set", user}}
	userUpdate := bson.M{
	    "$set": bson.M{
	        "channelAliases": append(user.ChannelAliases, newAliasChannel),
	        "updatedAt":      time.Now(),
	    },
	}
	_, err = collUser.UpdateOne(context.TODO(), filterUser, userUpdate)
	if err != nil {
	  log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
	}

	session, err = common.PushReGenerate(session)
	if err != nil {
		log.Printf("ReGenerateData: %v; Req:", err, r.URL.Path, r.Form)
	}
	responseData := struct {
		ChannelID    string    `json:"channelID"`
		Csrf         string    `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		ChannelID:    channelID,
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
