package database

// ReceptionMenuItemStruct maps to the reception_menu_item table.
// Nullable SQL fields should be selected with COALESCE when scanning into
// these non-pointer Go fields.
type ReceptionMenuItemStruct struct {
	ID              int64  `db:"id" json:"id"`
	ReceptionID     int64  `db:"reception_id" json:"receptionID"`
	SortOrder       int    `db:"sort_order" json:"sortOrder"`
	MenuID          int64  `db:"menu_id" json:"menuID"`
	MenuName        string `db:"menu_name" json:"menuName"`
	Price           int    `db:"price" json:"price"`
	PrepaidPrice    int    `db:"prepaid_price" json:"prepaidPrice"`
	NeedSkill       string `db:"need_skill" json:"needSkill,omitempty"`
	NeedFacility    string `db:"need_facility" json:"needFacility,omitempty"`
	SpecifyNameFlag bool   `db:"specify_name_flag" json:"specifyNameFlag"`
	SpendMinute     int    `db:"spend_minute" json:"spendMinute,omitempty"`
	ForBookType     int    `db:"for_book_type" json:"forBookType,omitempty"`
}
