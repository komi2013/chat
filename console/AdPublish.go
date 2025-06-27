package console

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
	"chat/common"
)

func AdPublish() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	if err != nil {
		log.Fatalf("MongoDB接続エラー: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(common.MongoDb1)

	now := time.Now()

	var targetHour int
	if now.Minute() >= 30 {
		targetHour = (now.Hour() + 11) % 24
	} else {
		targetHour = (now.Hour() + 10) % 24
	}

	weekDay := int(now.Weekday()) + 1
	if weekDay > 7 {
		weekDay = 1
	}

	adStartKey := weekDay*100 + targetHour
	log.Printf("adStartKey: %v", adStartKey)
	filter := bson.M{
		"ad_start": adStartKey,
	}

	adColl := db.Collection("ad")
	cursor, err := adColl.Find(ctx, filter)
	if err != nil {
		log.Fatalf("Findエラー: %v", err)
	}

	var ads []collection.AdStruct
	if err := cursor.All(ctx, &ads); err != nil {
		log.Fatalf("Decodeエラー: %v", err)
	}

	userColl := db.Collection("user")
	sessionColl := db.Collection("session")

	for _, ad := range ads {
		delta := float64(ad.Distance) / 100.0

		minLat := ad.Latitude - delta
		maxLat := ad.Latitude + delta
		minLng := ad.Longitude - delta
		maxLng := ad.Longitude + delta

		fmt.Printf("Ad [%s] 範囲: Lat %.2f~%.2f, Lng %.2f~%.2f\n", ad.UserID, minLat, maxLat, minLng, maxLng)

		twoWeeksAgo := time.Now().AddDate(0, 0, -14)

		userFilter := bson.M{
			"latitude":  bson.M{"$gte": minLat, "$lte": maxLat},
			"longitude": bson.M{"$gte": minLng, "$lte": maxLng},
			"signed_at": bson.M{"$gte": twoWeeksAgo},
		}

		var users []collection.UserStruct
		cursor, err := userColl.Find(ctx, userFilter)
		if err != nil {
			log.Printf("ユーザー検索エラー: %v", err)
			continue
		}
		if err := cursor.All(ctx, &users); err != nil {
			log.Printf("ユーザーデコードエラー: %v", err)
			continue
		}

		var userIDs []string
		for _, u := range users {
			fmt.Printf("  - UserID: %s (Lat: %.2f, Lng: %.2f)\n", u.UserID, u.Latitude, u.Longitude)
			userIDs = append(userIDs, u.UserID)
		}

		if len(userIDs) == 0 {
			fmt.Println("  → 該当ユーザーなし → セッション検索スキップ")
			continue
		}

		// Session検索
		sessionFilter := bson.M{
			"user_id": bson.M{"$in": userIDs},
		}

		var sessions []collection.SessionStruct
		cursor, err = sessionColl.Find(ctx, sessionFilter)
		if err != nil {
			log.Printf("セッション検索エラー: %v", err)
			continue
		}
		if err := cursor.All(ctx, &sessions); err != nil {
			log.Printf("セッションデコードエラー: %v", err)
			continue
		}

		fmt.Printf("  → 該当セッション: %d 件\n", len(sessions))
		for _, s := range sessions {
			fmt.Printf("     - SessionID: %s | UserID: %s | IsMobile: %v | CreatedAt: %v\n",
				s.SessionID, s.UserID, s.IsMobile, s.CreatedAt)
		}

	  var arr []interface{}
	  arr = append(arr, "advertisement")
	  arr = append(arr, ad.AdStart)
	  arr = append(arr, ad.AdEnd)
	  arr = append(arr, ad.AdLink)
	  arr = append(arr, ad.PathSquare)
	  // arr = append(arr, ad.PathBanner)
	  // arr = append(arr, ad.AdText)
		common.ChunkPush(sessions, db, arr)

		filter := bson.M{
			"user_id": ad.UserID, // ← AdStruct に `ID primitive.ObjectID` が必要
		}
		update := bson.M{
			"$set": bson.M{
				"active_flag": false,
				"updated_at":  time.Now(),
			},
		}

		res, err := adColl.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("広告 [%s] の非アクティブ化失敗: %v", ad.UserID, err)
			continue
		}

		fmt.Printf("広告 [%s] を非アクティブ化（matched: %d, modified: %d）\n", ad.UserID, res.MatchedCount, res.ModifiedCount)
	}
}
