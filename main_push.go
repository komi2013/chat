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
	json.Unmarshal([]byte(`{"endpoint":"https://updates.push.services.mozilla.com/wpush/v2/gAAAAABlq5Q4O5-V94gXatdpcEJgqNwipN5ayxQbJwHtN2UyrkK1oiYsIdV603yPOV-EFVj1zlu4eiPjj2-OlJGmZ14w2T_kygx53UmpSzJ3IeubnCvfNUAfnPuE8bCrrcZCfsCmK7Avtl3CpyfnftCIp2mPdQVa_rxoYzxJ4oZFTUdnCVNcXzs","expirationTime":null,"keys":{"auth":"zJ_Ed7bkr_3jRK9FzBjdIQ","p256dh":"BOZu7f-8VI17S9dElGab1tnFR111IG-0BKzD-I5wobXn_jVDvMuz9XPwW4aXgn3Fe6j9VA0xLH9oBTyjN9iE2FA"}}`), s)

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
