package collection

import (
  "time"
)

type SessionStruct struct {
  SessionID  string    `bson:"_id,omitempty"`
  UserID  string    `bson:"userID,omitempty"`
  CreatedAt  time.Time `bson:"createdAt,omitempty"`
  UpdatedAt  time.Time `bson:"updatedAt,omitempty"`
  // AliasArray  [][]string      `bson:"aliasArray,omitempty"`
  ChannelAliases []ChannelAlias `bson:"channelAliases,omitempty" json:"channelAliases,omitempty"`
  Subscription  string      `bson:"subscription,omitempty"`
  Csrf  string      `bson:"csrf,omitempty"`
  PushContents []string `bson:"pushContents"`
  IsMobile bool `bson:"isMobile" json:"isMobile"`
  FcmToken string `bson:"fcmToken,omitempty" json:"fcmToken,omitempty"`
  Mail         string    `bson:"mail,omitempty" json:"mail,omitempty"`
  Telephone    string    `bson:"telephone,omitempty" json:"telephone,omitempty"`
  Nickname     string    `bson:"nickname,omitempty" json:"nickname,omitempty"`
  NickImg     string    `bson:"nickImg,omitempty" json:"nickImg,omitempty"`
  TweetPosts  []TweetPost    `bson:"tweetPosts,omitempty" json:"tweetPosts,omitempty"`
}

type ChannelAlias struct {
    ChannelID string `bson:"channelID" json:"channelID"`
    Alias     string `bson:"alias" json:"alias"`
    GuestFlag     bool `bson:"guestFlag,omitempty" json:"guestFlag,omitempty"`
}

type TweetPost struct {
    ParentID       string    `bson:"parentID" json:"parentID"`
    PostCount      int    `bson:"postCount" json:"postCount"`
    PostAdminFlag  bool    `bson:"postAdminFlag" json:"postAdminFlag"`
    PostedAt       time.Time `bson:"postedAt" json:"postedAt"`
}

// type Content struct {
// 	PushID string      `bson:"push_id"`
// 	Data   string `bson:"data"`
// }
