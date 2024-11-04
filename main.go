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
		http.HandleFunc("/BookmarkToggle/", controller.BookmarkToggle)
		http.HandleFunc("/ChannelAdd/", controller.ChannelAdd)
		http.HandleFunc("/ChannelEdit/", controller.ChannelEdit)
		http.HandleFunc("/CommunityMatch/", controller.CommunityMatch)
		http.HandleFunc("/ContentsPush/", controller.ContentsPush)
		http.HandleFunc("/EmojiToggle/", controller.EmojiToggle)
		http.HandleFunc("/GoogleIdentity/", controller.GoogleIdentity)
		http.HandleFunc("/GroupAliasEdit/", controller.GroupAliasEdit)
		http.HandleFunc("/img/", controller.Img)
		// http.HandleFunc("/HubPush/", controller.HubPush)
		// http.HandleFunc("/MessageEdit/", controller.MessageEdit)
		// http.HandleFunc("/MessagePost/", controller.MessagePost)
		http.HandleFunc("/PushGet/", controller.PushGet)
		http.HandleFunc("/PushResponse/", controller.PushResponse)
		http.HandleFunc("/PushSubscribe/", controller.PushSubscribe)
		http.HandleFunc("/ShiftStaffEdit/", controller.ShiftStaffEdit)
		http.HandleFunc("/ShiftStaffGet/", controller.ShiftStaffGet)
		http.HandleFunc("/ThreadEdit/", controller.ThreadEdit)
		http.HandleFunc("/ThreadPost/", controller.ThreadPost)
		http.HandleFunc("/TicketAdd/", controller.TicketAdd)
		http.HandleFunc("/TicketEdit/", controller.TicketEdit)
		http.HandleFunc("/TicketGet/", controller.TicketGet)
		http.HandleFunc("/TmpLogin/", controller.TmpLogin)
		http.HandleFunc("/upload/", controller.Upload)
		http.HandleFunc("/WindowAdd/", controller.WindowAdd)
		http.HandleFunc("/WindowGet/", controller.WindowGet)

		http.HandleFunc("/", controller.Top)

		fmt.Println("starting.." + common.CacheV)
		fmt.Println(common.MongoDb1 + common.GoPort)

		log.Fatal(http.ListenAndServe(common.GoPort, nil))

	} else {
		fmt.Printf("console is running %#v\n", os.Args)
	}
}
