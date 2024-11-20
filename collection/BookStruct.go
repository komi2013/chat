package collection

import (
  "go.mongodb.org/mongo-driver/bson/primitive"
)

type BookStruct struct {
  ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
  WindowID     string             `bson:"window_id" json:"windowID"`
  BookStart    string             `bson:"book_start" json:"bookStart"`
  BookEnd      string             `bson:"book_end" json:"bookEnd"`
  Answers      []string             `bson:"answers,omitempty" json:"answers,omitempty"`
  MenuID       int             `bson:"menu_id,omitempty" json:"menuID,omitempty"`
  UseRole      string             `bson:"use_role,omitempty" json:"useRole,omitempty"`
  UseFacility  string             `bson:"use_facility,omitempty" json:"useFacility,omitempty"`
  SpecifyName  string             `bson:"specify_name,omitempty" json:"specifyName,omitempty"`
}
