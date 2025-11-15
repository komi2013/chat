package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  "net/http"
  "strconv"
  "strings"
  "time"

  _ "image/gif"

  "go.mongodb.org/mongo-driver/bson"
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
  if err != nil || err2 != nil || startDay < 1 || startDay > 7 || startHour < 0 || startHour > 24 {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+err2.Error()+";開始時刻形式が不正（曜日:1-7, 時:00-24）:", http.StatusOK)
    return
  }
  endDay, err := strconv.Atoi(adEndStr[:1])
  endHour, err2 := strconv.Atoi(adEndStr[1:])
  if err != nil || err2 != nil || endDay < 1 || endDay > 7 || endHour < 0 || endHour > 24 {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+err2.Error()+";終了時刻形式が不正（曜日:1-7, 時:00-24）:", http.StatusOK)
    return
  }
  ad.AdStart = startDay*100 + startHour
  ad.AdEnd   = endDay*100 + endHour
  distanceStr := r.FormValue("distance")
  distance := 0
  if distanceStr != "" {
    if d, err := strconv.Atoi(distanceStr); err == nil && d >= 1 && d <= 999 {
      distance = d
    }
  }
  ad.Distance = distance

  // === セッション確認 ===
  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    common.WriteResponseWithoutSession(w, csrf, err.Error()+";SessionCheckTake", http.StatusOK)
    return
  }

  adID := r.FormValue("adID")

  // === 🔍 ユーザーの既存広告チェック ===
  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()

  adCollection := common.DB.AdDB.Collection("ad")
  var existingAds []collection.AdStruct
  cursor, err := adCollection.Find(ctx, bson.M{"userID": session.UserID})
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";Find error", http.StatusOK)
    return
  }
  if err := cursor.All(ctx, &existingAds); err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";cursor decode error", http.StatusOK)
    return
  }

  // === 条件チェック ===
  if len(existingAds) >= 2 {
    existing := existingAds[0]
    if adID == "" {
      // 新規登録だけど既に1件存在
      common.WriteResponseWithSession(w, session, "現在は1ユーザーにつき2件のみ登録可能です", http.StatusOK)
      return
    }
    if adID != existing.AdID {
      // 既存のadIDと一致しない（他人や別データ）
      common.WriteResponseWithSession(w, session, "指定された広告はこのユーザーに属していません", http.StatusOK)
      return
    }
  }

	previewSquare := r.FormValue("previewSquare")
	if previewSquare != "" {
	  path, err := common.ImgSave(previewSquare, session.UserID, adID, "-ad-", 0, 5)
	  if err != nil {
	    common.WriteResponseWithSession(w, session, err.Error()+";Square画像保存失敗", http.StatusOK)
	    return
	  }
	  ad.PathSquare = path
	}

  priceList, err := common.GetPrices(ctx, ad)
  if err != nil {
    // priceList = []collection.AdPriceStruct{
    //   {AdPriceYen: 200},
    // }
  }

  maxPrice := 0
  for _, p := range priceList {
    if p.AdYen > maxPrice {
      maxPrice = p.AdYen
    }
  }
  ad.AdYen = maxPrice

  ad.UserID = session.UserID
  ad.UpdatedAt = time.Now()

  // === ID割り当て ===
	if adID == "" {
		newID, err := common.CountUpID("adID")
		if err != nil {
			common.WriteResponseWithSession(w, session, err.Error()+" CountUpID error", http.StatusOK)
			return
		}
		ad.AdID = newID
	} else {
		ad.AdID = adID
	}

	// === Upsert ===
	filter := bson.M{
		"adID": ad.AdID,
		"userID": ad.UserID,
	}

  update := bson.M{
    "$set": ad,
  }

  opts := options.Update().SetUpsert(true)
  _, err = adCollection.UpdateOne(ctx, filter, update, opts)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";UpdateOne", http.StatusOK)
    return
  }

  // === 正常レスポンス ===
  responseData := struct {
    Csrf          string       `json:"csrf"`
    PushContents  []string     `json:"pushContents"`
    Ad            collection.AdStruct  `json:"ad"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Ad:           ad,
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
