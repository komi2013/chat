package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"
	"unicode/utf8"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func StorePush(w http.ResponseWriter, r *http.Request) {
  session, err := common.Session(w,r)
  if err != nil {
    http.Error(w, "Unauthorized: Session expired or login required", http.StatusUnauthorized)
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

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")

  trueAccess := false
  for _, arrayData := range session.AliasArray {
    if arrayData[0] == aliasName && arrayData[1] == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    fmt.Printf(" err %s\n", session.AliasArray, aliasName)
    return
  }

	coll := db1.Collection("push_select")
	filter := bson.D{{"_id", r.FormValue("pushSelectID")}}
	result, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		fmt.Printf("push_select Failed to delete document: %v\n", err)
		return
	}

	deletedCount := result.DeletedCount
	fmt.Printf("Deleted %d document(s)\n", deletedCount)
	if deletedCount == 0 {
		return
	}

  jsonBytes := []byte(r.FormValue("userIDs"))
  userIDs := []string{}
  json.Unmarshal(jsonBytes, &userIDs)
  jsonBytes = []byte(r.FormValue("contents"))
  var contents interface{}
  json.Unmarshal(jsonBytes, &contents)
  fmt.Println("contents:", contents)
  coll := db1.Collection("session")
  filter := bson.D{{
    "user_id", bson.D{{"$in", userIDs}}}}
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

  var arr []interface{}
  arr = append(arr, r.FormValue("storePush"))
  arr = append(arr, channelID)
  arr = append(arr, aliasName)
  arr = append(arr, contents)
  jsonData, err := json.Marshal(arr)
  if err != nil {
    fmt.Println("JSON変換エラー:", err)
  }
  chunk := false
  var chunks []string
	if len(jsonData) > 2000 {
	  // fmt.Println("JSON data exceeds 2000 bytes")
	  chunk = true
	  chunks = common.SplitIntoByteChunks(string(jsonData), 2000 / utf8.UTFMax)
	} else {
		chunks = []string{"no chunk"}
	}
	chunkLength := len(chunks)
	chunkPass := common.StringRand(2)
	for chIndex, ch := range chunks {
	  for _, r4 := range results4 {
	    pushID := common.StringRand(12)
	    var arrForJson []interface{}
	    if chunk {
	    	arrForJson = append(arrForJson, pushID)
			  arrForJson = append(arrForJson, "chunk")
			  // arr = append(arr, channelID)
			  // arr = append(arr, aliasName)
			  arrForJson = append(arrForJson, ch)
			  arrForJson = append(arrForJson, chunkPass)
			  arrForJson = append(arrForJson, chIndex)
			  arrForJson = append(arrForJson, chunkLength)
	    } else {
		    arrForJson = append([]interface{}{pushID}, arr...)
	    }

	    jsonD, err := json.Marshal(arrForJson)
	    if err != nil {
	      fmt.Println("JSON変換エラー:", err)
	    }
	    coll = db1.Collection("push")
	    document := bson.M{
	      "_id": pushID,
	      "pushJson": string(jsonD),
	      "created_at": time.Now().Format("2006-01-02 15:04:05"),
	    }
	    _, err = coll.InsertOne(context.TODO(), document)
	    if err != nil {
	        fmt.Printf("err %s\n", err)
	    }
	    cursor.Decode(&r4)
	    webpushSub := &webpush.Subscription{}
	    json.Unmarshal([]byte(r4.Subscription), webpushSub)

	    // Send Notification
	    resp, err := webpush.SendNotification([]byte(string(jsonD)), webpushSub, &webpush.Options{
	      Subscriber:      "example@example.com",
	      VAPIDPublicKey:  common.VAPIDPublicKey,
	      VAPIDPrivateKey: common.VAPIDPrivateKey,
	      TTL:             30,
	    })
	    if err != nil {
	      // TODO: Handle error
	      fmt.Printf(" err %s\n", err)
	    }
	    defer resp.Body.Close()
	  }
	}

  fmt.Fprint(w, `{"Status":"1"}`)
}
