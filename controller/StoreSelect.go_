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

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func StoreSelect(w http.ResponseWriter, r *http.Request) {
	var userIDs []string
  if err := json.Unmarshal([]byte(r.FormValue("userIDs")), &userIDs); err != nil {
  	log.Printf("userIDs: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON userIDs", http.StatusBadRequest)
    return
  }

  var param interface{}
  if err := json.Unmarshal([]byte(r.FormValue("param")), &param); err != nil {
  	log.Printf("param: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON param", http.StatusBadRequest)
    return
  }

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")
  targetStore := r.FormValue("targetStore")

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck: %v; Req: ", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}

  trueAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    log.Printf("ChannelAliases !trueAccess: %v; Req: ", session.ChannelAliases, aliasName, channelID, r.URL.Path, r.Form)
    return
  }

  coll := db1.Collection("session")
  filter := bson.D{{"user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"subscription", 1},{"user_id", 1}}
  opts := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts)
  if err != nil {
    log.Printf("coll.Find session: %v", err, userIDs, r.URL.Path, r.Form)
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
  	log.Printf("cursor.All sessions: %v", err, userIDs, r.URL.Path, r.Form)
  }
	for _, ssData := range sessions {
		pushID := common.StringRand(1)
		var arr []interface{}
		arr = append(arr, pushID)
		arr = append(arr, "storeSelect")
		arr = append(arr, channelID)
		arr = append(arr, aliasName)
		arr = append(arr, targetStore)
		arr = append(arr, param)
		jsonD, err := json.Marshal(arr)
		if err != nil {
			log.Printf("json.Marshal storeSelect: %v", err, userIDs, r.URL.Path, r.Form)
			continue
		}
		webpushSub := &webpush.Subscription{}
		err = json.Unmarshal([]byte(ssData.Subscription), webpushSub)
		if err != nil {
			log.Printf("json.Marshal storeSelect: %v", err, userIDs, r.URL.Path, r.Form)
			continue
		}

		resp, err := webpush.SendNotification([]byte(string(jsonD)), webpushSub, &webpush.Options{
			Subscriber:      "example@example.com",
			VAPIDPublicKey:  common.VAPIDPublicKey,
			VAPIDPrivateKey: common.VAPIDPrivateKey,
			TTL:             30,
		})
		if err != nil {
			log.Printf("SendNotification storeSelect: %v", err, string(jsonD), r.URL.Path, r.Form)
			continue
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
