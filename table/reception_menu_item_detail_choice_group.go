package database

// ReceptionMenuItemDetailChoiceGroupStruct maps to
// reception_menu_item_detail_choice_group.
// One row represents one inner []string in ItemDetail.Choices.
type ReceptionMenuItemDetailChoiceGroupStruct struct {
	ID           int64 `db:"id" json:"id"`
	ItemDetailID int64 `db:"item_detail_id" json:"itemDetailID"`
	SortOrder    int   `db:"sort_order" json:"sortOrder"`
}
