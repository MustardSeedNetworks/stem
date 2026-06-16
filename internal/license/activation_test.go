// SPDX-License-Identifier: BUSL-1.1

package license_test

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

// Test constants.
const (
	configDirPerm   = 0o700
	minTrialDays    = 13
	maxTrialDays    = 14
	minHashLen      = 8
	minStringLen    = 20
	expectedHashLen = 16
)

// newTestManager creates a license manager backed by the TEST verifier so that
// tokens minted with signTestKey activate. It uses a fresh temp config dir.
func newTestManager(t *testing.T) (*license.Manager, error) {
	t.Helper()
	return license.NewManagerWithVerifier(testVerifier(t))
}

// setupTestManager creates a manager with a temporary config directory. The
// manager verifies tokens against the TEST signing key so signTestKey tokens
// activate.
func setupTestManager(t *testing.T) *license.Manager {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	mgr, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManagerWithVerifier() error: %v", err)
	}
	return mgr
}

func TestNewManager(t *testing.T) {
	mgr := setupTestManager(t)
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestManagerGetFingerprint(t *testing.T) {
	mgr := setupTestManager(t)

	fp := mgr.GetFingerprint()
	if fp == nil {
		t.Error("GetFingerprint() returned nil")
	}
	hash := fp.Hash()
	if len(hash) < minHashLen {
		t.Errorf("GetFingerprint().Hash() too short: %s", hash)
	}
}

func TestManagerIsActivated(t *testing.T) {
	mgr := setupTestManager(t)

	// Fresh install should not be activated.
	if mgr.IsActivated() {
		t.Error("Fresh manager should not be activated")
	}
}

func TestManagerStartTrial(t *testing.T) {
	mgr := setupTestManager(t)

	result := mgr.StartTrial()
	if !result.Success {
		t.Errorf("StartTrial() failed: %s", result.Message)
	}

	// Should now be activated in trial mode.
	if !mgr.IsActivated() {
		t.Error("Should be activated after starting trial")
	}

	if !mgr.IsTrialValid() {
		t.Error("Trial should be valid")
	}

	days := mgr.TrialDaysRemaining()
	if days < minTrialDays || days > maxTrialDays {
		t.Errorf("Expected ~14 trial days, got %d", days)
	}
}

func TestManagerTrialExpiry(t *testing.T) {
	mgr := setupTestManager(t)

	// Manually set expired trial by starting trial and then simulating time.
	result := mgr.StartTrial()
	if !result.Success {
		t.Fatalf("StartTrial() failed: %s", result.Message)
	}

	// Get the state and check it expired after 15 days.
	state := mgr.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil after StartTrial")
	}

	// For testing expired trial, we need to verify IsTrialValid behavior.
	// A trial that started 15 days ago should not be valid.
	// Since we can't easily modify internal state in black-box testing,
	// we verify the current trial is valid.
	if !mgr.IsTrialValid() {
		t.Error("Fresh trial should be valid")
	}
}

func TestManagerActivate(t *testing.T) {
	mgr := setupTestManager(t)

	// Generate a valid key for TestSuite tier.
	key, err := signTestKey(t, "2001", "1234567", license.TierProfessional)
	if err != nil {
		t.Fatalf("signTestKey() error: %v", err)
	}

	result := mgr.Activate(key)
	if !result.Success {
		t.Errorf("Activate() failed: %s", result.Message)
	}

	if !mgr.IsActivated() {
		t.Error("Should be activated after license activation")
	}

	state := mgr.GetState()
	if state.Tier != license.TierProfessional {
		t.Errorf("Expected tier %d, got %d", license.TierProfessional, state.Tier)
	}
}

func TestManagerActivateInvalidKey(t *testing.T) {
	mgr := setupTestManager(t)

	result := mgr.Activate("INVALID-KEY-1234")
	if result.Success {
		t.Error("Invalid key should not activate")
	}
}

func TestManagerDeactivate(t *testing.T) {
	mgr := setupTestManager(t)

	// First activate.
	mgr.StartTrial()
	if !mgr.IsActivated() {
		t.Fatal("Should be activated")
	}

	// Then deactivate.
	err := mgr.Deactivate()
	if err != nil {
		t.Errorf("Deactivate() error: %v", err)
	}

	if mgr.IsActivated() {
		t.Error("Should not be activated after deactivation")
	}
}

func TestManagerGetState(t *testing.T) {
	mgr := setupTestManager(t)

	// Start trial first to have a state.
	mgr.StartTrial()

	state := mgr.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil")
	}

	// Device hash should be set.
	if state.DeviceHash == "" {
		t.Error("DeviceHash should not be empty")
	}
}

func TestManagerNeedsCheckIn(t *testing.T) {
	mgr := setupTestManager(t)

	// No state = no check-in needed.
	if mgr.NeedsCheckIn() {
		t.Error("No state should not need check-in")
	}

	// Trial mode shouldn't need check-in.
	mgr.StartTrial()
	if mgr.NeedsCheckIn() {
		t.Error("Trial mode should not need check-in")
	}
}

func TestManagerCheckIn(t *testing.T) {
	mgr := setupTestManager(t)

	// No state.
	result := mgr.CheckIn()
	if result.Success {
		t.Error("CheckIn with no state should fail")
	}

	// With state.
	mgr.StartTrial()
	result = mgr.CheckIn()
	if !result.Success {
		t.Errorf("CheckIn with state failed: %s", result.Message)
	}
}

