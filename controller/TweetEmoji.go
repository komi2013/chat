package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/common"
	"chat/collection"
)

// TweetEmoji : Tweetに絵文字を登録・編集する
func TweetEmoji(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	messageID := r.FormValue("messageID") // 対象TweetのMessageID
	emoji := r.FormValue("emoji")         // 絵文字（例: "👍"）
	parentID := r.FormValue("parentID")   // スレッド親ID
	csrf := r.FormValue("csrf")

	if messageID == "" || emoji == "" {
		common.WriteResponseWithoutSession(w, csrf, "messageIDまたはemojiが指定されていません。", http.StatusOK)
		return
	}

	// === セッションチェック ===
	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, err.Error()+";session check error", http.StatusOK)
		return
	}

	coll := common.DB.TweetDB.Collection("tweet")
	var tweetDoc collection.TweetStruct

	// 親スレッド取得
	err = coll.FindOne(ctx, bson.M{"_id": parentID}).Decode(&tweetDoc)
	if err != nil {
		common.WriteResponseWithSession(w, session, "スレッドが見つかりません:"+err.Error(), http.StatusOK)
		return
	}

	// === 対象ツイートを探索 ===
	var message collection.Tweet
	found := false
	for i, t := range tweetDoc.Tweets {
		if t.MessageID == messageID {
			found = true
			emojis := t.Emojis
			index := -1

			// === 同じemoji + ニックネームが存在するか確認 ===
			for j, e := range emojis {
				if e.Emoji == emoji && e.AliasName == session.Nickname {
					index = j
					break
				}
			}

			if index >= 0 {
				// === 既に存在している場合は削除 ===
				emojis = append(emojis[:index], emojis[index+1:]...)
			} else {
				// === 存在しない場合は追加 ===
				emojis = append(emojis, collection.Emoji{
					AliasName: session.Nickname,
					Emoji:     emoji,
				})
			}

			// 更新を反映
			tweetDoc.Tweets[i].Emojis = emojis
			message = tweetDoc.Tweets[i]
			break
		}
	}


	if !found {
		common.WriteResponseWithSession(w, session, "指定されたメッセージが見つかりません。", http.StatusOK)
		return
	}

	// === MongoDBに反映 ===
	_, err = coll.UpdateOne(ctx,
		bson.M{"_id": parentID},
		bson.M{"$set": bson.M{
			"tweets": tweetDoc.Tweets,
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		common.WriteResponseWithSession(w, session, "DB更新エラー:"+err.Error(), http.StatusOK)
		return
	}

	// ======== Response ========
	responseData := struct {
		Csrf           string   `json:"csrf"`
		PushContents   []string `json:"pushContents"`
		Message          collection.Tweet   `json:"message"`
		// Nickname       string   `json:"nickname"`
	}{
		Csrf:           session.Csrf,
		PushContents:   session.PushContents,
		Message: message,
		// Nickname: session.Nickname,

	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
