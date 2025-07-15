package common

import (
	"context"
  "encoding/base64"
  // "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "math"
  "net/http"
  "os"
  "path/filepath"
  // "runtime"
  "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/bson"

  "chat/collection"

)

func ImgSave(db1 *mongo.Database, img string, userID string, name string, channelID string, fileIDLength int) (string, error) {
	imgPath := img
	// /img/user/seijiro/seijiro_kom1.png
	if strings.HasPrefix(img, "data:image") {
		base64Data := strings.Split(img, ",")[1]
		imageData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			LogError("failed to decode base64", err)
			return "", fmt.Errorf("failed to decode base64: %w", err)
		}
		fileID := StringRand(fileIDLength)
    if channelID == "" {
      channelID = "tweet"
    }
		dirPath := OSImgDir + "/img/" + channelID + "/" + fileID + "/"
		err = os.MkdirAll("."+dirPath, 0755)
		if err != nil {
			LogError("failed to create directory", err)
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
		imgPath = PublicImgPath + "/img/" + channelID + "/" + fileID + "/" + name + ".png"
		filePath := "." + dirPath + name + ".png"
		err = ioutil.WriteFile(filePath, imageData, 0644)
		if err != nil {
			LogError("failed to write file", err)
			return "", fmt.Errorf("failed to write file: %w", err)
		}
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			LogError("failed to get file info", err)
			return "", fmt.Errorf("failed to get file info: %w", err)
		}
		fileSize := fileInfo.Size()
		fileSizeMB := float64(fileSize) / (1024 * 1024)
		fileSizeMB = math.Floor(fileSizeMB*100) / 100
		coll := db1.Collection("file")
		fileDocument := collection.FileStruct{
			FileID:     fileID,
			ChannelID:  channelID,
			UploadedBy: userID,
			ImgPath:    imgPath,
			FileSize:   fileSizeMB,
			CreatedAt:  time.Now(),
		}
		_, err = coll.InsertOne(context.TODO(), fileDocument)
		if err != nil {
			LogError("insert file", err)
			return "", fmt.Errorf("insert file: %w", err)
		}
	}
	return imgPath, nil
}

func FileSave(r *http.Request, db1 *mongo.Database, channelID string, uploadedBy string, userIDs []string) ([]string, error) {
	const (
		maxFileSize      = 100 << 20 // 100MB (1ファイルあたりの上限)
		maxTotalSize     = 500 << 20 // 500MB (全体の上限)
		maxFileCount     = 10        // 最大10ファイル
	)
	allowedExtensions := map[string]bool{".jpg": true, ".png": true, ".txt": true, ".pdf": true}

	var fileLinks []string
	files := r.MultipartForm.File["files[]"]

	if len(files) > maxFileCount {
		return nil, fmt.Errorf("file limit exceeded: maximum %d files allowed", maxFileCount)
	}

	coll := db1.Collection("file")
	var totalSize int64 = 0

	for _, fileHeader := range files {
		if fileHeader.Size > maxFileSize {
			return nil, fmt.Errorf("file '%s' exceeds max size of %dMB", fileHeader.Filename, maxFileSize/(1<<20))
		}

		totalSize += fileHeader.Size
		if totalSize > maxTotalSize {
			return nil, fmt.Errorf("total upload size exceeds %dMB", maxTotalSize/(1<<20))
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename)) // 🔴 修正後も問題なし
		if !allowedExtensions[ext] {
			return nil, fmt.Errorf("file '%s' has an invalid extension: %s", fileHeader.Filename, ext)
		}

		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Failed to open file: %v", err)
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		fileID := StringRand(4)
		filePath := fmt.Sprintf("/upload/file/%s/%s/%s", channelID, fileID, fileHeader.Filename)
		saveDir := fmt.Sprintf(UploadDir + "/upload_data/file/%s/%s/", channelID, fileID)

		if err := os.MkdirAll(saveDir, 0755); err != nil {
			log.Printf("Failed to create directory: %v", err)
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		dst, err := os.Create(filepath.Join(saveDir, fileHeader.Filename))
		if err != nil {
			log.Printf("Failed to create file: %v", err)
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			log.Printf("Failed to copy file: %v", err)
			return nil, fmt.Errorf("failed to copy file: %w", err)
		}

		fileSizeMB := float64(fileHeader.Size) / (1024 * 1024)
		fileSizeMB = math.Floor(fileSizeMB*100) / 100
		fileLinks = append(fileLinks, filePath)

		fileDoc := collection.FileStruct{
			FileID:      fileID,
			ChannelID:   channelID,
			UploadedBy:  uploadedBy,
			ImgPath:     filePath,
			FileSize:    fileSizeMB,
			CreatedAt:   time.Now(),
			AvailableBy: userIDs,
		}

		_, err = coll.InsertOne(context.TODO(), fileDoc)
		if err != nil {
			log.Printf("Failed to insert document into MongoDB: %v", err)
			return nil, fmt.Errorf("failed to insert document into MongoDB: %w", err)
		}
	}

	return fileLinks, nil
}
