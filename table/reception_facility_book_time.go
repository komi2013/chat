package database
type ReceptionFacilityBookTimeStruct struct {
	ID int64 `db:"id" json:"id"`
	FacilityID int64 `db:"facility_id" json:"facilityID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	BookStart string `db:"book_start" json:"bookStart"`
	BookEnd string `db:"book_end" json:"bookEnd"`
}
