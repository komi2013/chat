package controller

import (
  "context"
  "encoding/base64"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func CommunityMatch (w http.ResponseWriter, r *http.Request) {

	session, err := common.Session(w,r)
	if err != nil {
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
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

  // fd.append('channelID', props.channel_id);
  // fd.append('code', props.code);
  // fd.append('aliasName', aliasName.value);
  // fd.append('toAliasName', props.toAliasName);

  // fmt.Printf("session: %v\n", session)

	aliasImg := r.FormValue("aliasImg")
	if (strings.HasPrefix(r.FormValue("aliasImg"), "data:image")) {
	  base64Data := strings.Split(r.FormValue("aliasImg"), ",")[1]
	  imageData, err := base64.StdEncoding.DecodeString(base64Data)
	  if err != nil {
	      log.Println(err)
	  }
	  randPath := common.StringRand(4)
		dirPath := "./aliasImg/"
		os.MkdirAll(dirPath, 0755)
		filePath := dirPath + randPath + r.FormValue("aliasName") + ".png"
	  err = ioutil.WriteFile(filePath, imageData, 0644)
	  if err != nil {
	      log.Println(err)
	  }
	  log.Println("PNG image file saved successfully.")
	  aliasImg = "/aliasImg/" + randPath + r.FormValue("aliasName") + ".png"
	}

	coll := db1.Collection("window")
	filter := bson.M{"_id": r.FormValue("code")}

	var window collection.WindowStruct
	err = coll.FindOne(context.TODO(), filter).Decode(&window)
	if err != nil {
		log.Print(err, "window", r.FormValue("windowID"))
	}
	fmt.Printf("Subscription: %v\n", window.Subscriptions)

  for _, subscription := range window.Subscriptions {
		var arr []interface{}
		arr = append(arr, common.StringRand(12))
		arr = append(arr, "rookie")
		arr = append(arr, window.ChannelID)
		arr = append(arr, r.FormValue("aliasName"))
		arr = append(arr, aliasImg)
		arr = append(arr, session.UserID)
		jsonData, err := json.Marshal(arr)
		if err != nil {
			fmt.Println("JSON変換エラー:", err)
		}
		// JSON 文字列を表示
		fmt.Println(string(jsonData))

    // cursor.Decode(&r)
		webpushSub := &webpush.Subscription{}
		json.Unmarshal([]byte(subscription), webpushSub)

		// Send Notification
		resp, err := webpush.SendNotification([]byte(string(jsonData)), webpushSub, &webpush.Options{
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

	coll = db1.Collection("session")
	filter = bson.M{"user_id": session.UserID}

	err = coll.FindOne(context.TODO(), filter).Decode(&window)
	if err != nil {
		log.Print(err, "window", r.FormValue("windowID"))
	}

  coll = db1.Collection("session")
  filter = bson.M{"user_id": session.UserID}
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
  log.Print("session results4 ", session.UserID)
  for _, ss := range results4 {
  	pushID := common.StringRand(12)
	  var arr []interface{}
		arr = append(arr, pushID)
		arr = append(arr, "channelJoin")
		arr = append(arr, window.ChannelID)
		arr = append(arr, r.FormValue("aliasName"))
		arr = append(arr, aliasImg)
		arr = append(arr, session.UserID)
		jsonData, err := json.Marshal(arr)
		if err != nil {
			fmt.Println("JSON変換エラー:", err)
		}
		// JSON 文字列を表示
		fmt.Println(string(jsonData))

		coll = db1.Collection("push")
		document := bson.M{
	    "_id": pushID,
	    "contents": window.Contents,
	    "pushJson": string(jsonData),
	    "created_at": time.Now().Format("2006-01-02 15:04:05"),
		}
		_, err = coll.InsertOne(context.TODO(), document)
		if err != nil {
		    fmt.Printf("err %s\n", err)
		}

    // cursor.Decode(&r)
		webpushSub := &webpush.Subscription{}
		json.Unmarshal([]byte(ss.Subscription), webpushSub)

		// Send Notification
		resp, err := webpush.SendNotification([]byte(string(jsonData)), webpushSub, &webpush.Options{
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

	// JSON 文字列を表示
	// fmt.Println(string(jsonData))
  // fmt.Fprint(w, string(jsonData))
}
