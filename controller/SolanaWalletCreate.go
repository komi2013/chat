package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"chat/common"
	"chat/collection"
	"github.com/gagliardetto/solana-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SolanaWalletCreate ユーザーのSolanaウォレットを生成・保存
func SolanaWalletCreate(w http.ResponseWriter, r *http.Request) {
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

	// 既にウォレットが存在する場合はエラー
	if user.SolanaPrivateKey != "" && user.SolanaWalletAddress != "" {
		responseData := struct {
			Csrf              string `json:"csrf"`
			PushContents      []string `json:"pushContents"`
			Error             string `json:"error,omitempty"`
			SolanaWalletAddress string `json:"solanaWalletAddress,omitempty"`
		}{
			Csrf:              session.Csrf,
			PushContents:      session.PushContents,
			Error:             "ウォレットは既に作成されています",
			SolanaWalletAddress: user.SolanaWalletAddress,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseData)
		return
	}

	// 新しいSolanaウォレットを生成
	keypair, err := solana.NewRandomPrivateKey()
	if err != nil {
		common.WriteResponseWithSession(w, session, "ウォレット生成に失敗しました: "+err.Error(), http.StatusOK)
		return
	}

	privateKeyBase58 := keypair.String()
	publicKey := keypair.PublicKey()
	walletAddress := publicKey.String()

	// ユーザー情報を更新
	update := bson.M{
		"$set": bson.M{
			"solanaPrivateKey":    privateKeyBase58,
			"solanaWalletAddress": walletAddress,
			"updatedAt":           time.Now(),
		},
	}
	opts := options.Update().SetUpsert(false)
	_, err = collUser.UpdateOne(ctx, filterUser, update, opts)
	if err != nil {
		common.WriteResponseWithSession(w, session, "ウォレット保存に失敗しました: "+err.Error(), http.StatusOK)
		return
	}

	responseData := struct {
		Csrf              string   `json:"csrf"`
		PushContents      []string `json:"pushContents"`
		SolanaWalletAddress string `json:"solanaWalletAddress"`
		Message           string   `json:"message"`
	}{
		Csrf:              session.Csrf,
		PushContents:      session.PushContents,
		SolanaWalletAddress: walletAddress,
		Message:           "ウォレットが正常に作成されました",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
