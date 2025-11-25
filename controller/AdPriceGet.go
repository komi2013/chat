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


	// === 時刻パース (ISO形式 2025-11-23T16:02) ===
	adStartStr := r.FormValue("adStart")
	adEndStr := r.FormValue("adEnd")

	const layout = "2006-01-02T15:04"

	reqStart, err := time.Parse(layout, adStartStr)
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "開始時刻の形式が不正です:"+err.Error(), http.StatusOK)
		return
	}

	reqEnd, err := time.Parse(layout, adEndStr)
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "終了時刻の形式が不正です:"+err.Error(), http.StatusOK)
		return
	}

	if reqEnd.Before(reqStart) {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "終了時刻は開始より後である必要があります", http.StatusOK)
		return
	}

	ad.AdStart = reqStart
	ad.AdEnd = reqEnd

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

