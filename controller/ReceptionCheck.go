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
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func ReceptionCheck(w http.ResponseWriter, r *http.Request) {
	receptionID := r.FormValue("receptionID")

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Code is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	if err != nil {
		log.Printf("mongo.Connect error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer c.Disconnect(ctx)

	db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck error: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

  coll := db1.Collection("reception")

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
	for _, table := range reception.Seats {
		for _, passcode := range table.Passcodes {
			if passcode.Passkey == code {
				found = true
				// reception.Seats[i].CurrentCode = code
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
		"$set": bson.M{"seats": reception.Seats},
	}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Failed to update seat status", http.StatusInternalServerError)
		return
	}
	// メニューとアイテム詳細のみを返却
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
		Menus        []collection.Menu       `json:"menus"`
		ItemDetails  []collection.ItemDetail `json:"itemDetails"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Menus:       reception.Menus,
		ItemDetails: reception.ItemDetails,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response to JSON", http.StatusInternalServerError)
	}
}

