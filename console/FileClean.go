package console

import (
	"context"
	"os"
	"time"

	"chat/common"
	"chat/table"
)

// FileClean - 古いファイルを削除する
func FileClean() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()

	fileCleanLog := common.NewDailyLogger("file_clean_log_")
	fileCleanError := common.NewDailyLogger("file_clean_error_")

	// PostgreSQLから削除対象のファイル情報を取得する。
	// usage_type:
	// 0 = contentsPush
	// 4 = tweet content
	//
	// まず更新から1時間以上経過したファイルを取得し、
	// 用途ごとの保存期間はGo側で判定する。
	query := `
		SELECT file_path, updated_at, usage_type
		FROM file
		WHERE updated_at < $1
		  AND usage_type IN (0, 4)
	`

	rows, err := common.DB.FileDB.QueryContext(
		ctx,
		query,
		now.Add(-1*time.Hour),
	)
	if err != nil {
		fileCleanError.Printf(
			"PostgreSQL query error: %v",
			err,
		)
		return
	}
	defer rows.Close()

	var files []table.FileStruct

	for rows.Next() {
		var f table.FileStruct

		if err := rows.Scan(
			&f.FilePath,
			&f.UpdatedAt,
			&f.UsageType,
		); err != nil {
			fileCleanError.Printf(
				"PostgreSQL row scan error: %v",
				err,
			)
			return
		}

		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		fileCleanError.Printf(
			"PostgreSQL rows error: %v",
			err,
		)
		return
	}

	deletedCount := 0

	for _, f := range files {
		var limit time.Duration

		switch f.UsageType {
		case 0:
			// contentsPush: 5日
			limit = 5 * 24 * time.Hour

		case 4:
			// tweet content: 1時間
			limit = time.Hour

		default:
			continue
		}

		// 用途ごとの保存期間を超えていなければ削除しない。
		if now.Sub(f.UpdatedAt) < limit {
			continue
		}

		// ファイルを削除する。
		if err := os.Remove(f.FilePath); err != nil {
			if os.IsNotExist(err) {
				fileCleanError.Printf(
					"File not found: %s",
					f.FilePath,
				)
			} else {
				fileCleanError.Printf(
					"File delete error: path=%s error=%v",
					f.FilePath,
					err,
				)
			}

			continue
		}

		deletedCount++

		fileCleanLog.Printf(
			"Deleted file: %s",
			f.FilePath,
		)
	}

	fileCleanLog.Printf(
		"File cleanup completed. Deleted %d files.",
		deletedCount,
	)
}
