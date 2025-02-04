package collection

import (
  "time"
)

type UserStruct struct {
  UserID       string    `bson:"_id,omitempty"`
  GoogleJWTSub string    `bson:"google_jwt_sub,omitempty"`
  Mail         string    `bson:"mail,omitempty"`
  Telephone    string    `bson:"telephone,omitempty"`
  CreatedAt    time.Time `bson:"created_at,omitempty"`
  UpdatedAt    time.Time `bson:"updated_at,omitempty"`
  SignedAt     time.Time `bson:"signed_at,omitempty"`
	ChannelAliases []ChannelAlias `bson:"channel_aliases,omitempty" json:"channel_aliases,omitempty"`
	Yen   int `bson:"yen,omitempty"` // must delete
}
