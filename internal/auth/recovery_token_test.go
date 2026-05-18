// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
)

func TestNewRecoveryTokenManager(t *testing.T) {
	t.Parallel()

	manager := auth.NewRecoveryTokenManager("/tmp/test-recovery")
	if manager == nil {
		t.Fatal("NewRecoveryTokenManager returned nil")
	}

	if manager.IsActive() {
		t.Error("new manager should not be active")
	}
}

func TestRecoveryTokenManager_TriggerFilePath(t *testing.T) {
	t.Parallel()

	manager := auth.NewRecoveryTokenManager("/tmp/test-data")
	expected := filepath.Join("/tmp/test-data", ".recovery")

	if got := manager.TriggerFilePath(); got != expected {
		t.Errorf("TriggerFilePath() = %q, want %q", got, expected)
	}
}

func TestRecoveryTokenManager_TokenFilePath(t *testing.T) {
	t.Parallel()

	manager := auth.NewRecoveryTokenManager("/tmp/test-data")
	expected := filepath.Join("/tmp/test-data", ".recovery-token")

	if got := manager.TokenFilePath(); got != expected {
		t.Errorf("TokenFilePath() = %q, want %q", got, expected)
	}
}

func TestRecoveryTokenManager_TokenExpiryDuration(t *testing.T) {
	t.Parallel()

	manager := auth.NewRecoveryTokenManager("/tmp/test-data")
	if got := manager.TokenExpiryDuration(); got != auth.RecoveryTokenExpiry {
		t.Errorf("TokenExpiryDuration() = %v, want %v", got, auth.RecoveryTokenExpiry)
	}
}

func TestRecoveryTokenManager_CheckRecoveryMode_NoTrigger(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	if manager.CheckRecoveryMode() {
		t.Error("CheckRecoveryMode should return false when no trigger file exists")
	}
}

func TestRecoveryTokenManager_CheckRecoveryMode_WithTrigger(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	// Create trigger file.
	triggerPath := manager.TriggerFilePath()
	err := os.WriteFile(triggerPath, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("failed to create trigger file: %v", err)
	}
	defer os.Remove(triggerPath)

	if !manager.CheckRecoveryMode() {
		t.Error("CheckRecoveryMode should return true when trigger file exists")
	}

	// Token file should be created.
	tokenPath := manager.TokenFilePath()
	_, err = os.Stat(tokenPath)
	if os.IsNotExist(err) {
		t.Error("token file should be created after CheckRecoveryMode")
	}

	// Read and verify token content.
	content, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("failed to read token file: %v", err)
	}

	// Token should not be empty (includes newline).
	if len(content) < 40 {
		t.Errorf("token content too short: %d bytes", len(content))
	}
}

func TestRecoveryTokenManager_ValidateAndConsume(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	// Create trigger file to activate recovery mode.
	triggerPath := manager.TriggerFilePath()
	err := os.WriteFile(triggerPath, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("failed to create trigger file: %v", err)
	}

	manager.CheckRecoveryMode()

	// Read the token from file.
	tokenPath := manager.TokenFilePath()
	content, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("failed to read token file: %v", err)
	}
	token := string(content[:len(content)-1]) // Remove newline.

	// Valid token should succeed.
	if !manager.ValidateAndConsume(token) {
		t.Error("ValidateAndConsume should succeed for valid token")
	}

	// Token should be single-use.
	if manager.ValidateAndConsume(token) {
		t.Error("ValidateAndConsume should fail for already-used token")
	}
}

func TestRecoveryTokenManager_ValidateAndConsume_Empty(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	if manager.ValidateAndConsume("") {
		t.Error("ValidateAndConsume should fail for empty token")
	}
}

func TestRecoveryTokenManager_ValidateAndConsume_NoToken(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	if manager.ValidateAndConsume("sometoken") {
		t.Error("ValidateAndConsume should fail when no token exists")
	}
}

func TestRecoveryTokenManager_ValidateAndConsume_WrongToken(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	// Create trigger file to activate recovery mode.
	triggerPath := manager.TriggerFilePath()
	err := os.WriteFile(triggerPath, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("failed to create trigger file: %v", err)
	}

	manager.CheckRecoveryMode()

	if manager.ValidateAndConsume("wrongtoken") {
		t.Error("ValidateAndConsume should fail for wrong token")
	}
}

