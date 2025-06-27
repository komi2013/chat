package collection

import (
  "time"
)

type AdStruct struct {
	PathBanner string  `bson:"path_banner" json:"pathBanner"`
	PathSquare string  `bson:"path_square" json:"pathSquare"`
	AdText     string  `bson:"ad_text" json:"adText"`
	AdLink     string  `bson:"ad_link" json:"adLink"`
	AdStart    int     `bson:"ad_start" json:"adStart"`    // 例: 100（日曜0時）
	AdEnd      int     `bson:"ad_end" json:"adEnd"`        // 例: 223（月曜23時）
	UserID     string  `bson:"user_id" json:"userID"`
	Latitude   float64 `bson:"latitude" json:"latitude"`   // 例: 35.73
	Longitude  float64 `bson:"longitude" json:"longitude"` // 例: 139.53
	Distance   int     `bson:"distance" json:"distance"`   // 半径のスケール
	AdYen      int     `bson:"ad_yen,omitempty" json:"adYen,omitempty"`
	UpdatedAt  time.Time    `bson:"updated_at"`
	ActiveFlag bool    `bson:"active_flag,omitempty" json:"activeFlag,omitempty"`
}

type AdPriceStruct struct {
	LatitudeNorth  float64 `bson:"latitude_north" json:"latitudeNorth"`   // 例: 35.74
	LatitudeSouth  float64 `bson:"latitude_south" json:"latitudeSouth"`   // 例: 35.72
	LongitudeEast  float64 `bson:"longitude_east" json:"longitudeEast"`   // 例: 139.54
	LongitudeWest  float64 `bson:"longitude_west" json:"longitudeWest"`   // 例: 139.52
	AdStart        int     `bson:"ad_start" json:"adStart"`               // 例: 000
	AdEnd          int     `bson:"ad_end" json:"adEnd"`                   // 例: 059
	AdPriceYen     int     `bson:"ad_price_yen" json:"adPriceYen"`
	UpdatedAt      time.Time    `bson:"updated_at"`
}

// 近畿の範囲
// 34.76625859516642, 134.82831731136247
// 34.73466468313902, 136.76191111565907
// 35.726272800793296, 135.6907440933925
// 33.648742984368326, 135.8500458556783

// 3364 ~ 3476
// 13482 ~ 13676