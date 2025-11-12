package controller

import (
	"context"
	"encoding/json"
	"log"
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

	now := time.Now()

	// --- get ad ---
	collAd := common.DB.AdDB.Collection("ad")
	filter := bson.M{"_id": adID, "userID": session.UserID}

	var ad collection.AdStruct
	if err := collAd.FindOne(ctx, filter).Decode(&ad); err != nil {
		common.WriteResponseWithSession(w, session, "広告が見つかりません: "+err.Error(), http.StatusOK)
		return
	}

	// --- 二重請求防止 ---
	// PaidAt が InvoicedAt より古い（＝既に請求済み以降の支払いがある）場合は中断
	if !ad.InvoicedAt.IsZero() && ad.PaidAt.Before(ad.InvoicedAt) {
		common.WriteResponseWithSession(w, session, "既に請求済み、または支払いデータが古いため請求できません", http.StatusOK)
		return
	}

	// --- nextPayment フラグ（互換性: currentFlag を受け取る実装もサポート） ---
	nextPayment := false
	if r.FormValue("nextPayment") != "" || r.FormValue("currentFlag") != "" {
		nextPayment = true
	}

	// --- 配信期間（曜日時刻 -> 実日時刻）計算 ---
	startDay := ad.AdStart / 100
	startHour := ad.AdStart % 100
	endDay := ad.AdEnd / 100
	endHour := ad.AdEnd % 100

	// durationHours (曜日ベース、週またぎ対応)
	durationHours := (endDay*24 + endHour) - (startDay*24 + startHour)
	if durationHours <= 0 {
		durationHours += 7 * 24
	}

	// paidWeekStart: PaidAt を含む「週の0日（日曜）」の0:00
	paidWeekStart := ad.PaidAt.Truncate(24 * time.Hour).Add(-time.Duration(int(ad.PaidAt.Weekday())) * 24 * time.Hour)
	startTime := paidWeekStart.Add(time.Duration(startDay*24+startHour) * time.Hour)
	endTime := paidWeekStart.Add(time.Duration(endDay*24+endHour) * time.Hour)
	if endTime.Before(startTime) {
		endTime = endTime.Add(7 * 24 * time.Hour) // 週跨ぎ補正
	}

	// isActive: 現在が配信期間に含まれているか
	isActive := now.After(startTime) && now.Before(endTime)

	// --- nextInvoicedFlag と PaidAt 関係の短期スキップ ---
	// nextInvoicedFlag が true かつ PaidAt から 14日未満 の場合は既請求とみなしてスキップ
	if ad.NextInvoicedFlag && now.Sub(ad.PaidAt) < 14*24*time.Hour {
		common.WriteResponseWithSession(w, session, "既に次回請求があり（14日未満）、新しい請求は不要です", http.StatusOK)
		return
	}
	// （注）NextPayFlag==true でも PaidAt から 14日以上経過していれば 再請求は許可（下で判定）

	// --- 請求判定 ---
	shouldInvoice := false

	// 1) ユーザーが明示的に次回支払いを希望（配信中でも請求許可）
	if nextPayment {
		// ただし二重請求防止（上で PaidAt.Before(InvoicedAt) を確認済み）
		shouldInvoice = true
	}

	// 2) PaidAt から 7日以上経過（通常の更新請求）
	if !shouldInvoice && now.Sub(ad.PaidAt) > 7*24*time.Hour {
		// ただし「現在配信中」で、ユーザーが nextPayment を指定していない場合は請求しない
		if !isActive {
			shouldInvoice = true
		} else {
			// isActive && !nextPayment -> 配信中で勝手に請求しない
			shouldInvoice = false
		}
	}

	// --- 最終確認: まだ請求しない条件なら中断 ---
	if !shouldInvoice {
		common.WriteResponseWithSession(w, session, "請求条件に該当しません（請求不要）", http.StatusOK)
		return
	}

	// --- Invoice 作成 ---
	newID, err := common.CountUpID("invoiceID")
	if err != nil {
		common.WriteResponseWithSession(w, session, "請求ID生成エラー: "+err.Error(), http.StatusOK)
		return
	}

	// addr, err := common.MyAddress
	// if err != nil {
	// 	common.WriteResponseWithSession(w, session, "ウォレット生成エラー: "+err.Error(), http.StatusOK)
	// 	return
	// }

  collUser := common.DB.UserDB.Collection("user")
  filterUser := bson.M{"_id": session.UserID}
  var user collection.UserResponse
  err = collUser.FindOne(context.TODO(), filterUser).Decode(&user)
  if err != nil {
		common.WriteResponseWithSession(w, session, "user FindOne:"+err.Error(), http.StatusOK)
		return
  }

	invoice := collection.InvoiceStruct{
		InvoiceID:      newID,
		UserID:         session.UserID,
		AdID:           ad.AdID,
		AmountJPYC:     ad.AdYen,
		FromAddress:    user.WalletAddress,
		InvoiceAddress: common.SystemWalletAddress,
		InvoiceStatus:  1, // pending
		CreatedAt:      now,
		ExpiresAt:      now.Add(24 * time.Hour),
	}

	collInvoice := common.DB.InvoiceDB.Collection("invoice")
	if _, err := collInvoice.InsertOne(ctx, invoice); err != nil {
		common.WriteResponseWithSession(w, session, "請求書作成エラー: "+err.Error(), http.StatusOK)
		return
	}

	// --- Ad 更新: InvoicedAt, NextPayFlag, UpdatedAt ---
	ad.InvoicedAt = now
	// ad.nextInvoicedFlag = true
	ad.UpdatedAt = now

	update := bson.M{"$set": bson.M{
		"invoicedAt":  ad.InvoicedAt,
		// "nextPayFlag": ad.NextPayFlag,
		"updatedAt":   ad.UpdatedAt,
	}}
	opts := options.Update().SetUpsert(true)
	if _, err := collAd.UpdateOne(ctx, filter, update, opts); err != nil {
		common.WriteResponseWithSession(w, session, "Ad更新エラー: "+err.Error(), http.StatusOK)
		return
	}

	log.Printf("🧾 Invoice created for adID=%s amount=%d (nextPayment=%v, isActive=%v)", ad.AdID, ad.AdYen, nextPayment, isActive)

	// --- レスポンス ---
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
