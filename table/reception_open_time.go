package database
type ReceptionOpenTimeStruct struct {
	ID int64 `db:"id" json:"id"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	LimitStart string `db:"limit_start" json:"limitStart"`
	LimitEnd string `db:"limit_end" json:"limitEnd"`
}
