package controller

import (
	"chat/collection"
	"chat/common"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

func AdPublicGet(w http.ResponseWriter, r *http.Request) {
	session, _ := common.SessionGet(w, r)
	session, err := common.PushReGenerate(session)
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";ReGenerateCSRF", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := common.DB.AdDB.Collection("ad")
	filter := bson.M{
		"distance": -1,
	}
	cursor, err := coll.Find(ctx, filter)
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
	var jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	nowJST := time.Now().In(jst)
	activeAds := make([]collection.AdResponse, 0)
	for _, ad := range ads {
		adStartJST := ad.AdStart.In(jst)
		adEndJST := ad.AdEnd.In(jst)
		paidAtJST := ad.PaidAt
		if !paidAtJST.IsZero() {
			paidAtJST = paidAtJST.In(jst)
		}
		if adEndJST.Before(nowJST) {
			_, delErr := coll.DeleteOne(ctx, bson.M{"_id": ad.AdID})
			if delErr != nil {
				log.Println("DELETE ERROR:", delErr)
			}
			continue
		}
		isCurrentlyActive := nowJST.After(adStartJST) && nowJST.Before(adEndJST.Add(1*time.Second))
		isPaidBeforeEnd := false
		if !ad.PaidAt.IsZero() {
			if paidAtJST.Before(adEndJST.Add(1*time.Second)) {
				isPaidBeforeEnd = true
			}
		}
		if isCurrentlyActive && isPaidBeforeEnd {
			ad.AdStart = adStartJST
			ad.AdEnd = adEndJST
			ad.PaidAt = paidAtJST
			activeAds = append(activeAds, ad)
		}
	}
	sort.Slice(activeAds, func(i, j int) bool {
		return activeAds[i].AdYen > activeAds[j].AdYen
	})

	responseData := struct {
		Csrf         string                `json:"csrf"`
		PushContents []string              `json:"pushContents"`
		Ads          []collection.AdResponse `json:"ads"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Ads:          activeAds,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}