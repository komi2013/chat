package console

import (
	"fmt"
	"os"

	"chat/common"
)

// TestPushFCM sends a test notification to one FCM registration token.
// The token is read from FCM_TEST_TOKEN so it is not stored in source code.
func TestPushFCM() {
	// Load infrastructure/.env first, while allowing a process environment
	// variable to override it for one-off tests.
	cfg := common.LoadConfig()
	token := os.Getenv("FCM_TEST_TOKEN")
	if token == "" {
		token = cfg.FCMTestToken
	}
	if token == "" {
		fmt.Println("FCM_TEST_TOKEN is not set in the process environment or infrastructure/.env")
		return
	}

	if err := common.SendFCMPushNotification(
		"Goからのテスト通知",
		"FCMのテストPush通知です",
		"test-channel",
		token,
	); err != nil {
		fmt.Printf("FCM push failed: %v\n", err)
		return
	}

	fmt.Println("FCM push sent successfully")
}
