package controller

import (
	"context"
	"encoding/json"
	// "math/rand"
	"log"
	"net/http"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/common"
	"chat/collection"
)

// TweetPost : 新しいTweetを投稿する
func TweetPost(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// === リクエストパラメータ取得 ===
	parentID := r.FormValue("parentID")
	parentMessageID := r.FormValue("parentMessageID")
	messageTxt := r.FormValue("messageTxt")

	if len(messageTxt) == 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "本文が空です。", http.StatusOK)
		return
	}
	if len([]rune(messageTxt)) > 500 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "文字数が500を超えています。", http.StatusOK)
		return
	}

	// === セッションチェック ===
	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";session check error", http.StatusOK)
		return
	}

	coll := common.DB.TweetDB.Collection("tweet")
	var tweetDoc collection.TweetStruct

	// parentID が指定されているときのみ既存スレッドを取得
	if parentID != "" {
		err = coll.FindOne(ctx, bson.M{"_id": parentID}).Decode(&tweetDoc)
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+";tweet not found", http.StatusOK)
			return
		}
	} else {
		parentID = common.StringRand(8)
		tweetDoc = collection.TweetStruct{
			ID:         parentID,
			TweetHeads: []collection.TweetHead{},
		}		
	}

	now := time.Now()
	userID := session.UserID
	nickname := session.Nickname
	nickImg := session.NickImg

	var userTweets []collection.Tweet
	threadCount := make(map[string]int)
	var lastTweet *collection.Tweet

	// === 投稿制限チェック ===
	for _, t := range tweetDoc.Tweets {
		if t.UserID == userID {
			userTweets = append(userTweets, t)
			threadCount[t.ParentID]++
			if lastTweet == nil || t.CreatedAt > lastTweet.CreatedAt {
				tmp := t
				lastTweet = &tmp
			}
		}
	}

	if len(threadCount) >= 5 {
		if lastTweet != nil {
			lastTime, _ := time.Parse(time.RFC3339, lastTweet.CreatedAt)
			if now.Sub(lastTime) < 20*time.Hour {
				common.WriteResponseWithSession(w, session, "全体で20時間経過していません。投稿できません。", http.StatusOK)
				return
			}
		}
	} else {
		for _, t := range userTweets {
			if t.ParentID == parentID {
				lastTime, _ := time.Parse(time.RFC3339, t.CreatedAt)
				if now.Sub(lastTime) < 20*time.Hour {
					common.WriteResponseWithSession(w, session, "このスレッドでは20時間経過していません。", http.StatusOK)
					return
				}
			}
		}
	}

	// === TweetHead更新 or 新規作成 ===
	headFound := false
	for i, head := range tweetDoc.TweetHeads {
		if head.ParentID == parentID {
			tweetDoc.TweetHeads[i].UpdatedAt = now
			headFound = true
			break
		}
	}

	if headFound {
		// === 新規Tweet作成 ===
		newTweet := collection.Tweet{
			MessageID:  common.StringRand(8),
			ParentID:   parentID,
			MessageTxt: messageTxt,
			Nickname:   nickname,
			NickImg:    nickImg,
			UserID:     userID,
			CreatedAt:  now.Format(time.RFC3339),
			BackID:     "",
			Emojis:     []collection.Emoji{},
		}
		tweetDoc.Tweets = append(tweetDoc.Tweets, newTweet)
	} else {
		title := GenerateTitle(messageTxt)
		newHead := collection.TweetHead{
			ParentID:   parentID,
			MessageID:  parentMessageID,
			Nickname:   nickname,
			NickImg:    nickImg,
			Title:      title,
			MessageTxt: messageTxt,
			UpdatedAt:  now,
			CreatedAt:  now.Format(time.RFC3339),
			NickNames:  []string{nickname},
			Emojis:     []collection.Emoji{},
		}
		tweetDoc.TweetHeads = append(tweetDoc.TweetHeads, newHead)
	}

	// tweetDoc.Tweets = append(tweetDoc.Tweets, newTweet)
	log.Printf("tweetDoc=%s", common.ToJSON(tweetDoc))
	_, err = coll.UpdateOne(ctx,
		bson.M{"_id": parentID},
		bson.M{
			"$set": bson.M{
				"tweets":     tweetDoc.Tweets,
				"tweetHeads": tweetDoc.TweetHeads,
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+";tweet upsert", http.StatusOK)
		return
	}

	// === レスポンス ===
	responseData := common.BaseResponse{
		Csrf:     session.Csrf,
		PushContents: session.PushContents,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func GenerateTitle(messageTxt string) string {
	// HTMLタグ除去
	re := regexp.MustCompile(`<[^>]*>`)
	clean := re.ReplaceAllString(messageTxt, "")

	// 改行などの削除
	clean = regexp.MustCompile(`\r?\n`).ReplaceAllString(clean, "")

	// 不要な記号や装飾などの軽い除去（JSの removeMark 相当の一部）
	clean = regexp.MustCompile(`[*_~>`+"`"+`]`).ReplaceAllString(clean, "")

	// 必要に応じてさらに特殊なMarkdown除去処理を追加可

	// 30文字に制限（マルチバイト対応）
	runes := []rune(clean)
	if len(runes) > 30 {
		return string(runes[:30]) + "…"
	}
	return clean
}
