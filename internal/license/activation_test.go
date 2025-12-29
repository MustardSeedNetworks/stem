// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package license

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	// Use temp directory for test
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create .config/seed-test-suite directory
	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestManagerGetFingerprint(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	fp := mgr.GetFingerprint()
	if fp == nil {
		t.Error("GetFingerprint() returned nil")
	}
	hash := fp.Hash()
	if len(hash) < 8 {
		t.Errorf("GetFingerprint().Hash() too short: %s", hash)
	}
}

func TestManagerIsActivated(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Fresh install should not be activated
	if mgr.IsActivated() {
		t.Error("Fresh manager should not be activated")
	}
}

func TestManagerStartTrial(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	result := mgr.StartTrial()
	if !result.Success {
		t.Errorf("StartTrial() failed: %s", result.Message)
	}

	// Should now be activated in trial mode
	if !mgr.IsActivated() {
		t.Error("Should be activated after starting trial")
	}

	if !mgr.IsTrialValid() {
		t.Error("Trial should be valid")
	}

	days := mgr.TrialDaysRemaining()
	if days < 13 || days > 14 {
		t.Errorf("Expected ~14 trial days, got %d", days)
	}
}

func TestManagerTrialExpiry(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Manually set expired trial
	mgr.state = &ActivationState{
		TrialStartedAt: time.Now().Add(-15 * 24 * time.Hour),
		Tier:           TierTestSuite,
		IsTrialMode:    true,
		DeviceHash:     mgr.fingerprint.Hash(),
	}

	if mgr.IsTrialValid() {
		t.Error("Expired trial should not be valid")
	}

	days := mgr.TrialDaysRemaining()
	if days != 0 {
		t.Errorf("Expired trial should have 0 days remaining, got %d", days)
	}
}

func TestManagerActivate(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Generate a valid key for TestSuite tier
	key, err := GenerateLicenseKey("2001", "1234567", TierTestSuite)
	if err != nil {
		t.Fatalf("GenerateLicenseKey() error: %v", err)
	}

	result := mgr.Activate(key)
	if !result.Success {
		t.Errorf("Activate() failed: %s", result.Message)
	}

	if !mgr.IsActivated() {
		t.Error("Should be activated after license activation")
	}

	state := mgr.GetState()
	if state.Tier != TierTestSuite {
		t.Errorf("Expected tier %d, got %d", TierTestSuite, state.Tier)
	}
}

func TestManagerActivateInvalidKey(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	result := mgr.Activate("INVALID-KEY-1234")
	if result.Success {
		t.Error("Invalid key should not activate")
	}
}

func TestManagerDeactivate(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// First activate
	mgr.StartTrial()
	if !mgr.IsActivated() {
		t.Fatal("Should be activated")
	}

	// Then deactivate
	err = mgr.Deactivate()
	if err != nil {
		t.Errorf("Deactivate() error: %v", err)
	}

	if mgr.IsActivated() {
		t.Error("Should not be activated after deactivation")
	}
}

func TestManagerGetState(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Start trial first to have a state
	mgr.StartTrial()

	state := mgr.GetState()
	if state == nil {
		t.Fatal("GetState() returned nil")
	}

	// Device hash should be set
	if state.DeviceHash == "" {
		t.Error("DeviceHash should not be empty")
	}
}

func TestManagerNeedsCheckIn(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// No state = no check-in needed
	if mgr.NeedsCheckIn() {
		t.Error("No state should not need check-in")
	}

	// Trial mode shouldn't need check-in
	mgr.StartTrial()
	if mgr.NeedsCheckIn() {
		t.Error("Trial mode should not need check-in")
	}
}

