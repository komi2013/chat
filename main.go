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
	if len(os.Args) == 1 {
	  cfg := common.LoadConfig()
	  common.InitMongo(cfg)
		http.HandleFunc("/AdDelete/", controller.AdDelete)
		http.HandleFunc("/AdEdit/", controller.AdEdit)
		http.HandleFunc("/AdGet/", controller.AdGet)
		http.HandleFunc("/AdInvoice/", controller.AdInvoice)
		// http.HandleFunc("/AdPriceGet/", controller.AdPriceGet)
		http.HandleFunc("/AdPublicGet/", controller.AdPublicGet)
		http.HandleFunc("/ChannelAdd/", controller.ChannelAdd)
		http.HandleFunc("/ChannelDelete/", controller.ChannelDelete)
		http.HandleFunc("/ChannelEdit/", controller.ChannelEdit)
		http.HandleFunc("/ChannelJoin/", controller.ChannelJoin)
		http.HandleFunc("/ContentsJustPush/", controller.ContentsJustPush)
		http.HandleFunc("/ContentsPush/", controller.ContentsPush)
		http.HandleFunc("/LogFromJS/", controller.LogFromJS)
		http.HandleFunc("/NicknameGet/", controller.NicknameGet)
		http.HandleFunc("/PushSubscribe/", controller.PushSubscribe)
http.HandleFunc("/PushSubscribeMobile/", controller.PushSubscribeMobile)
		http.HandleFunc("/ReceptionBook/", controller.ReceptionBook)
		// http.HandleFunc("/ReceptionCheck/", controller.ReceptionCheck)
		http.HandleFunc("/ReceptionEdit/", controller.ReceptionEdit)
		http.HandleFunc("/ReceptionGet/", controller.ReceptionGet)
		http.HandleFunc("/ReceptionShift/", controller.ReceptionShift)
		http.HandleFunc("/ReceptionOrder/", controller.ReceptionOrder)
		http.HandleFunc("/ReceptionOrderDelete/", controller.ReceptionOrderDelete)
		http.HandleFunc("/ReceptionQueueEdit/", controller.ReceptionQueueEdit)
		http.HandleFunc("/ReceptionThreadCustomer/", controller.ReceptionThreadCustomer)
		// http.HandleFunc("/ShiftStaffEdit/", controller.ShiftStaffEdit)
		// http.HandleFunc("/StorePush/", controller.StorePush)
		// http.HandleFunc("/StoreSelect/", controller.StoreSelect)
		http.HandleFunc("/SignInGoogle/", controller.SignInGoogle)
		http.HandleFunc("/SignInGoogleMobile/", controller.SignInGoogleMobile)
		http.HandleFunc("/SignInTmp/", controller.SignInTmp)
		http.HandleFunc("/TweetEmoji/", controller.TweetEmoji)
		http.HandleFunc("/TweetGet/", controller.TweetGet)
		http.HandleFunc("/TweetGetLatest/", controller.TweetGetLatest)
		http.HandleFunc("/TweetPost/", controller.TweetPost)
		http.HandleFunc("/upload/", controller.Upload)
		http.HandleFunc("/UserEdit/", controller.UserEdit)
		http.HandleFunc("/UserGet/", controller.UserGet)
		http.HandleFunc("/WebRTCTokenGet/", controller.WebRTCTokenGet)

		// default
		http.HandleFunc("/", controller.Top)

		fmt.Println("starting.." + cfg.GoPort + " " + cfg.CacheV )
		log.Fatal(http.ListenAndServe(cfg.GoPort, nil))
	} else {
		fmt.Printf("console is running %#v\n", os.Args)
	  cfg := common.LoadConsoleConfig()
	  common.InitMongo(cfg)
		switch os.Args[1] {
			case "AdPublish":	console.AdPublish()
			case "FileClean":	console.FileClean()
			case "TestPush": console.TestPush()
			case "TestCode": console.TestCode()
		}
	}
}
