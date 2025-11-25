package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  // "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
)

// AdTop5Get は Latitude, Longitude, Distance が 0 の広告のうち、
// AdYen の降順で上位5件を返す API
func AdPublicGet(w http.ResponseWriter, r *http.Request) {

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  coll := common.DB.AdDB.Collection("ad")

  // 条件: 緯度経度・距離が0
  filter := bson.M{
    // "latitude":  0,
    // "longitude": 0,
    "distance":  -1,
  }

  // 並び順: AdYen降順 & 上位5件
  findOptions := options.Find().
    SetSort(bson.D{{Key: "adYen", Value: -1}}).
    SetLimit(5)

  cursor, err := coll.Find(ctx, filter, findOptions)
  if err != nil {
  	common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";coll.Find", http.StatusOK)
    return
  }
  defer cursor.Close(ctx)

  var ads []collection.AdResponse
  if err := cursor.All(ctx, &ads); err != nil {
  	common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";cursor.All", http.StatusOK)
    return
  }

  responseData := struct {
    Csrf         string                  `json:"csrf"`
    PushContents []string                `json:"pushContents"`
    Ads          []collection.AdResponse   `json:"ads"`
  }{
    Csrf:         r.FormValue("csrf"),
    PushContents: []string{},
    Ads:          ads,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
