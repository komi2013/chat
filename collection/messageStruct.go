package collection

import (
  "time"
)

type MessageStruct struct {
  MessageID  string    `bson:"_id,omitempty"`
  ChannelID  string    `bson:"channel_id,omitempty"`
  MessageTxt  string      `bson:"message_txt,omitempty"`
  From  string      `bson:"from,omitempty"`
  FromImg  string      `bson:"from_img,omitempty"`
  Task  string      `bson:"task,omitempty"`
  CreatedAt  time.Time `bson:"created_at,omitempty"`
}
