package collection

import (
  "time"
)

type ChannelGroupStruct struct {
  ChannelID  string    `bson:"channel_id,omitempty"`
  AliasName  string    `bson:"alias_name,omitempty"`
  UpdatedAt  time.Time `bson:"updated_at,omitempty"`
  UnreadFlg  bool      `bson:"unread_flg,omitempty"`
}
