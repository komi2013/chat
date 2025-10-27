package controller

import (
	"context"
	"encoding/json"
	// "log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/collection"
	"chat/common"
)

func UserEdit(w http.ResponseWriter, r *http.Request) {
	var user collection.UserStruct

	latStr := r.FormValue("latitude")
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "緯度 (latitude) が不正です", http.StatusOK)
		return
	}
	lngStr := r.FormValue("longitude")
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "経度 (longitude) が不正です", http.StatusOK)
		return
	}

	user.Latitude = lat
	user.Longitude = lng
	user.Mail = r.FormValue("mail")
	user.Telephone = r.FormValue("telephone")

	var nickname collection.NicknameStruct
	nickname.Nickname = r.FormValue("nickname")
	nickname.NickBio = r.FormValue("nickBio")

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	nickname.NickImg, err = common.ImgSave(r.FormValue("nickImg"), session.UserID, nickname.Nickname, "", 3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := "ユーザー情報は更新されました "

	if nickname.Nickname != "" {
		collNickname := common.DB.NicknameDB.Collection("nickname")
		filterNickname := bson.M{"userID": session.UserID}
		cursor, err := collNickname.Find(ctx, filterNickname)
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
			return
		}

		var nicknames []collection.NicknameStruct
		if err = cursor.All(ctx, &nicknames); err != nil {
			common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
			return
		}

		// 既に同じニックネームが存在するか確認
		var existingNick *collection.NicknameStruct
		for _, n := range nicknames {
			if n.Nickname == nickname.Nickname {
				existingNick = &n
				break
			}
		}

		if existingNick != nil {
			// 更新処理（NickImg, NickBio の更新）
			update := bson.M{
				"$set": bson.M{
					"nickImg":   nickname.NickImg,
					"nickBio":   nickname.NickBio,
					"updatedAt": time.Now(),
				},
			}
			filter := bson.M{"_id": existingNick.Nickname}
			_, err := collNickname.UpdateOne(ctx, filter, update)
			if err != nil {
				common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
				return
			}
			message += "ニックネーム情報が更新されました"
		} else {
			// 3件制限チェック
			if len(nicknames) >= 3 {
				common.WriteResponseWithSession(w, session, "ニックネーム情報は3件以上は登録できません", http.StatusOK)
				return
			} else {
				nickname.UserID = session.UserID
				nickname.UpdatedAt = time.Now()
				_, err := collNickname.InsertOne(ctx, nickname)
				if err != nil {
					common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
					return
				}
				message += "ニックネーム情報が追加されました"
			}
		}
	}

	// SessionDB 更新
	update := bson.M{
		"$set": bson.M{
			"mail":      user.Mail,
			"telephone": user.Telephone,
			"nickname": nickname.Nickname,
			"nickImg": nickname.NickImg,
		},
	}
	filterUser := bson.M{"userID": session.UserID}
	opts := options.Update().SetUpsert(false)
	collSessions := common.DB.SessionDB.Collection("session")
	_, err = collSessions.UpdateMany(ctx, filterUser, update, opts)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	// UserDB 更新
	update = bson.M{"$set": user}
	filterUser = bson.M{"_id": session.UserID}
	opts = options.Update().SetUpsert(true)
	collUser := common.DB.UserDB.Collection("user")
	_, err = collUser.UpdateOne(ctx, filterUser, update, opts)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}

	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
		Message      string   `json:"message"`
		Nickname     string   `json:"nickname"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Message:      message,
		Nickname:     nickname.Nickname,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
