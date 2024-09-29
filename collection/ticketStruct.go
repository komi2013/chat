package collection

import (
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
