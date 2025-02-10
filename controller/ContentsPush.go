package controller

import (
  "context"
  "encoding/json"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"
)

func ContentsPush(w http.ResponseWriter, r *http.Request) {
	var userIDs []string
  if err := json.Unmarshal([]byte(r.FormValue("userIDs")), &userIDs); err != nil {
  	log.Printf("userIDs: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON userIDs", http.StatusBadRequest)
    return
  }

  updatedBy := r.FormValue("updatedBy")
  channelID := r.FormValue("channelID")
  pushTitle := r.FormValue("pushTitle")

  var contents interface{}
  if err := json.Unmarshal([]byte(r.FormValue("contents")), &contents); err != nil {
  	log.Printf("contents: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON contents", http.StatusBadRequest)
    return
  }

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 最大10MB
		log.Printf("ParseMultipartForm: %v; Req: ", err, r.URL.Path, r.Form)
		http.Error(w, "Failed to parse form because more than 10MB", http.StatusBadRequest)
		return
	}

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
    if d.Alias == updatedBy && d.ChannelID == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    log.Printf("ChannelAliases !trueAccess: %v; Req: ", session.ChannelAliases, updatedBy, channelID, r.URL.Path, r.Form)
    return
  }
  imgPath, err := common.ImgSave(db1, r.FormValue("imgPath"), session.UserID, updatedBy, channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileLinks, err := common.FileSave(r, db1, channelID, updatedBy, userIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  coll := db1.Collection("session")
  filter := bson.D{{"user_id", bson.D{{"$in", userIDs}}}}
  cursor, err := coll.Find(context.TODO(), filter)
  if err != nil {
    log.Printf("coll.Find: %v; Req: ", err, userIDs, r.URL.Path, r.Form)
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
    log.Printf("cursor.All: %v; Req: ", err, userIDs, r.URL.Path, r.Form)
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
  
	common.ChunkPush(filteredSessions, db1, arr)

	session, err = common.ReGenerateData(db1, session)
	if err != nil {
		log.Printf("ReGenerateData: %v; Req:", err, r.URL.Path, r.Form)
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

