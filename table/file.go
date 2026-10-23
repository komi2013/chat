package table

import "time"

// FileStruct represents a row in the file table.
type FileStruct struct {
	FileID      int64     `json:"fileId" db:"file_id"`
	FilePath    string    `json:"filePath,omitempty" db:"file_path"`
	ChannelID   int64     `json:"channelId,omitempty" db:"channel_id"`
	UploadedBy  int64     `json:"uploadedBy,omitempty" db:"uploaded_by"`
	PublicPath  string    `json:"publicPath,omitempty" db:"public_path"`
	FileSize    float64   `json:"fileSize,omitempty" db:"file_size"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty" db:"updated_at"`
	AvailableBy []int64   `json:"availableBy" db:"available_by"`
	UsageType   int16     `json:"usageType" db:"usage_type"`
}
