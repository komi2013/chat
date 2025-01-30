package collection

import "time"

type FileStruct struct {
  FileID       string    `bson:"_id"`
  ChannelID    string    `bson:"channel_id,omitempty"`
  UploadedBy   string    `bson:"uploaded_by,omitempty"`
  ImgPath      string    `bson:"img_path,omitempty"`
  FileSize     float64   `bson:"file_size,omitempty"`
  CreatedAt    time.Time `bson:"created_at,omitempty"`
  AvailableBy  []string  `bson:"available_by,omitempty"`
}
