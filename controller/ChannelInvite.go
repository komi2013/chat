package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/common"
  "chat/collection"
)

func ChannelInvite(w http.ResponseWriter, r *http.Request) {
	var userIDs []string
  if err := json.Unmarshal([]byte(r.FormValue("userIDs")), &userIDs); err != nil {
  	log.Printf("userIDs: %v; Request:", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON userIDs", http.StatusBadRequest)
    return
  }

	var aliasNames []string
  if err := json.Unmarshal([]byte(r.FormValue("aliasNames")), &aliasNames); err != nil {
  	log.Printf("aliasNames: %v; Request:", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON aliasNames", http.StatusBadRequest)
    return
  }

  channelID := r.FormValue("channelID")
  channelName := r.FormValue("channelName")
  channelDescription := r.FormValue("channelDescription")
  updatedBy := r.FormValue("updatedBy")

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
		log.Printf("SessionCheck: %v; Request:", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}

  trueAccess := false
  for _, d := range session.AliasChannels {
    if d.Alias == updatedBy && d.ChannelID == channelID {
      trueAccess = true
    }
  }

  if !trueAccess {
    log.Printf("AliasChannels !trueAccess: %v; Request:", session.AliasChannels, updatedBy, channelID, r.URL.Path, r.Form)
    return
  }
  var subscriptions []string
  coll := db1.Collection("session")
  filter := bson.D{{
    "user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"created_at", 0}}
  opts4 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts4)
  if err != nil {
    log.Printf("Find: %v; Request:", err, r.URL.Path, r.Form)
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
    log.Printf("sessions: %v; Request:%v;%v", err, r.URL.Path, r.Form)
  }
	for _, s := range sessions {
	  trueAccess := false
	  for _, d := range s.AliasChannels {
	    if d.Alias == updatedBy && d.ChannelID == channelID {
	      trueAccess = true
	    }
	  }
	  if trueAccess {
	  	subscriptions = append(subscriptions, s.Subscription)
	  }
	}
	coll = db1.Collection("invitation")
  invitation := collection.InvitationStruct{
  	InvitationCode: common.StringRand(16),
		ChannelID: channelID,
		ChannelName: channelName,
		ChannelDescription: channelDescription,
		CreatedBy: updatedBy,
    CreatedAt: time.Now(),
		Subscriptions: subscriptions,
		AliasNames: aliasNames,
	}
	_, err = coll.InsertOne(context.TODO(), invitation)
	if err != nil {
		log.Printf("InsertOne: %v; Request:", err, r.URL.Path, r.Form)
	}

	var jsonArr []interface{}
	jsonArr = append(jsonArr, invitation.InvitationCode)
	jsonData, _ := json.Marshal(jsonArr)
  fmt.Fprint(w, string(jsonData))
}
