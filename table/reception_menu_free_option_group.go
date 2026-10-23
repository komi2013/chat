package database

// ReceptionMenuFreeOptionGroupStruct maps to reception_menu_free_option_group.
// One row represents one inner []int in Menu.FreeOptions.
type ReceptionMenuFreeOptionGroupStruct struct {
	ID         int64 `db:"id" json:"id"`
	MenuItemID int64 `db:"menu_item_id" json:"menuItemID"`
	SortOrder  int   `db:"sort_order" json:"sortOrder"`
}
