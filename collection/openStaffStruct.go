package collection

import (
  "time"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

type OpenStaffStruct struct {
  ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
  WindowID  string             `bson:"window_id" json:"window_id"`
  AliasName string             `bson:"alias_name" json:"alias_name"`
  OpenStart time.Time          `bson:"open_start" json:"open_start"`
  OpenEnd   time.Time          `bson:"open_end" json:"open_end"`
  Role      string             `bson:"role" json:"role"`
}
