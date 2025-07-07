package controller

import (
  "context"
  "encoding/json"
  // "io"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"
)

func ReceptionEdit(w http.ResponseWriter, r *http.Request) {
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  client, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect error: %v", err)
    http.Error(w, "Database connection error", http.StatusInternalServerError)
    return
  }
  defer client.Disconnect(ctx)
  db := client.Database(common.MongoDb1)
  session, err := common.SessionCheck(db, w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheck error: %v", err)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  var reception collection.ReceptionStruct
  if err := json.Unmarshal([]byte(r.FormValue("reception")), &reception); err != nil {
    log.Printf("reception: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON reception", http.StatusBadRequest)
    return
  }

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")

  staffAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      staffAccess = true
      break
    }
  }

  coll := db.Collection("reception")
	var existing collection.ReceptionStruct
	filter := bson.M{"_id": channelID}
	err = coll.FindOne(ctx, filter).Decode(&existing)

	if err == mongo.ErrNoDocuments {
		reception.AdminNames = append(reception.AdminNames, aliasName)
	} else if err != nil {
		log.Printf("DB error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	} else {
		adminAccess := false
		for _, admin := range existing.AdminNames {
			if admin == aliasName {
				adminAccess = true
				break
			}
		}
		if !adminAccess && !staffAccess {
			log.Printf("Unauthorized edit attempt: aliasName=%s, channelID=%s, receptionID=%s", aliasName, channelID, reception.ReceptionID)
			http.Error(w, "You do not have permission to edit this reception", http.StatusForbidden)
			return
		}
	}
  reception.UpdatedAt = time.Now()
  update := bson.M{"$set": reception}
	opts := options.Update().SetUpsert(true)
  _, err = coll.UpdateOne(ctx, filter, update, opts)
  if err != nil {
    log.Printf("Failed to update reception: %v", err)
    http.Error(w, "Failed to update reception", http.StatusInternalServerError)
    return
  }
  responseData := struct {
    Csrf         string                     `json:"csrf"`
    PushContents []string                   `json:"pushContents"`
    // Reception    collection.ReceptionStruct `json:"reception"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    // Reception:    updated,
  }

  w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(responseData); err != nil {
    log.Printf("Response encode error: %v", err)
    http.Error(w, "Failed to encode response", http.StatusInternalServerError)
  }
}
