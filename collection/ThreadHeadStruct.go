package collection

import "time"

// ThreadHeadStruct はスレッドのヘッダ情報を表す構造体です。
type ThreadHeadStruct struct {
  ParentID       string        `bson:"parentID,omitempty" json:"parentID,omitempty"`
  BackID         string        `bson:"backID,omitempty" json:"backID,omitempty"`
  MessageTxt     string        `bson:"messageTxt,omitempty" json:"messageTxt,omitempty"`
  AliasName      string        `bson:"aliasName,omitempty" json:"aliasName,omitempty"`
  AliasImg       string        `bson:"aliasImg,omitempty" json:"aliasImg,omitempty"`
  Emojis         []EmojiStruct `bson:"emojis,omitempty" json:"emojis,omitempty"`
  UpdatedAt      time.Time     `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
  ThreadCount    int           `bson:"threadCount,omitempty" json:"threadCount,omitempty"`
  ChannelID      string        `bson:"channelID,omitempty" json:"channelID,omitempty"`
  Bookmark       bool          `bson:"bookmark,omitempty" json:"bookmark,omitempty"`
  Description    string        `bson:"description,omitempty" json:"description,omitempty"`
  DisplayStatus  *int          `bson:"displayStatus,omitempty" json:"displayStatus,omitempty"`
  Title          string        `bson:"title,omitempty" json:"title,omitempty"`
  AliasNames     []string      `bson:"aliasNames,omitempty" json:"aliasNames,omitempty"`
  BroadcastFlag  bool          `bson:"broadcastFlag,omitempty" json:"broadcastFlag,omitempty"`
  AdminNames     []string      `bson:"adminNames,omitempty" json:"adminNames,omitempty"`
  InquiryFlag    bool          `bson:"inquiryFlag,omitempty" json:"inquiryFlag,omitempty"`
  NewThread      bool          `bson:"newThread,omitempty" json:"newThread,omitempty"`
}

// EmojiStruct は threadHead の emojis 内で使用される構造体です。
type EmojiStruct struct {
  AliasName string `bson:"aliasName,omitempty" json:"aliasName,omitempty"`
  Emoji     string `bson:"emoji,omitempty" json:"emoji,omitempty"`
}
