package common

import (
	"bytes"
	"context"
  "encoding/base64"
  "errors"
  "image"
  "image/jpeg"
  "image/png"
	"log"
  // "math"
  "os"
  "path/filepath"
	"sort"
	"strings"
	// "time"

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

	// --- 💰 adYen計算ロジック ---
	startDay := ad.AdStart / 100
	startHour := ad.AdStart % 100
	endDay := ad.AdEnd / 100
	endHour := ad.AdEnd % 100

	durationHours := (endDay*24 + endHour) - (startDay*24 + startHour)
	if durationHours < 0 {
		endDay += 7
		durationHours = (endDay*24 + endHour) - (startDay*24 + startHour)
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
			// priceDoc.AdYen = basePrice + bonus
			matched = append(matched, priceDoc)
		}
	}

	if len(matched) == 0 {
		adPriceYen := 100
		matched = append(matched, collection.AdPriceStruct{
			AdPriceYen: adPriceYen,
			AdYen: int(float64(2*ad.Distance+1) * float64(adPriceYen) * float64(durationHours)),
		})
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].AdPriceYen > matched[j].AdPriceYen
	})
	// log.Printf("matched: %v", ToJSON(matched), ToJSON(ad))
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
		// log.Printf("priceDoc: %v", priceDoc, ad.Latitude, ad.Longitude)
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
			AdPriceYen:     10,
		})
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].AdPriceYen > matched[j].AdPriceYen
	})

	return matched, nil
}

func isTimeInRange(reqStart, reqEnd, priceStart, priceEnd int) bool {
	return !(reqEnd < priceStart || reqStart > priceEnd)
}

func SaveBase64Image(base64Str, userID, suffix string) (string, error) {
  if idx := strings.Index(base64Str, "base64,"); idx != -1 {
    base64Str = base64Str[idx+7:]
  }
  data, err := base64.StdEncoding.DecodeString(base64Str)
  if err != nil {
    return "", err
  }

  img, format, err := image.Decode(bytes.NewReader(data))
  if err != nil {
    return "", err
  }

  filename := OSImgDir + "/ad/" + userID + "_" + suffix + "." + format
  log.Printf("filename error: %v", filename)
  outFile, err := createFileWithDirs(filename) // 自動ディレクトリ作成も
  if err != nil {
  	log.Printf("filename error: %v", err)
    return "", err
  }
  defer outFile.Close()

  switch format {
  case "jpeg":
    err = jpeg.Encode(outFile, img, nil)
  case "png":
    err = png.Encode(outFile, img)
  default:
    return "", errors.New("対応していない画像形式")
  }

  if err != nil {
    return "", err
  }

  return PublicImgPath + "/ad/" + userID + "_" + suffix + "." + format, nil
}

func createFileWithDirs(path string) (*os.File, error) {
  dir := filepath.Dir(path)
  if err := os.MkdirAll(dir, 0755); err != nil {
    return nil, err
  }
  return os.Create(path)
}

