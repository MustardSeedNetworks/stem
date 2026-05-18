// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
	"time"
)

// Internal tests that can access unexported functions.

func TestGenerateTokenID_Success(t *testing.T) {
	t.Parallel()

	// Create a manager to access generateTokenID method.
	mgr, mgrErr := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if mgrErr != nil {
		t.Fatalf("NewManager() error: %v", mgrErr)
	}

	// Generate multiple token IDs and verify they are unique.
	ids := make(map[string]bool)
	for range 100 {
		id, err := mgr.generateTokenID()
		if err != nil {
			t.Fatalf("generateTokenID() error: %v", err)
		}
		if id == "" {
			t.Error("generateTokenID() returned empty string")
		}
		if ids[id] {
			t.Errorf("generateTokenID() generated duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestGenerateTokenID_Length(t *testing.T) {
	t.Parallel()

	// Create a manager to access generateTokenID method.
	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	id, err := mgr.generateTokenID()
	if err != nil {
		t.Fatalf("generateTokenID() error: %v", err)
	}

	// Token ID is base64url encoded 16 bytes = 22 characters.
	const expectedLen = 22
	if len(id) != expectedLen {
		t.Errorf("generateTokenID() length = %d, want %d", len(id), expectedLen)
	}
}

func TestGenerateTokenWithType_Access(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	token, err := mgr.generateTokenWithType("testuser", "access", 10*time.Minute)
	if err != nil {
		t.Fatalf("generateTokenWithType() error: %v", err)
	}

	claims, err := mgr.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken() error: %v", err)
	}

	if claims.TokenType != "access" {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, "access")
	}
	if claims.Username != "testuser" {
		t.Errorf("Username = %q, want %q", claims.Username, "testuser")
	}
}

func TestGenerateTokenWithType_Refresh(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	token, err := mgr.generateTokenWithType("testuser", "refresh", RefreshTokenDuration)
	if err != nil {
		t.Fatalf("generateTokenWithType() error: %v", err)
	}

	claims, err := mgr.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken() error: %v", err)
	}

	if claims.TokenType != "refresh" {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, "refresh")
	}
}

func TestGenerateTokenWithType_CustomDuration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Generate token with very short duration.
	token, err := mgr.generateTokenWithType("testuser", "access", 1*time.Second)
	if err != nil {
		t.Fatalf("generateTokenWithType() error: %v", err)
	}

	// Token should be valid immediately.
	claims, err := mgr.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken() error: %v", err)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt should not be nil")
	}

	// Expiration should be approximately 1 second from now.
	expectedExpiry := time.Now().Add(1 * time.Second)
	if claims.ExpiresAt.After(expectedExpiry.Add(100 * time.Millisecond)) {
		t.Error("Token expires too late")
	}
}

func TestGenerateToken_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	token, err := mgr.generateToken("testuser")
	if err != nil {
		t.Fatalf("generateToken() error: %v", err)
	}

	claims, err := mgr.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken() error: %v", err)
	}

	if claims.TokenType != "access" {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, "access")
	}
}

func TestCleanupLoop_StartedByNewTokenBlacklist(t *testing.T) {
	t.Parallel()

	// Create a blacklist - this starts the cleanup goroutine.
	bl := NewTokenBlacklist()

	// Add an expired token.
	bl.Add("expired", time.Now().Add(-time.Hour))

	// The cleanup loop runs every 5 minutes, so we can't easily test it.
	// But we can verify the blacklist was created successfully.
	if bl == nil {
		t.Error("NewTokenBlacklist() returned nil")
	}

	// Manually call cleanup to verify the cleanup function works.
	bl.cleanup()

	// Expired token should be removed.
	if bl.IsBlacklisted("expired") {
		t.Error("expired token should be removed after cleanup")
	}
}

func TestCleanup_DirectCall(t *testing.T) {
	t.Parallel()

	bl := NewTokenBlacklist()

	// Add tokens with various expiration times.
	bl.Add("expired1", time.Now().Add(-time.Hour))
	bl.Add("expired2", time.Now().Add(-time.Minute))
	bl.Add("valid1", time.Now().Add(time.Hour))
	bl.Add("valid2", time.Now().Add(time.Minute))

	// Call cleanup directly.
	bl.cleanup()

	// Expired tokens should be removed.
	if bl.IsBlacklisted("expired1") {
		t.Error("expired1 should be removed")
	}
	if bl.IsBlacklisted("expired2") {
		t.Error("expired2 should be removed")
	}

	// Valid tokens should remain.
	if !bl.IsBlacklisted("valid1") {
		t.Error("valid1 should still be present")
	}
	if !bl.IsBlacklisted("valid2") {
		t.Error("valid2 should still be present")
	}
}

