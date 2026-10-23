package database
type ReceptionWaitConfigStruct struct {
	ID int64 `db:"id" json:"id"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	GuestRange []int `db:"guest_range" json:"guestRange,omitempty"`
	WaitRatio int `db:"wait_ratio" json:"waitRatio,omitempty"`
}
