package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"chat/common"
	"chat/collection"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"go.mongodb.org/mongo-driver/bson"
)

func PaymentExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	csrf := r.FormValue("csrf")
	adID := r.FormValue("adID")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, err.Error()+"; SessionCheckTake", http.StatusOK)
		return
	}

	// ユーザー情報を取得
	collUser := common.DB.UserDB.Collection("user")
	filterUser := bson.M{"_id": session.UserID}
	var user collection.UserStruct
	err = collUser.FindOne(ctx, filterUser).Decode(&user)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+"; FindOne user", http.StatusOK)
		return
	}

	// ウォレットが存在しない場合
	if user.SolanaPrivateKey == "" || user.SolanaWalletAddress == "" {
		common.WriteResponseWithSession(w, session, "ウォレットが作成されていません", http.StatusOK)
		return
	}

	// 広告情報を取得
	collAd := common.DB.AdDB.Collection("ad")
	var ad collection.AdStruct
	err = collAd.FindOne(ctx, bson.M{"_id": adID}).Decode(&ad)
	if err != nil {
		common.WriteResponseWithSession(w, session, err.Error()+"; FindOne ad", http.StatusOK)
		return
	}

	// 設定を取得
	cfg := common.LoadConfig()
	if cfg.SolanaFeePayerPrivateKey == "" {
		common.WriteResponseWithSession(w, session, "Solana fee payer private key is not configured", http.StatusOK)
		return
	}

	// システム利用料ウォレットアドレス（Solana形式）を取得（Fee Payerの公開鍵から）
	feePayerPrivateKey, err := solana.PrivateKeyFromBase58(cfg.SolanaFeePayerPrivateKey)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Invalid fee payer private key: %v", err), http.StatusOK)
		return
	}
	systemFeeWallet := feePayerPrivateKey.PublicKey()
	systemWalletAddress := systemFeeWallet.String()

	// ユーザーの秘密鍵を取得
	userPrivateKey, err := solana.PrivateKeyFromBase58(user.SolanaPrivateKey)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Invalid user private key: %v", err), http.StatusOK)
		return
	}
	userPubkey := userPrivateKey.PublicKey()

	// 広告料金をJPYCに変換（1円 = 1 JPYC）
	jpycAmount := ad.AdYen // 円単位
	jpycAmountLamports := uint64(jpycAmount * 1_000_000) // JPYCは通常decimals=6

	if jpycAmountLamports <= 0 {
		common.WriteResponseWithSession(w, session, "送金金額が0以下です", http.StatusOK)
		return
	}

	// システム利用料（1 JPYC）
	systemFeeAmount := common.SystemFeeAmount // 1 JPYC = 1,000,000 lamports (decimals=6)

	// Solana RPC接続
	client := rpc.New(rpc.MainNetBeta_RPC)

	// 最新 blockhash を取得
	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to get latest blockhash: %v", err), http.StatusOK)
		return
	}

	jpycMint, err := solana.PublicKeyFromBase58(common.JpycSolanaMint)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Invalid JPYC mint address: %v", err), http.StatusOK)
		return
	}

	merchantWallet, err := solana.PublicKeyFromBase58(systemWalletAddress)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Invalid merchant wallet address: %v", err), http.StatusOK)
		return
	}

	// トークンアカウントアドレスを取得
	userTokenAccount, err := common.FindAssociatedTokenAddress(
		userPubkey,
		jpycMint,
	)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to get user token account: %v", err), http.StatusOK)
		return
	}

	merchantTokenAccount, err := common.FindAssociatedTokenAddress(
		merchantWallet,
		jpycMint,
	)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to get merchant token account: %v", err), http.StatusOK)
		return
	}

	systemFeeTokenAccount, err := common.FindAssociatedTokenAddress(
		systemFeeWallet,
		jpycMint,
	)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to get system fee token account: %v", err), http.StatusOK)
		return
	}

	// トランザクションを作成
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			// Instruction 1: ユーザー -> 店舗アドレス: 商品代金（JPYC）
			token.NewTransferInstruction(
				jpycAmountLamports,
				userTokenAccount,
				merchantTokenAccount,
				userPubkey,
				[]solana.PublicKey{},
			).Build(),
			// Instruction 2: ユーザー -> システムアドレス: 手数料（1 JPYC）
			token.NewTransferInstruction(
				systemFeeAmount,
				userTokenAccount,
				systemFeeTokenAccount,
				userPubkey,
				[]solana.PublicKey{},
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(systemFeeWallet), // Fee Payerはシステムウォレット
	)

	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to create transaction: %v", err), http.StatusOK)
		return
	}

	// ユーザーで署名
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(userPubkey) {
			return &userPrivateKey
		}
		return nil
	})
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to sign transaction: %v", err), http.StatusOK)
		return
	}

	// Fee Payerとして署名を追加
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(systemFeeWallet) {
			return &feePayerPrivateKey
		}
		return nil
	})
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to sign transaction with fee payer: %v", err), http.StatusOK)
		return
	}

	// トランザクションを送信
	sig, err := client.SendTransaction(ctx, tx)
	if err != nil {
		common.WriteResponseWithSession(w, session, fmt.Sprintf("Failed to send transaction: %v", err), http.StatusOK)
		return
	}

	signature := sig.String()

	// Invoiceを作成または更新（pending状態で保存）
	if adID != "" {
		ctxInvoice, cancelInvoice := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelInvoice()

		collInvoice := common.DB.InvoiceDB.Collection("invoice")
		
		// 既存のInvoiceを検索（adIDとpending状態）
		filter := bson.M{
			"adID":          adID,
			"invoiceStatus": 1, // pending
		}
		
		var existingInvoice collection.InvoiceStruct
		err := collInvoice.FindOne(ctxInvoice, filter).Decode(&existingInvoice)
		
		if err == nil {
			// 既存のInvoiceを更新（paidTxHashを保存）
			update := bson.M{
				"$set": bson.M{
					"paidTxHash": signature,
					"updatedAt":  time.Now(),
				},
			}
			_, err = collInvoice.UpdateOne(ctxInvoice, filter, update)
			if err != nil {
				// エラーログを出力（決済は成功しているので、レスポンスは返す）
				fmt.Printf("Failed to update invoice: %v\n", err)
			}
		} else {
			// Invoiceが見つからない場合は、新規作成
			var jst = time.FixedZone("Asia/Tokyo", 9*60*60)
			now := time.Now().In(jst)
			
			newID, err := common.CountUpID("invoiceID")
			if err == nil {
				invoice := collection.InvoiceStruct{
					InvoiceID:      newID,
					UserID:         session.UserID,
					AdID:           adID,
					AmountJPYC:     ad.AdYen,
					FromAddress:    user.SolanaWalletAddress,
					InvoiceAddress: systemWalletAddress, // Solanaアドレスを使用
					InvoiceStatus:  1, // pending
					PaidTxHash:     signature,
					CreatedAt:      now,
					ExpiresAt:      now.Add(24 * time.Hour),
				}
				collInvoice.InsertOne(ctxInvoice, invoice)
			}
		}
	}

	response := struct {
		Csrf      string `json:"csrf"`
		Signature string `json:"signature"`
		Status    string `json:"status"`
	}{
		Csrf:      session.Csrf,
		Signature: signature,
		Status:    "pending", // バッチで確定待ち
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
