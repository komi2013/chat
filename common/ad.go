package common

import (
	"bytes"
	"chat/collection"
	"context"
  "encoding/base64"
  "errors"
  "image"
  "image/jpeg"
  "image/png"
	"log"
  "os"
  "path/filepath"
	"sort"
	"strings"
	// "time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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

  filename := "./vue/public/data/ad/" + userID + "_" + suffix + "." + format
  outFile, err := createFileWithDirs(filename) // 自動ディレクトリ作成も
  if err != nil {
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

  return "/data/ad/" + userID + "_" + suffix + "." + format, nil
}

func createFileWithDirs(path string) (*os.File, error) {
  dir := filepath.Dir(path)
  if err := os.MkdirAll(dir, 0755); err != nil {
    return nil, err
  }
  return os.Create(path)
}

