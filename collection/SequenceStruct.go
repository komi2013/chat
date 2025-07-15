package collection

import (
  "time"
)

type SequenceStruct struct {
  SequenceID  string    `bson:"_id"`
  Count  string    `bson:"count"`
  Lock  int    `bson:"lock"`
  Description  string `bson:"description,omitempty"`
  UpdatedAt  time.Time `bson:"updatedAt,omitempty"`
}
