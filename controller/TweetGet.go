package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"chat/common"
	"chat/collection"
)

func TweetGet(w http.ResponseWriter, r *http.Request) {
	parentID := r.FormValue("parentID")
	backID := r.FormValue("backID")
	// skipStr := r.FormValue("skip")
	csrf := r.FormValue("csrf")
	if parentID == "" {
		common.WriteResponseWithoutSession(w, csrf, "parentID is required", http.StatusBadRequest)
		return
	}

	// === パラメータ処理 ===
	skip := 0
	if r.FormValue("skip") != "" {
		if s, err := strconv.Atoi(r.FormValue("skip")); err == nil && s >= -1 {
			skip = s
		}
	}
	const limit = 10

	// === セッションチェック ===
	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		log.Printf("SessionCheckTake: %v; Req: %v", err, r.URL.Path)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := common.DB.TweetDB.Collection("tweet")

	// === ドキュメント取得 ===
	var tweetDoc collection.TweetStruct
	err = coll.FindOne(ctx, bson.M{"_id": parentID}).Decode(&tweetDoc)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+";tweet not found", http.StatusOK)
		return
	}

	// === 絞り込み処理 ===
	var filteredTweets []collection.Tweet
	// var filteredHead collection.TweetHead

	// if backID != "" {
	// 	// log.Printf("tweetDoc.Tweets: %s", common.ToJSON(tweetDoc.Tweets))

	// } else {
	// 	// 全スレッド
	// 	filteredTweets = tweetDoc.Tweets
		
	// }
	filteredHead := tweetDoc.TweetHead
	for _, t := range tweetDoc.Tweets {
		// log.Printf("Tweets: %s", common.ToJSON(t))
		if t.ParentID == backID {
			filteredTweets = append(filteredTweets, t)
		} else if backID == "" && t.ParentID == parentID {
			filteredTweets = append(filteredTweets, t)
		}
		if t.MessageID == backID {
			head := collection.TweetHead{
				ParentID:   t.ParentID,
				MessageTxt: t.MessageTxt,
				Nickname:   t.Nickname,
				NickImg:    t.NickImg,
				// UpdatedAt:  time.Now(),
				Nicknames:  tweetDoc.TweetHead.Nicknames,
				BackID:     t.BackID,
				CreatedAt:  t.CreatedAt,
				Emojis:     t.Emojis,
				MessageID:  t.MessageID,
			}
			filteredHead = head
		}
	}
	// log.Printf("filteredTweets: %s", common.ToJSON(filteredTweets))
	// === ページング (skip, limit) ===
	start := skip
	end := skip + limit + 1

	if skip < 0 {
		// --- 全件取得（そのまま） ---
		// filteredTweets = filteredTweets
	} else {
		total := len(filteredTweets)

		// --- 範囲外チェック ---
		if start >= total {
			filteredTweets = []collection.Tweet{}
		} else {
			// --- 後ろからのインデックス計算 ---
			revStart := total - end
			if revStart < 0 {
				revStart = 0
			}
			revEnd := total - start
			if revEnd > total {
				revEnd = total
			}
			filteredTweets = filteredTweets[revStart:revEnd]
		}
	}

	// === レスポンス形式統一 ===
	responseData := struct {
		Csrf         string                 `json:"csrf"`
		PushContents []string               `json:"pushContents"`
		Tweet        collection.TweetStruct `json:"tweet"`
		Nickname     string                 `json:"nickname"`
		TweetPosts   []collection.TweetPost `json:"tweetPosts"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Tweet: collection.TweetStruct{
			ID:         tweetDoc.ID,
			Tweets:     filteredTweets,
			TweetHead: filteredHead,
		},
		Nickname: session.Nickname,
		TweetPosts: session.TweetPosts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
