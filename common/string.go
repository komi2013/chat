package common

import (
	"math/rand"
	"strings"
	"time"

)

const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// StringRand random step 2
func StringRand(length int) string {
	return stringWithCharset(length, charset)
}

// StringReverse random step 3
// func StringReverse(s string) string {
// 	rs := []rune(s)
// 	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
// 		rs[i], rs[j] = rs[j], rs[i]
// 	}
// 	return string(rs)
// }

// func SliceUnique(target []int) (unique []int) {
// 	m := map[int]bool{}
// 	for _, v := range target {
// 		if !m[v] {
// 			m[v] = true
// 			unique = append(unique, v)
// 		}
// 	}
// 	return unique
// }

func base62Decode(s string) int64 {
	var result int64
	for _, c := range s {
		result = result*62 + int64(strings.IndexRune(charset, c))
	}
	return result
}

func base62Encode(num int64) string {
	var result strings.Builder
	for num > 0 {
		remainder := num % 62
		result.WriteByte(charset[remainder])
		num /= 62
	}
	return reverseString(result.String())
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func incrementBase62(s string) string {
	num := base62Decode(s)
	num++
	return base62Encode(num)
}
