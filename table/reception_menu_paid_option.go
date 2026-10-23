package database

// ReceptionMenuPaidOptionStruct maps to reception_menu_paid_option.
// Each row represents one ItemOption in Menu.PaidOptions.
type ReceptionMenuPaidOptionStruct struct {
	ID         int64 `db:"id" json:"id"`
	MenuItemID int64 `db:"menu_item_id" json:"menuItemID"`
	SortOrder  int   `db:"sort_order" json:"sortOrder"`
	ItemID     int64 `db:"item_id" json:"itemID"`
	Price      int   `db:"price" json:"price"`
}
