package database

// ReceptionMenuItemDetailStruct maps to reception_menu_item_detail.
type ReceptionMenuItemDetailStruct struct {
	ID          int64  `db:"id" json:"id"`
	ReceptionID int64  `db:"reception_id" json:"receptionID"`
	SortOrder   int    `db:"sort_order" json:"sortOrder"`
	ItemID      int64  `db:"item_id" json:"itemID"`
	ItemName    string `db:"item_name" json:"itemName"`
	ImgPath     string `db:"img_path" json:"imgPath,omitempty"`
}
