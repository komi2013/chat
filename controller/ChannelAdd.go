package controller

import (
  "context"
  // "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"

  // webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func ChannelAdd(w http.ResponseWriter, r *http.Request) {
  cookie, _ := r.Cookie("ss")
  // if err != nil {
  //  return ""
  // }

  channelName := r.FormValue("channelName")
  aliasName := r.FormValue("aliasName")
  channelDescription := r.FormValue("channelDescription")

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
  var aliasImg string
  var userIDs []string
  for _, arrayData := range session.AliasArray {
  	if arrayData[0] == aliasName {
  		trueAccess = true
  		aliasImg = arrayData[1]
  		userIDs = append(userIDs, aliasName)
  	}
  }
  if !trueAccess {
  	fmt.Printf(" err %s\n", session.AliasArray, aliasName)
  	return
  }

  coll = db1.Collection("channel")
  // var channel collection.ChannelStruct
  channel := collection.ChannelStruct{
		ChannelName:   channelName,
		ChannelDescription:  channelDescription,
		UpdatedAt:  time.Now(),
	}
	insertResult, err := coll.InsertOne(context.TODO(), channel)
	if err != nil {
		log.Fatal(err)
	}
	insertedID := insertResult.InsertedID.(primitive.ObjectID)
	fmt.Println("Inserted document ID:", insertedID)

  coll = db1.Collection("community")
  // var channel collection.ChannelStruct
  community := collection.CommunityStruct{
		ChannelID:   insertedID,
		AliasName:   aliasName,
		AliasImg:   aliasImg,
		UserIDs: userIDs,
		UpdatedAt:  time.Now(),
	}
	insertResult, err = coll.InsertOne(context.TODO(), community)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(" insertResult %s\n", insertResult)
  fmt.Fprint(w, `{"Status":"1"}`)
}

func containsString(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}
