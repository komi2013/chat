package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "net/http"
  "os"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func ChannelEdit(w http.ResponseWriter, r *http.Request) {
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
  channelID := r.FormValue("channelID")
  if r.FormValue("channelID") == "" {
  	channelID = common.StringRand(4)
  }
	aliasImg := common.AliasImgSave(r.FormValue("aliasImg"), channelID, 0)

	fileLinks := ""
	files := r.MultipartForm.File["files[]"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()
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
  var userIDs []interface{}
  json.Unmarshal(jsonBytes, &userIDs)
  userIDs = append(userIDs, session.UserID)
  fmt.Printf("userIDs %s\n", userIDs)
	coll := db1.Collection("session")
  filter := bson.D{{
  	"user_id", bson.D{{"$in", userIDs}}}}
	project := bson.D{{"subscription", 1}}
	opts4 := options.Find().SetProjection(project)
	cursor, err := coll.Find(context.TODO(), filter, opts4)
	if err != nil {
	    fmt.Printf("err %s\n", err)
	}
	var results4 []collection.SessionStruct
	if err = cursor.All(context.TODO(), &results4); err != nil {
	    fmt.Printf("err %s\n", err)
	}

  for _, d := range results4 {
	  pushID := common.StringRand(12)
		var arr []interface{}
		arr = append(arr, pushID)
		arr = append(arr, "channelEdit")
		arr = append(arr, channelID)
		arr = append(arr, r.FormValue("channelName"))
		arr = append(arr, r.FormValue("description") + fileLinks)
		arr = append(arr, aliasImg)
		arr = append(arr, r.FormValue("aliasName"))
		arr = append(arr, time.Now().Format("2006-01-02"))
		msgJson, err := json.Marshal(arr)
		if err != nil {
			fmt.Println("JSON変換エラー:", err)
		}
		fmt.Println(string(msgJson))

		coll = db1.Collection("push")
		document := bson.M{
	    "_id": pushID,
	    "contents": string(msgJson),
	    "created_at": time.Now().Format("2006-01-02 15:04:05"),
		}
		_, err = coll.InsertOne(context.TODO(), document)
		if err != nil {
		    fmt.Printf("InsertOne %s\n", err)
		}

    cursor.Decode(&d)
		webpushSub := &webpush.Subscription{}
		json.Unmarshal([]byte(d.Subscription), webpushSub)
		resp, err := webpush.SendNotification([]byte(string(msgJson)), webpushSub, &webpush.Options{
			Subscriber:      "example@example.com",
      VAPIDPublicKey:  common.VAPIDPublicKey,
      VAPIDPrivateKey: common.VAPIDPrivateKey,
			TTL:             30,
		})
		if err != nil {
	    fmt.Printf(" err %s\n", err)
		}
		defer resp.Body.Close()
  }
  fmt.Fprint(w, `{"Status":"1"}`)
}
