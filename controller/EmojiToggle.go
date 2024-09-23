package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  // "strconv"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func EmojiToggle(w http.ResponseWriter, r *http.Request) {
	session, err := common.Session(w,r)
	if err != nil {
  	http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
    return
	}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  channelID := r.FormValue("channelID")
  messageID := r.FormValue("messageID")
  aliasName := r.FormValue("aliasName")
  parentID := r.FormValue("parentID")
  emojiValue := r.FormValue("emojiValue")

  trueAccess := false
  for _, arrayData := range session.AliasArray {
    if arrayData[0] == aliasName && arrayData[1] == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    fmt.Printf(" err %s\n", session.AliasArray, aliasName)
    return
  }
  jsonBytes := []byte(r.FormValue("userIDs"))
  userIDs := []string{}
  json.Unmarshal(jsonBytes, &userIDs)
  fmt.Println("userIDs:", userIDs)

  coll := db1.Collection("session")
  filter := bson.D{{
    "user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"subscription", 1}}
  opts4 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts4)
  if err != nil {
    fmt.Printf(" err %s\n", err)
  }
  var results4 []collection.SessionStruct
  if err = cursor.All(context.TODO(), &results4); err != nil {
    fmt.Printf(" err %s\n", err)
  }

  editType := 1
  if r.FormValue("clicked") == "1" {
    editType = 3 // delete
  }
  for _, r4 := range results4 {
  	pushID := common.StringRand(12)
	  var arr []interface{}
	  arr = append(arr, pushID)
	  arr = append(arr, "emoji")
	  arr = append(arr, messageID)
	  arr = append(arr, aliasName)
	  arr = append(arr, editType)
	  arr = append(arr, emojiValue)
	  arr = append(arr, parentID)
	  arr = append(arr, time.Now())

	  jsonData, err := json.Marshal(arr)
	  if err != nil {
	    fmt.Println("JSON変換エラー:", err)
	  }
	  fmt.Println("JSONs成功:", jsonData)
	  // JSON 文字列を表示
	  fmt.Println(string(jsonData))

		coll = db1.Collection("push")
		document := bson.M{
	    "_id": pushID,
	    "pushJson": string(jsonData),
	    "created_at": time.Now().Format("2006-01-02 15:04:05"),
		}
		_, err = coll.InsertOne(context.TODO(), document)
		if err != nil {
		  fmt.Printf("err %s\n", err)
		}

    cursor.Decode(&r4)
    webpushSub := &webpush.Subscription{}
    json.Unmarshal([]byte(r4.Subscription), webpushSub)

    // Send Notification
    resp, err := webpush.SendNotification([]byte(string(jsonData)), webpushSub, &webpush.Options{
      Subscriber:      "example@example.com",
      VAPIDPublicKey:  common.VAPIDPublicKey,
      VAPIDPrivateKey: common.VAPIDPrivateKey,
      TTL:             30,
    })
    if err != nil {
      // TODO: Handle error
      fmt.Printf(" err %s\n", err)
    }
    defer resp.Body.Close()
  }

  fmt.Fprint(w, `{"Status":"1"}`)
}
