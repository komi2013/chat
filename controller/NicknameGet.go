package controller

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    // "go.mongodb.org/mongo-driver/mongo/options"

    "chat/common"
    "chat/collection"
)

// GetNickname : ニックネーム情報取得API
func NicknameGet(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    csrf := r.FormValue("csrf")
    nickname := r.FormValue("nickname")
    if nickname == "" {
        common.WriteResponseWithoutSession(w, csrf, "nickname is required", http.StatusOK)
        return
    }

    session, err := common.SessionCheckTake(w, r, csrf)
    if err != nil {
    	common.WriteResponseWithoutSession(w, csrf, err.Error(), http.StatusOK)
      return
    }

    coll := common.DB.NicknameDB.Collection("nickname")
    filter := bson.M{"_id": nickname}
    var nickDoc collection.NicknameResponse
    err = coll.FindOne(ctx, filter).Decode(&nickDoc)
    if err != nil {
        common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
        return
    }

    response := struct {
        Csrf     string           `json:"csrf"`
        NickData collection.NicknameResponse `json:"nickData"`
    }{
        Csrf:     session.Csrf,
        NickData: nickDoc,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
