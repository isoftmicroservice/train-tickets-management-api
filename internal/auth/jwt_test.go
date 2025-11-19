package auth

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

func TestExtractUserFromContext(t *testing.T) {
	// Create a simple JWT token (header.payload.signature format)
	// For testing, we'll create a minimal valid JWT structure
	// Note: ParseUnverified doesn't validate signature, so we just need valid structure
	claims := jwt.MapClaims{
		"email":      "test@example.com",
		"first_name": "John",
		"last_name":  "Doe",
		"role":       "user",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// For testing, we'll use a dummy secret to sign
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Create context with metadata
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + tokenString,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// Test extraction
	userClaims, err := ExtractUserFromContext(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if userClaims.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", userClaims.Email)
	}

	if userClaims.FirstName != "John" {
		t.Errorf("Expected first name John, got %s", userClaims.FirstName)
	}

	if userClaims.Role != "user" {
		t.Errorf("Expected role user, got %s", userClaims.Role)
	}
}

func TestExtractUserFromContext_Admin(t *testing.T) {
	claims := jwt.MapClaims{
		"email":      "admin@example.com",
		"first_name": "Admin",
		"last_name":  "User",
		"role":       "admin",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + tokenString,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	userClaims, err := ExtractUserFromContext(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !userClaims.IsAdmin() {
		t.Error("Expected user to be admin")
	}
}

func TestExtractUserFromContext_NoMetadata(t *testing.T) {
	ctx := context.Background()

	_, err := ExtractUserFromContext(ctx)
	if err != ErrNoMetadata {
		t.Errorf("Expected ErrNoMetadata, got: %v", err)
	}
}

func TestExtractUserFromContext_NoAuthHeader(t *testing.T) {
	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := ExtractUserFromContext(ctx)
	if err != ErrNoAuthHeader {
		t.Errorf("Expected ErrNoAuthHeader, got: %v", err)
	}
}

func TestExtractUserFromContext_DefaultRole(t *testing.T) {
	claims := jwt.MapClaims{
		"email": "test@example.com",
		// No role specified
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + tokenString,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	userClaims, err := ExtractUserFromContext(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if userClaims.Role != "user" {
		t.Errorf("Expected default role 'user', got %s", userClaims.Role)
	}
}

