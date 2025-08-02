package controller

import (
	"context"
  "fmt"
  "log"
  // "encoding/json"
  "net/http"
  "time"

  "github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"

)

var hmacSampleSecret []byte

func SignInGoogle(w http.ResponseWriter, r *http.Request) {
  cookie, err := r.Cookie("g_csrf_token")
	if err != nil {
		log.Printf("g_csrf_token: %v; Req: ", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}


  if r.FormValue("g_csrf_token") != cookie.Value {
		log.Printf("csrf is wrong: %v; Req: ", err, r.URL.Path, r.Form)
  	http.Error(w, "csrf is wrong", http.StatusServiceUnavailable)
    return
  }

  token, err := jwt.Parse(r.FormValue("credential"), func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
      return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
    }

    hmacSampleSecret = []byte("")
    return hmacSampleSecret, nil
  })
	if err != nil {
		log.Printf("Unexpected signing method: %v; Req: ", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}

  claims, _ := token.Claims.(jwt.MapClaims)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	if err != nil {
		log.Print(err)
	}
	defer c.Disconnect(ctx)
	db1 := c.Database(common.MongoDb1)

	var user collection.UserStruct
	collUser := db1.Collection("user")

	filterUser := bson.D{{"googleJWTSub", claims["sub"]}}

	err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			userID, err := common.CountUpID("2")
			if err != nil {
				log.Printf("CountUpID userID error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			newUser := bson.D{
				{"googleJWTSub", claims["sub"]},
				{"userID", userID},
				{"signedAt", time.Now()},
			}

			_, err = collUser.InsertOne(context.TODO(), newUser)
			if err != nil {
				log.Printf("InsertOne new user error: %v", err)
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		} else {
			log.Printf("FindOne user error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// 既存ユーザー: signedAtのみ更新
		update := bson.D{{"$set", bson.D{{"signedAt", time.Now()}}}}
		_, err = collUser.UpdateOne(context.TODO(), filterUser, update)
		if err != nil {
			log.Printf("UpdateOne user error: %v; Req: %v %v", err, r.URL.Path, r.Form)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}


  collUser = db1.Collection("user")
	filterUser2 := bson.D{{"googleJWTSub", claims["sub"]}}
  optsUser := options.FindOne().SetProjection(bson.D{
    {"userID", 1},
  })
  collUser.FindOne(context.TODO(), filterUser2, optsUser).Decode(&user)

  sessionID := common.StringRand(16)
  cookie = &http.Cookie{
    Name:     "ss",
    Value:    sessionID,
    MaxAge:   2592000,
    Secure:   true,
    HttpOnly: true,
    Path:     "/",
  }
  http.SetCookie(w, cookie)

  var ssAlready collection.SessionStruct
  collSession := db1.Collection("session")
	filterSession := bson.D{{"userID", user.UserID}}
  optsSession := options.FindOne().SetProjection(bson.D{
    {"userID", 1},
    {"channelAliases", 1},
  }).SetSort(bson.D{
    {"createdAt", -1},
	})
  err = collSession.FindOne(context.TODO(), filterSession, optsSession).Decode(&ssAlready)

	if err == mongo.ErrNoDocuments {
	  collSession = db1.Collection("session")
	  session := collection.SessionStruct{
			SessionID: sessionID,
			UserID: user.UserID,
			CreatedAt: time.Now(),
		}
		_, err := collSession.InsertOne(context.TODO(), session)
		if err != nil {
			log.Printf("collSession.InsertOne: %v; Req: ", err, r.URL.Path, r.Form)
	  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
	    return
		}
	} else if err != nil {
		log.Printf("else if err != nil: %v; Req: ", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	} else {
	  collSession = db1.Collection("session")
	  session := collection.SessionStruct{
			SessionID: sessionID,
			UserID: user.UserID,
			ChannelAliases: ssAlready.ChannelAliases,
			CreatedAt: time.Now(),
		}
		_, err := collSession.InsertOne(context.TODO(), session)
		if err != nil {
			log.Printf("already collSession.InsertOne: %v; Req: ", err, r.URL.Path, r.Form)
	  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
	    return
		}
	}
  http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
}
