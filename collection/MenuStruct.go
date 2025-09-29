package collection

// booking > reception > QR code at table, open > menu > order

import (
  "time"

  // "go.mongodb.org/mongo-driver/bson/primitive"
)


type MenuStruct struct {
  ReceptionID string     `bson:"-_id" json:"receptionID,omitempty"`
  Menus       []Menu     `bson:"menus" json:"menus"`
  ItemDetails []ItemDetail `bson:"itemDetails" json:"itemDetails"`
  UpdatedAt   time.Time  `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

type Menu struct {
  MenuID            int             `bson:"menuID" json:"menuID"`
  MenuName          string          `bson:"menuName" json:"menuName"`
  Price             int             `bson:"price" json:"price"`
  PrepaidPrice      int             `bson:"prepaidPrice" json:"prepaidPrice"`
  NeedSkill         string          `bson:"needSkill,omitempty" json:"needSkill,omitempty"`
  NeedFacility      string          `bson:"needFacility,omitempty" json:"needFacility,omitempty"`
  SpecifyNameFlag   bool            `bson:"specifyNameFlag" json:"specifyNameFlag"`
  SpendMinute       int             `bson:"spendMinute,omitempty" json:"spendMinute,omitempty"`
  Items             []int           `bson:"items" json:"items"`
  PaidOptions       []ItemOption    `bson:"paidOptions,omitempty" json:"paidOptions"`         // [[item_id, price]]
  FreeOptions       [][]int         `bson:"freeOptions,omitempty" json:"freeOptions"`         // [[item_id, ...]]
  FreeMultiOptions  []int           `bson:"freeMultiOptions,omitempty" json:"freeMultiOptions"` // [item_id, ...]
  Bookable          bool `bson:"bookable,omitempty" json:"bookable,omitempty"`
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
