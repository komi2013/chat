package table

import "time"

// UserStruct represents a row in the PostgreSQL "user" table.
type UserStruct struct {
	UserID        int64     `json:"userId" db:"user_id"`
	GoogleJWTSub  string    `json:"-" db:"google_jwt_sub"`
	Mail          string    `json:"mail,omitempty" db:"mail"`
	Telephone     string    `json:"telephone,omitempty" db:"telephone"`
	CreatedAt     time.Time `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt,omitempty" db:"updated_at"`
	SignedAt      time.Time `json:"signedAt,omitempty" db:"signed_at"`
	Latitude      float64   `json:"latitude,omitempty" db:"latitude"`
	Longitude     float64   `json:"longitude,omitempty" db:"longitude"`
	WalletAddress string    `json:"walletAddress,omitempty" db:"wallet_address"`
}

type UserResponse struct {
	Mail          string  `bson:"mail,omitempty" json:"mail,omitempty"`
	Telephone     string  `bson:"telephone,omitempty" json:"telephone,omitempty"`
	WalletAddress string  `bson:"walletAddress" json:"walletAddress"`
	Latitude      float64 `bson:"latitude" json:"latitude"`   // 例: 35.73
	Longitude     float64 `bson:"longitude" json:"longitude"` // 例: 139.53
}
