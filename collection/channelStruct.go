package collection

import (
  "time"
)

type ChannelStruct struct {
  ChannelID  string    `bson:"_id,omitempty"`
  ChannelName  string    `bson:"channel_name,omitempty"`
  ChannelDescription  string      `bson:"channel_description,omitempty"`
  UpdatedAt  time.Time `bson:"updated_at,omitempty"`
}

// channel_id
// channel_name
// channel_description
// updated_at