package common

import (
  "context"
  "errors"
  // "log"
  "net/http"
  "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
)

func SessionGet(db1 *mongo.Database, w http.ResponseWriter, r *http.Request) (collection.SessionStruct, error) {
  var session collection.SessionStruct
  cookie, err := r.Cookie("ss")
  if err != nil {
    return session, err
  }
  coll := db1.Collection("session")
  filter := bson.D{{"_id", cookie.Value}}
  opts := options.FindOne().SetProjection(bson.D{})
  err = coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  return session, err
}

func SessionCheck(db1 *mongo.Database, w http.ResponseWriter, r *http.Request, token string) (collection.SessionStruct, error) {
  var session collection.SessionStruct
  cookie, err := r.Cookie("ss")
  if err != nil {
    return session, err
  }
  coll := db1.Collection("session")
  filter := bson.D{{"_id", cookie.Value}}
  opts := options.FindOne().SetProjection(bson.D{})
  err = coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  if err != nil {
    return session, err
  }
	if session.Csrf != token {
		return session, errors.New("token error")
		// LogError("SessionCheck:", nil, session.Csrf, token)
	}

	if time.Since(session.UpdatedAt) > 20*24*time.Hour {
		return session, errors.New("session expired")
	}	else if time.Since(session.UpdatedAt) > 10*24*time.Hour {
		session, err = RegenerateSessionData(db1, session, w)
	} else {
		// UpdateSessionTimestamp(db, session)
		session, err = CheckMakeCSRFToken(db1, session, token)
	}
  // session, err = CheckMakeCSRFToken(db1, session, token)
  return session, err
}

func CheckMakeCSRFToken(db1 *mongo.Database, session collection.SessionStruct, token string) (collection.SessionStruct, error) {
	if session.Csrf != token {
		return session, errors.New("token error ")
		// LogError("CheckMakeCSRFToken:", nil, session.Csrf, token)
	}
	session, err := ReGenerateData(db1, session)
	return session, err
}

func ReGenerateData(db1 *mongo.Database, session collection.SessionStruct) (collection.SessionStruct, error) {
	coll := db1.Collection("session")
	token := StringRand(16)
	session.Csrf = token
	var returnContents []string
	if len(session.PushContents) > 100 {
		returnContents = session.PushContents[:100]
		session.PushContents = session.PushContents[100:] // 残りを保持
	} else {
		returnContents = session.PushContents
		session.PushContents = []string{} // 全部消費したら空
	}
	filter := bson.D{{"_id", session.SessionID}}
	update := bson.D{
		{"$set", bson.D{
			{"csrf", session.Csrf},
			{"updatedAt", time.Now()},
			{"pushContents", session.PushContents}, // 残りだけを保存
		}},
	}
	opts := options.Update().SetUpsert(false)
	_, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	session.PushContents = returnContents
	return session, err
}

func RegenerateSessionData(db *mongo.Database, session collection.SessionStruct, w http.ResponseWriter) (collection.SessionStruct, error) {
	newSession := session
	newSession.SessionID = StringRand(16)
	newSession.Csrf = StringRand(16)
	newSession.UpdatedAt = time.Now()
	newSession.PushContents = []string{}
	// newSession.AliasArray = [][]string{}
	// newSession.ChannelAliases = []collection.ChannelAlias{}

	cookie := &http.Cookie{
		Name:     "ss",
		Value:    newSession.SessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
	coll := db.Collection("session")

	_, err := coll.InsertOne(context.TODO(), newSession)
	if err != nil {
		return session, err
	}

	_, err = coll.DeleteOne(context.TODO(), bson.M{"_id": session.SessionID})
	if err != nil {
		return session, err
	}
	return newSession, nil
}

func ReGenerateCSRF(db1 *mongo.Database, session collection.SessionStruct) (collection.SessionStruct, error) {
	coll := db1.Collection("session")
	session.Csrf = StringRand(16)
	filter := bson.D{{"_id", session.SessionID}}
	update := bson.D{
		{"$set", bson.D{
			{"csrf", session.Csrf},
		}},
	}
	opts := options.Update().SetUpsert(false)
	_, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	return session, err
}

func FilterSessionsByChannelID(sessions []collection.SessionStruct, channelID string) []collection.SessionStruct {
	var filteredSessions []collection.SessionStruct
	for _, session := range sessions {
		var filteredAliases []collection.ChannelAlias
		for _, alias := range session.ChannelAliases {
			if alias.ChannelID == channelID {
				filteredAliases = append(filteredAliases, alias)
			}
		}
		if len(filteredAliases) > 0 {
			session.ChannelAliases = filteredAliases
			filteredSessions = append(filteredSessions, session)
		}
	}
	return filteredSessions
}

func IsMobile(userAgent string) bool {
	mobileKeywords := []string{
		"Mobile", "Android", "iPhone", "iPad", "iPod", "Windows Phone",
	}

	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}
	return false
}

