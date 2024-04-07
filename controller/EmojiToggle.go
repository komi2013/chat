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
  "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func EmojiToggle(w http.ResponseWriter, r *http.Request) {
  cookie, _ := r.Cookie("ss")
  // if err != nil {
  //  return ""
  // }
  fmt.Printf("channelID %+v\n", r.FormValue("messageID"))
  primitiveChannelID, err := primitive.ObjectIDFromHex(r.FormValue("channelID"))
  if err != nil {
    log.Print(err)
  }
  // channelID := r.FormValue("channelID")
  messageID := r.FormValue("messageID")
  aliasName := r.FormValue("aliasName")
  parentID := r.FormValue("parentID")
  emojiValue := r.FormValue("emojiValue")

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
    if arrayData[0] == aliasName {
      trueAccess = true
    }
  }
  if !trueAccess {
    fmt.Printf(" err %s\n", session.AliasArray, aliasName)
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
  })
  coll.FindOne(context.TODO(), filter2, opts2).Decode(&channel)
  if err != nil {
    fmt.Printf(" err %s\n", err)
  }
  fmt.Printf("channel.UserIDs %+v\n", channel)
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
  var deleteType int
  if r.FormValue("clicked") == "1" {
    deleteType = 1
  }
  coll = db1.Collection("emoji")
  emoji := collection.EmojiStruct{
    MessageID: messageID,
    AliasName: aliasName,
    EmojiValue: emojiValue,
    CreatedAt: time.Now(),
    DeleteType: deleteType,
  }
  _, err = coll.InsertOne(context.TODO(), emoji)
  if err != nil {
    log.Fatal(err)
  }

  editType := 1
  if r.FormValue("clicked") == "1" {
    editType = 3 // delete
  }
  var arr []interface{}
  arr = append(arr, "emoji")
  arr = append(arr, messageID)
  arr = append(arr, aliasName)
  arr = append(arr, editType)
  arr = append(arr, emojiValue)
  arr = append(arr, parentID)
  arr = append(arr, time.Now())

  msgJson, err := json.Marshal(arr)
  if err != nil {
    fmt.Println("JSON変換エラー:", err)
  }
  fmt.Println("JSONs成功:", msgJson)
  // JSON 文字列を表示
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
