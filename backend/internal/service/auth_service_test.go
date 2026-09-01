package service_test

import (
	"testing"

	"github.com/taskflow/backend/internal/config"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

func TestAuthService_FullFlow(t *testing.T) {
	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("InitTestDB failed: %v", err)
	}

	userRepo := repository.NewUserRepository(testDB)
	jwtCfg := &config.JWTConfig{
		Secret:              "super-secret-jwt-key-for-testing-only-123456789",
		AccessExpiryMinutes: 15,
		RefreshExpiryDays:   7,
	}
	authService := service.NewAuthService(userRepo, jwtCfg)

	// 1. Register new user
	user, access, refresh, err := authService.Register("Alice Engineer", "alice@example.com", "securepassword123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user.Email != "alice@example.com" || user.Name != "Alice Engineer" {
		t.Errorf("Unexpected user: %+v", user)
	}
	if access == "" || refresh == "" {
		t.Fatal("Expected non-empty token pair")
	}

	// 2. Prevent duplicate registration
	_, _, _, err = authService.Register("Alice Fake", "alice@example.com", "otherpassword")
	if err != service.ErrUserAlreadyExists {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}

	// 3. Login with correct password
	loginUser, loginAccess, loginRefresh, err := authService.Login("alice@example.com", "securepassword123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loginUser.ID != user.ID {
		t.Errorf("Logged in user ID mismatch: got %v, want %v", loginUser.ID, user.ID)
	}
	if loginAccess == "" || loginRefresh == "" {
		t.Fatal("Expected non-empty tokens on login")
	}

	// 4. Login with wrong password
	_, _, _, err = authService.Login("alice@example.com", "wrongpassword")
	if err != service.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}

	// 5. Validate Token
	claims, err := authService.ValidateToken(access)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != user.ID || claims.Email != user.Email {
		t.Errorf("Claims mismatch: got %+v", claims)
	}

	// 6. Refresh Token rotation
	newAccess, newRefresh, err := authService.RefreshToken(refresh)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if newAccess == "" || newRefresh == "" {
		t.Fatal("Expected new tokens from refresh")
	}
}
