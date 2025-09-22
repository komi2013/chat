package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"chat/common"
	"chat/console"
	"chat/controller"
)

func main() {
  common.InitMongo()
	if len(os.Args) == 1 {
		http.HandleFunc("/AdEdit/", controller.AdEdit)
		http.HandleFunc("/AdGet/", controller.AdGet)
		http.HandleFunc("/AdPriceGet/", controller.AdPriceGet)
		http.HandleFunc("/ChannelAdd/", controller.ChannelAdd)
		http.HandleFunc("/ChannelDelete/", controller.ChannelDelete)
		http.HandleFunc("/ChannelInvite/", controller.ChannelInvite)
		http.HandleFunc("/ChannelJoin/", controller.ChannelJoin)
		http.HandleFunc("/ContentsJustPush/", controller.ContentsJustPush)
		http.HandleFunc("/ContentsPush/", controller.ContentsPush)
		http.HandleFunc("/LogFromJS/", controller.LogFromJS)
		http.HandleFunc("/PushSubscribe/", controller.PushSubscribe)
		http.HandleFunc("/ReceptionBook/", controller.ReceptionBook)
		// http.HandleFunc("/ReceptionCheck/", controller.ReceptionCheck)
		http.HandleFunc("/ReceptionEdit/", controller.ReceptionEdit)
		http.HandleFunc("/ReceptionGet/", controller.ReceptionGet)
		http.HandleFunc("/ReceptionShift/", controller.ReceptionShift)
		http.HandleFunc("/ReceptionOrder/", controller.ReceptionOrder)
		http.HandleFunc("/ReceptionOrderDelete/", controller.ReceptionOrderDelete)
		// http.HandleFunc("/ShiftStaffEdit/", controller.ShiftStaffEdit)
		// http.HandleFunc("/StorePush/", controller.StorePush)
		// http.HandleFunc("/StoreSelect/", controller.StoreSelect)
		http.HandleFunc("/SignInGoogle/", controller.SignInGoogle)
		http.HandleFunc("/SignInTmp/", controller.SignInTmp)
		http.HandleFunc("/upload/", controller.Upload)
		http.HandleFunc("/UserEdit/", controller.UserEdit)
		http.HandleFunc("/UserGet/", controller.UserGet)

		// default
		http.HandleFunc("/", controller.Top)

		fmt.Println("starting.." + common.GoPort + " " + common.CacheV )
		log.Fatal(http.ListenAndServe(common.GoPort, nil))
	} else {
		fmt.Printf("console is running %#v\n", os.Args)
		switch os.Args[1] {
		case "AdPublish":
			console.AdPublish()
		case "TestPush":
			console.TestPush()
		case "TestCountUp":
			console.TestCountUp()
		}
	}
}
