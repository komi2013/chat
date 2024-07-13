package controller

import (
  "io"
  "mime"
  "net/http"
  "path/filepath"
  "strings"
)

func Img(w http.ResponseWriter, r *http.Request) {

	filePath := strings.TrimPrefix(r.URL.Path, "/img")
  file, err := http.Dir("./img").Open(filePath)
  if err != nil {
    http.Error(w, "File not found", http.StatusNotFound)
    return
  }
  defer file.Close()

  contentType := "application/octet-stream"
  if ext := filepath.Ext(filePath); ext != "" {
    contentType = mime.TypeByExtension(ext)
  }

  w.Header().Set("Content-Type", contentType)

  _, err = io.Copy(w, file)
  if err != nil {
    http.Error(w, "Failed to read file", http.StatusInternalServerError)
    return
  }

}
