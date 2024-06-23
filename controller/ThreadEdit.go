package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  // "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func ThreadEdit(w http.ResponseWriter, r *http.Request) {
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

  trueAccess := false
  for _, arrayData := range session.AliasArray {
    if arrayData[0] == r.FormValue("aliasName") {
      trueAccess = true
    }
  }
  if !trueAccess {
    fmt.Printf(" err %s\n", session.AliasArray, r.FormValue("aliasName"))
    return
  }

 	var yets [][]string
  // for _, arrayData := range channel.AliasArray {
  // 	atName := "＠＠" + arrayData[0] + "・＠＠"
  // 	strings.Contains(r.FormValue("messageTxt"), atName)
  //   if strings.Contains(r.FormValue("messageTxt"), atName) {
  //     yets = append(yets, []string{arrayData[0], "/img/yet.png"})
  //   }
  // }

  coll := db1.Collection("session")
  filter := bson.D{{
    "user_id", bson.D{{"$in", r.FormValue("userIDs")}}}}
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
  messageID := common.Base62Encode(time.Now().Unix())
  messageID = messageID + common.StringRand(1)
  var arr []interface{}
  arr = append(arr, "threadEdit")
  arr = append(arr, messageID)
  arr = append(arr, r.FormValue("messageTxt"))
  if r.FormValue("messageTxt") != "" {
		arr = append(arr, yets)
  } else {
  	arr = append(arr, "")
  }
  arr = append(arr, time.Now())

  msgJson, err := json.Marshal(arr)
  if err != nil {
    fmt.Println("JSON変換エラー:", err)
  }
  fmt.Println(string(msgJson))

  for _, r := range results4 {
    cursor.Decode(&r)
    webpushSub := &webpush.Subscription{}
    json.Unmarshal([]byte(r.Subscription), webpushSub)

    // Send Notification
    resp, err := webpush.SendNotification([]byte(string(msgJson)), webpushSub, &webpush.Options{
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
