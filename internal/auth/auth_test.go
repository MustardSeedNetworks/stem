// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/krisarmstrong/stem/internal/auth"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		jwtSecret      string
		sessionTimeout time.Duration
		username       string
		password       string
		wantDuration   time.Duration
		wantErr        bool
	}{
		{
			name:           "missing credentials returns error",
			jwtSecret:      "",
			sessionTimeout: 0,
			username:       "",
			password:       "",
			wantDuration:   0,
			wantErr:        true,
		},
		{
			name:           "missing username returns error",
			jwtSecret:      "secret",
			sessionTimeout: 0,
			username:       "",
			password:       "password",
			wantDuration:   0,
			wantErr:        true,
		},
		{
			name:           "missing password returns error",
			jwtSecret:      "secret",
			sessionTimeout: 0,
			username:       "admin",
			password:       "",
			wantDuration:   0,
			wantErr:        true,
		},
		{
			name:           "custom values",
			jwtSecret:      "test-secret-key-at-least-32-chars",
			sessionTimeout: 15 * time.Minute,
			username:       "testuser",
			password:       "testpass",
			wantDuration:   15 * time.Minute,
			wantErr:        false,
		},
		{
			name:           "negative timeout uses default",
			jwtSecret:      "secret",
			sessionTimeout: -1 * time.Minute,
			username:       "admin",
			password:       "admin",
			wantDuration:   auth.AccessTokenDuration,
			wantErr:        false,
		},
		{
			name:           "auto-generates JWT secret",
			jwtSecret:      "",
			sessionTimeout: 0,
			username:       "admin",
			password:       "admin",
			wantDuration:   auth.AccessTokenDuration,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mgr, err := auth.NewManager(tt.jwtSecret, tt.sessionTimeout, tt.username, tt.password)
			if tt.wantErr {
				if err == nil {
					t.Fatal("NewManager() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewManager() unexpected error: %v", err)
			}
			if mgr == nil {
				t.Fatal("NewManager returned nil")
			}
			if got := mgr.SessionDuration(); got != tt.wantDuration {
				t.Errorf("SessionDuration() = %v, want %v", got, tt.wantDuration)
			}
		})
	}
}

func TestNewManager_MissingCredentialsError(t *testing.T) {
	t.Parallel()

	_, err := auth.NewManager("secret", auth.AccessTokenDuration, "", "")
	if !errors.Is(err, auth.ErrMissingCredentials) {
		t.Errorf("NewManager() error = %v, want %v", err, auth.ErrMissingCredentials)
	}
}

// mustNewManager creates a test auth manager or fails the test.
func mustNewManager(
	t *testing.T, jwtSecret string, sessionTimeout time.Duration, username, password string,
) *auth.Manager {
	t.Helper()
	mgr, err := auth.NewManager(jwtSecret, sessionTimeout, username, password)
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}
	return mgr
}

func TestAuthenticate_ValidCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "admin", "admin")
	if err != nil {
		t.Errorf("Authenticate() unexpected error: %v", err)
	}
	if token == "" {
		t.Error("Authenticate() returned empty token")
	}
}

func TestAuthenticate_CaseInsensitiveUsername(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "ADMIN", "admin")
	if err != nil {
		t.Errorf("Authenticate() unexpected error: %v", err)
	}
	if token == "" {
		t.Error("Authenticate() returned empty token")
	}
}

func TestAuthenticate_WrongUsername(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "wronguser", "admin")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("Authenticate() error = %v, want %v", err, auth.ErrInvalidCredentials)
	}
	if token != "" {
		t.Error("Authenticate() returned token on error")
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "admin", "wrongpass")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("Authenticate() error = %v, want %v", err, auth.ErrInvalidCredentials)
	}
	if token != "" {
		t.Error("Authenticate() returned token on error")
	}
}

func TestAuthenticate_EmptyUsername(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "", "admin")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("Authenticate() error = %v, want %v", err, auth.ErrInvalidCredentials)
	}
	if token != "" {
		t.Error("Authenticate() returned token on error")
	}
}

func TestAuthenticate_EmptyPassword(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "admin", "")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("Authenticate() error = %v, want %v", err, auth.ErrInvalidCredentials)
	}
	if token != "" {
		t.Error("Authenticate() returned token on error")
	}
}

func TestAuthenticate_PasswordCaseSensitive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "admin", "ADMIN")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("Authenticate() error = %v, want %v", err, auth.ErrInvalidCredentials)
	}
	if token != "" {
		t.Error("Authenticate() returned token on error")
	}
}

