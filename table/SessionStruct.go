package table

import "time"

// SessionStruct represents a session stored in PostgreSQL.
//
// DeviceID is a device identifier used only for session counting.
// Authentication is performed using SessionID.
type SessionStruct struct {
  SessionID        string    `json:"sessionId,omitempty" db:"session_id"`
  SessionIDRotated bool      `json:"-"`
  DeviceID         string    `json:"deviceID,omitempty" db:"device_id"`
  UserID           int64     `json:"userID,omitempty" db:"user_id"`
  CreatedAt        time.Time `json:"createdAt,omitempty" db:"created_at"`
  UpdatedAt        time.Time `json:"updatedAt,omitempty" db:"updated_at"`

  PushToken  string   `json:"-" db:"push_token"`
  DeviceType int      `json:"-" db:"device_type"`
  Csrf       string   `json:"csrf,omitempty" db:"csrf"`
  PushContents []string `json:"pushContents" db:"push_contents"`

  Mail      string `json:"mail,omitempty" db:"mail"`
  Telephone string `json:"telephone,omitempty" db:"telephone"`
}