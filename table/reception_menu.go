package database

import "time"

// ReceptionMenuStruct maps to the reception_menu table.
// Menu entries and item details are stored in their own tables.
type ReceptionMenuStruct struct {
	ReceptionID int64     `db:"reception_id" json:"receptionID"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt,omitempty"`
}
