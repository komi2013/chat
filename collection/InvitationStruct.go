package collection

import (
  "time"
)

type InvitationStruct struct {
  InvitationCode     string    `bson:"_id"`
  ChannelID     string    `bson:"channel_id"`
  ChannelName      string    `bson:"channel_name"`
  ChannelDescription      string    `bson:"channel_description"`
  CreatedAt     time.Time    `bson:"created_at"`
  CreatedBy     string    `bson:"created_by"`
  Subscriptions      []string    `bson:"subscriptions"`
  AliasNames      []string    `bson:"aliasNames"`
  Aliases          []Alias `bson:"aliases"`
  Groups          []Group `bson:"groups"`
}

type Alias struct {
	AliasName string
	AliasImg string
	UserID string
}

type Group struct {
	GroupName string
	GroupImg string
	AliasNames []string
}