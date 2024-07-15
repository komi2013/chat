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
  "strconv"
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

func ChannelEdit(w http.ResponseWriter, r *http.Request) {

  // r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB
  // _, err := ioutil.ReadAll(r.Body)
  // if err != nil {
  //     http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
  //     return
  // }


	_, err := common.Session(w,r)
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

	aliasImg := aliasImgSave(r.FormValue("aliasImg"), r.FormValue("channelID"), 0)
  jsonBytes := []byte(r.FormValue("aliases"))
  var aliases [][]string
  json.Unmarshal(jsonBytes, &aliases)

 //  for i := range aliases {
	// 	if aliases[i][0] == r.FormValue("aliasName") {
	// 		aliases[i][1] = aliasImg
	// 		break
	// 	}
	// }

  jsonBytes = []byte(r.FormValue("groupAliases"))
  var groupAliases [][]interface{}
  json.Unmarshal(jsonBytes, &groupAliases)

  for i := range groupAliases {
		groupAliases[i][1] = aliasImgSave(groupAliases[i][1].(string), r.FormValue("channelID"), i)
	}

  jsonBytes = []byte(r.FormValue("userIDs"))
  var userIDs []interface{}
  json.Unmarshal(jsonBytes, &userIDs)
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
		arr = append(arr, r.FormValue("channelID"))
		arr = append(arr, r.FormValue("channelName"))
		arr = append(arr, r.FormValue("description"))
		arr = append(arr, aliasImg)
		arr = append(arr, r.FormValue("aliasName"))
		arr = append(arr, time.Now().Format("2006-01-02"))
		arr = append(arr, groupAliases)
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

func aliasImgSave(aliasImg string, channelID string, i int ) string {
	imgPath := aliasImg
	if (strings.HasPrefix(aliasImg, "data:image")) {
	  base64Data := strings.Split(aliasImg, ",")[1]
	  imageData, err := base64.StdEncoding.DecodeString(base64Data)
	  if err != nil {
	      log.Println(err)
	  }
	  randPath := common.StringRand(4)
		dirPath := "/img/group/" + channelID + "/"
		os.MkdirAll("." + dirPath, 0755)
		imgPath = dirPath + strconv.Itoa(i) + "_" + randPath + ".png"
		filePath := "." + imgPath
	  err = ioutil.WriteFile(filePath, imageData, 0644)
	  if err != nil {
	      log.Println(err)
	  }
	  log.Println("PNG image file saved successfully.")
	  // imgPath = ""
	}
	return imgPath
}
