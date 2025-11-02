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

func TweetEmoji(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	messageID := r.FormValue("messageID")
	emoji := r.FormValue("emoji")
	parentID := r.FormValue("parentID")
	csrf := r.FormValue("csrf")

	if messageID == "" || emoji == "" {
		common.WriteResponseWithoutSession(w, csrf, "messageIDまたはemojiが指定されていません。", http.StatusOK)
		return
	}

	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, err.Error()+";session check error", http.StatusOK)
		return
	}

	coll := common.DB.TweetDB.Collection("tweet")
	var tweetDoc collection.TweetStruct

	err = coll.FindOne(ctx, bson.M{"_id": parentID}).Decode(&tweetDoc)
	if err != nil {
		common.WriteResponseWithSession(w, session, "スレッドが見つかりません:"+err.Error(), http.StatusOK)
		return
	}

	for _, b := range tweetDoc.TweetHead.BlockNames {
		if b == session.Nickname {
			common.WriteResponseWithSession(w, session, "あなたはブロックされています", http.StatusOK)
			return
		}
	}

	var message collection.Tweet
	found := false
	emojiCountUpDown := 0
	tweeter := ""
	for i, t := range tweetDoc.Tweets {
		if t.MessageID == messageID {
			found = true
			emojis := t.Emojis
			index := -1

			for j, e := range emojis {
				if e.Emoji == emoji && e.AliasName == session.Nickname {
					index = j
					break
				}
			}

			if index >= 0 {
				emojis = append(emojis[:index], emojis[index+1:]...)
				if (emoji == "👍" || emoji == "👎" || emoji == "🚫") && session.Nickname != t.Nickname {
					emojiCountUpDown = -1
				}
			} else {
				emojis = append(emojis, collection.Emoji{
					AliasName: session.Nickname,
					Emoji:     emoji,
				})
				if (emoji == "👍" || emoji == "👎" || emoji == "🚫") && session.Nickname != t.Nickname {
					emojiCountUpDown = 1
				}
			}
			tweetDoc.Tweets[i].Emojis = emojis
			message = tweetDoc.Tweets[i]
			tweeter = t.Nickname
			break
		}
	}

	if !found {
		common.WriteResponseWithSession(w, session, "指定されたメッセージが見つかりません。", http.StatusOK)
		return
	}

	if emojiCountUpDown != 0 {
		nickColl := common.DB.NicknameDB.Collection("nickname")
		filter := bson.M{"_id": tweeter}
		var existingNick collection.NicknameStruct
		err := nickColl.FindOne(ctx, filter).Decode(&existingNick)
		if err == nil {
			updateFields := bson.M{}
			switch emoji {
			case "👍":
				updateFields["good"] = existingNick.Good + emojiCountUpDown
			case "👎":
				updateFields["bad"] = existingNick.Bad + emojiCountUpDown
			case "🚫":
				updateFields["report"] = existingNick.Report + emojiCountUpDown
			}

			update := bson.M{"$set": updateFields}
			_, err := nickColl.UpdateOne(ctx, filter, update)
			if err != nil {
				common.WriteResponseWithSession(w, session, "nickname更新エラー:"+err.Error(), http.StatusOK)
				return
			}
		}
	}

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

	responseData := struct {
		Csrf           string   `json:"csrf"`
		PushContents   []string `json:"pushContents"`
		Message          collection.Tweet   `json:"message"`
	}{
		Csrf:           session.Csrf,
		PushContents:   session.PushContents,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
