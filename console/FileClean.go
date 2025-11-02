package console

import (
	"context"
	// "log"
	"os"
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

	var fileCleanLog = common.NewDailyLogger("file_clean_log_")
	var fileCleanError = common.NewDailyLogger("file_clean_error_")
	filter := bson.M{
		"updatedAt": bson.M{"$lt": now.Add(-1 * time.Hour)},
	}
	cursor, err := coll.Find(ctx, filter, options.Find())
	if err != nil {
		fileCleanError.Printf("⚠️ MongoDB find error: %v", err)
		return
	}
	defer cursor.Close(ctx)
	var files []collection.FileStruct
	if err = cursor.All(ctx, &files); err != nil {
		fileCleanError.Printf("⚠️ MongoDB cursor decode error: %v", err)
		return
	}
	deletedCount := 0
	for _, f := range files {
		var limit time.Duration
		if f.ChannelID == "-tweet-" {
			limit = 1 * time.Hour
		} else {
			limit = 24 * time.Hour
		}
		if now.Sub(f.UpdatedAt) < limit {
			continue
		}

		switch f.FileType {
		case 1:
			if err := os.Remove(f.FilePath); err != nil {
				if os.IsNotExist(err) {
					fileCleanError.Printf("⚠️ Not found (img): %s", f.FilePath)
				} else {
					fileCleanError.Printf("⚠️ Delete error (img): %v", err)
				}
			} else {
				fileCleanLog.Printf("🗑️ Deleted image: %s", f.FilePath)
			}
		case 2:
			// UploadDir + "/upload_data/file/" + channelID + "/" + fileID + "/"
			// saveDir := filepath.Join(common.UploadDir, "upload_data", "file", f.ChannelID, f.FileID)
			if err := os.Remove(f.FilePath); err != nil {
				if os.IsNotExist(err) {
					fileCleanError.Printf("⚠️ Not found (img): %s", f.FilePath)
				} else {
					fileCleanError.Printf("⚠️ Delete error (img): %v", err)
				}
			} else {
				fileCleanLog.Printf("🗑️ Deleted image: %s", f.FilePath)
			}
		case 3:
			continue

		default:
			fileCleanError.Printf("⚠️ Unknown FileType=%d (FileID=%s)", f.FileType, f.FileID)
			continue
		}

		if f.FileType != 3 {
			_, err := coll.DeleteOne(ctx, bson.M{"_id": f.FileID})
			if err != nil {
				fileCleanError.Printf("⚠️ Failed to delete MongoDB doc for %s: %v", f.FileID, err)
			} else {
				deletedCount++
			}			
		}
	}
	fileCleanLog.Printf("✅ File cleanup completed. Deleted %d documents.", deletedCount)
}
