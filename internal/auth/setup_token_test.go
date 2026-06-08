// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"sync"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

func TestNewSetupTokenManager(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()
	if manager == nil {
		t.Fatal("NewSetupTokenManager returned nil")
	}

	if manager.HasValidToken() {
		t.Error("new manager should not have a valid token")
	}
}

func TestSetupTokenManager_GenerateToken(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()

	token, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("GenerateToken returned empty token")
	}

	// Token should be base64-encoded 32 bytes = 43-44 chars.
	if len(token) < 40 {
		t.Errorf("GenerateToken token too short: %d chars", len(token))
	}

	if !manager.HasValidToken() {
		t.Error("manager should have valid token after generation")
	}
}

func TestSetupTokenManager_GenerateToken_Unique(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()
	tokens := make(map[string]bool)

	for range 100 {
		token, err := manager.GenerateToken()
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}
		if tokens[token] {
			t.Error("GenerateToken produced duplicate token")
		}
		tokens[token] = true
	}
}

func TestSetupTokenManager_ValidateToken(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()

	token, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Valid token should succeed.
	if !manager.ValidateToken(token) {
		t.Error("ValidateToken should succeed for valid token")
	}

	// Token should be single-use.
	if manager.ValidateToken(token) {
		t.Error("ValidateToken should fail for already-used token")
	}
}

func TestSetupTokenManager_ValidateToken_Empty(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()

	if manager.ValidateToken("") {
		t.Error("ValidateToken should fail for empty token")
	}
}

func TestSetupTokenManager_ValidateToken_NoToken(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()

	if manager.ValidateToken("sometoken") {
		t.Error("ValidateToken should fail when no token exists")
	}
}

func TestSetupTokenManager_ValidateToken_WrongToken(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()

	_, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if manager.ValidateToken("wrongtoken") {
		t.Error("ValidateToken should fail for wrong token")
	}
}

func TestSetupTokenManager_ValidateToken_NewTokenInvalidatesOld(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()

	token1, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	token2, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Old token should be invalid.
	if manager.ValidateToken(token1) {
		t.Error("old token should be invalid after new token generated")
	}

	// New token should be valid.
	if !manager.ValidateToken(token2) {
		t.Error("new token should be valid")
	}
}

func TestSetupTokenManager_Invalidate(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()

	token, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if !manager.HasValidToken() {
		t.Error("should have valid token before invalidate")
	}

	manager.Invalidate()

	if manager.HasValidToken() {
		t.Error("should not have valid token after invalidate")
	}

	if manager.ValidateToken(token) {
		t.Error("token should be invalid after Invalidate")
	}
}

func TestSetupTokenManager_HasValidToken_UsedToken(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()

	token, err := manager.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if !manager.HasValidToken() {
		t.Error("should have valid token")
	}

	manager.ValidateToken(token)

	if manager.HasValidToken() {
		t.Error("should not have valid token after use")
	}
}

func TestSetupTokenManager_Concurrent(t *testing.T) {
	t.Parallel()

	manager := auth.NewSetupTokenManager()
	var wg sync.WaitGroup

	// Generate tokens concurrently.
	for range 10 {
		wg.Go(func() {
			_, _ = manager.GenerateToken()
		})
	}

	// Validate tokens concurrently.
	for range 10 {
		wg.Go(func() {
			_ = manager.ValidateToken("sometoken")
		})
	}

	// Check HasValidToken concurrently.
	for range 10 {
		wg.Go(func() {
			_ = manager.HasValidToken()
		})
	}

	wg.Wait()
}

func TestSetupTokenExpiry(t *testing.T) {
	// Verify the expiry constant is reasonable.
	if auth.SetupTokenExpiry != 15*time.Minute {
		t.Errorf("SetupTokenExpiry = %v, want 15 minutes", auth.SetupTokenExpiry)
	}
}
