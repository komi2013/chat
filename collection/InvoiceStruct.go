package collection

import (
  "time"
)

// InvoiceStruct represents a payment invoice (注文/支払い要求)
type InvoiceStruct struct {
  InvoiceID    string    `bson:"_id" json:"invoiceID"`          // MongoDB _id
  UserID       string    `bson:"userID" json:"userID"`           // 誰の注文か
  AmountJPYC   int       `bson:"amountJPYC" json:"amountJPYC"`   // 金額（JPYC）
  FromAddress      string    `bson:"fromAddress" json:"fromAddress"`
  InvoiceAddress      string    `bson:"invoiceAddress" json:"invoiceAddress"`         // 受取用ウォレットアドレス（invoice用）
  InvoiceStatus       int    `bson:"invoiceStatus" json:"invoiceStatus"`           // 1=pending, 2=paid, 3=expired
  PaidTxHash   string    `bson:"paidTxHash,omitempty" json:"paidTxHash,omitempty"` // 入金されたTX
  CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
  PaidAt       time.Time `bson:"paidAt,omitempty" json:"paidAt,omitempty"`
  ExpiresAt    time.Time `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
  AdID         string    `bson:"adID" json:"adID"`               // どの広告の注文か（AdStruct への紐付け）
  BookID       string    `bson:"bookID" json:"bookID"`
}
