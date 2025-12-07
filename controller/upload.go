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
    "strings"
    "time"

    "go.mongodb.org/mongo-driver/bson"

    "chat/collection"
    "chat/common"
)

func Upload(w http.ResponseWriter, r *http.Request) {
    cfg := common.LoadConfig()
    log.Printf("[Upload] START RawReqPath=%s", r.URL.Path)

    // --- URL デコード ---
    decodedPath, err := url.PathUnescape(r.URL.Path)
    if err != nil {
        log.Printf("[Upload] URL decode error: %v", err)
        http.Error(w, "invalid URL", http.StatusBadRequest)
        return
    }
    log.Printf("[Upload] DecodedPath=%s", decodedPath)

    // /upload/<fileType>/<channelID>/<fileID>
    parts := strings.Split(decodedPath, "/")
    log.Printf("[Upload] URL Split: %+v", parts)

    if len(parts) < 5 {
        log.Printf("[Upload] Invalid URL length (%d) → %+v", len(parts), parts)
        http.Error(w, "invalid URL path", http.StatusNotFound)
        return
    }

    fileType := parts[2]   // "file" or "img"
    channelID := parts[3]
    fileID := parts[4]     // 実ファイル名。日本語OK

    log.Printf("[Upload] fileType=%s channelID=%s fileID=%s",
        fileType, channelID, fileID)

    // --- セッションチェック ---
    session, err := common.SessionGet(w, r)
    if err != nil {
        log.Printf("[Upload] SessionGet error: %v", err)
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    log.Printf("[Upload] SessionLoad OK: UserID=%s ChannelAliases=%+v",
        session.UserID, session.ChannelAliases)

    // --- チャンネルアクセス権確認 ---
    allowed := false
    for _, d := range session.ChannelAliases {
        if d.ChannelID == channelID {
            allowed = true
            break
        }
    }
    if !allowed {
        log.Printf("[Upload] NO CHANNEL ACCESS UserID=%s ChannelID=%s",
            session.UserID, channelID)
        http.Error(w, "no access right for file", http.StatusNotFound)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // --- fileType=file の場合だけ DB アクセス権チェック ---
    if fileType == "file" {
        var fileData collection.FileStruct
        coll := common.DB.FileDB.Collection("file")

        dbID := fmt.Sprintf("./upload_data/%s/%s/%s", fileType, channelID, fileID)
        log.Printf("[Upload] DB lookup _id = %s", dbID)

        err = coll.FindOne(ctx, bson.M{"_id": dbID}).Decode(&fileData)
        if err != nil {
            log.Printf("[Upload] DB Find Error: %v", err)
            http.Error(w, "file not found", http.StatusNotFound)
            return
        }

        // ユーザーが許可されてるか？
        ok := false
        for _, uid := range fileData.AvailableBy {
            if uid == session.UserID {
                ok = true
                break
            }
        }
        if !ok {
            log.Printf("[Upload] User %s NOT allowed for file %s",
                session.UserID, dbID)
            http.Error(w, "file not found", http.StatusNotFound)
            return
        }
    }

    // --- ファイルパス作成（aliasName 廃止） ---
    filePath := fmt.Sprintf("%s/%s", channelID, fileID)
    fullDir := cfg.UploadDir + "/upload_data/" + fileType
    fullPath := fullDir + "/" + filePath

    log.Printf("[Upload] Opening file: %s", fullPath)

    file, err := http.Dir(fullDir).Open(filePath)
    if err != nil {
        log.Printf("[Upload] File OPEN ERROR path=%s error=%v", filePath, err)
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    defer file.Close()

    // --- Content-Type 判定（画像なら inline 表示） ---
    ext := strings.ToLower(filepath.Ext(fileID))
    contentType := mime.TypeByExtension(ext)
    if contentType == "" {
        contentType = "application/octet-stream"
    }

    w.Header().Set("Content-Type", contentType)

    // 画像なら表示、非画像はダウンロード
    if fileType != "img" && !strings.HasPrefix(contentType, "image/") {
        w.Header().Set("Content-Disposition",
            fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(fileID)))
        log.Printf("[Upload] Non-image → download mode")
    } else {
        log.Printf("[Upload] Image → inline mode")
    }

    log.Printf("[Upload] Content-Type: %s", contentType)

    // --- Stream 出力 ---
    _, err = io.Copy(w, file)
    if err != nil {
        log.Printf("[Upload] Streaming ERROR: %v", err)
        http.Error(w, "Failed to read file", http.StatusInternalServerError)
        return
    }

    log.Printf("[Upload] FINISHED OK file=%s", filePath)
}
