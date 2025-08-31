package collection

import (
  "time"
)

type SessionStruct struct {
  SessionID  string    `bson:"_id,omitempty"`
  UserID  string    `bson:"userID,omitempty"`
  CreatedAt  time.Time `bson:"createdAt,omitempty"`
  UpdatedAt  time.Time `bson:"updatedAt,omitempty"`
  // AliasArray  [][]string      `bson:"aliasArray,omitempty"`
  ChannelAliases []ChannelAlias `bson:"channelAliases,omitempty" json:"channelAliases,omitempty"`
  Subscription  string      `bson:"subscription,omitempty"`
  Csrf  string      `bson:"csrf,omitempty"`
  PushContents []string `bson:"pushContents"`
  IsMobile bool `bson:"isMobile" json:"isMobile"`
}

type ChannelAlias struct {
    ChannelID string `bson:"channelID" json:"channelID"`
    Alias     string `bson:"alias" json:"alias"`
    Guest     bool `bson:"guest,omitempty" json:"guest,omitempty"`
}


// type Content struct {
// 	PushID string      `bson:"push_id"`
// 	Data   string `bson:"data"`
// }
