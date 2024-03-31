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
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"
)

func Upload(w http.ResponseWriter, r *http.Request) {
  cookie, _ := r.Cookie("ss")
  // if err != nil {
  //  return ""
  // }

	u := strings.Split(r.URL.Path, "/")
  // if err = nil {
  //   log.Fatal(err)
  // }
  aliasName := u[3]
  // fmt.Printf(" u %s\n", u)
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  var session collection.SessionStruct

  coll := db1.Collection("session")
  filter := bson.D{{"_id", cookie.Value}}
  // opts := options.FindOne().SetProjection(projection)
  opts := options.FindOne().SetProjection(bson.D{
    {"user_id", 1},
    {"alias_array", 1},
  })
  coll.FindOne(context.TODO(), filter, opts).Decode(&session)
  if err != nil {
    panic(err)
  }
  trueAccess := false
  for _, arrayData := range session.AliasArray {
    if arrayData[0] == aliasName {
      trueAccess = true
    }
  }
  if !trueAccess {
    fmt.Printf(" err %s\n", session.AliasArray, aliasName)
    return
  }

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
