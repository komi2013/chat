package common

import (
	"context"
  "encoding/base64"
  // "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
  // "log"
  "math"
  "net/http"
  "os"
  "path/filepath"
  // "runtime"
  "regexp"
  "strings"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"

)

func EmojiImgValid (imgStr string) (bool) {
	pattern := regexp.MustCompile(`^,.{1,2},#[0-9a-fA-F]{6}$`)
	return pattern.MatchString(imgStr)
	// if !pattern.MatchString(imgStr) {
	// 	common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "画像が不正", http.StatusOK)
}

func ImgSave(img string, userID string, name string, channelID string, fileIDLength int, usageType int) (string, error) {
  cfg := LoadConfig()

	imgPath := img
	// /img/user/seijiro/seijiro_kom1.png
	if strings.HasPrefix(img, "data:image") {
		base64Data := strings.Split(img, ",")[1]
		imageData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			LogError("failed to decode base64", err)
			return "", fmt.Errorf("failed to decode base64: %w", err)
		}
		nameTail := StringRand(fileIDLength)
		// fileID, err := CountUpID("fileID")
		// if err != nil {
		// 	return "", fmt.Errorf("fileID CountUpID: %w", err)
		// }
    if channelID == "" {
      channelID = "-"
    }
    fileName := name + nameTail
		dirPath := cfg.OSImgDir + "/img/" + channelID + "/"
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			LogError("failed to create directory", err)
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
		imgPath = cfg.PublicImgPath + "/img/" + channelID + "/" + fileName + ".png"
		filePath := dirPath + fileName + ".png"
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
		coll := DB.FileDB.Collection("file")
		fileDocument := collection.FileStruct{
			FilePath:     filePath,
			ChannelID:  channelID,
			UploadedBy: userID,
			// FilePath:   filePath,
			PublicPath:  imgPath + "?" + time.Now().Format("0102150405"),
			FileSize:   fileSizeMB,
			UpdatedAt:  time.Now(),
			UsageType: usageType,
		}
		filter := bson.M{"_id": filePath}
		update := bson.M{"$set": fileDocument}
		opts := options.Update().SetUpsert(true)
		_, err = coll.UpdateOne(context.TODO(), filter, update, opts)
		if err != nil {
	    LogError("upsert file", err)
	    return "", fmt.Errorf("upsert file: %w", err)
		}
	} else if !EmojiImgValid(img) && img != "" {
		return "", fmt.Errorf("Emoji invalid:")
	}
	return imgPath, nil
}

func FileSave(r *http.Request, channelID string, uploadedBy string, userIDs []string, usageType int) ([]string, error) {
  cfg := LoadConfig()
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

	// coll := db1.Collection("file")
	coll := DB.FileDB.Collection("file")
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
			LogError("failed to open file", err)
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		// fileID := StringRand(4)
		// fileID, err := CountUpID("fileID")
		// if err != nil {
		// 	return nil, fmt.Errorf("fileID CountUpID: %w", err)
		// }
		filePath := fmt.Sprintf("/upload/file/%s/%s", channelID, fileHeader.Filename)
		saveDir := fmt.Sprintf(cfg.UploadDir + "/upload_data/file/%s/", channelID)

		if err := os.MkdirAll(saveDir, 0755); err != nil {
			LogError("Failed to create directory", err)
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		dst, err := os.Create(filepath.Join(saveDir, fileHeader.Filename))
		if err != nil {
			LogError("Failed to create file", err)
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			LogError("Failed to copy file", err)
			return nil, fmt.Errorf("failed to copy file: %w", err)
		}

		fileSizeMB := float64(fileHeader.Size) / (1024 * 1024)
		fileSizeMB = math.Floor(fileSizeMB*100) / 100
		fileLinks = append(fileLinks, filePath)

		fileDoc := collection.FileStruct{
			FilePath:      saveDir + fileHeader.Filename,
			ChannelID:   channelID,
			UploadedBy:  uploadedBy,
			// FilePath:     saveDir + fileHeader.Filename,
			PublicPath:  filePath,
			FileSize:    fileSizeMB,
			UpdatedAt:   time.Now(),
			AvailableBy: userIDs,
			UsageType: usageType,
		}

		opts := options.Update().SetUpsert(true)
		filter := bson.M{"_id": fileDoc.FilePath}
		update := bson.M{
		    "$set": fileDoc,
		}
		_, err = coll.UpdateOne(context.TODO(), filter, update, opts)
		if err != nil {
		    return nil, fmt.Errorf("failed to upsert document into MongoDB: %w", err)
		}
	}

	return fileLinks, nil
}

