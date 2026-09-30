package collection

import (
  "time"
)

// 0=none, 1=VAPID JSON, 2=FCM token, 3=APNs reserved.

type SessionStruct struct {
  SessionID  string    `bson:"_id,omitempty"`
  SessionIDRotated bool `bson:"-" json:"-"`
  // DeviceID はセッションを所有する端末（アプリインストール）を表す識別子。
  // クライアントが生成する UUIDv4 で、認証情報ではない（認証はサーバー生成の
  // SessionID のみで行う）。ログイン時の同一端末セッション置換と、push通知の
  // 「1端末1通」判定に使う。
  DeviceID string    `bson:"deviceID,omitempty" json:"deviceID,omitempty"`
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
