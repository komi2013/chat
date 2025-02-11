package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  "log"
  "net/http"
  "time"
	"unicode/utf8"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"

)

func StorePush(w http.ResponseWriter, r *http.Request) {
	var userIDs []string
  if err := json.Unmarshal([]byte(r.FormValue("userIDs")), &userIDs); err != nil {
  	log.Printf("userIDs: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON userIDs", http.StatusBadRequest)
    return
  }
  var contents interface{}
  if err := json.Unmarshal([]byte(r.FormValue("contents")), &contents); err != nil {
  	log.Printf("contents: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON contents", http.StatusBadRequest)
    return
  }
  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")

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
	log.Printf("trueAccess dayo: %v; Req: ", nil, r.URL.Path, r.Form)
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

	// coll := db1.Collection("push_select")
	// filter := bson.D{{"_id", r.FormValue("pushSelectID")}}
	// result, err := coll.DeleteOne(context.TODO(), filter)
	// if err != nil {
	// 	log.Printf("coll.DeleteOne session: %v", err, userIDs, r.URL.Path, r.Form)
	// 	return
	// }

	// deletedCount := result.DeletedCount
	// fmt.Printf("Deleted %d document(s)\n", deletedCount)
	// if deletedCount == 0 {
	// 	return
	// }

  coll := db1.Collection("session")
  filter := bson.D{{"user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"subscription", 1}}
  opts4 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts4)
  if err != nil {
    log.Printf("coll.Find session: %v", err, userIDs, r.URL.Path, r.Form)
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
    log.Printf("coll.Find session: %v", err, userIDs, r.URL.Path, r.Form)
  }

  var arr []interface{}
  arr = append(arr, "storePush")
  arr = append(arr, channelID)
  arr = append(arr, aliasName)
  arr = append(arr, contents)
  arr = append(arr, r.FormValue("targetStore"))
  jsonData, err := json.Marshal(arr)
  if err != nil {
    log.Printf("json.Marshal storePush: %v", err, jsonData, r.URL.Path, r.Form)
  }
  chunk := false
  chunks := []string{string(jsonData)}
  if len(jsonData) > 2000 {
    chunk = true
    // chunks = SplitIntoByteChunks(string(jsonData), maxChunkSize)
    chunks = common.SplitIntoByteChunks(string(jsonData), 2000 / utf8.UTFMax)
  }
  chunkLength := len(chunks)
  chunkPass := common.StringRand(2)
	for chIndex, ch := range chunks {
	  for _, ssData := range sessions {
	    pushID := common.StringRand(1)
	    var arrForJson []interface{}
	    if chunk {
	    	arrForJson = append(arrForJson, pushID)
			  arrForJson = append(arrForJson, "chunk")
			  arrForJson = append(arrForJson, ch)
			  arrForJson = append(arrForJson, chunkPass)
			  arrForJson = append(arrForJson, chIndex)
			  arrForJson = append(arrForJson, chunkLength)
	    } else {
		    arrForJson = append([]interface{}{pushID}, arr...)
	    }
	    jsonD, err := json.Marshal(arrForJson)
	    if err != nil {
	      log.Printf("json.Marshal chunks sessions: %v", err, arrForJson, r.URL.Path, r.Form)
	    }
	    webpushSub := &webpush.Subscription{}
	    json.Unmarshal([]byte(ssData.Subscription), webpushSub)
	    resp, err := webpush.SendNotification([]byte(string(jsonD)), webpushSub, &webpush.Options{
	      Subscriber:      "example@example.com",
	      VAPIDPublicKey:  common.VAPIDPublicKey,
	      VAPIDPrivateKey: common.VAPIDPrivateKey,
	      TTL:             30,
	    })
	    if err != nil {
	      log.Printf("SendNotification storePush: %v", err, string(jsonD), r.URL.Path, r.Form)
	    }
	    defer resp.Body.Close()
	  }
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
