package console

import (
	"encoding/json"
	// "encoding/base64"
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"

	"chat/common"
)

func TestPush() {
	// str := base64.StdEncoding.EncodeToString([]byte("BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8"))
	// fmt.Println(str)
	// Decode subscription
  cfg := common.LoadConsoleConfig()

	s := &webpush.Subscription{}
	// json.Unmarshal([]byte(``), s)
	json.Unmarshal([]byte(`{"endpoint":"https://fcm.googleapis.com/fcm/send/dXJER6yxO4Q:APA91bGNiRcLoBe1GDw2BeZZaXshjJe2tUtonWxpsIt2P7eJcfp46UwWynqlbKH269n29bqbIaXGIUNnXll6guCh0MSAUaV-i0-FQIhXUg6vk8bgy6SZT2QsQeoGgTGXv6sUeL98an9_","expirationTime":null,"keys":{"p256dh":"BHGyoJX53YYhlwDjb6GhnU7-2mv1ndZo9eQ21nvRhMyKXDCCxKjHOupCeGPUQ-D_rmrNMnfJyfYgfQo8jCclE7I","auth":"covFZxc0wZPifKFFFPIG7w"}}`), s)
	payload := map[string]interface{}{
	    "type":   "thread",
	    "channel": "C",
	    "parent":  "65acf63d8309d8da55154dea",
	    "message": "＊p＊苔こけ・＊p＊",
	    "icon":    "/ivan.png",
	    "url":     "/thread/663019ee703f7d8a80884514/65acf63d8309d8da55154dea/",
	}

	jsonPayload, _ := json.Marshal(payload)

	resp, err := webpush.SendNotification(jsonPayload, s, &webpush.Options{
	    Subscriber:      "example@example.com",
	    VAPIDPublicKey:  "BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8",
	    VAPIDPrivateKey: cfg.VAPIDPrivateKey,
	    TTL:             30,
	})


	// Send Notification
	// resp, err := webpush.SendNotification([]byte(`["thread","663019ee703f7d8a80884514","662f55cb18eccdac4820cab7","＊p＊苔こけ・＊p＊","sei2","/ivan.png","2024-04-30T07:06:38.643361442+09:00","65acf63d8309d8da55154dea","","","",""]`), s, &webpush.Options{
	// 	Subscriber:      "example@example.com",
	// 	VAPIDPublicKey:  "BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8",
	// 	VAPIDPrivateKey: cfg.VAPIDPrivateKey,
	// 	TTL:             30,
	// })
	if err != nil {
		// TODO: Handle error
    fmt.Printf(" err %s\n", err)
	}
	defer resp.Body.Close()
}
