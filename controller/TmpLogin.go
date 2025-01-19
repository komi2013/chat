package controller

import (
  "context"
  "fmt"
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

func TmpLogin(w http.ResponseWriter, r *http.Request) {

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
  var ssAlready collection.SessionStruct
  coll := db1.Collection("session")
  filter3 := bson.D{{"user_id", userID}}
  opts3 := options.FindOne().SetProjection(bson.D{
    {"user_id", 1},
    {"alias_channels", 1},
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
      UserID: userID,
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
      UserID: userID,
      ChannelAliases: ssAlready.ChannelAliases,
      CreatedAt: time.Now(),
    }
    _, err := coll.InsertOne(context.TODO(), session)
    if err != nil {
      log.Fatal(err)
    }
  }
  http.Redirect(w, r, "/pushSubscription/", http.StatusSeeOther)
}
