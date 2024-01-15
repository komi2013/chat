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

  "chat/common"
  "chat/collection"
)

func InvitationGet(w http.ResponseWriter, r *http.Request) {
  cookie, _ := r.Cookie("ss")
  // if err != nil {
  //  return ""
  // }

  channelID := r.FormValue("channelID")
  // aliasName := r.FormValue("aliasName")
  // channelDescription := r.FormValue("channelDescription")

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
  var aliasNames []string
  for _, arrayData := range session.AliasArray {
  	aliasNames = append(aliasNames, arrayData[0])
  }
  var alias collection.AliasStruct
	coll = db1.Collection("alias")
	filter = bson.D{{"alias_name", bson.D{{"$in", aliasNames}}}}
  opts2 := options.FindOne().SetProjection(bson.D{{"channel_ids", 1},})
	coll.FindOne(context.TODO(), filter, opts2).Decode(&alias)
	trueAccess := false
	for _, r := range alias.ChannelIDs {
		if channelID == r {
			trueAccess = true
		}
	}
  if !trueAccess {
  	fmt.Printf(" err %s\n", session.AliasArray, alias)
  	return
  }
  rand := common.StringRand(15)
	coll = db1.Collection("channel")
	filter = bson.D{{"_id", channelID}}
	update := bson.D{
		{"$set", bson.D{
			{"invitation_code", rand},
			{"invited_at", time.Now()},
		}},
	}
	opts3 := options.Update().SetUpsert(true)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts3)
	if err != nil {
		panic(err)
	}

	var jsonArr []interface{}
	jsonArr = append(jsonArr, 1)
	jsonArr = append(jsonArr, rand)
	jsonData, _ := json.Marshal(jsonArr)
  fmt.Fprint(w, string(jsonData))
}
