package collection

import (
	"errors"
  "time"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

// Ticket represents a ticket document in the ticket collection
type TicketLog struct {
  TicketID         primitive.ObjectID   `bson:"ticket_id,omitempty" json:"ticketID"`         
  ChangedTexts     []bson.M               `bson:"changed_texts" json:"changed_texts"`
  CreatedBy        string               `bson:"created_by" json:"createdBy"`         
  CreatedAt        time.Time            `bson:"created_at" json:"createdAt"`        
}

// Main validation function that combines all the individual validators
func ValidateTicketLog(ticketLog TicketLog) error {
	if err := ValidateCreatedBy(ticketLog.CreatedBy); err != nil {
		return err
	}

	if err := ValidateChangedTexts(ticketLog.ChangedTexts); err != nil {
		return err
	}

	if err := ValidateCreatedAt(ticketLog.CreatedAt); err != nil {
		return err
	}

	if err := ValidateTicketID(ticketLog.TicketID); err != nil {
		return err
	}

	return nil
}

// ValidateCreatedBy validates the CreatedBy field of TicketLog
func ValidateCreatedBy(createdBy string) error {
	if len(createdBy) == 0 {
		return errors.New("CreatedBy field cannot be empty")
	}
	if len(createdBy) > 50 {
		return errors.New("CreatedBy field exceeds 50 characters")
	}
	return nil
}

// ValidateChangedTexts validates the ChangedTexts field of TicketLog
func ValidateChangedTexts(changedTexts []bson.M) error {
	if len(changedTexts) == 0 {
		return errors.New("ChangedTexts field cannot be empty")
	}
	for _, text := range changedTexts {
		if title, ok := text["title"].(string); ok {
			if len(title) > 100 {
				return errors.New("Title in ChangedTexts exceeds 100 characters")
			}
		}
	}
	return nil
}

// ValidateCreatedAt validates the CreatedAt field of TicketLog
func ValidateCreatedAt(createdAt time.Time) error {
	// Here you can apply more rules if needed, such as date restrictions
	if createdAt.IsZero() {
		return errors.New("CreatedAt field cannot be zero")
	}
	return nil
}

// ValidateTicketID validates the TicketID field of TicketLog
func ValidateTicketID(ticketID primitive.ObjectID) error {
	if ticketID.IsZero() {
		return errors.New("TicketID is required and cannot be zero")
	}
	return nil
}
