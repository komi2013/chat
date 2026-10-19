package controller

import (
	"encoding/json"
	"net/http"

	"chat/collection"
	"chat/common"
)

func UserGet(w http.ResponseWriter, r *http.Request) {
	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(
			w,
			r.FormValue("csrf"),
			err.Error()+";Session Check",
			http.StatusOK,
		)
		return
	}

	var user collection.UserResponse

	err = common.DB.UserDB.QueryRowContext(
		r.Context(),
		`
		SELECT
			COALESCE(mail, ''),
			COALESCE(telephone, ''),
			COALESCE(wallet_address, ''),
			COALESCE(latitude, 0),
			COALESCE(longitude, 0)
		FROM "user"
		WHERE user_id = $1
		`,
		session.UserID,
	).Scan(
		&user.Mail,
		&user.Telephone,
		&user.WalletAddress,
		&user.Latitude,
		&user.Longitude,
	)
	if err != nil {
		common.WriteResponseWithSession(
			w,
			session,
			err.Error()+";UserGet",
			http.StatusOK,
		)
		return
	}

	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
		SessionID    string   `json:"sessionId,omitempty"`
		User         userResponse `json:"user"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		User:         user,
	}

	if session.SessionIDRotated {
		responseData.SessionID = session.SessionID
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}

type userResponse struct {
	Mail          string  `json:"mail,omitempty"`
	Telephone     string  `json:"telephone,omitempty"`
	WalletAddress string  `json:"walletAddress"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
}