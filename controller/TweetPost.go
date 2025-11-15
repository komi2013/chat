package controller

import (
	"context"
	"encoding/json"
	// "math/rand"
	// "log"
	"net/http"
	// "regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/common"
	"chat/collection"
)

// TweetPost : 新しいTweetを投稿する
func TweetPost(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	parentID := r.FormValue("parentID")
	backID := r.FormValue("backID")
	messageTxt := r.FormValue("messageTxt")
	if len(messageTxt) == 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "本文が空です。", http.StatusOK)
		return
	}
	if len([]rune(messageTxt)) > 500 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "文字数が500を超えています。", http.StatusOK)
		return
	}

	if r.FormValue("anonymousImg") != "" && !common.EmojiImgValid(r.FormValue("anonymousImg")) {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "画像が不正", http.StatusOK)
		return
	}

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";session check error", http.StatusOK)
		return
	}
	if session.Nickname == "" {
		common.WriteResponseWithSession(w, session, "ニックネームが登録されていません", http.StatusOK)
		return
	}
	nickname := session.Nickname
	if r.FormValue("anonymous") != "" {
		nickname = r.FormValue("anonymous")
	}
	nickImg := session.NickImg
	if r.FormValue("anonymousImg") != "" {
		nickImg = r.FormValue("anonymousImg")
	}

	now := time.Now()
	// restriction post 
	limitedThreads := 0
	for _, post := range session.TweetPosts {
		if now.Sub(post.PostedAt) < 20*time.Hour {
			if !post.PostAdminFlag && post.ParentID == parentID && post.PostCount >= 3 {
				common.WriteResponseWithSession(w, session, "このスレッドでは20時間以内に3回投稿しています", http.StatusOK)
				return
			}
		}
		if post.ParentID != parentID {
			limitedThreads++			
		}
	}

	if limitedThreads >= 3 {
		common.WriteResponseWithSession(w, session, "20時間以内に3つのスレッドで上限投稿しています", http.StatusOK)
		return
	}
	// restriction post 

  imgPath, err := common.ImgSave(r.FormValue("imgPath"), session.UserID, session.Nickname, "", 3, 4)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+";ImgSave", http.StatusOK)
		return
	}
	if imgPath != "" {
		messageTxt = messageTxt + "＊img＊" + imgPath + "・＊img＊"
	}
	// take and make tweet data
	coll := common.DB.TweetDB.Collection("tweet")
	var tweetDoc collection.TweetStruct
	if parentID != "" {
		err = coll.FindOne(ctx, bson.M{"_id": parentID}).Decode(&tweetDoc)
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+";tweet not found", http.StatusOK)
			return
		}
	} else {
		tweetID, err := common.CountUpID("tweetID")
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+" CountUpID error", http.StatusOK)
			return
		}
		parentID = tweetID
		tweetDoc = collection.TweetStruct{
			ID:         parentID,
			TweetHead: collection.TweetHead{},
		}		
	}
	iamAdmin := tweetDoc.TweetHead.UserID == session.UserID
	for _, b := range tweetDoc.TweetHead.BlockUserIDs {
		if b == session.UserID && !iamAdmin {
			common.WriteResponseWithSession(w, session, "あなたはブロックされています", http.StatusOK)
			return
		}
	}
	newTweets := make([]collection.Tweet, 0, len(tweetDoc.Tweets))
	for i, t := range tweetDoc.Tweets {
		if t.MessageID == backID {
			tweetDoc.Tweets[i].TweetCount++
		}
		newTweets = append(newTweets, tweetDoc.Tweets[i])
	}
	tweetDoc.Tweets = newTweets
	parentMessageID := parentID
	if backID != "" {
		parentMessageID = backID
	}
	if tweetDoc.TweetHead.ParentID == parentID { // head found
		newTweet := collection.Tweet{
			MessageID:  common.StringRand(8),
			ParentID:   parentMessageID,
			MessageTxt: messageTxt,
			Nickname:   nickname,
			NickImg:    nickImg,
			CreatedAt:  now.Format(time.RFC3339),
			BackID:     "",
			Emojis:     []collection.Emoji{},
			UserID:     session.UserID,
			HiddenName:  session.Nickname,
		}
		if messageTxt != "" {
			tweetDoc.Tweets = append(tweetDoc.Tweets, newTweet)
		}
		exists := false
		for _, n := range tweetDoc.TweetHead.Nicknames {
			if n == nickname {
				exists = true
				break
			}
		}
		if !exists {
			tweetDoc.TweetHead.Nicknames = append(tweetDoc.TweetHead.Nicknames, nickname)
		}
		exists = false
		for _, n := range tweetDoc.TweetHead.HiddenNames {
			if n == session.Nickname {
				exists = true
				break
			}
		}
		if !exists {
			tweetDoc.TweetHead.HiddenNames = append(tweetDoc.TweetHead.HiddenNames, session.Nickname)
		}
	} else {
		newHead := collection.TweetHead{
			ParentID:   parentID,
			MessageID:  parentID,
			Nickname:   nickname,
			NickImg:    nickImg,
			// Title:      title,
			MessageTxt: messageTxt,
			CreatedAt:  now.Format(time.RFC3339),
			Nicknames:  []string{nickname},
			Emojis:     []collection.Emoji{},
			UserID:     session.UserID,
			HiddenName:  session.Nickname,
			HiddenNames:  []string{session.Nickname},
		}
		// tweetDoc.TweetHeads = append(tweetDoc.TweetHeads, newHead)
		tweetDoc.TweetHead = newHead
	}
	// take and make tweet data
	// take sessions for push from nickname
	var nicknameDocs []collection.NicknameStruct
	nicknameColl := common.DB.NicknameDB.Collection("nickname")
	cursor, err := nicknameColl.Find(ctx, bson.M{"_id": bson.M{"$in": tweetDoc.TweetHead.HiddenNames}})
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" nickname find  error", http.StatusOK)
		return
	}
	if err = cursor.All(ctx, &nicknameDocs); err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" nickname cursor decode error", http.StatusOK)
		return
	}
	userIDs := []string{}
	for _, n := range nicknameDocs {
		userIDs = append(userIDs, n.UserID)
	}
	sessionColl := common.DB.SessionDB.Collection("session")
	var filteredSessions []collection.SessionStruct
	sessCur, err := sessionColl.Find(ctx, bson.M{"userID": bson.M{"$in": userIDs}})
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" session find error", http.StatusOK)
		return
	}
	if err = sessCur.All(ctx, &filteredSessions); err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" session cursor decode error", http.StatusOK)
		return
	}
	liteHead := collection.TweetHeadLite{
		ParentID:   tweetDoc.TweetHead.ParentID,
		MessageTxt: tweetDoc.TweetHead.MessageTxt,
		CreatedAt:  tweetDoc.TweetHead.CreatedAt,
	}
	tweetHeadArray := []interface{}{
		"tweetHead",        // Push識別子
		parentID,           // parentID
		session.Nickname,   // 更新者（SessionStructのNickname）
		liteHead, // 通知データ本体
	}
	// take sessions for push from nickname

	// update data
	_, err = coll.UpdateOne(ctx,
		bson.M{"_id": parentID},
		bson.M{
			"$set": bson.M{
				"tweets":     tweetDoc.Tweets,
				"tweetHead": tweetDoc.TweetHead,
				"updatedAt": now,
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+";tweet upsert", http.StatusOK)
		return
	}

	postAdminFlag := false
	if tweetDoc.TweetHead.Nickname == session.Nickname || tweetDoc.TweetHead.HiddenName == session.Nickname {
		postAdminFlag = true
	}

	collSession := common.DB.SessionDB.Collection("session")
	var sessionDoc collection.SessionStruct
	err = collSession.FindOne(ctx, bson.M{"userID": session.UserID}).Decode(&sessionDoc)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+";session not found", http.StatusOK)
		return
	}

	// === TweetPosts 更新処理 ===
	var updatedPosts []collection.TweetPost
	found := false
	for _, post := range sessionDoc.TweetPosts {
		// 20時間以上前の投稿は削除
		if now.Sub(post.PostedAt) >= 20*time.Hour {
			continue
		}

		if post.ParentID == parentID {
			found = true
			post.PostCount++
			post.PostAdminFlag = postAdminFlag
			post.PostedAt = now
		}
		updatedPosts = append(updatedPosts, post)
	}
	// log.Printf("updatedPosts=%s", common.ToJSON(updatedPosts))
	// 新規スレッド投稿の場合
	if !found {
		newPost := collection.TweetPost{
			ParentID:      parentID,
			PostCount:     1,
			PostAdminFlag: postAdminFlag,
			PostedAt:      now,
		}
		updatedPosts = append(updatedPosts, newPost)
	}

	// === MongoDBに更新（Upsert相当） ===
	_, err = collSession.UpdateOne(ctx,
		bson.M{"userID": session.UserID},
		bson.M{"$set": bson.M{"tweetPosts": updatedPosts}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+";tweetPosts update failed", http.StatusOK)
		return
	}
	// update data

	common.ChunkPush(filteredSessions, tweetHeadArray)

	responseData := struct {
		Csrf         string                 `json:"csrf"`
		PushContents []string               `json:"pushContents"`
		Date     string                 `json:"date"`
		ParentID     string                 `json:"parentID"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Date:         now.Format("20060102"),
		ParentID:     parentID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

