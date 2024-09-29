package collection

import (
  "time"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

type Flow struct {
  ID          primitive.ObjectID `bson:"_id,omitempty"`    // MongoDB's object ID
  TicketID    primitive.ObjectID `bson:"ticket_id"`        // Reference to the ticket (foreign key)
  CreatedBy   string             `bson:"created_by"`       // User who created the history entry
  Description string             `bson:"description"`      // Description of the action or change
  CreatedAt   time.Time          `bson:"created_at"`       // Timestamp of when the entry was created
}
