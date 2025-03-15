package collection

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	BookStart    string   `bson:"book_start" json:"bookStart"`
	BookEnd      string   `bson:"book_end" json:"bookEnd"`
	Answers      []string `bson:"answers,omitempty" json:"answers"`
	ServiceID    int      `bson:"service_id,omitempty" json:"serviceID,omitempty"`
}

type WorkStaff struct {
	AliasName  string   `bson:"alias_name" json:"aliasName"`
	WorkStart  string   `bson:"work_start" json:"workStart"`
	WorkEnd    string   `bson:"work_end" json:"workEnd"`
	Seq        int      `bson:"seq" json:"seq"`
  // Delete     bool     `json:"delete,omitempty"`
}

type Shift struct {
  AliasNames []string             `bson:"alias_names" json:"aliasNames"`
  ShiftStart string         `bson:"shift_start" json:"shiftStart"`
  ShiftEnd   string         `bson:"shift_end" json:"shiftEnd"`
  Open      int                 `bson:"open" json:"open"`
  Role      string             `bson:"role" json:"role"`
  Fix         bool              `bson:"fix,omitempty" json:"fix,omitempty"`
  // Delete      bool              `bson:"delete,omitempty" json:"delete,omitempty"`
}

type OpenTime struct {
	LimitStart  string                 `bson:"limit_start" json:"limitStart"`
	LimitEnd    string                 `bson:"limit_end" json:"limitEnd"`
}

type Service struct {
	ServiceID    int    `bson:"service_id" json:"serviceID"`
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
	Books            []Book             `bson:"books,omitempty" json:"books,omitempty"`
	Asks             []string           `bson:"asks,omitempty" json:"asks,omitempty"`
	AskChoices       [][]string         `bson:"ask_choices,omitempty" json:"askChoices,omitempty"`
	AskMultiChoices  [][]string         `bson:"ask_multi_choices,omitempty" json:"askMultiChoices,omitempty"`
	Facilities       []Facility         `bson:"facilities,omitempty" json:"facilities,omitempty"` // Mixed types require interface{}
	OpenTimes        []OpenTime         `bson:"open_times,omitempty" json:"openTimes,omitempty"`
	Shifts           []Shift            `bson:"shifts,omitempty" json:"shifts,omitempty"`
	Services         []Service          `bson:"services,omitempty" json:"services,omitempty"`
	Skills           []string           `bson:"skills,omitempty" json:"skills,omitempty"`
	StaffSkills      []StaffSkill       `bson:"staff_skills,omitempty" json:"staffSkills,omitempty"`
	WorkStaffs       []WorkStaff        `bson:"work_staffs,omitempty" json:"workStaffs,omitempty"`
	WorkStaffNeed    bool               `bson:"work_staff_need,omitempty" json:"workStaffNeed,omitempty"`
}

type Facility struct {
	FacilityCount  int    `bson:"facility_count" json:"facilityCount"`
	FacilityName   string `bson:"facility_name" json:"facilityName"`
}

type StaffSkill struct {
	AliasName        string            `bson:"alias_name,omitempty" json:"aliasName,omitempty"`
  Skills           []string          `bson:"skills,omitempty" json:"skills,omitempty"`
}
