package controller

import (
	"context"
  "fmt"
  "log"
  "net/http"
  "time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	 "chat/common"
)

func SetCookie(w http.ResponseWriter, r *http.Request) {
  log.Println(r.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	if err != nil {
		log.Print(err)
	}
	defer c.Disconnect(ctx)
	db1 := c.Database(common.MongoDb1)

	coll := db1.Collection("session")

	filter := bson.D{{"_id", r.FormValue("userID")}}
	// arrAlias := []string{r.FormValue("userID")}
	update := bson.D{{"$set", bson.D{
		{"user_id", r.FormValue("userID")},
		{"alias_names", []string{r.FormValue("alias")}},
		{"subscription", r.FormValue("subscription")},
		{"created_at", time.Now()}}}}
	opts := options.Update().SetUpsert(true)
	_, err = coll.UpdateOne(context.TODO(), filter, update, opts)

  cookie := &http.Cookie{
    Name:     "ss",
    Value:    r.FormValue("userID"),
    MaxAge:   101556952,
    Secure:   true,
    HttpOnly: true,
    Path:     "/",
  }
  http.SetCookie(w, cookie)
  
  fmt.Printf("%s", r.FormValue("userID"))

  fmt.Fprint(w, `[1]`)
}
