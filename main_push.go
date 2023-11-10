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
	json.Unmarshal([]byte(`{"endpoint":"https://updates.push.services.mozilla.com/wpush/v2/gAAAAABlTXo2I7rbbOe7OmVPBabQP1khUxpO-HFoN0CwY_P0R1om-7Z3IKGsZXikZAqDtv7gmAYZb1DjSn22DfOu-MIkJP5OrSnRU5oxOLLcJcYa5EGOte4KtmqOdCXqbOR1zj2C0m4QlcBNFPq1OdTbyEjfzoSh_pvW0shC_O29ulQbi_xw32s","expirationTime":null,"keys":{"auth":"SxgeXnUzoL5SHjhZu--5OA","p256dh":"BNGgh55-SVVBzZ6o1NY_uDHfcrNQ2ivkcRbE6VxzdRFCQK-XNoQNUHJTvK2LJ8Ks0BH4ub4NDb-3zjTg-l-WgRQ"}}`), s)

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
