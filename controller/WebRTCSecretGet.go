package controller

import (
  // "chat/collection"
  "chat/common"
  // "context"
  "encoding/json"
  // "fmt"
  // "log"
  // "math"
  "net/http"
  // "time"

  // "go.mongodb.org/mongo-driver/bson"
)

func WebRTCTokenGet(w http.ResponseWriter, r *http.Request) {

  // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  // defer cancel()

  // session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  // if err != nil {
  //   common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
  //   return
  // }

  token, err := common.GenerateSkyWayToken()
  if err != nil {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";GenerateSkyWayToken", http.StatusOK)
    // common.WriteResponseWithSession(w, session, err.Error()+";GenerateSkyWayToken", http.StatusOK)
    return
  }

  responseData := struct {
    Csrf         string       `json:"csrf"`
    PushContents []string     `json:"pushContents"`
    Token   string       `json:"token"`
  }{
    // Csrf:         session.Csrf,
    Csrf:         r.FormValue("csrf"),
    // PushContents: session.PushContents,
    PushContents: []string{},
    Token: token,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

