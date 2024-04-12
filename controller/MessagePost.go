package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "io"
  "net/http"
  "strings"
  "os"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
  // "chat/logic/quiz"
)

func MessagePost(w http.ResponseWriter, r *http.Request) {
  cookie, _ := r.Cookie("ss")
  // if err != nil {
  //  return ""
  // }

	channelID, err := primitive.ObjectIDFromHex(r.FormValue("channelID"))
	if err != nil {
		log.Fatal(err)
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

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
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

  var channel collection.ChannelStruct
  coll = db1.Collection("channel")
	filter2 := bson.D{
		{"_id", channelID},
	}
  opts2 := options.FindOne().SetProjection(bson.D{
    {"user_ids", 1},
    {"alias_array", 1},
  })
  coll.FindOne(context.TODO(), filter2, opts2).Decode(&channel)
  if err != nil {
    fmt.Printf(" err %s\n", err)
  }
  fmt.Printf("channel %+v\n", channel)
 	var yets [][]string
  for _, arrayData := range channel.AliasArray {
  	atName := "＠＠" + arrayData[0] + "・＠＠"
  	strings.Contains(r.FormValue("messageTxt"), atName)
    if strings.Contains(r.FormValue("messageTxt"), atName) {
      yets = append(yets, []string{arrayData[0], "/img/yet.png"})
    }
  }
  coll = db1.Collection("session")
  filter = bson.D{{
  	"user_id", bson.D{{"$in", channel.UserIDs}}}}
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

  coll = db1.Collection("message")
  message := collection.MessageStruct{
		ChannelID: r.FormValue("channelID"),
		MessageTxt: r.FormValue("messageTxt") + fileLinks,
		From: aliasName,
		FromImg: aliasImg,
    Task: r.FormValue("task"),
		CreatedAt: time.Now(),
	}
	insertResult, err := coll.InsertOne(context.TODO(), message)
	if err != nil {
		log.Fatal(err)
	}
	insertedID := insertResult.InsertedID.(primitive.ObjectID)

	var arr []interface{}
	arr = append(arr, "message")
	arr = append(arr, insertedID)
	arr = append(arr, channelID)
	arr = append(arr, message.MessageTxt)
	arr = append(arr, aliasName)
	arr = append(arr, aliasImg)
	arr = append(arr, time.Now())
  if message.Task != "" {
		arr = append(arr, yets)
  } else {
  	arr = append(arr, "")
  }
	msgJson, err := json.Marshal(arr)
	if err != nil {
		fmt.Println("JSON変換エラー:", err)
	}
	fmt.Println(string(msgJson))

  for _, r := range results4 {
    cursor.Decode(&r)
		webpushSub := &webpush.Subscription{}
		json.Unmarshal([]byte(r.Subscription), webpushSub)

		// Send Notification
		resp, err := webpush.SendNotification([]byte(string(msgJson)), webpushSub, &webpush.Options{
			Subscriber:      "example@example.com",
			VAPIDPublicKey:  "BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8",
			VAPIDPrivateKey: "bdSiNzUhUP6piAxLH-tW88zfBlWWveIx0dAsDO66aVU",
			TTL:             30,
		})
		if err != nil {
			// TODO: Handle error
	    fmt.Printf(" err %s\n", err)
		}
		defer resp.Body.Close()
  }

  fmt.Fprint(w, `{"Status":"1"}`)
}