func TestCleanup_EmptyBlacklist(t *testing.T) {
	t.Parallel()

	bl := NewTokenBlacklist()

	// Should not panic on empty blacklist.
	bl.cleanup()
}

func TestCleanup_AllExpired(t *testing.T) {
	t.Parallel()

	bl := NewTokenBlacklist()

	// Add only expired tokens.
	for i := range 10 {
		bl.tokens.Store("expired"+string(rune('0'+i)), time.Now().Add(-time.Hour))
	}

	// Call cleanup.
	bl.cleanup()

	// All tokens should be removed.
	for i := range 10 {
		tokenID := "expired" + string(rune('0'+i))
		if bl.IsBlacklisted(tokenID) {
			t.Errorf("%s should be removed", tokenID)
		}
	}
}

func TestCleanup_NoneExpired(t *testing.T) {
	t.Parallel()

	bl := NewTokenBlacklist()

	// Add only valid tokens.
	for i := range 10 {
		bl.tokens.Store("valid"+string(rune('0'+i)), time.Now().Add(time.Hour))
	}

	// Call cleanup.
	bl.cleanup()

	// All tokens should remain.
	for i := range 10 {
		tokenID := "valid" + string(rune('0'+i))
		if !bl.IsBlacklisted(tokenID) {
			t.Errorf("%s should still be present", tokenID)
		}
	}
}

func TestCleanup_MixedExpiry(t *testing.T) {
	t.Parallel()

	bl := NewTokenBlacklist()

	// Add mixed tokens directly to sync.Map.
	bl.tokens.Store("expired", time.Now().Add(-time.Hour))
	bl.tokens.Store("valid", time.Now().Add(time.Hour))

	// Call cleanup.
	bl.cleanup()

	// Verify results.
	if bl.IsBlacklisted("expired") {
		t.Error("expired should be removed")
	}
	if !bl.IsBlacklisted("valid") {
		t.Error("valid should still be present")
	}
}

func TestCleanup_EdgeCaseExactlyExpired(t *testing.T) {
	t.Parallel()

	bl := NewTokenBlacklist()

	// Add token that expired at exactly now (edge case).
	bl.tokens.Store("justExpired", time.Now())

	// Small delay to ensure time has passed.
	time.Sleep(time.Millisecond)

	// Call cleanup.
	bl.cleanup()

	// Token should be removed since it's now in the past.
	if bl.IsBlacklisted("justExpired") {
		t.Error("justExpired should be removed")
	}
}

