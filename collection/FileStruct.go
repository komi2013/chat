package collection

import "time"

type FileStruct struct {
  FileID       string    `bson:"_id"`
  ChannelID    string    `bson:"channelID,omitempty"`
  UploadedBy   string    `bson:"uploadedBy,omitempty"`
  FilePath      string    `bson:"filePath,omitempty"`
  PublicPath   string    `bson:"publicPath,omitempty"`
  FileSize     float64   `bson:"fileSize,omitempty"`
  UpdatedAt    time.Time `bson:"updatedAt,omitempty"`
  AvailableBy  []string  `bson:"availableBy,omitempty"`
  FileType     int    `bson:"fileType,omitempty"` // 0=other, 1=img, 2=file, 3=icon 
}
