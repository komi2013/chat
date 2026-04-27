package controller

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/collection"
	"chat/common"
)

func SignInGoogleMobile(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: SignInGoogleMobile called from %s\n", r.RemoteAddr)

	// In mobile, we trust the g_csrf_token form value.
	formCsrf := r.FormValue("g_csrf_token")
	if formCsrf == "" {
		fmt.Printf("DEBUG: Missing CSRF in form\n")
		http.Error(w, "missing csrf", http.StatusBadRequest)
		return
	}

	credential := r.FormValue("credential")
	if credential == "" {
		http.Error(w, "missing credential", http.StatusBadRequest)
		return
	}

	token, err := jwt.Parse(credential, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		return getGooglePublicKey(kid)
	})
	if err != nil {
		fmt.Printf("DEBUG: Token validation failed: %v\n", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	googleSub := claims["sub"]

	collUser := common.DB.UserDB.Collection("user")
	collSession := common.DB.SessionDB.Collection("session")

	var user collection.UserStruct
	err = collUser.FindOne(context.TODO(), bson.M{"googleJWTSub": googleSub}).Decode(&user)
	if err != nil && err != mongo.ErrNoDocuments {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := user.UserID
	if userID == "" {
		userID, err = common.CountUpID("userID")
		if err != nil {
			http.Error(w, "cannot create user", http.StatusInternalServerError)
			return
		}
	}

	_, err = collUser.UpdateOne(
		context.TODO(),
		bson.M{"googleJWTSub": googleSub},
		bson.M{
			"$set": bson.M{
				"_id":          userID,
				"googleJWTSub": googleSub,
				"signedAt":     time.Now(),
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		http.Error(w, "user update fail", http.StatusInternalServerError)
		return
	}

	// Fetch existing sessions to migrate tweet posts (rate limiting)
	cursor, _ := collSession.Find(context.TODO(), bson.M{"userID": userID})
	var sessions []collection.SessionStruct
	var tweetPosts []collection.TweetPost
	if err := cursor.All(context.TODO(), &sessions); err == nil {
		for _, s := range sessions {
			tweetPosts = append(tweetPosts, s.TweetPosts...)
		}
	}

	newSession := collection.SessionStruct{
		SessionID:      common.StringRand(16),
		Csrf:           common.StringRand(16),
		UserID:         userID,
		IsMobile:       true,
		Nickname:       user.Nickname,
		NickImg:        user.NickImg,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ChannelAliases: user.ChannelAliases,
		Mail:           user.Mail,
		Telephone:      user.Telephone,
		TweetPosts:     tweetPosts,
	}

	// Set session cookie for subsequent requests
	http.SetCookie(w, &http.Cookie{
		Name:     "ss",
		Value:    newSession.SessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})

	_, err = collSession.InsertOne(context.TODO(), newSession)
	if err != nil {
		http.Error(w, "cannot create session", http.StatusInternalServerError)
		return
	}

	// Clean up old sessions (keep latest 5)
	if len(sessions) > 5 {
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].UpdatedAt.Before(sessions[j].UpdatedAt)
		})
		deleteCount := len(sessions) - 5
		for i := 0; i < deleteCount; i++ {
			_, _ = collSession.DeleteOne(context.TODO(), bson.M{"_id": sessions[i].SessionID})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success": true, "csrf": "%s", "userId": "%s", "nickname": "%s", "message": "Success"}`, newSession.Csrf, userID, newSession.Nickname)
}
