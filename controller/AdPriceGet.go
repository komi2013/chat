package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  // "log"
  // "math"
  "net/http"
  "strconv"
  // "strings"
  "time"

)

func AdPriceGet(w http.ResponseWriter, r *http.Request) {

	var ad collection.AdStruct
	latStr := r.FormValue("latitude")
	lngStr := r.FormValue("longitude")
	adStartStr := r.FormValue("adStart")
	adEndStr := r.FormValue("adEnd")
	distanceStr := r.FormValue("distance")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "緯度 (latitude) が不正です: "+latStr, http.StatusOK)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "経度 (longitude) が不正です: "+lngStr, http.StatusOK)
		return
	}

	ad.Latitude = lat
	ad.Longitude = lng

	if len(adStartStr) != 3 || len(adEndStr) != 3 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "adStart および adEnd は3桁の文字列である必要があります", http.StatusOK)
		return
	}

	startDay, err := strconv.Atoi(adStartStr[:1])
	startHour, err2 := strconv.Atoi(adStartStr[1:])
	if err != nil || err2 != nil || startDay < 1 || startDay > 7 || startHour < 0 || startHour > 24 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "adStart の形式が不正です（曜日:1-7, 時:00-24）: "+adStartStr, http.StatusOK)
		return
	}

	endDay, err := strconv.Atoi(adEndStr[:1])
	endHour, err2 := strconv.Atoi(adEndStr[1:])
	if err != nil || err2 != nil || endDay < 1 || endDay > 7 || endHour < 0 || endHour > 24 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "adEnd の形式が不正です（曜日:1-7, 時:00-24）: "+adEndStr, http.StatusOK)
		return
	}

	ad.AdStart = startDay*100 + startHour // 例: 0*100 + 0 = 000
	ad.AdEnd   = endDay*100 + endHour     // 例: 0*100 + 1 = 001

	// 距離チェック
	distance := 0
	if distanceStr != "" {
		if d, err := strconv.Atoi(distanceStr); err == nil && d >= 1 && d <= 999 {
			distance = d
		}
	}
	ad.Distance = distance

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+" session check some error", http.StatusOK)
		return
	}

	priceList, err := common.GetPrices(ctx, ad)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+" session Find", http.StatusOK)
		return
	}

	responseData := struct {
		Csrf          string       `json:"csrf"`
		PushContents  []string     `json:"pushContents"`
		AdPrices      []collection.AdPriceStruct  `json:"adPrices"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		AdPrices:     priceList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)


}

