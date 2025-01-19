package collection

import (
  "time"
)

type SessionStruct struct {
  SessionID  string    `bson:"_id,omitempty"`
  UserID  string    `bson:"user_id,omitempty"`
  CreatedAt  time.Time `bson:"created_at,omitempty"`
  UpdatedAt  time.Time `bson:"updated_at,omitempty"`
  AliasArray  [][]string      `bson:"alias_array,omitempty"`
  // AliasChannels []AliasChannel `bson:"alias_channels,omitempty" json:"alias_channels,omitempty"`
  ChannelAliases []ChannelAlias `bson:"channel_aliases,omitempty" json:"channel_aliases,omitempty"`
  Subscription  string      `bson:"subscription,omitempty"`
  Csrf  string      `bson:"csrf,omitempty"`
  Yen   int `bson:"yen,omitempty"` // must delete
}

// type AliasChannel struct {
//     ChannelID string `bson:"channel_id" json:"channel_id"`
//     Alias     string `bson:"alias" json:"alias"`
// }

type ChannelAlias struct {
    ChannelID string `bson:"channel_id" json:"channel_id"`
    Alias     string `bson:"alias" json:"alias"`
}
