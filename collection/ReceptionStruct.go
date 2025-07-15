package collection

import (
  "time"

	// "go.mongodb.org/mongo-driver/bson/primitive"
)

type ReceptionStruct struct {
	ReceptionID      string   				  `bson:"_id,omitempty" json:"receptionID"`
  ChannelID        string             `bson:"channelID,omitempty" json:"channelID,omitempty"`
	AdminNames       []string           `bson:"adminNames" json:"adminNames"`
	Passcodes        []Passcode         `bson:"passcodes" json:"passcodes"`
	JoinNames        []string           `bson:"joinNames" json:"joinNames"`
	Subscriptions    []string           `bson:"subscriptions,omitempty" json:"subscriptions,omitempty"`
	UpdatedAt        time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
	ReceptionTitle   string             `bson:"receptionTitle,omitempty" json:"receptionTitle,omitempty"`
	Books            []Book             `bson:"books,omitempty" json:"books,omitempty"`
	Asks             []string           `bson:"asks,omitempty" json:"asks,omitempty"`
	AskChoices       [][]string         `bson:"askChoices,omitempty" json:"askChoices,omitempty"`
	AskMultiChoices  [][]string         `bson:"askMultiChoices,omitempty" json:"askMultiChoices,omitempty"`
	Facilities       []Facility         `bson:"facilities,omitempty" json:"facilities,omitempty"` // Mixed types require interface{}
	OpenTimes        []OpenTime         `bson:"openTimes,omitempty" json:"openTimes,omitempty"`
	Shifts           []Shift            `bson:"shifts,omitempty" json:"shifts,omitempty"`
	Menus            []Menu             `bson:"menus,omitempty" json:"menus,omitempty"`
	Skills           []string           `bson:"skills,omitempty" json:"skills,omitempty"`
	StaffSkills      []StaffSkill       `bson:"staffSkills,omitempty" json:"staffSkills,omitempty"`
	WorkStaffs       []WorkStaff        `bson:"workStaffs,omitempty" json:"workStaffs,omitempty"`
	WorkStaffNeed    bool               `bson:"workStaffNeed,omitempty" json:"workStaffNeed,omitempty"`
	Seats            []Seat             `bson:"seats,omitempty" json:"seats,omitempty"`
	ItemDetails      []ItemDetail       `bson:"itemDetails,omitempty" json:"itemDetails,omitempty"`
	Queues           []Queue            `bson:"queues,omitempty" json:"queues,omitempty"`
	WaitConfigs      []WaitConfig       `bson:"waitConfigs,omitempty" json:"waitConfigs,omitempty"`
}

type Book struct {
	BookStart    string   `bson:"bookStart" json:"bookStart"`
	BookEnd      string   `bson:"bookEnd" json:"bookEnd"`
	Answers      []string `bson:"answers,omitempty" json:"answers"`
	MenuID    int      `bson:"menuID,omitempty" json:"menuID,omitempty"`
	People    int      `bson:"people,omitempty" json:"people,omitempty"`
	CreatedAt time.Time `bson:"createdAt,omitempty"`
}

type WorkStaff struct {
	AliasName  string   `bson:"aliasName" json:"aliasName"`
	WorkStart  string   `bson:"workStart" json:"workStart"`
	WorkEnd    string   `bson:"workEnd" json:"workEnd"`
	Seq        int      `bson:"seq" json:"seq"`
  // Delete     bool     `json:"delete,omitempty"`
}

type Shift struct {
  AliasNames []string             `bson:"aliasNames" json:"aliasNames"`
  ShiftStart string         `bson:"shiftStart" json:"shiftStart"`
  ShiftEnd   string         `bson:"shiftEnd" json:"shiftEnd"`
  Open      int                 `bson:"open" json:"open"`
  Role      string             `bson:"role" json:"role"`
  Fix         bool              `bson:"fix,omitempty" json:"fix,omitempty"`
  // Delete      bool              `bson:"delete,omitempty" json:"delete,omitempty"`
}

type OpenTime struct {
	LimitStart  string                 `bson:"limitStart" json:"limitStart"`
	LimitEnd    string                 `bson:"limitEnd" json:"limitEnd"`
}

type Facility struct {
	FacilityCount  int    `bson:"facilityCount" json:"facilityCount"`
	FacilityName   string `bson:"facilityName" json:"facilityName"`
}

type StaffSkill struct {
	AliasName        string            `bson:"aliasName,omitempty" json:"aliasName,omitempty"`
  Skills           []string          `bson:"skills,omitempty" json:"skills,omitempty"`
}

type Menu struct {
	MenuID            int         		`bson:"menuID" json:"menuID"`
	MenuName        	string      		`bson:"menuName" json:"menuName"`
	Price           	int         		`bson:"price" json:"price"`
	PrepaidPrice 			int    					`bson:"prepaidPrice" json:"prepaidPrice"`
	NeedSkill    			string 					`bson:"needSkill,omitempty" json:"needSkill,omitempty"`
	NeedFacility 			string 					`bson:"needFacility,omitempty" json:"needFacility,omitempty"`
	SpecifyNameFlag   bool           	`bson:"specifyNameFlag" json:"specifyNameFlag"`
	SpendMinute  			int    					`bson:"spendMinute,omitempty" json:"spendMinute,omitempty"`
	Items             []int       		`bson:"items" json:"items"`
	PaidOptions       []ItemOption    `bson:"paidOptions,omitempty" json:"paidOptions,omitempty"`         // [[item_id, price]]
	FreeOptions       [][]int         `bson:"freeOptions,omitempty" json:"freeOptions,omitempty"`         // [[item_id, ...]]
	FreeMultiOptions  []int       		`bson:"freeMultiOptions,omitempty" json:"freeMultiOptions,omitempty"` // [item_id, ...]
}

type ItemDetail struct {
	ItemID   int      `bson:"itemID" json:"itemID"`
	ItemName string   `bson:"itemName" json:"itemName"`
	ImgPath  string   `bson:"imgPath,omitempty" json:"imgPath,omitempty"`
	Choices  [][]string `bson:"choices,omitempty" json:"choices,omitempty"` // choices: [['硬い','普通','柔らかい'],['油多め','普通','油少なめ']]
}

type ItemOption struct {
	ItemID    int      `bson:"itemID" json:"itemID"`
	Price     int      `bson:"price" json:"price"`
}

type Seat struct {
	SeatName string   `bson:"seatName" json:"seatName"`
	Capacity  int      `bson:"capacity" json:"capacity"`
	Passcodes []Passcode `bson:"passcodes" json:"passcodes"`
	CurrentCode string `bson:"currentCode,omitempty" json:"currentCode,omitempty"`
}

type Queue struct {
	WaitingGuest  int      `bson:"waitingGuest,omitempty" json:"waitingGuest,omitempty"`
	QueueName 		string   `bson:"queueName,omitempty" json:"queueName,omitempty"`
	QueuedAt     string   `bson:"queuedAt,omitempty" json:"queuedAt,omitempty"`
}

type WaitConfig struct {
	GuestRange    [2]int      `bson:"guestRange,omitempty" json:"guestRange,omitempty"`
	WaitRatio     int   `bson:"waitRatio,omitempty" json:"waitRatio,omitempty"`
}

type Passcode struct {
	Passkey        string   `bson:"passkey" json:"passkey"`
	UsageLimit  int      `bson:"usageLimit" json:"usageLimit"`
	PassStart   string   `bson:"passStart" json:"passStart"`
	PassEnd   string   `bson:"passEnd" json:"passEnd"`
}