package controller

import (
  // "context"
  // "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/mongo/options"

  // "chat/common"
  // "chat/logic/quiz"
)

func SubscriptionPost(w http.ResponseWriter, r *http.Request) {
  log.Println(r.URL)
  // if r.URL.Path != "/" {
  //  http.Error(w, "Not found", http.StatusNotFound)
  //  return
  // }
  time.Sleep(3 * time.Second)

  // resp, err := http.Get(url)
  // if err != nil {
  //         panic(err)
  // }
  // defer resp.Body.Close()
  // fmt.Printf("%s", j)

  // var j interface{}
  // err = json.NewDecoder(r.Body.Read(body)).Decode(&j)
  // if err != nil {
  //   panic(err)
  // }
  // fmt.Printf("%s", j)

  
  fmt.Printf("%s", r.FormValue("json"))

  fmt.Fprint(w, `[[1]`)
}
