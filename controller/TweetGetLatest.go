package controller

import (
  "context"
  "encoding/json"
  "log"
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

  // === セッション確認 ===
  session, err := common.SessionCheckTake(w, r, csrf)
  if err != nil {
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  coll := common.DB.TweetDB.Collection("tweet")

  // === MongoDBクエリ ===
  // findOpts := options.Find().
  //   SetSort(bson.M{"updatedAt": -1}). // UpdatedAt降順（新しい順）
  //   SetLimit(100)

	findOpts := options.Find().
		SetProjection(bson.M{"tweetHead": 1, "_id": 1}). // ← tweetHeadだけ
		SetSort(bson.D{{"updatedAt", -1}}).              // 更新日時の降順
		SetLimit(100)

  cursor, err := coll.Find(ctx, bson.M{}, findOpts)
  if err != nil {
    common.WriteResponseWithSession(w, session, "DB検索エラー:"+err.Error(), http.StatusInternalServerError)
    return
  }
  defer cursor.Close(ctx)

  var tweets []collection.TweetStruct
  if err := cursor.All(ctx, &tweets); err != nil {
    common.WriteResponseWithSession(w, session, "デコードエラー:"+err.Error(), http.StatusInternalServerError)
    return
  }

  log.Printf("✅ 最新Tweet %d件取得", len(tweets))

  // === レスポンス構築 ===
  responseData := struct {
    Csrf         string                   `json:"csrf"`
    PushContents []string                 `json:"pushContents"`
    Tweets       []collection.TweetStruct `json:"tweets"`
    Nickname     string                   `json:"nickname"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Tweets:       tweets,
    Nickname:     session.Nickname,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