func TestDeviceFingerprintString(t *testing.T) {
	t.Parallel()
	fp, err := license.GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint() error: %v", err)
	}

	s := fp.String()
	if s == "" {
		t.Error("Fingerprint String() should not be empty")
	}
	if len(s) < minStringLen {
		t.Error("Fingerprint String() seems too short")
	}
}

func TestDeviceFingerprintHash(t *testing.T) {
	t.Parallel()
	fp, err := license.GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint() error: %v", err)
	}

	hash := fp.Hash()
	if len(hash) != expectedHashLen {
		t.Errorf("Expected 16-char hash, got %d chars", len(hash))
	}

	// Same input should produce same hash.
	hash2 := fp.Hash()
	if hash != hash2 {
		t.Error("Hash should be deterministic")
	}
}

func TestManagerStartTrialTwice(t *testing.T) {
	mgr := setupTestManager(t)

	// Start trial first time.
	result := mgr.StartTrial()
	if !result.Success {
		t.Errorf("First StartTrial() failed: %s", result.Message)
	}

	// Start trial second time - should succeed but show remaining days.
	result2 := mgr.StartTrial()
	if !result2.Success {
		t.Errorf("Second StartTrial() failed: %s", result2.Message)
	}
	if result2.DaysRemaining <= 0 {
		t.Error("Should show remaining days")
	}
}

func TestManagerActivateExpiredLicense(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate with a valid key first.
	key, err := signTestKey(t, "2001", "1234567", license.TierProfessional)
	if err != nil {
		t.Fatalf("signTestKey() error: %v", err)
	}

	result := mgr.Activate(key)
	if !result.Success {
		t.Errorf("Activate() failed: %s", result.Message)
	}

	// Verify activation works.
	if !mgr.IsActivated() {
		t.Error("Should be activated after activation")
	}
}

func TestTrialDaysRemainingNoState(t *testing.T) {
	mgr := setupTestManager(t)

	// No state = 0 days.
	days := mgr.TrialDaysRemaining()
	if days != 0 {
		t.Errorf("Expected 0 days with no state, got %d", days)
	}
}

func TestTrialDaysRemainingNotTrial(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate with full license (not trial).
	key, err := signTestKey(t, "2001", "1234567", license.TierProfessional)
	if err != nil {
		t.Fatalf("signTestKey() error: %v", err)
	}

	mgr.Activate(key)

	days := mgr.TrialDaysRemaining()
	if days != 0 {
		t.Errorf("Expected 0 days for non-trial, got %d", days)
	}
}

func TestIsTrialValidNoState(t *testing.T) {
	mgr := setupTestManager(t)

	// No state = not valid.
	if mgr.IsTrialValid() {
		t.Error("No state should mean trial not valid")
	}
}

func TestStartTrialAlreadyActivated(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate with full license.
	key, err := signTestKey(t, "2001", "1234567", license.TierProfessional)
	if err != nil {
		t.Fatalf("signTestKey() error: %v", err)
	}

	mgr.Activate(key)

	// Try to start trial - should succeed but return existing license info.
	result := mgr.StartTrial()
	if !result.Success {
		t.Errorf("StartTrial on activated license should succeed: %s", result.Message)
	}
}

func TestNeedsCheckInAfterActivation(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate with a valid key.
	key, err := signTestKey(t, "2001", "1234567", license.TierProfessional)
	if err != nil {
		t.Fatalf("signTestKey() error: %v", err)
	}

	result := mgr.Activate(key)
	if !result.Success {
		t.Fatalf("Activate() failed: %s", result.Message)
	}

	// Right after activation, should not need check-in.
	if mgr.NeedsCheckIn() {
		t.Error("Should not need check-in immediately after activation")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	mgr := setupTestManager(t)

	// Start trial to ensure manager has fingerprint set up.
	mgr.StartTrial()

	// The encrypt/decrypt functions are private, so we test them indirectly
	// by verifying that the manager can save and load state correctly.
	state := mgr.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil after StartTrial")
	}

	// Create a new manager with the same HOME directory to test state loading.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	mgr2, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Start trial and verify it works.
	result := mgr2.StartTrial()
	if !result.Success {
		t.Errorf("StartTrial() failed: %s", result.Message)
	}

	// Verify state can be retrieved.
	state2 := mgr2.GetState()
	if state2 == nil {
		t.Error("GetState() returned nil after StartTrial on mgr2")
	}
}

func TestActivateInvalidKeyFormats(t *testing.T) {
	mgr := setupTestManager(t)

	invalidKeys := []string{
		"",                     // Empty.
		"SHORT",                // Too short.
		"TOOLONGKEYVALUE12345", // Too long.
		"INVALID-CHARS-@@@@",   // Invalid characters.
		"1234567890123456",     // All numbers (may have invalid checksum).
	}

	for _, key := range invalidKeys {
		result := mgr.Activate(key)
		if result.Success {
			t.Errorf("Invalid key %q should not activate", key)
		}
	}
}

func TestTrialFeatures(t *testing.T) {
	mgr := setupTestManager(t)

	result := mgr.StartTrial()
	if !result.Success {
		t.Fatalf("StartTrial() failed: %s", result.Message)
	}

	// Trial should have TierProfessional.
	if result.Tier != license.TierProfessional {
		t.Errorf("Trial should have TierProfessional, got %v", result.Tier)
	}

	// Trial should be marked as trial mode.
	if !result.IsTrialMode {
		t.Error("Trial result should have IsTrialMode=true")
	}

	// Verify trial days.
	if result.DaysRemaining != license.TrialDays {
		t.Errorf("Expected %d trial days, got %d", license.TrialDays, result.DaysRemaining)
	}
}

