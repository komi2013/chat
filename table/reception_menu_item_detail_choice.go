package database

// ReceptionMenuItemDetailChoiceStruct maps to reception_menu_item_detail_choice.
// Each row represents one choice string in an ItemDetail.Choices group.
type ReceptionMenuItemDetailChoiceStruct struct {
	ID            int64  `db:"id" json:"id"`
	ChoiceGroupID int64  `db:"choice_group_id" json:"choiceGroupID"`
	SortOrder     int    `db:"sort_order" json:"sortOrder"`
	ChoiceText    string `db:"choice_text" json:"choiceText"`
}
