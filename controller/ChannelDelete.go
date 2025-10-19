package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  // "log"
  "net/http"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func ChannelDelete(w http.ResponseWriter, r *http.Request) {
	channelID := r.FormValue("channelID")
	updatedBy := r.FormValue("updatedBy")

	// === セッションチェック ===
	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Session Error", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// === チャンネル取得 ===
	collChannel := common.DB.ChannelDB.Collection("channel")
	var channelData collection.ChannelStruct
	if err := collChannel.FindOne(ctx, bson.M{"_id": channelID}).Decode(&channelData); err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+":Channel not found", http.StatusOK)
		return
	}

	// === updatedBy の権限確認 ===
	admin := false
	for _, alias := range channelData.Aliases {
		if alias.AliasName == updatedBy && alias.AccessRight == "admin" {
			admin = true
			break
		}
	}
	if !admin {
		common.WriteResponseWithSession(w, session, "No admin access right", http.StatusOK)
		return
	}

	// === すべての userID 抽出 ===
	userIDs := make([]string, 0)
	unique := make(map[string]struct{})
	for _, alias := range channelData.Aliases {
		if _, ok := unique[alias.UserID]; !ok {
			unique[alias.UserID] = struct{}{}
			userIDs = append(userIDs, alias.UserID)
		}
	}

	// === セッション取得 ===
	collSession := common.DB.SessionDB.Collection("session")
	cursor, err := collSession.Find(ctx, bson.M{"userID": bson.M{"$in": userIDs}})
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+":Session find error", http.StatusOK)
	}
	var sessions []collection.SessionStruct
	if err := cursor.All(ctx, &sessions); err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+":Session decode error", http.StatusOK)
	}

	for _, s := range sessions {
	    // === ChannelAliases から削除対象以外を残す ===
	    filteredAliases := make([]collection.ChannelAlias, 0)
	    for _, alias := range s.ChannelAliases {
	        if alias.ChannelID != channelID {
	            filteredAliases = append(filteredAliases, alias)
	        }
	    }

	    // === DB 更新 ===
	    _, err := collSession.UpdateOne(
	        ctx,
	        bson.M{"_id": s.SessionID},
	        bson.M{
	            "$set": bson.M{
	                "channelAliases": filteredAliases,
	                "updatedAt":      time.Now(),
	            },
	        },
	    )
	    if err != nil {
	        common.WriteResponseWithSession(w, session, err.Error()+":Session update error", http.StatusOK)
	        return
	    }
	}

	// === userコレクションから対象ユーザー取得 ===
	collUser := common.DB.UserDB.Collection("user")

	cursor, err = collUser.Find(ctx, bson.M{"_id": bson.M{"$in": userIDs}})
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+":User find error", http.StatusOK)
		return
	}

	var users []collection.UserStruct
	if err := cursor.All(ctx, &users); err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+":User decode error", http.StatusOK)
		return
	}

	// === 各ユーザーのChannelAliasesを更新（ループ版） ===
	for _, u := range users {
		// 削除対象のchannelID以外を残す
		filteredAliases := make([]collection.ChannelAlias, 0)
		for _, alias := range u.ChannelAliases {
			if alias.ChannelID != channelID {
				filteredAliases = append(filteredAliases, alias)
			}
		}

		// === UpdateOne実行 ===
		update := bson.M{
			"$set": bson.M{
				"channelAliases": filteredAliases,
				"updatedAt":      time.Now(),
			},
		}

		_, err := collUser.UpdateOne(
			ctx,
			bson.M{"_id": u.UserID},
			update,
		)
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+":User update error", http.StatusOK)
			return
		}
	}

	// === チャンネルをDBから削除 ===
	_, err = collChannel.DeleteOne(ctx, bson.M{"_id": channelID})
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+":Failed to delete channel", http.StatusOK)
		return
	}

	// === チャンネル削除通知 push ===
	channelPushArray := []interface{}{
		"channelEdit", // pushTitle
		channelID,       // channelID
		updatedBy,       // updatedBy
		"delete",             // contents
		// nil,             // imgPath or files
	}
	common.ChunkPush(sessions, channelPushArray)

	// === レスポンス ===
	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
