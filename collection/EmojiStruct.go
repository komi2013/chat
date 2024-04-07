package collection

import (
  "time"
)

type EmojiStruct struct {
  MessageID  string    `bson:"message_id,omitempty"`
  AliasName  string      `bson:"alias_name,omitempty"`
  EmojiValue  string      `bson:"emoji_value,omitempty"`
  CreatedAt  time.Time `bson:"created_at,omitempty"`
  DeleteType  int      `bson:"delete_type,omitempty"`
  ParentID  int      `bson:"parentID,omitempty"`
}
