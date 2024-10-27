package collection

import (
	"errors"
	"log"
	"runtime"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Ticket represents a ticket document in the ticket collection
type Ticket struct {
	TicketID         primitive.ObjectID   `bson:"_id,omitempty" json:"ticketID"`       
	Status           int                  `bson:"status" json:"status"`              
	Title            string               `bson:"title" json:"title"`              
	Assignee         string               `bson:"assignee" json:"assignee"`          
	CreatedBy        string               `bson:"created_by" json:"createdBy"`         
	Description      string               `bson:"description" json:"description"`
	AccessNames      []string             `bson:"access_names" json:"accessNames"`
	// AssigneeOptions  []string             `bson:"assignee_options" json:"assigneeOptions"`
	CreatedAt        time.Time            `bson:"created_at" json:"createdAt"`        
	ChannelID        string               `bson:"channel_id" json:"channelID"`          
   
	UpdatedAt        *time.Time            `bson:"updated_at,omitempty" json:"updatedAt"` 
	Contents         []bson.M               `bson:"contents,omitempty" json:"contents"`  
	ContentsType     int               `bson:"contents_type,omitempty" json:"contentsType"`  
	Progress         int               `bson:"progress,omitempty"`      
	Priority         int               `bson:"priority,omitempty"`      
	StartDate        string               `bson:"start_date,omitempty"` 
	Deadline         string               `bson:"deadline,omitempty"`   
	Category         int               `bson:"category,omitempty"`     
	Approver1        string               `bson:"approver1,omitempty"`    
	Approver2        string               `bson:"approver2,omitempty"`    
	Approver3        string               `bson:"approver3,omitempty"`    
	ParentID         string               `bson:"parent_id,omitempty"`    
	ChildrenIDs      []string             `bson:"children_ids,omitempty"` 
}

func ValidateTicketStatus(status float64) (int, error) {
	if status > 5 {
		return int(status), errors.New("status must be no more than 5")
	}
	return int(status), nil
}

func ValidateTicketTitle(title string) (string, error) {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		// fmt.Println("Could not retrieve caller information")
		return title, errors.New("Could not retrieve caller information")
	}
	fn := runtime.FuncForPC(pc).Name()
	if len(title) < 3 {
		// log.Print("ValidateTicketTitle was called from %s:%d (function: %s)\n", file, line, fn)
		log.Printf("ValidateTicketTitle was called from %s:%d (function: %s)\n", file, line, fn)
		return title, errors.New("Title must be at least 3 characters")
	}
	if len(title) > 50 {
		log.Printf("ValidateTicketTitle was called from %s:%d (function: %s)\n", file, line, fn)
		return title, errors.New("Title must be no more than 50 characters")
	}
	return title, nil
}

func ValidateTicketDescription(description string) (string, error) {
	if len(description) > 5000 {
		return description, errors.New("Description must be no more than 5000 characters")
	}
	return description, nil
}

func ValidateTicketNewComment(newComment string) (string, error) {
	if len(newComment) > 500 {
		return newComment, errors.New("newComment must be no more than 500 characters")
	}
	return newComment, nil
}

func ValidateTicketContentsType(contentsType float64) (int, error) {
	if contentsType > 10 {
		return int(contentsType), errors.New("contentsType must be be no more than 10")
	}
	return int(contentsType), nil
}


// func ValidateTicketContents(contents []bson.M) ([]bson.M, error) {
// 	if len(contents) == 0 {
// 		return contents, errors.New("contents cannot be empty")
// 	}
// 	return contents, nil
// }
// func ValidateTicket(ticket Ticket) error {
// 	if err := ValidateTitle(ticket.Title); err != nil {
// 		return err
// 	}

// 	if err := ValidateDescription(ticket.Description); err != nil {
// 		return err
// 	}

// 	return nil
// }