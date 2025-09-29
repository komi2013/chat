package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  "log"
  "net/http"
  // "strconv"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  // webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func ReceptionOrderDelete(w http.ResponseWriter, r *http.Request) {

	aliasName := r.FormValue("aliasName")
  receptionID := r.FormValue("receptionID")
  code := r.FormValue("code")
  if code == "" {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "code is required", http.StatusOK)
    return
  }

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck error: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := common.DB.ReceptionDB.Collection("reception")

  var reception collection.ReceptionStruct
  filter := bson.M{"_id": receptionID}
  err = coll.FindOne(ctx, filter).Decode(&reception)
  if err != nil {
    if err == mongo.ErrNoDocuments {
      common.WriteResponseWithSession(w, session, "Reception not found", http.StatusOK)
    } else {
      common.WriteResponseWithSession(w, session, "Database query failed", http.StatusOK)
    }
    return
  }
  seatName := ""
	updatedTables := make([]collection.Facility, 0, len(reception.Facilities))
	for _, table := range reception.Facilities {
		if table.CurrentCode == code {
			table.CurrentCode = ""
			seatName = table.FacilityName
		}
		updatedTables = append(updatedTables, table)
	}
	var receptionLog = common.NewDailyLogger("reception_")

	update := bson.M{
		"$set": bson.M{"facilities": updatedTables},
	}
	opts := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		receptionLog.Printf("failed to update facilities reception", err, r.URL.Path, r.Form)
		common.WriteResponseWithSession(w, session, "failed to update facilities reception", http.StatusOK)
		return
	}

  collSessions := common.DB.SessionDB.Collection("session")
	filter = bson.M{"userID": bson.M{"$in": reception.OrderUserIDs},}
  cursor, err := collSessions.Find(context.TODO(), filter)
  if err != nil {
    receptionLog.Printf("collSessions.Find: %v; Req: ", err, reception.OrderUserIDs, r.URL.Path, r.Form)
    return
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
    receptionLog.Printf("cursor.All: %v; Req: ", err, reception.OrderUserIDs, r.URL.Path, r.Form)
    return
  }

  var arr []interface{}
	arr = append(arr, "receptionOrder")
	arr = append(arr, reception.ChannelID)
	arr = append(arr, aliasName)
	arr = append(arr, seatName)
	common.ChunkPush(sessions, arr)

	responseData := common.BaseResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