func TestNewManager_WithZeroSessionTimeout(t *testing.T) {
	t.Parallel()

	mgr, err := NewManager("test-secret", 0, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	if mgr.SessionDuration() != AccessTokenDuration {
		t.Errorf("SessionDuration() = %v, want %v", mgr.SessionDuration(), AccessTokenDuration)
	}
}

func TestManagerFields(t *testing.T) {
	t.Parallel()

	mgr, err := NewManager("test-secret", 30*time.Minute, "testuser", "testpass")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Verify internal fields are set correctly.
	if string(mgr.jwtSecret) != "test-secret" {
		t.Error("jwtSecret not set correctly")
	}
	if mgr.sessionTimeout != 30*time.Minute {
		t.Errorf("sessionTimeout = %v, want %v", mgr.sessionTimeout, 30*time.Minute)
	}
	if mgr.username != "testuser" {
		t.Errorf("username = %q, want %q", mgr.username, "testuser")
	}
	if mgr.issuer != "The Stem" {
		t.Errorf("issuer = %q, want %q", mgr.issuer, "The Stem")
	}
	if mgr.blacklist == nil {
		t.Error("blacklist should not be nil")
	}
}

func TestCleanup_WithNonTimeValue(t *testing.T) {
	t.Parallel()

	bl := NewTokenBlacklist()

	// Add a token with the correct time.Time value.
	bl.tokens.Store("valid", time.Now().Add(time.Hour))

	// Try to add an invalid value type directly (edge case testing).
	// This simulates a potential bug where wrong type is stored.
	bl.tokens.Store("invalid", "not-a-time")

	// Cleanup should handle this gracefully (the type assertion will fail).
	bl.cleanup()

	// Valid token should still be present.
	if !bl.IsBlacklisted("valid") {
		t.Error("valid should still be present")
	}

	// The "invalid" entry should still be present since it's not a time.Time.
	_, exists := bl.tokens.Load("invalid")
	if !exists {
		t.Error("invalid entry should still exist (cleanup doesn't remove non-time values)")
	}
}

func TestTokenConstants(t *testing.T) {
	t.Parallel()

	// Verify constant values.
	if jwtSecretLength != 32 {
		t.Errorf("jwtSecretLength = %d, want 32", jwtSecretLength)
	}
	if tokenIDLength != 16 {
		t.Errorf("tokenIDLength = %d, want 16", tokenIDLength)
	}
}

func TestCleanupInterval(t *testing.T) {
	t.Parallel()

	// Verify cleanup interval is 5 minutes.
	if cleanupInterval != 5*time.Minute {
		t.Errorf("cleanupInterval = %v, want 5m", cleanupInterval)
	}
}

func TestBlacklistTokensMapOperations(t *testing.T) {
	t.Parallel()

	bl := NewTokenBlacklist()

	// Test direct sync.Map operations.
	bl.tokens.Store("test1", time.Now().Add(time.Hour))
	bl.tokens.Store("test2", time.Now().Add(-time.Hour))

	// Verify stored values.
	val, ok := bl.tokens.Load("test1")
	if !ok {
		t.Error("test1 should be stored")
	}
	if _, valid := val.(time.Time); !valid {
		t.Error("test1 value should be time.Time")
	}

	// Run cleanup.
	bl.cleanup()

	// test1 should remain, test2 should be deleted.
	_, exists1 := bl.tokens.Load("test1")
	if !exists1 {
		t.Error("test1 should still be present")
	}
	_, exists2 := bl.tokens.Load("test2")
	if exists2 {
		t.Error("test2 should be removed")
	}
}

func TestGenerateTokenWithType_AllFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mgr, err := NewManager("test-secret", 30*time.Minute, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	token, err := mgr.generateTokenWithType("testuser", "custom", 1*time.Hour)
	if err != nil {
		t.Fatalf("generateTokenWithType() error: %v", err)
	}

	claims, err := mgr.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken() error: %v", err)
	}

	// Verify all claims.
	if claims.Username != "testuser" {
		t.Errorf("Username = %q, want %q", claims.Username, "testuser")
	}
	if claims.TokenType != "custom" {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, "custom")
	}
	if claims.Issuer != "The Stem" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "The Stem")
	}
	if claims.Subject != "testuser" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "testuser")
	}
	if claims.ID == "" {
		t.Error("ID should not be empty")
	}
	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
	}
	if claims.IssuedAt == nil {
		t.Error("IssuedAt should not be nil")
	}
	if claims.NotBefore == nil {
		t.Error("NotBefore should not be nil")
	}
}

func TestManagerBlacklistIntegration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Verify blacklist is initialized.
	if mgr.blacklist == nil {
		t.Error("blacklist should not be nil")
	}

	// Create and revoke a token.
	token, err := mgr.Authenticate(ctx, "admin", "admin")
	if err != nil {
		t.Fatalf("Authenticate() error: %v", err)
	}

	claims, err := mgr.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("ValidateToken() error: %v", err)
	}

	mgr.RevokeToken(claims)

	// Verify token is in blacklist.
	if !mgr.blacklist.IsBlacklisted(claims.ID) {
		t.Error("Token ID should be in blacklist")
	}
}

func TestGenerateTokenID_Base64URLSafe(t *testing.T) {
	t.Parallel()

	// Create a manager to access generateTokenID method.
	mgr, mgrErr := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if mgrErr != nil {
		t.Fatalf("NewManager() error: %v", mgrErr)
	}

	for range 100 {
		id, err := mgr.generateTokenID()
		if err != nil {
			t.Fatalf("generateTokenID() error: %v", err)
		}

		// Verify the ID only contains base64url-safe characters.
		for _, c := range id {
			isUpperAlpha := c >= 'A' && c <= 'Z'
			isLowerAlpha := c >= 'a' && c <= 'z'
			isDigit := c >= '0' && c <= '9'
			isSpecial := c == '-' || c == '_'
			if !isUpperAlpha && !isLowerAlpha && !isDigit && !isSpecial {
				t.Errorf("Invalid character in token ID: %c", c)
			}
		}
	}
}

func TestCleanup_RangeCallback(t *testing.T) {
	t.Parallel()

	bl := NewTokenBlacklist()

	// Add tokens with specific expiry patterns.
	now := time.Now()

	// Expired tokens.
	for i := range 5 {
		bl.tokens.Store("expired"+string(rune('A'+i)), now.Add(-time.Duration(i+1)*time.Minute))
	}

	// Valid tokens.
	for i := range 5 {
		bl.tokens.Store("valid"+string(rune('A'+i)), now.Add(time.Duration(i+1)*time.Minute))
	}

	// Run cleanup.
	bl.cleanup()

	// Count remaining tokens.
	count := 0
	bl.tokens.Range(func(_, _ any) bool {
		count++
		return true
	})

	// Should have 5 valid tokens remaining.
	if count != 5 {
		t.Errorf("Expected 5 tokens remaining, got %d", count)
	}
}

