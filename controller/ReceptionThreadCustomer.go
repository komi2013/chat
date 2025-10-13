package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"chat/collection"
	"chat/common"
)

func ReceptionThreadCustomer(w http.ResponseWriter, r *http.Request) {
	channelID := r.FormValue("channelID")
	messageID := r.FormValue("messageID")
	parentID := r.FormValue("parentID")
	messageTxt := r.FormValue("messageTxt")
	// threadHead := r.FormValue("threadHead")
	threadHeadNew := false
	var threadHead collection.ThreadHeadStruct
	if r.FormValue("threadHead") != "" {
		if err := json.Unmarshal([]byte(r.FormValue("threadHead")), &threadHead); err != nil {
			common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "threadHead Invalid JSON", http.StatusOK)
			return
		}
		threadHeadNew = true
	}

	// セッションチェック
	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "。サインインし直してください", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// receptionデータを取得してOrderUserIDsを取得
	var reception collection.ReceptionStruct
	receptionColl := common.DB.ReceptionDB.Collection("reception")
	receptionFilter := bson.M{"_id": channelID}
	err = receptionColl.FindOne(ctx, receptionFilter).Decode(&reception)
	if err != nil {
		log.Printf("receptionColl.FindOne: %v; Req: ", err, r.URL.Path, r.Form)
		http.Error(w, "Failed to get reception data", http.StatusInternalServerError)
		return
	}

	threadHead.AliasNames = append(threadHead.AliasNames, reception.JoinNames...)
	threadHead.AdminNames = reception.JoinNames
	threadHead.AliasImg   = session.NickImg

	threadHeadArray := []interface{}{
		"threadHead",    
		channelID,       
		session.Nickname, //updatedBy
		threadHead,      //
	}

	// aliasデータの準備
	aliasData := []interface{}{
		session.UserID,  // [0] userID
		session.Nickname, // [1] aliasName
		"",              // [2] bio
		"inquirer",      // [3] accessRight
	}

	// alias用のpush配列
	aliasPushArray := []interface{}{
		"alias",         // [0] pushTitle
		channelID,       // [1] channelID
		session.Nickname, // [2] updatedBy
		aliasData,       // [3] aliasData
		session.NickImg, // [4] aliasImg
		1,                // inquiry flag
	}

	// threadデータの準備
	threadData := []interface{}{
		parentID,          // [0] parentID
		messageID,         // [1] messageID
		messageTxt,        // [2] messageTxt
		session.NickImg,  // [3] aliasImg
		[]string{session.Nickname}, // [4] aliasNames
		"",              // [5] backID
		[]interface{}{}, // [6] emojis
		session.Nickname, // [7] aliasName
	}

	// push配列の準備
	pushArray := []interface{}{
		"thread",        // [0] pushTitle
		channelID,       // [1] channelID
		session.Nickname, // [2] updatedBy
		threadData,      // [3] threadData
		[]string{},      // [4] fileLinks
	}

	pushTest, _ := json.Marshal(threadHeadArray)
	log.Printf("threadHeadArray JSON: %s", string(pushTest))

	// チャンネルコレクションの取得（1回でOK）
	collChannel := common.DB.ChannelDB.Collection("channel")

	// 1️⃣ 既存チャンネルの取得
	var channel collection.ChannelStruct
	filterChannel := bson.M{"_id": channelID}
	err = collChannel.FindOne(ctx, filterChannel).Decode(&channel)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	// 2️⃣ 新しいAliasを作成
	newAlias := collection.Alias{
		AliasName:   session.Nickname,
		AliasImg:    session.NickImg,
		UserID:      session.UserID,
		Bio:         "",
		AccessRight: "inquirer",
	}

	// 3️⃣ MongoDBへ更新
	update := bson.M{
		"$push": bson.M{
			"aliases":    newAlias,
		},
	}

	_, err = collChannel.UpdateOne(ctx, filterChannel, update)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("alias登録失敗: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Alias '%s' (userID: %s) を channel '%s' に登録しました",
		session.Nickname, session.UserID, channelID)


	uniqueIDs := make(map[string]struct{})
	var userIDs []string
	pushNameSet := make(map[string]struct{}, len(reception.InquiryNames))
	for _, name := range reception.InquiryNames {
		pushNameSet[name] = struct{}{}
	}
	for _, alias := range channel.Aliases {
		if _, ok := pushNameSet[alias.AliasName]; ok {
			if _, exists := uniqueIDs[alias.UserID]; !exists {
				uniqueIDs[alias.UserID] = struct{}{}
				userIDs = append(userIDs, alias.UserID)
			}
		}
	}
	userIDs = append(userIDs, session.UserID) // カスタマー側のセッションを追加
	log.Printf("Filtered userIDs (matched pushNames): %v", userIDs)



	// セッションを取得してプッシュ
	sessionColl := common.DB.SessionDB.Collection("session")
	sessionFilter := bson.D{{"userID", bson.D{{"$in", userIDs}}}}
	cursor, err := sessionColl.Find(ctx, sessionFilter)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}
	defer cursor.Close(ctx)

	var sessions []collection.SessionStruct
	if err = cursor.All(ctx, &sessions); err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	filteredSessions := common.FilterSessionsByChannelID(sessions, channelID)

	var pushTargetSessions []collection.SessionStruct
	for _, s := range filteredSessions {
		if s.UserID != session.UserID {
			pushTargetSessions = append(pushTargetSessions, s)
		}
	}
	common.ChunkPush(pushTargetSessions, aliasPushArray)

	if threadHeadNew {
		common.ChunkPush(filteredSessions, threadHeadArray)
	} else {
		common.ChunkPush(filteredSessions, pushArray)
	}

	responseData := common.BaseResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
