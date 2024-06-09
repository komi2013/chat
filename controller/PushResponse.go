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

  // "chat/collection"
  "chat/common"
)

func PushResponse(w http.ResponseWriter, r *http.Request) {
  // session, err := common.Session(w,r)
  // if err != nil {
  //   log.Print(err)
  //   return
  // }
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)
  coll := db1.Collection("push")
  _, err = coll.DeleteOne(context.Background(), bson.M{"_id": r.FormValue("pushID")})
  if err != nil {
      log.Print(err)
      return
  }
  fmt.Fprint(w, `{"Status":"deleted"}`)
}

