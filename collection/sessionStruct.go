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
  PushContents []string `bson:"push_contents,omitempty"`
}

// type Content struct {
// 	PushID string      `bson:"push_id"`
// 	Data   string `bson:"data"`
// }
