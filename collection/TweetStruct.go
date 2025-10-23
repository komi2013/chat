package collection

import "time"

// TweetStruct : MongoDBに格納される1つのTweetドキュメント
type TweetStruct struct {
  ID         string       `bson:"_id,omitempty" json:"id,omitempty"`
  Tweets     []Tweet      `bson:"tweets" json:"tweets"`
  TweetHeads []TweetHead  `bson:"tweetHeads" json:"tweetHeads"`
}

type Tweet struct {
  MessageID   string   `bson:"messageID" json:"messageID"`
  ParentID    string   `bson:"parentID" json:"parentID"`
  MessageTxt  string   `bson:"messageTxt" json:"messageTxt"`
  Nickname     string    `bson:"nickname,omitempty" json:"nickname,omitempty"`
  NickImg     string    `bson:"nickImg,omitempty" json:"nickImg,omitempty"`
  UserID 			 string    `bson:"userID,omitempty" json:"userID,omitempty"`
  CreatedAt   string   `bson:"createdAt" json:"createdAt"`
  BackID      string   `bson:"backID" json:"backID"`
  Emojis      []Emoji  `bson:"emojis" json:"emojis"`
  TweetCount int      `bson:"tweetCount,omitempty" json:"tweetCount,omitempty"`
  Reply       bool     `bson:"reply,omitempty" json:"reply,omitempty"`
}

type Emoji struct {
	AliasName string `bson:"aliasName" json:"aliasName"`
	Emoji     string `bson:"emoji" json:"emoji"`
}

// TweetHead : ツイートスレッドのヘッダ（旧threadHeadのvalue部分）
type TweetHead struct {
  ParentID      string    `bson:"parentID" json:"parentID"`
  MessageTxt    string    `bson:"messageTxt" json:"messageTxt"`
  Nickname     string    `bson:"nickname,omitempty" json:"nickname,omitempty"`
  NickImg     string    `bson:"nickImg,omitempty" json:"nickImg,omitempty"`
  UpdatedAt     time.Time `bson:"updatedAt" json:"updatedAt"`
  Title         string    `bson:"title" json:"title"`
  NickNames    []string  `bson:"nickNames" json:"nickNames"`
  BackID        string    `bson:"backID,omitempty" json:"backID,omitempty"`
  CreatedAt     string    `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
  Emojis        []Emoji  `bson:"emojis,omitempty" json:"emojis,omitempty"`
  MessageID     string    `bson:"messageID,omitempty" json:"messageID,omitempty"`
}
