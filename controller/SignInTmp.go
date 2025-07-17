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

  sessionID := common.StringRand(16)
  cookie := &http.Cookie{
    Name:     "ss",
    Value:    sessionID,
    MaxAge:   2592000,
    Secure:   true,
    HttpOnly: true,
    Path:     "/",
  }
  http.SetCookie(w, cookie)
  userID := r.FormValue("userID")

  collUser := db1.Collection("user")
  filterUser := bson.M{"_id": userID}
  var user collection.UserStruct
  err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
  if err != nil {
    log.Printf("user FindOne: %v; Req:", err, r.URL.Path, r.Form)
  }
  coll := db1.Collection("session")
  session := collection.SessionStruct{
    SessionID: sessionID,
    UserID: userID,
    ChannelAliases: user.ChannelAliases,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
    IsMobile: common.IsMobile(r.UserAgent()),
  }
  _, err = coll.InsertOne(context.TODO(), session)
  if err != nil {
    log.Fatal(err)
  }

	coll = db1.Collection("user")
	userFilter := bson.D{{"_id", userID}}
	update := bson.D{{"$set", bson.D{
		{"signedAt", time.Now()}}}}
	opts := options.Update().SetUpsert(true)
	_, err = coll.UpdateOne(context.TODO(), userFilter, update, opts)
  // http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
}
