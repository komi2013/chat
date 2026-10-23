package common

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	GoPort string
	CacheV string

	// PostgreSQL connection strings
	PgChannel   string
	PgAlias     string
	PgGroup     string
	PgFile      string
	PgInvoice   string
	PgReception string
	PgSession   string
	PgUser      string

	VAPIDPublicKey            string
	VAPIDPrivateKey           string
	OSImgDir                  string
	PublicImgPath             string
	UploadDir                 string
	Domain                    string
	GoogleClientID            string
	GoogleClientMobileID      string
	EtherscanApiKey           string
	WebRTC                    string
	FirebaseProjectID         string
	FirebaseServiceAccountKey string
	FCMTestToken              string
	FCMAccessToken            string
}

func LoadConfig() *Config {
	const pathRelative = "infrastructure/.env"

	if err := godotenv.Load(pathRelative); err != nil {
		log.Fatalf(
			"FATAL: Console Config loading failed! Required path not found: %s. Error: %v",
			pathRelative,
			err,
		)
	}

	return loadFromEnv()
}

// LoadConsoleConfig loads environment variables for console tasks.
func LoadConsoleConfig() *Config {
	prodAbsolute := "/Work/chat/infrastructure/.env"
	devAbsolute := "/root/infrastructure/.env"

	if err := godotenv.Load(prodAbsolute); err != nil {
		if err := godotenv.Load(devAbsolute); err != nil {
			log.Fatalf(
				"Console .env loading failed; prod:<%s> dev:<%s> err: %v",
				prodAbsolute,
				devAbsolute,
				err,
			)
		}
	}

	return loadFromEnv()
}

// loadFromEnv reads environment variables into Config.
func loadFromEnv() *Config {
	return &Config{
		GoPort: os.Getenv("GO_PORT"),
		CacheV:  fixCache(),

		// PostgreSQL
		PgChannel:   os.Getenv("PG_CHANNEL"),
		PgAlias:     os.Getenv("PG_ALIAS"),
		PgGroup:     os.Getenv("PG_GROUP"),
		PgFile:      os.Getenv("PG_FILE"),
		PgInvoice:   os.Getenv("PG_INVOICE"),
		PgReception: os.Getenv("PG_RECEPTION"),
		PgSession:   os.Getenv("PG_SESSION"),
		PgUser:      os.Getenv("PG_USER"),

		VAPIDPublicKey:            "BIN2Jc5Vmkmy-S3AUrcMlpKxJpLeVRAfu9WBqUbJ70SJOCWGCGXKY-Xzyh7HDr6KbRDGYHjqZ06OcS3BjD7uAm8",
		VAPIDPrivateKey:           os.Getenv("VAPID_PRIVATE_KEY"),
		OSImgDir:                  os.Getenv("OS_IMG_DIR"),
		PublicImgPath:             os.Getenv("PUBLIC_IMG_PATH"),
		UploadDir:                 os.Getenv("UPLOAD_DIR"),
		Domain:                    os.Getenv("DOMAIN"),
		GoogleClientID:            os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientMobileID:      os.Getenv("GOOGLE_CLIENT_MOBILE_ID"),
		EtherscanApiKey:           os.Getenv("ETHER_SCAN_APIKEY"),
		WebRTC:                    os.Getenv("WEB_RTC"),
		FirebaseProjectID:         os.Getenv("FIREBASE_PROJECT_ID"),
		FirebaseServiceAccountKey: os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY"),
		FCMTestToken:              os.Getenv("FCM_TEST_TOKEN"),
		FCMAccessToken:            os.Getenv("FCM_ACCESS_TOKEN"),
	}
}

func fixCache() string {
	f, err := os.Open("infrastructure/cache_v.env")
	if err != nil {
		log.Fatalf("FATAL: cache_v.env not found: %v", err)
	}
	defer f.Close()

	envMap, err := godotenv.Parse(f)
	if err != nil {
		log.Fatalf("FATAL: cache_v.env parse error: %v", err)
	}

	cacheStr := envMap["CACHE_V"]

	if os.Getenv("ENV") == "dev" {
		cacheStr = time.Now().Format("20060102T1504")
	}

	return cacheStr
}
