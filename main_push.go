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
	json.Unmarshal([]byte(`{"endpoint":"https://fcm.googleapis.com/fcm/send/cYSrnHSWCtI:APA91bFNaBgx9eRYSZiAGYCxlPAh6e3j8WkFI_1wRdWzqYtpbpxcck76oB9oRFEYiuABy2nWO-O8KsnF9Xo3cPqRhhop7q_BuO0qMYXihDY2kWGhzsdriztgfsei3T6SUnm24eq1YA8Q","expirationTime":null,"keys":{"p256dh":"BBkbQ5jb1u60NSZhxg5fEvsWKkTpW34y-sMxRJP8Oyiscdnj89oZzND7qZEeIEmEUVbgDj5jtkB0d8MTkj7lpoQ","auth":"C8GdBJWx4fAB-HJkvVEzOQ"}}`), s)

	// Send Notification
	resp, err := webpush.SendNotification([]byte(`["message","seijiro"]`), s, &webpush.Options{
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
