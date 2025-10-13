package controller

import (
  "context"
  "encoding/json"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/common"
  "chat/collection"
)

func ChannelEdit(w http.ResponseWriter, r *http.Request) {
  var aliases []collection.Alias
  if err := json.Unmarshal([]byte(r.FormValue("aliases")), &aliases); err != nil {
    log.Printf("aliases: %v; Request:", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON aliases", http.StatusBadRequest)
    return
  }

  var aliasNames []string
  if err := json.Unmarshal([]byte(r.FormValue("aliasNames")), &aliasNames); err != nil {
    log.Printf("aliasNames: %v; Request:", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid JSON aliasNames", http.StatusBadRequest)
    return
  }

  untilDate, err := time.Parse("2006-01-02", r.FormValue("untilDate"))
  if err != nil {
    log.Printf("Invalid untilDate: %v; Request:", err, r.URL.Path, r.Form)
    http.Error(w, "Invalid untilDate: must be in YYYY-MM-DD format", http.StatusBadRequest)
    return
  }

  channelID := r.FormValue("channelID")
  channelName := r.FormValue("channelName")
  channelDescription := r.FormValue("channelDescription")
  updatedBy := r.FormValue("updatedBy")
  guest := r.FormValue("guest") != ""

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheck: %v; Request:", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  trueAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == updatedBy && d.ChannelID == channelID && !d.Guest {
      trueAccess = true
    }
  }

  if !trueAccess {
    log.Printf("ChannelAliases !trueAccess: %v; Request:", session.ChannelAliases, updatedBy, channelID, r.URL.Path, r.Form)
    common.WriteResponseWithSession(w, session, "ChannelAliases !trueAccess", http.StatusOK)
    return
  }

  // ===== Channel の更新処理 =====
  coll := common.DB.ChannelDB.Collection("channel")
  filter := bson.M{"_id": channelID}
  invitationCode := common.StringRand(16)
  update := bson.M{
    "$set": bson.M{
      "channelName":        channelName,
      "channelDescription": channelDescription,
      "aliases":            aliases,
      "aliasNames":         aliasNames,
      "guest":              guest,
      "untilDate":          untilDate,
      "updatedBy":          updatedBy,
      "updatedAt":          time.Now(),
      "invitationCode":     invitationCode,
    },
  }

  opts := options.Update().SetUpsert(false)
  result, err := coll.UpdateOne(context.TODO(), filter, update, opts)
  if err != nil {
    log.Printf("UpdateOne: %v; Request:", err, r.URL.Path, r.Form)
    http.Error(w, "Failed to update channel", http.StatusInternalServerError)
    return
  }

  if result.MatchedCount == 0 {
    log.Printf("No channel found for ID: %s", channelID)
    http.Error(w, "Channel not found", http.StatusNotFound)
    return
  }

  responseData := struct {
    InvitationCode    string        `json:"invitationCode"`
    Csrf         string   `json:"csrf"`
    PushContents []string `json:"pushContents"`
  }{
    InvitationCode:    invitationCode,
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
