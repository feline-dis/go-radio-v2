package services

import (
	"testing"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/config"
)

func TestJWTService_GenerateToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: time.Hour,
		},
	}

	jwtService := NewJWTService(cfg)

	token, err := jwtService.GenerateToken("testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	t.Logf("Generated token: %s", token)
}

func TestJWTService_ValidateToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: time.Hour,
		},
	}

	jwtService := NewJWTService(cfg)

	// Generate a token
	username := "testuser"
	token, err := jwtService.GenerateToken(username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate the token
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.Username != username {
		t.Fatalf("Expected username %s, got %s", username, claims.Username)
	}

	t.Logf("Token validated successfully for user: %s", claims.Username)
}

func TestJWTService_ValidateInvalidToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: time.Hour,
		},
	}

	jwtService := NewJWTService(cfg)

	// Try to validate an invalid token
	_, err := jwtService.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token, but got none")
	}

	t.Logf("Correctly rejected invalid token with error: %v", err)
}

func TestJWTService_RefreshToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: time.Hour,
		},
	}

	jwtService := NewJWTService(cfg)

	// Generate a token
	username := "testuser"
	originalToken, err := jwtService.GenerateToken(username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait a moment to ensure different timestamps
	time.Sleep(time.Second * 1)

	// Refresh the token
	refreshedToken, err := jwtService.RefreshToken(originalToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if refreshedToken == originalToken {
		t.Fatal("Refreshed token should be different from original")
	}

	// Validate the refreshed token
	claims, err := jwtService.ValidateToken(refreshedToken)
	if err != nil {
		t.Fatalf("Failed to validate refreshed token: %v", err)
	}

	if claims.Username != username {
		t.Fatalf("Expected username %s, got %s", username, claims.Username)
	}

	t.Logf("Token refreshed successfully for user: %s", claims.Username)
}

func TestJWTService_NoSecret(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "", // Empty secret
			Expiration: time.Hour,
		},
	}

	jwtService := NewJWTService(cfg)

	// Try to generate a token without secret
	_, err := jwtService.GenerateToken("testuser")
	if err == nil {
		t.Fatal("Expected error when JWT secret is not configured")
	}

	t.Logf("Correctly rejected token generation without secret: %v", err)
} 