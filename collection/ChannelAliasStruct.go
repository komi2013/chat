package collection

type ChannelAlias struct {
    ChannelID string `bson:"channel_id" json:"channel_id"`
    Alias     string `bson:"alias" json:"alias"`
}
