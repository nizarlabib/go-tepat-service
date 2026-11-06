package spesial

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func setupJWTTest(t *testing.T) {
	t.Helper()
	viper.Set("JWT_KEY", "test-secret-key-for-testing-only")
	viper.Set("APP_ENV", "test")
	viper.Set("APP_SERVICE_NAME", "test-service")
	viper.Set("APP_VERSION", "1.0.0")
	viper.Set("JWT_ACCESS_TOKEN_EXPIRY", 60)
	viper.Set("JWT_REFRESH_TOKEN_EXPIRY", 10080)
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	setupJWTTest(t)

	payload := &Payload{
		ID:             1,
		RoleID:         uuid.New(),
		Role:           "user",
		SessionID:      uuid.NewString(),
		PermissionMask: "test",
	}

	tokenStr, err := GenerateAccessToken(payload)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.ID != payload.ID {
		t.Errorf("expected ID %d, got %d", payload.ID, claims.ID)
	}
	if claims.Type != PayloadTypeAccess {
		t.Errorf("expected PayloadTypeAccess, got %v", claims.Type)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	setupJWTTest(t)

	payload := &Payload{
		ID:             1,
		RoleID:         uuid.New(),
		Role:           "user",
		SessionID:      uuid.NewString(),
		PermissionMask: "test",
	}

	tokenStr, err := GenerateRefreshToken(payload)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	claims := &Payload{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, GetJWTKeyFunc())
	if err != nil || !token.Valid {
		t.Fatalf("failed to parse refresh token: %v", err)
	}
	if claims.Type != PayloadTypeRefresh {
		t.Errorf("expected PayloadTypeRefresh, got %v", claims.Type)
	}
}

func TestValidateTokenInvalidSignature(t *testing.T) {
	setupJWTTest(t)

	viper.Set("JWT_KEY", "wrong-key")

	payload := &Payload{
		ID:        1,
		RoleID:    uuid.New(),
		Role:      "user",
		SessionID: uuid.NewString(),
	}
	tokenStr, err := GenerateAccessToken(payload)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	viper.Set("JWT_KEY", "test-secret-key-for-testing-only")

	_, err = ValidateToken(tokenStr)
	if err == nil {
		t.Error("expected error for token signed with wrong key")
	}
}

func TestValidateTokenMissingToken(t *testing.T) {
	setupJWTTest(t)

	_, err := ValidateToken("")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestValidateTokenMalformedToken(t *testing.T) {
	setupJWTTest(t)

	_, err := ValidateToken("not-a-valid-token")
	if err == nil {
		t.Error("expected error for malformed token")
	}
}

func TestValidateTokenExpiredToken(t *testing.T) {
	setupJWTTest(t)

	viper.Set("JWT_ACCESS_TOKEN_EXPIRY", -1)

	payload := &Payload{
		ID:        1,
		RoleID:    uuid.New(),
		Role:      "user",
		SessionID: uuid.NewString(),
	}
	tokenStr, err := GenerateAccessToken(payload)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	viper.Set("JWT_ACCESS_TOKEN_EXPIRY", 15)

	_, err = ValidateToken(tokenStr)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestGetJWTKeyFuncRejectsNonHS256(t *testing.T) {
	setupJWTTest(t)

	keyFunc := GetJWTKeyFunc()

	// Create a token with a non-HS256 signing method (none algorithm)
	token := jwt.New(jwt.SigningMethodNone)
	_, err := keyFunc(token)
	if err == nil {
		t.Error("expected error for non-HS256 signing method")
	}
	if !errors.Is(err, errors.New("unexpected signing method")) && err.Error() != "unexpected signing method" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseBearerToken(t *testing.T) {
	setupJWTTest(t)

	// Valid bearer token
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer abc123")
	token, err := ParseBearerToken(req)
	if err != nil {
		t.Fatalf("ParseBearerToken failed: %v", err)
	}
	if token != "abc123" {
		t.Errorf("expected 'abc123', got '%s'", token)
	}

	// Missing authorization header
	req2 := httptest.NewRequest("GET", "/", nil)
	_, err = ParseBearerToken(req2)
	if err == nil {
		t.Error("expected error for missing Authorization header")
	}

	// Malformed authorization header
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Header.Set("Authorization", "InvalidFormat")
	_, err = ParseBearerToken(req3)
	if err == nil {
		t.Error("expected error for malformed Authorization header")
	}
}

func TestPayloadTypeConstants(t *testing.T) {
	if PayloadTypeAccess != 0 {
		t.Errorf("expected PayloadTypeAccess=0, got %d", PayloadTypeAccess)
	}
	if PayloadTypeRefresh != 1 {
		t.Errorf("expected PayloadTypeRefresh=1, got %d", PayloadTypeRefresh)
	}
}

func TestBcryptPasswordHashing(t *testing.T) {
	password := "test-password-123"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword failed: %v", err)
	}

	err = bcrypt.CompareHashAndPassword(hashed, []byte(password))
	if err != nil {
		t.Errorf("bcrypt.CompareHashAndPassword failed for correct password: %v", err)
	}

	err = bcrypt.CompareHashAndPassword(hashed, []byte("wrong-password"))
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestValidateTokenAccessTypeEnforced(t *testing.T) {
	setupJWTTest(t)

	// Generate a refresh token and try to validate it as access - should fail
	payload := &Payload{
		ID:             1,
		RoleID:         uuid.New(),
		Role:           "user",
		SessionID:      uuid.NewString(),
		PermissionMask: "test",
	}
	tokenStr, err := GenerateRefreshToken(payload)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	_, err = ValidateToken(tokenStr)
	if err == nil {
		t.Error("expected error when refresh token validated as access token")
	}
}

var _ = httptest.NewRequest
