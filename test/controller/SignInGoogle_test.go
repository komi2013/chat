package controller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"chat/collection"
	"chat/common"
	"chat/test"
)

// Mock Google JWT token for testing
func createMockGoogleJWT(googleSub, email string) (string, error) {
	// Create a mock RSA key pair for testing
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(mockRSAPrivateKey))
	if err != nil {
		return "", err
	}

	// Create token claims
	claims := jwt.MapClaims{
		"sub":   googleSub,
		"email": email,
		"iss":   "https://accounts.google.com",
		"aud":   "test-google-client-id",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	// Create token with mock key
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"

	return token.SignedString(privateKey)
}

// Mock RSA private key for testing (in real implementation, this would be properly generated)
const mockRSAPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890
abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnop
qrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABC
DEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOP
QRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234
567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdef
ghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz
ABCDEFGHIJKLMNOPQRSTUVWXYZwIDAQAB
-----END RSA PRIVATE KEY-----`

// Mock public key function for testing
func mockGetGooglePublicKey(kid string) (interface{}, error) {
	if kid == "test-key-id" {
		// Return mock public key that matches the private key above
		publicKey, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(mockRSAPublicKey))
		return publicKey, nil
	}
	return nil, errors.New("unknown key id")
}

const mockRSAPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1234567890abcdefghijklmnopqrstuvwxyz
ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnop
qrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABC
DEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOP
QRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234
567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdef
ghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz
ABCDEFGHIJKLMNOPQRSTUVWXYZwIDAQAB
-----END PUBLIC KEY-----`

func TestSignInGoogle_NewUser(t *testing.T) {
	// Setup test environment
	cleanup := test.SetupTestWithCleanup()
	defer cleanup()

	// Create mock JWT token for new user
	googleSub := "new-google-sub-123"
	email := "newuser@example.com"
	mockToken, err := createMockGoogleJWT(googleSub, email)
	if err != nil {
		t.Fatalf("Failed to create mock JWT: %v", err)
	}

	// Prepare form data
	form := url.Values{}
	form.Add("g_csrf_token", "test-csrf-value")
	form.Add("credential", mockToken)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/SignInGoogle/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "g_csrf_token", Value: "test-csrf-value"})

	// Create response recorder
	w := httptest.NewRecorder()

	// Mock the getGooglePublicKey function for testing
	// In a real test, you might use dependency injection or a test double
	originalGetGooglePublicKey := getGooglePublicKey
	getGooglePublicKey = func(kid string) (*rsa.PublicKey, error) {
		if kid == "test-key-id" {
			pubKey, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(mockRSAPublicKey))
			return pubKey, nil
		}
		return nil, errors.New("unknown key id")
	}
	defer func() {
		getGooglePublicKey = originalGetGooglePublicKey
	}()

	// Execute the controller function
	SignInGoogle(w, req)

	// Check response - should redirect to pushSubscription
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}

	// Verify location header
	location := w.Header().Get("Location")
	if location != "/pushSubscription/" {
		t.Errorf("Expected redirect to /pushSubscription/, got %s", location)
	}

	// Verify user was created in database
	user, err := test.FindUserByGoogleSub(test.TestDB, googleSub)
	if err != nil {
		t.Fatalf("Failed to find created user: %v", err)
	}

	if user.Mail != email {
		t.Errorf("Expected email %s, got %s", email, user.Mail)
	}

	if user.GoogleJWTSub != googleSub {
		t.Errorf("Expected Google sub %s, got %s", googleSub, user.GoogleJWTSub)
	}

	// Verify session was created
	sessions, err := test.TestDB.Database("chatSession").Collection("session").Find(context.TODO(), map[string]interface{}{
		"userID": user.UserID,
	})
	if err != nil {
		t.Fatalf("Failed to find sessions: %v", err)
	}
	defer sessions.Close(context.TODO())

	var sessionList []collection.SessionStruct
	if err = sessions.All(context.TODO(), &sessionList); err != nil {
		t.Fatalf("Failed to decode sessions: %v", err)
	}

	if len(sessionList) == 0 {
		t.Error("Expected at least one session to be created")
	}

	// Verify cookie was set
	cookies := w.Result().Cookies()
	var ssCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "ss" {
			ssCookie = cookie
			break
		}
	}

	if ssCookie == nil {
		t.Error("Expected ss cookie to be set")
	}

	if ssCookie != nil && ssCookie.Value == "" {
		t.Error("Expected ss cookie to have a value")
	}
}