func SessionCheckTake(w http.ResponseWriter, r *http.Request, token string) (collection.SessionStruct, error) {
  var session collection.SessionStruct
  cookie, err := r.Cookie("ss")
  if err != nil {
    return session, err
  }
  coll := DB.SessionDB.Collection("session")
  filter := bson.D{{"_id", cookie.Value}}
  opts := options.FindOne().SetProjection(bson.D{})
  err = coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  if err != nil {
    return session, err
  }
	// if session.Csrf != token {
	// 	return session, errors.New("token error")
	// 	// LogError("SessionCheck:", nil, session.Csrf, token)
	// }

	if time.Since(session.UpdatedAt) > 20*24*time.Hour {
		return session, errors.New("session expired")
	}	else if time.Since(session.UpdatedAt) > 10*24*time.Hour {
		session, err = SessionRegenerate(session, w)
	} else {
		// UpdateSessionTimestamp(db, session)
		// session, err = CSRFcheckMake(session, token)
	}
  // session, err = CheckMakeCSRFToken(db1, session, token)
  return session, err
}

func CSRFcheckMake(session collection.SessionStruct, token string) (collection.SessionStruct, error) {
	if session.Csrf != token {
		return session, errors.New("token error")
	}
	session, err := PushReGenerate(session)
	return session, err
}

func PushReGenerate(session collection.SessionStruct) (collection.SessionStruct, error) {
	coll := DB.SessionDB.Collection("session")
	token := StringRand(16)
	session.Csrf = token

	var returnContents []string
	if len(session.PushContents) > 100 {
		returnContents = session.PushContents[:100]
		session.PushContents = session.PushContents[100:]
	} else {
		returnContents = session.PushContents
		session.PushContents = []string{}
	}

	filter := bson.D{{"_id", session.SessionID}}
	update := bson.D{
		{"$set", bson.D{
			{"csrf", session.Csrf},
			{"updatedAt", time.Now()},
			{"pushContents", session.PushContents},
		}},
	}
	opts := options.Update().SetUpsert(false)
	_, err := coll.UpdateOne(context.TODO(), filter, update, opts)

	session.PushContents = returnContents
	return session, err
}

func SessionRegenerate(session collection.SessionStruct, w http.ResponseWriter) (collection.SessionStruct, error) {
	newSession := session
	newSession.SessionID = StringRand(16)
	newSession.Csrf = StringRand(16)
	newSession.UpdatedAt = time.Now()
	newSession.PushContents = []string{}

	cookie := &http.Cookie{
		Name:     "ss",
		Value:    newSession.SessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)

	coll := DB.SessionDB.Collection("session")

	_, err := coll.InsertOne(context.TODO(), newSession)
	if err != nil {
		return session, err
	}

	_, err = coll.DeleteOne(context.TODO(), bson.M{"_id": session.SessionID})
	if err != nil {
		return session, err
	}
	return newSession, nil
}

// func CSRFreGenerate(session collection.SessionStruct) (collection.SessionStruct, error) {
// 	coll := DB.SessionDB.Collection("session")
// 	session.Csrf = StringRand(16)

// 	filter := bson.D{{"_id", session.SessionID}}
// 	update := bson.D{
// 		{"$set", bson.D{
// 			{"csrf", session.Csrf},
// 		}},
// 	}
// 	opts := options.Update().SetUpsert(false)
// 	_, err := coll.UpdateOne(context.TODO(), filter, update, opts)
// 	return session, err
// }


// func Session(w http.ResponseWriter, r *http.Request) (collection.SessionStruct, error) {
//   var session collection.SessionStruct
//   cookie, err := r.Cookie("ss")
//   if err != nil {
//     return session, err
//   }
//   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//   defer cancel()
//   c, err := mongo.Connect(ctx, options.Client().ApplyURI(Mongo1))
//   if err != nil {
//     log.Print(err)
//   }
//   defer c.Disconnect(ctx)
//   db1 := c.Database(MongoDb1)

//   coll := db1.Collection("session")
//   filter := bson.D{{"_id", cookie.Value}}
//   opts := options.FindOne().SetProjection(bson.D{})
//   err = coll.FindOne(context.TODO(), filter, opts).Decode(&session)
//   if err != nil {
//     return session, err
//   }
//   return session, nil
// }


// func CreateNewSessionAndSetCookie(db *mongo.Database, w http.ResponseWriter) (collection.SessionStruct, error) {
// 	sessionID := common.StringRand(16)
// 	session := collection.SessionStruct{
// 		SessionID:      sessionID,
// 		Csrf:           common.StringRand(16),
// 		CreatedAt:      time.Now(),
// 		UpdatedAt:      time.Now(),
// 		PushContents:   []string{},
// 		AliasArray:     [][]string{},
// 		ChannelAliases: []collection.ChannelAlias{},
// 		IsMobile:       false, // 必要に応じて初期値設定
// 	}
// 	coll := db.Collection("session")
// 	_, err := coll.InsertOne(context.TODO(), session)
// 	if err != nil {
// 		return session, err
// 	}

// 	// Cookie 登録
// 	cookie := &http.Cookie{
// 		Name:     "ss",
// 		Value:    sessionID,
// 		MaxAge:   2592000,
// 		Secure:   true,
// 		HttpOnly: true,
// 		Path:     "/",
// 	}
// 	http.SetCookie(w, cookie)

// 	return session, nil
// }
