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
		fmt.Printf("DEBUG: Missing credential in form\n")
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
	fmt.Printf("DEBUG: Google Sub: %v\n", googleSub)

	collUser := common.DB.UserDB.Collection("user")
	collSession := common.DB.SessionDB.Collection("session")

	var user collection.UserStruct
	err = collUser.FindOne(context.TODO(), bson.M{"googleJWTSub": googleSub}).Decode(&user)
	if err != nil && err != mongo.ErrNoDocuments {
		fmt.Printf("DEBUG: User lookup failed: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := user.UserID
	if userID == "" {
		userID, err = common.CountUpID("userID")
		if err != nil {
			fmt.Printf("DEBUG: CountUpID failed: %v\n", err)
			http.Error(w, "cannot create user", http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUG: Generated new userID: %s\n", userID)
	} else {
		fmt.Printf("DEBUG: Found existing user: %s\n", userID)
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
		fmt.Printf("DEBUG: collUser.UpdateOne failed: %v\n", err)
		http.Error(w, "user update fail", http.StatusInternalServerError)
		return
	}

	// Fetch existing sessions
	cursor, err := collSession.Find(context.TODO(), bson.M{"userID": userID})
	var sessions []collection.SessionStruct
	var tweetPosts []collection.TweetPost
	var previousSession collection.SessionStruct
	if err == nil {
		if err := cursor.All(context.TODO(), &sessions); err == nil {
			fmt.Printf("DEBUG: Found %d existing sessions for user %s\n", len(sessions), userID)
			for _, s := range sessions {
				tweetPosts = append(tweetPosts, s.TweetPosts...)
			}
			if len(sessions) > 0 {
				sort.Slice(sessions, func(i, j int) bool {
					return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
				})
				previousSession = sessions[0]
			}
		}
	} else {
		fmt.Printf("DEBUG: collSession.Find error (non-fatal): %v\n", err)
	}

	newSession := collection.SessionStruct{
		SessionID:      common.StringRand(16),
		Csrf:           common.StringRand(16),
		UserID:         userID,
		IsMobile:       true,
		Nickname:       previousSession.Nickname,
		NickImg:        previousSession.NickImg,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ChannelAliases: user.ChannelAliases,
		Mail:           user.Mail,
		Telephone:      user.Telephone,
		TweetPosts:     tweetPosts,
		PushContents:   []string{},
	}

	fmt.Printf("DEBUG: Attempting to insert new session ID: %s\n", newSession.SessionID)

	// Set session cookie
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
		errorMsg := fmt.Sprintf("cannot create session: %v", err)
		fmt.Printf("DEBUG: CRITICAL - %s\n", errorMsg)
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return
	}

	fmt.Printf("DEBUG: Session created successfully for user %s\n", userID)

	// Clean up old sessions
	if len(sessions) > 5 {
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].UpdatedAt.Before(sessions[j].UpdatedAt)
		})
		deleteCount := len(sessions) - 5
		for i := 0; i < deleteCount; i++ {
			_, err := collSession.DeleteOne(context.TODO(), bson.M{"_id": sessions[i].SessionID})
			if err != nil {
				fmt.Printf("DEBUG: Failed to delete old session %s: %v\n", sessions[i].SessionID, err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success": true, "csrf": "%s", "userId": "%s", "nickname": "%s", "message": "Success"}`, newSession.Csrf, userID, newSession.Nickname)
}
