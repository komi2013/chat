package console

import (
	"context"
	"encoding/binary"
	"time"

	"chat/common"
	"chat/collection"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"go.mongodb.org/mongo-driver/bson"
)

func CheckSolanaDeposit() {
	now := time.Now()

	// ログ準備
	var checkSolanaLog = common.NewDailyLogger("check_solana_log_")
	var checkSolanaErr = common.NewDailyLogger("check_solana_err_")

	collInvoice := common.DB.InvoiceDB.Collection("invoice")

	// InvoiceStatus = 1 かつ paidTxHash があるものを取得
	cursor, err := collInvoice.Find(
		context.TODO(),
		bson.M{
			"invoiceStatus": 1,
			"paidTxHash":    bson.M{"$exists": true, "$ne": ""},
		},
	)
	if err != nil {
		checkSolanaErr.Printf("Invoice Find error: %v\n", err)
		return
	}

	var pendingInvoices []collection.InvoiceStruct
	if err := cursor.All(context.TODO(), &pendingInvoices); err != nil {
		checkSolanaErr.Printf("Invoice decode error: %v\n", err)
		return
	}

	if len(pendingInvoices) == 0 {
		checkSolanaLog.Printf("No pending invoices with transaction hash\n")
		return
	}

	checkSolanaLog.Printf("Found %d pending invoices to check\n", len(pendingInvoices))

	client := rpc.New(rpc.MainNetBeta_RPC)

	for _, inv := range pendingInvoices {
		// ExpiresAt が過ぎてたら削除
		if !inv.ExpiresAt.IsZero() && inv.ExpiresAt.Before(now) {
			checkSolanaLog.Printf("Invoice expired and removed: invoiceID=%s\n", inv.InvoiceID)
			collInvoice.DeleteOne(context.TODO(), bson.M{"_id": inv.InvoiceID})
			continue
		}

		// トランザクション署名を取得
		if inv.PaidTxHash == "" {
			checkSolanaErr.Printf("Invoice %s has no paidTxHash\n", inv.InvoiceID)
			continue
		}

		// トランザクション署名をパース
		sig, err := solana.SignatureFromBase58(inv.PaidTxHash)
		if err != nil {
			checkSolanaErr.Printf("Invalid signature for invoice %s: %v\n", inv.InvoiceID, err)
			continue
		}

		// トランザクションを取得（Finalizedで確認）
		out, err := client.GetTransaction(
			context.Background(),
			sig,
			&rpc.GetTransactionOpts{
				Commitment: rpc.CommitmentFinalized,
				Encoding:   solana.EncodingJSON,
			},
		)

		if err != nil || out == nil {
			// まだ確定していない、または見つからない
			checkSolanaLog.Printf("Transaction not finalized yet: invoiceID=%s txHash=%s\n", inv.InvoiceID, inv.PaidTxHash)
			continue
		}

		// トランザクションが失敗していないか確認
		if out.Meta == nil || out.Meta.Err != nil {
			checkSolanaErr.Printf("Transaction failed: invoiceID=%s txHash=%s err=%v\n", inv.InvoiceID, inv.PaidTxHash, out.Meta.Err)
			continue
		}

		// トランザクション内容を取得
		tx, err := out.Transaction.GetTransaction()
		if err != nil {
			checkSolanaErr.Printf("Failed to get transaction: invoiceID=%s error=%v\n", inv.InvoiceID, err)
			continue
		}

		// システムウォレットアドレスを取得（InvoiceAddressから）
		systemWallet, err := solana.PublicKeyFromBase58(inv.InvoiceAddress)
		if err != nil {
			// Ethereumアドレスの場合は、別の方法で取得する必要がある
			// ここではエラーとして処理
			checkSolanaErr.Printf("Invalid invoice address (not Solana format): invoiceID=%s address=%s\n", inv.InvoiceID, inv.InvoiceAddress)
			continue
		}

		// JPYC送金を検証
		expectedAmount := uint64(inv.AmountJPYC * 1_000_000) // JPYC amount in lamports (decimals=6)
		foundPayment := false

		msg := tx.Message
		for _, inst := range msg.Instructions {
			// ProgramIDIndex を使ってプログラムIDを取得
			if int(inst.ProgramIDIndex) >= len(msg.AccountKeys) {
				continue
			}
			programID := msg.AccountKeys[inst.ProgramIDIndex]

			// Token Program IDを確認
			if programID.Equals(token.ProgramID) {
				// Transfer instruction (3) を確認
				if len(inst.Data) < 9 || inst.Data[0] != 3 {
					continue
				}

				amount := binary.LittleEndian.Uint64(inst.Data[1:9])

				// アカウントを確認（Transfer instructionは通常、source, destination, authorityの順）
				if len(inst.Accounts) < 2 {
					continue
				}

				// destination accountを取得
				if int(inst.Accounts[1]) >= len(msg.AccountKeys) {
					continue
				}
				destination := msg.AccountKeys[inst.Accounts[1]]

				// システムウォレットへの送金を確認（商品代金）
				if destination.Equals(systemWallet) && amount == expectedAmount {
					foundPayment = true
					checkSolanaLog.Printf("Found payment: invoiceID=%s amount=%d\n", inv.InvoiceID, amount)
				}
			}
		}

		// 商品代金の送金が確認できたかチェック
		if !foundPayment {
			checkSolanaErr.Printf("Payment not found in transaction: invoiceID=%s expectedAmount=%d\n", inv.InvoiceID, expectedAmount)
			continue
		}

		// システム利用料の送金が確認できたかチェック（オプション）
		// システム利用料が別のウォレットに送金される場合は、ここで確認
		// 現在の実装では、システム利用料も同じウォレットに送金されることを想定
		// 必要に応じて、別のウォレットアドレスを環境変数から取得して検証

		checkSolanaLog.Printf("MATCH: invoiceID=%s txHash=%s\n", inv.InvoiceID, inv.PaidTxHash)

		// Invoice status を paid(2) に更新
		_, err = collInvoice.UpdateOne(
			context.TODO(),
			bson.M{"_id": inv.InvoiceID},
			bson.M{
				"$set": bson.M{
					"invoiceStatus": 2,
					"paidAt":        now,
				},
			},
		)
		if err != nil {
			checkSolanaErr.Printf("Invoice update error: invoiceID=%s error=%v\n", inv.InvoiceID, err)
			continue
		}

		// 対象 ad の PaidAt も更新
		adColl := common.DB.AdDB.Collection("ad")
		_, err = adColl.UpdateOne(
			context.TODO(),
			bson.M{"_id": inv.AdID},
			bson.M{"$set": bson.M{"paidAt": now}},
		)
		if err != nil {
			checkSolanaErr.Printf("Ad update error: invoiceID=%s adID=%s error=%v\n", inv.InvoiceID, inv.AdID, err)
		}

		checkSolanaLog.Printf("Invoice %s paid and Ad %s updated\n", inv.InvoiceID, inv.AdID)
	}
}
