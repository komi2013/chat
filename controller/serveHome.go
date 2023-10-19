package controller

import (
  // "context"
  // "fmt"
  "log"
  "net/http"
  // "time"

  // "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/mongo/options"

  // "chat/common"
  // "chat/logic/quiz"
)

func ServeHome(w http.ResponseWriter, r *http.Request) {
  log.Println(r.URL)
  // if r.URL.Path != "/" {
  //  http.Error(w, "Not found", http.StatusNotFound)
  //  return
  // }
  if r.Method != http.MethodGet {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }
  http.ServeFile(w, r, "home.html")
}
