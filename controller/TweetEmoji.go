package controller

import (
  "context"
  "encoding/json"
  "net/http"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"

  "chat/common"
  "chat/collection"
)

// /TweetEmoji/ : 絵文字の追加・削除を行うAPI
func TweetEmoji(w http.ResponseWriter, r *http.Request) {
  ctx := context.Background()
  r.ParseMultipartForm(10 << 20)

  coll := common.DB.TweetDB.Collection("tweet")

  parentID := r.FormValue("parentID")   // 親ドキュメントID
  tweetID := r.FormValue("tweetID")     // 絵文字を付ける対象のtweet
  aliasName := r.FormValue("aliasName") // 操作するユーザー
  emojiValue := r.FormValue("emoji")    // 絵文字
  deleteFlag := r.FormValue("delete") == "true"

  if parentID == "" || tweetID == "" || aliasName == "" || emojiValue == "" {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "missing parameters", http.StatusOK)
    return
  }

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";session check some error", http.StatusOK)
		return
	}

  // 該当tweetを含むドキュメントを取得
  var tweetDoc collection.TweetStruct
  err = coll.FindOne(ctx, bson.M{"_id": parentID}).Decode(&tweetDoc)
  if err != nil {
    if err == mongo.ErrNoDocuments {
      common.WriteResponseWithSession(w, session, err.Error()+";tweet", http.StatusOK)
      return
    }
    common.WriteResponseWithSession(w, session, err.Error()+";tweet other error", http.StatusOK)
    return
  }

  found := false
  for i, t := range tweetDoc.Tweets {
    if t.MessageID == tweetID {
      found = true
      if deleteFlag {
        // 削除処理
        newEmojis := []collection.Emoji{}
        for _, e := range t.Emojis {
          if !(e.AliasName == aliasName && e.Emoji == emojiValue) {
            newEmojis = append(newEmojis, e)
          }
        }
        tweetDoc.Tweets[i].Emojis = newEmojis
      } else {
        // 追加処理（重複防止）
        exists := false
        for _, e := range t.Emojis {
          if e.AliasName == aliasName && e.Emoji == emojiValue {
            exists = true
            break
          }
        }
        if !exists {
          tweetDoc.Tweets[i].Emojis = append(tweetDoc.Tweets[i].Emojis,
            collection.Emoji{AliasName: aliasName, Emoji: emojiValue})
        }
      }
      break
    }
  }

  if !found {
    common.WriteResponseWithSession(w, session, "tweet not found", http.StatusOK)
    return
  }

  // 更新（ドキュメント全体置き換え）
  _, err = coll.ReplaceOne(ctx, bson.M{"_id": parentID}, tweetDoc)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";tweet ReplaceOne", http.StatusOK)
    return
  }
	responseData := common.BaseResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		// Mail:         session.Mail,
		// Telephone:    session.Telephone,
		// Nickname:     session.Nickname,
	}
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
