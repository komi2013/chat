package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chat/common"
	"chat/collection"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.mongodb.org/mongo-driver/bson"
)

// SolanaJpycBalance ユーザーのJPYC残高を取得
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

	// JPYC Mintアドレス
	jpycMint, err := solana.PublicKeyFromBase58(common.JpycSolanaMint)
	if err != nil {
		common.WriteResponseWithSession(w, session, "無効なJPYC Mintアドレス: "+err.Error(), http.StatusOK)
		return
	}

	// トークンアカウントアドレスを取得
	tokenAccount, err := common.FindAssociatedTokenAddress(
		walletPubkey,
		jpycMint,
	)
	if err != nil {
		common.WriteResponseWithSession(w, session, "トークンアカウント取得エラー: "+err.Error(), http.StatusOK)
		return
	}

	// 残高を取得
	balance, err := client.GetTokenAccountBalance(ctx, tokenAccount, rpc.CommitmentConfirmed)
	if err != nil {
		// トークンアカウントが存在しない場合（残高0）もエラーになる可能性がある
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

	// 残高をJPYC単位に変換
	var jpycBalance float64
	if balance.Value != nil {
		decimals := balance.Value.Decimals
		if decimals == 0 {
			decimals = 6 // デフォルト値
		}
		amountStr := balance.Value.Amount
		amount, err := strconv.ParseUint(amountStr, 10, 64)
		if err == nil {
			jpycBalance = float64(amount) / float64(1e6) // decimals=6を想定
		}
	}

	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
		Balance      float64  `json:"balance"`
		WalletAddress string  `json:"walletAddress"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Balance:      jpycBalance,
		WalletAddress: user.SolanaWalletAddress,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
