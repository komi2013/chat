package table

// GroupStruct represents a row in the PostgreSQL "group" table.
type GroupStruct struct {
	GroupID    string   `json:"groupId" db:"group_id"`
	ChannelID  int64    `json:"channelId" db:"channel_id"`
	GroupName  string   `json:"groupName,omitempty" db:"group_name"`
	GroupImg   string   `json:"groupImg,omitempty" db:"group_img"`
	GroupBio   string   `json:"groupBio,omitempty" db:"group_bio"`
	AliasNames []string `json:"aliasNames,omitempty" db:"alias_names"`
}
