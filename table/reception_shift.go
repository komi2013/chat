package database
type ReceptionShiftStruct struct {
	ShiftID     int64    `db:"shift_id" json:"shiftID"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	AliasNames []string `db:"alias_names" json:"aliasNames"`
	ShiftStart string `db:"shift_start" json:"shiftStart"`
	ShiftEnd string `db:"shift_end" json:"shiftEnd"`
	Open int `db:"open" json:"open"`
	Skill string `db:"skill" json:"skill"`
	Fix bool `db:"fix" json:"fix,omitempty"`
}
