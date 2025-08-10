package controller

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/collection"
	"chat/common"
)

func getGooglePublicKey(kid string) (*rsa.PublicKey, error) {
    ctx := context.Background()
    set, err := jwk.Fetch(ctx, "https://www.googleapis.com/oauth2/v3/certs")
    if err != nil {
        return nil, fmt.Errorf("failed to fetch Google's certs: %v", err)
    }

    key, found := set.LookupKeyID(kid)
    if !found {
        return nil, fmt.Errorf("unable to find key %q", kid)
    }

    var pubkey rsa.PublicKey
    if err := key.Raw(&pubkey); err != nil {
        return nil, fmt.Errorf("failed to create public key: %v", err)
    }
    return &pubkey, nil
}

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
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		return getGooglePublicKey(kid)
	})
	if err != nil {
		log.Printf("Token parse error: %v; Req: ", err, r.URL.Path, r.Form)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	if err != nil {
		log.Printf("mongo.Connect(ctx: %v; Req: ", err, r.URL.Path, r.Form)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer c.Disconnect(ctx)
	db1 := c.Database(common.MongoDb1)

	var user collection.UserStruct
	filterUser := bson.D{{"googleJWTSub", claims["sub"]}}
	collUser := db1.Collection("user")
	err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
	var userID string
	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("FindOne user error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		userID = user.UserID
	}
	var oldSession collection.SessionStruct
	isMobile := common.IsMobile(r.Header.Get("User-Agent"))
	if userID != "" {
		collSession := db1.Collection("session")
		filterSession := bson.D{
			{"userID", user.UserID},
			{"isMobile", isMobile},
		}
		opts := options.FindOne().SetProjection(bson.D{
			{"_id", 1},
			{"pushContents", 1},
		})
		err = collSession.FindOne(context.TODO(), filterSession, opts).Decode(&oldSession)
		log.Printf("oldSession: %v", oldSession)
		if err != nil && err != mongo.ErrNoDocuments {
			log.Printf("FindOne session error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if userID == "" {
		userID, err = common.CountUpID("userID")
		if err != nil {
			log.Printf("CountUpID userID error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	update := bson.D{{"$set", bson.D{
		{"_id", userID},
		{"googleJWTSub", claims["sub"]},
		{"signedAt", time.Now()},
	}}}
	opts := options.Update().SetUpsert(true)
	_, err = collUser.UpdateOne(context.TODO(), filterUser, update, opts)
	if err != nil {
		log.Printf("coll.UpdateOne user error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newSession := oldSession
	newSession.SessionID = common.StringRand(16)
	newSession.Csrf = common.StringRand(16)
	newSession.CreatedAt = time.Now()
	newSession.UpdatedAt = time.Now()
	newSession.UserID = userID
	newSession.ChannelAliases = user.ChannelAliases
	newSession.PushContents = oldSession.PushContents
	newSession.IsMobile = isMobile
	newCookie := &http.Cookie{
		Name:     "ss",
		Value:    newSession.SessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, newCookie)
	collSession := db1.Collection("session")
	_, err = collSession.InsertOne(context.TODO(), newSession)
	if err != nil {
		log.Printf("coll.InsertOne session error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = collSession.DeleteOne(context.TODO(), bson.M{"_id": oldSession.SessionID})
	if err != nil {
		log.Printf("coll.DeleteOne session error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
}
