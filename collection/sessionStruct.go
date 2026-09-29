package collection

import (
  "time"
)

// 0=none, 1=VAPID JSON, 2=FCM token, 3=APNs reserved.

type SessionStruct struct {
  SessionID  string    `bson:"_id,omitempty"`
  SessionIDRotated bool `bson:"-" json:"-"`
  UserID  string    `bson:"userID,omitempty"`
  CreatedAt  time.Time `bson:"createdAt,omitempty"`
  UpdatedAt  time.Time `bson:"updatedAt,omitempty"`
  // AliasArray  [][]string      `bson:"aliasArray,omitempty"`
  ChannelAliases []ChannelAlias `bson:"channelAliases,omitempty" json:"channelAliases,omitempty"`
  // PushToken is the single registered push destination for this session:
  // web subscription JSON (1), an FCM token (2) or an APNs token (3).
  PushToken  string `bson:"pushToken,omitempty" json:"-"`
  // Doubles as the web-vs-native session marker: >= 2 came from a native login.
  DeviceType int    `bson:"deviceType,omitempty" json:"-"`
  Csrf  string      `bson:"csrf,omitempty"`
  PushContents []string `bson:"pushContents"`
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
