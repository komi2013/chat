package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
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

func ReceptionDelete(w http.ResponseWriter, r *http.Request) {
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

  apiKey := r.FormValue("apiKey")
  if apiKey == "" {
    http.Error(w, "apiKey is required", http.StatusBadRequest)
    return
  }
  tableName := r.FormValue("tableName")
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
  if reception.ApiKey != apiKey {
  	http.Error(w, "apiKey is different", http.StatusInternalServerError)
  	return
  }

	// Find the matching table and update CurrentCode and Passcodes
	updatedTables := make([]collection.Table, 0, len(reception.Tables))
	var removedCode string
	for _, table := range reception.Tables {
		if table.TableName == tableName {
			// Save the current code for removal from Passcodes
			removedCode = table.CurrentCode
			table.CurrentCode = "" // Clear CurrentCode
			// Remove the code from Passcodes
			newPasscodes := make([]string, 0, len(table.Passcodes))
			for _, passcode := range table.Passcodes {
				if passcode != removedCode {
					newPasscodes = append(newPasscodes, passcode)
				}
			}
			table.Passcodes = newPasscodes
		}
		updatedTables = append(updatedTables, table)
	}

	// Prepare the update
	update := bson.M{
		"$set": bson.M{"tables": updatedTables},
	}
	opts := options.Update().SetUpsert(false)

	// Apply the update to MongoDB
	_, err = coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		http.Error(w, "failed to update reception", http.StatusInternalServerError)
		return
	}

  var arr []interface{}
  arr = append(arr, "receptionOrder")
  arr = append(arr, tableName)
  for _, subscription := range reception.Subscription {
    pushID := common.StringRand(12)
    arrForPush := append([]interface{}{pushID}, arr...)
    jsonData, err := json.Marshal(arrForPush)
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
  if err := json.NewEncoder(w).Encode(arr); err != nil {
    http.Error(w, "Failed to encode reception to JSON", http.StatusInternalServerError)
  }
}

