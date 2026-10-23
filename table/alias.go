package table

// AliasStruct represents a row in the alias table.
type AliasStruct struct {
	AliasID     int64  `json:"aliasId" db:"alias_id"`
	ChannelID   int64  `json:"channelId" db:"channel_id"`
	AliasName   string `json:"aliasName,omitempty" db:"alias_name"`
	AliasImg    string `json:"aliasImg,omitempty" db:"alias_img"`
	UserID      int64  `json:"userId,omitempty" db:"user_id"`
	AliasBio    string `json:"aliasBio,omitempty" db:"alias_bio"`
	AccessRight int16  `json:"accessRight,omitempty" db:"access_right"`
}