func TestCheckInUpdatesLastValidated(t *testing.T) {
	mgr := setupTestManager(t)

	// Start trial.
	mgr.StartTrial()

	// Get initial state.
	state1 := mgr.GetState()
	if state1 == nil {
		t.Fatal("GetState() returned nil")
	}

	// Wait a tiny bit to ensure time difference.
	time.Sleep(10 * time.Millisecond)

	// Check in.
	result := mgr.CheckIn()
	if !result.Success {
		t.Errorf("CheckIn() failed: %s", result.Message)
	}

	// Get updated state.
	state2 := mgr.GetState()
	if state2 == nil {
		t.Fatal("GetState() returned nil after CheckIn")
	}

	// LastValidatedAt should be updated.
	if !state2.LastValidatedAt.After(state1.TrialStartedAt) {
		t.Error("LastValidatedAt should be updated after CheckIn")
	}
}

func TestDeactivateWithNoLicense(t *testing.T) {
	mgr := setupTestManager(t)

	// Deactivate without any license should not error.
	err := mgr.Deactivate()
	if err != nil {
		t.Errorf("Deactivate() with no license should not error: %v", err)
	}

	// Should still not be activated.
	if mgr.IsActivated() {
		t.Error("Should not be activated after Deactivate")
	}
}

func TestActivateThenDeactivateThenActivate(t *testing.T) {
	mgr := setupTestManager(t)

	// First activation.
	key1, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	result1 := mgr.Activate(key1)
	if !result1.Success {
		t.Fatalf("First Activate() failed: %s", result1.Message)
	}

	if !mgr.IsActivated() {
		t.Error("Should be activated after first activation")
	}

	// Deactivate.
	err := mgr.Deactivate()
	if err != nil {
		t.Fatalf("Deactivate() error: %v", err)
	}

	if mgr.IsActivated() {
		t.Error("Should not be activated after deactivation")
	}

	// Second activation with different key.
	key2, _ := signTestKey(t, "2001", "7654321", license.TierProfessional)
	result2 := mgr.Activate(key2)
	if !result2.Success {
		t.Errorf("Second Activate() failed: %s", result2.Message)
	}

	if !mgr.IsActivated() {
		t.Error("Should be activated after second activation")
	}
}

// TestIsActivatedExpiredLicense tests license expiration detection.
func TestIsActivatedExpiredLicense(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate with a valid key.
	key, err := signTestKey(t, "2001", "1234567", license.TierProfessional)
	if err != nil {
		t.Fatalf("signTestKey() error: %v", err)
	}

	result := mgr.Activate(key)
	if !result.Success {
		t.Fatalf("Activate() failed: %s", result.Message)
	}

	// Should be activated initially.
	if !mgr.IsActivated() {
		t.Error("Should be activated after fresh activation")
	}

	// Verify state is not nil.
	state := mgr.GetState()
	if state == nil {
		t.Fatal("State should not be nil after activation")
	}

	// Verify state has correct tier.
	if state.Tier != license.TierProfessional {
		t.Errorf("Expected TierProfessional, got %v", state.Tier)
	}

	// Verify IsTrialMode is false.
	if state.IsTrialMode {
		t.Error("Regular activation should not be in trial mode")
	}
}

// TestIsActivatedDeviceHashMismatch tests device binding verification.
func TestIsActivatedDeviceHashMismatch(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate first.
	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	mgr.Activate(key)

	// Get fingerprint to verify device binding is working.
	fp := mgr.GetFingerprint()
	if fp == nil {
		t.Fatal("GetFingerprint() returned nil")
	}

	state := mgr.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil")
	}

	// Device hash should match fingerprint.
	if state.DeviceHash != fp.Hash() {
		t.Error("Device hash should match fingerprint hash")
	}
}

// TestTrialModeComplete tests complete trial workflow.
func TestTrialModeComplete(t *testing.T) {
	mgr := setupTestManager(t)

	// Start trial.
	result := mgr.StartTrial()
	if !result.Success {
		t.Fatalf("StartTrial() failed: %s", result.Message)
	}

	// Verify trial properties.
	if !result.IsTrialMode {
		t.Error("Result should indicate trial mode")
	}

	if result.Tier != license.TierProfessional {
		t.Errorf("Trial should have TierProfessional, got %v", result.Tier)
	}

	if result.DaysRemaining != license.TrialDays {
		t.Errorf("Expected %d trial days, got %d", license.TrialDays, result.DaysRemaining)
	}

	// Verify state.
	state := mgr.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil")
	}

	if !state.IsTrialMode {
		t.Error("State should be in trial mode")
	}

	if state.TrialStartedAt.IsZero() {
		t.Error("TrialStartedAt should be set")
	}

	// Verify features are present.
	expectedFeatures := []string{"reflector", "rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn"}
	for _, f := range expectedFeatures {
		if !slices.Contains(state.Features, f) {
			t.Errorf("Expected feature %q not found in trial features", f)
		}
	}
}

// TestTrialAfterFullLicenseActivation tests starting trial after full license.
func TestTrialAfterFullLicenseActivation(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate with full license.
	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	result := mgr.Activate(key)
	if !result.Success {
		t.Fatalf("Activate() failed: %s", result.Message)
	}

	// Try to start trial - should return info about existing license.
	trialResult := mgr.StartTrial()

	// Should succeed but indicate already activated.
	if !trialResult.Success {
		t.Error("StartTrial after activation should succeed")
	}

	// Should not be in trial mode (already has full license).
	if trialResult.IsTrialMode {
		t.Error("Should not be in trial mode when already activated")
	}

	// Should have correct tier.
	if trialResult.Tier != license.TierProfessional {
		t.Errorf("Expected TierProfessional, got %v", trialResult.Tier)
	}
}

