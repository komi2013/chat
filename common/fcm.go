package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2/google"
	"chat/collection"
)

type FCMMessage struct {
	Message struct {
		Token        string            `json:"token"`
		Topic        string            `json:"topic,omitempty"`
		Condition    string            `json:"condition,omitempty"`
		Notification *Notification     `json:"notification,omitempty"`
		Data         map[string]string `json:"data,omitempty"`
		Android      *AndroidConfig    `json:"android,omitempty"`
	} `json:"message"`
}

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Image string `json:"image,omitempty"`
}

type AndroidConfig struct {
	Priority         string             `json:"priority,omitempty"`
	CollapseKey      string             `json:"collapse_key,omitempty"`
	TTL              string             `json:"ttl,omitempty"`
	RestrictedPackageName string        `json:"restricted_package_name,omitempty"`
	Notification     *AndroidNotification `json:"notification,omitempty"`
}

type AndroidNotification struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Color        string `json:"color,omitempty"`
	Sound        string `json:"sound,omitempty"`
	Tag          string `json:"tag,omitempty"`
	ClickAction  string `json:"click_action,omitempty"`
	BodyLocKey   string `json:"body_loc_key,omitempty"`
	BodyLocArgs  []string `json:"body_loc_args,omitempty"`
	TitleLocKey  string `json:"title_loc_key,omitempty"`
	TitleLocArgs []string `json:"title_loc_args,omitempty"`
}

type FCMManager struct {
	client *http.Client
	config *Config
}

func NewFCMManager(cfg *Config) *FCMManager {
	return &FCMManager{
		client: &http.Client{Timeout: 30 * time.Second},
		config: cfg,
	}
}

func (f *FCMManager) getAccessToken() (string, error) {
	// Load service account key from file
	data, err := ioutil.ReadFile(f.config.FirebaseServiceAccountKey)
	if err != nil {
		return "", fmt.Errorf("failed to read service account key: %v", err)
	}

	// Create credentials from service account key
	creds, err := google.CredentialsFromJSON(context.Background(), data, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return "", fmt.Errorf("failed to create credentials: %v", err)
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %v", err)
	}

	return token.AccessToken, nil
}

func (f *FCMManager) SendNotification(token, title, body, channelId string) error {
	accessToken, err := f.getAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}

	message := FCMMessage{}
	message.Message.Token = token
	message.Message.Notification = &Notification{
		Title: title,
		Body:  body,
	}
	message.Message.Data = map[string]string{
		"channelId": channelId,
		"title":     title,
		"body":      body,
	}
	message.Message.Android = &AndroidConfig{
		Priority: "high",
		Notification: &AndroidNotification{
			Title:       title,
			Body:        body,
			Icon:        "ic_notification",
			Color:       "#3F51B5",
			Sound:       "default",
			ClickAction: "MainActivity",
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", f.config.FirebaseProjectID)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("FCM API returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (f *FCMManager) SendToMultipleTokens(tokens []string, title, body, channelId string) error {
	// For multiple tokens, use multicast messaging
	// Or send individual messages in goroutines
	for _, token := range tokens {
		if err := f.SendNotification(token, title, body, channelId); err != nil {
			// Log error but continue with other tokens
			fmt.Printf("Failed to send to token %s: %v\n", token, err)
		}
	}
	return nil
}

// Helper function to get FCM tokens from sessions
func GetFCMTokensFromSessions(sessions []collection.SessionStruct) []string {
	var tokens []string
	for _, session := range sessions {
		if session.FcmToken != "" && session.IsMobile {
			tokens = append(tokens, session.FcmToken)
		}
	}
	return tokens
}
