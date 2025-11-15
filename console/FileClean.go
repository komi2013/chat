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
		"usageType": bson.M{"$nin": []int{1, 2, 3}},
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
		switch f.UsageType {
		case 0: // contentsPush
			limit = 24 * 5 * time.Hour
		case 4: //tweet
			limit = 1 * time.Hour
		default:
			limit = 24 * time.Hour
			continue
		}
		if now.Sub(f.UpdatedAt) < limit {
			continue
		}
		if err := os.Remove(f.FilePath); err != nil {
			if os.IsNotExist(err) {
				fileCleanError.Printf("⚠️ Not found (img): %s", f.FilePath)
			} else {
				fileCleanError.Printf("⚠️ Delete error (img): %v", err)
			}
		} else {
			fileCleanLog.Printf("🗑️ Deleted image: %s", f.FilePath)
		}
	}
	fileCleanLog.Printf("✅ File cleanup completed. Deleted %d documents.", deletedCount)
}
