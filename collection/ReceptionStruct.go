package collection

// booking > reception > QR code at table, open > menu > order

import (
  "time"

	// "go.mongodb.org/mongo-driver/bson/primitive"
)

type ReceptionStruct struct {
	ReceptionID      string   				  `bson:"_id" json:"receptionID"`
  ChannelID        string             `bson:"channelID" json:"channelID"`
	AdminNames       []string           `bson:"adminNames" json:"adminNames"`
	Passcodes        []Passcode         `bson:"passcodes" json:"passcodes"`
	JoinNames        []string           `bson:"joinNames" json:"joinNames"`
	Subscriptions    []string           `bson:"subscriptions,omitempty" json:"subscriptions,omitempty"`
	OrderUserIDs     []string           `bson:"orderUserIDs,omitempty" json:"orderUserIDs,omitempty"`
	UpdatedAt        time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
	ReceptionTitle   string             `bson:"receptionTitle,omitempty" json:"receptionTitle,omitempty"`
	Books            []Book             `bson:"books,omitempty" json:"books,omitempty"`
	Asks             []Ask              `bson:"asks,omitempty" json:"asks,omitempty"`
	AskChoices       []AskChoice        `bson:"askChoices,omitempty" json:"askChoices,omitempty"`
	AskMultiChoices  []AskMultiChoice   `bson:"askMultiChoices,omitempty" json:"askMultiChoices,omitempty"`
	Facilities       []Facility         `bson:"facilities,omitempty" json:"facilities,omitempty"` // Mixed types require interface{}
	OpenTimes        []OpenTime         `bson:"openTimes,omitempty" json:"openTimes,omitempty"`
	Shifts           []Shift            `bson:"shifts,omitempty" json:"shifts,omitempty"`
	// Menus            []Menu             `bson:"menus,omitempty" json:"menus,omitempty"`
	Skills           []string           `bson:"skills,omitempty" json:"skills,omitempty"`
	StaffSkills      []StaffSkill       `bson:"staffSkills,omitempty" json:"staffSkills,omitempty"`
	WorkStaffs       []WorkStaff        `bson:"workStaffs,omitempty" json:"workStaffs,omitempty"`
	WorkStaffNeed    bool               `bson:"workStaffNeed,omitempty" json:"workStaffNeed,omitempty"`
	// Seats            []Seat             `bson:"seats,omitempty" json:"seats,omitempty"`
	// ItemDetails      []ItemDetail       `bson:"itemDetails,omitempty" json:"itemDetails,omitempty"`
	Queues           []Queue            `bson:"queues,omitempty" json:"queues,omitempty"`
	WaitConfigs      []WaitConfig       `bson:"waitConfigs,omitempty" json:"waitConfigs,omitempty"`
}

type Facility struct {
	// FacilityCount  int    `bson:"facilityCount" json:"facilityCount"`
	FacilityName   string `bson:"facilityName" json:"facilityName"`
	Capacity  int      `bson:"capacity" json:"capacity"`
	Passcodes []Passcode `bson:"passcodes" json:"passcodes"`
	CurrentCode string `bson:"currentCode,omitempty" json:"currentCode,omitempty"`
	Bookable    bool   `bson:"bookable,omitempty" json:"bookable,omitempty"`
	BookTimes   []BookTime  `bson:"bookTimes,omitempty" json:"bookTimes,omitempty"`
}

type WaitConfig struct {
	GuestRange    [2]int      `bson:"guestRange,omitempty" json:"guestRange,omitempty"`
	WaitRatio     int   `bson:"waitRatio,omitempty" json:"waitRatio,omitempty"` // minutes
}

type OpenTime struct {
	LimitStart  string                 `bson:"limitStart" json:"limitStart"`
	LimitEnd    string                 `bson:"limitEnd" json:"limitEnd"`
}

type AskChoice struct {
	Question   string   `bson:"question" json:"question"`
	Choices    []string `bson:"choices" json:"choices"`
	Sequence   int      `bson:"sequence" json:"sequence"`
}

type AskMultiChoice struct {
	Question   string   `bson:"question" json:"question"`
	Choices    []string `bson:"choices" json:"choices"`
	Sequence   int      `bson:"sequence" json:"sequence"`
}

type Ask struct {
	Question   string   `bson:"question" json:"question"`
	Sequence   int      `bson:"sequence" json:"sequence"`
}

type Passcode struct {
	Passkey        string   `bson:"passkey" json:"passkey"`
	// AvailableUsage  int      `bson:"availableUsage,omitempty" json:"availableUsage,omitempty"`
	PassStart   string   `bson:"passStart,omitempty" json:"passStart,omitempty"`
	PassEnd   string   `bson:"passEnd,omitempty" json:"passEnd,omitempty"`
}

type StaffSkill struct {
	AliasName        string            `bson:"aliasName,omitempty" json:"aliasName,omitempty"`
  Skills           []string          `bson:"skills,omitempty" json:"skills,omitempty"`
}

// ↑ master data, ↓ staff input

type Shift struct {
  AliasNames []string             `bson:"aliasNames" json:"aliasNames"`
  ShiftStart string         `bson:"shiftStart" json:"shiftStart"`
  ShiftEnd   string         `bson:"shiftEnd" json:"shiftEnd"`
  Open      int                 `bson:"open" json:"open"`
  Skill      string             `bson:"skill" json:"skill"`
  Fix         bool              `bson:"fix,omitempty" json:"fix,omitempty"`
  // Delete      bool              `bson:"delete,omitempty" json:"delete,omitempty"`
}

type WorkStaff struct {
	AliasName  string   `bson:"aliasName" json:"aliasName"`
	WorkStart  string   `bson:"workStart" json:"workStart"`
	WorkEnd    string   `bson:"workEnd" json:"workEnd"`
	Seq        int      `bson:"seq" json:"seq"`
  // Delete     bool     `json:"delete,omitempty"`
}

// ↓ customer input

type Book struct {
	BookStart    string   `bson:"bookStart" json:"bookStart"`
	BookEnd      string   `bson:"bookEnd" json:"bookEnd"`
	Answers      []string `bson:"answers,omitempty" json:"answers"`
	MenuID    int      `bson:"menuID,omitempty" json:"menuID,omitempty"`
	People    int      `bson:"people,omitempty" json:"people,omitempty"`
	CreatedAt time.Time `bson:"createdAt,omitempty"`
	Nickname     string   `bson:"nickname" json:"nickname"`
}

type BookTime struct {
	BookStart  string                 `bson:"bookStart" json:"bookStart"`
	BookEnd    string                 `bson:"bookEnd" json:"bookEnd"`
}

type Queue struct {
	WaitingGuest  int      `bson:"waitingGuest,omitempty" json:"waitingGuest,omitempty"`
	QueueName 		string   `bson:"queueName,omitempty" json:"queueName,omitempty"`
	QueuedAt     string   `bson:"queuedAt,omitempty" json:"queuedAt,omitempty"`
	UserID     string   `bson:"userID,omitempty" json:"userID,omitempty"`
}
