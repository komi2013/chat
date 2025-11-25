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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AdEdit(w http.ResponseWriter, r *http.Request) {

	csrf := r.FormValue("csrf")

	// === 基本項目 ===
	var ad collection.AdStruct
	ad.AdText = r.FormValue("adText")
	if ad.AdText != "" {
		if len(ad.AdText) > 200 {
			common.WriteResponseWithoutSession(w, csrf, "広告文は200文字以内です", http.StatusOK)
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

	// === 緯度経度 ===
	latStr := r.FormValue("latitude")
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		common.WriteResponseWithoutSession(w, csrf, "緯度が不正です:"+latStr, http.StatusOK)
		return
	}

	lngStr := r.FormValue("longitude")
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		common.WriteResponseWithoutSession(w, csrf, "経度が不正です:"+lngStr, http.StatusOK)
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
		common.WriteResponseWithoutSession(w, csrf, "開始時刻の形式が不正です:"+err.Error(), http.StatusOK)
		return
	}

	reqEnd, err := time.Parse(layout, adEndStr)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, "終了時刻の形式が不正です:"+err.Error(), http.StatusOK)
		return
	}

	if reqEnd.Before(reqStart) {
		common.WriteResponseWithoutSession(w, csrf, "終了時刻は開始より後である必要があります", http.StatusOK)
		return
	}

	ad.AdStart = reqStart
	ad.AdEnd = reqEnd

	// === 距離 ===
	distanceStr := r.FormValue("distance")
	distance := 0
	if distanceStr != "" {
		if d, err := strconv.Atoi(distanceStr); err == nil && d >= 1 && d <= 999 {
			distance = d
		}
	}
	ad.Distance = distance

	// === セッション ===
	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, "セッション失敗:"+err.Error(), http.StatusOK)
		return
	}

	adID := r.FormValue("adID")

	// === 既存広告上限 2件チェック ===
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	adCollection := common.DB.AdDB.Collection("ad")
	var existingAds []collection.AdStruct
	cursor, err := adCollection.Find(ctx, bson.M{"userID": session.UserID})
	if err == nil {
		cursor.All(ctx, &existingAds)
	}

	if len(existingAds) >= 2 {
		if adID == "" {
			common.WriteResponseWithSession(w, session, "現在は2件まで登録可能です", http.StatusOK)
			return
		}
		found := false
		for _, e := range existingAds {
			if e.AdID == adID {
				found = true
				break
			}
		}
		if !found {
			common.WriteResponseWithSession(w, session, "指定された広告はこのユーザーに属していません", http.StatusOK)
			return
		}
	}

	// === 請求済み未払いチェック ===
	if len(existingAds) > 0 {
		for _, e := range existingAds {
			if !e.InvoicedAt.IsZero() {
				if ( e.PaidAt.IsZero() || e.PaidAt.Before(e.InvoicedAt) ) && e.AdID == adID {
					common.WriteResponseWithSession(w, session, "請求済みで未払いのため編集できません", http.StatusOK)
					return
				}
			}
		}
	}

	// === 画像保存 ===
	previewSquare := r.FormValue("previewSquare")
	if previewSquare != "" {
		path, err := common.ImgSave(previewSquare, session.UserID, adID, "-ad-", 0, 5)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Square画像保存失敗:"+err.Error(), http.StatusOK)
			return
		}
		ad.PathSquare = path
	}

	// === 価格取得 ===
	priceList, _ := common.GetPrices(ctx, ad)
	maxPrice := 0
	for _, p := range priceList {
		if p.AdYen > maxPrice {
			maxPrice = p.AdYen
		}
	}
	ad.AdYen = maxPrice

	ad.UserID = session.UserID
	ad.UpdatedAt = time.Now()

	// === adID生成 ===
	if adID == "" {
		newID, err := common.CountUpID("adID")
		if err != nil {
			common.WriteResponseWithSession(w, session, "ID生成失敗:"+err.Error(), http.StatusOK)
			return
		}
		ad.AdID = newID
	} else {
		ad.AdID = adID
	}

	// === Upsert ===
	filter := bson.M{
		"adID":   ad.AdID,
		"userID": ad.UserID,
	}

	update := bson.M{"$set": ad}
	opts := options.Update().SetUpsert(true)

	_, err = adCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		common.WriteResponseWithSession(w, session, "保存失敗:"+err.Error(), http.StatusOK)
		return
	}

	// === 正常レスポンス ===
	resp := struct {
		Csrf         string                `json:"csrf"`
		PushContents []string              `json:"pushContents"`
		Ad           collection.AdStruct   `json:"ad"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Ad:           ad,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
