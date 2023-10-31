package controller

import (
  // "context"
  "fmt"
  "log"
  "net/http"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/mongo/options"

  // "chat/common"
  // "chat/logic/quiz"
)

func Init(w http.ResponseWriter, r *http.Request) {
  log.Println(r.URL)
  // if r.URL.Path != "/" {
  //  http.Error(w, "Not found", http.StatusNotFound)
  //  return
  // }
  time.Sleep(3 * time.Second)

  fmt.Fprint(w, `[[1],[[1,"channel name 1","description 1"],[1,"channel name 2","description2 1"]],"dfjdkosjo"]`)
}
