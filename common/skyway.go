package common

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type SkyWayToken struct {
	JTI   string      `json:"jti"`
	IAT   int64       `json:"iat"`
	EXP   int64       `json:"exp"`
	Scope SkyWayScope `json:"scope"`
}

type SkyWayScope struct {
	App SkyWayApp `json:"app"`
}

type SkyWayApp struct {
	ID       string          `json:"id"`
	Turn     bool            `json:"turn"`
	Actions  []string        `json:"actions"`
	Channels []SkyWayChannel `json:"channels"`
}

type SkyWayChannel struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	Actions []string            `json:"actions"`
	Members []SkyWayChannelUser `json:"members"`
	SfuBots []SkyWaySfuBot      `json:"sfuBots"`
}

type SkyWayChannelUser struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Actions      []string           `json:"actions"`
	Publication  SkyWayActionHolder `json:"publication"`
	Subscription SkyWayActionHolder `json:"subscription"`
}

type SkyWaySfuBot struct {
	Actions     []string           `json:"actions"`
	Forwardings []SkyWayActionHolder `json:"forwardings"`
}

type SkyWayActionHolder struct {
	Actions []string `json:"actions"`
}

// struct → map に変換（JWT 署名用）
func structToMap(obj interface{}) map[string]interface{} {
	b, _ := json.Marshal(obj)
	var result map[string]interface{}
	json.Unmarshal(b, &result)
	return result
}

func GenerateSkyWayToken(appId, secret string) (string, error) {

	now := time.Now().Unix() // JSTで問題なし（Unix時刻は同じ）

	tokenObj := SkyWayToken{
		JTI: uuid.New().String(),
		IAT: now,
		EXP: now + 60*60*24*365, // 1年
		Scope: SkyWayScope{
			App: SkyWayApp{
				ID:      appId,
				Turn:    true,
				Actions: []string{"read"},
				Channels: []SkyWayChannel{
					{
						ID:      "*",
						Name:    "*",
						Actions: []string{"write"},
						Members: []SkyWayChannelUser{
							{
								ID:      "*",
								Name:    "*",
								Actions: []string{"write"},
								Publication: SkyWayActionHolder{
									Actions: []string{"write"},
								},
								Subscription: SkyWayActionHolder{
									Actions: []string{"write"},
								},
							},
						},
						SfuBots: []SkyWaySfuBot{
							{
								Actions: []string{"write"},
								Forwardings: []SkyWayActionHolder{
									{Actions: []string{"write"}},
								},
							},
						},
					},
				},
			},
		},
	}

	claims := jwt.MapClaims(structToMap(tokenObj))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
