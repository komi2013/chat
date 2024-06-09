package controller

import (
  "context"
  // "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"
)

func PushGet(w http.ResponseWriter, r *http.Request) {
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

  var push collection.PrivateStruct
	coll := db1.Collection("push")
	filter := bson.M{"_id": r.FormValue("pushID")}
	err = coll.FindOne(context.TODO(), filter).Decode(&push)
	if err != nil {
		log.Print(err, "push", r.FormValue("pushID"))
	}

  coll = db1.Collection("push")
  _, err = coll.DeleteOne(context.Background(), bson.M{"_id": r.FormValue("pushID")})
  if err != nil {
      log.Print(err)
      return
  }
 //  var arr []interface{}
	// arr = append(arr, "got")
	// arr = append(arr, push.Contents)
	// jsonData, err := json.Marshal(arr)
	// if err != nil {
	// 	fmt.Println("JSON変換エラー:", err)
	// }
  fmt.Fprint(w, push.Contents)
}