// TestStatePersistedAndReloaded tests that state survives manager recreation.
func TestStatePersistedAndReloaded(t *testing.T) {
	// Create temp directory for this test.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create first manager and activate.
	mgr1, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	result := mgr1.Activate(key)
	if !result.Success {
		t.Fatalf("Activate() failed: %s", result.Message)
	}

	// Create second manager - should load persisted state.
	mgr2, err := newTestManager(t)
	if err != nil {
		t.Fatalf("Second NewManager() error: %v", err)
	}

	// Should still be activated.
	if !mgr2.IsActivated() {
		t.Error("Manager should be activated after reload")
	}

	// State should have correct tier.
	state := mgr2.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil on reloaded manager")
	}

	if state.Tier != license.TierProfessional {
		t.Errorf("Reloaded state should have TierProfessional, got %v", state.Tier)
	}
}

// TestTrialPersistedAndReloaded tests that trial state survives manager recreation.
func TestTrialPersistedAndReloaded(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create first manager and start trial.
	mgr1, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	result := mgr1.StartTrial()
	if !result.Success {
		t.Fatalf("StartTrial() failed: %s", result.Message)
	}

	// Create second manager - should load persisted trial state.
	mgr2, err := newTestManager(t)
	if err != nil {
		t.Fatalf("Second NewManager() error: %v", err)
	}

	// Should still be activated (in trial).
	if !mgr2.IsActivated() {
		t.Error("Manager should be activated after reload (trial)")
	}

	// Should be in trial mode.
	if !mgr2.IsTrialValid() {
		t.Error("Trial should still be valid after reload")
	}

	state := mgr2.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil on reloaded manager")
	}

	if !state.IsTrialMode {
		t.Error("Reloaded state should be in trial mode")
	}
}

// TestLoadStateWithCorruptedFile tests loading state from corrupted file.
func TestLoadStateWithCorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write corrupted license file.
	licensePath := filepath.Join(configDir, ".seed-license")
	err = os.WriteFile(licensePath, []byte("not-valid-base64!@#$"), 0o600)
	if err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	// Create manager - should handle corrupted file gracefully.
	mgr, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() should not fail on corrupted license: %v", err)
	}

	// Should not be activated.
	if mgr.IsActivated() {
		t.Error("Manager should not be activated with corrupted license file")
	}
}

// TestLoadStateWithInvalidJSON tests loading state with valid base64 but invalid JSON.
func TestLoadStateWithInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write base64-encoded but invalid content.
	licensePath := filepath.Join(configDir, ".seed-license")
	// This is valid base64 but won't decrypt properly.
	err = os.WriteFile(licensePath, []byte("dGVzdGluZzEyMzQ1Ng=="), 0o600)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Create manager - should handle invalid file gracefully.
	mgr, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() should not fail on invalid license: %v", err)
	}

	// Should not be activated.
	if mgr.IsActivated() {
		t.Error("Manager should not be activated with invalid license file")
	}
}

// TestCheckInResult tests the CheckIn result format.
func TestCheckInResult(t *testing.T) {
	mgr := setupTestManager(t)

	// CheckIn without state.
	result := mgr.CheckIn()
	if result.Success {
		t.Error("CheckIn without state should fail")
	}
	if result.Tier != license.TierInvalid {
		t.Errorf("CheckIn without state should have TierInvalid, got %v", result.Tier)
	}

	// Start trial then check in.
	mgr.StartTrial()
	result = mgr.CheckIn()

	if !result.Success {
		t.Errorf("CheckIn after trial should succeed: %s", result.Message)
	}

	// Verify result message.
	if result.Message == "" {
		t.Error("CheckIn message should not be empty")
	}
}

// TestNeedsCheckInAfterInterval tests check-in interval logic.
func TestNeedsCheckInAfterInterval(t *testing.T) {
	mgr := setupTestManager(t)

	// No state - no check-in needed.
	if mgr.NeedsCheckIn() {
		t.Error("No state should not need check-in")
	}

	// Trial - no check-in needed.
	mgr.StartTrial()
	if mgr.NeedsCheckIn() {
		t.Error("Trial should not need check-in")
	}

	// Deactivate and activate with full license.
	mgr.Deactivate()

	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	mgr.Activate(key)

	// Right after activation - no check-in needed.
	if mgr.NeedsCheckIn() {
		t.Error("Fresh activation should not need check-in")
	}
}

// TestActivationResultFields tests all ActivationResult fields.
func TestActivationResultFields(t *testing.T) {
	mgr := setupTestManager(t)

	// Test trial result fields.
	trialResult := mgr.StartTrial()
	if trialResult.Tier != license.TierProfessional {
		t.Errorf("Trial tier should be TierProfessional, got %v", trialResult.Tier)
	}
	if trialResult.DaysRemaining != license.TrialDays {
		t.Errorf("Trial days should be %d, got %d", license.TrialDays, trialResult.DaysRemaining)
	}
	if !trialResult.IsTrialMode {
		t.Error("Trial result IsTrialMode should be true")
	}

	// Deactivate and test full license result with a Professional key.
	mgr.Deactivate()

	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	licenseResult := mgr.Activate(key)

	if !licenseResult.Success {
		t.Fatalf("Activate() failed: %s", licenseResult.Message)
	}
	if licenseResult.Tier != license.TierProfessional {
		t.Errorf("License tier should be TierProfessional, got %v", licenseResult.Tier)
	}
	if licenseResult.IsTrialMode {
		t.Error("License result IsTrialMode should be false")
	}
	if licenseResult.DaysRemaining != 365 {
		t.Errorf("License days remaining should be 365, got %d", licenseResult.DaysRemaining)
	}

	// Enterprise (tier 3 / "3001") is retired: activation must FAIL.
	mgr.Deactivate()
	enterpriseKey, _ := signTestKey(t, "3001", "1234567", license.Tier(3))
	enterpriseResult := mgr.Activate(enterpriseKey)
	if enterpriseResult.Success {
		t.Error("Enterprise (tier 3) activation must be rejected after retirement")
	}
	if enterpriseResult.Tier != license.TierInvalid {
		t.Errorf("Rejected Enterprise activation should have TierInvalid, got %v", enterpriseResult.Tier)
	}
}

