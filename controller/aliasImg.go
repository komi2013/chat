package controller

import (
  // "context"
  // "encoding/json"
  // "fmt"
  "io"
  "log"
  "mime"
  "net/http"
  // "os"
  "path/filepath"
  "strings"
  // "time"

  // "chat/common"
)

func AliasImg(w http.ResponseWriter, r *http.Request) {
	// session, err := common.Session(w,r)
	// if err != nil {
 //  	http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
 //    return
	// }
	// log.Print(session)
  // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  // defer cancel()
  // c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  // if err != nil {
  //   log.Print(err)
  // }
  // defer c.Disconnect(ctx)
  // db1 := c.Database(common.MongoDb1)
	u := strings.Split(r.URL.Path, "/")
  // if err = nil {
  //   log.Fatal(err)
  // }
  // aliasName := u[3]
  
  log.Print("log u", u[2])
  // return
  filePath := u[2]
  file, err := http.Dir("./aliasImg").Open(filePath)
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
