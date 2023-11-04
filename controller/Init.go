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

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func Init(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("ss")
	// if err != nil {
	// 	return ""
	// }

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
	// fmt.Printf(" session.AliasNames %s\n", session.AliasNames)
	// dataJson := `["1","2","3"]`
	// jsonParser := json.NewDecoder(session.AliasNames)
	bytes := []byte(session.AliasNames)
	// var arr [][]interface{}
  var names [][]interface{}
  if err := json.Unmarshal(bytes, &names); err != nil {
    log.Fatal(err)
  }

  var arrName []string
  for _, d := range names {
		str, ok := d[0].(string)
		if !ok {
		  fmt.Printf("ERROR: not a string -> %#v\n", d[0])
		}
  	arrName = append(arrName, str)
  }
	coll = db1.Collection("community")
	filter = bson.D{{"alias_name", bson.D{{"$in", arrName}}}}
	project := bson.D{
		{"channel_id", 1},
		{"unread_flg", 1},
		{"channel_db", 1}}
	opts2 := options.Find().SetProjection(project)
	cursor, err := coll.Find(context.TODO(), filter, opts2)
	if err != nil {
		fmt.Printf(" err %s\n", err)
	}
	var results []collection.CommunityStruct
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	// fmt.Printf(" results %s\n", results)
	var arrChannelID []string
	for _, r := range results {
		cursor.Decode(&r)
		arrChannelID = append(arrChannelID, r.ChannelID)
	}
	// coll = db1.Collection("channel")
	// filter = bson.D{{"_id", bson.D{{"$in", arrChannelID}}}}
	// project = bson.D{
	// 	{"channel_id", 1},
	// 	{"channel_name", 1},
	// 	{"channel_description", 1}}
	// opts := options.Find().SetProjection(project)
	// cursor, err := coll.Find(context.TODO(), filter, opts)
	// if err != nil {
	// 	return err
	// }
	// var results []collection.ChannelGroupStruct
	// if err = cursor.All(context.TODO(), &results); err != nil {
	// 	return err
	// }
	// var channelArr []collection.ChannelStruct
	// for _, r := range results {
	// 	cursor.Decode(&r)
	// 	channelArr = append(channelArr, r)
	// }

	coll = db1.Collection("message")
	filter = bson.D{{"channel_id", bson.D{{"$in", arrChannelID}}}}
	project = bson.D{
		{"_id", 1},
		{"channel_id", 1},
		{"message_txt", 1},
		{"message_type", 1},
		{"from", 1},
		{"edit_flg", 1},
		{"parent_id", 1},
		{"emojis", 1},
		{"updated_at", 1}}
	opts3 := options.Find().SetProjection(project)
	cursor, _ = coll.Find(context.TODO(), filter, opts3)
	// if err != nil {
	// 	return err
	// }
	var results2 []collection.MessageStruct
	if err = cursor.All(context.TODO(), &results2); err != nil {
		fmt.Printf(" err %s\n", err)
	}
	var msgArr []interface{}
	for _, r := range results2 {
		cursor.Decode(&r)
		var arr []interface{}
		arr = append(arr, r.MessageID)
		arr = append(arr, r.MessageType)
		// arr = append(arr, r.MessageID)
		// arr = append(arr, r.MessageID)
		// msgArr = append(msgArr, interface{r.MessageID, r.ChannelID, r.MessageTxt, r.MessageType, r.From, r.EditFlg, r.ParentID, r.Emojis, r.UpdatedAt })

		msgArr = append(msgArr, arr)

	}
	fmt.Printf(" msgArr %s\n", msgArr)
	msgJson, _ := json.Marshal(msgArr)
	fmt.Printf(" msgJson %s\n", msgJson)
  fmt.Fprint(w, `[1,[[1,"channel name 1","description 1"],[2,"channel name 2","description2 1"]],"dfjdkosjo"]`)
}
