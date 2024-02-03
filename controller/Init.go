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

  "chat/collection"
  "chat/common"

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
		{"alias_array", 1},
	})
	coll.FindOne(context.TODO(), filter, opts).Decode(&session)
	if err != nil {
		panic(err)
	}

	coll = db1.Collection("alias")
	filter = bson.D{{"user_id", session.UserID}}
	project := bson.D{
		{"alias_name", 1},
		{"alias_img", 1},
		{"group_flg", 1},
		{"channel_ids", 1}}
	opts3 := options.Find().SetProjection(project)
	cursor, _ := coll.Find(context.TODO(), filter, opts3)
	var results2 []collection.AliasStruct
	if err = cursor.All(context.TODO(), &results2); err != nil {
		fmt.Printf(" err %s\n", err)
	}
	var aliasArr []interface{}
	var arrChannelID  []string
	for _, r := range results2 {
		cursor.Decode(&r)
		var arr []interface{}
		arr = append(arr, r.AliasName)
		arr = append(arr, r.AliasImg)
		arr = append(arr, r.GroupFlg)
		aliasArr = append(aliasArr, arr)
		arrChannelID = append(arrChannelID, r.ChannelIDs...)
	}
	fmt.Printf(" arrChannelID %s\n", arrChannelID)
	var hexChannelIDs []primitive.ObjectID
	for _, r := range arrChannelID {
		primitiveID, err := primitive.ObjectIDFromHex(r)
		if err != nil {
			log.Fatal(err)
		}		
		hexChannelIDs = append(hexChannelIDs, primitiveID)
	}
	coll = db1.Collection("channel")
	filter = bson.D{{"_id", bson.D{{"$in", hexChannelIDs}}}}
	// filter = bson.D{}
	project = bson.D{
		{"_id", 1},
		{"channel_name", 1},
		{"channel_description", 1},
		{"alias_array", 1},
		{"updated_at", 1}}
	opts4 := options.Find().SetProjection(project)
	cursor, err = coll.Find(context.TODO(), filter, opts4)
	if err != nil {
		fmt.Printf(" err %s\n", err)
	}
	var results4 []collection.ChannelStruct
	if err = cursor.All(context.TODO(), &results4); err != nil {
		fmt.Printf(" err %s\n", err)
	}
	var channelArr []interface{}
	for _, r := range results4 {
		cursor.Decode(&r)
		var arr []interface{}
		arr = append(arr, r.ChannelID)
		arr = append(arr, r.ChannelName)
		arr = append(arr, r.ChannelDescription)
		arr = append(arr, r.UpdatedAt)
		arr = append(arr, r.AliasArray)
		var myAliasName string
		for _, r2 := range r.AliasArray {
			for i3, r3 := range results2 {
        if r2[0] == r3.AliasName && ( i3 == len(results2)-1 || r3.GroupFlg != 1 ) {
        	myAliasName = r3.AliasName
        }
			}
		}
		arr = append(arr, myAliasName)
		channelArr = append(channelArr, arr)
	}
	fmt.Printf(" channelArr %s\n", channelArr)
	status := 1
	var jsonArr []interface{}
	jsonArr = append(jsonArr, status)
	jsonArr = append(jsonArr, channelArr)
	jsonArr = append(jsonArr, aliasArr)
	fmt.Printf(" jsonArr %s\n", jsonArr)
	jsonData, _ := json.Marshal(jsonArr)
	fmt.Printf(" jsonData %s\n", jsonData)
  fmt.Fprint(w, string(jsonData))
}
