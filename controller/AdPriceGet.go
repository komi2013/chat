package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  "log"
  // "math"
  "net/http"
  "strconv"
  // "strings"
  "time"

  // "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

func AdPriceGet(w http.ResponseWriter, r *http.Request) {

	var ad collection.AdStruct
	// ad.PathBanner = r.FormValue("pathBanner")
	// ad.PathSquare = r.FormValue("pathSquare")
	latStr := r.FormValue("latitude")
	lngStr := r.FormValue("longitude")
	adStartStr := r.FormValue("adStart")
	adEndStr := r.FormValue("adEnd")
	distanceStr := r.FormValue("distance")

	// if ad.PathBanner != "" && !strings.HasPrefix(ad.PathBanner, "http://") && !strings.HasPrefix(ad.PathBanner, "https://") {
	// 	log.Print("pathBanner は http:// または https:// で始まる必要があります")
	// 	http.Error(w, "pathBanner は http:// または https:// で始まる必要があります", http.StatusBadRequest)
	// 	return
	// }

	// if ad.PathSquare != "" && !strings.HasPrefix(ad.PathSquare, "http://") && !strings.HasPrefix(ad.PathSquare, "https://") {
	// 	log.Print("pathSquare は http:// または https:// で始まる必要があります")
	// 	http.Error(w, "pathSquare は http:// または https:// で始まる必要があります", http.StatusBadRequest)
	// 	return
	// }

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		log.Printf("緯度 (latitude) が不正です: %v", latStr)
		http.Error(w, "緯度 (latitude) が不正です", http.StatusBadRequest)
		return
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		log.Printf("経度 (longitude) が不正です: %v", lngStr)
		http.Error(w, "経度 (longitude) が不正です", http.StatusBadRequest)
		return
	}
	ad.Latitude = lat
	ad.Longitude = lng

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

	distance := 1
	if distanceStr != "" {
		if d, err := strconv.Atoi(distanceStr); err == nil && d >= 1 && d <= 999 {
			distance = d
		}
	}
	ad.Distance = distance

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
  // var ad collection.AdStruct

  coll := db1.Collection("adPrice")

	priceList, err := common.GetMatchedPrices(ctx, coll, ad)
	if err != nil {
		log.Printf("Price lookup failed: %s", err)
	}

	var priceInfos []collection.AdPriceStruct
	for _, p := range priceList {
		priceInfos = append(priceInfos, collection.AdPriceStruct{
			AdPriceYen:           p.AdPriceYen,
			AdStart:       p.AdStart,
			AdEnd:         p.AdEnd,
			LatitudeNorth: p.LatitudeNorth,
			LatitudeSouth: p.LatitudeSouth,
			LongitudeEast: p.LongitudeEast,
			LongitudeWest: p.LongitudeWest,
		})
	}

	responseData := struct {
		Csrf          string       `json:"csrf"`
		PushContents  []string     `json:"pushContents"`
		AdPrices      []collection.AdPriceStruct  `json:"adPrices"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		AdPrices:     priceInfos,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)


}

