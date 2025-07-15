package common

import (
  "context"
  "fmt"
  "strings"

  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
)

func IncrementBase62Smart(s string) string {
	runes := []rune(s)
	n := len(runes)

	// Base62文字からインデックスを得るマップ
	charIndex := make(map[rune]int)
	for i, c := range charset {
		charIndex[c] = i
	}

	// カウントアップ処理（末尾から）
	carry := true
	for i := n - 1; i >= 0 && carry; i-- {
		index := charIndex[runes[i]]
		if index < 61 {
			runes[i] = rune(charset[index+1])
			carry = false
		} else {
			runes[i] = rune(charset[0]) // 'z' -> '0'
		}
	}

	// すべて繰り上がった場合（例: "z" → "00", "zz" → "000"）
	if carry {
		return strings.Repeat("0", n+1)
	}

	return string(runes)
}

func CountUpID(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(Mongo1))
	if err != nil {
		return "", err
	}
	defer client.Disconnect(ctx)

	db := client.Database(MongoDb1)
	coll := db.Collection("sequence")

	var seq collection.SequenceStruct
	var locked bool

	// 1〜3回ロック試行
	for i := 0; i < 3; i++ {
		res, err := coll.UpdateOne(ctx,
			bson.M{"_id": key, "lock": 0},
			bson.M{
				"$set": bson.M{
					"lock":      1,
					"updatedAt": time.Now(),
				},
			},
		)
		if err != nil {
			return "", err
		}
		if res.ModifiedCount > 0 {
			locked = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !locked {
		return "", fmt.Errorf("could not acquire lock for key %s after 3 attempts", key)
	}

	// 値を取得
	err = coll.FindOne(ctx, bson.M{"_id": key}).Decode(&seq)
	if err != nil {
		return "", err
	}

	// カウントアップ
	newCount := IncrementBase62Smart(seq.Count)

	// カウントを保存
	_, err = coll.UpdateOne(ctx,
		bson.M{"_id": key},
		bson.M{
			"$set": bson.M{
				"count":     newCount,
				"updatedAt": time.Now(),
				"lock":      0,
			},
		})
	if err != nil {
		return "", err
	}

	return newCount, nil
}
