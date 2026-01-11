package controller

import (
	"context"
	"encoding/json"
	// "log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/collection"
	"chat/common"
)

func AdInvoice(w http.ResponseWriter, r *http.Request) {
	csrf := r.FormValue("csrf")
	adID := r.FormValue("adID")

	if adID == "" {
		common.WriteResponseWithoutSession(w, csrf, "adID is required", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	session, err := common.SessionCheckTake(w, r, csrf)
	if err != nil {
		common.WriteResponseWithoutSession(w, csrf, err.Error()+"; SessionCheckTake", http.StatusOK)
		return
	}

	var jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().In(jst)

	collAd := common.DB.AdDB.Collection("ad")
	filter := bson.M{"_id": adID, "userID": session.UserID}

	var ad collection.AdStruct
	if err := collAd.FindOne(ctx, filter).Decode(&ad); err != nil {
		common.WriteResponseWithSession(w, session, "広告が見つかりません: "+err.Error(), http.StatusOK)
		return
	}

	paidAtJST := ad.PaidAt
	if !paidAtJST.IsZero() {
		paidAtJST = paidAtJST.In(jst)
	}

	if !ad.InvoicedAt.IsZero() && ad.PaidAt.Before(ad.InvoicedAt) {
		common.WriteResponseWithSession(w, session, "既に請求済み、または支払いデータが古いため請求できません", http.StatusOK)
		return
	}

	newID, err := common.CountUpID("invoiceID")
	if err != nil {
		common.WriteResponseWithSession(w, session, "請求ID生成エラー: "+err.Error(), http.StatusOK)
		return
	}

	collUser := common.DB.UserDB.Collection("user")
	filterUser := bson.M{"_id": session.UserID}

	var user collection.UserResponse
	if err := collUser.FindOne(context.TODO(), filterUser).Decode(&user); err != nil {
		common.WriteResponseWithSession(w, session, "user FindOne:"+err.Error(), http.StatusOK)
		return
	}

	if user.WalletAddress == "" {
		common.WriteResponseWithSession(w, session, "JPYC アドレスをユーザーページで入力してください", http.StatusOK)
		return
	}

	// Solanaシステムウォレットアドレスを取得（Fee Payerの公開鍵から）
	cfg := common.LoadConfig()
	var systemSolanaWalletAddress string
	if cfg.SolanaFeePayerPrivateKey != "" {
		pubkey, err := common.GetSystemFeePayerPublicKey(cfg.SolanaFeePayerPrivateKey)
		if err == nil {
			systemSolanaWalletAddress = pubkey.String()
		}
	}
	if systemSolanaWalletAddress == "" {
		common.WriteResponseWithSession(w, session, "システムウォレットアドレスが設定されていません", http.StatusOK)
		return
	}

	invoice := collection.InvoiceStruct{
		InvoiceID:      newID,
		UserID:         session.UserID,
		AdID:           ad.AdID,
		AmountJPYC:     ad.AdYen,
		FromAddress:    user.WalletAddress,
		InvoiceAddress: systemSolanaWalletAddress, // Solanaアドレスを使用
		InvoiceStatus:  1, // pending
		CreatedAt:      now,
		ExpiresAt:      now.Add(24 * time.Hour),
	}

	collInvoice := common.DB.InvoiceDB.Collection("invoice")
	if _, err := collInvoice.InsertOne(ctx, invoice); err != nil {
		common.WriteResponseWithSession(w, session, "請求書作成エラー: "+err.Error(), http.StatusOK)
		return
	}

	ad.InvoicedAt = now
	ad.UpdatedAt = now

	update := bson.M{
		"$set": bson.M{
			"invoicedAt": ad.InvoicedAt,
			"updatedAt":  ad.UpdatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	if _, err := collAd.UpdateOne(ctx, filter, update, opts); err != nil {
		common.WriteResponseWithSession(w, session, "Ad更新エラー: "+err.Error(), http.StatusOK)
		return
	}

	resp := struct {
		Csrf   string              `json:"csrf"`
		Ad     collection.AdStruct `json:"ad"`
		Status string              `json:"status"`
	}{
		Csrf:   session.Csrf,
		Ad:     ad,
		Status: "invoice_created",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
