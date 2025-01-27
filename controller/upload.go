package controller

import (
  "context"
  // "encoding/json"
  "fmt"
  "io"
  "log"
  "mime"
  "net/http"
  // "os"
  "path/filepath"
  "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  // "chat/collection"
  "chat/common"
)

func Upload(w http.ResponseWriter, r *http.Request) {
	u := strings.Split(r.URL.Path, "/")
  aliasName := u[3]
  
	fmt.Printf("/upload/I0JH/ u %s\n", u[1], u[2]) // upload I0JH
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck: %v; Req: ", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}

  trueAccess := false
  // for _, d := range session.ChannelAliases {
  //   if d.Alias == updatedBy && d.ChannelID == channelID {
  //     trueAccess = true
  //   }
  // }
  // if !trueAccess {
  //   log.Printf("ChannelAliases !trueAccess: %v; Req: ", session.ChannelAliases, updatedBy, channelID, r.URL.Path, r.Form)
  //   return
  // }

  // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  // defer cancel()
  // c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  // if err != nil {
  //   log.Print(err)
  // }
  // defer c.Disconnect(ctx)
  // db1 := c.Database(common.MongoDb1)

  // var session collection.SessionStruct

  // coll := db1.Collection("session")
  // filter := bson.D{{"_id", cookie.Value}}
  // // opts := options.FindOne().SetProjection(projection)
  // opts := options.FindOne().SetProjection(bson.D{
  //   {"user_id", 1},
  //   {"alias_array", 1},
  // })
  // coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  // if err != nil {
  //   panic(err)
  // }
  // trueAccess := false
  // for _, arrayData := range session.AliasArray {
  //   if arrayData[0] == aliasName {
  //     trueAccess = true
  //   }
  // }
  // if !trueAccess {
  //   fmt.Printf(" err %s\n", session.AliasArray, aliasName)
  //   return
  // }

  fmt.Printf(" u %s\n", aliasName, session, trueAccess)

	filePath := u[2] + "/" + u[4]

	file, err := http.Dir("./upload").Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	contentType := "application/octet-stream"
	if ext := filepath.Ext(filePath); ext != "" {
		contentType = mime.TypeByExtension(ext)
	}

	w.Header().Set("Content-Type", contentType)

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

}
