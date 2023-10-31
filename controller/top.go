package controller

import (
  "log"
  "net/http"
)

func Top(w http.ResponseWriter, r *http.Request) {
  log.Println(r.URL)
  // if r.URL.Path != "/" {
  //  http.Error(w, "Not found", http.StatusNotFound)
  //  return
  // }
  // if r.Method != http.MethodGet {
  //   http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  //   return
  // }
  http.ServeFile(w, r, "public/index.html")
}
