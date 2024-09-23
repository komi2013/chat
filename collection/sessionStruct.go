package collection

import (
  "time"
)

type SessionStruct struct {
  SessionID  string    `bson:"_id,omitempty"`
  UserID  string    `bson:"user_id,omitempty"`
  CreatedAt  time.Time `bson:"created_at,omitempty"`
  AliasArray  [][]string      `bson:"alias_array,omitempty"`
  Subscription  string      `bson:"subscription,omitempty"`
}

// session_id
// user_id
// created_at
// alias_names
// subscription

// [
// 	['sei1','hQKP'],
// 	['komi1','d']
// ]
