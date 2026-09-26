package common

import (
  "context"
  "errors"
  "net/http"
	"strings"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
)

func SessionGet(w http.ResponseWriter, r *http.Request) (collection.SessionStruct, error) {
  var session collection.SessionStruct
	sessionID, _, err := sessionIDFromRequest(r)
  if err != nil {
    return session, err
  }
  coll := DB.SessionDB.Collection("session")
	filter := bson.D{{"_id", sessionID}}
  opts := options.FindOne().SetProjection(bson.D{})
  err = coll.FindOne(context.TODO(), filter, opts).Decode(&session)
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

func SessionCheckTake(w http.ResponseWriter, r *http.Request, token string) (collection.SessionStruct, error) {
  var session collection.SessionStruct
	sessionID, bearerToken, err := sessionIDFromRequest(r)
  if err != nil {
    return session, err
  }
  coll := DB.SessionDB.Collection("session")
	filter := bson.D{{"_id", sessionID}}
  opts := options.FindOne().SetProjection(bson.D{})
  err = coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  if err != nil {
    return session, err
  }
	if session.Csrf != token {
		LogError("session.Csrf != token", nil)
		return session, errors.New("SessionCheckTake token error")
	}

	if time.Since(session.UpdatedAt) > 20*24*time.Hour {
		return session, errors.New("session expired")
	}	else if time.Since(session.UpdatedAt) > 10*24*time.Hour {
		session, err = SessionRegenerate(session, w)
		if err == nil && bearerToken {
			session.SessionIDRotated = true
			w.Header().Set("X-Session-Token", session.SessionID)
		}
	} else {
		// UpdateSessionTimestamp(db, session)
		session, err = CSRFcheckMake(session, token)
	}
  // session, err = CheckMakeCSRFToken(db1, session, token)
  return session, err
}

func sessionIDFromRequest(r *http.Request) (string, bool, error) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization != "" {
		parts := strings.Fields(authorization)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return "", false, errors.New("invalid Authorization header")
		}
		return parts[1], true, nil
	}

	cookie, err := r.Cookie("ss")
	if err != nil {
		return "", false, err
	}
	return cookie.Value, false, nil
}

func CSRFcheckMake(session collection.SessionStruct, token string) (collection.SessionStruct, error) {
	if session.Csrf != token {
		return session, errors.New("CSRFcheckMake token error")
	}
	session = CSRFReGenerate(session)
  // if err != nil {
		// LogError("CSRFReGenerate:", err)
  //   return session, err
  // }
	session, err := PushReGenerate(session)
	return session, err
}

func CSRFReGenerate(session collection.SessionStruct) (collection.SessionStruct) {
	randomPart := StringRand(16)
	timestamp := time.Now().Unix()
	timePart := Base62Encode(timestamp)
	session.Csrf = randomPart + timePart
	return session
}

func PushReGenerate(session collection.SessionStruct) (collection.SessionStruct, error) {
	coll := DB.SessionDB.Collection("session")
	// randomPart := StringRand(16)
	// timestamp := time.Now().Unix()
	// timePart := Base62Encode(timestamp)
	// session.Csrf = randomPart + timePart

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

	coll := DB.SessionDB.Collection("session")

	_, err := coll.InsertOne(context.TODO(), newSession)
	if err != nil {
		return session, err
	}

	_, err = coll.DeleteOne(context.TODO(), bson.M{"_id": session.SessionID})
	if err != nil {
		return session, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "ss",
		Value:    newSession.SessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})
	return newSession, nil
}

