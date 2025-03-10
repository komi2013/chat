package collection

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Books struct {
	BookStart    string   `bson:"book_start" json:"bookStart"`
	BookEnd      string   `bson:"book_end" json:"bookEnd"`
	Answers      []string `bson:"answers" json:"answers"`
	ServiceID    int      `bson:"service_id,omitempty" json:"serviceID,omitempty"`
}

type WorkStaff struct {
	AliasName  string   `bson:"alias_name" json:"aliasName"`
	WorkStart  string   `bson:"work_start" json:"workStart"`
	WorkEnd    string   `bson:"work_end" json:"workEnd"`
	Skills     []string `bson:"skills,omitempty" json:"skills,omitempty"`
	Seq        int      `bson:"seq" json:"seq"`
  Delete     bool     `json:"delete,omitempty"`
}

type Shift struct {
	Date        string                 `bson:"date" json:"date"`
  ShiftStart string         `bson:"shift_start" json:"shiftStart"`
  ShiftEnd   string         `bson:"shift_end" json:"shiftEnd"`
  Open      int                 `bson:"open" json:"open"`
  Role      string             `bson:"role" json:"role"`
  AliasNames []string             `bson:"alias_names" json:"aliasNames"`
}

type TimeSlot struct {
	Date        string                 `bson:"date" json:"date"`
	LimitStart  string                 `bson:"limit_start" json:"limitStart"`
	LimitEnd    string                 `bson:"limit_end" json:"limitEnd"`
	WorkStaffs  []WorkStaff            `bson:"work_staffs,omitempty" json:"workStaffs,omitempty"`
	Books       []Books                `bson:"books,omitempty" json:"books,omitempty"`
}

type Service struct {
	ID           int    `bson:"id" json:"id"`
	ServiceName  string `bson:"service_name" json:"serviceName"`
	Price        int    `bson:"price" json:"price"`
	PrepaidPrice int    `bson:"prepaid_price" json:"prepaidPrice"`
	NeedSkill    string `bson:"need_skill,omitempty" json:"needSkill,omitempty"`
	NeedFacility string `bson:"need_facility,omitempty" json:"needFacility,omitempty"`
	SpendMinute  int    `bson:"spend_minute,omitempty" json:"spendMinute,omitempty"`
}

type BookPatternStruct struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdminGroup       string             `bson:"admin_group" json:"adminGroup"`
	JoinNames        []string           `bson:"join_names" json:"joinNames"`
	BookTitle        string             `bson:"book_title,omitempty" json:"bookTitle,omitempty"`
	Asks             []string           `bson:"asks,omitempty" json:"asks,omitempty"`
	AskChoices       [][]string         `bson:"ask_choices,omitempty" json:"askChoices,omitempty"`
	AskMultiChoices  [][]string         `bson:"ask_multi_choices,omitempty" json:"askMultiChoices,omitempty"`
	Facilities       []Facility         `bson:"facilities,omitempty" json:"facilities,omitempty"` // Mixed types require interface{}
	Times            []TimeSlot         `bson:"times,omitempty" json:"times,omitempty"`
	Shifts           []Shift            `bson:"shifts,omitempty" json:"shifts,omitempty"`
	Services         []Service          `bson:"services,omitempty" json:"services,omitempty"`
	Skills           []string           `bson:"skills,omitempty" json:"skills,omitempty"`
}

type Facility struct {
	FacilityCount  int    `bson:"facility_count" json:"facilityCount"`
	FacilityName   string `bson:"facility_name" json:"facilityName"`
}

