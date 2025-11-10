package controller

import (
  "chat/common"
  "context"
  "encoding/json"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/bson"
)

func AdDelete(w http.ResponseWriter, r *http.Request) {
  csrf := r.FormValue("csrf")
  adID := r.FormValue("adID")

  if adID == "" {
    common.WriteResponseWithoutSession(w, csrf, "adIDが指定されていません", http.StatusOK)
    return
  }

  // セッション確認
  session, err := common.SessionCheckTake(w, r, csrf)
  if err != nil {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+"; セッション確認エラー", http.StatusOK)
    return
  }

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  coll := common.DB.AdDB.Collection("ad")

  // 対象データの存在確認
  filter := bson.M{
    "adID":   adID,
    "userID": session.UserID,
  }
  var existing bson.M
  err = coll.FindOne(ctx, filter).Decode(&existing)
  if err != nil {
    common.WriteResponseWithSession(w, session, "指定された広告が存在しません", http.StatusOK)
    return
  }

  // 削除実行
  _, err = coll.DeleteOne(ctx, filter)
  if err != nil {
    log.Printf("AdDelete DeleteOne error: %v", err)
    common.WriteResponseWithSession(w, session, "削除中にエラーが発生しました: "+err.Error(), http.StatusOK)
    return
  }

  // 正常終了レスポンス
  responseData := struct {
    Csrf         string   `json:"csrf"`
    PushContents []string `json:"pushContents"`
    Message      string   `json:"message"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Message:      "広告を削除しました",
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
