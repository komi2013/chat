package common

import (
	"fmt"
  "net/http"
  "runtime"
)

func ResponseErrorStatus(w http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}

func LogError(message string, err error) {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	fmt.Printf("[ERROR] %s: %s - %v\n", funcName, message, err)
}