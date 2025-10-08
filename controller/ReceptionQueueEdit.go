package controller

import (
  "context"
  "encoding/json"
  // "log"
  "net/http"
  "strconv"
  "time"

  "chat/collection"
  "chat/common"

  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo"
)

func ReceptionQueueEdit(w http.ResponseWriter, r *http.Request) {
  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")
  receptionID := r.FormValue("receptionID")
  editType := r.FormValue("editType") // "1"=追加, "2"=削除
	if editType != "1" && editType != "2" {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid editType", http.StatusOK)
		return
	}
  queueName := r.FormValue("queueName")
  guestCount, err := strconv.Atoi(r.FormValue("guestCount"))
	if err != nil && editType == "1" {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid guestCount format", http.StatusOK)
		return
	}

	code := r.FormValue("code")

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    // log.Printf("SessionCheckTake: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  staffAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      staffAccess = true
    }
  }

	if !staffAccess && code == "" {
		common.WriteResponseWithSession(w, session, "コードがない", http.StatusOK)
		return
	}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  var reception collection.ReceptionStruct
  coll := common.DB.ReceptionDB.Collection("reception")
  filter := bson.M{"_id": receptionID}
  err = coll.FindOne(ctx, filter).Decode(&reception)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
    return
  }

	// after you get reception DB, more strict check
	staffAccess = false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID && channelID == reception.ChannelID {
      staffAccess = true
    }
  }

  validCode := false
	now := time.Now().In(time.FixedZone("JST", 9*60*60))

  for j, passcode := range reception.Passcodes {
		// receptionLog.Printf("passcode.Passkey == code:", passcode.Passkey, code)
    if passcode.Passkey == code {
			startTime, err1 := time.Parse("2006-01-02T15:04", passcode.PassStart)
			endTime, err2 := time.Parse("2006-01-02T15:04", passcode.PassEnd)
      if err1 == nil && err2 == nil && now.After(startTime) && now.Before(endTime) {
        validCode = true
        break
      } else if now.After(endTime) { //remove expired code
        reception.Passcodes = append(
            reception.Passcodes[:j],
            reception.Passcodes[j+1:]...,
        )
        update := bson.M{
            "$set": bson.M{
                "passcodes": reception.Passcodes,
            },
        }
        _, err := coll.UpdateOne(ctx, filter, update)
        if err != nil {
            common.WriteResponseWithSession(w, session, "Update failed: "+err.Error(), http.StatusInternalServerError)
            return
        }
        break
      }
    }
  }


  if !staffAccess && !validCode{
  	common.WriteResponseWithSession(w, session, "コードが一致してません。スタッフでもありません", http.StatusOK)
  	return
  }

	queuedAt := now.Format("2006-01-02T15:04")

	switch editType {
	case "1": // 追加
	  newQueue := collection.Queue{
	    WaitingGuest: guestCount,
	    QueueName:    session.Nickname,
	    QueuedAt:     queuedAt,
	    UserID:       session.UserID,
	  }

	  update := bson.M{
	    "$push": bson.M{
	      "queues": newQueue,
	    },
	  }

	  _, err = coll.UpdateOne(ctx, filter, update)
	  if err != nil {
	    common.WriteResponseWithSession(w, session, "Update failed: "+err.Error(), http.StatusInternalServerError)
	    return
	  }

	case "2": // 削除
	  if !staffAccess {
	    common.WriteResponseWithSession(w, session, "スタッフのみ削除できます", http.StatusForbidden)
	    return
	  }

	  // QueueName（もしくはUserID）に一致するものを削除
	  update := bson.M{
	    "$pull": bson.M{
	      "queues": bson.M{
	        "queueName": queueName, // Queue構造体のフィールド名に合わせて要確認
	      },
	    },
	  }

	  _, err = coll.UpdateOne(ctx, filter, update)
	  if err != nil {
	    common.WriteResponseWithSession(w, session, "Update failed: "+err.Error(), http.StatusInternalServerError)
	    return
	  }
	}

  responseData := common.BaseResponse{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
