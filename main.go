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
	json.Unmarshal([]byte(`{"endpoint":"https://updates.push.services.mozilla.com/wpush/v2/gAAAAABlOx7ZU4p6pF0nDL8aP81kVZWfx_3PCcVveQ-SUkgB8f_4O_kP9DWoZUhUJkZgsHJA-XmfMEPKIiIzWRZ1nqj8BCJ4r35O97sx35h8zGeb3x61WATT9A9NlRZU8xx5sypVbkukvJhBgQjoLULxNO1mNSAcPJK34PploTgMrau51KOAr04","expirationTime":null,"keys":{"auth":"OvlhDZFMPcn-LUd1_p-S7g","p256dh":"BCpyO5Cu7F4PYgW0jnLoevAejFH6aUbnqcrykewh1HwM2SdAD-ncRzDJsZEQ5sYahlmZfmYOFnRGy2NwDATs-Fg"}}`), s)

	// Send Notification
	resp, err := webpush.SendNotification([]byte("seijiro"), s, &webpush.Options{
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
 // privateKey p_FTSk4LAiFQfv-gkuRhyFrThGP5nI7Ot79bESuMUtA
 // publicKey BDEkA4OaXL61KXrtH_jfTovZDJ4cJbHlVRWmvyXoCDPu6Y7ikPjL4MEmmIjzdDULXUWkVjYOZ0b0NxVyHDgTsh8

// curl -X POST "https://fcm.googleapis.com/fcm/send/epXPEfN5M5U:APA91bHjfV9TzztVAcHJ1fHT3qeDlj8AWYRckVTHb5V19IghQRJFBo1Xms3_W1ifWsCYvdHJwbHIQEiNX0Za2TqAwzaNxGv9nht39o5WXnOsGZgq19uJ29Td3wHQf65FNPOpI4Vodxdj" -H "Authorization: Bearer p_FTSk4LAiFQfv-gkuRhyFrThGP5nI7Ot79bESuMUtA" -H "crypto-key: p256dh=QkRFa0E0T2FYTDYxS1hydEhfamZUb3ZaREo0Y0piSGxWUldtdnlYb0NEUHU2WTdpa1BqTDRNRW1tSWp6ZERVTFhVV2tWallPWjBiME54VnlIRGdUc2g4" -H "Content-Type: application/json" --data '{"subscription":{"endpoint":"https://fcm.googleapis.com/fcm/send/epXPEfN5M5U:APA91bHjfV9TzztVAcHJ1fHT3qeDlj8AWYRckVTHb5V19IghQRJFBo1Xms3_W1ifWsCYvdHJwbHIQEiNX0Za2TqAwzaNxGv9nht39o5WXnOsGZgq19uJ29Td3wHQf65FNPOpI4Vodxdj","expirationTime":null,"keys":{"p256dh":"BBhLg70KQBk0dg4YtFFqPOnqxERbG0VNSFnC4REAzNTJJjMx8UUI0dhy-_hXkMMZAL0SEQQ1bYamlAyedhx9XlI","auth":"hjVATUVVtxtVvDvpyE-dzw"}}}
// }'

// {"subscription":{"endpoint":"https://fcm.googleapis.com/fcm/send/dvWJYNHxMmk:APA91bEgzGORDpNSX5QB3L6-Q8PT9WF3ENGyr6Mo55zu4C6Ic4g-_D0YMBsEmSIxDFgSu-GPAHgekacX3S418AFw2bEqK67MyzN161JRIBPEcd0t9qOm6-x6f-eL7B-X2eBIWxMEaL1B","expirationTime":null,"keys":{"p256dh":"BF1NHDRkn5ylbc2JhDjepa8MxaU_w1eM1i7bAZyJDtF4Ii4-Vc0I6UjS0sjBXWV-JluQKNDEaPe4wghERdzAo1k","auth":"Djc969tPNyhg4zQvirN9nw"}}}

// npx web-push send-notification --endpoint=https://fcm.googleapis.com/fcm/send/cYvLdSB6Fx4:APA91bFR3zehFpyJXfGbBW2JF_tXEYCK-ZEqDvrCc0N4N2WSk6V4nk_CEOIPMEc82poxlRw0u69lPewKjAe9_dF9q_z4MbpAyu29Lg-FyFsPvfkQOZuNkVBHQkGZKErwkDVGNt9bXdkV --auth=WaGri_Sa9R-_3jAhiL9qvg --key=BHOWS_HG_fQBzaKdYQfuUl3_YP-bf0sy2-1tz04MzENRnIWFR9MlNnp-s_dHRX4UFBABcuF8qaXnZl_RI9EXvS4 --payload=hello --vapid-pubkey=BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8 --vapid-pvtkey=bdSiNzUhUP6piAxLH-tW88zfBlWWveIx0dAsDO66aVU --vapid-subject=mailto:asdf@gmail.com


// curl -X POST "https://fcm.googleapis.com/fcm/send/cYvLdSB6Fx4:APA91bFR3zehFpyJXfGbBW2JF_tXEYCK-ZEqDvrCc0N4N2WSk6V4nk_CEOIPMEc82poxlRw0u69lPewKjAe9_dF9q_z4MbpAyu29Lg-FyFsPvfkQOZuNkVBHQkGZKErwkDVGNt9bXdkV" \
// -H "Authorization: Bearer WaGri_Sa9R-_3jAhiL9qvg" \
// -H "Content-Type: application/json" \
// --data '{
//   "notification": {
//     "title": "Your Notification Title",
//     "body": "Your Notification Body"
//   },
//   "data": {
//     "crypto-key": "p256ecdsa=QklOMkpjNVZta215LVMzQVVyY01scEt4SnBMZVZSQWZ1OVdCcVViSjcwU0pPQ1dHQ0dYS1ktWHp5aDdIRHI2S2JSREdZSGpxWjA2T2NTM0JqRDd1QW04",
//     "payload": "your_payload",
//     "vapid": {
//       "public_key": "BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8",
//       "private_key": "bdSiNzUhUP6piAxLH-tW88zfBlWWveIx0dAsDO66aVU",
//       "subject": "your_email@example.com"
//     }
//   }
// }'


