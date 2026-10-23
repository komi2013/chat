package database

import "time"

type ReceptionBookStruct struct {
	BookID int64 `db:"book_id" json:"bookID"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	BookStart string `db:"book_start" json:"bookStart"`
	BookEnd string `db:"book_end" json:"bookEnd"`
	Answers []string `db:"answers" json:"answers,omitempty"`
	MenuID int64 `db:"menu_id" json:"menuID,omitempty"`
	People int `db:"people" json:"people,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"createdAt,omitempty"`
	Nickname string `db:"nickname" json:"nickname"`
}
