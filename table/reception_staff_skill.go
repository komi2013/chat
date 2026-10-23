package database
type ReceptionStaffSkillStruct struct {
	ID int64 `db:"id" json:"id"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	AliasName string `db:"alias_name" json:"aliasName,omitempty"`
	Skills []string `db:"skills" json:"skills,omitempty"`
}
