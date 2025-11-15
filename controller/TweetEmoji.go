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

// --- 👇 TweetEmoji本体 ---
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

	for _, b := range tweetDoc.TweetHead.BlockUserIDs {
		if b == session.UserID {
			common.WriteResponseWithSession(w, session, "あなたはブロックされています", http.StatusOK)
			return
		}
	}
	iAmAdmin := tweetDoc.TweetHead.UserID == session.UserID
	var message collection.Tweet
	found := false
	// deleteFlag := false
	emojiCountUpDown := 0
	tweeter := ""

	// --- Tweetの中を探索 ---
	for i, t := range tweetDoc.Tweets {
		if t.MessageID == messageID {
			found = true
			newEmojis, countChange := toggleEmoji(t.Emojis, emoji, session.Nickname, t.Nickname, session.UserID)
			isMyTweet := session.UserID == t.UserID
			reportCount := len(newEmojis)
			shouldDelete := false
			if emoji == "⚠️" && reportCount >= 5 {
				shouldDelete = true
			}
			if emoji == "🗑" && (isMyTweet || iAmAdmin) {
				shouldDelete = true
			}
	    if shouldDelete {
				newTweets := make([]collection.Tweet, 0, len(tweetDoc.Tweets)-1)
				for j, tw := range tweetDoc.Tweets {
				    if j != i {
				        newTweets = append(newTweets, tw)
				    }
				}
				tweetDoc.Tweets = newTweets
	    } else {
				tweetDoc.Tweets[i].Emojis = newEmojis
				emojiCountUpDown = countChange
				message = tweetDoc.Tweets[i]

				tweeter = t.Nickname
				if t.Nickname != t.HiddenName {
					tweeter = t.HiddenName
				}
	    }
	    if emoji == "🚫" && iAmAdmin {
		    tweetDoc.TweetHead.BlockUserIDs = append(tweetDoc.TweetHead.BlockUserIDs, t.UserID)
	    }
			break
		}
	}

	if !found && tweetDoc.TweetHead.MessageID == messageID {
		found = true
		newEmojis, countChange := toggleEmoji(tweetDoc.TweetHead.Emojis, emoji, session.Nickname, tweetDoc.TweetHead.Nickname, session.UserID)
		reportCount := len(newEmojis)
		shouldDelete := false
		if emoji == "⚠️" && reportCount >= 5 {
			shouldDelete = true
		}
		if emoji == "🗑" && iAmAdmin {
			shouldDelete = true
		}
    if shouldDelete {
        _, err = coll.DeleteOne(ctx, bson.M{"_id": parentID})
        if err != nil {
            common.WriteResponseWithSession(w, session, "DB削除エラー:"+err.Error(), http.StatusOK)
            return
        }
        common.WriteResponseWithSession(w, session, "このスレッドは削除されました", http.StatusOK)
        return
    }
		tweetDoc.TweetHead.Emojis = newEmojis
		emojiCountUpDown = countChange

		message = collection.Tweet{
			MessageID: tweetDoc.TweetHead.MessageID,
			Emojis:    tweetDoc.TweetHead.Emojis,
			Nickname:  tweetDoc.TweetHead.Nickname,
		}
		tweeter = tweetDoc.TweetHead.Nickname
		if tweetDoc.TweetHead.Nickname != tweetDoc.TweetHead.HiddenName {
			tweeter = tweetDoc.TweetHead.HiddenName
		}
	}

	if !found {
		common.WriteResponseWithSession(w, session, "指定されたメッセージが見つかりません。", http.StatusOK)
		return
	}

	// 👍👎⚠️ によるスコア変動
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
			case "⚠️":
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

	// --- DB更新 ---
	_, err = coll.UpdateOne(ctx,
		bson.M{"_id": parentID},
		bson.M{"$set": bson.M{
			"tweets":     tweetDoc.Tweets,
			"tweetHead":  tweetDoc.TweetHead,
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		common.WriteResponseWithSession(w, session, "DB更新エラー:"+err.Error(), http.StatusOK)
		return
	}

	responseData := struct {
		Csrf         string             `json:"csrf"`
		PushContents []string           `json:"pushContents"`
		Message      collection.Tweet   `json:"message"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Message:      message,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func toggleEmoji(
	emojis []collection.Emoji,
	emoji string,
	sessionNickname string,
	targetNickname string,
	userID string,
	) ([]collection.Emoji, int) {
	index := -1
	for j, e := range emojis {
		if e.Emoji == emoji && e.UserID == userID {
			index = j
			break
		}
	}
	emojiCountUpDown := 0
	if index >= 0 {
		emojis = append(emojis[:index], emojis[index+1:]...)
		if (emoji == "👍" || emoji == "👎" || emoji == "🚫") && sessionNickname != targetNickname {
			emojiCountUpDown = -1
		}
	} else {
		emojis = append(emojis, collection.Emoji{
			AliasName: sessionNickname,
			Emoji:     emoji,
			UserID:    userID,
		})
		if (emoji == "👍" || emoji == "👎" || emoji == "🚫") && sessionNickname != targetNickname {
			emojiCountUpDown = 1
		}
	}
	return emojis, emojiCountUpDown
}
