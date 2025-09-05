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
  // "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func ReceptionGet(w http.ResponseWriter, r *http.Request) {

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")
  receptionID := r.FormValue("receptionID")
  passkey := r.FormValue("passkey")

  // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  // defer cancel()
  // c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  // if err != nil {
  //   log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  // }
  // defer c.Disconnect(ctx)
  // db1 := c.Database(common.MongoDb1)

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheckTake: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  staffAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      staffAccess = true
    }
  }

  var reception collection.ReceptionStruct
  responseData := struct {
    Csrf         string        `json:"csrf"`
    PushContents []string `json:"pushContents"`
    Reception collection.ReceptionStruct `json:"reception"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Reception: reception,
  }

	// coll := db1.Collection("reception")
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
	coll := common.DB.ReceptionDB.Collection("reception")
	filter := bson.M{"_id": receptionID}
	err = coll.FindOne(ctx, filter).Decode(&reception)
	if err != nil {
		log.Print(err, " reception ", receptionID)
		// http.Error(w, "reception not found", http.StatusNotFound)
	  w.Header().Set("Content-Type", "application/json")
	  json.NewEncoder(w).Encode(responseData)
		return
	}

  now := time.Now()
  valid := false
  for _, pc := range reception.Passcodes {
  	// log.Printf("Invalid or expired passkey: Req: ", pc)
    if pc.Passkey == passkey {
      startTime, err1 := time.Parse("2006-01-02T15:04", pc.PassStart)
      endTime, err2 := time.Parse("2006-01-02T15:04", pc.PassEnd)
      if err1 == nil && err2 == nil && now.After(startTime) && now.Before(endTime) {
        valid = true
        break
      }
    }
  }

  if !valid && !staffAccess {
	  if !staffAccess {
	    log.Printf("ChannelAliases !staffAccess: %v; Req: ", session.ChannelAliases, r.URL.Path, r.Form)
	  }
	  if !valid {
	  	log.Printf("Invalid or expired passkey: Req: ", r.URL.Path, r.Form)
	  }
    http.Error(w, "invalid or expired passkey or not staff", http.StatusForbidden)
    return
  }

	if !staffAccess {
		reception.ChannelID = ""
	  reception.AdminNames = nil
	  reception.Passcodes = nil
	  reception.JoinNames = nil
	  // reception.Subscriptions = nil
	  for i := range reception.Seats {
	    reception.Seats[i].Passcodes = nil
	  }
	}
  // reception の値だけここでセット
  responseData.Reception = reception
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