func TestSignInGoogle_ExistingUserWithSession(t *testing.T) {
	// Setup test environment
	cleanup := test.SetupTestWithCleanup()
	defer cleanup()

	// Create existing user
	existingUser := test.CreateTestUser("existing-user-1", "existing-google-sub-456", "existing@example.com")
	err := test.InsertTestData(test.TestDB, existingUser, "chatUser", "user")
	if err != nil {
		t.Fatalf("Failed to insert existing user: %v", err)
	}

	// Create existing session
	existingSession := test.CreateTestSession("existing-session-1", "existing-user-1", "existing-csrf")
	err = test.InsertTestData(test.TestDB, existingSession, "chatSession", "session")
	if err != nil {
		t.Fatalf("Failed to insert existing session: %v", err)
	}

	// Create mock JWT token for existing user
	mockToken, err := createMockGoogleJWT("existing-google-sub-456", "existing@example.com")
	if err != nil {
		t.Fatalf("Failed to create mock JWT: %v", err)
	}

	// Prepare form data
	form := url.Values{}
	form.Add("g_csrf_token", "test-csrf-value")
	form.Add("credential", mockToken)

	// Create HTTP request with existing session cookie
	req := httptest.NewRequest("POST", "/SignInGoogle/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.AddCookie(&http.Cookie{Name: "g_csrf_token", Value: "test-csrf-value"})
	req.AddCookie(&http.Cookie{Name: "ss", Value: "existing-session-1"})

	// Create response recorder
	w := httptest.NewRecorder()

	// Mock the getGooglePublicKey function for testing
	originalGetGooglePublicKey := getGooglePublicKey
	getGooglePublicKey = func(kid string) (*rsa.PublicKey, error) {
		if kid == "test-key-id" {
			pubKey, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(mockRSAPublicKey))
			return pubKey, nil
		}
		return nil, errors.New("unknown key id")
	}
	defer func() {
		getGooglePublicKey = originalGetGooglePublicKey
	}()

	// Execute the controller function
	SignInGoogle(w, req)

	// Check response - should redirect to pushSubscription
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}

	// Verify old session was deleted (same browser case)
	oldSession, err := test.FindSessionByID(test.TestDB, "existing-session-1")
	if err == nil {
		t.Error("Expected old session to be deleted, but it still exists")
	}

	// Verify new session was created
	newSessions, err := test.TestDB.Database("chatSession").Collection("session").Find(context.TODO(), map[string]interface{}{
		"userID": "existing-user-1",
	})
	if err != nil {
		t.Fatalf("Failed to find new sessions: %v", err)
	}
	defer newSessions.Close(context.TODO())

	var newSessionList []collection.SessionStruct
	if err = newSessions.All(context.TODO(), &newSessionList); err != nil {
		t.Fatalf("Failed to decode new sessions: %v", err)
	}

	// Should have exactly 1 new session (old one deleted)
	if len(newSessionList) != 1 {
		t.Errorf("Expected 1 new session, got %d", len(newSessionList))
	}

	// Verify new session inherits data from old session
	newSession := newSessionList[0]
	if newSession.Nickname != existingSession.Nickname {
		t.Errorf("Expected nickname %s, got %s", existingSession.Nickname, newSession.Nickname)
	}

	if newSession.Mail != existingSession.Mail {
		t.Errorf("Expected mail %s, got %s", existingSession.Mail, newSession.Mail)
	}
}

func TestSignInGoogle_InvalidCSRF(t *testing.T) {
	// Setup test environment
	cleanup := test.SetupTestWithCleanup()
	defer cleanup()

	// Prepare form data with invalid CSRF
	form := url.Values{}
	form.Add("g_csrf_token", "invalid-csrf")
	form.Add("credential", "mock-token")

	// Create HTTP request with different CSRF cookie
	req := httptest.NewRequest("POST", "/SignInGoogle/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "g_csrf_token", Value: "different-csrf"})

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute the controller function
	SignInGoogle(w, req)

	// Should return error for CSRF mismatch
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestSignInGoogle_MissingCSRFCookie(t *testing.T) {
	// Setup test environment
	cleanup := test.SetupTestWithCleanup()
	defer cleanup()

	// Prepare form data
	form := url.Values{}
	form.Add("g_csrf_token", "test-csrf")
	form.Add("credential", "mock-token")

	// Create HTTP request without CSRF cookie
	req := httptest.NewRequest("POST", "/SignInGoogle/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute the controller function
	SignInGoogle(w, req)

	// Should return error for missing CSRF cookie
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}
