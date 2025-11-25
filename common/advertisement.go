package common

import (
	"context"
	"log"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"chat/collection"
)

func GetPrices(ctx context.Context, ad collection.AdStruct) ([]collection.AdPriceStruct, error) {
	delta := 0.01 * float64(ad.Distance)

	coll := DB.AdPriceDB.Collection("adPrice")
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matched []collection.AdPriceStruct

	startTime := ad.AdStart
	endTime := ad.AdEnd

	durationHours := int(endTime.Sub(startTime).Hours())
	if durationHours < 1 {
		durationHours = 1
	}

	for cursor.Next(ctx) {
		var priceDoc collection.AdPriceStruct
		if err := cursor.Decode(&priceDoc); err != nil {
			log.Printf("decode error: %v", err)
			continue
		}

		if priceDoc.LatitudeSouth <= ad.Latitude+delta &&
			priceDoc.LatitudeNorth >= ad.Latitude-delta &&
			priceDoc.LongitudeWest <= ad.Longitude+delta &&
			priceDoc.LongitudeEast >= ad.Longitude-delta &&
			isTimeInRange(ad.AdStart, ad.AdEnd, priceDoc.AdStart, priceDoc.AdEnd) {

			priceDoc.AdYen = int(float64(2*ad.Distance+1) *
				float64(priceDoc.AdPriceYen) *
				float64(durationHours))

			matched = append(matched, priceDoc)
		}
	}

	if len(matched) == 0 {
		adPriceYen := 100
		matched = append(matched, collection.AdPriceStruct{
			AdPriceYen: adPriceYen,
			AdYen:      int(float64(2*ad.Distance+1) * float64(adPriceYen) * float64(durationHours)),
		})
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].AdPriceYen > matched[j].AdPriceYen
	})

	return matched, nil
}

func GetMatchedPrices(ctx context.Context, coll *mongo.Collection, ad collection.AdStruct) ([]collection.AdPriceStruct, error) {
	delta := 0.01 * float64(ad.Distance)

	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matched []collection.AdPriceStruct

	for cursor.Next(ctx) {
		var priceDoc collection.AdPriceStruct
		if err := cursor.Decode(&priceDoc); err != nil {
			log.Printf("decode error: %v", err)
			continue
		}

		if priceDoc.LatitudeSouth <= ad.Latitude+delta &&
			priceDoc.LatitudeNorth >= ad.Latitude-delta &&
			priceDoc.LongitudeWest <= ad.Longitude+delta &&
			priceDoc.LongitudeEast >= ad.Longitude-delta {

			if isTimeInRange(ad.AdStart, ad.AdEnd, priceDoc.AdStart, priceDoc.AdEnd) {
				matched = append(matched, priceDoc)
			}
		}
	}

	if len(matched) == 0 {
		matched = append(matched, collection.AdPriceStruct{
			AdPriceYen: 10,
		})
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].AdPriceYen > matched[j].AdPriceYen
	})

	return matched, nil
}

const weekMinutes = 7 * 24 * 60

func priceIntToMinutes(priceInt int) int {
	day := priceInt / 100
	hour := priceInt % 100
	return ((day - 1) * 24 * 60) + (hour * 60)
}

func isTimeInRange(reqStart, reqEnd time.Time, priceStartInt, priceEndInt int) bool {
	reqStartMin := (int(reqStart.Weekday()) * 24 * 60) + (reqStart.Hour() * 60) + reqStart.Minute()
	reqEndMin := (int(reqEnd.Weekday()) * 24 * 60) + (reqEnd.Hour() * 60) + reqEnd.Minute()

	if reqEndMin <= reqStartMin {
		reqEndMin += weekMinutes
	}

	priceStartMin := priceIntToMinutes(priceStartInt)
	priceEndMin := priceIntToMinutes(priceEndInt)
	if priceEndMin <= priceStartMin {
		priceEndMin += weekMinutes
	}

	for shift := 0; shift <= 1; shift++ {
		ps := priceStartMin + shift*weekMinutes
		pe := priceEndMin + shift*weekMinutes

		if reqStartMin < pe && reqEndMin > ps {
			return true
		}
	}

	return false
}
