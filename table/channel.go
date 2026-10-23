package table

import "time"

// ChannelStruct represents a row in the channel table.
type ChannelStruct struct {
	ChannelID           int64     `json:"channelId" db:"channel_id"`
	ChannelName         string    `json:"channelName" db:"channel_name"`
	ChannelDescription  string    `json:"channelDescription,omitempty" db:"channel_description"`
	CreatedAt           time.Time `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt           time.Time `json:"updatedAt,omitempty" db:"updated_at"`
	UpdatedBy           string    `json:"updatedBy,omitempty" db:"updated_by"`
	InvitationCode      string    `json:"invitationCode,omitempty" db:"invitation_code"`
	InvitationGuestCode string    `json:"invitationGuestCode,omitempty" db:"invitation_guest_code"`
	InvitedAt           time.Time `json:"invitedAt,omitempty" db:"invited_at"`
}
