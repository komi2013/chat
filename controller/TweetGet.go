package controller

import (
	"context"
	"encoding/json"
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
	csrf := r.FormValue("csrf")
	if parentID == "" {
		common.WriteResponseWithoutSession(w, csrf, "parentID is required", http.StatusOK)
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
	session, _ := common.SessionCheckTake(w, r, csrf)
	// if err != nil {
	// 	// common.WriteResponseWithoutSession(w, csrf, err.Error()+";SessionCheckTake", http.StatusOK)
	// 	// return
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := common.DB.TweetDB.Collection("tweet")
	var tweetDoc collection.TweetStruct
	err := coll.FindOne(ctx, bson.M{"_id": parentID}).Decode(&tweetDoc)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, err.Error()+";tweet not found", http.StatusOK)
		return
	}
	nickname := session.Nickname
	var filteredTweets []collection.Tweet
	filteredHead := tweetDoc.TweetHead
	if filteredHead.Nickname != filteredHead.HiddenName {
		nickname = filteredHead.Nickname
		filteredHead.AnonymousFlag = true
	}
	alreadyJoinFlag := false
	for _, n := range filteredHead.HiddenNames {
		if session.Nickname == n {
			alreadyJoinFlag = true
		}
	}
	filteredHead.UserID = ""
	filteredHead.HiddenName = ""
	filteredHead.HiddenNames = nil
	filteredHead.Emojis = removeUserID(filteredHead.Emojis)
	for _, t := range tweetDoc.Tweets {
		if t.HiddenName != t.Nickname {
			t.AnonymousFlag = true
		}
		t.UserID = ""
		t.HiddenName = ""
		t.Emojis = removeUserID(t.Emojis)
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
				Nicknames:  []string{nickname},
				BackID:     t.BackID,
				CreatedAt:  t.CreatedAt,
				Emojis:     t.Emojis,
				MessageID:  t.MessageID,
				// HiddenName: "", // ← HiddenName を空文字に設定
				// HiddenNames: []string{t.Nickname},
			}
			filteredHead = head
		}
	}

	start := skip
	end := skip + limit + 1

	// === messageID 指定があれば、その tweet を逆順の先頭にする ===
	messageID := r.FormValue("messageID")
	if messageID != "" {
	    idx := -1
	    for i, t := range filteredTweets {
	        if t.MessageID == messageID {
	            idx = i
	            break
	        }
	    }

	    if idx != -1 {
	        total := len(filteredTweets)

	        // ★逆順位置に変換（元ロジックに合わせる）
	        revPos := total - 1 - idx

	        start = revPos
	        end = revPos + limit + 1
	    }
	}

	if skip < 0 {
	    // 全件取得なら slice 不使用
	} else {
	    total := len(filteredTweets)

	    if start >= total {
	        filteredTweets = []collection.Tweet{}
	    } else {
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

	responseData := struct {
		Csrf         string                 `json:"csrf"`
		PushContents []string               `json:"pushContents"`
		Tweet        collection.TweetStruct `json:"tweet"`
		Nickname     string                 `json:"nickname"`
		TweetPosts   []collection.TweetPost `json:"tweetPosts"`
		AlreadyJoinFlag bool `json:"alreadyJoinFlag"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Tweet: collection.TweetStruct{
			ID:        tweetDoc.ID,
			Tweets:    filteredTweets,
			TweetHead: filteredHead,
			UpdatedAt: tweetDoc.UpdatedAt,
		},
		Nickname:   nickname,
		TweetPosts: session.TweetPosts,
		AlreadyJoinFlag: alreadyJoinFlag,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func removeUserID(emojis []collection.Emoji) []collection.Emoji {
	if len(emojis) == 0 {
		return emojis
	}

	cleaned := make([]collection.Emoji, len(emojis))
	for i, e := range emojis {
		e.UserID = "" // ← UserIDを削除（空文字に）
		cleaned[i] = e
	}
	return cleaned
}
