package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/collection"
	"chat/common"
)

func ChannelAdd(w http.ResponseWriter, r *http.Request) {
	myname := r.FormValue("myname")
	myimg := r.FormValue("myimg")
	channelName := r.FormValue("channelName")
	channelDescription := r.FormValue("channelDescription")

	// === セッション確認 ===
	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+" session check some error", http.StatusOK)
		return
	}

	collUser := common.DB.UserDB.Collection("user")
	filterUser := bson.M{"_id": session.UserID}

	// === ① すべての Find を最初に行う ===

	// 🧩 ユーザー情報取得
	var user collection.UserStruct
	err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" user FindOne", http.StatusOK)
		return
	}

	// チャンネル作成上限チェック
	uniqueChannels := make(map[string]struct{})
	for _, alias := range user.ChannelAliases {
		uniqueChannels[alias.ChannelID] = struct{}{}
	}
	if len(uniqueChannels) >= 3 {
		common.WriteResponseWithSession(w, session, "すでにチャネル作成の上限です", http.StatusOK)
		return
	}

	// 🧩 ユーザーの全セッション情報取得
	collSession := common.DB.SessionDB.Collection("session")
	filter := bson.M{"userID": session.UserID}
	project := bson.D{{"updatedAt", 0}}
	opts := options.Find().SetProjection(project)

	cursor, err := collSession.Find(context.TODO(), filter, opts)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" session Find", http.StatusOK)
		return
	}

	var mySessions []collection.SessionStruct
	if err = cursor.All(context.TODO(), &mySessions); err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" session cursor.All", http.StatusOK)
		return
	}

	// === ② 登録・更新処理 ===

	// チャンネルID採番
	channelID, err := common.CountUpID("channelID")
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" CountUpID error", http.StatusOK)
		return
	}

	// アイコン画像保存
	aliasImg, err := common.ImgSave(myimg, session.UserID, myname, channelID, 0, 1)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" image save error", http.StatusOK)
		return
	}

	// === 🧩 新しいチャンネル構造体作成 ===
	newChannel := collection.ChannelStruct{
		ChannelID:          channelID,
		ChannelName:        channelName,
		ChannelDescription: channelDescription,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		InvitedAt:          time.Now(),
		UpdatedBy:          session.UserID,
		Aliases: []collection.Alias{
			{
				AliasID:   channelID + myname,
				ChannelID:   channelID,
				UserID:   session.UserID,
				AliasName:    myname,
				AliasImg: aliasImg,
				AccessRight:     "admin",
			},
		},
		Groups:              []collection.Group{},
		InvitationCode:      common.StringRand(6),
		InvitationGuestCode: common.StringRand(8),
	}

	// 🧩 Channel登録
	collChannel := common.DB.ChannelDB.Collection("channel")
	_, err = collChannel.InsertOne(context.TODO(), newChannel)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" channel InsertOne", http.StatusOK)
		return
	}

	// 🧩 User更新
	newAliasChannel := collection.ChannelAlias{
		ChannelID: channelID,
		Alias:     myname,
	}
	userUpdate := bson.M{
		"$set": bson.M{
			"channelAliases": append(user.ChannelAliases, newAliasChannel),
			"updatedAt":      time.Now(),
		},
	}
	_, err = collUser.UpdateOne(context.TODO(), filterUser, userUpdate)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" UpdateOne user", http.StatusOK)
		return
	}

	// 🧩 Session更新
	for _, d := range mySessions {
		update := bson.M{
			"$set": bson.M{
				"channelAliases": append(d.ChannelAliases, newAliasChannel),
				"updatedAt":      time.Now(),
			},
		}
		_, err = collSession.UpdateOne(context.TODO(), bson.M{"_id": d.SessionID}, update)
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+" UpdateOne session", http.StatusOK)
			return
		}
	}

	// === ③ DB更新完了後 → Push通知・ChunkPush ===

	// 📦 ChunkPush
	contents := []string{channelName, channelDescription}
	var arr []interface{}
	arr = append(arr, "channelEdit")
	arr = append(arr, channelID)
	arr = append(arr, myname)
	arr = append(arr, contents)
	common.ChunkPush(mySessions, arr)

	// 📬 SendWebPushNotification
	for _, d := range mySessions {
		pushID := common.StringRand(1)
		pushArr := []interface{}{
			pushID,
			"alias",
			channelID,
			myname,
			[]string{session.UserID, myname, "", "admin"},
			aliasImg,
		}

		resp, err := common.SendWebPushNotification(pushArr, pushID, d)
		if err != nil {
			log.Println("SendWebPushNotification error:", err)
			continue
		}
		if resp != nil {
			defer resp.Body.Close()
		}
	}

	// === レスポンス ===
	responseData := struct {
		ChannelID    string   `json:"channelID"`
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		ChannelID:    channelID,
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