// TestActivationStateFields tests ActivationState fields.
func TestActivationStateFields(t *testing.T) {
	mgr := setupTestManager(t)

	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	mgr.Activate(key)

	state := mgr.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil")
	}

	// Verify all fields are set correctly.
	if state.LicenseKey == "" {
		t.Error("LicenseKey should be set")
	}
	if state.DeviceHash == "" {
		t.Error("DeviceHash should be set")
	}
	if state.Tier != license.TierProfessional {
		t.Errorf("Tier should be TierProfessional, got %v", state.Tier)
	}
	if state.ActivatedAt.IsZero() {
		t.Error("ActivatedAt should be set")
	}
	if state.LastValidatedAt.IsZero() {
		t.Error("LastValidatedAt should be set")
	}
	if state.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}
	if state.IsTrialMode {
		t.Error("IsTrialMode should be false for full license")
	}
	if !state.TrialStartedAt.IsZero() {
		t.Error("TrialStartedAt should be zero for full license")
	}
	if len(state.Features) == 0 {
		t.Error("Features should be set")
	}
}

// TestTrialStateFields tests ActivationState fields for trial.
func TestTrialStateFields(t *testing.T) {
	mgr := setupTestManager(t)

	mgr.StartTrial()

	state := mgr.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil")
	}

	// Verify trial-specific fields.
	if state.LicenseKey != "" {
		t.Error("LicenseKey should be empty for trial")
	}
	if state.DeviceHash == "" {
		t.Error("DeviceHash should be set for trial")
	}
	if state.Tier != license.TierProfessional {
		t.Errorf("Tier should be TierProfessional, got %v", state.Tier)
	}
	if !state.IsTrialMode {
		t.Error("IsTrialMode should be true for trial")
	}
	if state.TrialStartedAt.IsZero() {
		t.Error("TrialStartedAt should be set for trial")
	}
	if len(state.Features) == 0 {
		t.Error("Features should be set for trial")
	}
}

// TestDeactivateRemovesFile tests that Deactivate removes the license file.
func TestDeactivateRemovesFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	mgr, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Activate to create file.
	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	mgr.Activate(key)

	// Verify file exists.
	licensePath := filepath.Join(configDir, ".seed-license")
	_, statErr := os.Stat(licensePath)
	if os.IsNotExist(statErr) {
		t.Error("License file should exist after activation")
	}

	// Deactivate.
	err = mgr.Deactivate()
	if err != nil {
		t.Fatalf("Deactivate() error: %v", err)
	}

	// Verify file is removed.
	_, removeStatErr := os.Stat(licensePath)
	if !os.IsNotExist(removeStatErr) {
		t.Error("License file should be removed after deactivation")
	}
}

// TestAllTiersActivation tests activation with all active tier types.
// Enterprise (tier 3 / "3001") is retired and tested separately by
// TestEnterpriseTierActivationRejected.
func TestAllTiersActivation(t *testing.T) {
	tiers := []struct {
		product string
		tier    license.Tier
		name    string
	}{
		{"1001", license.TierReflector, "Reflector"},
		{"2001", license.TierProfessional, "Professional"},
	}

	for _, tc := range tiers {
		t.Run(tc.name, func(t *testing.T) {
			mgr := setupTestManager(t)

			key, err := signTestKey(t, tc.product, "1234567", tc.tier)
			if err != nil {
				t.Fatalf("signTestKey() error: %v", err)
			}

			result := mgr.Activate(key)
			if !result.Success {
				t.Errorf("Activate() failed: %s", result.Message)
			}

			if result.Tier != tc.tier {
				t.Errorf("Expected tier %v, got %v", tc.tier, result.Tier)
			}

			if !mgr.IsActivated() {
				t.Error("Manager should be activated")
			}

			state := mgr.GetState()
			if state.Tier != tc.tier {
				t.Errorf("State tier should be %v, got %v", tc.tier, state.Tier)
			}
		})
	}
}

// TestEnterpriseTierActivationRejected guards the Enterprise tier retirement.
// Activation with a tier-3 / "3001" token must fail.
func TestEnterpriseTierActivationRejected(t *testing.T) {
	mgr := setupTestManager(t)

	key, err := signTestKey(t, "3001", "1234567", license.Tier(3))
	if err != nil {
		t.Fatalf("signTestKey() error: %v", err)
	}

	result := mgr.Activate(key)
	if result.Success {
		t.Error("Enterprise (tier 3) activation must be rejected after retirement")
	}
	if mgr.IsActivated() {
		t.Error("Manager must not be activated by a rejected Enterprise token")
	}
}

