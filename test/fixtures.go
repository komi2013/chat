package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chat/collection"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// LoadTestData loads test data from JSON files in test/data directory
func LoadTestData(testDB *mongo.Client) error {
	dataDir := "test/data"
	
	// Load users
	if err := loadJSONToCollection(testDB, dataDir, "users.json", "chatUser", "user", &[]collection.UserStruct{}); err != nil {
		return fmt.Errorf("failed to load users: %v", err)
	}
	
	// Load sessions
	if err := loadJSONToCollection(testDB, dataDir, "sessions.json", "chatSession", "session", &[]collection.SessionStruct{}); err != nil {
		return fmt.Errorf("failed to load sessions: %v", err)
	}
	
	return nil
}

// loadJSONToCollection loads JSON data into a MongoDB collection
func loadJSONToCollection(testDB *mongo.Client, dataDir, filename, dbName, collectionName string, target interface{}) error {
	filePath := filepath.Join(dataDir, filename)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}
	
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %v", filePath, err)
	}
	
	collection := testDB.Database(dbName).Collection(collectionName)
	
	switch v := target.(type) {
	case *[]collection.UserStruct:
		for _, user := range *v {
			_, err := collection.InsertOne(context.TODO(), user)
			if err != nil {
				return fmt.Errorf("failed to insert user %s: %v", user.UserID, err)
			}
		}
	case *[]collection.SessionStruct:
		for _, session := range *v {
			_, err := collection.InsertOne(context.TODO(), session)
			if err != nil {
				return fmt.Errorf("failed to insert session %s: %v", session.SessionID, err)
			}
		}
	}
	
	return nil
}

// CleanupTestDatabase cleans up all test data from MongoDB
func CleanupTestDatabase(testDB *mongo.Client) error {
	databases := []string{"chatUser", "chatSession", "chatNickname", "chatAd", "chatReception"}
	
	for _, dbName := range databases {
		collections, err := testDB.Database(dbName).ListCollectionNames(context.TODO(), nil)
		if err != nil {
			return fmt.Errorf("failed to list collections in %s: %v", dbName, err)
		}
		
		for _, collName := range collections {
			err := testDB.Database(dbName).Collection(collName).Drop(context.TODO())
			if err != nil {
				return fmt.Errorf("failed to drop collection %s.%s: %v", dbName, collName, err)
			}
		}
	}
	
	return nil
}

// CreateTestUser creates a test user with specified parameters
func CreateTestUser(userID, googleSub, email string) *collection.UserStruct {
	return &collection.UserStruct{
		UserID:       userID,
		GoogleJWTSub: googleSub,
		Mail:         email,
		Telephone:    "123-456-7890",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		SignedAt:     time.Now(),
		ChannelAliases: []collection.ChannelAlias{
			{ChannelID: "test-channel", Alias: "Test User", GuestFlag: false},
		},
		Latitude:     35.6762,
		Longitude:    139.6503,
		WalletAddress: "0x1234567890123456789012345678901234567890",
	}
}

// CreateTestSession creates a test session with specified parameters
func CreateTestSession(sessionID, userID, csrf string) *collection.SessionStruct {
	return &collection.SessionStruct{
		SessionID:    sessionID,
		UserID:       userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Csrf:         csrf,
		PushContents: []string{"test-content"},
		IsMobile:     false,
		Mail:         "test@example.com",
		Telephone:    "123-456-7890",
		Nickname:     "TestNick",
		NickImg:      "/img/test.png",
		TweetPosts:   []collection.TweetPost{},
	}
}

// InsertTestData inserts test data into the specified collection
func InsertTestData(testDB *mongo.Client, data interface{}, dbName, collectionName string) error {
	collection := testDB.Database(dbName).Collection(collectionName)
	_, err := collection.InsertOne(context.TODO(), data)
	return err
}

// FindUserByGoogleSub finds a user by Google JWT subject
func FindUserByGoogleSub(testDB *mongo.Client, googleSub string) (*collection.UserStruct, error) {
	var user collection.UserStruct
	err := testDB.Database("chatUser").Collection("user").FindOne(context.TODO(), bson.M{"googleJWTSub": googleSub}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindSessionByID finds a session by session ID
func FindSessionByID(testDB *mongo.Client, sessionID string) (*collection.SessionStruct, error) {
	var session collection.SessionStruct
	err := testDB.Database("chatSession").Collection("session").FindOne(context.TODO(), bson.M{"_id": sessionID}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}
