package test

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"chat/common"
)

var TestDB *mongo.Client
var TestConfig *common.Config

// SetupTestEnvironment initializes the test environment
func SetupTestEnvironment() error {
	// Load test configuration
	TestConfig = loadTestConfig()
	
	// Connect to MongoDB
	if err := connectTestMongoDB(); err != nil {
		return fmt.Errorf("failed to connect to test MongoDB: %v", err)
	}
	
	// Initialize common.DB with test connection
	initTestDB()
	
	return nil
}

// loadTestConfig loads configuration for testing
func loadTestConfig() *common.Config {
	// Use environment variable MONGO_URL or default to localhost
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://test:test@localhost:27017"
	}
	
	return &common.Config{
		GoPort:          ":8080",
		CacheV:          "test-cache",
		MongoAd:         mongoURL,
		MongoAdPrice:    mongoURL,
		MongoChannel:    mongoURL,
		MongoFile:       mongoURL,
		MongoInvoice:    mongoURL,
		MongoNickname:   mongoURL,
		MongoReception:  mongoURL,
		MongoSequence:   mongoURL,
		MongoSession:    mongoURL,
		MongoTweet:      mongoURL,
		MongoUser:       mongoURL,
		VAPIDPublicKey:  "test-public-key",
		VAPIDPrivateKey: "test-private-key",
		OSImgDir:        "/tmp/test-img",
		PublicImgPath:   "/public/test-img",
		UploadDir:       "/tmp/test-upload",
		Domain:          "localhost:8080",
		GoogleClientID:  "test-google-client-id",
		EtherscanApiKey: "test-etherscan-key",
		WebRTC:          "test-webrtc",
	}
}

// connectTestMongoDB connects to the test MongoDB instance
func connectTestMongoDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(TestConfig.MongoUser))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}
	
	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}
	
	TestDB = client
	return nil
}

// initTestDB initializes the common.DB with test connections
func initTestDB() {
	common.DB = &common.DBs{
		AdDB:        TestDB.Database("chatAd"),
		AdPriceDB:   TestDB.Database("chatAdPrice"),
		ChannelDB:   TestDB.Database("chatChannel"),
		FileDB:      TestDB.Database("chatFile"),
		InvoiceDB:   TestDB.Database("chatInvoice"),
		NicknameDB:  TestDB.Database("chatNickname"),
		ReceptionDB: TestDB.Database("chatReception"),
		SequenceDB:  TestDB.Database("chatSequence"),
		SessionDB:   TestDB.Database("chatSession"),
		TweetDB:     TestDB.Database("chatTweet"),
		UserDB:      TestDB.Database("chatUser"),
	}
}

// CleanupTestEnvironment cleans up the test environment
func CleanupTestEnvironment() {
	if TestDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Disconnect from MongoDB
		TestDB.Disconnect(ctx)
		TestDB = nil
	}
}

// SetupTestWithCleanup sets up test environment and returns cleanup function
func SetupTestWithCleanup() func() {
	if err := SetupTestEnvironment(); err != nil {
		panic(fmt.Sprintf("Failed to setup test environment: %v", err))
	}
	
	return func() {
		CleanupTestDatabase(TestDB)
		CleanupTestEnvironment()
	}
}

// IsTestEnvironment returns true if running in test environment
func IsTestEnvironment() bool {
	return os.Getenv("GO_TEST") == "1" || TestConfig != nil
}