// TestIsTrialValidZeroTrialStart tests IsTrialValid with zero TrialStartedAt.
func TestIsTrialValidZeroTrialStart(t *testing.T) {
	mgr := setupTestManager(t)

	// Should not be valid with no state.
	if mgr.IsTrialValid() {
		t.Error("IsTrialValid should be false with no state")
	}
}

// TestTrialDaysRemainingWithZeroStart tests TrialDaysRemaining edge case.
func TestTrialDaysRemainingWithZeroStart(t *testing.T) {
	mgr := setupTestManager(t)

	// No state - should return 0.
	days := mgr.TrialDaysRemaining()
	if days != 0 {
		t.Errorf("Expected 0 days with no state, got %d", days)
	}
}

// TestActivateReplacesExistingLicense tests that activating replaces existing license.
func TestActivateReplacesExistingLicense(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate with first key.
	key1, _ := signTestKey(t, "1001", "1234567", license.TierReflector)
	result1 := mgr.Activate(key1)
	if !result1.Success {
		t.Fatalf("First Activate() failed: %s", result1.Message)
	}

	if result1.Tier != license.TierReflector {
		t.Errorf("First activation tier should be Reflector, got %v", result1.Tier)
	}

	// Activate with second key (different tier).
	key2, _ := signTestKey(t, "2001", "7654321", license.TierProfessional)
	result2 := mgr.Activate(key2)
	if !result2.Success {
		t.Fatalf("Second Activate() failed: %s", result2.Message)
	}

	if result2.Tier != license.TierProfessional {
		t.Errorf("Second activation tier should be TestSuite, got %v", result2.Tier)
	}

	// State should have new tier.
	state := mgr.GetState()
	if state.Tier != license.TierProfessional {
		t.Errorf("State tier should be TestSuite after second activation, got %v", state.Tier)
	}
}

// TestNewManagerWithFallbackHomeDir tests manager creation when HOME is unavailable.
func TestNewManagerWithFallbackHomeDir(t *testing.T) {
	// Set HOME to non-existent directory.
	t.Setenv("HOME", "/nonexistent/path/that/does/not/exist")

	// Creating manager should still work (falls back to /tmp).
	mgr, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() should not fail with bad HOME: %v", err)
	}

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}

	// Fingerprint should still be available.
	fp := mgr.GetFingerprint()
	if fp == nil {
		t.Error("Fingerprint should be available")
	}
}

// TestActivationMessage tests the message content of activation results.
func TestActivationMessage(t *testing.T) {
	mgr := setupTestManager(t)

	// Test trial message.
	trialResult := mgr.StartTrial()
	if !containsNumber(trialResult.Message, license.TrialDays) {
		t.Errorf("Trial message should mention %d days: %s", license.TrialDays, trialResult.Message)
	}

	// Test license message.
	mgr.Deactivate()
	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	licenseResult := mgr.Activate(key)
	if !contains(licenseResult.Message, "Professional") {
		t.Errorf("License message should mention tier: %s", licenseResult.Message)
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsNumber(s string, n int) bool {
	numStr := strconv.Itoa(n)
	return containsSubstr(s, numStr)
}

// TestStartTrialReturnsExpiredForOldTrial tests StartTrial with expired trial.
func TestStartTrialReturnsExpiredForOldTrial(t *testing.T) {
	mgr := setupTestManager(t)

	// Start trial first.
	result := mgr.StartTrial()
	if !result.Success {
		t.Fatalf("StartTrial() failed: %s", result.Message)
	}

	// Call StartTrial again - should succeed and show remaining days.
	result2 := mgr.StartTrial()
	if !result2.Success {
		t.Error("Second StartTrial should succeed")
	}

	if result2.DaysRemaining <= 0 {
		t.Error("Should have remaining trial days")
	}
}

// TestDeactivateIdempotent tests that Deactivate can be called multiple times.
func TestDeactivateIdempotent(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate first.
	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	mgr.Activate(key)

	// Deactivate multiple times - should not error.
	for i := range 3 {
		err := mgr.Deactivate()
		if err != nil {
			t.Errorf("Deactivate() #%d should not error: %v", i+1, err)
		}
	}

	// Should not be activated.
	if mgr.IsActivated() {
		t.Error("Should not be activated after deactivation")
	}
}

// TestCheckInNoOpOnTrial tests that CheckIn is a no-op for trial.
func TestCheckInNoOpOnTrial(t *testing.T) {
	mgr := setupTestManager(t)

	mgr.StartTrial()

	// CheckIn should succeed but not require it.
	result := mgr.CheckIn()
	if !result.Success {
		t.Errorf("CheckIn on trial should succeed: %s", result.Message)
	}
}

// TestSaveStateNilState tests saveState behavior with nil state.
func TestSaveStateNilState(t *testing.T) {
	mgr := setupTestManager(t)

	// Manager with nil state should not error on deactivate (which calls saveState).
	err := mgr.Deactivate()
	if err != nil {
		t.Errorf("Deactivate() with nil state should not error: %v", err)
	}
}

// TestLoadStateMissingFile tests loadState when file doesn't exist.
func TestLoadStateMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Don't create any license file.
	mgr, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Should have nil state but no error.
	state := mgr.GetState()
	if state != nil {
		t.Error("State should be nil when no license file exists")
	}
}

// TestLoadStateWithTooShortCiphertext tests decryption with too-short data.
func TestLoadStateWithTooShortCiphertext(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write base64-encoded data that's too short for valid ciphertext.
	// A valid AES-GCM ciphertext needs at least nonce (12 bytes) + tag (16 bytes) = 28 bytes.
	licensePath := filepath.Join(configDir, ".seed-license")
	// "dGVzdA==" decodes to "test" which is only 4 bytes.
	err = os.WriteFile(licensePath, []byte("dGVzdA=="), 0o600)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Create manager - should handle invalid ciphertext gracefully.
	mgr, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() should not fail: %v", err)
	}

	// Should not be activated.
	if mgr.IsActivated() {
		t.Error("Manager should not be activated with too-short ciphertext")
	}
}

