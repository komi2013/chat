package controller

import (
	"context"
	"encoding/json"
	// "log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/common"
	"chat/collection"
)

func ChannelEdit(w http.ResponseWriter, r *http.Request) {
	var (
		pushNames     []string
		contents      interface{}
		deleteAliases []string
	)

	if err := json.Unmarshal([]byte(r.FormValue("pushNames")), &pushNames); r.FormValue("pushNames") != "" && err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "pushNames JSON Unmarshal Error", http.StatusOK)
		return
	}
	if err := json.Unmarshal([]byte(r.FormValue("contents")), &contents); r.FormValue("contents") != "" && err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "contents JSON Unmarshal Error", http.StatusOK)
		return
	}
	if err := json.Unmarshal([]byte(r.FormValue("deleteAliases")), &deleteAliases); r.FormValue("deleteAliases") != "" && err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "deleteAliases JSON Unmarshal Error", http.StatusOK)
		return
	}

	var admin collection.Alias
	if err := json.Unmarshal([]byte(r.FormValue("admin")), &admin); r.FormValue("admin") != "" && err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "admin JSON Unmarshal Error", http.StatusOK)
		return
	}

	var groups []collection.Group
	if err := json.Unmarshal([]byte(r.FormValue("groups")), &groups); r.FormValue("groups") != "" && err != nil {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "groups JSON Unmarshal Error", http.StatusOK)
    return
	}


	channelID := r.FormValue("channelID")
	channelName := r.FormValue("channelName")
	channelDescription := r.FormValue("channelDescription")
	updatedBy := r.FormValue("updatedBy")
	guest := r.FormValue("guest") != ""
	csrf := r.FormValue("csrf")
	generateInvitation := r.FormValue("generateInvitation") != ""

	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+"session check some error", http.StatusOK)
		return
	}

	trueAccess := false
	for _, d := range session.ChannelAliases {
		if d.Alias == updatedBy && d.ChannelID == channelID && !d.GuestFlag {
			trueAccess = true
			break
		}
	}
	if !trueAccess {
		common.WriteResponseWithSession(w, session, "ChannelAliases !trueAccess", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := common.DB.ChannelDB.Collection("channel")
	filter := bson.M{"_id": channelID}
	var channel collection.ChannelStruct
	err = coll.FindOne(ctx, filter).Decode(&channel)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+"Channel not found", http.StatusOK)
		return
	}

	responseCode := channel.InvitationCode
	if guest {
		responseCode = channel.InvitationGuestCode
	}

	// ======== Step 6: PushNamesからPush対象ユーザーを抽出 ========
	uniqueIDs := make(map[string]struct{})
	var pushUserIDs []string
	pushNameSet := make(map[string]struct{}, len(pushNames))
	for _, name := range pushNames {
		pushNameSet[name] = struct{}{}
	}
	for _, alias := range channel.Aliases {
		if _, ok := pushNameSet[alias.AliasName]; ok {
			if _, exists := uniqueIDs[alias.UserID]; !exists {
				uniqueIDs[alias.UserID] = struct{}{}
				pushUserIDs = append(pushUserIDs, alias.UserID)
			}
		}
	}

	// ======== ★ 変更点: collSession.Find() はここで1回のみ実行 ========
	var allSessions []collection.SessionStruct
	if len(pushUserIDs) > 0 {
		collSession := common.DB.SessionDB.Collection("session")
		cursor, err := collSession.Find(ctx, bson.M{"userID": bson.M{"$in": pushUserIDs}})
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+"Push session fetch error", http.StatusOK)
			return
		}
		if err := cursor.All(ctx, &allSessions); err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+"Session decode error", http.StatusOK)
			return
		}
	}

	aliasUpdateAdmins := make([]collection.Alias, len(channel.Aliases))
	updateFields := bson.M{}

	if channelName != "" {
		updateFields["channelName"] = channelName
	}
	if channelDescription != "" {
		updateFields["channelDescription"] = channelDescription
	}
	if generateInvitation {
		if guest {
			invitationGuestCode := common.StringRand(20)
			updateFields["invitationGuestCode"] = invitationGuestCode
			responseCode = invitationGuestCode
		} else {
			invitationCode := common.StringRand(16)
			updateFields["invitationCode"] = invitationCode
			responseCode = invitationCode
		}
		updateFields["invitedAt"] = time.Now()
	}
	if len(groups) > 0 {
    updateFields["groups"] = groups
	}

	// ======== Step 7: deleteAliases処理（別UserIDリストを使う） ========
	if len(deleteAliases) > 0 {
		newAliases := make([]collection.Alias, 0, len(channel.Aliases))
		deleteUserIDs := make([]string, 0)

		// 削除対象エイリアスのUserID収集と新しいAliases作成
		for _, a := range channel.Aliases {
			shouldDelete := false
			for _, del := range deleteAliases {
				if a.AliasName == del {
					shouldDelete = true
					deleteUserIDs = append(deleteUserIDs, a.UserID)
					break
				}
			}
			if !shouldDelete {
				newAliases = append(newAliases, a)
			}
		}
		updateFields["aliases"] = newAliases

		// ======== NEW: remove empty groups after alias deletion ========
		if len(channel.Groups) > 0 {
		    updatedGroups := make([]collection.Group, 0, len(channel.Groups))
		    for _, g := range channel.Groups {
		        // remove deleted aliases from each group
		        filteredMembers := []string{}
		        for _, aliasName := range g.AliasNames {
		            remove := false
		            for _, del := range deleteAliases {
		                if aliasName == del {
		                    remove = true
		                    break
		                }
		            }
		            if !remove {
		                filteredMembers = append(filteredMembers, aliasName)
		            }
		        }
		        // if group has no members left, skip (delete group)
		        if len(filteredMembers) == 0 {
		            continue
		        }
		        g.AliasNames = filteredMembers
		        updatedGroups = append(updatedGroups, g)
		    }
		    updateFields["groups"] = updatedGroups
		}

		// ======== ★ 変更点: すでに取得したallSessionsから対象ユーザーを抽出 ========
		deleteSessions := make([]collection.SessionStruct, 0)
		for _, s := range allSessions {
			for _, id := range deleteUserIDs {
				if s.UserID == id {
					deleteSessions = append(deleteSessions, s)
					break
				}
			}
		}

		// 該当セッションをDB更新
		collSession := common.DB.SessionDB.Collection("session")
		for _, s := range deleteSessions {
			filteredAliases := make([]collection.ChannelAlias, 0)
			for _, alias := range s.ChannelAliases {
				if alias.ChannelID != channelID {
					filteredAliases = append(filteredAliases, alias)
				}
			}
			_, err := collSession.UpdateOne(
				ctx,
				bson.M{"_id": s.SessionID},
				bson.M{"$set": bson.M{
					"channelAliases": filteredAliases,
					"updatedAt":      time.Now(),
				}},
			)
			if err != nil {
				common.WriteResponseWithSession(w, session, err.Error()+": Session update error", http.StatusOK)
				return
			}
		}

		// ======== ユーザー削除も同様に更新 ========
		collUser := common.DB.UserDB.Collection("user")
		cursor, err := collUser.Find(ctx, bson.M{"_id": bson.M{"$in": deleteUserIDs}})
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+":User find error", http.StatusOK)
			return
		}
		var users []collection.UserStruct
		if err := cursor.All(ctx, &users); err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+":User decode error", http.StatusOK)
			return
		}
		for _, u := range users {
			filteredAliases := make([]collection.ChannelAlias, 0)
			for _, alias := range u.ChannelAliases {
				if alias.ChannelID != channelID {
					filteredAliases = append(filteredAliases, alias)
				}
			}
			update := bson.M{
				"$set": bson.M{
					"channelAliases": filteredAliases,
					"updatedAt":      time.Now(),
				},
			}
			_, err := collUser.UpdateOne(ctx, bson.M{"_id": u.UserID}, update)
			if err != nil {
				common.WriteResponseWithSession(w, session, err.Error()+":User update error", http.StatusOK)
				return
			}
		}
	} else if r.FormValue("admin") != "" {
		copy(aliasUpdateAdmins, channel.Aliases)
		updated := false
		for i, a := range aliasUpdateAdmins {
			if a.AliasName == admin.AliasName {
				aliasUpdateAdmins[i] = admin
				updated = true
				break
			}
		}
		if !updated {
			aliasUpdateAdmins = append(aliasUpdateAdmins, admin)
		}
		updateFields["aliases"] = aliasUpdateAdmins
	}

	update := bson.M{"$set": updateFields}
	opts := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+"Failed to update channel", http.StatusOK)
		return
	}

	// ======== Push通知処理 ========
	channelArr := []interface{}{
		"channelEdit",
		channelID,
		updatedBy,
		[]string{channelName, channelDescription},
	}
	if channelName != "" {
		common.ChunkPush(allSessions, channelArr)
	}

	for _, aliasName := range deleteAliases {
		aliasData := []interface{}{"", aliasName, "", "delete"}
		aliasPushArray := []interface{}{"alias", channelID, updatedBy, aliasData, ""}
		common.ChunkPush(allSessions, aliasPushArray)

		for _, s := range allSessions {
			// log.Printf("[DEBUG] Checking session: ChannelAliasCount=%d",s.ChannelAliases)

			for _, chAlias := range s.ChannelAliases {
				// log.Printf("[DEBUG]   ChannelAlias: ChannelID=%s, Alias=%s, GuestFlag=%v",
				// 	chAlias.ChannelID, chAlias.Alias, chAlias.GuestFlag)

				// 一致チェック
				if chAlias.ChannelID == channelID && chAlias.Alias == aliasName {
					// log.Printf("[MATCH] Found matching alias! aliasName=%s, channelID=%s, updatedBy=%s, sessionUserID=%s",
					// 	aliasName, channelID, updatedBy, s.UserID)

					// push データ内容を出力
					channelPushArray := []interface{}{
						"channelEdit", // pushTitle
						channelID,     // channelID
						updatedBy,     // updatedBy
						"delete",      // contents
					}
					// log.Printf("[PUSH DATA] %+v", channelPushArray)

					// push 実行
					common.ChunkPush([]collection.SessionStruct{s}, channelPushArray)
					// log.Printf("[PUSH SENT] To session: %s (UserID=%s)", s.SessionID, s.UserID)

					break
				}
			}
		}
	}

	if r.FormValue("admin") != "" {
		aliasData := []interface{}{
			admin.UserID,
			admin.AliasName,
			admin.AliasBio,
			admin.AccessRight,
		}
		aliasPushArray := []interface{}{
			"alias",
			channelID,
			updatedBy,
			aliasData,
			admin.AliasImg,
		}
		common.ChunkPush(allSessions, aliasPushArray)
	}

	// ======== Response ========
	responseData := struct {
		InvitationCode string   `json:"invitationCode"`
		Csrf           string   `json:"csrf"`
		PushContents   []string `json:"pushContents"`
	}{
		InvitationCode: responseCode,
		Csrf:           session.Csrf,
		PushContents:   session.PushContents,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
