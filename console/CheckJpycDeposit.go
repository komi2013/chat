package console

import (
    "context"
    "encoding/json"
    "fmt"
    // "log"
    "net/http"
    "strings"
    "time"

    "chat/common"
    "chat/collection"

    "go.mongodb.org/mongo-driver/bson"
)

type TokenTxResponse struct {
    Status  string `json:"status"`
    Message string `json:"message"`
    Result  []struct {
        Hash        string `json:"hash"`
        From        string `json:"from"`
        To          string `json:"to"`
        Value       string `json:"value"`
        TimeStamp   string `json:"timeStamp"`
        TokenSymbol string `json:"tokenSymbol"`
    } `json:"result"`
}

type Tx struct {
    Hash        string `json:"hash"`
    From        string `json:"from"`
    To          string `json:"to"`
    Value       string `json:"value"`
    TimeStamp   string `json:"timeStamp"`
    TokenSymbol string `json:"tokenSymbol"`
}

func CheckJpycDeposit() {
		cfg := common.LoadConfig()
    now := time.Now()

    // ログ準備
    var checkJpycLog = common.NewDailyLogger("check_jpyc_log_")
    var checkJpycErr = common.NewDailyLogger("check_jpyc_err_")

    collInvoice := common.DB.InvoiceDB.Collection("invoice")

    // InvoiceStatus = 1 のみ取得
    cursor, err := collInvoice.Find(
        context.TODO(),
        bson.M{"invoiceStatus": 1},
    )
    if err != nil {
        checkJpycErr.Printf("Invoice Find error:", err)
        return
    }

    var pendingInvoices []collection.InvoiceStruct
    if err := cursor.All(context.TODO(), &pendingInvoices); err != nil {
        checkJpycErr.Printf("Invoice decode error:", err)
        return
    }

    if len(pendingInvoices) == 0 {
        checkJpycLog.Printf("No pending invoices")
        return
    }

    addrMap := map[string]bool{}
    var uniqueAddrs []string

    for _, inv := range pendingInvoices {
        if _, exists := addrMap[inv.InvoiceAddress]; !exists {
            addrMap[inv.InvoiceAddress] = true
            uniqueAddrs = append(uniqueAddrs, inv.InvoiceAddress)
        }
    }

    checkJpycLog.Printf("Unique invoice addresses: %v\n", uniqueAddrs)

    var mergedTx []Tx

    for _, walletAddr := range uniqueAddrs {

				url := fmt.Sprintf(
				    "https://api.etherscan.io/v2/api?chain=polygon&module=account&action=tokentx&address=%s&contractaddresstoken=%s&page=1&offset=50&sort=desc",
				    walletAddr,
				    common.JpycContract,
				)

				req, _ := http.NewRequest("GET", url, nil)
				req.Header.Set("x-api-key", cfg.EtherscanApiKey)

				resp, err := http.DefaultClient.Do(req)

        if err != nil {
            checkJpycErr.Printf("HTTP error:", err)
            continue
        }
        defer resp.Body.Close()

        var apiResp TokenTxResponse
        if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
            checkJpycErr.Printf("JSON decode error:", err)
            continue
        }

        if apiResp.Status != "1" {
            checkJpycLog.Printf("No JPYC transfers for address=%s\n", walletAddr)
            continue
        }

        // ✅ マージ
        for _, tx := range apiResp.Result {
            mergedTx = append(mergedTx, Tx{
                Hash:        tx.Hash,
                From:        tx.From,
                To:          tx.To,
                Value:       tx.Value,
                TimeStamp:   tx.TimeStamp,
                TokenSymbol: tx.TokenSymbol,
            })
        }
    }

    if len(mergedTx) == 0 {
        checkJpycLog.Printf("No transfers found for all addresses")
        return
    }

    // ✅ mergedTx をループ処理（以前の data.Result と同じ）
    for _, tx := range mergedTx {

        // まずログに全部残す（後に金額・作成日付とのマッチング用）
        checkJpycLog.Printf("TX: Hash=%s From=%s To=%s Value=%s\n",
            tx.Hash, tx.From, tx.To, tx.Value)

        // ✅ invoice のループを外側に
        for _, inv := range pendingInvoices {

            // ✅ ExpiresAt が過ぎてたら削除
            if inv.ExpiresAt.Before(now) {
                checkJpycLog.Printf("Invoice expired and removed: invoiceID=%s\n", inv.InvoiceID)
                collInvoice.DeleteOne(context.TODO(), bson.M{"_id": inv.InvoiceID})
                continue
            }

            // ✅ Fromアドレス一致チェック
            if strings.ToLower(tx.From) == strings.ToLower(inv.FromAddress) {
								if tx.Value != fmt.Sprintf("%d000000000000000000", inv.AmountJPYC) {
								    continue
								}
								if inv.PaidTxHash == tx.Hash {
								    continue
								}
								if strings.ToLower(tx.To) != strings.ToLower(inv.InvoiceAddress) {
								    continue
								}

                checkJpycLog.Printf("MATCH: invoiceID=%s txHash=%s\n", inv.InvoiceID, tx.Hash)

                // ✅ invoice status を paid(2) に更新
                _, err := collInvoice.UpdateOne(
                    context.TODO(),
                    bson.M{"_id": inv.InvoiceID},
                    bson.M{
                        "$set": bson.M{
                            "invoiceStatus": 2,
                            "paidAt":        now,
                            "paidTxHash":    tx.Hash,
                        },
                    },
                )
                if err != nil {
                    checkJpycErr.Printf("Invoice update error: %s\n", err)
                    continue
                }

                // ✅ 対象 ad の PaidAt も更新
                adColl := common.DB.AdDB.Collection("ad")
                _, err = adColl.UpdateOne(
                    context.TODO(),
                    bson.M{"_id": inv.AdID},
                    bson.M{"$set": bson.M{"paidAt": now}},
                )
                if err != nil {
                    checkJpycErr.Printf("Ad update error: %s\n", err)
                }

                checkJpycLog.Printf("Invoice %s paid and Ad %s updated\n",
                    inv.InvoiceID, inv.AdID)
            }
        }
    }
}
