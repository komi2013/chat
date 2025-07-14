package collection

import (
  "time"
)

type UserStruct struct {
  UserID       string    `bson:"_id,omitempty" json:"userID"`
  GoogleJWTSub string    `bson:"google_jwt_sub,omitempty" json:"googleJWTSub,omitempty"`
  Mail         string    `bson:"mail,omitempty" json:"mail,omitempty"`
  Telephone    string    `bson:"telephone,omitempty" json:"telephone,omitempty"`
  CreatedAt    time.Time `bson:"created_at,omitempty" json:"createdAt,omitempty"`
  UpdatedAt    time.Time `bson:"updated_at,omitempty" json:"updatedAt,omitempty"`
  SignedAt     time.Time `bson:"signed_at,omitempty" json:"signedAt,omitempty"`
	ChannelAliases []ChannelAlias `bson:"channel_aliases,omitempty" json:"channelAliases,omitempty"`
	Yen   int `bson:"yen,omitempty" json:"yen,omitempty"`
	Latitude   float64 `bson:"latitude" json:"latitude,omitempty"`   // 例: 35.73
	Longitude  float64 `bson:"longitude" json:"longitude,omitempty"` // 例: 139.53
}
