package controller

import (
  "context"
  "encoding/json"
  // "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/common"
  "chat/collection"
)

// TweetGetLatest : UpdatedAt順で最新100件のTweetStructを取得
func TweetGetLatest(w http.ResponseWriter, r *http.Request) {
  csrf := r.FormValue("csrf")
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  coll := common.DB.TweetDB.Collection("tweet")
	findOpts := options.Find().
		SetProjection(bson.M{"tweetHead": 1, "_id": 1}). // ← tweetHeadだけ
		SetSort(bson.D{{"updatedAt", -1}}).              // 更新日時の降順
		SetLimit(100)

  cursor, err := coll.Find(ctx, bson.M{}, findOpts)
  if err != nil {
    common.WriteResponseWithoutSession(w, csrf, "DB検索エラー:"+err.Error(), http.StatusOK) 
    return
  }
  defer cursor.Close(ctx)
  var tweets []collection.TweetStruct
  if err := cursor.All(ctx, &tweets); err != nil {
    common.WriteResponseWithoutSession(w, csrf, "cursor.All:"+err.Error(), http.StatusOK) 
    return
  }
  responseData := struct {
    Csrf         string                   `json:"csrf"`
    PushContents []string                 `json:"pushContents"`
    Tweets       []collection.TweetStruct `json:"tweets"`
  }{
    Csrf:         csrf,
    PushContents: []string{},
    Tweets:       tweets,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
