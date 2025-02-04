package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"
  // "unicode/utf8"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func StoreSelect(w http.ResponseWriter, r *http.Request) {
	var userIDs []string
  if err := json.Unmarshal([]byte(r.FormValue("userIDs")), &userIDs); err != nil {
  	log.Printf("userIDs: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON userIDs", http.StatusBadRequest)
    return
  }

  var param interface{}
  if err := json.Unmarshal([]byte(r.FormValue("param")), &param); err != nil {
  	log.Printf("param: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON param", http.StatusBadRequest)
    return
  }

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")
  targetStore := r.FormValue("targetStore")

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck: %v; Req: ", err, r.URL.Path, r.Form)
  	http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
	}

  trueAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    log.Printf("ChannelAliases !trueAccess: %v; Req: ", session.ChannelAliases, aliasName, channelID, r.URL.Path, r.Form)
    return
  }

  coll := db1.Collection("session")
  filter := bson.D{{
    "user_id", bson.D{{"$in", userIDs}}}}
  project := bson.D{{"subscription", 1},{"user_id", 1}}
  opts4 := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts4)
  if err != nil {
    fmt.Printf(" err %s\n", err)
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
    fmt.Printf(" err %s\n", err)
  }

	pushSelectIDMap := make(map[string]string)
	for _, r4 := range sessions {
		// UserID に対応する pushSelectID が既に生成済みか確認
		pushSelectID, exists := pushSelectIDMap[r4.UserID]
		if !exists {
			// まだ生成されていない場合、新しい pushSelectID を生成して保存
			pushSelectID = common.StringRand(12)
			pushSelectIDMap[r4.UserID] = pushSelectID

			// push_select コレクションに挿入
			coll = db1.Collection("push_select")
			document := bson.M{
				"_id":        pushSelectID,
				"user_id":    r4.UserID,
				"target_store":    targetStore,
				"created_at": time.Now().Format("2006-01-02 15:04:05"),
			}
			_, err := coll.InsertOne(context.TODO(), document)
			if err != nil {
				fmt.Printf("push_select error %s\n", err)
				continue
			}
		}

		var arr []interface{}
		arr = append(arr, pushSelectID)
		arr = append(arr, "storeSelect")
		arr = append(arr, channelID)
		arr = append(arr, aliasName)
		arr = append(arr, targetStore)
		arr = append(arr, param)

		jsonD, err := json.Marshal(arr)
		if err != nil {
			fmt.Println("arr JSON変換エラー:", err)
			continue
		}

		webpushSub := &webpush.Subscription{}
		err = json.Unmarshal([]byte(r4.Subscription), webpushSub)
		if err != nil {
			fmt.Printf("Subscription Unmarshal error: %s\n", err)
			continue
		}

		resp, err := webpush.SendNotification([]byte(string(jsonD)), webpushSub, &webpush.Options{
			Subscriber:      "example@example.com",
			VAPIDPublicKey:  common.VAPIDPublicKey,
			VAPIDPrivateKey: common.VAPIDPrivateKey,
			TTL:             30,
		})
		if err != nil {
			fmt.Printf("Notification send error: %s\n", err)
			continue
		}
		defer resp.Body.Close()
	}
	responseData := struct {
		Csrf         string        `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
