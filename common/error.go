package common

import (
  "encoding/json"
	"log"
  "net/http"
  "runtime"
  "strings"

  "chat/collection"
)

func ResponseErrorStatus(w http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}

func LogError(message string, err error, errs ...interface{}) {
	funcName := getControllerFuncName()
	log.Printf("[LogError] %s: %s - %v %v", funcName, message, err, errs)
}

func getControllerFuncName() string {
	var lastValidFunc string // 直前の関数を保存する
	for i := 2; i < 15; i++ { // 2 〜 15 の範囲で関数名を探る
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break // これ以上スタックがない場合はループ終了
		}
		funcName := runtime.FuncForPC(pc).Name()

		// net/http の関数なら、直前の関数を返す
		if strings.Contains(funcName, "net/http.") {
			return lastValidFunc
		}

		// runtime や net/http でない関数を lastValidFunc に保存
		if !strings.Contains(funcName, "runtime.") {
			lastValidFunc = funcName
		}
	}
	return "unknown"
}

func WriteResponseWithSession(w http.ResponseWriter, session collection.SessionStruct, errMsg string, status int) {
    w.Header().Set("Content-Type", "application/json")
    if status != http.StatusOK {
        w.WriteHeader(status)
    }

    responseData := struct {
        Csrf         string   `json:"csrf"`
        PushContents []string `json:"pushContents"`
        Error        string   `json:"error,omitempty"`
    }{
        Csrf:         session.Csrf,
        PushContents: session.PushContents,
        Error:        errMsg,
    }

    json.NewEncoder(w).Encode(responseData)
}
