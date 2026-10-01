package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/common"
	"chat/collection"
)

// PushSubscribeMobile handles FCM token registration for mobile apps
func PushSubscribeMobile(w http.ResponseWriter, r *http.Request) {
	// FCM token (Android) via the merged pushToken field.
	// NOTE: form-encoded only. Do NOT json.Decode(r.Body) here: doing so
	// consumes the body before SessionCheckTake runs, and the iOS JSON client
	// has no cookie/session yet, so reading the body would also bypass the
	// CSRF check in the future.
	pushToken := r.FormValue("pushToken")
	if pushToken == "" {
		http.Error(w, "push token is required", http.StatusBadRequest)
		return
	}

	deviceType := 2
	if deviceTypeValue := r.FormValue("deviceType"); deviceTypeValue != "" {
		if parsed, err := strconv.Atoi(deviceTypeValue); err == nil {
			switch parsed {
			case 1, 2, 3:
				deviceType = parsed
			default:
				log.Printf("PushSubscribeMobile unknown deviceType %q, defaulting to FCM", deviceTypeValue)
			}
		} else {
			log.Printf("PushSubscribeMobile invalid deviceType %q: %v, defaulting to FCM", deviceTypeValue, err)
		}
	}

	// Validate session
	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	// Get nickname if not set
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var nickname collection.NicknameStruct
	if session.Nickname == "" {
		err = common.DB.NicknameDB.Collection("nickname").
			FindOne(ctx, bson.M{"userID": session.UserID}).Decode(&nickname)
		if err != nil {
			log.Printf("Nickname not found for userID=%s: %v", session.UserID, err)
		}
		session.Nickname = nickname.Nickname
		session.NickImg = nickname.NickImg
	}

	// Update session with the merged push registration.
	coll := common.DB.SessionDB.Collection("session")
	filter := bson.D{{"_id", session.SessionID}}
	update := bson.D{
		{"$set", bson.D{
			{"pushToken", pushToken},
			{"deviceType", deviceType},
			{"nickname", session.Nickname},
			{"nickImg", session.NickImg},
			{"updatedAt", time.Now()}}},
	}
	opts := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts)

	if err != nil {
		log.Printf("Failed to update session with push token: %v", err)
		http.Error(w, "Failed to register push token", http.StatusInternalServerError)
		return
	}

	// Push ターゲットの重複排除。同一ユーザーの他セッションが同じ pushToken を
	// 保持している場合、同一端末へ同じ通知が複数回届くため（1セッション=1通知の
	// ループがそのままでは1端末に6通届く）、他セッションからは pushToken を外す。
	// SessionID(_id) 自体はサーバー生成のままで、認証情報の性質は変えない。
	if result, err := coll.UpdateMany(context.TODO(),
		bson.M{
			"userID":    session.UserID,
			"_id":       bson.M{"$ne": session.SessionID},
			"pushToken": pushToken,
		},
		bson.M{"$unset": bson.M{"pushToken": "", "deviceType": ""}},
	); err != nil {
		log.Printf("Failed to clear duplicate push tokens of user %s: %v", session.UserID, err)
	} else if result.ModifiedCount > 0 {
		log.Printf("Cleared duplicate push token from %d session(s) of user %s", result.ModifiedCount, session.UserID)
	}

	// Update session object
	session.PushToken = pushToken
	session.DeviceType = deviceType

	// Send confirmation push notification
	var arr []interface{}
	arr = append(arr, "pushCheck")
	arr = append(arr, "FCM token registration completed")
	var sessions []collection.SessionStruct
	sessions = append(sessions, session)
	common.ChunkPush(sessions, arr)

	// ======== 所属チャネルを応答に含める ========
	// Web 版の PushSubscribe（controller/PushSubscribe.go）は応答に channels を
	// 含めており、vue はそれを受け取って IndexedDB を復元する。
	// モバイル版は channels を返しておらず、アプリはローカル SQLite を唯一の
	// データ源にしていたため、再インストール（= SQLite 消去）するとドロワーの
	// チャネル一覧が戻らない状態になっていた。モバイルでも同じ配列を返す。
	channelIDSet := make(map[string]struct{})
	var channelIDs []string
	for _, ca := range session.ChannelAliases {
		if _, ok := channelIDSet[ca.ChannelID]; !ok {
			channelIDSet[ca.ChannelID] = struct{}{}
			channelIDs = append(channelIDs, ca.ChannelID)
		}
	}

	safeChannels := make([]collection.ChannelStruct, 0, len(channelIDs))
	if len(channelIDs) > 0 {
		collChannel := common.DB.ChannelDB.Collection("channel")
		cursor, err := collChannel.Find(ctx, bson.M{"_id": bson.M{"$in": channelIDs}})
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
			return
		}
		var channels []collection.ChannelStruct
		if err := cursor.All(ctx, &channels); err != nil {
			common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
			return
		}

		channelAliasMap := make(map[string]collection.ChannelAlias, len(session.ChannelAliases))
		for _, ca := range session.ChannelAliases {
			channelAliasMap[ca.ChannelID] = ca
		}
		for _, ch := range channels {
			safeChannels = append(safeChannels, sanitizeMyChannel(ch, channelAliasMap))
		}
	}

	// Return success response
	responseData := struct {
		Csrf         string                     `json:"csrf"`
		PushContents []string                   `json:"pushContents"`
		Channels     []collection.ChannelStruct `json:"channels"`
		Success      bool                       `json:"success"`
		Message      string                     `json:"message"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Channels:     safeChannels,
		Success:      true,
		Message:      "FCM token registered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

// PushUnsubscribeMobile handles FCM token removal for mobile apps
func PushUnsubscribeMobile(w http.ResponseWriter, r *http.Request) {
	// Validate session
	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	// Remove the registered push destination from the session.
	coll := common.DB.SessionDB.Collection("session")
	filter := bson.D{{"_id", session.SessionID}}
	update := bson.D{{"$unset", bson.D{
		{"pushToken", ""},
		{"deviceType", ""}}}}
	opts := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts)

	if err != nil {
		log.Printf("Failed to remove FCM token: %v", err)
		http.Error(w, "Failed to unregister FCM token", http.StatusInternalServerError)
		return
	}

	// Return success response
	responseData := struct {
		Csrf    string `json:"csrf"`
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Csrf:    session.Csrf,
		Success: true,
		Message: "FCM token unregistered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
