package collection

import "time"

// TweetStruct : MongoDBに格納される1つのTweetドキュメント
type TweetStruct struct {
  ID         string       `bson:"_id,omitempty" json:"id,omitempty"`
  Tweets     []Tweet      `bson:"tweets" json:"tweets"`
  TweetHead  TweetHead  `bson:"tweetHead" json:"tweetHead"`
  UpdatedAt     time.Time `bson:"updatedAt" json:"updatedAt"`
}

type Tweet struct {
  MessageID   string   `bson:"messageID" json:"messageID"`
  ParentID    string   `bson:"parentID" json:"parentID"`
  MessageTxt  string   `bson:"messageTxt" json:"messageTxt"`
  Nickname     string    `bson:"nickname,omitempty" json:"nickname,omitempty"`
  NickImg     string    `bson:"nickImg,omitempty" json:"nickImg,omitempty"`
  CreatedAt   string   `bson:"createdAt" json:"createdAt"`
  BackID      string   `bson:"backID" json:"backID"`
  Emojis      []Emoji  `bson:"emojis" json:"emojis"`
  TweetCount int      `bson:"tweetCount,omitempty" json:"tweetCount,omitempty"`
  Reply       bool     `bson:"reply,omitempty" json:"reply,omitempty"`
  UserID 			 string    `bson:"userID,omitempty" json:"userID,omitempty"`
  AnonymousFlag   bool `json:"anonymousFlag,omitempty"`
  HiddenName   string `bson:"hiddenName,omitempty" json:"hiddenName,omitempty"`
}

type Emoji struct {
	AliasName string `bson:"aliasName" json:"aliasName"`
	Emoji     string `bson:"emoji" json:"emoji"`
	UserID 	  string  `bson:"userID,omitempty" json:"userID,omitempty"`
}

// TweetHead : ツイートスレッドのヘッダ（旧threadHeadのvalue部分）
type TweetHead struct {
  ParentID      string    `bson:"parentID" json:"parentID"`
  MessageTxt    string    `bson:"messageTxt" json:"messageTxt"`
  Nickname     string    `bson:"nickname,omitempty" json:"nickname,omitempty"`
  NickImg     string    `bson:"nickImg,omitempty" json:"nickImg,omitempty"`
  // UpdatedAt     time.Time `bson:"updatedAt" json:"updatedAt"`
  Nicknames    []string  `bson:"nicknames" json:"nicknames"`
  BackID        string    `bson:"backID,omitempty" json:"backID,omitempty"`
  CreatedAt     string    `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
  Emojis        []Emoji  `bson:"emojis,omitempty" json:"emojis,omitempty"`
  MessageID     string    `bson:"messageID,omitempty" json:"messageID,omitempty"`
  BlockUserIDs   []string  `bson:"blockUserIDs"`
  UserID 			 string    `bson:"userID,omitempty" json:"userID,omitempty"`
  AnonymousFlag   bool `json:"anonymousFlag,omitempty"`
  HiddenName   string `bson:"hiddenName,omitempty" json:"hiddenName,omitempty"`
  HiddenNames    []string  `bson:"hiddenNames" json:"hiddenNames"`
}

type TweetHeadLite struct {
	ParentID   string `json:"parentID"`
	MessageTxt      string `json:"messageTxt"`
	CreatedAt  string `json:"createdAt,omitempty"`
}
