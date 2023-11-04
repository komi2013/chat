package collection

import (
  "time"
)

type MessageStruct struct {
  MessageID  string    `bson:"_id,omitempty"`
  ChannelID  string    `bson:"channel_id,omitempty"`
  MessageTxt  string      `bson:"message_txt,omitempty"`
  MessageType  int      `bson:"message_type,omitempty"`
  From  string      `bson:"from,omitempty"`
  EditFlg  int      `bson:"edit_flg,omitempty"`
  ParentID  string      `bson:"parent_id,omitempty"`
  Emojis  string      `bson:"emojis,omitempty"`
  UpdatedAt  time.Time `bson:"updated_at,omitempty"`
}

// message_id
// channel_id
// message_txt
// message_type
// from
// edit_flg
// parent_id
// emojis
// UpdatedAt