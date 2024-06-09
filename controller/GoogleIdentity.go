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

func GoogleIdentity(w http.ResponseWriter, r *http.Request) {
  // fmt.Printf("r %#v\n", r)
  // fmt.Printf("credential %#v\n", r.FormValue("credential"))
  cookie, err := r.Cookie("g_csrf_token")
  if err != nil {
    fmt.Printf("err %#v\n", err)
  }
  if r.FormValue("g_csrf_token") != cookie.Value {
    fmt.Printf("csrf is wrong %#v\n", r.FormValue("g_csrf_token"))
  }

  token, err := jwt.Parse(r.FormValue("credential"), func(token *jwt.Token) (interface{}, error) {
    // Don't forget to validate the alg is what you expect:
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
      return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
      // !ok but no need to worry this error somehow 
    }

    // hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
    hmacSampleSecret = []byte("")
    return hmacSampleSecret, nil
  })
  claims, _ := token.Claims.(jwt.MapClaims)
  fmt.Printf("claims %#v\n", claims)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	if err != nil {
		log.Print(err)
	}
	defer c.Disconnect(ctx)
	db1 := c.Database(common.MongoDb1)

  var user collection.UserStruct
	coll := db1.Collection("user")
	filter := bson.D{{"google_jwt_sub", claims["sub"]}}
	update := bson.D{{"$set", bson.D{
		{"signed_at", time.Now()}}}}
	opts := options.Update().SetUpsert(true)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts)

  coll = db1.Collection("user")
	filter2 := bson.D{{"google_jwt_sub", claims["sub"]}}
  opts2 := options.FindOne().SetProjection(bson.D{
    {"user_id", 1},
  })
  coll.FindOne(context.TODO(), filter2, opts2).Decode(&user)
  if err != nil {
    fmt.Printf(" err %s\n", err)
  }
  fmt.Printf("user %+v\n", user)

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
  coll = db1.Collection("session")
	filter3 := bson.D{{"user_id", user.UserID}}
  opts3 := options.FindOne().SetProjection(bson.D{
    {"user_id", 1},
    {"alias_array", 1},
  }).SetSort(bson.D{
    {"created_at", -1},
	})
  err = coll.FindOne(context.TODO(), filter3, opts3).Decode(&ssAlready)

	if err == mongo.ErrNoDocuments {
		// ドキュメントが見つからない場合の処理
		fmt.Println("Document not found")
	  coll = db1.Collection("session")
	  session := collection.SessionStruct{
			SessionID: sessionID,
			UserID: user.UserID,
			CreatedAt: time.Now(),
		}
		_, err := coll.InsertOne(context.TODO(), session)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
    // エラーが発生した場合の処理
    fmt.Printf("Error: %s\n", err.Error())
	} else {
    // ドキュメントが見つかった場合の処理
    fmt.Println("Document found")
	  coll = db1.Collection("session")
	  session := collection.SessionStruct{
			SessionID: sessionID,
			UserID: user.UserID,
			AliasArray: ssAlready.AliasArray,
			CreatedAt: time.Now(),
		}
		_, err := coll.InsertOne(context.TODO(), session)
		if err != nil {
			log.Fatal(err)
		}
	}
  http.Redirect(w, r, "/pushSubscribe.html", http.StatusSeeOther)
}
