package console

import (
	"context"
	// "fmt"
	// "log"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"chat/collection"
	"chat/common"
)

func AdPublish() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adPublishLog := common.NewDailyLogger("ad_publish_")
	adPublishErrLog := common.NewDailyLogger("ad_publish_err_")

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	nowJST := time.Now().In(jst)

	adPublishLog.Printf("AdPublish start at: %v", nowJST)

	adColl := common.DB.AdDB.Collection("ad")

	// ✅ Public取得条件と完全一致させる
	filter := bson.M{
		"distance": -1,
		"paidAt":   bson.M{"$ne": time.Time{}},
	}

	cursor, err := adColl.Find(ctx, filter)
	if err != nil {
		adPublishErrLog.Printf("Find error: %v", err)
		return
	}

	var ads []collection.AdStruct
	if err := cursor.All(ctx, &ads); err != nil {
		adPublishErrLog.Printf("Decode error: %v", err)
		return
	}

	userColl := common.DB.UserDB.Collection("user")
	sessionColl := common.DB.SessionDB.Collection("session")

	for _, ad := range ads {

		adStartJST := ad.AdStart.In(jst)
		adEndJST := ad.AdEnd.In(jst)

		// 期間外はスキップ（削除・更新しない）
		if nowJST.Before(adStartJST) || nowJST.After(adEndJST.Add(1*time.Second)) {
			continue
		}

		paidAtJST := ad.PaidAt.In(jst)
		if paidAtJST.After(adEndJST.Add(1 * time.Second)) {
			continue
		}

		// ---- ユーザー検索（位置条件なし広告なので全体対象）----
		twoWeeksAgo := nowJST.AddDate(0, 0, -14)

		userFilter := bson.M{
			"signedAt": bson.M{"$gte": twoWeeksAgo},
		}

		var users []collection.UserStruct
		cursor, err := userColl.Find(ctx, userFilter)
		if err != nil {
			adPublishErrLog.Printf("User find error: %v", err)
			continue
		}
		if err := cursor.All(ctx, &users); err != nil {
			adPublishErrLog.Printf("User decode error: %v", err)
			continue
		}

		if len(users) == 0 {
			continue
		}

		userIDs := make([]string, 0, len(users))
		for _, u := range users {
			userIDs = append(userIDs, u.UserID)
		}

		// ---- セッション検索 ----
		sessionFilter := bson.M{
			"userID": bson.M{"$in": userIDs},
		}

		var sessions []collection.SessionStruct
		cursor, err = sessionColl.Find(ctx, sessionFilter)
		if err != nil {
			adPublishErrLog.Printf("Session find error: %v", err)
			continue
		}
		if err := cursor.All(ctx, &sessions); err != nil {
			adPublishErrLog.Printf("Session decode error: %v", err)
			continue
		}

		if len(sessions) == 0 {
			continue
		}

		// ---- Push ----
		arr := []interface{}{
			"advertisement",
			ad.AdStart,     // time.Time
			ad.AdEnd,       // time.Time
			ad.AdLink,
			ad.PathSquare,
 		}
		common.ChunkPush(sessions, arr)

		adPublishLog.Printf(
			"Ad pushed: adID=%s users=%d",
			ad.AdID, len(sessions),
		)
	}
}
