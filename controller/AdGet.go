package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  "log"
  // "math"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/bson"
)

func AdGet(w http.ResponseWriter, r *http.Request) {

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheckTake: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  coll := common.DB.AdDB.Collection("ad")
  filter := bson.M{
    "userID": session.UserID,
  }
	var ads []collection.AdStruct
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
	    log.Fatal(err)
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &ads); err != nil {
	    log.Fatal(err)
	}

  responseData := struct {
    Csrf         string       `json:"csrf"`
    PushContents []string     `json:"pushContents"`
    Ads       []collection.AdStruct  `json:"ads"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Ads         : ads,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