func TestAuthenticate_CustomCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "testuser", "testpassword123")

	token, err := mgr.Authenticate(ctx, "testuser", "testpassword123")
	if err != nil {
		t.Errorf("Authenticate() unexpected error: %v", err)
	}
	if token == "" {
		t.Error("Authenticate() returned empty token")
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "admin", "admin")
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	claims, err := mgr.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken() failed: %v", err)
	}

	if claims.Username != "admin" {
		t.Errorf("claims.Username = %q, want %q", claims.Username, "admin")
	}
	if claims.TokenType != "access" {
		t.Errorf("claims.TokenType = %q, want %q", claims.TokenType, "access")
	}
	if claims.Subject != "admin" {
		t.Errorf("claims.Subject = %q, want %q", claims.Subject, "admin")
	}
	if claims.Issuer != "The Stem" {
		t.Errorf("claims.Issuer = %q, want %q", claims.Issuer, "The Stem")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", 1*time.Millisecond, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "admin", "admin")
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = mgr.ValidateToken(ctx, token)
	if !errors.Is(err, auth.ErrTokenExpired) {
		t.Errorf("ValidateToken() error = %v, want %v", err, auth.ErrTokenExpired)
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	_, err := mgr.ValidateToken(ctx, "not-a-valid-jwt")
	if err == nil {
		t.Error("ValidateToken() expected error for invalid token")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	_, err := mgr.ValidateToken(ctx, "")
	if err == nil {
		t.Error("ValidateToken() expected error for empty token")
	}
}

func TestValidateToken_TamperedToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr.Authenticate(ctx, "admin", "admin")
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	tamperedToken := token[:len(token)-1] + "X"
	_, err = mgr.ValidateToken(ctx, tamperedToken)
	if err == nil {
		t.Error("ValidateToken() expected error for tampered token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr1 := mustNewManager(t, "secret-one", auth.AccessTokenDuration, "admin", "admin")
	mgr2 := mustNewManager(t, "secret-two", auth.AccessTokenDuration, "admin", "admin")

	token, err := mgr1.Authenticate(ctx, "admin", "admin")
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	_, err = mgr2.ValidateToken(ctx, token)
	if err == nil {
		t.Error("ValidateToken() expected error for token signed with different secret")
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	claims := jwt.MapClaims{
		"username":   "admin",
		"token_type": "access",
		"exp":        time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := mgr.ValidateToken(ctx, tokenString)
	if err == nil {
		t.Error("ValidateToken() expected error for wrong signing method")
	}
}

func TestSessionDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "default duration",
			timeout:  0,
			expected: auth.AccessTokenDuration,
		},
		{
			name:     "custom duration",
			timeout:  1 * time.Hour,
			expected: 1 * time.Hour,
		},
		{
			name:     "short duration",
			timeout:  5 * time.Minute,
			expected: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mgr := mustNewManager(t, "secret", tt.timeout, "admin", "admin")
			if got := mgr.SessionDuration(); got != tt.expected {
				t.Errorf("SessionDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateJWTSecret_Unique(t *testing.T) {
	t.Parallel()
	secret1, err1 := auth.GenerateJWTSecret()
	if err1 != nil {
		t.Fatalf("GenerateJWTSecret() error: %v", err1)
	}
	secret2, err2 := auth.GenerateJWTSecret()
	if err2 != nil {
		t.Fatalf("GenerateJWTSecret() error: %v", err2)
	}

	if secret1 == secret2 {
		t.Error("GenerateJWTSecret() generated identical secrets")
	}
}

func TestGenerateJWTSecret_NonEmpty(t *testing.T) {
	t.Parallel()
	secret, err := auth.GenerateJWTSecret()
	if err != nil {
		t.Fatalf("GenerateJWTSecret() error: %v", err)
	}
	if secret == "" {
		t.Error("GenerateJWTSecret() returned empty string")
	}
}

func TestGenerateJWTSecret_Base64URLEncoded(t *testing.T) {
	t.Parallel()
	secret, err := auth.GenerateJWTSecret()
	if err != nil {
		t.Fatalf("GenerateJWTSecret() error: %v", err)
	}
	for _, c := range secret {
		if !isBase64URLChar(c) {
			t.Errorf("GenerateJWTSecret() contains invalid character: %c", c)
		}
	}
}

func isBase64URLChar(c rune) bool {
	isUpperAlpha := c >= 'A' && c <= 'Z'
	isLowerAlpha := c >= 'a' && c <= 'z'
	isDigit := c >= '0' && c <= '9'
	isSpecial := c == '-' || c == '_'
	return isUpperAlpha || isLowerAlpha || isDigit || isSpecial
}

func TestGenerateJWTSecret_ExpectedLength(t *testing.T) {
	t.Parallel()
	secret, err := auth.GenerateJWTSecret()
	if err != nil {
		t.Fatalf("GenerateJWTSecret() error: %v", err)
	}
	const expectedLen = 43
	if len(secret) != expectedLen {
		t.Errorf("GenerateJWTSecret() length = %d, want %d", len(secret), expectedLen)
	}
}

func TestAuthenticateConcurrency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", auth.AccessTokenDuration, "admin", "admin")

	const numGoroutines = 100
	errChan := make(chan error, numGoroutines)

	for range numGoroutines {
		go func() {
			token, err := mgr.Authenticate(ctx, "admin", "admin")
			if err != nil {
				errChan <- err
				return
			}
			_, err = mgr.ValidateToken(ctx, token)
			errChan <- err
		}()
	}

	for range numGoroutines {
		err := <-errChan
		if err != nil {
			t.Errorf("concurrent operation failed: %v", err)
		}
	}
}

func TestTokenClaims_Username(t *testing.T) {
	t.Parallel()
	claims := getTestClaims(t)
	if claims.Username != "testuser" {
		t.Errorf("Username = %q, want %q", claims.Username, "testuser")
	}
}

func TestTokenClaims_Subject(t *testing.T) {
	t.Parallel()
	claims := getTestClaims(t)
	if claims.Subject != "testuser" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "testuser")
	}
}

func TestTokenClaims_TokenType(t *testing.T) {
	t.Parallel()
	claims := getTestClaims(t)
	if claims.TokenType != "access" {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, "access")
	}
}

func TestTokenClaims_Issuer(t *testing.T) {
	t.Parallel()
	claims := getTestClaims(t)
	if claims.Issuer != "The Stem" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "The Stem")
	}
}

