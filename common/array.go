package common

// import (
//   "math/rand"
//   "strings"
//   "time"
//   "unicode/utf8"

// )
// スライスに特定の要素が含まれているかチェック
func SliceStrContains(slice []string, target string) bool {
  for _, s := range slice {
    if s == target {
      return true
    }
  }
  return false
}