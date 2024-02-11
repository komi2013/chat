package collection

import (
  "time"
)

type MessageEditStruct struct {
  MessageID  string    `bson:"message_id,omitempty"`
  AliasName  string      `bson:"alias_name,omitempty"`
  EmojiPath  string      `bson:"emoji_path,omitempty"`
  MessageTxt  string      `bson:"message_txt,omitempty"`
  CreatedAt  time.Time `bson:"created_at,omitempty"`
  DeleteType  int      `bson:"delete_type,omitempty"`
}