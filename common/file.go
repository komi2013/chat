package common

import (
	"context"
  "encoding/base64"
  "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "runtime"
  "strings"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"

  "chat/collection"

)

func ImgSave(db1 *mongo.Database, img string, userID string, name string, channelID string) (string, error) {
	imgPath := img

	// 画像がBase64形式か確認
	if strings.HasPrefix(img, "data:image") {
		base64Data := strings.Split(img, ",")[1]
		imageData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			LogError("failed to decode base64", err)
			return "", fmt.Errorf("failed to decode base64: %w", err)
		}

		// 画像保存ディレクトリ作成
		dirPath := "/img/user/" + userID + "/"
		err = os.MkdirAll("."+dirPath, 0755)
		if err != nil {
			LogError("failed to create directory", err)
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		// ファイル保存パスの設定
		imgPath = dirPath + userID + "_" + name + ".png"
		filePath := "." + imgPath

		// 画像ファイルを書き込み
		err = ioutil.WriteFile(filePath, imageData, 0644)
		if err != nil {
			LogError("failed to write file", err)
			return "", fmt.Errorf("failed to write file: %w", err)
		}

		// ファイルサイズ取得
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			LogError("failed to get file info", err)
			return "", fmt.Errorf("failed to get file info: %w", err)
		}
		fileSize := fileInfo.Size()

		// ランダムなファイルIDを生成
		fileID := StringRand(8)

		// MongoDBにドキュメントを挿入
		coll := db1.Collection("file")
		fileDocument := collection.File{
			FileID:     fileID,
			ChannelID:  channelID,
			UploadedBy: userID,
			ImgPath:    imgPath,
			FileSize:   fileSize,
			CreatedAt:  time.Now(),
		}

		_, err = coll.InsertOne(context.TODO(), fileDocument)
		if err != nil {
			LogError("failed to insert document into MongoDB", err)
			return "", fmt.Errorf("failed to insert document into MongoDB: %w", err)
		}
	}

	return imgPath, nil
}

func FileSave(r *http.Request, db1 *mongo.Database, channelID string, uploadedBy string) ([]string, error) {
	var fileLinks []string
	files := r.MultipartForm.File["files[]"]

	// コレクションの取得
	coll := db1.Collection("files")

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			LogError("Failed to open file", err)
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		// ファイルの保存先パスを設定
		fileID := StringRand(4)
		filePath := fmt.Sprintf("/upload/%s/%s/%s", channelID, fileID, fileHeader.Filename)
		saveDir := fmt.Sprintf("./upload_data/%s/%s/", channelID, fileID)

		// ディレクトリ作成
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			LogError("Failed to create directory", err)
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		// ファイルを保存
		dst, err := os.Create(saveDir + fileHeader.Filename)
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

		// 画像リンクを追加
		fileLinks = append(fileLinks, filePath)

		// MongoDB にドキュメントを挿入
		fileDoc := collection.File{
			FileID:     fileID,
			ChannelID:  channelID,
			UploadedBy: uploadedBy,
			ImgPath:    filePath,
			FileSize:   fileHeader.Size,
			CreatedAt:  time.Now(),
		}

		_, err = coll.InsertOne(context.TODO(), fileDoc)
		if err != nil {
			LogError("Failed to insert document into MongoDB", err)
			return nil, fmt.Errorf("failed to insert document into MongoDB: %w", err)
		}
	}

	return fileLinks, nil
}


func ImgsSaveBK(imgs string, userID string, name string, db1 *mongo.Database) []string {
	var paths []string
  if err := json.Unmarshal([]byte(imgs), &paths); err != nil {
		pc, _, _, _ := runtime.Caller(1)
		funcName := runtime.FuncForPC(pc).Name()
  	log.Printf("paths: %v; From:", err, funcName)
    return []string{}
  }
  var imgPaths []string
  for _, imgPath := range paths {
	  if (strings.HasPrefix(imgPath, "data:image")) {
	    base64Data := strings.Split(imgPath, ",")[1]
	    imageData, err := base64.StdEncoding.DecodeString(base64Data)
	    if err != nil {
				pc, _, _, _ := runtime.Caller(1)
				funcName := runtime.FuncForPC(pc).Name()
	      log.Printf("base64.StdEncoding: %v; From:", err, funcName)
	      return []string{}
	    }

	    dirPath := "/img/" + userID + "/"
	    os.MkdirAll("." + dirPath, 0755)
	    imgPath = dirPath + userID + "_" + name + ".png"
	    filePath := "." + imgPath
	    err = ioutil.WriteFile(filePath, imageData, 0644)
	    if err != nil {
				pc, _, _, _ := runtime.Caller(1)
				funcName := runtime.FuncForPC(pc).Name()
	      log.Printf("ioutil.WriteFile: %v; From:", err, funcName)
	      return []string{}
	    }
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				pc, _, _, _ := runtime.Caller(1)
				funcName := runtime.FuncForPC(pc).Name()
	      log.Printf("os.Stat: %v; From:", err, funcName)
	      return []string{}
			}
			fileSize := fileInfo.Size()

			coll := db1.Collection("file")
			document := bson.M{
		    "user_id": userID,
		    "img_path": imgPath,
		    "file_size": fileSize,
		    "created_at": time.Now().Format("2006-01-02 15:04:05"),
			}
			_, err = coll.InsertOne(context.TODO(), document)
			if err != nil {
				pc, _, _, _ := runtime.Caller(1)
				funcName := runtime.FuncForPC(pc).Name()
				log.Printf("coll.InsertOne: %v; From:", err, funcName)
				return []string{}
			}
	  }
	  imgPaths = append(imgPaths, imgPath)
  }

  return imgPaths
}
