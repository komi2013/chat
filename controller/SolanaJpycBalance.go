package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"chat/common"
	"chat/collection"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.mongodb.org/mongo-driver/bson"
)

// SolanaJpycBalance ユーザーのSOL残高を取得
func SolanaJpycBalance(w http.ResponseWriter, r *http.Request) {
	csrf := r.FormValue("csrf")

	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, err.Error()+"; SessionCheckTake", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collUser := common.DB.UserDB.Collection("user")
	filterUser := bson.M{"_id": session.UserID}
	var user collection.UserStruct
	err = collUser.FindOne(ctx, filterUser).Decode(&user)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+"; FindOne", http.StatusOK)
		return
	}

	// ウォレットが存在しない場合
	if user.SolanaWalletAddress == "" {
		responseData := struct {
			Csrf         string   `json:"csrf"`
			PushContents []string `json:"pushContents"`
			Balance      *float64 `json:"balance,omitempty"`
			Error        string   `json:"error,omitempty"`
		}{
			Csrf:         session.Csrf,
			PushContents: session.PushContents,
			Balance:      nil,
			Error:        "ウォレットが作成されていません",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseData)
		return
	}

	// Solana RPC接続
	client := rpc.New(rpc.MainNetBeta_RPC)

	// ウォレットアドレスをパース
	walletPubkey, err := solana.PublicKeyFromBase58(user.SolanaWalletAddress)
	if err != nil {
		common.WriteResponseWithSession(w, session, "無効なウォレットアドレス: "+err.Error(), http.StatusOK)
		return
	}

	// SOL残高を取得
	balance, err := client.GetBalance(ctx, walletPubkey, rpc.CommitmentConfirmed)
	if err != nil {
		// エラーの場合（残高0）
		var balanceValue *float64
		zeroBalance := 0.0
		balanceValue = &zeroBalance

		responseData := struct {
			Csrf         string   `json:"csrf"`
			PushContents []string `json:"pushContents"`
			Balance      *float64 `json:"balance"`
			WalletAddress string  `json:"walletAddress"`
		}{
			Csrf:         session.Csrf,
			PushContents: session.PushContents,
			Balance:      balanceValue,
			WalletAddress: user.SolanaWalletAddress,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseData)
		return
	}

	// 残高をSOL単位に変換（1 SOL = 1,000,000,000 lamports）
	var solBalance float64
	solBalance = float64(balance) / float64(1_000_000_000)

	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
		Balance      float64  `json:"balance"`
		WalletAddress string  `json:"walletAddress"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Balance:      solBalance,
		WalletAddress: user.SolanaWalletAddress,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
