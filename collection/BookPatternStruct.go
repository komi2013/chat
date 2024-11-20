package collection

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Book represents an individual booking entry in the times array
type Book struct {
	BookStart    string   `bson:"book_start" json:"bookStart"`
	BookEnd      string   `bson:"book_end" json:"bookEnd"`
	Answers      []string `bson:"answers" json:"answers"`
	MenuID       int      `bson:"menu_id" json:"menuId"`
	UseRole      string   `bson:"use_role" json:"useRole"`
	UseFacility  string   `bson:"use_facility" json:"useFacility"`
}

// ShiftStaff represents shift details for a staff member within a times entry
type ShiftStaff struct {
	AliasName  string   `bson:"alias_name" json:"aliasName"`
	ShiftStart string   `bson:"shift_start" json:"shiftStart"`
	ShiftEnd   string   `bson:"shift_end" json:"shiftEnd"`
	Skills     []string `bson:"skills,omitempty" json:"skills,omitempty"`
	Seq        int      `bson:"seq" json:"seq"`
}

// TimeSlot represents a time slot in the schedule
type TimeSlot struct {
	Date        string       `bson:"date" json:"date"`
	LimitStart  string       `bson:"limit_start" json:"limitStart"`
	LimitEnd    string       `bson:"limit_end" json:"limitEnd"`
	ShiftStaff  []ShiftStaff `bson:"shift_staff" json:"shiftStaff"`
	Book        []Book       `bson:"book,omitempty" json:"book,omitempty"`
}

// MenuItem represents a menu item with optional role and facility requirements
type MenuItem struct {
	ID           int    `bson:"id" json:"id"`
	Name         string `bson:"name" json:"name"`
	Price        int    `bson:"price" json:"price"`
	NeedRole     string `bson:"need_role,omitempty" json:"needRole,omitempty"`
	NeedFacility string `bson:"need_facility,omitempty" json:"needFacility,omitempty"`
}

// BookPatternStruct represents the overall booking pattern structure
type BookPatternStruct struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdminGroup    string             `bson:"admin_group" json:"adminGroup"`
	JoinNames     []string           `bson:"join_names" json:"joinNames"`
	BookTitle     string             `bson:"book_title" json:"bookTitle"`
	Asks          []string           `bson:"asks" json:"asks"`
	AskChoices    [][]string         `bson:"ask_choices" json:"askChoices"`
	Facilities    []interface{}      `bson:"facilities" json:"facilities"` // Mixed types require interface{}
	Times         []TimeSlot         `bson:"times" json:"times"`
	Menu          []MenuItem         `bson:"menu" json:"menu"`
}
