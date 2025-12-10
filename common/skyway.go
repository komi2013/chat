package common

import (
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
    ID       string           `json:"id"`
    Turn     bool             `json:"turn"`
    Actions  []string         `json:"actions"`  // read + write 必須！
    Channels []SkyWayChannel  `json:"channels"`
}

type SkyWayChannel struct {
    ID      string              `json:"id"`
    Name    string              `json:"name"`
    Actions []string            `json:"actions"`
    Members []SkyWayChannelUser `json:"members"`
    SfuBots []SkyWaySfuBot      `json:"sfuBots"`
}

type SkyWayChannelUser struct {
    ID           string              `json:"id"`
    Name         string              `json:"name"`
    Actions      []string            `json:"actions"`
    Publication  SkyWayActionHolder  `json:"publication"`
    Subscription SkyWayActionHolder  `json:"subscription"`
}

type SkyWaySfuBot struct {
    Actions     []string             `json:"actions"`
    Forwardings []SkyWayActionHolder `json:"forwardings"`
}

type SkyWayActionHolder struct {
    Actions []string `json:"actions"`
}

func GenerateSkyWayToken(appId, secret string) (string, error) {
	now := time.Now().Unix()

	claims := jwt.MapClaims{
		"jti": uuid.New().String(),
		"iat": now,
		"exp": now + 60*60*24*365,
		"scope": map[string]interface{}{
			"app": map[string]interface{}{
				"id":   appId,
				"turn": true,
				"actions": []string{
					"read",
					"write",
				},
				"channels": []interface{}{
					map[string]interface{}{
						"id":      "*",
						"name":    "*",
						"actions": []string{"write"},
						"members": []interface{}{
							map[string]interface{}{
								"id":      "*",
								"name":    "*",
								"actions": []string{"write"},
								"publication": map[string]interface{}{
									"actions": []string{"write"},
								},
								"subscription": map[string]interface{}{
									"actions": []string{"write"},
								},
							},
						},
						"sfuBots": []interface{}{
							map[string]interface{}{
								"actions": []string{"write"},
								"forwardings": []interface{}{
									map[string]interface{}{
										"actions": []string{"write"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// func GenerateSkyWayToken(appId, secret string) (string, error) {
// 	now := time.Now().Unix()

// 	scope := map[string]interface{}{
// 		"app": map[string]interface{}{
// 			"id":   appId,
// 			"turn": true,
// 			"actions": []string{"read", "write"},
// 			"channels": []interface{}{
// 				map[string]interface{}{
// 					"id":      "*",
// 					"name":    "*",
// 					"actions": []string{"write"},
// 					"members": []interface{}{
// 						map[string]interface{}{
// 							"id":      "*",
// 							"name":    "*",
// 							"actions": []string{"write"},
// 							"publication": map[string]interface{}{
// 								"actions": []string{"write"},
// 							},
// 							"subscription": map[string]interface{}{
// 								"actions": []string{"write"},
// 							},
// 						},
// 					},
// 					"sfuBots": []interface{}{
// 						map[string]interface{}{
// 							"actions": []string{"write"},
// 							"forwardings": []interface{}{
// 								map[string]interface{}{
// 									"actions": []string{"write"},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	claims := jwt.MapClaims{
// 		"jti":   uuid.New().String(),
// 		"iat":   now,
// 		"exp":   now + 60*60*24*365,
// 		"scope": scope,
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	return token.SignedString([]byte(secret))
// }
