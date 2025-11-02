package controller

import (
  "context"
  "encoding/json"
  // "log"
  "net/http"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"
)

func ContentsPush(w http.ResponseWriter, r *http.Request) {
  csrf := r.FormValue("csrf")

	var pushNames []string
  if err := json.Unmarshal([]byte(r.FormValue("pushNames")), &pushNames); err != nil {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+";Invalid JSON pushNames", http.StatusOK)
    return
  }

  updatedBy := r.FormValue("updatedBy")
  channelID := r.FormValue("channelID")
  pushTitle := r.FormValue("pushTitle")

  var contents interface{}
  if err := json.Unmarshal([]byte(r.FormValue("contents")), &contents); err != nil {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+";Invalid JSON contents", http.StatusOK)
    return
  }

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 最大10MB
		common.WriteResponseWithoutSession(w, csrf, err.Error()+";files more than 10MB", http.StatusOK)
		return
	}

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
  	common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
    return
	}

  trueAccess := false
  for _, d := range session.ChannelAliases { // to prevent bad request with different name
    if d.Alias == updatedBy && d.ChannelID == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    common.WriteResponseWithSession(w, session, "no true access right", http.StatusOK)
    return
  }
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  var channel collection.ChannelStruct
	collChannel := common.DB.ChannelDB.Collection("channel")
	filterChannel := bson.M{"_id": channelID}
	err = collChannel.FindOne(ctx, filterChannel).Decode(&channel)
	if err != nil {
	  common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}
	uniqueIDs := make(map[string]struct{})
	var userIDs []string
	pushNameSet := make(map[string]struct{}, len(pushNames))
	for _, name := range pushNames {
		pushNameSet[name] = struct{}{}
	}
	for _, alias := range channel.Aliases {
		if _, ok := pushNameSet[alias.AliasName]; ok {
			if _, exists := uniqueIDs[alias.UserID]; !exists {
				uniqueIDs[alias.UserID] = struct{}{}
				userIDs = append(userIDs, alias.UserID)
			}
		}
	}
	var contentsPush = common.NewDailyLogger("contents_push_channel_id")
	contentsPush.Printf(channelID)

  imgPath, err := common.ImgSave(r.FormValue("imgPath"), session.UserID, updatedBy, channelID, 0, 1)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+";imgPath", http.StatusOK)
		return
	}

	fileLinks, err := common.FileSave(r, channelID, updatedBy, userIDs, 2)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+";FileSave", http.StatusOK)
		return
	}

  coll := common.DB.SessionDB.Collection("session")
  filter := bson.D{{"userID", bson.D{{"$in", userIDs}}}}
  cursor, err := coll.Find(context.TODO(), filter)
  if err != nil {
  	common.WriteResponseWithSession(w, session, err.Error()+";coll.Find", http.StatusOK)
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
  	common.WriteResponseWithSession(w, session, err.Error()+";cursor.All", http.StatusOK)
  }
  filteredSessions := common.FilterSessionsByChannelID(sessions, channelID)
  var arr []interface{}
  arr = append(arr, pushTitle)
  arr = append(arr, channelID)
  arr = append(arr, updatedBy)
  arr = append(arr, contents)
  if imgPath != "" {
  	arr = append(arr, imgPath)
  } else {
  	arr = append(arr, fileLinks)
  }
  
	common.ChunkPush(filteredSessions, arr)

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

