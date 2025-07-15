package collection

import (
  "time"
)

type ChannelStruct struct {
  InvitationCode     string    `bson:"_id"`
  ChannelID     string    `bson:"channelID"`
  ChannelName      string    `bson:"channelName"`
  ChannelDescription      string    `bson:"channelDescription"`
  CreatedAt     time.Time    `bson:"createdAt"`
  UpdatedAt     time.Time    `bson:"updatedAt"`
  CreatedBy     string    `bson:"createdBy"`
  // Subscriptions      []string    `bson:"subscriptions"`
  AliasNames      []string    `bson:"aliasNames"`
  Aliases          []Alias `bson:"aliases"`
  // Groups          []Group `bson:"groups"`
  NoRightMention  bool `bson:"noRightMention,omitempty"`
  UntilDate  time.Time `bson:"untilDate,omitempty"`
  PushSessions    []SessionStruct `bson:"pushSessions,omitempty"`
}

// type PushSession struct {
//   SessionID  string
//   Subscription  string
// }


type Alias struct {
	AliasName string `bson:"aliasName"`
	AliasImg string `bson:"aliasImg"`
	UserID string `bson:"userID"`
}

// type Group struct {
// 	GroupName string
// 	GroupImg string
// 	AliasNames []string
// }