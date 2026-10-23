package database

import "time"

// ReceptionStruct maps to the reception table.
type ReceptionStruct struct {
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	ChannelID int64 `db:"channel_id" json:"channelID"`
	AdminNames []string `db:"admin_names" json:"adminNames"`
	JoinNames []string `db:"join_names" json:"joinNames"`
	InquiryNames []string `db:"inquiry_names" json:"inquiryNames,omitempty"`
	Subscriptions []string `db:"subscriptions" json:"subscriptions,omitempty"`
	OrderUserIDs []int64 `db:"order_user_ids" json:"orderUserIDs,omitempty"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt,omitempty"`
	ReceptionTitle string `db:"reception_title" json:"receptionTitle,omitempty"`
	Skills []string `db:"skills" json:"skills,omitempty"`
	WorkStaffNeed bool `db:"work_staff_need" json:"workStaffNeed,omitempty"`
}