func TestManagerCheckIn(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// No state
	result := mgr.CheckIn()
	if result.Success {
		t.Error("CheckIn with no state should fail")
	}

	// With state
	mgr.StartTrial()
	result = mgr.CheckIn()
	if !result.Success {
		t.Errorf("CheckIn with state failed: %s", result.Message)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	testCases := []string{
		"Hello, World!",
		"Short",
		"A longer string with special characters: !@#$%^&*()",
		`{"key": "value", "number": 123}`,
	}

	for _, original := range testCases {
		encrypted, err := mgr.encrypt([]byte(original))
		if err != nil {
			t.Errorf("encrypt(%q) error: %v", original, err)
			continue
		}

		decrypted, err := mgr.decrypt(encrypted)
		if err != nil {
			t.Errorf("decrypt() error: %v", err)
			continue
		}

		if string(decrypted) != original {
			t.Errorf("Roundtrip failed: got %q, want %q", string(decrypted), original)
		}
	}
}

func TestDeviceFingerprintString(t *testing.T) {
	fp := &DeviceFingerprint{
		MACAddress: "00:11:22:33:44:55",
		CPUSerial:  "CPU12345",
		DiskSerial: "DISK6789",
		Hostname:   "testhost",
		Platform:   "linux",
	}

	s := fp.String()
	if s == "" {
		t.Error("Fingerprint String() should not be empty")
	}
	if len(s) < 20 {
		t.Error("Fingerprint String() seems too short")
	}
}

func TestDeviceFingerprintHash(t *testing.T) {
	fp := &DeviceFingerprint{
		MACAddress: "00:11:22:33:44:55",
		CPUSerial:  "CPU12345",
		DiskSerial: "DISK6789",
		Hostname:   "testhost",
		Platform:   "linux",
	}

	hash := fp.Hash()
	if len(hash) != 16 {
		t.Errorf("Expected 16-char hash, got %d chars", len(hash))
	}

	// Same input should produce same hash
	hash2 := fp.Hash()
	if hash != hash2 {
		t.Error("Hash should be deterministic")
	}
}

func TestRotorCipherEncodeDecode(t *testing.T) {
	// Test roundtrip encoding/decoding
	testCases := []struct {
		input    string
		startPos int
	}{
		{"HELLO", 0},
		{"12345", 7},
		{"ABCD1234", 15},
		{"test123", 25},
	}

	for _, tc := range testCases {
		encoder := NewRotorCipher(tc.startPos)
		encoded := encoder.EncodeString(tc.input)

		decoder := NewRotorCipher(tc.startPos)
		decoded := decoder.DecodeString(encoded)

		if decoded != tc.input {
			t.Errorf("Roundtrip failed: input=%q, encoded=%q, decoded=%q", tc.input, encoded, decoded)
		}
	}
}

func TestRotorCipherNonAlpha(t *testing.T) {
	cipher := NewRotorCipher(0)
	// Non-alphanumeric characters should pass through unchanged
	input := "TEST-123!"
	encoded := cipher.EncodeString(input)
	if encoded[4] != '-' || encoded[8] != '!' {
		t.Error("Non-alphanumeric characters should pass through")
	}
}

func TestCalculateChecksumDeterministic(t *testing.T) {
	// Checksum should be consistent
	cs1 := CalculateChecksum("HELLO")
	cs2 := CalculateChecksum("HELLO")
	if cs1 != cs2 {
		t.Error("Checksum should be deterministic")
	}

	// Different inputs should (usually) produce different checksums
	cs3 := CalculateChecksum("WORLD")
	if cs1 == cs3 {
		t.Log("Warning: collision detected (rare but possible)")
	}

	// Checksum should be 2 characters
	if len(cs1) != 2 {
		t.Errorf("Checksum should be 2 chars, got %d", len(cs1))
	}
}

func TestValidateChecksumRoundtrip(t *testing.T) {
	// Generate valid checksum
	payload := "TEST1234"
	checksum := CalculateChecksum(payload)
	valid := payload + checksum

	if !ValidateChecksum(valid) {
		t.Error("Valid checksum should validate")
	}

	// Invalid checksum
	if ValidateChecksum(payload + "XX") {
		t.Error("Invalid checksum should not validate")
	}

	// Too short
	if ValidateChecksum("AB") {
		t.Error("Too short string should not validate")
	}
}

func TestManagerStartTrialTwice(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Start trial first time
	result := mgr.StartTrial()
	if !result.Success {
		t.Errorf("First StartTrial() failed: %s", result.Message)
	}

	// Start trial second time - should succeed but show remaining days
	result2 := mgr.StartTrial()
	if !result2.Success {
		t.Errorf("Second StartTrial() failed: %s", result2.Message)
	}
	if result2.DaysRemaining <= 0 {
		t.Error("Should show remaining days")
	}
}

func TestManagerActivateExpiredLicense(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Set an expired license state
	mgr.state = &ActivationState{
		LicenseKey:  "TESTKEY",
		DeviceHash:  mgr.fingerprint.Hash(),
		Tier:        TierTestSuite,
		ExpiresAt:   time.Now().Add(-1 * 24 * time.Hour), // Expired yesterday
		IsTrialMode: false,
	}

	// Should not be activated with expired license
	if mgr.IsActivated() {
		t.Error("Expired license should not be activated")
	}
}

func TestManagerDeviceBindingCheck(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Set a license with wrong device hash
	mgr.state = &ActivationState{
		LicenseKey:  "TESTKEY",
		DeviceHash:  "WRONGDEVICEHASH1", // Different from actual device
		Tier:        TierTestSuite,
		ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
		IsTrialMode: false,
	}

	// Should not be activated with wrong device
	if mgr.IsActivated() {
		t.Error("License with wrong device hash should not be activated")
	}
}

func TestTrialDaysRemainingNoState(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// No state = 0 days
	days := mgr.TrialDaysRemaining()
	if days != 0 {
		t.Errorf("Expected 0 days with no state, got %d", days)
	}
}

func TestTrialDaysRemainingNotTrial(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Full license (not trial)
	mgr.state = &ActivationState{
		Tier:        TierTestSuite,
		IsTrialMode: false,
	}

	days := mgr.TrialDaysRemaining()
	if days != 0 {
		t.Errorf("Expected 0 days for non-trial, got %d", days)
	}
}

func TestIsTrialValidNoState(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// No state = not valid
	if mgr.IsTrialValid() {
		t.Error("No state should mean trial not valid")
	}
}

func TestIsTrialValidZeroTrialStart(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Trial mode but no start time
	mgr.state = &ActivationState{
		IsTrialMode: true,
		// TrialStartedAt is zero
	}

	if mgr.IsTrialValid() {
		t.Error("Zero trial start time should not be valid")
	}
}

func TestStartTrialAlreadyActivated(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Activate with full license
	key, _ := GenerateLicenseKey("2001", "1234567", TierTestSuite)
	mgr.Activate(key)

	// Try to start trial - should succeed but return existing license info
	result := mgr.StartTrial()
	if !result.Success {
		t.Errorf("StartTrial on activated license should succeed: %s", result.Message)
	}
}

func TestStartTrialExpired(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Set expired trial
	mgr.state = &ActivationState{
		TrialStartedAt: time.Now().Add(-30 * 24 * time.Hour), // 30 days ago
		IsTrialMode:    true,
		DeviceHash:     mgr.fingerprint.Hash(),
	}

	// Try to start trial again - should fail
	result := mgr.StartTrial()
	if result.Success {
		t.Error("StartTrial on expired trial should fail")
	}
}

func TestGenerateLicenseKeyAllTiers(t *testing.T) {
	tiers := []struct {
		product string
		tier    Tier
	}{
		{"1001", TierReflector},
		{"2001", TierTestSuite},
		{"3001", TierEnterprise},
	}

	for _, tc := range tiers {
		key, err := GenerateLicenseKey(tc.product, "ABCDEFG", tc.tier)
		if err != nil {
			t.Errorf("GenerateLicenseKey(%s, %d) error: %v", tc.product, tc.tier, err)
			continue
		}
		if len(key) != 16 {
			t.Errorf("Expected 16-char key, got %d chars", len(key))
		}
	}
}

func TestNeedsCheckInWithOldValidation(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "seed-test-suite")
	os.MkdirAll(configDir, 0755)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}

	// Set a full license with old validation time
	mgr.state = &ActivationState{
		Tier:            TierTestSuite,
		IsTrialMode:     false,
		LastValidatedAt: time.Now().Add(-60 * 24 * time.Hour), // 60 days ago
		DeviceHash:      mgr.fingerprint.Hash(),
		ExpiresAt:       time.Now().Add(365 * 24 * time.Hour),
	}

	if !mgr.NeedsCheckIn() {
		t.Error("Should need check-in after 30+ days")
	}
}
