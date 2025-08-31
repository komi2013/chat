package controller

import (
  "context"
  "encoding/json"
  // "fmt"
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

	var aliases []collection.Alias
  if err := json.Unmarshal([]byte(r.FormValue("aliases")), &aliases); err != nil {
  	log.Printf("aliases: %v; Request:", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON aliases", http.StatusBadRequest)
    return
  }

	// var groups []collection.Group
 //  if err := json.Unmarshal([]byte(r.FormValue("groups")), &groups); err != nil {
 //  	log.Printf("groups: %v; Request:", err, r.URL.Path, r.Form)
 //    http.Error(w, "Invalid JSON groups", http.StatusBadRequest)
 //    return
 //  }

	var aliasNames []string
  if err := json.Unmarshal([]byte(r.FormValue("aliasNames")), &aliasNames); err != nil {
  	log.Printf("aliasNames: %v; Request:", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON aliasNames", http.StatusBadRequest)
    return
  }

	untilDate, err := time.Parse("2006-01-02", r.FormValue("untilDate"))
	if err != nil {
		log.Printf("Invalid untilDate: %v; Request:", err, r.URL.Path, r.Form)
		http.Error(w, "Invalid untilDate: must be in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}

  channelID := r.FormValue("channelID")
  channelName := r.FormValue("channelName")
  channelDescription := r.FormValue("channelDescription")
  updatedBy := r.FormValue("updatedBy")
  guest := r.FormValue("guest") != ""

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect: %v; Request:", err, r.URL.Path, r.Form)
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
  for _, d := range session.ChannelAliases {
    if d.Alias == updatedBy && d.ChannelID == channelID && !d.Guest {
      trueAccess = true
    }
  }

  if !trueAccess {
    log.Printf("ChannelAliases !trueAccess: %v; Request:", session.ChannelAliases, updatedBy, channelID, r.URL.Path, r.Form)
    common.WriteResponseWithSession(w, session, "ChannelAliases !trueAccess", http.StatusOK)
    return
  }
  // var subscriptions []string
  var addSessions []collection.SessionStruct
  coll := db1.Collection("session")
  filter := bson.D{{
    "userID", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"createdAt", 0}}
  opts4 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts4)
  if err != nil {
    log.Printf("Find: %v; Request:", err, r.URL.Path, r.Form)
    common.WriteResponseWithSession(w, session, "ChannelAliases !trueAccess", http.StatusOK)
    return
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
    log.Printf("sessions: %v; Request:", err, r.URL.Path, r.Form)
  }
	for _, s := range sessions {
	  trueAccess := false
	  for _, d := range s.ChannelAliases {
	    if d.Alias == updatedBy && d.ChannelID == channelID {
	      trueAccess = true
	    }
	  }
	  if trueAccess {
	  	// subscriptions = append(subscriptions, s.Subscription)
	    addSessions = append(addSessions, collection.SessionStruct{
	        SessionID:    s.SessionID,
	        Subscription: s.Subscription,
	    })
	  }
	}

	coll = db1.Collection("channel")
	now := time.Now()
	filterInvitation := bson.M{
    "$or": []interface{}{
      bson.M{"channelID": channelID},
      bson.M{"untilDate": bson.M{"$lt": now}},
    },
	}

	_, err = coll.DeleteMany(context.TODO(), filterInvitation)
	if err != nil {
    log.Printf("DeleteMany: %v; Request:", err, r.URL.Path, r.Form)
	}
  invitation := collection.ChannelStruct{
  	InvitationCode: common.StringRand(16),
		ChannelID: channelID,
		ChannelName: channelName,
		ChannelDescription: channelDescription,
		CreatedBy: updatedBy,
    CreatedAt: time.Now(),
		PushSessions: addSessions,
		AliasNames: aliasNames,
		Aliases: aliases,
		Guest: guest,
		UntilDate: untilDate,
	}
	_, err = coll.InsertOne(context.TODO(), invitation)
	if err != nil {
		log.Printf("InsertOne: %v; Request:", err, r.URL.Path, r.Form)
	}

	session, err = common.ReGenerateData(db1, session)
	if err != nil {
		log.Printf("ReGenerateData: %v; Req:", err, r.URL.Path, r.Form)
	}

	responseData := struct {
		InvitationCode    string        `json:"invitationCode"`
		Csrf         string        `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		InvitationCode:    invitation.InvitationCode,
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
