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

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func GroupAliasEdit(w http.ResponseWriter, r *http.Request) {
  session, err := common.Session(w,r)
  if err != nil {
    http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
    return
  }
  aliasName := "";
  for i := range session.AliasArray {
    if session.AliasArray[i][0] == r.FormValue("aliasName") {
    	aliasName = r.FormValue("aliasName")
    }
  }
  if aliasName == "" {
  	log.Print("bad aliasName", aliasName)
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

  jsonBytes := []byte(r.FormValue("groupAliases"))
  var groupAliases [][]interface{}
  json.Unmarshal(jsonBytes, &groupAliases)

  for i := range groupAliases {
    groupAliases[i][1] = common.AliasImgSave(groupAliases[i][1].(string), r.FormValue("channelID"), "hii")
  }

  jsonBytes = []byte(r.FormValue("userIDs"))
  var userIDs []interface{}
  json.Unmarshal(jsonBytes, &userIDs)
  fmt.Printf("userIDs %s\n", userIDs)
  coll := db1.Collection("session")
  filter := bson.D{{
    "user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"subscription", 1}}
  opts4 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts4)
  if err != nil {
      fmt.Printf("err %s\n", err)
  }
  var results4 []collection.SessionStruct
  if err = cursor.All(context.TODO(), &results4); err != nil {
      fmt.Printf("err %s\n", err)
  }

  for _, d := range results4 {
    pushID := common.StringRand(12)
    var arr []interface{}
    arr = append(arr, pushID)
    arr = append(arr, "groupAliasEdit")
    arr = append(arr, r.FormValue("channelID"))
    arr = append(arr, groupAliases)
    arr = append(arr, aliasName)
    // arr = append(arr, aliasImg)

    msgJson, err := json.Marshal(arr)
    if err != nil {
      fmt.Println("JSON変換エラー:", err)
    }
    fmt.Println(string(msgJson))

    coll = db1.Collection("push")
    document := bson.M{
      "_id": pushID,
      "contents": string(msgJson),
      "created_at": time.Now().Format("2006-01-02 15:04:05"),
    }
    _, err = coll.InsertOne(context.TODO(), document)
    if err != nil {
        fmt.Printf("InsertOne %s\n", err)
    }

    cursor.Decode(&d)
    webpushSub := &webpush.Subscription{}
    json.Unmarshal([]byte(d.Subscription), webpushSub)
    resp, err := webpush.SendNotification([]byte(string(msgJson)), webpushSub, &webpush.Options{
      Subscriber:      "example@example.com",
      VAPIDPublicKey:  common.VAPIDPublicKey,
      VAPIDPrivateKey: common.VAPIDPrivateKey,
      TTL:             30,
    })
    if err != nil {
      fmt.Printf(" err %s\n", err)
    }
    defer resp.Body.Close()
  }
  fmt.Fprint(w, `{"Status":"1"}`)
}
