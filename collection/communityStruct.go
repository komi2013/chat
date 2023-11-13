package collection

import (
  "time"

  "go.mongodb.org/mongo-driver/bson/primitive"
)

type CommunityStruct struct {
	CommunityID  string    `bson:"_id,omitempty"`
  ChannelID  primitive.ObjectID    `bson:"channel_id,omitempty"`
  AliasName  string    `bson:"alias_name,omitempty"`
  AliasImg  string    `bson:"alias_img,omitempty"`
  UnreadFlg  int      `bson:"unread_flg,omitempty"`
  UserIDs  []string `bson:"user_ids,omitempty"`
  UpdatedAt  time.Time `bson:"updated_at,omitempty"`
  ChannelDB  int `bson:"channel_db,omitempty"`
  AliasDB  int `bson:"alias_db,omitempty"`
}