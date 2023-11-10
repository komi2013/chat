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

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func MessagePost(w http.ResponseWriter, r *http.Request) {
  cookie, _ := r.Cookie("ss")
  // if err != nil {
  //  return ""
  // }

  channelID := r.FormValue("channelID")
  // messageID := r.FormValue("MessageID")
  // messageTxt := r.FormValue("MessageTxt")
  // messageType := r.FormValue("MessageType")
  // editFlg := r.FormValue("EditFlg")


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
    {"alias_names", 1},
  })
  coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  if err != nil {
    panic(err)
  }
  var community collection.CommunityStruct
  coll = db1.Collection("community")
	filter2 := bson.D{
		{"alias_name", bson.D{{"$in", session.AliasNames}}},
		{"channel_id", channelID},
	}
	// filter2 := bson.D{}
  opts2 := options.FindOne().SetProjection(bson.D{
    {"_id", 1},
    {"channel_id", 1},
    {"alias_name", 1},
    {"user_ids", 1},
    {"channel_db", 1},
    {"updated_at", 1},
  })
  coll.FindOne(context.TODO(), filter2, opts2).Decode(&community)
  if err != nil {
    fmt.Printf(" err %s\n", err)
  }
  fmt.Printf("community %+v\n", community)
  coll = db1.Collection("session")
  filter = bson.D{{
  	"user_id", bson.D{{"$in", community.UserIDs}}}}
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


//     arr = append(arr, r.MessageID)  0
//     arr = append(arr, r.ChannelID)  1
//     arr = append(arr, r.MessageTxt) 2
//     arr = append(arr, r.MessageType)3
//     arr = append(arr, r.From)       4
//     arr = append(arr, r.FromImg)    5
//     arr = append(arr, r.EditFlg)    6
//     arr = append(arr, r.ParentID)   7
//     arr = append(arr, r.Emojis)     8
//     arr = append(arr, r.CreatedAt)  9
// props.msgs = [
//   ["MessageID1", "ChannelID1", "MessageTxt A", 0, "alias A", "/me.jpg", 0, "ParentID1", [["aliasA","🙇"]], "09:00" ],
//   ["id2", "alias A", "/me.jpg", "09:30", 0, "message text,message textmessage textmessage textmessage text", "", [["aliasB","🙇"]]]
//   ]

  for _, r := range results4 {
    cursor.Decode(&r)
		webpushSub := &webpush.Subscription{}
		json.Unmarshal([]byte(r.Subscription), webpushSub)

		// Send Notification
		resp, err := webpush.SendNotification([]byte(`["MessageID1", "ChannelID1", "MessageTxt A", 0, "alias A", "/me.jpg", 0, "ParentID1", [["aliasA","🙇"]], "09:00" ]`), webpushSub, &webpush.Options{
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

	// coll := db1.Collection("message")
	// filter := bson.D{{"_id", messageID}}

// message_id
// channel_id
// message_txt
// message_type
// from
// from_img
// edit_flg
// parent_id
// emojis
// created_at

	// update := bson.D{{"$set", bson.D{
	// 	{"note_title", n.NoteTitle},
	// 	{"category_id", n.CategoryID},
	// 	{"note_txt", n.NoteTxt},
	// 	{"created_at", n.CreatedAt},
	// 	{"updated_at", n.UpdatedAt}}}}
	// opts := options.Update().SetUpsert(true)
	// _, err := coll.UpdateOne(context.TODO(), filter, update, opts)

  fmt.Fprint(w, `{"Status":"1"}`)
}
