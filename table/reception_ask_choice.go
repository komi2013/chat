package database
type ReceptionAskChoiceStruct struct {
	ID int64 `db:"id" json:"id"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	Question string `db:"question" json:"question"`
	Choices []string `db:"choices" json:"choices"`
	Sequence int `db:"sequence" json:"sequence"`
}
