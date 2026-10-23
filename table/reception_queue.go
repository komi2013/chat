package database
type ReceptionQueueStruct struct {
	QueueID int64 `db:"queue_id" json:"queueID"`
	ReceptionID int64 `db:"reception_id" json:"receptionID"`
	SortOrder int `db:"sort_order" json:"sortOrder"`
	WaitingGuest int `db:"waiting_guest" json:"waitingGuest,omitempty"`
	QueueName string `db:"queue_name" json:"queueName,omitempty"`
	QueuedAt string `db:"queued_at" json:"queuedAt,omitempty"`
	UserID int64 `db:"user_id" json:"userID,omitempty"`
}
