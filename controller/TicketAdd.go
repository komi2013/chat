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
  "strconv"
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

  // need
	status, _ := strconv.Atoi(r.FormValue("status"))
	title := r.FormValue("title")
	assignee := r.FormValue("assignee")
	// createdBy := r.FormValue("created_by")
	description := r.FormValue("text")

	// option
	jsonBytes := []byte(r.FormValue("contents"))
	var contents []bson.M
	err = json.Unmarshal(jsonBytes, &contents)
	if err != nil && r.FormValue("contents") != "" {
    log.Print(err)
    log.Print("contents: ", err)
  }
	contentsType, _ := strconv.Atoi(r.FormValue("contents_type"))
	// progress, _ := strconv.Atoi(r.FormValue("progress"))
	priority, _ := strconv.Atoi(r.FormValue("priority"))
	startDate := r.FormValue("start_date")
	deadline := r.FormValue("deadline")
	category, _ := strconv.Atoi(r.FormValue("category"))
	approver1 := r.FormValue("approver1")
	approver2 := r.FormValue("approver2")
	approver3 := r.FormValue("approver3")
	parentID := r.FormValue("parent_id")

	jsonBytes = []byte(r.FormValue("children_ids"))
	var childrenIDs []string
	err = json.Unmarshal(jsonBytes, &childrenIDs)
	if err != nil && r.FormValue("children_ids") != "" {
    log.Print("childrenIDs: ", err)
  }
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
	ticket := collection.Ticket{
		Status:      status,
		Title:       title,
		Assignee:    assignee,
		CreatedBy:   aliasName,
		Description: description,
		CreatedAt:   time.Now(),
		ChannelID:   channelID,
		AccessNames: []string{assignee, aliasName},
	}

	if r.FormValue("contents") != "" {
		ticket.Contents = contents
	}
	if contentsType != 0 {
		ticket.ContentsType = contentsType
	}
	if priority != 0 {
		ticket.Priority = priority
	}
	if startDate != "" {
		ticket.StartDate = startDate
	}
	if deadline != "" {
		ticket.Deadline = deadline
	}
	if category != 0 {
		ticket.Category = category
	}
	if approver1 != "" {
		ticket.Approver1 = approver1
		ticket.AccessNames = append(ticket.AccessNames, approver1)
	}
	if approver2 != "" {
		ticket.Approver2 = approver2
		ticket.AccessNames = append(ticket.AccessNames, approver2)
	}
	if approver3 != "" {
		ticket.Approver3 = approver3
		ticket.AccessNames = append(ticket.AccessNames, approver3)
	}
	if parentID != "" {
		ticket.ParentID = parentID
	}
	if r.FormValue("children_ids") != "" {
		ticket.ChildrenIDs = childrenIDs
	}

	ticket.AccessNames = common.UniqueStrings(ticket.AccessNames)

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
	  arr = append(arr, "ticketAdd")
	  arr = append(arr, insertedID.Hex())
	  arr = append(arr, title)
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