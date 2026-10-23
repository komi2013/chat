package table

import (
	// "encoding/json"
	"time"
)

// ReceptionStruct is the single-row reception master. The nested master data
// columns are JSONB arrays; marshal/unmarshal those values in repository code.
type ReceptionStruct struct {
	ReceptionID    int64     `json:"receptionID" db:"reception_id"`
	ChannelID      int64     `json:"channelID" db:"channel_id"`
	AdminNames     []string  `json:"adminNames" db:"admin_names"`
	JoinNames      []string  `json:"joinNames" db:"join_names"`
	InquiryNames   []string  `json:"inquiryNames,omitempty" db:"inquiry_names"`
	OrderUserIDs   []int64   `json:"orderUserIDs,omitempty" db:"order_user_ids"`
	UpdatedAt      time.Time `json:"updatedAt,omitempty" db:"updated_at"`
	ReceptionTitle string    `json:"receptionTitle,omitempty" db:"reception_title"`
	Skills         []string  `json:"skills,omitempty" db:"skills"`
	WorkStaffNeed  bool      `json:"workStaffNeed" db:"work_staff_need"`

	Passcodes       []Passcode       `json:"passcodes" db:"passcodes"`
	Asks            []Ask            `json:"asks" db:"asks"`
	AskChoices      []AskChoice      `json:"askChoices" db:"ask_choices"`
	AskMultiChoices []AskMultiChoice `json:"askMultiChoices" db:"ask_multi_choices"`
	Facilities      []FacilityMaster `json:"facilities" db:"facilities"`
	OpenTimes       []OpenTime       `json:"openTimes" db:"open_times"`
	Shifts          []Shift          `json:"shifts" db:"shifts"`
	StaffSkills     []StaffSkill     `json:"staffSkills" db:"staff_skills"`
	WorkStaffs      []WorkStaff      `json:"workStaffs" db:"work_staffs"`
	WaitConfigs     []WaitConfig     `json:"waitConfigs" db:"wait_configs"`
	Menus           []Menu           `json:"menus" db:"menus"`
	ItemDetails     []ItemDetail     `json:"itemDetails" db:"item_details"`
}

// Passcode is stored inside reception.passcodes or a facility master object.
type Passcode struct {
	Passkey   string `json:"passkey"`
	PassStart string `json:"passStart,omitempty"`
	PassEnd   string `json:"passEnd,omitempty"`
}

type Ask struct {
	Question string `json:"question"`
	Sequence int    `json:"sequence"`
}

type AskChoice struct {
	Question string   `json:"question"`
	Choices  []string `json:"choices"`
	Sequence int      `json:"sequence"`
}

type AskMultiChoice struct {
	Question string   `json:"question"`
	Choices  []string `json:"choices"`
	Sequence int      `json:"sequence"`
}

type FacilityMaster struct {
	FacilityID   int      `json:"facilityID"`
	FacilityName string     `json:"facilityName"`
	Capacity     int        `json:"capacity"`
	Passcodes    []Passcode `json:"passcodes,omitempty"`
	Bookable     bool       `json:"bookable,omitempty"`
}

// FacilityView is the response shape after active state is combined from the
// code registry and facility-book-time transaction table. Do not write this
// type directly into reception.facilities JSONB.
type FacilityView struct {
	FacilityID   int      `json:"facilityID"`
	FacilityName string     `json:"facilityName"`
	Capacity     int        `json:"capacity"`
	Passcodes    []Passcode `json:"passcodes,omitempty"`
	Bookable     bool       `json:"bookable,omitempty"`
	CurrentCode  string     `json:"currentCode,omitempty"`
	BookTimes    []BookTime `json:"bookTimes,omitempty"`
}

type BookTime struct {
	BookStart string `json:"bookStart"`
	BookEnd   string `json:"bookEnd"`
}

type OpenTime struct {
	LimitStart string `json:"limitStart"`
	LimitEnd   string `json:"limitEnd"`
}

type Shift struct {
	ShiftID    int64    `json:"shiftID"`
	AliasNames []string `json:"aliasNames"`
	ShiftStart string   `json:"shiftStart"`
	ShiftEnd   string   `json:"shiftEnd"`
	Open       int      `json:"open"`
	Skill      string   `json:"skill"`
	Fix        bool     `json:"fix,omitempty"`
}

type StaffSkill struct {
	AliasName string   `json:"aliasName,omitempty"`
	Skills    []string `json:"skills,omitempty"`
}

type WorkStaff struct {
	ShiftID   int64  `json:"shiftID"`
	AliasName string `json:"aliasName"`
	WorkStart string `json:"workStart"`
	WorkEnd   string `json:"workEnd"`
	Seq       int    `json:"seq"`
}

type WaitConfig struct {
	GuestRange []int `json:"guestRange,omitempty"`
	WaitRatio  int   `json:"waitRatio,omitempty"`
}

type Menu struct {
	MenuID           int64        `json:"menuID"`
	MenuName         string       `json:"menuName"`
	Price            int          `json:"price"`
	PrepaidPrice     int          `json:"prepaidPrice"`
	NeedSkill        string       `json:"needSkill,omitempty"`
	NeedFacility     string       `json:"needFacility,omitempty"`
	SpecifyNameFlag  bool         `json:"specifyNameFlag"`
	SpendMinute      int          `json:"spendMinute,omitempty"`
	Items            []int64      `json:"items"`
	PaidOptions      []ItemOption `json:"paidOptions,omitempty"`
	FreeOptions      [][]int64    `json:"freeOptions,omitempty"`
	FreeMultiOptions []int64      `json:"freeMultiOptions,omitempty"`
	ForBookType      int          `json:"forBookType,omitempty"`
	NeedFacilityID   int          `json:"needFacilityID,omitempty"`
}

type ItemDetail struct {
	ItemID   int64      `json:"itemID"`
	ItemName string     `json:"itemName"`
	ImgPath  string     `json:"imgPath,omitempty"`
	Choices  [][]string `json:"choices,omitempty"`
}

type ItemOption struct {
	ItemID int64 `json:"itemID"`
	Price  int   `json:"price"`
}

