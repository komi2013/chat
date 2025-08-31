package controller

import (
  "context"
  // "encoding/base64"
  // "encoding/json"
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
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck: %v; Req:", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}

	collInvitation := db1.Collection("channel")
	invitationFilter := bson.M{"_id": code}
	var invitation collection.ChannelStruct
	err = collInvitation.FindOne(context.TODO(), invitationFilter).Decode(&invitation)
	if err != nil {
		log.Printf("invitation FindOne: %v; Req:", err, r.URL.Path, r.Form)
  	common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
	}
	if invitation.ChannelID != channelID {
		log.Printf("invitation.ChannelID != channelID Req:", r.URL.Path, r.Form)
  	common.WriteResponseWithSession(w, session, "invitation.ChannelID != channelID", http.StatusOK)
    return
	}

	nameFound := false
	for _, name := range invitation.AliasNames {
		if name == myname {
			nameFound = true
			break
		}
	}

  collUser := db1.Collection("user")
  filterUser := bson.M{"_id": session.UserID}
  var user collection.UserStruct
  err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
  if err != nil {
    log.Printf("user FindOne: %v; Req:", err, r.URL.Path, r.Form)
  	// http.Error(w, err.Error(), http.StatusServiceUnavailable)
  	common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
  }
  userFound := false
  for _, ca := range user.ChannelAliases {
    if ca.ChannelID == channelID {
      userFound = true
      // myname = ca.Alias
      break
    }
  }

	if nameFound && !userFound {
		log.Printf("nameFound && !userFound: %v; Req:", err, r.URL.Path, r.Form)
    common.WriteResponseWithSession(w, session, "すでに同じ名前が存在しています", http.StatusOK)
    return
	}
	var accessRight string
	if invitation.Guest {
		accessRight = "guest"
	}

	coll := db1.Collection("session")
  filter := bson.M{"userID": session.UserID}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
	  log.Printf("coll.Find: %v; Req:", err, r.URL.Path, r.Form)
  	common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
	}
	var mySessions []collection.SessionStruct
	if err = cursor.All(context.TODO(), &mySessions); err != nil {
	  log.Printf("cursor.All: %v; Req:", err, r.URL.Path, r.Form)
  	common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
	}

// ここから

	aliasImg, err := common.ImgSave(db1, myimg, session.UserID, myname, channelID, 3)
	if err != nil {
		log.Printf("ImgSave: %v; Req:", err, r.URL.Path, r.Form)
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	newAlias := collection.Alias{
	  AliasName:  myname,
	  AliasImg:   aliasImg,
	  UserID:     session.UserID,
	  AccessRight: accessRight,
	}

	contents := []string{invitation.ChannelName, invitation.ChannelDescription}
  var arr []interface{}
  arr = append(arr, "channelEdit")
  arr = append(arr, channelID)
  arr = append(arr, myname)
  arr = append(arr, contents)
	common.ChunkPush(mySessions, db1, arr)

	invitation.Aliases = append(invitation.Aliases, newAlias)

	for _, d := range invitation.Aliases {
		aliasData := []string{d.UserID, d.AliasName, d.Bio, d.AccessRight}
	  var arr []interface{}
		arr = append(arr, "alias")
		arr = append(arr, channelID)
		arr = append(arr, myname)
		arr = append(arr, aliasData)
		arr = append(arr, d.AliasImg)
		common.ChunkPush(mySessions, db1, arr)
	}

	contents = []string{session.UserID, myname, "", accessRight}
  for _, sess := range invitation.PushSessions {
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
	  	common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
	    return
		}
		defer resp.Body.Close()
  }

  newAliasChannel := collection.ChannelAlias{
		ChannelID: channelID,
		Alias:     myname,
		Guest:		invitation.Guest,
	}
	var addSessions []collection.SessionStruct
  for _, d := range mySessions {
  	if !userFound {
			d.ChannelAliases = append(d.ChannelAliases, newAliasChannel)
  	}
		d.UpdatedAt = time.Now()
		coll := db1.Collection("session")
		filter := bson.D{{"_id", d.SessionID}}
		update := bson.D{{"$set", d}}
		_, err = coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
		  log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
	  	common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
	    return
		}
    addSessions = append(addSessions, collection.SessionStruct{
      SessionID:    d.SessionID,
      Subscription: d.Subscription,
    })
  }

	if !userFound {
		user.ChannelAliases = append(user.ChannelAliases, newAliasChannel)
	}

  user.UpdatedAt = time.Now()
	userUpdate := bson.D{{"$set", user}}
	_, err = collUser.UpdateOne(context.TODO(), filterUser, userUpdate)
	if err != nil {
	  log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
  	common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
	}
	var pushUpd bson.M
	// if userFound {
	// 	pushUpd = bson.M{
	// 	    "$push": bson.M{
	//         "pushSessions": bson.M{"$each": addSessions},
	// 	    },
	//     }
	// } else {
	// 	pushUpd = bson.M{
	// 	    "$push": bson.M{
	//         "aliases": newAlias,
	//         "aliasNames": myname,
	//         "pushSessions": bson.M{"$each": addSessions},
	// 	    },
	//     }
	// }
	pushUpd = bson.M{
	    "$push": bson.M{
        "aliases": newAlias,
        "aliasNames": myname,
        "pushSessions": bson.M{"$each": addSessions},
	    },
    }
	_, err = collInvitation.UpdateOne(context.TODO(), invitationFilter, pushUpd)
	if err != nil {
	  log.Printf("collInvitation.UpdateOne pushUpd: %v; Req:", err, r.URL.Path, r.Form)
  	common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
	}
// ここまで
  common.WriteResponseWithSession(w, session, "", http.StatusOK)

}
