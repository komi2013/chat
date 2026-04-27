package controller

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"sort"
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
	fmt.Printf("DEBUG: SignInGoogle called from %s\n", r.RemoteAddr)

	cookie, err := r.Cookie("g_csrf_token")
	if err != nil {
		fmt.Printf("DEBUG: Missing CSRF cookie\n")
		http.Error(w, "missing csrf cookie", http.StatusServiceUnavailable)
		return
	}

	formCsrf := r.FormValue("g_csrf_token")
	if formCsrf != cookie.Value {
		fmt.Printf("DEBUG: CSRF mismatch. Cookie: %s, Form: %s\n", cookie.Value, formCsrf)
		http.Error(w, "csrf mismatch", http.StatusServiceUnavailable)
		return
	}

	credential := r.FormValue("credential")
	fmt.Printf("DEBUG: Received credential (first 20 chars): %s...\n", credential[:20])

	token, err := jwt.Parse(credential, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		return getGooglePublicKey(kid)
	})
	if err != nil {
		http.Error(w, "invalid token", http.StatusServiceUnavailable)
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
	isMobile := common.IsMobile(r.Header.Get("User-Agent"))

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

	var currentSessionID string
	cookieSS, cookieErr := r.Cookie("ss")
	if cookieErr == nil {
		currentSessionID = cookieSS.Value
	}

	cursor, err := collSession.Find(context.TODO(), bson.M{"userID": userID})
	if err != nil {
		http.Error(w, "session lookup fail", http.StatusInternalServerError)
		return
	}

	var sessions []collection.SessionStruct
	if err := cursor.All(context.TODO(), &sessions); err != nil {
		http.Error(w, "session decode fail", http.StatusInternalServerError)
		return
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.Before(sessions[j].UpdatedAt)
	})

	sameBrowser := false
	var previousSession collection.SessionStruct
	var tweetPosts []collection.TweetPost
	for _, s := range sessions {
		if s.SessionID == currentSessionID && currentSessionID != "" {
			previousSession = s
			sameBrowser = true
			break
		}
		tweetPosts = append(tweetPosts, s.TweetPosts...)  // prevent bad user post unlimited
	}

	newSession := collection.SessionStruct{
		SessionID:      common.StringRand(16),
		Csrf:           common.StringRand(16),
		UserID:         userID,
		IsMobile:       isMobile,
		PushContents:   previousSession.PushContents,
		Nickname:       previousSession.Nickname,
		NickImg:        previousSession.NickImg,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ChannelAliases: user.ChannelAliases,
		Mail: user.Mail,
		Telephone: user.Telephone,
		TweetPosts: tweetPosts,
	}

	// Cookie 更新（ブラウザ側に新しい ss をセット）
	http.SetCookie(w, &http.Cookie{
		Name:     "ss",
		Value:    newSession.SessionID,
		MaxAge:   2592000,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})

	// DB に新規セッションを追加
	_, err = collSession.InsertOne(context.TODO(), newSession)
	if err != nil {
		http.Error(w, "cannot create session", http.StatusInternalServerError)
		return
	}

	// =====================================================================
	// 挙動分岐
	//  - sameBrowser: 既存のこのブラウザの古い session を削除して終了
	//  - not sameBrowser: 別ブラウザなので最新 3 個まで保持（4 個目で古いものから削除）
	// =====================================================================

	if sameBrowser && currentSessionID != "" {
		// 同ブラウザ：古い自分のセッションを削除
		_, _ = collSession.DeleteOne(context.TODO(), bson.M{
			"_id": currentSessionID,
		})
		// 終了してリダイレクト
		http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
		return
	}

	// 別ブラウザ：新しい session を一覧に加えて最大 3 個を維持
	sessions = append(sessions, newSession)

	// =====================================================================
	// DELETE SESSIONS OLDER THAN 1 MONTH
	// =====================================================================
	oneMonthAgo := time.Now().AddDate(0, -1, 0) // 1 ヶ月前

	for _, s := range sessions {
	    if s.UpdatedAt.Before(oneMonthAgo) {
	        // 古いセッションは削除
	        _, _ = collSession.DeleteOne(context.TODO(), bson.M{
	            "_id": s.SessionID,
	        })
	    }
	}

	if len(sessions) > 5 {
		deleteCount := len(sessions) - 5
		for i := 0; i < deleteCount; i++ {
			_, _ = collSession.DeleteOne(context.TODO(), bson.M{
				"_id": sessions[i].SessionID,
			})
		}
	}

	if isMobile {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "csrf": "%s", "userId": "%s", "nickname": "%s", "message": "Success"}`, newSession.Csrf, userID, newSession.Nickname)
		return
	}

	http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
}
