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

  "go.mongodb.org/mongo-driver/mongo"
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

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect: %v; Req:", err, r.URL.Path, r.Form)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck: %v; Req:", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}
  channelID := common.StringRand(4)
	aliasImg, err := common.ImgSave(db1, myimg, session.UserID, myname, channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	coll := db1.Collection("session")
  filter := bson.M{"user_id": session.UserID}
	project := bson.D{{"updated_at", 0}}
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
	common.ChunkPush(mySessions, db1, arr)

  contents = []string{session.UserID, myname}
  newAliasChannel := collection.ChannelAlias{
		ChannelID: channelID,
		Alias:     myname,
	}
  for _, d := range mySessions { // go to alias
	  pushID := common.StringRand(12)
		var arr []interface{}
		arr = append(arr, pushID)
		arr = append(arr, "alias")
		arr = append(arr, channelID)
		arr = append(arr, myname)
		arr = append(arr, contents)
		arr = append(arr, aliasImg)
		// cursor.Decode(&d)
    resp, err := common.SendWebPushNotification(db1, arr, pushID, d.Subscription)
		if err != nil {
	    log.Printf("resp SendWebPushNotification: %v; Req:", err, r.URL.Path, r.Form)
		}
		defer resp.Body.Close()
		d.ChannelAliases = append(d.ChannelAliases, newAliasChannel)
		d.UpdatedAt = time.Now()
		coll := db1.Collection("session")
		filter := bson.D{{"_id", d.SessionID}}
		update := bson.D{{"$set", d}}
		_, err = coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
		  log.Printf("coll.UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
		}
  }

  collUser := db1.Collection("user")
  filterUser := bson.M{"_id": session.UserID}
  var user collection.UserStruct
  err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
  if err != nil {
    log.Printf("user FindOne: %v; Req:", err, r.URL.Path, r.Form)
  }
  user.ChannelAliases = append(user.ChannelAliases, newAliasChannel)
  user.UpdatedAt = time.Now()
	userUpdate := bson.D{{"$set", user}}
	_, err = collUser.UpdateOne(context.TODO(), filterUser, userUpdate)
	if err != nil {
	  log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
	}

  response := map[string]string{"channelID": channelID}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "JSONエンコードエラー", http.StatusInternalServerError)
	}
}
