// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	// "chat"
	"chat/common"
	"chat/controller"
	"chat/logic/ws"

)

var addr = flag.String("addr", common.GoPort, "http service address")

func main() {
	flag.Parse()
	hub1 := ws.NewHub()
	go hub1.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub1, w, r)
	})
	http.HandleFunc("/", controller.ServeHome)
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
