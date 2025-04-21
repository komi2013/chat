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
		http.HandleFunc("/ChannelAdd/", controller.ChannelAdd)
		http.HandleFunc("/ChannelDelete/", controller.ChannelDelete)
		http.HandleFunc("/ChannelInvite/", controller.ChannelInvite)
		http.HandleFunc("/ChannelJoin/", controller.ChannelJoin)
		http.HandleFunc("/ContentsJustPush/", controller.ContentsJustPush)
		http.HandleFunc("/ContentsPush/", controller.ContentsPush)
		http.HandleFunc("/GoogleIdentity/", controller.GoogleIdentity)
		// http.HandleFunc("/PushResponse/", controller.PushResponse)
		http.HandleFunc("/PushSubscribe/", controller.PushSubscribe)
		http.HandleFunc("/ReceptionBook/", controller.ReceptionBook)
		http.HandleFunc("/ReceptionCheck/", controller.ReceptionCheck)
		http.HandleFunc("/ReceptionDelete/", controller.ReceptionDelete)
		http.HandleFunc("/ReceptionGet/", controller.ReceptionGet)
		http.HandleFunc("/ReceptionShift/", controller.ReceptionShift)
		http.HandleFunc("/ReceptionOrder/", controller.ReceptionOrder)
		// http.HandleFunc("/ShiftStaffEdit/", controller.ShiftStaffEdit)
		// http.HandleFunc("/StorePush/", controller.StorePush)
		// http.HandleFunc("/StoreSelect/", controller.StoreSelect)
		http.HandleFunc("/TmpLogin/", controller.TmpLogin)
		http.HandleFunc("/upload/", controller.Upload)
		http.HandleFunc("/", controller.Top)

		fmt.Println("starting.." + common.CacheV)
		fmt.Println(common.MongoDb1 + common.GoPort)
		log.Fatal(http.ListenAndServe(common.GoPort, nil))
	} else {
		fmt.Printf("console is running %#v\n", os.Args)
	}
}
