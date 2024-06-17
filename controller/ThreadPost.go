package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "net/http"
  "strings"
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

  aliasName := r.FormValue("aliasName")

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  trueAccess := false
  var aliasImg string
  for _, arrayData := range session.AliasArray {
    if arrayData[0] == aliasName {
      aliasImg = arrayData[1]
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
		saveDir := "./upload/" + r.FormValue("channelID") + "/"
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
  jsonBytes := []byte(r.FormValue("allAliases"))
  allAliases := [][]string{}
  userIDs := []string{}
  json.Unmarshal(jsonBytes, &allAliases)
 	var yets [][]string
  for _, arrayData := range allAliases {
  	atName := "＠＠" + arrayData[0] + "・＠＠"
  	strings.Contains(r.FormValue("messageTxt"), atName)
    if strings.Contains(r.FormValue("messageTxt"), atName) {
      yets = append(yets, []string{arrayData[0], "/img/yet.png"})
    }
    userIDs = append(userIDs, arrayData[2])
  }
  fmt.Println("userIDs:", userIDs)
  fmt.Println("unixTime:", time.Now().Unix())

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
  messageID := common.Base62Encode(time.Now().Unix())
  for _, r4 := range results4 {
	  pushID := common.StringRand(12)
	  var arr []interface{}
	  arr = append(arr, pushID)
	  arr = append(arr, "thread")
	  arr = append(arr, messageID)
	  arr = append(arr, r.FormValue("parentID"))
	  arr = append(arr, r.FormValue("messageTxt") + fileLinks)
	  arr = append(arr, aliasName)
	  arr = append(arr, aliasImg)
	  arr = append(arr, time.Now())
	  arr = append(arr, r.FormValue("channelID"))
	  arr = append(arr, r.FormValue("type"))
	  arr = append(arr, r.FormValue("names"))
	  arr = append(arr, r.FormValue("backID"))
	  if r.FormValue("task") != "" {
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
