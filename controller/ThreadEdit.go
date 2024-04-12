package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func ThreadEdit(w http.ResponseWriter, r *http.Request) {
  cookie, _ := r.Cookie("ss")
  // if err != nil {
  //  return ""
  // }
  primitiveChannelID, err := primitive.ObjectIDFromHex(r.FormValue("channelID"))
  if err != nil {
    log.Fatal(err)
  }

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  var session collection.SessionStruct

  coll := db1.Collection("session")
  filter := bson.D{{"_id", cookie.Value}}
  // opts := options.FindOne().SetProjection(projection)
  opts := options.FindOne().SetProjection(bson.D{
    {"user_id", 1},
    {"alias_array", 1},
  })
  coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  if err != nil {
    panic(err)
  }
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

  var channel collection.ChannelStruct
  coll = db1.Collection("channel")
  // check access right
  filter2 := bson.D{
    {"_id", primitiveChannelID},
  }
  opts2 := options.FindOne().SetProjection(bson.D{
    {"user_ids", 1},
    {"alias_array", 1},
  })
  coll.FindOne(context.TODO(), filter2, opts2).Decode(&channel)
  if err != nil {
    fmt.Printf(" err %s\n", err)
  }

 	var yets [][]string
  for _, arrayData := range channel.AliasArray {
  	atName := "＠＠" + arrayData[0] + "・＠＠"
  	strings.Contains(r.FormValue("messageTxt"), atName)
    if strings.Contains(r.FormValue("messageTxt"), atName) {
      yets = append(yets, []string{arrayData[0], "/img/yet.png"})
    }
  }

  coll = db1.Collection("session")
  filter = bson.D{{
    "user_id", bson.D{{"$in", channel.UserIDs}}}}
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
  coll = db1.Collection("thread_edit")
  messageEdit := collection.MessageEditStruct{
    MessageID: r.FormValue("messageID"),
    MessageTxt: r.FormValue("messageTxt"),
    Task: r.FormValue("task"),
    CreatedAt: time.Now(),
  }
  _, err = coll.InsertOne(context.TODO(), messageEdit)
  if err != nil {
    log.Fatal(err)
  }
  var arr []interface{}
  arr = append(arr, "threadEdit")
  arr = append(arr, messageEdit.MessageID)
  arr = append(arr, messageEdit.MessageTxt)
  if messageEdit.Task != "" {
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
      VAPIDPublicKey:  "BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8",
      VAPIDPrivateKey: "bdSiNzUhUP6piAxLH-tW88zfBlWWveIx0dAsDO66aVU",
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
