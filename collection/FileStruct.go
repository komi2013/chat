package collection

import "time"

type FileStruct struct {
  FileID       string    `bson:"_id"`
  ChannelID    string    `bson:"channelID,omitempty"`
  UploadedBy   string    `bson:"uploadedBy,omitempty"`
  ImgPath      string    `bson:"imgPath,omitempty"`
  FileSize     float64   `bson:"fileSize,omitempty"`
  CreatedAt    time.Time `bson:"createdAt,omitempty"`
  AvailableBy  []string  `bson:"availableBy,omitempty"`
}
