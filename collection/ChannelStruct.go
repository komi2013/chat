package collection

import (
  "time"
)

type ChannelStruct struct {
  ChannelID     string    `bson:"_id"`
  InvitationCode     string    `bson:"invitationCode"`
  ChannelName      string    `bson:"channelName"`
  ChannelDescription      string    `bson:"channelDescription"`
  CreatedAt     time.Time    `bson:"createdAt"`
  // UpdatedAt     time.Time    `bson:"updatedAt"`
  CreatedBy     string    `bson:"createdBy"`
  // Subscriptions      []string    `bson:"subscriptions"`
  AliasNames      []string    `bson:"aliasNames"`
  Aliases          []Alias `bson:"aliases"`
  // Groups          []Group `bson:"groups"`
  Guest       bool `bson:"guest,omitempty"`
  UntilDate  time.Time `bson:"untilDate,omitempty"`
  PushSessions    []SessionStruct `bson:"pushSessions,omitempty"`
}

type Alias struct {
	AliasName string `bson:"aliasName"`
	AliasImg string `bson:"aliasImg"`
	UserID string `bson:"userID"`
	Bio    string `bson:"bio,omitempty"`
	AccessRight string `bson:"accessRight,omitempty"`
}

// type PushSession struct {
//   SessionID  string
//   Subscription  string
// }


// type Group struct {
// 	GroupName string
// 	GroupImg string
// 	AliasNames []string
// }