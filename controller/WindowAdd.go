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

  "chat/common"
  "chat/collection"
)

func WindowAdd(w http.ResponseWriter, r *http.Request) {
	session, err := common.Session(w,r)
	if err != nil {
  	http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
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

  verifiedName := ""
  for _, aliasArray := range session.AliasArray {
  	if (r.FormValue("aliasName") == aliasArray[0]) {
  		verifiedName = r.FormValue("aliasName")
  	}
  }
  if verifiedName == "" {
  	fmt.Printf("verifiedName err %s\n", r.FormValue("aliasName"))
  	return
  }
  var subscriptions []string
  if r.FormValue("userIDs") != "" {
	  jsonBytes := []byte(r.FormValue("userIDs"))
	  var userIDs []interface{}
	  json.Unmarshal(jsonBytes, &userIDs)

	  coll := db1.Collection("session")
	  filter := bson.D{{
	    "user_id", bson.D{{"$in", userIDs}}}}
	  project := bson.D{{"subscription", 1}}
	  opts4 := options.Find().SetProjection(project)
	  cursor, err := coll.Find(context.TODO(), filter, opts4)
	  if err != nil {
	    fmt.Printf("err %s\n", err)
	  }
	  var results []collection.SessionStruct
	  if err = cursor.All(context.TODO(), &results); err != nil {
	      fmt.Printf("err %s\n", err)
	  }
		for _, result := range results {
		  subscriptions = append(subscriptions, result.Subscription)
		}
	}
	windowID := r.FormValue("windowID")
	if windowID == "" {
		windowID = common.StringRand(4)
	}
	
	coll := db1.Collection("window")
  window := collection.WindowStruct{
  	ID: windowID,
		ChannelID: r.FormValue("channelID"),
		CreatorAlias: verifiedName,
		CreatorUser: session.UserID,
		Contents: r.FormValue("contents"),
    UpdatedAt: time.Now(),
		Subscriptions: subscriptions,
		Contents2: r.FormValue("contents2"),
	}
	_, err = coll.InsertOne(context.TODO(), window)
	if err != nil {
		log.Print(err)
	}

	var jsonArr []interface{}
	jsonArr = append(jsonArr, 1)
	jsonArr = append(jsonArr, windowID)
	jsonData, _ := json.Marshal(jsonArr)
  fmt.Fprint(w, string(jsonData))
}
