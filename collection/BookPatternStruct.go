package collection

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Book represents an individual booking entry in the times array
type Books struct {
	BookStart    string   `bson:"book_start" json:"bookStart"`
	BookEnd      string   `bson:"book_end" json:"bookEnd"`
	Answers      []string `bson:"answers" json:"answers"`
	MenuID       int      `bson:"menu_id" json:"menuID"`
}

type ShiftStaffs struct {
	AliasName  string   `bson:"alias_name" json:"aliasName"`
	ShiftStart string   `bson:"shift_start" json:"shiftStart"`
	ShiftEnd   string   `bson:"shift_end" json:"shiftEnd"`
	Skills     []string `bson:"skills,omitempty" json:"skills,omitempty"`
	Seq        int      `bson:"seq" json:"seq"`
}

// TimeSlot represents a time slot in the schedule
type TimeSlot struct {
	Date        string         `bson:"date" json:"date"`
	LimitStart  string         `bson:"limit_start" json:"limitStart"`
	LimitEnd    string         `bson:"limit_end" json:"limitEnd"`
	ShiftStaffs []ShiftStaffs  `bson:"shift_staffs,omitempty" json:"shiftStaffs,omitempty"`
	Books       []Books        `bson:"books,omitempty" json:"books,omitempty"`
}

type MenuItem struct {
	ID           int    `bson:"id" json:"id"`
	Name         string `bson:"name" json:"name"`
	Price        int    `bson:"price" json:"price"`
	NeedSkill    string `bson:"need_skill,omitempty" json:"needSkill,omitempty"`
	NeedFacility string `bson:"need_facility,omitempty" json:"needFacility,omitempty"`
	SpendMinute  int    `bson:"spend_minute,omitempty" json:"spendMinute,omitempty"`
}

// BookPatternStruct represents the overall booking pattern structure
type BookPatternStruct struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdminGroup    string             `bson:"admin_group" json:"adminGroup"`
	JoinNames     []string           `bson:"join_names" json:"joinNames"`
	BookTitle     string             `bson:"book_title,omitempty" json:"bookTitle,omitempty"`
	Asks          []string           `bson:"asks,omitempty" json:"asks,omitempty"`
	AskChoices    [][]string         `bson:"ask_choices,omitempty" json:"askChoices,omitempty"`
	Facilities    []interface{}      `bson:"facilities,omitempty" json:"facilities,omitempty"` // Mixed types require interface{}
	Times         []TimeSlot         `bson:"times,omitempty" json:"times,omitempty"`
	Menus         []MenuItem         `bson:"menus,omitempty" json:"menus,omitempty"`
}
