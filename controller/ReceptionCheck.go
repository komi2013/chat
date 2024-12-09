package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func ReceptionCheck(w http.ResponseWriter, r *http.Request) {
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
	for i, table := range reception.Tables {
		for _, passcode := range table.Passcodes {
			if passcode == code && (table.CurrentCode == "" || table.CurrentCode == code) {
				found = true
				reception.Tables[i].CurrentCode = code
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

	update := bson.M{
		"$set": bson.M{"tables": reception.Tables},
	}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Failed to update table status", http.StatusInternalServerError)
		return
	}
	// メニューとアイテム詳細のみを返却
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Menus       []collection.Menu       `json:"menus"`
		ItemDetails []collection.ItemDetail `json:"itemDetails"`
	}{
		Menus:       reception.Menus,
		ItemDetails: reception.ItemDetails,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response to JSON", http.StatusInternalServerError)
	}
}

