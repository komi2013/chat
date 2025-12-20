package collection

import (
	"time"
)

type ChannelStruct struct {
	ChannelID           string    `bson:"_id" json:"channelID"`
	ChannelName         string    `bson:"channelName" json:"channelName"`
	ChannelDescription  string    `bson:"channelDescription" json:"channelDescription,omitempty"`
	CreatedAt           time.Time `bson:"createdAt" json:"createdAt,omitempty"`
	UpdatedAt           time.Time `bson:"updatedAt" json:"updatedAt,omitempty"`
	UpdatedBy           string    `bson:"createdBy" json:"createdBy,omitempty"`
	Aliases             []Alias   `bson:"aliases" json:"aliases,omitempty"`
	Groups              []Group   `bson:"groups" json:"groups,omitempty"`
	Myname              string    `bson:"myname" json:"myname"`
	// Invitations         []Invitation   `bson:"invitations,omitempty" json:"invitations,omitempty"`
	InvitationCode      string    `bson:"invitationCode" json:"invitationCode,omitempty"`
	InvitationGuestCode string    `bson:"invitationGuestCode,omitempty" json:"invitationGuestCode,omitempty"`
	InvitedAt time.Time `bson:"invitedAt" json:"invitedAt,omitempty"`
}

type Alias struct {
	AliasID     string `bson:"aliasID" json:"aliasID"`
	ChannelID   string    `bson:"channelID" json:"channelID,omitempty"`
	AliasName   string `bson:"aliasName" json:"aliasName,omitempty"`
	AliasImg    string `bson:"aliasImg" json:"aliasImg,omitempty"`
	UserID      string `bson:"userID" json:"userID,omitempty"`
	AliasBio    string `bson:"aliasBio,omitempty" json:"aliasBio,omitempty"`
	AccessRight string `bson:"accessRight,omitempty" json:"accessRight,omitempty"`
}

type Group struct {
	GroupID    string   `bson:"groupID" json:"groupID"`
	ChannelID   string    `bson:"channelID" json:"channelID,omitempty"`
	GroupName  string   `bson:"groupName" json:"groupName,omitempty"`
	GroupImg   string   `bson:"groupImg" json:"groupImg,omitempty"`
	AliasNames []string `bson:"aliasNames" json:"aliasNames,omitempty"`
	GroupBio   string   `bson:"groupBio,omitempty" json:"groupBio,omitempty"`
}

type Invitation struct {
	Code      string    `bson:"code" json:"code,omitempty"`
	GuestCode string    `bson:"guestCode,omitempty" json:"guestCode,omitempty"`
	InvitedAt time.Time `bson:"invitedAt" json:"invitedAt,omitempty"`
}
