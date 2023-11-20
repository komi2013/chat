package controller

import (
  "context"
  "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/common"
)

func InvitationGet(w http.ResponseWriter, r *http.Request) {
  cookie, _ := r.Cookie("ss")
  // if err != nil {
  //  return ""
  // }

  channelID := r.FormValue("channelID")
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
  var aliasNames []string
  for _, arrayData := range session.AliasArray {
  	aliasNames = append(aliasNames, arrayData[0])
  }

	coll = db1.Collection("community")
	filter = bson.D{{"alias_name", bson.D{{"$in", aliasNames}}}}
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
	trueAccess := false
	for _, r := range results {
		cursor.Decode(&r)
		if channelID == r.ChannelID {
			trueAccess = true
		}
	}
  if !trueAccess {
  	fmt.Printf(" err %s\n", session.AliasArray, results)
  	return
  }

	coll := db.Collection("channel")
	filter := bson.D{{"_id", channelID}}
	update := bson.D{
		{"$set", bson.D{
			{"invitation_code", "randome code must be here"},
			{"invited_at", time.Now()},
		}},
	}

	opts := options.Update().SetUpsert(false)
	_, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

  fmt.Fprint(w, `[1]`)
}
