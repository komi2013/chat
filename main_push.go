package main

import (
	"encoding/json"
	// "encoding/base64"
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	// str := base64.StdEncoding.EncodeToString([]byte("BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8"))
	// fmt.Println(str)
	// Decode subscription
	s := &webpush.Subscription{}
	json.Unmarshal([]byte(`{"endpoint":"https://updates.push.services.mozilla.com/wpush/v2/gAAAAABlPK-ybScB0Cb8jwaA1OtOlcRcQ4KBuq0u0Dl9fkoY35lu5-AJigi9kXifDf4HEksKhBkJ0DdsBDLYgfA634tskb24b7qE1zchSfQEI3ISvEMTqFohTFNrj86vqJs7w40WVy7gaOznnruTcz6pbQbKkPXE_AKXlcwNNqq1gaalYjSGCXQ","expirationTime":null,"keys":{"auth":"13BeZk32Ln0RoRqowv6OxQ","p256dh":"BGMnl19ssZeUVYVMR4T0xX0TpMGfu7oiFF-XW6GQ873axeVNCcZ0rVvaz1bwOSeyLstAHqnyXJ35MP2OgCM3ZvI"}}`), s)

	// Send Notification
	resp, err := webpush.SendNotification([]byte("['komatsu','seijiro']"), s, &webpush.Options{
		Subscriber:      "example@example.com",
		VAPIDPublicKey:  "BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8",
		VAPIDPrivateKey: "bdSiNzUhUP6piAxLH-tW88zfBlWWveIx0dAsDO66aVU",
		TTL:             30,
	})
	if err != nil {
		// TODO: Handle error
    fmt.Printf(" err %s\n", err)
	}
	defer resp.Body.Close()
}
