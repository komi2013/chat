package database

// ReceptionMenuFreeOptionItemStruct maps to reception_menu_free_option_item.
// Each row represents an item ID inside one FreeOptions group.
type ReceptionMenuFreeOptionItemStruct struct {
	ID                int64 `db:"id" json:"id"`
	FreeOptionGroupID int64 `db:"free_option_group_id" json:"freeOptionGroupID"`
	SortOrder         int   `db:"sort_order" json:"sortOrder"`
	ItemID            int64 `db:"item_id" json:"itemID"`
}
