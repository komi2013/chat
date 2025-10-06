package controller

import (
	"encoding/json"
	"log"
	"net/http"
	// "os"
	// "strings"
	"time"

	"chat/common"
)

// フロントから送られてくるエラーログの構造体
type ErrorLogPayload struct {
	Message   string      `json:"message,omitempty"`
	Stack     string      `json:"stack,omitempty"`
	Level     string      `json:"level,omitempty"`
	URL       string      `json:"url"`
	Timestamp int64       `json:"timestamp"` // JS: Date.now() (ms)
	Extra     interface{} `json:"extra,omitempty"`
	Csrf string       `json:"csrf"`
}

// LoggerHandler: /logger/ にPOSTされたエラーログを受け取り、client_error.log に出力する
func LogFromJS(w http.ResponseWriter, r *http.Request) {
	// common.ClientErrorLogger.Printf("aaa")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload ErrorLogPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// CSRFが無効ならログを書かずにstatus:okを返す
	if !isValidCSRF(payload.Csrf) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "invalid POST"})
		return
	}

	tm := time.Unix(payload.Timestamp/1000, 0).Format("2006-01-02 15:04:05")
	clientIP := common.GetClientIP(r)
	userAgent := r.UserAgent()
	isMobile := common.IsMobile(userAgent)
	browser := common.DetectBrowser(userAgent)

	var jsErrorLogger = common.NewDailyLogger("js_error_")

	jsErrorLogger.Printf(
		"Time=%s IP=%s Mobile=%t Browser=%s URL=%s Stack=%s Extra=%v",
		tm,
		clientIP,
		isMobile,
		browser,
		payload.URL,
		payload.Stack,
		payload.Extra,
	)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// csrfの有効性チェック（失敗したらfalse）
func isValidCSRF(csrf string) bool {
	if len(csrf) <= 16 {
		return false
	}

	// 末尾がBase62エンコードされたtimestamp
	timePart := csrf[16:]
	decodedTs := common.Base62Decode(timePart)

	now := time.Now().Unix()
	// 24時間以内ならOK
	log.Print("now-decodedTs", now, decodedTs, 24*60*60)
	return now-decodedTs <= 24*60*60
}

