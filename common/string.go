package common

import (
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"
	"regexp"

)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

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

func Base62Decode(s string) int64 {
	var result int64
	for _, c := range s {
		result = result*62 + int64(strings.IndexRune(charset, c))
	}
	return result
}

func Base62Encode(num int64) string {
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
	num := Base62Decode(s)
	num++
	return Base62Encode(num)
}

func UniqueStrings(input []string) []string {
	uniqueMap := make(map[string]struct{})
	uniqueList := []string{}

	for _, entry := range input {
		if _, exists := uniqueMap[entry]; !exists {
			uniqueMap[entry] = struct{}{} // Add to the map to track seen entries
			uniqueList = append(uniqueList, entry) // Add to the unique list
		}
	}

	return uniqueList
}

func SplitIntoByteChunks(data string, maxBytes int) []string {
	var chunks []string
	var currentChunk string
	currentBytes := 0
	for _, r := range data {
		runeBytes := utf8.RuneLen(r) // Get the byte length of the current rune
		if currentBytes+runeBytes > maxBytes {
			chunks = append(chunks, currentChunk)
			currentChunk = ""
			currentBytes = 0
		}
		currentChunk += string(r)
		currentBytes += runeBytes
	}
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}
	return chunks
}

// スライスに特定の要素が含まれているかチェック
func SliceStrContains(slice []string, target string) bool {
  for _, s := range slice {
    if s == target {
      return true
    }
  }
  return false
}

func GenerateTitle(messageTxt string) string {
	// HTMLタグ除去
	re := regexp.MustCompile(`<[^>]*>`)
	clean := re.ReplaceAllString(messageTxt, "")

	// 改行などの削除
	clean = regexp.MustCompile(`\r?\n`).ReplaceAllString(clean, "")

	// 不要な記号や装飾などの軽い除去（JSの removeMark 相当の一部）
	clean = regexp.MustCompile(`[*_~>`+"`"+`]`).ReplaceAllString(clean, "")

	// 必要に応じてさらに特殊なMarkdown除去処理を追加可

	// 30文字に制限（マルチバイト対応）
	runes := []rune(clean)
	if len(runes) > 30 {
		return string(runes[:30]) + "…"
	}
	return clean
}