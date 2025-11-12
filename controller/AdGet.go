package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  "fmt"
  // "log"
  // "math"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/bson"
)

func AdGet(w http.ResponseWriter, r *http.Request) {

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
    return
  }

  coll := common.DB.AdDB.Collection("ad")
  filter := bson.M{
    "userID": session.UserID,
  }
	var ads []collection.AdStruct
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";coll.Find", http.StatusOK)
    return
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &ads); err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";cursor.All", http.StatusOK)
    return
	}

  url := fmt.Sprintf(
    "%s/api?module=account&action=tokentx&address=%s&contractaddress=%s&sort=desc",
    common.PolygonAPI,
    common.SystemWalletAddress,
    common.JpycContract,
  )

  // resp, err := http.Get(url)
  // if err != nil {
  //   common.WriteResponseWitSession(w, csrf, err.Error()+";http.Get", http.StatusOK)
  //   return
  // }

  responseData := struct {
    Csrf         string       `json:"csrf"`
    PushContents []string     `json:"pushContents"`
    Ads       []collection.AdStruct  `json:"ads"`
    SystemWalletAddress   string       `json:"systemWalletAddress"`
    JpycCheckURL   string       `json:"jpycCheckURL"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Ads:          ads,
    SystemWalletAddress: common.SystemWalletAddress,
    JpycCheckURL: url,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

