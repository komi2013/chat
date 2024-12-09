package collection

import "go.mongodb.org/mongo-driver/bson/primitive"

type ReceptionStruct struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdminGroup    string             `bson:"admin_group" json:"adminGroup"`
	JoinNames     []string           `bson:"join_names" json:"joinNames"`
	ReceptTitle   string             `bson:"recept_title,omitempty" json:"receptTitle,omitempty"`
	Asks          []string           `bson:"asks,omitempty" json:"asks,omitempty"`
	// Rooms         [][]interface{}    `bson:"rooms,omitempty" json:"rooms,omitempty"`             // [[capacity, room_name], ...]
	AskChild      bool               `bson:"ask_child,omitempty" json:"askChild,omitempty"`      // optional field
	// SpendMinute   int                `bson:"spend_minute,omitempty" json:"spendMinute,omitempty"` // optional field
	WaitRatio     int                `bson:"wait_ratio,omitempty" json:"waitRatio,omitempty"`  // optional field
	Tables        []Table            `bson:"tables,omitempty" json:"tables,omitempty"`
	Subscription  []string           `bson:"subscription,omitempty" json:"subscription,omitempty"`
	Menus         []Menu             `bson:"menus,omitempty" json:"menus,omitempty"`
	ItemDetails   []ItemDetail       `bson:"item_details,omitempty" json:"itemDetails,omitempty"`
}

type Table struct {
	TableName string   `bson:"table_name" json:"tableName"`
	Capacity  int      `bson:"capacity" json:"capacity"`
	Passcodes []string `bson:"passcodes" json:"passcodes"`
	CurrentCode string `bson:"current_code,omitempty" json:"currentCode,omitempty"`
}

// MenuOption represents an option for a menu item
type MenuOption struct {
	ItemID int `bson:"item_id" json:"itemId"`
	Price  int `bson:"price" json:"price"`
}

// Menu represents a menu item with its details
type Menu struct {
	ID                int          `bson:"id" json:"id"`
	MenuName          string       `bson:"menu_name" json:"menuName"`
	Price             int          `bson:"price" json:"price"`
	Items             []int        `bson:"items" json:"items"`
	PaidOptions       [][]int      `bson:"paid_options,omitempty" json:"paidOptions,omitempty"`         // [[item_id, price]]
	FreeOptions       [][]int      `bson:"free_options,omitempty" json:"freeOptions,omitempty"`         // [[item_id, ...]]
	FreeMultiOptions  []int        `bson:"free_multi_options,omitempty" json:"freeMultiOptions,omitempty"` // [item_id, ...]
}

// ItemDetail represents detailed information about an item
type ItemDetail struct {
	ItemID   int      `bson:"item_id" json:"itemId"`
	ItemName string   `bson:"item_name" json:"itemName"`
	ImgPath  string   `bson:"img_path,omitempty" json:"imgPath,omitempty"`
	Choices  []string `bson:"choices,omitempty" json:"choices,omitempty"`
}

// Room represents a room with its capacity and name
// type Room struct {
// 	Capacity int    `bson:"capacity" json:"capacity"`
// 	RoomName string `bson:"room_name" json:"roomName"`
// }

