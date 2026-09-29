package test

import (
	"testing"

	"chat/collection"
	"chat/common"
)

func TestResolvePushTarget(t *testing.T) {
	const vapidToken = `{"endpoint":"https://push.example.com/1"}`

	tests := []struct {
		name       string
		session    collection.SessionStruct
		wantDevice int
		wantToken  string
	}{
		{
			name:       "vapid registration",
			session:    collection.SessionStruct{PushToken: vapidToken, DeviceType: 1},
			wantDevice: 1,
			wantToken:  vapidToken,
		},
		{
			name:       "fcm registration",
			session:    collection.SessionStruct{PushToken: "fcm-token", DeviceType: 2},
			wantDevice: 2,
			wantToken:  "fcm-token",
		},
		{
			name:       "apns registration",
			session:    collection.SessionStruct{PushToken: "apns-token", DeviceType: 3},
			wantDevice: 3,
			wantToken:  "apns-token",
		},
		{
			name:       "pushToken without deviceType is corrupt",
			session:    collection.SessionStruct{PushToken: "orphan-token"},
		},
		{
			name:       "pushToken with unknown deviceType is corrupt",
			session:    collection.SessionStruct{PushToken: "orphan-token", DeviceType: 99},
		},
		{
			name: "nothing registered",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotDevice, gotToken := common.ResolvePushTarget(test.session)
			if gotDevice != test.wantDevice {
				t.Errorf("ResolvePushTarget() deviceType = %d, want %d", gotDevice, test.wantDevice)
			}
			if gotToken != test.wantToken {
				t.Errorf("ResolvePushTarget() token = %q, want %q", gotToken, test.wantToken)
			}
		})
	}
}