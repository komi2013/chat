package collection

import (
  "time"
)

type AliasStruct struct {
  AliasID  string    `bson:"_id,omitempty"`
  UserID  string    `bson:"user_id,omitempty"`
  AliasName  string `bson:"alias_name,omitempty"`
  AliasImg  string `bson:"alias_img,omitempty"`
  GroupFlg  int `bson:"group_flg,omitempty"`
  UpdatedAt  time.Time `bson:"updated_at,omitempty"`
  ChannelIDs  []string `bson:"channel_ids,omitempty"`
}
