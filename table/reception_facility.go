package database
type ReceptionFacilityStruct struct {
	FacilityID int64 `db:"facility_id" json:"facilityID"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	FacilityName string `db:"facility_name" json:"facilityName"`
	Capacity int `db:"capacity" json:"capacity"`
	CurrentCode string `db:"current_code" json:"currentCode,omitempty"`
	Bookable bool `db:"bookable" json:"bookable,omitempty"`
}
