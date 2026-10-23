package common

import (
  "encoding/json"
	// "log"
  "net/http"
  // "strings"

  "chat/table"
)

type BaseResponse struct {
	Csrf         string   `json:"csrf"`
	PushContents []string `json:"pushContents"`
	Error        string   `json:"error,omitempty"`
}

type ReceptionResponse struct {
	Csrf         string   `json:"csrf"`
	PushContents []string `json:"pushContents"`
	Error        string   `json:"error,omitempty"`
  Mail         string   `json:"mail,omitempty"`
  Telephone    string   `json:"telephone,omitempty"`
	Reception 	 table.ReceptionStruct `json:"reception,omitempty"`
	Menu 	 			 table.Menu `json:"menu,omitempty"`
	FacilityName string    `json:"facilityName,omitempty"`
	Nickname     string    `json:"nickname,omitempty"`
}


func ResponseErrorStatus(w http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}

func WriteResponseWithSession(w http.ResponseWriter, session table.SessionStruct, errMsg string, status int) {
    w.Header().Set("Content-Type", "application/json")
    if status != http.StatusOK {
      w.WriteHeader(status)
    }

    responseData := struct {
      Csrf         string   `json:"csrf"`
      PushContents []string `json:"pushContents"`
      SessionID    string   `json:"sessionId,omitempty"`
      Error        string   `json:"error,omitempty"`
    }{
      Csrf:         session.Csrf,
      PushContents: session.PushContents,
      SessionID:    rotatedSessionID(session),
      Error:        errMsg,
    }

    json.NewEncoder(w).Encode(responseData)
}

func WriteResponseWithoutSession(w http.ResponseWriter, csrf string, errMsg string, status int) {
    w.Header().Set("Content-Type", "application/json")
    if status != http.StatusOK {
      w.WriteHeader(status)
    }
    responseData := struct {
      Csrf         string   `json:"csrf"`
      PushContents []string `json:"pushContents"`
      Error        string   `json:"error,omitempty"`
    }{
      Csrf:         csrf,
      PushContents: []string{},
      Error:        errMsg,
    }

    json.NewEncoder(w).Encode(responseData)
}

func rotatedSessionID(session table.SessionStruct) string {
  if session.SessionIDRotated {
    return session.SessionID
  }
  return ""
}