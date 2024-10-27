package controller

import (
  "context"
  // "encoding/base64"
  "encoding/json"
  "fmt"
  // "io/ioutil"
  "log"
  "net/http"
  // "os"
  // "strconv"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func TicketAdd(w http.ResponseWriter, r *http.Request) {
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

  jsonBytes := []byte(r.FormValue("ticket"))
  var tk bson.M
  err = json.Unmarshal(jsonBytes, &tk)
  var ticket collection.Ticket
	ticket.Status, err = collection.ValidateTicketStatus(tk["status"].(float64))
	if common.ResponseErrorStatus(w, err) { return }

	ticket.Title, err = collection.ValidateTicketTitle(tk["title"].(string))
	if common.ResponseErrorStatus(w, err) { return }

	ticket.CreatedAt = time.Now()
	ticket.ContentsType, err = collection.ValidateTicketContentsType(tk["contentsType"].(float64))
	if common.ResponseErrorStatus(w, err) { return }

	jsonInterface, ok := tk["contents"].([]interface{})
	if !ok {
	    log.Println("Type assertion failed: tk['contents'] is not of type []interface{}")
	    return
	}
	var contents []primitive.M
	for _, item := range jsonInterface {
	    if contentMap, ok := item.(map[string]interface{}); ok {
	        contents = append(contents, contentMap)
	    } else {
	        log.Println("Failed to cast an element of contents to primitive.M")
	        return
	    }
	}
	ticket.Contents = contents
	jsonInterface, ok = tk["accessNames"].([]interface{})
	if !ok {
	    log.Println("Type assertion failed: tk['accessNames'] is not of type []interface{}")
	    return
	}

	var accessNames []string
	for _, item := range jsonInterface {
	    if itemStr, ok := item.(string); ok {  // Cast each item to string
	        accessNames = append(accessNames, itemStr)
	    } else {
	        log.Println("Failed to cast an element of accessNames to string")
	        return
	    }
	}
	ticket.AccessNames = accessNames

	coll := db1.Collection("ticket")
	insertResult, err := coll.InsertOne(context.TODO(), ticket)
	if err != nil {
		http.Error(w, "Failed to insert document into MongoDB", http.StatusInternalServerError)
		return
	}
	insertedID := insertResult.InsertedID.(primitive.ObjectID)

  jsonBytes = []byte(r.FormValue("userIDs"))
  userIDs := []string{}
  json.Unmarshal(jsonBytes, &userIDs)
  coll = db1.Collection("session")
  filter := bson.D{{
    "user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"subscription", 1}}
  opts4 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts4)
  if err != nil {
    fmt.Printf("get subscription err %s\n", err)
  }
  var results4 []collection.SessionStruct
  if err = cursor.All(context.TODO(), &results4); err != nil {
    fmt.Printf(" err %s\n", err)
  }
  for _, r4 := range results4 {
	  pushID := common.StringRand(12)
	  var arr []interface{}
	  arr = append(arr, pushID)
	  arr = append(arr, "ticketEdit")
	  arr = append(arr, insertedID.Hex())
	  arr = append(arr, ticket.Title)
	  jsonData, err := json.Marshal(arr)
	  if err != nil {
	    fmt.Println("JSON変換エラー:", err)
	  }
		coll = db1.Collection("push")
		document := bson.M{
	    "_id": pushID,
	    "pushJson": string(jsonData),
	    "created_at": time.Now().Format("2006-01-02 15:04:05"),
		}
		_, err = coll.InsertOne(context.TODO(), document)
		if err != nil {
		    fmt.Printf("push DB err %s\n", err)
		}
    cursor.Decode(&r4)
    webpushSub := &webpush.Subscription{}
    json.Unmarshal([]byte(r4.Subscription), webpushSub)

    resp, err := webpush.SendNotification([]byte(string(jsonData)), webpushSub, &webpush.Options{
      Subscriber:      "example@example.com",
      VAPIDPublicKey:  common.VAPIDPublicKey,
      VAPIDPrivateKey: common.VAPIDPrivateKey,
      TTL:             30,
    })
    if err != nil {
      // TODO: Handle error
      fmt.Printf("push send err %s\n", err)
    }
    defer resp.Body.Close()
  }
	response := struct {
		TicketID string `json:"TicketID"`
	}{
		TicketID: insertedID.Hex(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}