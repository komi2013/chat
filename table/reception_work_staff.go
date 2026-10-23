package database
type ReceptionWorkStaffStruct struct {
	ID int64 `db:"id" json:"id"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	ShiftID     int64  `db:"shift_id" json:"shiftID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	AliasName string `db:"alias_name" json:"aliasName"`
	WorkStart string `db:"work_start" json:"workStart"`
	WorkEnd string `db:"work_end" json:"workEnd"`
	Seq int `db:"seq" json:"seq"`
	Skill     string `bson:"-" json:"-"`
}
