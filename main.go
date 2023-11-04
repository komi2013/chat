package main

import (
  "fmt"
  "log"
  "net/http"
  "os"

  "chat/common"
  // "chat/console"
  "chat/controller"
)

func main() {

  if len(os.Args) == 1 {
    http.HandleFunc("/Init/", controller.Init)
    http.HandleFunc("/SetCookie/", controller.SetCookie)
    http.HandleFunc("/SubscriptionPost/", controller.SubscriptionPost)

    http.HandleFunc("/", controller.Top)

    fmt.Println("starting.." + common.CacheV)
    fmt.Println(common.MongoDb1 + common.GoPort)

    log.Fatal(http.ListenAndServe(common.GoPort, nil))

  } else {
    fmt.Printf("console is running %#v\n", os.Args)
  }
}