// -----------------------------------------------------------------------------
// Error Injection Tests (using SetRandReader and GenerateJWTSecretFrom)
// -----------------------------------------------------------------------------

// errorReader is a mock [io.Reader] that always returns an error.
type errorReader struct {
	err error
}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, e.err
}

func TestGenerateTokenID_RandReadError(t *testing.T) {
	// Create a valid manager.
	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Inject error reader via SetRandReader.
	mgr.SetRandReader(&errorReader{err: errors.New("mock random error")})

	_, err = mgr.generateTokenID()
	if err == nil {
		t.Error("generateTokenID() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "read random bytes") {
		t.Errorf("Error should contain 'read random bytes', got: %v", err)
	}
}

func TestGenerateJWTSecretFrom_RandReadError(t *testing.T) {
	// Test GenerateJWTSecretFrom directly with error reader.
	_, err := GenerateJWTSecretFrom(&errorReader{err: errors.New("mock random error")})
	if err == nil {
		t.Error("GenerateJWTSecretFrom() expected error, got nil")
	}
	if !errors.Is(err, ErrSecretGenerationFailed) {
		t.Errorf("Error should be ErrSecretGenerationFailed, got: %v", err)
	}
}

func TestGenerateJWTSecretFrom_Success(t *testing.T) {
	// Test GenerateJWTSecretFrom with crypto/rand.Reader works.
	secret, err := GenerateJWTSecretFrom(rand.Reader)
	if err != nil {
		t.Errorf("GenerateJWTSecretFrom() unexpected error: %v", err)
	}
	if secret == "" {
		t.Error("GenerateJWTSecretFrom() returned empty secret")
	}
	// Check base64url encoding (no + or / chars).
	if strings.ContainsAny(secret, "+/") {
		t.Error("GenerateJWTSecretFrom() should use base64url encoding")
	}
}

func TestGenerateTokenWithType_TokenIDError(t *testing.T) {
	// First create a valid manager.
	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Inject error reader via SetRandReader.
	mgr.SetRandReader(&errorReader{err: errors.New("mock random error")})

	_, err = mgr.generateTokenWithType("testuser", "access", AccessTokenDuration)
	if err == nil {
		t.Error("generateTokenWithType() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "generate token ID") {
		t.Errorf("Error should contain 'generate token ID', got: %v", err)
	}
}

func TestAuthenticate_TokenGenerationError(t *testing.T) {
	ctx := context.Background()

	// First create a valid manager.
	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Inject error reader via SetRandReader.
	mgr.SetRandReader(&errorReader{err: errors.New("mock random error")})

	_, err = mgr.Authenticate(ctx, "admin", "admin")
	if err == nil {
		t.Error("Authenticate() expected error, got nil")
	}
}

func TestAuthenticateWithRefresh_TokenGenerationError(t *testing.T) {
	ctx := context.Background()

	// First create a valid manager.
	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Inject error reader via SetRandReader.
	mgr.SetRandReader(&errorReader{err: errors.New("mock random error")})

	_, _, err = mgr.AuthenticateWithRefresh(ctx, "admin", "admin")
	if err == nil {
		t.Error("AuthenticateWithRefresh() expected error, got nil")
	}
}

func TestGenerateRefreshToken_TokenIDError(t *testing.T) {
	// First create a valid manager.
	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Inject error reader via SetRandReader.
	mgr.SetRandReader(&errorReader{err: errors.New("mock random error")})

	_, err = mgr.GenerateRefreshToken("admin")
	if err == nil {
		t.Error("GenerateRefreshToken() expected error, got nil")
	}
}

// readerFunc adapts a function to the [io.Reader] interface.
type readerFunc func(p []byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) {
	return f(p)
}

func TestAuthenticateWithRefresh_RefreshTokenGenerationError(t *testing.T) {
	ctx := context.Background()

	// First create a valid manager.
	mgr, err := NewManager("test-secret", AccessTokenDuration, "admin", "admin")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Create a reader that succeeds once (for access token) then fails.
	callCount := 0
	mgr.SetRandReader(readerFunc(func(p []byte) (int, error) {
		callCount++
		if callCount == 1 {
			// First call - generate access token ID successfully.
			return rand.Reader.Read(p)
		}
		// Second call - fail for refresh token.
		return 0, errors.New("mock error on second call")
	}))

	_, _, err = mgr.AuthenticateWithRefresh(ctx, "admin", "admin")
	if err == nil {
		t.Error("AuthenticateWithRefresh() expected error on refresh token generation, got nil")
	}
}
