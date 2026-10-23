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
		common.InitPostgres(cfg)
		http.HandleFunc("/ChannelAdd/", controller.ChannelAdd)
		http.HandleFunc("/ChannelDelete/", controller.ChannelDelete)
		http.HandleFunc("/ChannelEdit/", controller.ChannelEdit)
		http.HandleFunc("/ChannelJoin/", controller.ChannelJoin)
		http.HandleFunc("/ContentsJustPush/", controller.ContentsJustPush)
		http.HandleFunc("/ContentsPush/", controller.ContentsPush)
		http.HandleFunc("/LogFromJS/", controller.LogFromJS)
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
		http.HandleFunc("/SignInTestMobile/", controller.SignInTestMobile)
		http.HandleFunc("/upload/", controller.Upload)
		http.HandleFunc("/UserEdit/", controller.UserEdit)
		http.HandleFunc("/UserGet/", controller.UserGet)
		http.HandleFunc("/WebRTCTokenGet/", controller.WebRTCTokenGet)

		// default
		http.HandleFunc("/", controller.Top)

		fmt.Println("starting.." + cfg.GoPort + " " + cfg.CacheV)
		log.Fatal(http.ListenAndServe(cfg.GoPort, nil))
	} else {
		fmt.Printf("console is running %#v\n", os.Args)
		commandIndex := 1
		if os.Args[commandIndex] == "." && len(os.Args) > 2 {
			commandIndex = 2
		}
		command := os.Args[commandIndex]

		if command == "TestPushFCM" {
			console.TestPushFCM()
			return
		}

		cfg := common.LoadConsoleConfig()
		common.InitPostgres(cfg)
		switch command {
		case "FileClean":
			console.FileClean()
		case "TestPush":
			console.TestPush()
		case "TestCode":
			console.TestCode()
		default:
			fmt.Printf("Unknown console command %q. Available commands: AdPublish, FileClean, TestPush, TestPushFCM, TestCode\n", command)
			return
		}
	}
}