func TestRecoveryTokenManager_Cleanup(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	// Create trigger file to activate recovery mode.
	triggerPath := manager.TriggerFilePath()
	err := os.WriteFile(triggerPath, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("failed to create trigger file: %v", err)
	}

	manager.CheckRecoveryMode()

	// Verify files exist.
	tokenPath := manager.TokenFilePath()
	_, err = os.Stat(triggerPath)
	if os.IsNotExist(err) {
		t.Error("trigger file should exist before cleanup")
	}
	_, err = os.Stat(tokenPath)
	if os.IsNotExist(err) {
		t.Error("token file should exist before cleanup")
	}

	// Cleanup.
	manager.Cleanup()

	// Verify files are removed.
	_, err = os.Stat(triggerPath)
	if !os.IsNotExist(err) {
		t.Error("trigger file should be removed after cleanup")
	}
	_, err = os.Stat(tokenPath)
	if !os.IsNotExist(err) {
		t.Error("token file should be removed after cleanup")
	}

	// Manager should no longer be active.
	if manager.IsActive() {
		t.Error("manager should not be active after cleanup")
	}
}

func TestRecoveryTokenManager_IsActive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	if manager.IsActive() {
		t.Error("should not be active without trigger file")
	}

	// Create trigger file.
	triggerPath := manager.TriggerFilePath()
	err := os.WriteFile(triggerPath, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("failed to create trigger file: %v", err)
	}

	// Still not active until CheckRecoveryMode generates token.
	if manager.IsActive() {
		t.Error("should not be active until token is generated")
	}

	manager.CheckRecoveryMode()

	if !manager.IsActive() {
		t.Error("should be active after CheckRecoveryMode")
	}
}

func TestRecoveryTokenManager_Invalidate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	// Create trigger file and generate token.
	triggerPath := manager.TriggerFilePath()
	err := os.WriteFile(triggerPath, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("failed to create trigger file: %v", err)
	}

	manager.CheckRecoveryMode()

	if !manager.IsActive() {
		t.Error("should be active before invalidate")
	}

	manager.Invalidate()

	// Note: IsActive also checks trigger file, so it may still return true.
	// The internal token is cleared though.
	if manager.RemainingTime() != 0 {
		t.Error("RemainingTime should be 0 after invalidate")
	}
}

func TestRecoveryTokenManager_RemainingTime(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	// No token - should be 0.
	if manager.RemainingTime() != 0 {
		t.Error("RemainingTime should be 0 when no token exists")
	}

	// Create trigger file and generate token.
	triggerPath := manager.TriggerFilePath()
	err := os.WriteFile(triggerPath, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("failed to create trigger file: %v", err)
	}

	manager.CheckRecoveryMode()

	remaining := manager.RemainingTime()
	if remaining <= 0 {
		t.Error("RemainingTime should be positive after token generation")
	}
	if remaining > auth.RecoveryTokenExpiry {
		t.Errorf("RemainingTime %v should not exceed RecoveryTokenExpiry %v", remaining, auth.RecoveryTokenExpiry)
	}
}

func TestRecoveryTokenManager_Concurrent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manager := auth.NewRecoveryTokenManager(tmpDir)

	// Create trigger file.
	triggerPath := manager.TriggerFilePath()
	err := os.WriteFile(triggerPath, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("failed to create trigger file: %v", err)
	}

	var wg sync.WaitGroup

	// Check recovery mode concurrently.
	for range 10 {
		wg.Go(func() {
			_ = manager.CheckRecoveryMode()
		})
	}

	// Validate tokens concurrently.
	for range 10 {
		wg.Go(func() {
			_ = manager.ValidateAndConsume("sometoken")
		})
	}

	// Check IsActive concurrently.
	for range 10 {
		wg.Go(func() {
			_ = manager.IsActive()
		})
	}

	// Check RemainingTime concurrently.
	for range 10 {
		wg.Go(func() {
			_ = manager.RemainingTime()
		})
	}

	wg.Wait()
}

func TestRecoveryTokenExpiry(t *testing.T) {
	// Verify the expiry constant is reasonable.
	if auth.RecoveryTokenExpiry != 15*time.Minute {
		t.Errorf("RecoveryTokenExpiry = %v, want 15 minutes", auth.RecoveryTokenExpiry)
	}
}
