package common

import (
  "context"
  // "errors"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
)

func Session(w http.ResponseWriter, r *http.Request) (collection.SessionStruct, error) {
  var session collection.SessionStruct
  cookie, err := r.Cookie("ss")
  if err != nil {
    return session, err
  }
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(MongoDb1)

  coll := db1.Collection("session")
  filter := bson.D{{"_id", cookie.Value}}
  opts := options.FindOne().SetProjection(bson.D{})
  err = coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  if err != nil {
    return session, err
  }
  return session, nil
}

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
  session, err = CheckMakeCSRFToken(db1, session, token)
  return session, err
}

func CheckMakeCSRFToken(db1 *mongo.Database, session collection.SessionStruct, token string) (collection.SessionStruct, error) {
	// if session.Csrf != token {
	// 	return session, errors.New("token error ")
	// }
	session, err := GenerateCSRFToken(db1, session)
	return session, err
}

func GenerateCSRFToken(db1 *mongo.Database, session collection.SessionStruct) (collection.SessionStruct, error) {
	coll := db1.Collection("session")
	token := StringRand(16)
	session.Csrf = token
	session.UpdatedAt = time.Now()
	filter := bson.D{{"_id", session.SessionID}}
	update := bson.D{{"$set", session}}
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


