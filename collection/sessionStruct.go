package collection

import (
  "time"
)

type SessionStruct struct {
  SessionID  string    `bson:"_id,omitempty"`
  UserID  string    `bson:"user_id,omitempty"`
  CreatedAt  time.Time `bson:"created_at,omitempty"`
  AliasNames  string      `bson:"alias_names,omitempty"`
  Subscription  bool      `bson:"subscription,omitempty"`
}

// session_id
// user_id
// created_at
// alias_names
// subscription