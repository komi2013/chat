package collection

import (
  "time"
)

type UserStruct struct {
  UserID       string    `bson:"_id,omitempty" json:"userID"`
  GoogleJWTSub string    `bson:"googleJWTSub,omitempty"`
  Mail         string    `bson:"mail,omitempty" json:"mail,omitempty"`
  Telephone    string    `bson:"telephone,omitempty" json:"telephone,omitempty"`
  CreatedAt    time.Time `bson:"createdAt,omitempty"`
  UpdatedAt    time.Time `bson:"updatedAt,omitempty"`
  SignedAt     time.Time `bson:"signedAt,omitempty"`
	ChannelAliases []ChannelAlias `bson:"channelAliases,omitempty" json:"channelAliases,omitempty"`
	// Yen   int `bson:"yen,omitempty" json:"yen,omitempty"`
	Latitude   float64 `bson:"latitude" json:"latitude"`   // 例: 35.73
	Longitude  float64 `bson:"longitude" json:"longitude"` // 例: 139.53
	WalletAddress string `bson:"walletAddress" json:"walletAddress"`
	SolanaPrivateKey string `bson:"solanaPrivateKey,omitempty" json:"-"` // 秘密鍵はJSONレスポンスに含めない
	SolanaWalletAddress string `bson:"solanaWalletAddress,omitempty" json:"solanaWalletAddress,omitempty"`
	// Nickname     
}

type UserResponse struct {
  Mail         string    `bson:"mail,omitempty" json:"mail,omitempty"`
  Telephone    string    `bson:"telephone,omitempty" json:"telephone,omitempty"`
	WalletAddress string `bson:"walletAddress" json:"walletAddress"`
	SolanaWalletAddress string `bson:"solanaWalletAddress,omitempty" json:"solanaWalletAddress,omitempty"`
	Latitude   float64 `bson:"latitude" json:"latitude"`   // 例: 35.73
	Longitude  float64 `bson:"longitude" json:"longitude"` // 例: 139.53
}
