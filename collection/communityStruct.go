package collection

import (
  "time"
)

type CommunityStruct struct {
	CommunityID  string    `bson:"_id,omitempty"`
  ChannelID  string    `bson:"channel_id,omitempty"`
  AliasName  string    `bson:"alias_name,omitempty"`
  AliasImg  string    `bson:"alias_img,omitempty"`
  UnreadFlg  int      `bson:"unread_flg,omitempty"`
  UserIDs  []string `bson:"user_ids,omitempty"`
  ChannelDB  int `bson:"channel_db,omitempty"`
  AliasDB  int `bson:"alias_db,omitempty"`
  UpdatedAt  time.Time `bson:"updated_at,omitempty"`
}
