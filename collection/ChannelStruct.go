package collection

import (
	"time"
)

type ChannelStruct struct {
	ChannelID           string    `bson:"_id" json:"channelID,omitempty"`
	ChannelName         string    `bson:"channelName" json:"channelName,omitempty"`
	ChannelDescription  string    `bson:"channelDescription" json:"channelDescription,omitempty"`
	CreatedAt           time.Time `bson:"createdAt" json:"createdAt,omitempty"`
	UpdatedAt           time.Time `bson:"updatedAt" json:"updatedAt,omitempty"`
	InvitedAt           time.Time `bson:"invitedAt" json:"invitedAt,omitempty"`
	UpdatedBy           string    `bson:"createdBy" json:"createdBy,omitempty"`
	Aliases             []Alias   `bson:"aliases" json:"aliases,omitempty"`
	Groups              []Group   `bson:"groups" json:"groups,omitempty"`
	InvitationCode      string    `bson:"invitationCode" json:"invitationCode,omitempty"`
	InvitationGuestCode string    `bson:"invitationGuestCode,omitempty" json:"invitationGuestCode,omitempty"`
}

type Alias struct {
	AliasName   string `bson:"aliasName" json:"aliasName,omitempty"`
	AliasImg    string `bson:"aliasImg" json:"aliasImg,omitempty"`
	UserID      string `bson:"userID" json:"userID,omitempty"`
	AliasBio    string `bson:"aliasBio,omitempty" json:"aliasBio,omitempty"`
	AccessRight string `bson:"accessRight,omitempty" json:"accessRight,omitempty"`
}

type Group struct {
	GroupName  string   `bson:"groupName" json:"groupName,omitempty"`
	GroupImg   string   `bson:"groupImg" json:"groupImg,omitempty"`
	AliasNames []string `bson:"aliasNames" json:"aliasNames,omitempty"`
	GroupBio   string   `bson:"groupBio,omitempty" json:"groupBio,omitempty"`
}
