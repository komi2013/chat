package controller

import (
  "context"
  // "fmt"
  "log"
  // "encoding/json"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"

)

func SignInTmp(w http.ResponseWriter, r *http.Request) {

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  userID := r.FormValue("userID")
  collUser := db1.Collection("user")
  filterUser := bson.M{"_id": userID}
  var user collection.UserStruct
  err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
  if err != nil {
    log.Printf("user FindOne: %v; Req:", err, r.URL.Path, r.Form)
  }

	coll := db1.Collection("session")
	cursor, err := coll.Find(context.TODO(), bson.D{{"userID", userID}})
	if err != nil {
		log.Printf("Find error: %v", err)
		return
	}
	var sessions []collection.SessionStruct
	if err = cursor.All(context.TODO(), &sessions); err != nil {
		log.Printf("Cursor decode error: %v", err)
		return
	}

	isMobile := common.IsMobile(r.UserAgent())
	var matchedSession *collection.SessionStruct
	for _, s := range sessions {
		if s.IsMobile == isMobile {
			matchedSession = &s
			break
		}
	}

	now := time.Now()
	var sessionID string
	if matchedSession != nil {
		// 既存のセッションを更新
		sessionID = matchedSession.SessionID
		filter := bson.D{
			{"_id", sessionID},
		}
		update := bson.D{{"$set", bson.D{
			{"userID", userID},
			{"channelAliases", user.ChannelAliases},
			{"updatedAt", now},
		}}}

		_, err = coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			log.Printf("Update error: %v", err, r.URL.Path, r.Form)
		}
	} else {
		// 新規挿入（Insert）：ここでのみ SessionID を生成
		sessionID = common.StringRand(16)
		session := collection.SessionStruct{
			SessionID:      sessionID,
			UserID:         userID,
			ChannelAliases: user.ChannelAliases,
			CreatedAt:      now,
			UpdatedAt:      now,
			IsMobile:       isMobile,
			PushContents:   []string{},
		}
		_, err = coll.InsertOne(context.TODO(), session)
		if err != nil {
			log.Printf("Insert error: %v", err, r.URL.Path, r.Form)
		}
	}

  cookie := &http.Cookie{
    Name:     "ss",
    Value:    sessionID,
    MaxAge:   1728000,
    Secure:   true,
    HttpOnly: true,
    Path:     "/",
  }
  http.SetCookie(w, cookie)

	coll = db1.Collection("user")
	userFilter := bson.D{{"_id", userID}}
	update := bson.D{{"$set", bson.D{
		{"signedAt", time.Now()}}}}
	opts := options.Update().SetUpsert(true)
	_, err = coll.UpdateOne(context.TODO(), userFilter, update, opts)
  // http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
}