func TestTokenClaims_ExpiresAt(t *testing.T) {
	t.Parallel()
	claims := getTestClaims(t)
	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt is nil")
	}
	if !claims.ExpiresAt.After(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestTokenClaims_IssuedAt(t *testing.T) {
	t.Parallel()
	claims := getTestClaims(t)
	if claims.IssuedAt == nil {
		t.Fatal("IssuedAt is nil")
	}
	if claims.IssuedAt.After(time.Now()) {
		t.Error("IssuedAt should be in the past or present")
	}
}

func TestTokenClaims_NotBefore(t *testing.T) {
	t.Parallel()
	claims := getTestClaims(t)
	if claims.NotBefore == nil {
		t.Fatal("NotBefore is nil")
	}
	if claims.NotBefore.After(time.Now()) {
		t.Error("NotBefore should be in the past or present")
	}
}

func getTestClaims(t *testing.T) *auth.Claims {
	t.Helper()
	ctx := context.Background()
	mgr := mustNewManager(t, "test-secret", 10*time.Minute, "testuser", "testpass")

	token, err := mgr.Authenticate(ctx, "testuser", "testpass")
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	claims, err := mgr.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken() failed: %v", err)
	}
	return claims
}

func TestErrInvalidCredentials(t *testing.T) {
	t.Parallel()
	if auth.ErrInvalidCredentials == nil {
		t.Error("ErrInvalidCredentials is nil")
	}
	if !strings.Contains(auth.ErrInvalidCredentials.Error(), "invalid") {
		t.Error("ErrInvalidCredentials should contain 'invalid'")
	}
}

func TestErrInvalidToken(t *testing.T) {
	t.Parallel()
	if auth.ErrInvalidToken == nil {
		t.Error("ErrInvalidToken is nil")
	}
	if !strings.Contains(auth.ErrInvalidToken.Error(), "invalid") {
		t.Error("ErrInvalidToken should contain 'invalid'")
	}
}

func TestErrTokenExpired(t *testing.T) {
	t.Parallel()
	if auth.ErrTokenExpired == nil {
		t.Error("ErrTokenExpired is nil")
	}
	if !strings.Contains(auth.ErrTokenExpired.Error(), "expired") {
		t.Error("ErrTokenExpired should contain 'expired'")
	}
}

func TestAccessTokenDurationConstant(t *testing.T) {
	t.Parallel()
	// Access tokens are now 15 minutes for better security (with refresh tokens for extended sessions).
	if auth.AccessTokenDuration != 15*time.Minute {
		t.Errorf("AccessTokenDuration = %v, want %v", auth.AccessTokenDuration, 15*time.Minute)
	}
}
