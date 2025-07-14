package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  "log"
  // "math"
  "net/http"
  "strconv"
  "time"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

func UserEdit(w http.ResponseWriter, r *http.Request) {

	var user collection.UserStruct
  latStr := r.FormValue("latitude")
  lat, err := strconv.ParseFloat(latStr, 64)
  if err != nil || lat < -90 || lat > 90 {
    log.Printf("緯度 (latitude) が不正です: %v", latStr)
    http.Error(w, "緯度 (latitude) が不正です", http.StatusBadRequest)
    return
  }
  lngStr := r.FormValue("longitude")
  lng, err := strconv.ParseFloat(lngStr, 64)
  if err != nil || lng < -180 || lng > 180 {
    log.Printf("経度 (longitude) が不正です: %v", lngStr)
    http.Error(w, "経度 (longitude) が不正です", http.StatusBadRequest)
    return
  }
  user.Latitude = lat
  user.Longitude = lng

  var nickname collection.NicknameStruct
  nickname.Nickname = r.FormValue("nickname")
  nicknameOld := r.FormValue("nicknameOld")

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

	nickname.NickImg, err = common.ImgSave(db1, r.FormValue("nickImg"), session.UserID, nickname.Nickname, "", 3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  update := bson.M{
    "$set": user,
  }
  filterUser := bson.M{"_id": session.UserID}
  opts := options.Update().SetUpsert(true)
  collUser := db1.Collection("user")
  _, err = collUser.UpdateOne(ctx, filterUser, update, opts)
  if err != nil {
    log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
  }

  collNickname := db1.Collection("nickname")
  filterNickname := bson.M{"userID": session.UserID}
  cursor, err := collNickname.Find(context.TODO(), filterNickname)
  if err != nil {
    log.Printf("coll.Find: %v; Req: ", err, session.UserID, r.URL.Path, r.Form)
  }
  var nicknames []collection.NicknameStruct
  if err = cursor.All(context.TODO(), &nicknames); err != nil {
    log.Printf("cursor.All: %v; Req: ", err, session.UserID, r.URL.Path, r.Form)
  }
  filterName := nickname.Nickname
  nameCount := 0
  for _, nick := range nicknames {
  	if nick.Nickname == nicknameOld {
  		filterName = nick.Nickname
  	} else {
  		nameCount = nameCount + 1
  	}
  	log.Printf("UpdateOne: %v; Req:", nameCount)
  }
  message := "ユーザー情報は更新されました。 "
  if nameCount < 3 {
  	nickname.UpdatedAt = time.Now()
  	nickname.UserID = session.UserID
	  update = bson.M{
	    "$set": nickname,
	  }
	  filterNickname := bson.M{"_id": filterName}
	  opts = options.Update().SetUpsert(true)
	  collNickname := db1.Collection("nickname")
	  _, err = collNickname.UpdateOne(ctx, filterNickname, update, opts)
	  if err != nil {
	    log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
	  }
	  message = message + "ニックネーム情報は更新されました。 "
  } else {
  	message = message + "ニックネーム情報は3件以上は登録できません。 "
  }

  responseData := struct {
    Csrf         string       `json:"csrf"`
    PushContents []string     `json:"pushContents"`
    Message				string 			`json:"message"`
    // User       collection.UserStruct  `json:"user"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Message: 			message,
    // User         : user,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

