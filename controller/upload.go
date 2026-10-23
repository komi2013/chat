package controller

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"chat/common"
)

func Upload(w http.ResponseWriter, r *http.Request) {
	cfg := common.LoadConfig()

	decodedPath, err := url.PathUnescape(r.URL.Path)
	if err != nil {
		log.Printf("[Upload] URL decode error: %v", err)
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	parts := strings.Split(decodedPath, "/")
	if len(parts) != 5 || parts[1] != "upload" {
		log.Printf("[Upload] Invalid URL path: %s", decodedPath)
		http.Error(w, "invalid URL path", http.StatusNotFound)
		return
	}

	fileType := parts[2] // "file" or "img"
	channelID := parts[3]
	fileID := parts[4] // 実ファイル名。日本語OK

	if fileType != "file" && fileType != "img" {
		http.Error(w, "invalid file type", http.StatusBadRequest)
		return
	}

	// channel_id is BIGINT in PostgreSQL.
	channelIDInt, err := strconv.ParseInt(channelID, 10, 64)
	if err != nil || channelIDInt <= 0 {
		http.Error(w, "invalid channel ID", http.StatusBadRequest)
		return
	}

	// Prevent path traversal. A file name must be a single path component.
	if fileID == "" ||
		fileID == "." ||
		fileID == ".." ||
		strings.ContainsAny(fileID, `/\`) ||
		filepath.Base(fileID) != fileID {
		http.Error(w, "invalid file name", http.StatusBadRequest)
		return
	}

	session, err := common.SessionGet(w, r)
	if err != nil {
		log.Printf("[Upload] SessionGet error: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Check channel access via PostgreSQL alias table.
	// Inquirers (access_right = 3) are not considered members.
	var allowed bool
	err = common.DB.AliasDB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM alias
			WHERE channel_id = $1
			  AND user_id = $2
			  AND access_right <> 3
		)
	`, channelIDInt, session.UserID).Scan(&allowed)
	if err != nil {
		log.Printf(
			"[Upload] Channel access query failed: userID=%d channelID=%d error=%v",
			session.UserID,
			channelIDInt,
			err,
		)
		http.Error(w, "failed to check channel access", http.StatusInternalServerError)
		return
	}

	if !allowed {
		log.Printf(
			"[Upload] NO CHANNEL ACCESS UserID=%d ChannelID=%d",
			session.UserID,
			channelIDInt,
		)
		http.Error(w, "no access right for file", http.StatusNotFound)
		return
	}

	// For downloadable files, check the file's available_by BIGINT[].
	if fileType == "file" {
		dbID := fmt.Sprintf(
			"%s/upload_data/%s/%s/%s",
			strings.TrimRight(cfg.UploadDir, "/"),
			fileType,
			channelID,
			fileID,
		)

		var fileAllowed bool
		err = common.DB.FileDB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM "file"
				WHERE file_id = $1
				  AND $2 = ANY(available_by)
			)
		`, dbID, session.UserID).Scan(&fileAllowed)
		if err != nil {
			log.Printf(
				"[Upload] File permission query failed: fileID=%s userID=%d error=%v",
				dbID,
				session.UserID,
				err,
			)
			http.Error(w, "failed to check file access", http.StatusInternalServerError)
			return
		}

		if !fileAllowed {
			log.Printf(
				"[Upload] User %d not allowed for file %s",
				session.UserID,
				dbID,
			)
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
	}

	// Build the physical file path.
	fullDir := filepath.Join(cfg.UploadDir, "upload_data", fileType)
	filePath := filepath.Join(channelID, fileID)

	file, err := http.Dir(fullDir).Open(filePath)
	if err != nil {
		log.Printf(
			"[Upload] File OPEN ERROR path=%s error=%v",
			filePath,
			err,
		)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(fileID))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)

	if fileType != "img" && !strings.HasPrefix(contentType, "image/") {
		w.Header().Set(
			"Content-Disposition",
			fmt.Sprintf(
				"attachment; filename*=UTF-8''%s",
				url.PathEscape(fileID),
			),
		)
	}

	if _, err := io.Copy(w, file); err != nil {
		log.Printf("[Upload] Streaming ERROR: %v", err)
		// The response may already have started, so avoid writing
		// a second HTTP error body after a partial file transfer.
		return
	}
}