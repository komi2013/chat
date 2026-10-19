package common

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/lib/pq"
)

var emojiImgPattern = regexp.MustCompile(`^,.{1,2},#[0-9a-fA-F]{6}$`)

func EmojiImgValid(imgStr string) bool {
	return emojiImgPattern.MatchString(imgStr)
}

// IsStoredImgPath checks whether an image is already stored.
func IsStoredImgPath(img string) bool {
	return strings.HasPrefix(img, "/")
}

func channelFolder(channelID int64) string {
	if channelID <= 0 {
		return "-"
	}
	return strconv.FormatInt(channelID, 10)
}

// Keep only characters safe for a generated image filename.
func safeImageName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))

	var result strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			result.WriteRune(r)
		}
	}

	if result.Len() == 0 {
		return "image"
	}

	return result.String()
}

// upsertFile saves file metadata to PostgreSQL.
// channel_id is nullable; channelID == 0 means NULL.
func upsertFile(
	ctx context.Context,
	filePath string,
	channelID int64,
	uploadedBy int64,
	publicPath string,
	fileSizeMB float64,
	updatedAt time.Time,
	availableBy []string,
	usageType int,
) error {
	if DB == nil || DB.FileDB == nil {
		return fmt.Errorf("file database is not initialized")
	}

	var dbChannelID any
	if channelID > 0 {
		dbChannelID = channelID
	}

	if availableBy == nil {
		availableBy = []string{}
	}

	_, err := DB.FileDB.ExecContext(
		ctx,
		`
		INSERT INTO file (
			file_path,
			channel_id,
			uploaded_by,
			public_path,
			file_size,
			updated_at,
			available_by,
			usage_type
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (file_path)
		DO UPDATE SET
			channel_id = EXCLUDED.channel_id,
			uploaded_by = EXCLUDED.uploaded_by,
			public_path = EXCLUDED.public_path,
			file_size = EXCLUDED.file_size,
			updated_at = EXCLUDED.updated_at,
			available_by = EXCLUDED.available_by,
			usage_type = EXCLUDED.usage_type
		`,
		filePath,
		dbChannelID,
		uploadedBy,
		publicPath,
		fileSizeMB,
		updatedAt,
		pq.Array(availableBy),
		usageType,
	)
	if err != nil {
		return fmt.Errorf("upsert file metadata: %w", err)
	}

	return nil
}

func ImgSave(
	img string,
	userID int64,
	name string,
	channelID int64,
	fileIDLength int,
	usageType int,
) (string, error) {
	if IsStoredImgPath(img) {
		return img, nil
	}

	if img == "" {
		return "", nil
	}

	if !strings.HasPrefix(img, "data:image") {
		if !EmojiImgValid(img) {
			return "", fmt.Errorf("invalid emoji image")
		}
		return img, nil
	}

	_, base64Data, found := strings.Cut(img, ",")
	if !found {
		return "", fmt.Errorf("invalid image data URL")
	}

	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		LogError("failed to decode base64", err)
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	cfg := LoadConfig()

	folder := channelFolder(channelID)
	fileName := safeImageName(name) + StringRand(fileIDLength) + ".png"

	dirPath := filepath.Join(cfg.OSImgDir, "img", folder)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		LogError("failed to create image directory", err)
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	filePath := filepath.Join(dirPath, fileName)
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		LogError("failed to write image", err)
		return "", fmt.Errorf("failed to write image: %w", err)
	}

	publicPath := strings.TrimRight(cfg.PublicImgPath, "/") +
		"/img/" + folder + "/" + fileName

	now := time.Now()
	fileSizeMB := math.Floor(float64(len(imageData))/(1024*1024)*100) / 100

	err = upsertFile(
		context.Background(),
		filePath,
		channelID,
		userID,
		publicPath+"?"+now.Format("0102150405"),
		fileSizeMB,
		now,
		nil,
		usageType,
	)
	if err != nil {
		LogError("upsert image metadata", err)
		return "", err
	}

	return publicPath, nil
}

func FileSave(
	r *http.Request,
	channelID int64,
	uploadedBy int64,
	userIDs []string,
	usageType int,
) ([]string, error) {
	if DB == nil || DB.FileDB == nil {
		return nil, fmt.Errorf("file database is not initialized")
	}

	cfg := LoadConfig()

	const (
		maxFileSize  = 100 << 20 // 100 MB per file
		maxTotalSize = 500 << 20 // 500 MB total
		maxFileCount = 10
	)

	forbiddenExtensions := map[string]bool{
		".exe": true,
		".apk": true,
	}

	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return nil, fmt.Errorf("parse multipart form: %w", err)
		}
	}

	files := r.MultipartForm.File["files[]"]
	if len(files) > maxFileCount {
		return nil, fmt.Errorf(
			"file limit exceeded: maximum %d files allowed",
			maxFileCount,
		)
	}

	folder := channelFolder(channelID)
	saveDir := filepath.Join(
		cfg.UploadDir,
		"upload_data",
		"file",
		folder,
	)

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		LogError("failed to create upload directory", err)
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	var totalSize int64
	var fileLinks []string

	for _, fileHeader := range files {
		if fileHeader.Size > maxFileSize {
			return nil, fmt.Errorf(
				"file %q exceeds maximum size of %d MB",
				fileHeader.Filename,
				maxFileSize/(1<<20),
			)
		}

		totalSize += fileHeader.Size
		if totalSize > maxTotalSize {
			return nil, fmt.Errorf(
				"total upload size exceeds %d MB",
				maxTotalSize/(1<<20),
			)
		}

		// Prevent a supplied filename from escaping the upload directory.
		filename := filepath.Base(
			strings.ReplaceAll(fileHeader.Filename, "\\", "/"),
		)
		if filename == "" || filename == "." || filename == ".." {
			return nil, fmt.Errorf("invalid uploaded filename")
		}

		ext := strings.ToLower(filepath.Ext(filename))
		if forbiddenExtensions[ext] {
			return nil, fmt.Errorf(
				"file %q has a forbidden extension: %s",
				filename,
				ext,
			)
		}

		src, err := fileHeader.Open()
		if err != nil {
			LogError("failed to open uploaded file", err)
			return nil, fmt.Errorf("failed to open uploaded file: %w", err)
		}

		savePath := filepath.Join(saveDir, filename)
		dst, err := os.Create(savePath)
		if err != nil {
			_ = src.Close()
			LogError("failed to create uploaded file", err)
			return nil, fmt.Errorf("failed to create uploaded file: %w", err)
		}

		written, copyErr := io.Copy(dst, src)
		srcCloseErr := src.Close()
		dstCloseErr := dst.Close()

		if copyErr != nil {
			return nil, fmt.Errorf("failed to copy uploaded file: %w", copyErr)
		}
		if srcCloseErr != nil {
			return nil, fmt.Errorf("failed to close uploaded file: %w", srcCloseErr)
		}
		if dstCloseErr != nil {
			return nil, fmt.Errorf("failed to close destination file: %w", dstCloseErr)
		}

		now := time.Now()
		fileSizeMB := math.Floor(float64(written)/(1024*1024)*100) / 100
		publicPath := fmt.Sprintf(
			"/upload/file/%s/%s",
			folder,
			filename,
		)

		err = upsertFile(
			r.Context(),
			savePath,
			channelID,
			uploadedBy,
			publicPath,
			fileSizeMB,
			now,
			userIDs,
			usageType,
		)
		if err != nil {
			return nil, err
		}

		fileLinks = append(fileLinks, publicPath)
	}

	return fileLinks, nil
}