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
  // "log"
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

	csrf := r.FormValue("csrf")
	var ad collection.AdStruct
	ad.AdText = r.FormValue("adText")
	if ad.AdText != "" {
		if len(ad.AdText) > 200 {
		  common.WriteResponseWithoutSession(w, csrf, "広告文200文字以内である必要があります", http.StatusOK)
		  return
		}		
	}

	ad.AdLink = r.FormValue("adLink")
	if ad.AdLink != "" {
	  if len(ad.AdLink) > 200 || !strings.HasPrefix(ad.AdLink, "https://") {
	    common.WriteResponseWithoutSession(w, csrf, "リンクは200文字以内かつ https:// で始まる必要があります", http.StatusOK)
	    return
	  }
	}

	latStr := r.FormValue("latitude")
  lat, err := strconv.ParseFloat(latStr, 64)
  if err != nil || lat < -90 || lat > 90 {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+";緯度が不正:"+latStr, http.StatusOK)
    return
  }
  lngStr := r.FormValue("longitude")
  lng, err := strconv.ParseFloat(lngStr, 64)
  if err != nil || lng < -180 || lng > 180 {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+";経度が不正:"+lngStr, http.StatusOK)
    return
  }
  ad.Latitude = lat
  ad.Longitude = lng
  adStartStr := r.FormValue("adStart")
  adEndStr := r.FormValue("adEnd")
  if len(adStartStr) != 3 || len(adEndStr) != 3 {
    common.WriteResponseWithoutSession(w, csrf, "開始終了は3桁の文字列である必要:"+adStartStr+"-"+adEndStr, http.StatusOK)
    return
  }
  startDay, err := strconv.Atoi(adStartStr[:1])
  startHour, err2 := strconv.Atoi(adStartStr[1:])
  if err != nil || err2 != nil || startDay < 0 || startDay > 6 || startHour < 0 || startHour > 23 {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+err2.Error()+";開始時刻形式が不正（曜日:0-6, 時:00-23）:", http.StatusOK)
    return
  }
  endDay, err := strconv.Atoi(adEndStr[:1])
  endHour, err2 := strconv.Atoi(adEndStr[1:])
  if err != nil || err2 != nil || endDay < 0 || endDay > 6 || endHour < 0 || endHour > 23 {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+err2.Error()+";終了時刻形式が不正（曜日:0-6, 時:00-23）:", http.StatusOK)
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
    common.WriteResponseWithoutSession(w, csrf, err.Error()+";SessionCheckTake", http.StatusOK)
    return
  }

	previewBanner := r.FormValue("previewBanner")
	if previewBanner != "" {
	  path, err := common.SaveBase64Image(previewBanner, session.UserID, "Banner")
	  if err != nil {
	    common.WriteResponseWithSession(w, session, err.Error()+";Banner画像保存失敗", http.StatusOK)
	    return
	  }
	  ad.PathBanner = path
	}

	previewSquare := r.FormValue("previewSquare")
	if previewSquare != "" {
	  path, err := common.SaveBase64Image(previewSquare, session.UserID, "Square")
	  if err != nil {
	    common.WriteResponseWithSession(w, session, err.Error()+";Square画像保存失敗", http.StatusOK)
	    return
	  }
	  ad.PathSquare = path
	}

  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()

  priceList, err := common.GetPrices(ctx, ad)
  if err != nil {
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

  adCollection := common.DB.AdDB.Collection("ad")
  _, err = adCollection.UpdateOne(ctx, filter, update, opts)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";UpdateOne", http.StatusOK)
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

