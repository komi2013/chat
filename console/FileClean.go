package console

import (
	"context"
	"log"
	"os"
	// "path/filepath"
	// "strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/collection"
	"chat/common"
)

// FileClean - 古いファイルとMongoDBドキュメントを削除
func FileClean() {
	coll := common.DB.FileDB.Collection("file")
	ctx := context.Background()
	now := time.Now()

	// 1時間以上前のファイルを取得
	filter := bson.M{
		"updatedAt": bson.M{"$lt": now.Add(-1 * time.Hour)},
	}
	cursor, err := coll.Find(ctx, filter, options.Find())
	if err != nil {
		log.Printf("⚠️ MongoDB find error: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var files []collection.FileStruct
	if err = cursor.All(ctx, &files); err != nil {
		log.Printf("⚠️ MongoDB cursor decode error: %v", err)
		return
	}

	deletedCount := 0

	for _, f := range files {
		// ===============================
		// 削除対象時間の制限
		// ===============================
		var limit time.Duration
		if f.ChannelID == "-tweet-" {
			limit = 1 * time.Hour
		} else {
			limit = 24 * time.Hour
		}
		if now.Sub(f.UpdatedAt) < limit {
			continue
		}

		// ===============================
		// FileType別削除処理
		// ===============================
		switch f.FileType {
		case 1:
			// --- 画像（ImgSave） ---
			// PublicImgPath + "/img/" + channelID + "/" + fileName + ".png"
			// const OSImgDir = "/vue/public/data"
			// filePath: '/data/img/-tweet-/mikSjS.png',
			// vue/public/data/img/-tweet-/mikSjS.png
			// imgPath := "." + strings.TrimPrefix(f.FilePath, common.OSImgDir)
			if err := os.Remove(f.FilePath); err != nil {
				if os.IsNotExist(err) {
					log.Printf("⚠️ Not found (img): %s", f.FilePath)
				} else {
					log.Printf("⚠️ Delete error (img): %v", err)
				}
			} else {
				log.Printf("🗑️ Deleted image: %s", f.FilePath)
			}

		case 2:
			// --- 通常ファイル（FileSave） ---
			// UploadDir + "/upload_data/file/" + channelID + "/" + fileID + "/"
			// saveDir := filepath.Join(common.UploadDir, "upload_data", "file", f.ChannelID, f.FileID)
			if err := os.Remove(f.FilePath); err != nil {
				if os.IsNotExist(err) {
					log.Printf("⚠️ Not found (img): %s", f.FilePath)
				} else {
					log.Printf("⚠️ Delete error (img): %v", err)
				}
			} else {
				log.Printf("🗑️ Deleted image: %s", f.FilePath)
			}

		case 3:
			// --- アイコンなど（削除しない） ---
			log.Printf("⏭️ Skip icon (FileID=%s)", f.FileID)
			continue

		default:
			log.Printf("⚠️ Unknown FileType=%d (FileID=%s)", f.FileType, f.FileID)
			continue
		}

		if f.FileType != 3 {
			// ===============================
			// MongoDBドキュメント削除
			// ===============================
			_, err := coll.DeleteOne(ctx, bson.M{"_id": f.FileID})
			if err != nil {
				log.Printf("⚠️ Failed to delete MongoDB doc for %s: %v", f.FileID, err)
			} else {
				deletedCount++
			}			
		}
	}

	log.Printf("✅ File cleanup completed. Deleted %d documents.", deletedCount)
}
