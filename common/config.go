package common

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// const (
// 	EnvFilePathRelative = "infrastructure/.env"
// 	EnvFilePathProdAbsolute = "/Work/chat/infrastructure/.env"
// 	EnvFilePathDevAbsolute = "/chat/infrastructure/.env"
// )

// const 

type Config struct {
	GoPort          string
	CacheV          string
	SsKey           string
	T1Key           string
	CsrfKey         string
	MongoAd         string
	MongoAdPrice    string
	MongoChannel    string
	MongoFile       string
	MongoInvoice    string
	MongoNickname   string
	MongoReception  string
	MongoSequence   string
	MongoSession    string
	MongoTweet      string
	MongoUser       string
	DateLanguage    string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	OSImgDir        string
	PublicImgPath   string
	UploadDir       string
	Domain          string
	GoogleClientID  string
	EtherscanApiKey string
}


func LoadConfig() *Config {
	const pathRelative = "infrastructure/.env"
	err := godotenv.Load(pathRelative)
	if err != nil {
		log.Fatalf("FATAL: Console Config loading failed! Required path not found: %s. Error: %v", pathRelative, err)
	}
	return loadFromEnv()
}

// -----------------------------------------------------------
// LoadConsoleConfig() : コンソールタスク (Crontab/CLI) の環境変数を読み込む
// Policy: STRICTLY use the Absolute Path for maximum Crontab reliability.
// -----------------------------------------------------------
func LoadConsoleConfig() *Config {

	prodAbsolute := "/Work/chat/infrastructure/.env"
	devAbsolute := "/root/infrastructure/.env"
	err := godotenv.Load(prodAbsolute)
	if err != nil {
		err = godotenv.Load(devAbsolute)
		if err != nil {
			log.Fatalf("Console .env loading failed; prod:<%s> dev:<%s> err: %v", prodAbsolute, devAbsolute, err)
		}
	// } else {
	// 	log.Printf("DEBUG: EnvFilePathProdAbsolute path failed (%v). Attempting dev path: %s", err, EnvFilePathDevAbsolute)
	// 	err = godotenv.Load(EnvFilePathDevAbsolute)
		// if err != nil {
		// 	// Both attempts failed. Issue a final warning.
		// 	log.Printf("WARNING: App Configuration file not found. Relying on existing environment variables.")
		// } else {
		// 	log.Println("DEBUG: App Config loaded from Relative Path:", EnvFilePathRelative)
		// }
	}
	return loadFromEnv()
}

// Helper function to read variables from the environment into the Config struct
func loadFromEnv() *Config {
	return &Config{
		GoPort:          os.Getenv("GO_PORT"),
		CacheV:          fixCache(),
		SsKey:           os.Getenv("SS_KEY"),
		T1Key:           os.Getenv("T1_KEY"),
		CsrfKey:         os.Getenv("CSRF_KEY"),
		MongoAd:         os.Getenv("MONGO_AD"),
		MongoAdPrice:    os.Getenv("MONGO_AD_PRICE"),
		MongoChannel:    os.Getenv("MONGO_CHANNEL"),
		MongoFile:       os.Getenv("MONGO_FILE"),
		MongoInvoice:    os.Getenv("MONGO_INVOICE"),
		MongoNickname:   os.Getenv("MONGO_NICKNAME"),
		MongoReception:  os.Getenv("MONGO_RECEPTION"),
		MongoSequence:   os.Getenv("MONGO_SEQUENCE"),
		MongoSession:    os.Getenv("MONGO_SESSION"),
		MongoTweet:      os.Getenv("MONGO_TWEET"),
		MongoUser:       os.Getenv("MONGO_USER"),
		DateLanguage:    os.Getenv("DATE_LANGUAGE"),
		VAPIDPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		OSImgDir:        os.Getenv("OS_IMG_DIR"),
		PublicImgPath:   os.Getenv("PUBLIC_IMG_PATH"),
		UploadDir:       os.Getenv("UPLOAD_DIR"),
		Domain:          os.Getenv("DOMAIN"),
		GoogleClientID:  os.Getenv("GOOGLE_CLIENT_ID"),
		EtherscanApiKey: os.Getenv("ETHER_SCAN_APIKEY"),
	}
}

func fixCache() string {
	err := godotenv.Load("infrastructure/cache_v.env")
	if err != nil {
		log.Fatalf("FATAL: Console Config loading failed! Required path not found: %s. Error: %v", pathRelative, err)
	}
	cacheStr := os.Getenv("CACHE_V")
	if os.Getenv("ENV") == "dev" {
		cacheStr = time.Now().Format("20060102T1504")
	}
	// os.Getenv("CACHE_V") + time.Now().Format("2006-01-02T15:04"),
	return cacheStr
}

