package collection

import (
  "time"
)

type UserStruct struct {
  UserID       string    `bson:"_id,omitempty" json:"userID"`
  GoogleJWTSub string    `bson:"googleJWTSub,omitempty" json:"googleJWTSub,omitempty"`
  Mail         string    `bson:"mail,omitempty" json:"mail,omitempty"`
  Telephone    string    `bson:"telephone,omitempty" json:"telephone,omitempty"`
  CreatedAt    time.Time `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
  UpdatedAt    time.Time `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
  SignedAt     time.Time `bson:"signedAt,omitempty" json:"signedAt,omitempty"`
	ChannelAliases []ChannelAlias `bson:"channelAliases,omitempty" json:"channelAliases,omitempty"`
	Yen   int `bson:"yen,omitempty" json:"yen,omitempty"`
	Latitude   float64 `bson:"latitude" json:"latitude"`   // 例: 35.73
	Longitude  float64 `bson:"longitude" json:"longitude"` // 例: 139.53
	// Nickname     
}
