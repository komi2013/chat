package controller

import (
  "context"
  "encoding/json"
  "fmt"
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
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck: %v; Request: %v", err, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}
  channelID := common.StringRand(4)
	aliasImg := common.AliasImgSave(myimg, session.UserID, myname)
  var userIDs = []string{}
  userIDs = append(userIDs, session.UserID)

	coll := db1.Collection("session")
  filter := bson.D{{
  	"user_id", bson.D{{"$in", userIDs}}}}
	project := bson.D{{"updated_at", 0}}
	opts4 := options.Find().SetProjection(project)
	cursor, err := coll.Find(context.TODO(), filter, opts4)
	if err != nil {
	    fmt.Printf("err %s\n", err)
	}
	var sessions []collection.SessionStruct
	if err = cursor.All(context.TODO(), &sessions); err != nil {
	    fmt.Printf("err %s\n", err)
	}
	contents := []string{channelName, channelDescription}
  for _, d := range sessions { // go to channelEdit
	  pushID := common.StringRand(12)
		var arr []interface{}
		arr = append(arr, pushID)
		arr = append(arr, "channelEdit")
		arr = append(arr, channelID)
		arr = append(arr, myname)
		arr = append(arr, contents)
    cursor.Decode(&d)
    resp, err := common.SendWebPushNotification(db1, arr, pushID, d.Subscription)
		if err != nil {
	    fmt.Printf(" err %s\n", err)
		}
		defer resp.Body.Close()
  }
  contents = []string{aliasImg, session.UserID}
  newAliasChannel := collection.AliasChannel{
		ChannelID: channelID,
		Alias:     myname,
	}
  for _, d := range sessions { // go to alias
	  pushID := common.StringRand(12)
		var arr []interface{}
		arr = append(arr, pushID)
		arr = append(arr, "alias")
		arr = append(arr, channelID)
		arr = append(arr, myname)
		arr = append(arr, contents)
		cursor.Decode(&d)
    resp, err := common.SendWebPushNotification(db1, arr, pushID, d.Subscription)
		if err != nil {
	    fmt.Printf(" err %s\n", err)
		}
		defer resp.Body.Close()
		log.Printf("d: %v", d)
		d.AliasChannels = append(d.AliasChannels, newAliasChannel)
		coll := db1.Collection("session")
		filter := bson.D{{"_id", d.SessionID}}
		update := bson.D{{"$set", bson.D{
			{"user_id", d.UserID},
			{"updated_at", time.Now()},
			{"alias_channels", d.AliasChannels},
		}}}
		// opts := options.Update().SetUpsert(true)
		_, err = coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
		    fmt.Printf("UpdateOne %s\n", err)
		}
  }
  response := map[string]string{"channelID": channelID}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "JSONエンコードエラー", http.StatusInternalServerError)
	}
}
