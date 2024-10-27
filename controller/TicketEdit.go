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

func TicketEdit(w http.ResponseWriter, r *http.Request) {
  session, err := common.Session(w,r)
  if err != nil {
    http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
    return
  }
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
  if err != nil && r.FormValue("ticket") != "" {
    log.Print("ticket: ", err)
  }

	ticketID, err := primitive.ObjectIDFromHex(tk["ticketID"].(string))
	if common.ResponseErrorStatus(w, err) { return }

	title, err := collection.ValidateTicketTitle(tk["title"].(string))
	if common.ResponseErrorStatus(w, err) { return }

	description, err := collection.ValidateTicketDescription(tk["description"].(string))
	if common.ResponseErrorStatus(w, err) { return }

	newComment, err := collection.ValidateTicketNewComment(r.FormValue("newComment"))
	if common.ResponseErrorStatus(w, err) { return }

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

	coll := db1.Collection("ticket")
	var preTK collection.Ticket
	filter := bson.M{"_id": ticketID}
	err = coll.FindOne(ctx, filter).Decode(&preTK)
	if err != nil {
		log.Print(err, "ticket")
	}
	log.Print("pre title", preTK.Title)
	var changedTexts []bson.M
  updateFields := bson.D{}
  status := int(tk["status"].(float64))
	if status != preTK.Status {
		updateFields = append(updateFields, bson.E{"status", status})
		changedTexts = append(changedTexts, bson.M{"status": preTK.Status})
	}

	if title != preTK.Title {
		updateFields = append(updateFields, bson.E{"title", title})
		changedTexts = append(changedTexts, bson.M{"title": preTK.Title})
	}
	if description != preTK.Description {
		updateFields = append(updateFields, bson.E{"description", description})
		changedTexts = append(changedTexts, bson.M{"description": preTK.Description})
	}
	if len(newComment) > 0 {
		changedTexts = append(changedTexts, bson.M{"comment": newComment})
	}

	// updateFields = append(updateFields, bson.E{"assignee", assignee})
	// updateFields = append(updateFields, bson.E{"description", description})
	// updateFields = append(updateFields, bson.E{"priority", priority})

 //  if approver1 != "" {
 //    updateFields = append(updateFields, bson.E{"approver1", approver1})
 //    // ticket.AccessNames = append(ticket.AccessNames, approver1)
 //  }
 //  if approver2 != "" {
 //    updateFields = append(updateFields, bson.E{"approver2", approver2})
 //  }
 //  if approver3 != "" {
 //    updateFields = append(updateFields, bson.E{"approver3", approver3})
 //  }
 //  if parentID != "" {
 //  	updateFields = append(updateFields, bson.E{"parent_id", parentID})
 //  }
	updateFields = append(updateFields, bson.E{"updated_at", time.Now()})
  // if r.FormValue("children_ids") != "" {
  //   ticket.ChildrenIDs = childrenIDs
  // }

  // ticket.AccessNames = common.UniqueStrings(ticket.AccessNames)

	coll = db1.Collection("ticket")
	filter = primitive.M{"_id": ticketID}
	update := bson.D{
		{"$set", updateFields},
	}
	// update := bson.D{
	// 	{"$set", bson.D{
	// 		{"status", status},
	// 		{"title", title},
	// 		{"updated_at", time.Now().Format("2006-01-02 15:04:05")},
	// 	}},
	// }
	opts := options.Update().SetUpsert(true)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts)

	ticketLog := collection.TicketLog{
		TicketID   	: ticketID,
		ChangedTexts	: changedTexts,
		CreatedBy  : aliasName,
		CreatedAt  : time.Now(),
	}

	coll = db1.Collection("ticket_log")
	_, err = coll.InsertOne(context.TODO(), ticketLog)
	if err != nil {
		http.Error(w, "Failed to insert document into MongoDB", http.StatusInternalServerError)
		return
	}

  jsonBytes = []byte(r.FormValue("userIDs"))
  userIDs := []string{}
  json.Unmarshal(jsonBytes, &userIDs)
  coll = db1.Collection("session")
	filter = primitive.M{
	    "user_id": primitive.M{
	        "$in": userIDs,
	    },
	}
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
    arr = append(arr, ticketID)
    arr = append(arr, tk["title"].(string))
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
    Status int
  }{
    Status: 1,
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(response)
}

