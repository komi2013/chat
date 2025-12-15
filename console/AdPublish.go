package console

import (
	"context"
	"fmt"
	"log"
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

	now := time.Now()

	adPublishLog.Printf("AdPublish start at: %v", now)

	adColl := common.DB.AdDB.Collection("ad")

	// ✅ Simple & correct filter using time.Time
	filter := bson.M{
		"adStart":    bson.M{"$lte": now},
		"adEnd":      bson.M{"$gt": now},
		"activeFlag": true,
	}

	cursor, err := adColl.Find(ctx, filter)
	if err != nil {
		adPublishErrLog.Printf("Find error: %v", err)
		return
	}

	var ads []collection.AdStruct
	if err := cursor.All(ctx, &ads); err != nil {
		log.Fatalf("Decode error: %v", err)
	}

	userColl := common.DB.UserDB.Collection("user")
	sessionColl := common.DB.SessionDB.Collection("session")

	for _, ad := range ads {

		// ---- 地理範囲計算 ----
		delta := float64(ad.Distance) / 100.0
		minLat := ad.Latitude - delta
		maxLat := ad.Latitude + delta
		minLng := ad.Longitude - delta
		maxLng := ad.Longitude + delta

		fmt.Printf(
			"Ad [%s] active [%v ~ %v], range Lat %.2f~%.2f Lng %.2f~%.2f\n",
			ad.UserID, ad.AdStart, ad.AdEnd,
			minLat, maxLat, minLng, maxLng,
		)

		// ---- ユーザー検索 ----
		twoWeeksAgo := now.AddDate(0, 0, -14)

		userFilter := bson.M{
			"latitude":  bson.M{"$gte": minLat, "$lte": maxLat},
			"longitude": bson.M{"$gte": minLng, "$lte": maxLng},
			"signedAt":  bson.M{"$gte": twoWeeksAgo},
		}

		var users []collection.UserStruct
		cursor, err := userColl.Find(ctx, userFilter)
		if err != nil {
			log.Printf("User find error: %v", err)
			continue
		}
		if err := cursor.All(ctx, &users); err != nil {
			log.Printf("User decode error: %v", err)
			continue
		}

		if len(users) == 0 {
			fmt.Println("  → No users found, skip")
			continue
		}

		var userIDs []string
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
			log.Printf("Session find error: %v", err)
			continue
		}
		if err := cursor.All(ctx, &sessions); err != nil {
			log.Printf("Session decode error: %v", err)
			continue
		}

		fmt.Printf("  → Target sessions: %d\n", len(sessions))

		// ---- Push ----
		arr := []interface{}{
			"advertisement",
			ad.AdStart,     // time.Time
			ad.AdEnd,       // time.Time
			ad.AdLink,
			ad.PathSquare,
		}

		common.ChunkPush(sessions, arr)

		// ---- Deactivate ad ----
		update := bson.M{
			"$set": bson.M{
				"activeFlag": false,
				"updatedAt":  now,
			},
		}

		res, err := adColl.UpdateOne(
			ctx,
			bson.M{"userID": ad.UserID},
			update,
		)

		if err != nil {
			log.Printf("Deactivate ad [%s] failed: %v", ad.UserID, err)
			continue
		}

		fmt.Printf(
			"Ad [%s] deactivated (matched: %d, modified: %d)\n",
			ad.UserID, res.MatchedCount, res.ModifiedCount,
		)
	}
}
