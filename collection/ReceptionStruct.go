package collection

import (
  "time"

	// "go.mongodb.org/mongo-driver/bson/primitive"
)

type ReceptionStruct struct {
	ReceptionID      string   				  `bson:"_id,omitempty" json:"receptionID"`
  ChannelID        string             `bson:"channel_id,omitempty" json:"channelID,omitempty"`
	AdminNames       []string           `bson:"admin_names" json:"adminNames"`
	Passcodes        []Passcode         `bson:"passcodes" json:"passcodes"`
	JoinNames        []string           `bson:"join_names" json:"joinNames"`
	Subscriptions    []string           `bson:"subscriptions,omitempty" json:"subscriptions,omitempty"`
	UpdatedAt        time.Time          `bson:"updated_at,omitempty" json:"updatedAt,omitempty"`
	ReceptionTitle   string             `bson:"reception_title,omitempty" json:"receptionTitle,omitempty"`
	Books            []Book             `bson:"books,omitempty" json:"books,omitempty"`
	Asks             []string           `bson:"asks,omitempty" json:"asks,omitempty"`
	AskChoices       [][]string         `bson:"ask_choices,omitempty" json:"askChoices,omitempty"`
	AskMultiChoices  [][]string         `bson:"ask_multi_choices,omitempty" json:"askMultiChoices,omitempty"`
	Facilities       []Facility         `bson:"facilities,omitempty" json:"facilities,omitempty"` // Mixed types require interface{}
	OpenTimes        []OpenTime         `bson:"open_times,omitempty" json:"openTimes,omitempty"`
	Shifts           []Shift            `bson:"shifts,omitempty" json:"shifts,omitempty"`
	Menus            []Menu             `bson:"menus,omitempty" json:"menus,omitempty"`
	Skills           []string           `bson:"skills,omitempty" json:"skills,omitempty"`
	StaffSkills      []StaffSkill       `bson:"staff_skills,omitempty" json:"staffSkills,omitempty"`
	WorkStaffs       []WorkStaff        `bson:"work_staffs,omitempty" json:"workStaffs,omitempty"`
	WorkStaffNeed    bool               `bson:"work_staff_need,omitempty" json:"workStaffNeed,omitempty"`
	Seats            []Seat             `bson:"seats,omitempty" json:"seats,omitempty"`
	ItemDetails      []ItemDetail       `bson:"item_details,omitempty" json:"itemDetails,omitempty"`
	Queues           []Queue            `bson:"queues,omitempty" json:"queues,omitempty"`
	WaitConfigs      []WaitConfig       `bson:"wait_configs,omitempty" json:"waitConfigs,omitempty"`
}

type Book struct {
	BookStart    string   `bson:"book_start" json:"bookStart"`
	BookEnd      string   `bson:"book_end" json:"bookEnd"`
	Answers      []string `bson:"answers,omitempty" json:"answers"`
	MenuID    int      `bson:"menu_id,omitempty" json:"menuID,omitempty"`
	People    int      `bson:"people,omitempty" json:"people,omitempty"`
	CreatedAt time.Time `bson:"created_at,omitempty" json:"-"`
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

type Facility struct {
	FacilityCount  int    `bson:"facility_count" json:"facilityCount"`
	FacilityName   string `bson:"facility_name" json:"facilityName"`
}

type StaffSkill struct {
	AliasName        string            `bson:"alias_name,omitempty" json:"aliasName,omitempty"`
  Skills           []string          `bson:"skills,omitempty" json:"skills,omitempty"`
}

type Menu struct {
	MenuID            int         		`bson:"menu_id" json:"menuID"`
	MenuName        	string      		`bson:"menu_name" json:"menuName"`
	Price           	int         		`bson:"price" json:"price"`
	PrepaidPrice 			int    					`bson:"prepaid_price" json:"prepaidPrice"`
	NeedSkill    			string 					`bson:"need_skill,omitempty" json:"needSkill,omitempty"`
	NeedFacility 			string 					`bson:"need_facility,omitempty" json:"needFacility,omitempty"`
	SpecifyNameFlag   bool           	`bson:"specify_name_flag" json:"specifyNameFlag"`
	SpendMinute  			int    					`bson:"spend_minute,omitempty" json:"spendMinute,omitempty"`
	Items             []int       		`bson:"items" json:"items"`
	PaidOptions       []ItemOption    `bson:"paid_options,omitempty" json:"paidOptions,omitempty"`         // [[item_id, price]]
	FreeOptions       [][]int         `bson:"free_options,omitempty" json:"freeOptions,omitempty"`         // [[item_id, ...]]
	FreeMultiOptions  []int       		`bson:"free_multi_options,omitempty" json:"freeMultiOptions,omitempty"` // [item_id, ...]
}

type ItemDetail struct {
	ItemID   int      `bson:"item_id" json:"itemID"`
	ItemName string   `bson:"item_name" json:"itemName"`
	ImgPath  string   `bson:"img_path,omitempty" json:"imgPath,omitempty"`
	Choices  [][]string `bson:"choices,omitempty" json:"choices,omitempty"` // choices: [['硬い','普通','柔らかい'],['油多め','普通','油少なめ']]
}

type ItemOption struct {
	ItemID    int      `bson:"item_id" json:"itemID"`
	Price     int      `bson:"price" json:"price"`
}

type Seat struct {
	SeatName string   `bson:"seat_name" json:"seatName"`
	Capacity  int      `bson:"capacity" json:"capacity"`
	Passcodes []Passcode `bson:"passcodes" json:"passcodes"`
	CurrentCode string `bson:"current_code,omitempty" json:"currentCode,omitempty"`
}

type Queue struct {
	WaitingGuest  int      `bson:"waiting_guest,omitempty" json:"waitingGuest,omitempty"`
	QueueName 		string   `bson:"queue_name,omitempty" json:"queueName,omitempty"`
	QueuedAt     string   `bson:"queued_at,omitempty" json:"queuedAt,omitempty"`
}

type WaitConfig struct {
	GuestRange    [2]int      `bson:"guest_range,omitempty" json:"guestRange,omitempty"`
	WaitRatio     int   `bson:"wait_ratio,omitempty" json:"waitRatio,omitempty"`
}

type Passcode struct {
	Passkey        string   `bson:"passkey" json:"passkey"`
	UsageLimit  int      `bson:"usage_limit" json:"usageLimit"`
	PassStart   string   `bson:"pass_start" json:"passStart"`
	PassEnd   string   `bson:"pass_end" json:"passEnd"`
}