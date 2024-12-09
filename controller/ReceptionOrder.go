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

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func ReceptionOrder(w http.ResponseWriter, r *http.Request) {
  _, err := common.Session(w,r)
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
  coll := db1.Collection("reception")

  receptionID, err := primitive.ObjectIDFromHex(r.FormValue("receptionID"))
  if err != nil {
    http.Error(w, "Invalid receptionID format", http.StatusBadRequest)
    return
  }

  code := r.FormValue("code")
  if code == "" {
    http.Error(w, "Code is required", http.StatusBadRequest)
    return
  }

  var reception collection.ReceptionStruct
  filter := bson.M{"_id": receptionID}
  err = coll.FindOne(ctx, filter).Decode(&reception)
  if err != nil {
    if err == mongo.ErrNoDocuments {
      http.Error(w, "Reception not found", http.StatusNotFound)
    } else {
      http.Error(w, "Database query failed", http.StatusInternalServerError)
    }
    return
  }

  found := false
  tableName := ""
  for i, table := range reception.Tables {
    for _, passcode := range table.Passcodes {
      if passcode == code && (table.CurrentCode == "" || table.CurrentCode == code) {
        found = true
        reception.Tables[i].CurrentCode = code
        tableName = table.TableName
      }
    }
    if found {
      break
    }
  }

  if !found {
    http.Error(w, "Code not associated with any table", http.StatusNotFound)
    return
  }

  for _, subscription := range reception.Subscription {
	  pushID := common.StringRand(12)
	  var arr []interface{}
	  arr = append(arr, pushID)
	  arr = append(arr, "receptionOrder")
	  arr = append(arr, r.FormValue("menuID"))
	  arr = append(arr, tableName)
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
		    fmt.Printf("err %s\n", err)
		}
    // cursor.Decode(&r4)
    webpushSub := &webpush.Subscription{}
    json.Unmarshal([]byte(subscription), webpushSub)

    // Send Notification
    resp, err := webpush.SendNotification([]byte(string(jsonData)), webpushSub, &webpush.Options{
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

  w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(reception); err != nil {
    http.Error(w, "Failed to encode reception to JSON", http.StatusInternalServerError)
  }
}

