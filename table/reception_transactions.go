package table

import "time"

// ReceptionBookStruct is a booking transaction. Nickname is intentionally not
// persisted; book_id is the externally displayed booking number.
type ReceptionBookStruct struct {
	BookID      int64     `json:"bookID" db:"book_id"`
	ReceptionID int64     `json:"receptionID" db:"reception_id"`
	SortOrder   int       `json:"sortOrder" db:"sort_order"`
	BookStart   string    `json:"bookStart" db:"book_start"`
	BookEnd     string    `json:"bookEnd" db:"book_end"`
	Answers     []string  `json:"answers,omitempty" db:"answers"`
	MenuID      int64     `json:"menuID,omitempty" db:"menu_id"`
	People      int       `json:"people,omitempty" db:"people"`
	CreatedAt   time.Time `json:"createdAt,omitempty" db:"created_at"`
}

// ReceptionFacilityBookTimeStruct records occupied time slots for a facility.
type ReceptionFacilityBookTimeStruct struct {
	ID          int64  `json:"id" db:"facility_book_time_id"`
	ReceptionID int64  `json:"receptionID" db:"reception_id"`
	FacilityID  int64  `json:"facilityID" db:"facility_id"`
	BookID      int64  `json:"bookID" db:"book_id"`
	SortOrder   int    `json:"sortOrder" db:"sort_order"`
	BookStart   string `json:"bookStart" db:"book_start"`
	BookEnd     string `json:"bookEnd" db:"book_end"`
}

// ReceptionFacilityCodeStruct represents a currently active facility code.
type ReceptionFacilityCodeStruct struct {
	CurrentCode string    `json:"currentCode" db:"current_code"`
	ReceptionID int64     `json:"receptionID" db:"reception_id"`
	FacilityID  int64     `json:"facilityID" db:"facility_id"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty" db:"updated_at"`
}

// ReceptionQueueStruct is a queue transaction.
type ReceptionQueueStruct struct {
	QueueID      int64  `json:"queueID" db:"queue_id"`
	ReceptionID  int64  `json:"receptionID" db:"reception_id"`
	WaitingGuest int    `json:"waitingGuest" db:"waiting_guest"`
	QueueName    string `json:"queueName,omitempty" db:"queue_name"`
	QueuedAt     string `json:"queuedAt,omitempty" db:"queued_at"`
	UserID       int64  `json:"userID,omitempty" db:"user_id"`
	SortOrder    int    `json:"sortOrder" db:"sort_order"`
}
