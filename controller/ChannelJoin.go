package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"chat/collection"
	"chat/common"
)

func ChannelJoin(w http.ResponseWriter, r *http.Request) {
	channelID := r.FormValue("channelID")
	myname := r.FormValue("myname")
	myimg := r.FormValue("myimg")
	code := r.FormValue("code")

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";session check", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ========== FIND PHASE ==========
	collInvitation := common.DB.ChannelDB.Collection("channel")
	var invitation collection.ChannelStruct
	invitationFilter := bson.M{"_id": channelID}
	if err := collInvitation.FindOne(ctx, invitationFilter).Decode(&invitation); err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	if (invitation.InvitationCode != code && invitation.InvitationGuestCode != code) ||
	    time.Since(invitation.InvitedAt) > 7*24*time.Hour {
		common.WriteResponseWithSession(w, session, "code is wrong or invitation expired", http.StatusOK)
		return
	}


	// 名前重複チェック
	for _, alias := range invitation.Aliases {
		if alias.AliasName == myname && alias.AccessRight != "inquirer" {
			common.WriteResponseWithSession(w, session, "すでに同じ名前が存在しています", http.StatusOK)
			return
		}
	}

	// "inquirer" 以外のUserID収集
	var userIDs []string
	for _, alias := range invitation.Aliases {
		if alias.AccessRight != "inquirer" {
			userIDs = append(userIDs, alias.UserID)
		}
	}

	// 権限設定
	accessRight := ""
	if invitation.InvitationGuestCode == code {
		accessRight = "guest"
	}

	// セッション一括取得
	collSession := common.DB.SessionDB.Collection("session")
	allUserIDs := append(userIDs, session.UserID)
	cursor, err := collSession.Find(ctx, bson.M{"userID": bson.M{"$in": allUserIDs}})
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+": Session find error", http.StatusOK)
		return
	}

	var allSessions []collection.SessionStruct
	if err := cursor.All(ctx, &allSessions); err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+": Session decode error", http.StatusOK)
		return
	}

	var (
		mySessions    []collection.SessionStruct
		otherSessions []collection.SessionStruct
	)
	for _, s := range allSessions {
		if s.UserID == session.UserID {
			mySessions = append(mySessions, s)
		} else {
			otherSessions = append(otherSessions, s)
		}
	}

	// ========== UPDATE PHASE ==========
	aliasImg, err := common.ImgSave(myimg, session.UserID, myname, channelID, 3, 1)
	if err != nil {
		log.Printf("ImgSave: %v; Req:", err, r.URL.Path, r.Form)
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	newAlias := collection.Alias{
		AliasID:     channelID + myname,
		ChannelID:   channelID,
		AliasName:   myname,
		AliasImg:    aliasImg,
		UserID:      session.UserID,
		AccessRight: accessRight,
	}
	newAliasChannel := collection.ChannelAlias{
		ChannelID: channelID,
		Alias:     myname,
	}

	// --- userコレクション更新 ---
	collUser := common.DB.UserDB.Collection("user")
	filterUser := bson.M{"_id": session.UserID}
	userUpdate := bson.M{
		"$addToSet": bson.M{"channelAliases": newAliasChannel},
		"$set":      bson.M{"updatedAt": time.Now()},
	}
	if _, err := collUser.UpdateOne(ctx, filterUser, userUpdate); err != nil {
		log.Printf("user UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	// --- 自分のセッション更新 ---
	for _, d := range mySessions {
		d.UpdatedAt = time.Now()
		d.ChannelAliases = append(d.ChannelAliases, newAliasChannel)

		filter := bson.M{"_id": d.SessionID}
		update := bson.M{"$set": d}
		if _, err := collSession.UpdateOne(ctx, filter, update); err != nil {
			common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
			return
		}
	}

	// --- チャンネルにAlias追加 ---
	pushUpd := bson.M{
		"$push": bson.M{
			"aliases":    newAlias,
		},
	}
	if _, err := collInvitation.UpdateOne(ctx, invitationFilter, pushUpd); err != nil {
		log.Printf("collInvitation.UpdateOne pushUpd: %v; Req:", err, r.URL.Path, r.Form)
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	// ========== PUSH PHASE ==========
	// 自分へチャンネル情報通知
	contents := []string{invitation.ChannelName, invitation.ChannelDescription}
	common.ChunkPush(mySessions, []interface{}{"channelEdit", channelID, myname, contents})

	// エイリアス一覧送信
	invitation.Aliases = append(invitation.Aliases, newAlias)
	for _, d := range invitation.Aliases {
		aliasData := []string{d.UserID, d.AliasName, d.AliasBio, d.AccessRight}
		common.ChunkPush(mySessions, []interface{}{"alias", channelID, myname, aliasData, d.AliasImg})
	}

	for _, d := range invitation.Groups {
		gData := []interface{}{d.GroupName, d.AliasNames}
		common.ChunkPush(mySessions, []interface{}{"group", channelID, myname, gData, d.GroupImg})
	}

	// 他ユーザーへ通知
	contents = []string{session.UserID, myname, "", accessRight}
	common.ChunkPush(otherSessions, []interface{}{"alias", channelID, myname, contents, aliasImg})

	responseData := struct {
		Csrf           string   `json:"csrf"`
		PushContents   []string `json:"pushContents"`
	}{
		Csrf:           session.Csrf,
		PushContents:   session.PushContents,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
