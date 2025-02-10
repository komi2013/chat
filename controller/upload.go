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
	u := strings.Split(r.URL.Path, "/")
	if len(u) < 6 {
		http.Error(w, "invalid URL path or File not found", http.StatusNotFound)
		return 
	}
	fileType := u[2]
	channelID := u[3]
	fileID := u[4]
	aliasName := u[5]
	if channelID == "" || fileID == "" {
		http.Error(w, "invalid URL path or File not found", http.StatusNotFound)
		return 
	}
	ext := filepath.Ext(aliasName)
	if fileType == "img" && ext != ".png" {
		http.Error(w, "invalid URL path or File not found", http.StatusNotFound)
		return 
	}
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

	session, err := common.SessionGet(db1, w, r)
	if err != nil {
		log.Printf("SessionGet: %v; Req: ", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}

  trueAccess := false
  for _, d := range session.ChannelAliases {
    if d.ChannelID == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    log.Printf("!trueAccess: %v; Req: ", session.ChannelAliases, r.URL.Path)
    http.Error(w, "no access right for file", http.StatusNotFound)
    return
  }
  // log.Printf("!trueAccess: %v; Req: ", fileID)
  if fileType == "file" {
		var fileData collection.FileStruct
		coll := db1.Collection("file")
		filter := bson.M{"_id": fileID}
		err = coll.FindOne(ctx, filter).Decode(&fileData)
		if err != nil {
			log.Printf("fileData: %v; Req: ", err, fileID, r.URL.Path)
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
    found := false
    for _, userID := range fileData.AvailableBy {
      if userID == session.UserID {
        found = true
        break
      }
    }
    if !found {
      log.Printf("fileData access denied for user %s; Req: ", session.UserID, r.URL.Path)
      http.Error(w, "file not found", http.StatusNotFound)
      return
    }
  }

	filePath := fmt.Sprintf("%s/%s/%s", channelID, fileID, aliasName)
	file, err := http.Dir("./upload_data/" + fileType).Open(filePath)
	if err != nil {
		log.Printf("File not found: %v; Req: ", err, r.URL.Path, r.Form)
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
