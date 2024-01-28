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
  "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func CommunityMatch(w http.ResponseWriter, r *http.Request) {
  cookie, _ := r.Cookie("ss")
  // if err != nil {
  //  return ""
  // }

	channelID, err := primitive.ObjectIDFromHex(r.FormValue("channelID"))
	if err != nil {
		log.Fatal(err)
	}
  code := r.FormValue("code")
  aliasName := r.FormValue("aliasName")

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
  aliasImg := ""
  for _, arrayData := range session.AliasArray {
    if arrayData[0] == aliasName {
      trueAccess = true
      aliasImg = arrayData[1]
    }
  }
  if !trueAccess {
    fmt.Printf(" err %s\n", session.AliasArray, aliasName)
    return
  }

  var channel collection.ChannelStruct
  coll = db1.Collection("channel")

  filter = bson.D{{"_id", channelID}}
  opts2 := options.FindOne().SetProjection(bson.D{
    {"_id", 1},
    {"channel_name", 1},
    {"invitation_code", 1},
    {"invited_at", 1},
  })
  coll.FindOne(context.TODO(), filter, opts2).Decode(&channel)
  if err != nil {
    fmt.Printf("mongo err %+v\n", err)
  }
	if channel.InvitationCode != code || time.Since(channel.InvitedAt).Hours() > 10 {
    fmt.Printf("channel err %+v", channelID, channel)
    fmt.Printf("code %+v", code)
    fmt.Printf("channel.InvitedAt %+v", time.Since(channel.InvitedAt).Hours())
    return
	}

  var userIDs []string
  userIDs = append(channel.UserIDs, session.UserID)

	coll = db1.Collection("channel")
	filter = bson.D{{"_id", channelID}}
	update := bson.D{
		{"$set", bson.D{
			{"user_ids", userIDs},
		}},
	}
	opts3 := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts3)
	if err != nil {
		fmt.Printf(" err %+v\n", err)
	}

  var alias collection.AliasStruct
  coll = db1.Collection("alias")
  filter = bson.D{{"alias_name", aliasName}}
  opts4 := options.FindOne().SetProjection(bson.D{
    {"channel_ids", 1},
  })
  coll.FindOne(context.TODO(), filter, opts4).Decode(&alias)
  if err != nil {
    fmt.Printf(" err %+v\n", err)
  }
  var channelIDs []primitive.ObjectID
	for _, str := range alias.ChannelIDs {
		obj, err := primitive.ObjectIDFromHex(str)
		if err != nil {
			log.Fatal(err)
		}
		channelIDs = append(channelIDs, obj)
	}
  channelIDs = append(channelIDs, channelID)
  fmt.Printf(" channelIDs %+v\n", channelIDs)
	coll = db1.Collection("alias")
	filter = bson.D{{"alias_name", aliasName}}
	update = bson.D{
		{"$set", bson.D{
			{"channel_ids", channelIDs},
		}},
	}
	opts5 := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts5)
	if err != nil {
		fmt.Printf(" err %+v\n", err)
	}
  fmt.Printf("channel.UserIDs %+v", channel.UserIDs)
  coll = db1.Collection("session")
  filter = bson.D{{
  	"user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"subscription", 1}}
  opts6 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts6)
  if err != nil {
    fmt.Printf(" err %+v\n", err)
  }
  var results4 []collection.SessionStruct
  if err = cursor.All(context.TODO(), &results4); err != nil {
    fmt.Printf(" err %+v\n", err)
  }

	var arr []interface{}
	arr = append(arr, "community_join")
	arr = append(arr, channelID)
	arr = append(arr, aliasName)
	arr = append(arr, aliasImg)

	msgJson, err := json.Marshal(arr)
	if err != nil {
		fmt.Println("JSON変換エラー:", err)
	}

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
