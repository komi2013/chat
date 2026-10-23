package database

// ReceptionMenuFreeMultiOptionStruct maps to reception_menu_free_multi_option.
// Each row represents one value in Menu.FreeMultiOptions.
type ReceptionMenuFreeMultiOptionStruct struct {
	ID         int64 `db:"id" json:"id"`
	MenuItemID int64 `db:"menu_item_id" json:"menuItemID"`
	SortOrder  int   `db:"sort_order" json:"sortOrder"`
	ItemID     int64 `db:"item_id" json:"itemID"`
}
