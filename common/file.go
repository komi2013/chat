package common

import (
	"context"
  "encoding/base64"
  "encoding/json"
  "io/ioutil"
  "log"
  "os"
  "runtime"
  "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"

)

func ImgSave(img string, userID string, name string, db1 *mongo.Database) string {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
  imgPath := img
  if (strings.HasPrefix(img, "data:image")) {
    base64Data := strings.Split(img, ",")[1]
    imageData, err := base64.StdEncoding.DecodeString(base64Data)
    if err != nil {
			log.Printf("base64.StdEncoding: %v; From:", err, funcName)
			return ""
    }
    dirPath := "/img/user/" + userID + "/"
    os.MkdirAll("." + dirPath, 0755)
    imgPath = dirPath + userID + "_" + name + ".png"
    filePath := "." + imgPath
    err = ioutil.WriteFile(filePath, imageData, 0644)
    if err != nil {
      log.Printf("ioutil.WriteFile: %v; From:", err, funcName)
      return ""
    }
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Printf("os.Stat: %v; From:", err, funcName)
			return ""
		}
		fileSize := fileInfo.Size()

		coll := db1.Collection("file")
		document := bson.M{
	    "user_id": userID,
	    "img_path": imgPath,
	    "file_size": fileSize,
	    "created_at": time.Now().Format("2006-01-02 15:04:05"),
		}
		_, err = coll.InsertOne(context.TODO(), document)
		if err != nil {
			log.Printf("coll.InsertOne: %v; From:", err, funcName)
			return ""
		}
  }
  return imgPath
}

func ImgsSaveBK(imgs string, userID string, name string, db1 *mongo.Database) []string {
	var paths []string
  if err := json.Unmarshal([]byte(imgs), &paths); err != nil {
		pc, _, _, _ := runtime.Caller(1)
		funcName := runtime.FuncForPC(pc).Name()
  	log.Printf("paths: %v; From:", err, funcName)
    return []string{}
  }
  var imgPaths []string
  for _, imgPath := range paths {
	  if (strings.HasPrefix(imgPath, "data:image")) {
	    base64Data := strings.Split(imgPath, ",")[1]
	    imageData, err := base64.StdEncoding.DecodeString(base64Data)
	    if err != nil {
				pc, _, _, _ := runtime.Caller(1)
				funcName := runtime.FuncForPC(pc).Name()
	      log.Printf("base64.StdEncoding: %v; From:", err, funcName)
	      return []string{}
	    }

	    dirPath := "/img/" + userID + "/"
	    os.MkdirAll("." + dirPath, 0755)
	    imgPath = dirPath + userID + "_" + name + ".png"
	    filePath := "." + imgPath
	    err = ioutil.WriteFile(filePath, imageData, 0644)
	    if err != nil {
				pc, _, _, _ := runtime.Caller(1)
				funcName := runtime.FuncForPC(pc).Name()
	      log.Printf("ioutil.WriteFile: %v; From:", err, funcName)
	      return []string{}
	    }
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				pc, _, _, _ := runtime.Caller(1)
				funcName := runtime.FuncForPC(pc).Name()
	      log.Printf("os.Stat: %v; From:", err, funcName)
	      return []string{}
			}
			fileSize := fileInfo.Size()

			coll := db1.Collection("file")
			document := bson.M{
		    "user_id": userID,
		    "img_path": imgPath,
		    "file_size": fileSize,
		    "created_at": time.Now().Format("2006-01-02 15:04:05"),
			}
			_, err = coll.InsertOne(context.TODO(), document)
			if err != nil {
				pc, _, _, _ := runtime.Caller(1)
				funcName := runtime.FuncForPC(pc).Name()
				log.Printf("coll.InsertOne: %v; From:", err, funcName)
				return []string{}
			}
	  }
	  imgPaths = append(imgPaths, imgPath)
  }

  return imgPaths
}
