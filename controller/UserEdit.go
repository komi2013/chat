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
  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

func UserEdit(w http.ResponseWriter, r *http.Request) {

	var user collection.UserStruct
  latStr := r.FormValue("latitude")
  lat, err := strconv.ParseFloat(latStr, 64)
  if err != nil || lat < -90 || lat > 90 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "緯度 (latitude) が不正です", http.StatusOK)
		return
  }
  lngStr := r.FormValue("longitude")
  lng, err := strconv.ParseFloat(lngStr, 64)
  if err != nil || lng < -180 || lng > 180 {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "経度 (longitude) が不正です", http.StatusOK)
    return
  }

  user.Latitude = lat
  user.Longitude = lng

  mail := r.FormValue("mail")
  telephone := r.FormValue("telephone")

  user.Mail = mail
  user.Telephone = telephone

  var nickname collection.NicknameStruct
  nickname.Nickname = r.FormValue("nickname")
  nicknameOld := r.FormValue("nicknameOld")

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheckTake: %v; Req: ", err, r.URL.Path)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

	nickname.NickImg, err = common.ImgSave(r.FormValue("nickImg"), session.UserID, nickname.Nickname, "", 3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  message := "ユーザー情報は更新されました "
  if nickname.Nickname != "" {
  	collNickname := common.DB.NicknameDB.Collection("nickname")
	  filterNickname := bson.M{"userID": session.UserID}
	  cursor, err := collNickname.Find(context.TODO(), filterNickname)
	  if err != nil {
	    log.Printf("coll.Find: %v; Req: ", err, session.UserID, r.URL.Path, r.Form)
	    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
	    return
	  }
	  var nicknames []collection.NicknameStruct
	  if err = cursor.All(context.TODO(), &nicknames); err != nil {
	    log.Printf("cursor.All: %v; Req: ", err, session.UserID, r.URL.Path, r.Form)
	    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
	    return
	  }
	  filterName := nickname.Nickname
	  nameCount := 0
	  for _, nick := range nicknames {
	  	if nick.Nickname == nicknameOld {
	  		filterName = nick.Nickname
	  	} else {
	  		nameCount = nameCount + 1
	  	}
	  }
	  if nameCount < 3 {
	  	nickname.UpdatedAt = time.Now()
	  	nickname.UserID = session.UserID
		  update := bson.M{
		    "$set": nickname,
		  }
		  filterNickname := bson.M{"_id": filterName}
		  opts := options.Update().SetUpsert(true)
		  collNickname := common.DB.NicknameDB.Collection("nickname")
		  _, err = collNickname.UpdateOne(ctx, filterNickname, update, opts)
		  if err != nil {
		    log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
		    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		    return
		  }
		  message = message + "ニックネーム情報は更新されました"
	  } else {
	  	message = message + "ニックネーム情報は3件以上は登録できません"
	  }
  }

	update := bson.M{
	    "$set": bson.M{
	        "mail":      user.Mail,
	        "telephone": user.Telephone,
	    },
	}

	filterUser := bson.M{"userID": session.UserID}
	opts := options.Update().SetUpsert(false)
	collSessions := common.DB.SessionDB.Collection("session")
	_, err = collSessions.UpdateMany(ctx, filterUser, update, opts)
	if err != nil {
	    log.Printf("UpdateMany: %v; Req: %s %v", err, r.URL.Path, r.Form)
	    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
	    return
	}

  update = bson.M{
    "$set": user,
  }
  filterUser = bson.M{"_id": session.UserID}
  opts = options.Update().SetUpsert(true)
  collUser := common.DB.UserDB.Collection("user")
  // collUser := db1.Collection("user")
  _, err = collUser.UpdateOne(ctx, filterUser, update, opts)
  if err != nil {
    log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
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