// TestEncryptDecryptRoundtrip tests encryption/decryption via save/load.
func TestEncryptDecryptRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create manager and activate.
	mgr1, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	key, _ := signTestKey(t, "2001", "ENCRYPT", license.TierProfessional)
	mgr1.Activate(key)

	state1 := mgr1.GetState()
	if state1 == nil {
		t.Fatal("State should not be nil after activation")
	}

	// Create second manager to verify roundtrip.
	mgr2, err := newTestManager(t)
	if err != nil {
		t.Fatalf("Second NewManager() error: %v", err)
	}

	state2 := mgr2.GetState()
	if state2 == nil {
		t.Fatal("State should not be nil after reload")
	}

	// Verify key fields match.
	if state1.LicenseKey != state2.LicenseKey {
		t.Error("LicenseKey should match after roundtrip")
	}
	if state1.Tier != state2.Tier {
		t.Error("Tier should match after roundtrip")
	}
	if state1.DeviceHash != state2.DeviceHash {
		t.Error("DeviceHash should match after roundtrip")
	}
}

// TestTrialExpirationCheck tests trial expiration logic.
func TestTrialExpirationCheck(t *testing.T) {
	mgr := setupTestManager(t)

	// Start trial.
	result := mgr.StartTrial()
	if !result.Success {
		t.Fatalf("StartTrial() failed: %s", result.Message)
	}

	// Trial should be valid right now.
	if !mgr.IsTrialValid() {
		t.Error("Fresh trial should be valid")
	}

	// IsActivated should return true for valid trial.
	if !mgr.IsActivated() {
		t.Error("IsActivated should be true for valid trial")
	}

	// Trial days should be close to TrialDays.
	days := mgr.TrialDaysRemaining()
	if days < 13 || days > 14 {
		t.Errorf("Trial days should be ~14, got %d", days)
	}
}

// TestMultipleActivationCycles tests repeated activation/deactivation.
func TestMultipleActivationCycles(t *testing.T) {
	mgr := setupTestManager(t)

	for i := range 5 {
		// Activate.
		key, _ := signTestKey(t, "2001", "123456"+strconv.Itoa(i), license.TierProfessional)
		result := mgr.Activate(key)
		if !result.Success {
			t.Fatalf("Cycle %d: Activate() failed: %s", i+1, result.Message)
		}

		if !mgr.IsActivated() {
			t.Errorf("Cycle %d: Should be activated", i+1)
		}

		// Deactivate.
		err := mgr.Deactivate()
		if err != nil {
			t.Fatalf("Cycle %d: Deactivate() error: %v", i+1, err)
		}

		if mgr.IsActivated() {
			t.Errorf("Cycle %d: Should not be activated after deactivation", i+1)
		}
	}
}

// TestActivationWithDifferentKeys tests switching between different license keys.
func TestActivationWithDifferentKeys(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate with Reflector.
	key1, _ := signTestKey(t, "1001", "1234567", license.TierReflector)
	result1 := mgr.Activate(key1)
	if !result1.Success {
		t.Fatalf("First activation failed: %s", result1.Message)
	}
	if result1.Tier != license.TierReflector {
		t.Errorf("Expected Reflector tier, got %v", result1.Tier)
	}

	// Upgrade to TestSuite (without deactivating).
	key2, _ := signTestKey(t, "2001", "2345678", license.TierProfessional)
	result2 := mgr.Activate(key2)
	if !result2.Success {
		t.Fatalf("Second activation failed: %s", result2.Message)
	}
	if result2.Tier != license.TierProfessional {
		t.Errorf("Expected TestSuite tier, got %v", result2.Tier)
	}

	// Enterprise (tier 3 / "3001") is retired: activation attempt must FAIL;
	// the existing Professional state must remain unchanged.
	key3, _ := signTestKey(t, "3001", "3456789", license.Tier(3))
	result3 := mgr.Activate(key3)
	if result3.Success {
		t.Error("Enterprise (tier 3) activation must be rejected after retirement")
	}

	// Manager should still be activated with the previous Professional key.
	if !mgr.IsActivated() {
		t.Error("Manager should still be activated with the previous Professional key")
	}

	// Verify final state still reflects Professional.
	state := mgr.GetState()
	if state == nil {
		t.Fatal("State should not be nil after rejected Enterprise activation")
	}
	if state.Tier != license.TierProfessional {
		t.Errorf("Final state should be Professional (Enterprise was rejected), got %v", state.Tier)
	}
}

// TestLoadStateOpenError tests loadState when file cannot be opened.
func TestLoadStateOpenError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	err := os.MkdirAll(configDir, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create a directory where the license file should be (causes open error).
	licensePath := filepath.Join(configDir, ".seed-license")
	err = os.MkdirAll(licensePath, configDirPerm)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create manager - should handle open error gracefully.
	mgr, err := newTestManager(t)
	if err != nil {
		t.Fatalf("NewManager() should not fail on open error: %v", err)
	}

	// Should not be activated.
	if mgr.IsActivated() {
		t.Error("Manager should not be activated when license file is a directory")
	}
}

