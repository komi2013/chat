package database

// ReceptionMenuItemContentStruct maps to reception_menu_item_content.
// Each row represents one value in the original Menu.Items slice.
type ReceptionMenuItemContentStruct struct {
	ID         int64 `db:"id" json:"id"`
	MenuItemID int64 `db:"menu_item_id" json:"menuItemID"`
	SortOrder  int   `db:"sort_order" json:"sortOrder"`
	ItemID     int64 `db:"item_id" json:"itemID"`
}
