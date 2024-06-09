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

func ChannelAdd(w http.ResponseWriter, r *http.Request) {
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

  //  channelName: 改めてグループ
	// description: <p>ここはディスクリプション</p>
	// aliasName: コマツ
	// aliasImg: 
	// jsonBytes := []byte(r.FormValue("aliasArray"))
	// var aliases [][]string
	// json.Unmarshal(jsonBytes, &aliases)

	// aliases
 //  fd.append('aliasName', aliasName.value);
 //  fd.append('aliasImg', aliasImg.value);
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

	alias := []string{
		r.FormValue("aliasName"),
		aliasImg,
		session.UserID}

  // for _, d := range aliases {
  if !isAliasExist(r.FormValue("aliasName"), session.AliasArray) {
    session.AliasArray = append(session.AliasArray, []string{
    	r.FormValue("aliasName"), aliasImg})
  }
  // }
  coll := db1.Collection("session")
  filter := bson.D{{"_id", session.SessionID}}
  update := bson.D{{"$set", bson.D{
      {"alias_array", session.AliasArray},
  }}}
  coll.UpdateOne(context.TODO(), filter, update)

	channelID := common.StringRand(4)
	userIDs := []string{session.UserID}

	coll = db1.Collection("session")
  filter = bson.D{{
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

  pushID := common.StringRand(12)
	var arr []interface{}
	arr = append(arr, pushID)
	arr = append(arr, "channel")
	arr = append(arr, channelID)
	arr = append(arr, r.FormValue("channelName"))
	arr = append(arr, r.FormValue("description"))
	arr = append(arr, alias)
	arr = append(arr, r.FormValue("aliasName"))
	arr = append(arr, time.Now().Format("2006-01-02"))
	arr = append(arr, 1) // 1 = add, 2 = edit

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
	    fmt.Printf("err %s\n", err)
	}
  for _, r := range results4 {
    cursor.Decode(&r)
		webpushSub := &webpush.Subscription{}
		json.Unmarshal([]byte(r.Subscription), webpushSub)
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

func isAliasExist(alias string, aliases [][]string) bool {
    for _, a := range aliases {
        if a[0] == alias {
            return true
        }
    }
    return false
}
