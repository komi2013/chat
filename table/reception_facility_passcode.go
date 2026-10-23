package database
type ReceptionFacilityPasscodeStruct struct {
	ID int64 `db:"id" json:"id"`
	FacilityID int64 `db:"facility_id" json:"facilityID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	Passkey string `db:"passkey" json:"passkey"`
	PassStart string `db:"pass_start" json:"passStart,omitempty"`
	PassEnd string `db:"pass_end" json:"passEnd,omitempty"`
}
