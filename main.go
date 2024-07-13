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
		http.HandleFunc("/EmojiToggle/", controller.EmojiToggle)
		http.HandleFunc("/GoogleIdentity/", controller.GoogleIdentity)
		// http.HandleFunc("/HubPush/", controller.HubPush)
		// http.HandleFunc("/MessageEdit/", controller.MessageEdit)
		// http.HandleFunc("/MessagePost/", controller.MessagePost)
		http.HandleFunc("/PrivateAdd/", controller.PrivateAdd)
		http.HandleFunc("/PushGet/", controller.PushGet)
		http.HandleFunc("/PushResponse/", controller.PushResponse)
		http.HandleFunc("/PushSubscribe/", controller.PushSubscribe)
		http.HandleFunc("/img/", controller.Img)
		http.HandleFunc("/Init/", controller.Init)
		http.HandleFunc("/ThreadEdit/", controller.ThreadEdit)
		http.HandleFunc("/ThreadPost/", controller.ThreadPost)
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
