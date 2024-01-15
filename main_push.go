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
	json.Unmarshal([]byte(`{"endpoint":"https://updates.push.services.mozilla.com/wpush/v2/gAAAAABloxQ5BTb70thKuHZ-WJxS_Q8IyJsvS-Aitq2IX_N_PPh0mrap9nS6xpir8VrcxT2xRuDwqC9_EmoGY_xveVmiEUaq9glUjFI-0qAZ6f4dz8mvR6Xe47SGSERzLBtfxbx7rReTFDP8ps1mv35RSv2NKRGkQ0Io-45AUE1R9bIdnZ5N8hw","expirationTime":null,"keys":{"auth":"LdVh3BA4Ce3kgwOTWfTvPQ","p256dh":"BB_yF2nHXVCk6mXghtvzkQOMY-QydjufH0G3N_0kSTgBnxPxvmuI3rdOjMe_hyrwIdfJikPU-5xYbVoeqJuLczg"}}`), s)

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
