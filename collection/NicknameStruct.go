package collection

import (
  "time"
)

type NicknameStruct struct {
  Nickname     string    `bson:"_id,omitempty" json:"nickname,omitempty"`
  UserID 			 string    `bson:"userID,omitempty" json:"userID,omitempty"`
  NickImg     string    `bson:"nickImg,omitempty" json:"nickImg,omitempty"`
  Good         int    	`bson:"good,omitempty" json:"good,omitempty"`
  Bad   			 int    		`bson:"bad,omitempty" json:"bad,omitempty"`
  CreatedAt    time.Time `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
  UpdatedAt    time.Time `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}