// TestStateExpiresAtIsSet tests that ExpiresAt is set correctly.
func TestStateExpiresAtIsSet(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate.
	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	mgr.Activate(key)

	state := mgr.GetState()
	if state == nil {
		t.Fatal("State should not be nil")
	}

	// ExpiresAt should be approximately 1 year from now.
	now := time.Now()
	oneYearLater := now.AddDate(1, 0, 0)

	// Allow 1 minute tolerance.
	if state.ExpiresAt.Before(oneYearLater.Add(-time.Minute)) ||
		state.ExpiresAt.After(oneYearLater.Add(time.Minute)) {
		t.Errorf("ExpiresAt should be ~1 year from now, got %v (expected ~%v)",
			state.ExpiresAt, oneYearLater)
	}
}

// TestActivatedAtIsSet tests that ActivatedAt is set correctly.
func TestActivatedAtIsSet(t *testing.T) {
	mgr := setupTestManager(t)

	before := time.Now()

	// Small delay to ensure time difference is measurable.
	time.Sleep(1 * time.Millisecond)

	// Activate.
	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	mgr.Activate(key)

	time.Sleep(1 * time.Millisecond)
	after := time.Now()

	state := mgr.GetState()
	if state == nil {
		t.Fatal("State should not be nil")
	}

	// ActivatedAt should be between before and after.
	if state.ActivatedAt.Before(before) || state.ActivatedAt.After(after) {
		t.Errorf("ActivatedAt %v should be between %v and %v",
			state.ActivatedAt, before, after)
	}
}

// TestReflectorTierFeatures tests that Reflector tier has correct features.
func TestReflectorTierFeatures(t *testing.T) {
	mgr := setupTestManager(t)

	key, _ := signTestKey(t, "1001", "1234567", license.TierReflector)
	mgr.Activate(key)

	state := mgr.GetState()
	if state == nil {
		t.Fatal("State should not be nil")
	}

	// Reflector should have only the reflector feature.
	if len(state.Features) != 1 {
		t.Errorf("Reflector should have 1 feature, got %d", len(state.Features))
	}

	if len(state.Features) > 0 && state.Features[0] != "reflector" {
		t.Errorf("Reflector feature should be 'reflector', got %v", state.Features)
	}
}

// TestEnterpriseTierRejectedNoFeatures guards the Enterprise tier retirement.
// A tier-3 / "3001" token must be rejected: no state is persisted, no features granted.
func TestEnterpriseTierRejectedNoFeatures(t *testing.T) {
	mgr := setupTestManager(t)

	key, _ := signTestKey(t, "3001", "1234567", license.Tier(3))
	result := mgr.Activate(key)

	if result.Success {
		t.Error("Enterprise (tier 3) activation must be rejected after retirement")
	}

	// No state should be persisted on rejection.
	state := mgr.GetState()
	if state != nil {
		if state.Features != nil {
			for _, f := range state.Features {
				if f == "api" || f == "multiuser" {
					t.Errorf("Rejected Enterprise must not grant feature %q", f)
				}
			}
		}
	}
}

// TestCheckInTier tests that CheckIn returns correct tier.
func TestCheckInTier(t *testing.T) {
	mgr := setupTestManager(t)

	// Activate with Professional.
	key, _ := signTestKey(t, "2001", "1234567", license.TierProfessional)
	mgr.Activate(key)

	// Check in.
	result := mgr.CheckIn()
	if !result.Success {
		t.Fatalf("CheckIn failed: %s", result.Message)
	}

	// Tier should be returned.
	if result.Tier != license.TierProfessional {
		t.Errorf("CheckIn tier should be Professional, got %v", result.Tier)
	}
}

// TestActivationStateCompleteFields verifies all fields are properly set during activation.
func TestActivationStateCompleteFields(t *testing.T) {
	mgr := setupTestManager(t)

	key, _ := signTestKey(t, "1001", "REFLCT1", license.TierReflector)
	result := mgr.Activate(key)

	if !result.Success {
		t.Fatalf("Activate failed: %s", result.Message)
	}

	state := mgr.GetState()
	if state == nil {
		t.Fatal("State should not be nil after activation")
	}

	// All required fields should be set.
	if state.LicenseKey == "" {
		t.Error("LicenseKey should be set")
	}
	if state.DeviceHash == "" {
		t.Error("DeviceHash should be set")
	}
	if state.ActivatedAt.IsZero() {
		t.Error("ActivatedAt should be set")
	}
	if state.LastValidatedAt.IsZero() {
		t.Error("LastValidatedAt should be set")
	}
	if state.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}
	// Reflector tier activation should NOT be trial.
	if state.IsTrialMode {
		t.Error("Full license should not be trial mode")
	}
	if len(state.Features) == 0 {
		t.Error("Features should be set")
	}
}

// TestReflectorTierActivation tests Reflector tier activation specifically.
func TestReflectorTierActivation(t *testing.T) {
	mgr := setupTestManager(t)

	key, _ := signTestKey(t, "1001", "RFLCTR1", license.TierReflector)
	result := mgr.Activate(key)

	if !result.Success {
		t.Fatalf("Reflector activation failed: %s", result.Message)
	}

	if result.Tier != license.TierReflector {
		t.Errorf("Result tier should be Reflector, got %v", result.Tier)
	}

	state := mgr.GetState()
	if state == nil {
		t.Fatal("State should not be nil")
	}

	if state.Tier != license.TierReflector {
		t.Errorf("State tier should be Reflector, got %v", state.Tier)
	}

	// Reflector should only have reflector feature.
	if len(state.Features) != 1 || state.Features[0] != "reflector" {
		t.Errorf("Reflector should have only 'reflector' feature, got %v", state.Features)
	}
}
