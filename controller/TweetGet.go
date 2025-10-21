package controller

import (
  "context"
  "encoding/json"
  "log"
  "net/http"
  "strconv"
  "time"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/common"
  "chat/collection"
)

func TweetGet(w http.ResponseWriter, r *http.Request) {

  parentID := r.FormValue("parentID")
  limitStr := r.FormValue("limit")
  csrf := r.FormValue("csrf")

  if parentID == "" {
    common.WriteResponseWithoutSession(w, csrf, "parentID is required", http.StatusBadRequest)
    return
  }

  limit := 20 // default latest 20 tweets
  if limitStr != "" {
    if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
      limit = l
    }
  }

  session, err := common.SessionCheckTake(w, r, csrf)
  if err != nil {
    log.Printf("SessionCheckTake: %v; Req: %v", err, r.URL.Path)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  coll := common.DB.TweetDB.Collection("tweet")

  // filter by document ID (parentID)
  filter := bson.M{"_id": parentID}

  // projection: only the latest `limit` tweets (lightweight slice)
  projection := bson.M{
    "tweets":     bson.M{"$slice": -limit},
    "tweetHeads": 1, // include tweetHeads if exists
  }

  opts := options.FindOne().SetProjection(projection)

  var tweet collection.TweetStruct
  err = coll.FindOne(ctx, filter, opts).Decode(&tweet)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
  }

	// ======== Response ========
	responseData := struct {
		Tweet          collection.TweetStruct   `json:"tweet"`
		Csrf           string   `json:"csrf"`
		PushContents   []string `json:"pushContents"`
	}{
		Tweet: tweet,
		Csrf:           session.Csrf,
		PushContents:   session.PushContents,
	}

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
