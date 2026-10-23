package database
type ReceptionPasscodeStruct struct {
	ID int64 `db:"id" json:"id"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	Passkey string `db:"passkey" json:"passkey"`
	PassStart string `db:"pass_start" json:"passStart,omitempty"`
	PassEnd string `db:"pass_end" json:"passEnd,omitempty"`
}
