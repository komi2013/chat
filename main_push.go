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
	json.Unmarshal([]byte(`{"endpoint":"https://updates.push.services.mozilla.com/wpush/v2/gAAAAABlTdgMzziffOf-4fWTXtPtM9hnTfgWO6_ABAtMsGmOY-fcG0E4VLozToAYOMC5ZJVKsHm_u5McMXN1uMp_j52UMzKz77Psg8vwxniS4PHZylvulOw6lHOVl5QJjCu4jXuATr-BOjd4tvQVCZhFwNBiwFpXhzTjJcP8O6nTRPD8SmFp9v8","expirationTime":null,"keys":{"auth":"iLkiDQ8fKctQWh3LQuHR7g","p256dh":"BK2xPFLsHuBoRF3Ng0pELgXOFTfqRMqX_rxg89ihC_tVIbmO74mGgjsIiYP7lZraiVzoAGNcl0uTDpwTDrkYKrE"}}`), s)

	// Send Notification
	resp, err := webpush.SendNotification([]byte("['なせ','seijiro']"), s, &webpush.Options{
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
