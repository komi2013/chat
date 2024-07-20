package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "net/http"
  // "strings"
  "os"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func ThreadPost(w http.ResponseWriter, r *http.Request) {
	session, err := common.Session(w,r)
	if err != nil {
  	http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
    return
	}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")

  trueAccess := false
  for _, arrayData := range session.AliasArray {
    if arrayData[0] == aliasName && arrayData[1] == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    fmt.Printf(" err %s\n", session.AliasArray, aliasName)
    return
  }

	err = r.ParseMultipartForm(10 << 20) // 最大10MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
// 「＊([^]*?)＊」（＊([^]*?)＊）/
	// fileNames := ""
	fileLinks := ""
	files := r.MultipartForm.File["files[]"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		// fileNames += "＊f＊" + fileHeader.Filename + "・＊f＊"
		// fileLinks += "「＊/upload/" + r.FormValue("channelID") + "/" + fileHeader.Filename + "＊」（＊" + fileHeader.Filename + "＊）"
		fileLinks += "＊f＊" + fileHeader.Filename + "・＊f＊"
		saveDir := "./upload/" + channelID + "/"
		os.MkdirAll(saveDir, 0755);
		dst, err := os.Create(saveDir + fileHeader.Filename)
		if err != nil {
			http.Error(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Failed to copy file", http.StatusInternalServerError)
			return
		}
	}
  jsonBytes := []byte(r.FormValue("userIDs"))
  userIDs := []string{}
  json.Unmarshal(jsonBytes, &userIDs)
  fmt.Println("userIDs:", userIDs)

  coll := db1.Collection("session")
  filter := bson.D{{
    "user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"subscription", 1}}
  opts4 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts4)
  if err != nil {
    fmt.Printf(" err %s\n", err)
  }
  var results4 []collection.SessionStruct
  if err = cursor.All(context.TODO(), &results4); err != nil {
    fmt.Printf(" err %s\n", err)
  }
  jsonBytes = []byte(r.FormValue("names"))
  names := []string{}
  json.Unmarshal(jsonBytes, &names)
  messageID := common.Base62Encode(time.Now().Unix())
  messageID = channelID + messageID + common.StringRand(1)
  for _, r4 := range results4 {
	  pushID := common.StringRand(12)
	  var arr []interface{}
	  arr = append(arr, pushID)
	  arr = append(arr, "thread")
	  arr = append(arr, messageID)
	  arr = append(arr, r.FormValue("parentID"))
	  arr = append(arr, r.FormValue("messageTxt") + fileLinks)
	  arr = append(arr, aliasName)
	  arr = append(arr, r.FormValue("aliasImg"))
	  arr = append(arr, r.FormValue("type"))
	  arr = append(arr, names)
	  arr = append(arr, r.FormValue("backID"))
	  if r.FormValue("yets") != "" {
		  jsonBytes = []byte(r.FormValue("yets"))
		  yets := [][]string{}
		  json.Unmarshal(jsonBytes, &yets)
		  fmt.Println("yets:", yets)
			arr = append(arr, yets)
	  } else {
	  	arr = append(arr, "")
	  }
	  jsonData, err := json.Marshal(arr)
	  if err != nil {
	    fmt.Println("JSON変換エラー:", err)
	  }
	  // fmt.Println(string(msgJson))

		coll = db1.Collection("push")
		document := bson.M{
	    "_id": pushID,
	    "pushJson": string(jsonData),
	    "created_at": time.Now().Format("2006-01-02 15:04:05"),
		}
		_, err = coll.InsertOne(context.TODO(), document)
		if err != nil {
		    fmt.Printf("err %s\n", err)
		}
    cursor.Decode(&r4)
    webpushSub := &webpush.Subscription{}
    json.Unmarshal([]byte(r4.Subscription), webpushSub)

    // Send Notification
    resp, err := webpush.SendNotification([]byte(string(jsonData)), webpushSub, &webpush.Options{
      Subscriber:      "example@example.com",
      VAPIDPublicKey:  common.VAPIDPublicKey,
      VAPIDPrivateKey: common.VAPIDPrivateKey,
      TTL:             30,
    })
    if err != nil {
      // TODO: Handle error
      fmt.Printf(" err %s\n", err)
    }
    defer resp.Body.Close()
  }
}
