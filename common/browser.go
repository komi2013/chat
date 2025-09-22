package common

import (
  // "math/rand"
	"net/http"
  "strings"
  // "time"
  // "unicode/utf8"

)

// クライアントIPアドレスを取得（X-Forwarded-For対応）
func GetClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		return host[:idx]
	}
	return host
}

func IsMobile(userAgent string) bool {
	mobileKeywords := []string{
		"Mobile", "Android", "iPhone", "iPad", "iPod", "Windows Phone",
	}

	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}
	return false
}

func DetectBrowser(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "Edg/"):
		return "Edge"
	case strings.Contains(userAgent, "Chrome/"):
		return "Chrome"
	case strings.Contains(userAgent, "Firefox/"):
		return "Firefox"
	case strings.Contains(userAgent, "Safari/") && !strings.Contains(userAgent, "Chrome/"):
		return "Safari"
	case strings.Contains(userAgent, "MSIE") || strings.Contains(userAgent, "Trident/"):
		return "Internet Explorer"
	default:
		return "Other"
	}
}
