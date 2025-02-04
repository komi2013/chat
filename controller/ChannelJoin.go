package controller

import (
  "context"
  // "encoding/base64"
  "encoding/json"
  // "fmt"
  // "io/ioutil"
  "log"
  "net/http"
  // "os"
  // "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  // "go.mongodb.org/mongo-driver/bson/primitive"

  // webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func ChannelJoin (w http.ResponseWriter, r *http.Request) {

  channelID := r.FormValue("channelID")
  myname := r.FormValue("myname")
  myimg := r.FormValue("myimg")
  code := r.FormValue("code")

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

	aliasImg, err := common.ImgSave(db1, myimg, session.UserID, myname, channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newAlias := collection.Alias{
	  AliasName:  myname,
	  AliasImg:   myimg,
	  UserID:     session.UserID,
	}

	coll := db1.Collection("invitation")
	filter := bson.M{"_id": code}
	var invitation collection.InvitationStruct
	err = coll.FindOne(context.TODO(), filter).Decode(&invitation)
	if err != nil {
		log.Printf("invitation FindOne: %v; Req:", err, r.URL.Path, r.Form)
	}
	if invitation.ChannelID != channelID {
		log.Printf("invitation.ChannelID != channelID:; Req:", r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}

	nameFound := false
	for _, name := range invitation.AliasNames {
		if name == myname {
			nameFound = true
			break
		}
	}

	if nameFound {
	    collUser := db1.Collection("user")
	    filterUser := bson.M{"_id": session.UserID}
	    var user collection.UserStruct
	    err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
	    if err != nil {
	        log.Printf("user FindOne: %v; Req:", err, r.URL.Path, r.Form)
	    }
	    userFound := false
	    for _, ca := range user.ChannelAliases {
	        if ca.ChannelID == channelID && ca.Alias == myname {
	            userFound = true
	            break
	        }
	    }
			if !userFound {
			    w.Header().Set("Content-Type", "application/json")
			    w.WriteHeader(http.StatusConflict) // 409 Conflict
			    json.NewEncoder(w).Encode(map[string]string{
			        "error": "すでに同じ名前が存在しています。",
			    })
			    return
			}
	}

	_, err = coll.UpdateOne(context.TODO(), filter,
    bson.M{
        "$push": bson.M{
            "aliases": newAlias,
            "alias_names": myname,
        },
    },
	)
	if err != nil {
	  log.Printf("coll.UpdateOne push aliases alias_names: %v; Req:", err, r.URL.Path, r.Form)
	}

	coll = db1.Collection("session")
  filter = bson.M{"user_id": session.UserID}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
	  log.Printf("coll.Find: %v; Req:", err, r.URL.Path, r.Form)
	}
	var mySessions []collection.SessionStruct
	if err = cursor.All(context.TODO(), &mySessions); err != nil {
	  log.Printf("cursor.All: %v; Req:", err, r.URL.Path, r.Form)
	}

	contents := []string{invitation.ChannelName, invitation.ChannelDescription}
  var arr []interface{}
  arr = append(arr, "channelEdit")
  arr = append(arr, channelID)
  arr = append(arr, myname)
  arr = append(arr, contents)
	common.ChunkPush(mySessions, db1, arr)

	for _, d := range invitation.Aliases {
		aliasData := []string{d.UserID, d.AliasName}
	  var arr []interface{}
		arr = append(arr, "alias")
		arr = append(arr, channelID)
		arr = append(arr, myname)
		arr = append(arr, aliasData)
		arr = append(arr, d.AliasImg)
		log.Printf("invitation.Aliases: %v; Req:", err, r.URL.Path, r.Form)
		common.ChunkPush(mySessions, db1, arr)
	}

	// for _, d := range invitation.Groups {
	// 	groupData := []interface{}{d.GroupName, d.AliasNames}
	//   var arr []interface{}
	// 	arr = append(arr, "group")
	// 	arr = append(arr, channelID)
	// 	arr = append(arr, myname)
	// 	arr = append(arr, groupData)
	// 	arr = append(arr, d.GroupImg)
	// 	log.Printf("invitation.Aliases: %v; Req:", err, r.URL.Path, r.Form)
	// 	common.ChunkPush(mySessions, db1, arr)
	// }

  newAliasChannel := collection.ChannelAlias{
		ChannelID: channelID,
		Alias:     myname,
	}
  for _, d := range mySessions {
  	invitation.Subscriptions = append(invitation.Subscriptions, d.Subscription)
		d.ChannelAliases = append(d.ChannelAliases, newAliasChannel)
		d.UpdatedAt = time.Now()
		coll := db1.Collection("session")
		filter := bson.D{{"_id", d.SessionID}}
		update := bson.D{{"$set", d}}
		_, err = coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
		  log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
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

	contents = []string{session.UserID, myname}
  for _, sess := range invitation.Sessions {
	  pushID := common.StringRand(12)
		var arr []interface{}
		arr = append(arr, pushID)
		arr = append(arr, "alias")
		arr = append(arr, channelID)
		arr = append(arr, myname)
		arr = append(arr, contents)
		arr = append(arr, aliasImg)
    resp, err := common.SendWebPushNotification(db1, arr, pushID, sess)
		if err != nil {
	    log.Printf("resp SendWebPushNotification: %v; Req: ", err, r.URL.Path, r.Form)
		}
		defer resp.Body.Close()
  }
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
