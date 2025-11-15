package collection

import (
  "time"
)

type AdStruct struct {
	AdID string  `bson:"_id" json:"adID"`
	PathBanner string  `bson:"pathBanner" json:"pathBanner"`
	PathSquare string  `bson:"pathSquare" json:"pathSquare"`
	AdText     string  `bson:"adText" json:"adText"`
	AdLink     string  `bson:"adLink" json:"adLink"`
	AdStart    int     `bson:"adStart" json:"adStart"`    // 例: 100（日曜0時）
	AdEnd      int     `bson:"adEnd" json:"adEnd"`        // 例: 223（月曜23時）
	UserID     string  `bson:"userID" json:"userID"`
	Latitude   float64 `bson:"latitude" json:"latitude"`   // 例: 35.73
	Longitude  float64 `bson:"longitude" json:"longitude"` // 例: 139.53
	Distance   int     `bson:"distance" json:"distance"`   // 半径のスケール
	AdYen      int     `bson:"adYen,omitempty" json:"adYen,omitempty"`
	UpdatedAt  time.Time  `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
	PaidAt time.Time  `bson:"paidAt,omitempty" json:"paidAt,omitempty"`
	InvoicedAt time.Time  `bson:"invoicedAt,omitempty" json:"invoicedAt,omitempty"`
	PublishedAt time.Time    `bson:"published,omitempty" json:"published,omitempty"`
	NextInvoicedFlag bool    `bson:"nextInvoicedFlag,omitempty" json:"nextInvoicedFlag,omitempty"` // when no geo ad, running current ad and want ad next week
  CurrentRunFlag bool    `bson:"currentRunFlag,omitempty" json:"currentRunFlag,omitempty"`
}

type AdPriceStruct struct {
	LatitudeNorth  float64 `bson:"latitudeNorth" json:"latitudeNorth"`   // 例: 35.74
	LatitudeSouth  float64 `bson:"latitudeSouth" json:"latitudeSouth"`   // 例: 35.72
	LongitudeEast  float64 `bson:"longitudeEast" json:"longitudeEast"`   // 例: 139.54
	LongitudeWest  float64 `bson:"longitudeWest" json:"longitudeWest"`   // 例: 139.52
	AdStart        int     `bson:"adStart" json:"adStart"`               // 例: 000
	AdEnd          int     `bson:"adEnd" json:"adEnd"`                   // 例: 059
	AdPriceYen     int     `bson:"adPriceYen" json:"adPriceYen"`
	UpdatedAt      time.Time   `bson:"updatedAt" json:"updatedAt"`
	AdYen      int     `bson:"adYen,omitempty" json:"adYen,omitempty"`
}

type AdResponse struct {
	// AdStart    int     `bson:"adStart" json:"adStart"`    // 例: 100（日曜0時）
	// AdEnd      int     `bson:"adEnd" json:"adEnd"`        // 例: 223（月曜23時）
	AdID string  `bson:"_id" json:"advertisementID"`
	PathBanner string  `bson:"pathBanner" json:"pathBanner"`
	PathSquare string  `bson:"pathSquare" json:"pathSquare"`
	AdText     string  `bson:"adText" json:"adText"`
	AdLink     string  `bson:"adLink" json:"adLink"`
	AdYen      int     `bson:"adYen,omitempty" json:"adYen,omitempty"`
	UpdatedAt  time.Time    `bson:"updatedAt" json:"updatedAt,omitempty"`
}

// 近畿の範囲
// 34.76625859516642, 134.82831731136247
// 34.73466468313902, 136.76191111565907
// 35.726272800793296, 135.6907440933925
// 33.648742984368326, 135.8500458556783

// 3364 ~ 3476
// 13482 ~ 13676