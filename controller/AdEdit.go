package controller

import (
  // "bytes"
  "chat/collection"
  "chat/common"
  "context"
  // "encoding/base64"
  "encoding/json"
  // "errors"
  // "image"
  // "image/jpeg"
  // "image/png"
  // "io"
  "log"
  "net/http"
  "strconv"
  "strings"
  "time"

  _ "image/gif"
  // _ "image/webp"

  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

func AdEdit(w http.ResponseWriter, r *http.Request) {

	var ad collection.AdStruct
	ad.AdText = r.FormValue("adText")
	if ad.AdText != "" {
		if len(ad.AdText) > 200 {
		  log.Print("adText は200文字以内である必要があります")
		  http.Error(w, "adText は200文字以内である必要があります", http.StatusBadRequest)
		  return
		}		
	}

	ad.AdLink = r.FormValue("adLink")
	if ad.AdLink != "" {
	  if len(ad.AdLink) > 200 || !strings.HasPrefix(ad.AdLink, "https://") {
	    log.Print("adLink は200文字以内かつ https:// で始まる必要があります")
	    http.Error(w, "adLink は200文字以内かつ https:// で始まる必要があります", http.StatusBadRequest)
	    return
	  }
	}

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
  ad.Latitude = lat
  ad.Longitude = lng

  adStartStr := r.FormValue("adStart")
  adEndStr := r.FormValue("adEnd")
  if len(adStartStr) != 3 || len(adEndStr) != 3 {
    log.Print("adStart および adEnd は3桁の文字列である必要があります")
    http.Error(w, "adStart および adEnd は3桁の文字列である必要があります", http.StatusBadRequest)
    return
  }
  startDay, err := strconv.Atoi(adStartStr[:1])
  startHour, err2 := strconv.Atoi(adStartStr[1:])
  if err != nil || err2 != nil || startDay < 0 || startDay > 6 || startHour < 0 || startHour > 23 {
    log.Printf("adStart の形式が不正です（曜日:0-6, 時:00-23）: %s", adStartStr)
    http.Error(w, "adStart の形式が不正です（曜日:0-6, 時:00-23）", http.StatusBadRequest)
    return
  }
  endDay, err := strconv.Atoi(adEndStr[:1])
  endHour, err2 := strconv.Atoi(adEndStr[1:])
  if err != nil || err2 != nil || endDay < 0 || endDay > 6 || endHour < 0 || endHour > 23 {
    log.Printf("adEnd の形式が不正です（曜日:0-6, 時:00-23）: %s", adEndStr)
    http.Error(w, "adEnd の形式が不正です（曜日:0-6, 時:00-23）", http.StatusBadRequest)
    return
  }
  ad.AdStart = startDay*100 + startHour // 例: 0*100 + 0 = 000
  ad.AdEnd   = endDay*100 + endHour     // 例: 0*100 + 1 = 001
  distanceStr := r.FormValue("distance")
  distance := 0
  if distanceStr != "" {
    if d, err := strconv.Atoi(distanceStr); err == nil && d >= 1 && d <= 999 {
      distance = d
    }
  }
  ad.Distance = distance

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheck: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

	previewBanner := r.FormValue("previewBanner")
	if previewBanner != "" {
	  path, err := common.SaveBase64Image(previewBanner, session.UserID, "Banner")
	  if err != nil {
	    log.Printf("previewBanner 保存失敗: %v", err)
	    http.Error(w, "previewBanner の保存に失敗しました", http.StatusBadRequest)
	    return
	  }
	  ad.PathBanner = path
	}

	previewSquare := r.FormValue("previewSquare")
	if previewSquare != "" {
	  path, err := common.SaveBase64Image(previewSquare, session.UserID, "Square")
	  if err != nil {
	    log.Printf("previewSquare 保存失敗: %v", err)
	    http.Error(w, "previewSquare の保存に失敗しました", http.StatusBadRequest)
	    return
	  }
	  ad.PathSquare = path
	}

  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()

  // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  // defer cancel()
  // c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  // if err != nil {
  //   log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  // }
  // defer c.Disconnect(ctx)
  // db1 := c.Database(common.MongoDb1)

  priceList, err := common.GetPrices(ctx, ad)
  if err != nil {
    log.Printf("Price lookup failed: %s", err)

    priceList = []collection.AdPriceStruct{
      {AdPriceYen: 200},
    }
  }

  maxPrice := 0
  for _, p := range priceList {
    if p.AdPriceYen > maxPrice {
      maxPrice = p.AdPriceYen
    }
  }
  ad.AdYen = maxPrice

  ad.UserID = session.UserID
  ad.UpdatedAt = time.Now()

  filter := bson.M{
    "userID":  ad.UserID,
  }

  update := bson.M{
    "$set": ad,
  }

  opts := options.Update().SetUpsert(true)

  // adCollection := db1.Collection("ad")
  adCollection := common.DB.AdDB.Collection("ad")
  _, err = adCollection.UpdateOne(ctx, filter, update, opts)
  if err != nil {
    log.Printf("UpdateOne: %v; Req:", err, r.URL.Path, r.Form)
  }

  responseData := struct {
    Csrf          string       `json:"csrf"`
    PushContents  []string     `json:"pushContents"`
    Ad      collection.AdStruct  `json:"ad"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Ad:     ad,
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

