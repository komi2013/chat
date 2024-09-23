package controller

import (
  "log"
  "net/http"
)

func Top(w http.ResponseWriter, r *http.Request) {
  log.Println(r.URL)
  http.ServeFile(w, r, "public/index.html")
}
