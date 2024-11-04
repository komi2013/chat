package collection

type ShiftStaffStruct struct {
  ShiftStaffID  string         `bson:"_id,omitempty" json:"shiftStaffID"`
  ChannelID  string             `bson:"channel_id,omitempty" json:"channelID,omitempty"`
  BookPatternID  string             `bson:"book_pattern_id,omitempty" json:"bookPatternID,omitempty"`
  AliasName string             `bson:"alias_name" json:"aliasName"`
  ShiftStart string         `bson:"shift_start" json:"shiftStart"`
  ShiftEnd   string         `bson:"shift_end" json:"shiftEnd"`
  Role      string             `bson:"role" json:"role"`
  Seq      int                 `bson:"seq" json:"seq"`
  Delete     int               `json:"delete,omitempty"`
}
