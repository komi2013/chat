package common

import (
    // "log"
    "os"
    "time"
)

type Config struct {
    GoPort           string
    CacheV           string
    SsKey            string
    T1Key            string
    CsrfKey          string
    MongoAd          string
    MongoAdPrice     string
    MongoChannel     string
    MongoFile        string
    MongoInvoice     string
    MongoNickname    string
    MongoReception   string
    MongoSequence    string
    MongoSession     string
    MongoTweet       string
    MongoUser        string
    DateLanguage     string
    VAPIDPublicKey   string
    VAPIDPrivateKey  string
    OSImgDir         string
    PublicImgPath    string
    UploadDir        string
    Domain           string
    GoogleClientID   string
}

// -----------------------------------------------------------
// LoadConfig() : すべての環境変数を読み込む
// -----------------------------------------------------------
func LoadConfig() *Config {

    cfg := &Config{
        GoPort:          os.Getenv("GO_PORT"),
        CacheV:          os.Getenv("CACHE_V") + time.Now().Format("2006-01-02T15:04"),
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
    }

    return cfg
}
