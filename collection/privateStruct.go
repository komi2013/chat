package collection

import (
  "time"
)

type PrivateStruct struct {
  ID     string    `bson:"_id"`
  ChannelID     string    `bson:"channel_id"`
  CreatorAlias     string    `bson:"creator_alias"`
  CreatorUser        string    `bson:"creator_user"`
  Contents      string    `bson:"contents"`
  UpdatedAt     time.Time    `bson:"updated_at"`
  Subscriptions      []string    `bson:"subscriptions"`
  Contents2      string    `bson:"contents2"`
}
