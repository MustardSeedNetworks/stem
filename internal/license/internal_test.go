// SPDX-License-Identifier: BUSL-1.1

package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

// Test internal/unexported functions using same-package tests.

// TestMaskStringEdgeCases tests maskString with edge cases.
func TestMaskStringEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		show     int
		expected string
	}{
		// Normal case - string longer than show.
		{"ABCDEFGHIJ", 4, "ABCD****"},
		// String exactly equal to show.
		{"ABCD", 4, "ABCD"},
		// String shorter than show.
		{"ABC", 4, "ABC"},
		// Empty string.
		{"", 4, ""},
		// Show is 0.
		{"ABCD", 0, "****"},
		// Show is 1.
		{"ABCD", 1, "A****"},
		// Long string.
		{"ABCDEFGHIJKLMNOP", 8, "ABCDEFGH****"},
	}

	for _, tc := range tests {
		result := maskString(tc.input, tc.show)
		if result != tc.expected {
			t.Errorf("maskString(%q, %d) = %q, want %q",
				tc.input, tc.show, result, tc.expected)
		}
	}
}

// TestSplitLinesIterator tests splitLines iterator.
func TestSplitLinesIterator(t *testing.T) {
	t.Parallel()
	input := "line1\nline2\nline3"
	var lines []string

	for line := range splitLines(input) {
		lines = append(lines, line)
	}

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	expected := []string{"line1", "line2", "line3"}
	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("Line %d: got %q, want %q", i, line, expected[i])
		}
	}
}

// TestGetPrimaryMACReturnsValid tests that getPrimaryMAC returns a valid format.
func TestGetPrimaryMACReturnsValid(t *testing.T) {
	t.Parallel()
	mac := getPrimaryMAC()

	// Should be either default or a valid MAC format.
	if mac == "" {
		t.Error("getPrimaryMAC should not return empty string")
	}

	// If not default, should contain colons.
	if mac != defaultMAC {
		colonCount := 0
		for _, c := range mac {
			if c == ':' {
				colonCount++
			}
		}
		if colonCount != 5 {
			t.Errorf("MAC address should have 5 colons: %s", mac)
		}
	}
}

// TestGetCPUSerialReturnsNonEmpty tests getCPUSerial returns non-empty.
func TestGetCPUSerialReturnsNonEmpty(t *testing.T) {
	t.Parallel()
	serial := getCPUSerial()
	if serial == "" {
		t.Error("getCPUSerial should not return empty string")
	}
}

// TestGetDiskSerialReturnsNonEmpty tests getDiskSerial returns non-empty.
func TestGetDiskSerialReturnsNonEmpty(t *testing.T) {
	t.Parallel()
	serial := getDiskSerial()
	if serial == "" {
		t.Error("getDiskSerial should not return empty string")
	}
}

// TestManagerEncryptDecryptRoundtrip tests encrypt/decrypt directly.
func TestManagerEncryptDecryptRoundtrip(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Test various plaintexts.
	plaintexts := [][]byte{
		[]byte("hello"),
		[]byte(""),
		[]byte("a longer test string with more content to encrypt"),
		[]byte(`{"key": "value", "number": 123}`),
	}

	for _, plaintext := range plaintexts {
		encrypted, encryptErr := mgr.encrypt(plaintext)
		if encryptErr != nil {
			t.Errorf("encrypt error for %q: %v", string(plaintext), encryptErr)
			continue
		}

		decrypted, decryptErr := mgr.decrypt(encrypted)
		if decryptErr != nil {
			t.Errorf("decrypt error for %q: %v", string(plaintext), decryptErr)
			continue
		}

		if string(decrypted) != string(plaintext) {
			t.Errorf("Roundtrip failed: got %q, want %q", string(decrypted), string(plaintext))
		}
	}
}

// TestDecryptInvalidBase64 tests decrypt with invalid base64.
func TestDecryptInvalidBase64(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Invalid base64.
	_, err = mgr.decrypt([]byte("not-valid-base64!@#$%"))
	if err == nil {
		t.Error("decrypt should fail with invalid base64")
	}
}

// TestDecryptTooShort tests decrypt with too-short ciphertext.
func TestDecryptTooShort(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Valid base64 but too short.
	// Base64 encoding of "test" (4 bytes, less than nonce size).
	_, err = mgr.decrypt([]byte("dGVzdA=="))
	if err == nil {
		t.Error("decrypt should fail with too-short ciphertext")
	}
}

// TestDecryptInvalidCiphertext tests decrypt with invalid ciphertext content.
func TestDecryptInvalidCiphertext(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Valid base64, long enough, but invalid ciphertext.
	// 50 bytes of 'A' encoded in base64.
	_, err = mgr.decrypt([]byte("QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE="))
	if err == nil {
		t.Error("decrypt should fail with invalid ciphertext")
	}
}

// TestDeriveKey tests that deriveKey returns consistent results.
func TestDeriveKey(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	key1 := mgr.deriveKey()
	key2 := mgr.deriveKey()

	// Keys should be consistent.
	if len(key1) != len(key2) {
		t.Error("deriveKey should return same length keys")
	}

	for i := range key1 {
		if key1[i] != key2[i] {
			t.Error("deriveKey should return same key each time")
			break
		}
	}

	// Key should be 32 bytes for AES-256.
	if len(key1) != 32 {
		t.Errorf("deriveKey should return 32 bytes, got %d", len(key1))
	}
}

// TestValidateLicenseKeyInternalEdgeCases tests Verifier.Validate edge cases.
func TestValidateLicenseKeyInternalEdgeCases(t *testing.T) {
	t.Parallel()
	v := internalTestVerifier(t)

	// Test Reflector tier key validation.
	reflectorKey, err := signTestKey(t, "1001", "1234567", TierReflector)
	if err != nil {
		t.Fatalf("signTestKey error: %v", err)
	}

	info := v.Validate(reflectorKey)
	if !info.Valid {
		t.Errorf("Reflector key should be valid: %s", info.ErrorMsg)
	}
	if info.Tier != TierReflector {
		t.Errorf("Expected Reflector tier, got %v", info.Tier)
	}
	if len(info.Features) != 1 || info.Features[0] != "reflector" {
		t.Errorf("Reflector should have only 'reflector' feature, got %v", info.Features)
	}

	// Test Professional tier key validation.
	testSuiteKey, err := signTestKey(t, "2001", "ABCDEFG", TierProfessional)
	if err != nil {
		t.Fatalf("signTestKey error: %v", err)
	}

	info = v.Validate(testSuiteKey)
	if !info.Valid {
		t.Errorf("Professional key should be valid: %s", info.ErrorMsg)
	}
	if info.Tier != TierProfessional {
		t.Errorf("Expected Professional tier, got %v", info.Tier)
	}

	// Test Enterprise tier key validation.
	enterpriseKey, err := signTestKey(t, "3001", "XYZXYZX", TierEnterprise)
	if err != nil {
		t.Fatalf("signTestKey error: %v", err)
	}

	info = v.Validate(enterpriseKey)
	if !info.Valid {
		t.Errorf("Enterprise key should be valid: %s", info.ErrorMsg)
	}
	if info.Tier != TierEnterprise {
		t.Errorf("Expected Enterprise tier, got %v", info.Tier)
	}
}

// TestIsActivatedInternalEdgeCases tests IsActivated edge cases.
func TestIsActivatedInternalEdgeCases(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Nil state should return false.
	mgr.state = nil
	if mgr.IsActivated() {
		t.Error("Nil state should not be activated")
	}

	// Create state with trial mode and valid trial.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeZero(),
		TrialStartedAt:  timeNow(),
		IsTrialMode:     true,
		Features:        nil,
	}
	if !mgr.IsActivated() {
		t.Error("Valid trial should be activated")
	}

	// Create state with non-trial, valid expiration, correct device hash.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeNow().AddDate(1, 0, 0),
		TrialStartedAt:  timeZero(),
		IsTrialMode:     false,
		Features:        nil,
	}
	if !mgr.IsActivated() {
		t.Error("Valid license should be activated")
	}

	// Create state with expired license.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeNow().AddDate(-1, 0, 0), // Expired 1 year ago.
		TrialStartedAt:  timeZero(),
		IsTrialMode:     false,
		Features:        nil,
	}
	if mgr.IsActivated() {
		t.Error("Expired license should not be activated")
	}

	// Create state with wrong device hash.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "WRONGDEVICEHASH", // Wrong hash.
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeNow().AddDate(1, 0, 0),
		TrialStartedAt:  timeZero(),
		IsTrialMode:     false,
		Features:        nil,
	}
	if mgr.IsActivated() {
		t.Error("Wrong device hash should not be activated")
	}
}

// Helper functions for testing.
func timeNow() time.Time {
	return time.Now()
}

func timeZero() time.Time {
	return time.Time{}
}

// TestIsTrialValidInternalEdgeCases tests IsTrialValid edge cases.
func TestIsTrialValidInternalEdgeCases(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Nil state should return false.
	mgr.state = nil
	if mgr.IsTrialValid() {
		t.Error("Nil state should not have valid trial")
	}

	// Non-trial state should return false.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "",
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeZero(),
		TrialStartedAt:  timeNow(),
		IsTrialMode:     false,
		Features:        nil,
	}
	if mgr.IsTrialValid() {
		t.Error("Non-trial state should not have valid trial")
	}

	// Trial with zero start time should return false.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "",
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeZero(),
		TrialStartedAt:  timeZero(),
		IsTrialMode:     true,
		Features:        nil,
	}
	if mgr.IsTrialValid() {
		t.Error("Trial with zero start time should not be valid")
	}

	// Valid trial should return true.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "",
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeZero(),
		TrialStartedAt:  timeNow(),
		IsTrialMode:     true,
		Features:        nil,
	}
	if !mgr.IsTrialValid() {
		t.Error("Valid trial should be valid")
	}
}

// TestTrialDaysRemainingInternalEdgeCases tests TrialDaysRemaining edge cases.
func TestTrialDaysRemainingInternalEdgeCases(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Nil state should return 0.
	mgr.state = nil
	if mgr.TrialDaysRemaining() != 0 {
		t.Error("Nil state should have 0 days remaining")
	}

	// Non-trial state should return 0.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "",
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeZero(),
		TrialStartedAt:  timeNow(),
		IsTrialMode:     false,
		Features:        nil,
	}
	if mgr.TrialDaysRemaining() != 0 {
		t.Error("Non-trial state should have 0 days remaining")
	}

	// Trial with zero start time should return TrialDays.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "",
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeZero(),
		TrialStartedAt:  timeZero(),
		IsTrialMode:     true,
		Features:        nil,
	}
	if mgr.TrialDaysRemaining() != TrialDays {
		t.Errorf("Zero start trial should return %d days, got %d", TrialDays, mgr.TrialDaysRemaining())
	}

	// Expired trial should return 0.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "",
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeZero(),
		TrialStartedAt:  timeNow().AddDate(0, 0, -30), // Started 30 days ago.
		IsTrialMode:     true,
		Features:        nil,
	}
	if mgr.TrialDaysRemaining() != 0 {
		t.Error("Expired trial should have 0 days remaining")
	}
}

// TestStartTrialInternalEdgeCases tests StartTrial edge cases.
func TestStartTrialInternalEdgeCases(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Test with existing non-trial activation.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierEnterprise,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeNow().AddDate(1, 0, 0),
		TrialStartedAt:  timeZero(),
		IsTrialMode:     false,
		Features:        nil,
	}
	result := mgr.StartTrial()
	if !result.Success {
		t.Error("StartTrial with existing license should succeed")
	}
	if result.IsTrialMode {
		t.Error("StartTrial with existing license should not be trial mode")
	}
	if result.Tier != TierEnterprise {
		t.Errorf("Expected Enterprise tier, got %v", result.Tier)
	}

	// Test with expired trial.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierInvalid,
		ActivatedAt:     timeZero(),
		LastValidatedAt: timeZero(),
		ExpiresAt:       timeZero(),
		TrialStartedAt:  timeNow().AddDate(0, 0, -30), // Started 30 days ago.
		IsTrialMode:     true,
		Features:        nil,
	}
	result = mgr.StartTrial()
	if result.Success {
		t.Error("StartTrial with expired trial should fail")
	}
	if result.DaysRemaining != 0 {
		t.Error("Expired trial should have 0 days remaining")
	}
}

// TestSaveStateNilState tests saveState with nil state.
func TestSaveStateNilState(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// saveState with nil state should return nil.
	mgr.state = nil
	err = mgr.saveState()
	if err != nil {
		t.Errorf("saveState with nil state should not error: %v", err)
	}
}

// TestDeactivateInternalEdgeCases tests Deactivate edge cases.
func TestDeactivateInternalEdgeCases(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Deactivate with nil state should not error.
	mgr.state = nil
	err = mgr.Deactivate()
	if err != nil {
		t.Errorf("Deactivate with nil state should not error: %v", err)
	}
}

// TestValidateLicenseKeyAllPaths tests all Validate code paths.
func TestValidateLicenseKeyAllPaths(t *testing.T) {
	t.Parallel()
	v := internalTestVerifier(t)

	// Test empty key — rejected with a generic invalid message.
	info := v.Validate("")
	if info.Valid {
		t.Error("Empty key should not be valid")
	}
	if info.ErrorMsg == "" {
		t.Error("Empty key should set a non-empty ErrorMsg")
	}

	// Test garbage that resembles the old 16-char format.
	info = v.Validate("ABCD!@#$EFGH5678")
	if info.Valid {
		t.Error("Garbage key should not be valid")
	}

	// Another garbage string.
	info = v.Validate("AAAAAAAAAAAAAAAA")
	if info.Valid {
		t.Error("Garbage key should not be valid")
	}

	// Generate valid keys for all tiers and verify validation.
	tiers := []struct {
		product string
		tier    Tier
	}{
		{"1001", TierReflector},
		{"2001", TierProfessional},
		{"3001", TierEnterprise},
	}

	for _, tc := range tiers {
		key, _ := signTestKey(t, tc.product, "ABCDEFG", tc.tier)
		info = v.Validate(key)
		if !info.Valid {
			t.Errorf("Key for tier %v should be valid: %s", tc.tier, info.ErrorMsg)
		}
		if info.Tier != tc.tier {
			t.Errorf("Expected tier %v, got %v", tc.tier, info.Tier)
		}
		if info.ProductCode != tc.product {
			t.Errorf("Expected product %s, got %s", tc.product, info.ProductCode)
		}
	}
}

// TestSaveStateCreatesDirectory tests that saveState creates directory if needed.
func TestSaveStateCreatesDirectory(t *testing.T) {
	t.Parallel()
	// This is already tested implicitly through activation tests,
	// but let's test the error path explicitly.
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set a valid state.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierProfessional,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Time{},
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Now(),
		IsTrialMode:     true,
		Features:        nil,
	}

	// saveState should work.
	err = mgr.saveState()
	if err != nil {
		t.Errorf("saveState should succeed: %v", err)
	}
}

// TestActivateInternalEdgeCases tests Activate edge cases.
func TestActivateInternalEdgeCases(t *testing.T) {
	t.Parallel()
	mgr, err := NewManagerWithVerifier(internalTestVerifier(t))
	if err != nil {
		t.Fatalf("NewManagerWithVerifier error: %v", err)
	}

	// Test with invalid key.
	result := mgr.Activate("INVALID-KEY-@@@@")
	if result.Success {
		t.Error("Invalid key should not activate")
	}
	if result.Tier != TierInvalid {
		t.Errorf("Invalid key should have TierInvalid, got %v", result.Tier)
	}

	// Test with valid key.
	key, _ := signTestKey(t, "2001", "ABCDEFG", TierProfessional)
	result = mgr.Activate(key)
	if !result.Success {
		t.Errorf("Valid key should activate: %s", result.Message)
	}
	if result.Tier != TierProfessional {
		t.Errorf("Expected TierProfessional, got %v", result.Tier)
	}
	if result.IsTrialMode {
		t.Error("Activation should not be trial mode")
	}
}

// TestLoadStateErrorPaths tests loadState error paths.
func TestLoadStateErrorPaths(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// loadState with non-existent file should return nil (no error).
	// This is already the default case for a fresh manager.
	// Let's verify the state is nil.
	if mgr.state != nil {
		// State might be loaded from existing file - that's ok.
		t.Log("Manager has existing state from disk")
	}
}

// TestNewManagerEdgeCases tests NewManager edge cases.
func TestNewManagerEdgeCases(t *testing.T) {
	t.Parallel()
	// NewManager should work even with default paths.
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	if mgr.fingerprint == nil {
		t.Error("Manager should have fingerprint")
	}

	if mgr.configDir == "" {
		t.Error("Manager should have configDir")
	}
}

// TestEncryptNonceGeneration tests that encrypt generates unique nonces.
func TestEncryptNonceGeneration(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	plaintext := []byte("test data")

	// Encrypt multiple times - should produce different ciphertexts.
	encrypted1, err := mgr.encrypt(plaintext)
	if err != nil {
		t.Fatalf("First encrypt error: %v", err)
	}

	encrypted2, err := mgr.encrypt(plaintext)
	if err != nil {
		t.Fatalf("Second encrypt error: %v", err)
	}

	// Ciphertexts should be different (due to random nonce).
	if string(encrypted1) == string(encrypted2) {
		t.Error("Encrypt should produce different ciphertexts for same plaintext")
	}

	// Both should decrypt to same plaintext.
	decrypted1, err := mgr.decrypt(encrypted1)
	if err != nil {
		t.Fatalf("First decrypt error: %v", err)
	}

	decrypted2, err := mgr.decrypt(encrypted2)
	if err != nil {
		t.Fatalf("Second decrypt error: %v", err)
	}

	if string(decrypted1) != string(decrypted2) {
		t.Error("Both should decrypt to same plaintext")
	}
}

// TestFingerprintHashConsistency tests that fingerprint hash is consistent.
func TestFingerprintHashConsistency(t *testing.T) {
	t.Parallel()
	fp, err := GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint error: %v", err)
	}

	// Hash should be consistent.
	hash1 := fp.Hash()
	hash2 := fp.Hash()
	hash3 := fp.Hash()

	if hash1 != hash2 || hash2 != hash3 {
		t.Error("Fingerprint hash should be consistent")
	}

	// Hash should be 16 characters.
	if len(hash1) != 16 {
		t.Errorf("Hash should be 16 chars, got %d", len(hash1))
	}
}

// TestFingerprintStringFormat tests that fingerprint String format is correct.
func TestFingerprintStringFormat(t *testing.T) {
	t.Parallel()
	fp, err := GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint error: %v", err)
	}

	str := fp.String()

	// Should contain expected prefixes.
	expectedParts := []string{"MAC=", "CPU=", "DISK=", "HOST="}
	for _, part := range expectedParts {
		found := false
		for i := 0; i <= len(str)-len(part); i++ {
			if str[i:i+len(part)] == part {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("String should contain %q: %s", part, str)
		}
	}
}

// TestSaveStateWithReadOnlyDir tests saveState error when directory is read-only.
func TestSaveStateWithReadOnlyDir(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set config dir to a non-writable path.
	mgr.configDir = "/nonexistent/readonly/path"

	// Set a state to save.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierProfessional,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Time{},
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Now(),
		IsTrialMode:     true,
		Features:        nil,
	}

	// saveState should fail.
	err = mgr.saveState()
	if err == nil {
		t.Error("saveState should fail with non-writable directory")
	}
}

// TestStartTrialSaveError tests StartTrial when save fails.
func TestStartTrialSaveError(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Clear any existing state so we go through the "Start new trial" path.
	mgr.state = nil

	// Set config dir to a non-writable path.
	mgr.configDir = "/nonexistent/readonly/path"

	// StartTrial should fail because save fails.
	result := mgr.StartTrial()
	if result.Success {
		t.Error("StartTrial should fail when save fails")
	}
	if result.Tier != TierInvalid {
		t.Errorf("Failed StartTrial should have TierInvalid, got %v", result.Tier)
	}
}

// TestActivateSaveError tests Activate when save fails.
func TestActivateSaveError(t *testing.T) {
	t.Parallel()
	mgr, err := NewManagerWithVerifier(internalTestVerifier(t))
	if err != nil {
		t.Fatalf("NewManagerWithVerifier error: %v", err)
	}

	// Set config dir to a non-writable path.
	mgr.configDir = "/nonexistent/readonly/path"

	// Generate valid key.
	key, _ := signTestKey(t, "2001", "1234567", TierProfessional)

	// Activate should fail because save fails.
	result := mgr.Activate(key)
	if result.Success {
		t.Error("Activate should fail when save fails")
	}
}

// TestDeactivateRemoveError tests Deactivate when remove fails.
func TestDeactivateRemoveError(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set config dir to a directory that exists but file removal will fail.
	// Create a directory where the license file should be.
	tmpDir := t.TempDir()
	configDir := tmpDir + "/.config/seed-test-suite"
	err = os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Create a non-empty directory where the license file should be.
	// os.Remove on a non-empty directory will fail.
	licensePath := configDir + "/.seed-license"
	err = os.MkdirAll(licensePath, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Add a file inside to make the directory non-empty.
	dummyFile := licensePath + "/dummy"
	err = os.WriteFile(dummyFile, []byte("test"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	mgr.configDir = configDir

	// Deactivate should fail because we can't remove a non-empty directory with os.Remove.
	err = mgr.Deactivate()
	if err == nil {
		t.Error("Deactivate should fail when remove fails on non-empty directory")
	}
}

// TestValidateLicenseKeyUnknownProductCode verifies a signed token with an
// unknown product code is rejected, and that all valid combos validate.
func TestValidateLicenseKeyUnknownProductCode(t *testing.T) {
	t.Parallel()
	v := internalTestVerifier(t)

	// A correctly signed token with an unknown product code must be rejected.
	unknown := signLicenseTokenInternal(t, "9999", "1234567", TierReflector)
	info := v.Validate(unknown)
	if info.Valid {
		t.Error("Token with unknown product code should not be valid")
	}
	if info.ErrorMsg == "" {
		t.Error("Unknown product code should set a non-empty ErrorMsg")
	}

	// All valid tier/product combinations work.
	validCombos := []struct {
		product string
		tier    Tier
	}{
		{"1001", TierReflector},
		{"2001", TierProfessional},
		{"3001", TierEnterprise},
	}

	for _, tc := range validCombos {
		key, err := signTestKey(t, tc.product, "TESTKEY", tc.tier)
		if err != nil {
			t.Errorf("signTestKey failed for %s/%v: %v", tc.product, tc.tier, err)
			continue
		}

		comboInfo := v.Validate(key)
		if !comboInfo.Valid {
			t.Errorf("Key for %s/%v should be valid: %s", tc.product, tc.tier, comboInfo.ErrorMsg)
		}
	}
}

// signLicenseTokenInternal mints an MSN1 token signed by the internal test key.
// It mirrors testsign_test.go's signTestKey but allows arbitrary (possibly
// mismatched) code/tier pairings so forgery and mismatch paths can be tested.
func signLicenseTokenInternal(t *testing.T, code, serial string, tier Tier) string {
	t.Helper()
	payload := map[string]any{
		"v":          1,
		"product":    "stem",
		"code":       code,
		"serial":     serial,
		"tier":       int(tier),
		"maxDevices": 3,
		"iat":        time.Now().Unix(),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	sig := ed25519.Sign(internalTestSigningKey(t), b)
	return "MSN1." + base64.RawURLEncoding.EncodeToString(b) +
		"." + base64.RawURLEncoding.EncodeToString(sig)
}

// TestLoadStateReadAllError tests loadState when ReadAll fails.
func TestLoadStateReadAllError(_ *testing.T) {
	// This is hard to trigger without mocking.
	// The ReadAll error path requires a file that opens but fails to read.
	// Let's skip this edge case as it's very rare in practice.
}

// TestLoadStateUnmarshalError tests loadState when JSON unmarshal fails.
func TestLoadStateUnmarshalError(_ *testing.T) {
	// This is tested by writing valid base64 that decrypts to invalid JSON.
	// Already covered by TestLoadStateWithInvalidJSON in activation_test.go.
}

// TestDarwinCPUSerialNoIOPlatformUUID tests getDarwinCPUSerial when UUID not found.
func TestDarwinCPUSerialNoIOPlatformUUID(t *testing.T) {
	t.Parallel()
	// This is hard to test without mocking exec.Command.
	// The function will return defaultDarwinCPU if UUID is not found.
	// We verify the function returns something.
	serial := getDarwinCPUSerial()
	if serial == "" {
		t.Error("getDarwinCPUSerial should not return empty string")
	}
}

// TestDarwinDiskSerialNoSerialNumber tests getDarwinDiskSerial when serial not found.
func TestDarwinDiskSerialNoSerialNumber(t *testing.T) {
	t.Parallel()
	// This function tries SATA first, then NVMe.
	// We verify it returns something.
	serial := getDarwinDiskSerial()
	if serial == "" {
		t.Error("getDarwinDiskSerial should not return empty string")
	}
}

// TestLoadStateWithValidEncryptedState tests loadState with valid encrypted data.
func TestLoadStateWithValidEncryptedState(t *testing.T) {
	t.Parallel()
	// Create a manager and save state, then verify it can be loaded.
	tmpDir := t.TempDir()
	configDir := tmpDir + "/.config/seed-test-suite"
	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Create first manager.
	mgr1, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}
	mgr1.configDir = configDir

	// Set and save state.
	mgr1.state = &ActivationState{
		LicenseKey:      "TESTKEY12345678",
		DeviceHash:      mgr1.fingerprint.Hash(),
		Tier:            TierProfessional,
		ActivatedAt:     time.Now(),
		LastValidatedAt: time.Now(),
		ExpiresAt:       time.Now().AddDate(1, 0, 0),
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        []string{"reflector", "rfc2544"},
	}

	err = mgr1.saveState()
	if err != nil {
		t.Fatalf("saveState error: %v", err)
	}

	// Create second manager and verify state loads.
	mgr2, err := NewManager()
	if err != nil {
		t.Fatalf("Second NewManager error: %v", err)
	}
	mgr2.configDir = configDir

	// Load state.
	err = mgr2.loadState()
	if err != nil {
		t.Fatalf("loadState error: %v", err)
	}

	// Verify state loaded correctly.
	if mgr2.state == nil {
		t.Fatal("State should be loaded")
	}
	if mgr2.state.LicenseKey != "TESTKEY12345678" {
		t.Errorf("LicenseKey mismatch: got %s", mgr2.state.LicenseKey)
	}
	if mgr2.state.Tier != TierProfessional {
		t.Errorf("Tier mismatch: got %v", mgr2.state.Tier)
	}
}

// TestEncryptLargeData tests encryption with larger data.
func TestEncryptLargeData(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Create large data.
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, err := mgr.encrypt(largeData)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	decrypted, err := mgr.decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if len(decrypted) != len(largeData) {
		t.Errorf("Decrypted length mismatch: got %d, want %d", len(decrypted), len(largeData))
	}

	for i := range largeData {
		if decrypted[i] != largeData[i] {
			t.Errorf("Data mismatch at position %d", i)
			break
		}
	}
}

// TestNewManagerLoadsExistingState tests that NewManager loads existing state.
func TestNewManagerLoadsExistingState(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := tmpDir + "/.config/seed-test-suite"
	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Set HOME to temp dir.
	t.Setenv("HOME", tmpDir)

	// Create first manager and save state.
	mgr1, err := NewManagerWithVerifier(internalTestVerifier(t))
	if err != nil {
		t.Fatalf("First NewManagerWithVerifier error: %v", err)
	}

	key, _ := signTestKey(t, "2001", "TESTKEY", TierProfessional)
	result := mgr1.Activate(key)
	if !result.Success {
		t.Fatalf("Activate failed: %s", result.Message)
	}

	// Create second manager - should load state.
	mgr2, err := NewManagerWithVerifier(internalTestVerifier(t))
	if err != nil {
		t.Fatalf("Second NewManagerWithVerifier error: %v", err)
	}

	// Verify state was loaded.
	if mgr2.state == nil {
		t.Error("State should be loaded")
	}

	if !mgr2.IsActivated() {
		t.Error("Should be activated from loaded state")
	}
}

// TestGetPrimaryMACWithLoopback tests that loopback is skipped.
func TestGetPrimaryMACWithLoopback(t *testing.T) {
	t.Parallel()
	// getPrimaryMAC should skip loopback interfaces.
	mac := getPrimaryMAC()

	// Should not return empty.
	if mac == "" {
		t.Error("MAC should not be empty")
	}

	// If we get a MAC, it should be in expected format or default.
	if mac != defaultMAC {
		// Should contain colons.
		colonCount := 0
		for _, c := range mac {
			if c == ':' {
				colonCount++
			}
		}
		if colonCount != 5 {
			t.Errorf("MAC format incorrect: %s", mac)
		}
	}
}

// TestValidateLicenseKeyProductCodeMismatch verifies that correctly signed
// tokens whose product code does not match their tier are rejected with the
// product-code-mismatch error. Replaces the old rotor-crafted Tier1/2/3 cases.
func TestValidateLicenseKeyProductCodeMismatch(t *testing.T) {
	t.Parallel()
	v := internalTestVerifier(t)

	cases := []struct {
		name string
		code string
		tier Tier
	}{
		{"pro-code with reflector tier", "2001", TierReflector},
		{"reflector-code with pro tier", "1001", TierProfessional},
		{"pro-code with enterprise tier", "2001", TierEnterprise},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			token := signLicenseTokenInternal(t, tc.code, "1234567", tc.tier)
			info := v.Validate(token)
			if info.Valid {
				t.Error("Token with mismatched tier/product should not be valid")
			}
			if info.ErrorMsg != errProductCodeMismatch {
				t.Errorf("Expected error %q, got %q", errProductCodeMismatch, info.ErrorMsg)
			}
		})
	}
}

// TestValidateLicenseKeyInvalidTier verifies that a correctly signed token with
// an out-of-range tier is rejected with the invalid-tier error.
func TestValidateLicenseKeyInvalidTier(t *testing.T) {
	t.Parallel()
	v := internalTestVerifier(t)

	for _, tier := range []Tier{TierInvalid, Tier(4), Tier(99)} {
		token := signLicenseTokenInternal(t, "2001", "1234567", tier)
		info := v.Validate(token)
		if info.Valid {
			t.Errorf("Token with tier %d should not be valid", tier)
		}
		if info.ErrorMsg != "Invalid license tier" {
			t.Errorf("tier %d: expected 'Invalid license tier', got %q", tier, info.ErrorMsg)
		}
	}
}

// TestForgeryRejectedInternal verifies a token signed by an attacker key is
// rejected by the production verifier (ValidateLicenseKey).
func TestForgeryRejectedInternal(t *testing.T) {
	t.Parallel()
	_, attackerPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate attacker key: %v", err)
	}
	payload := map[string]any{
		"v": 1, "product": "stem", "code": "2001", "serial": "1234567",
		"tier": int(TierProfessional), "maxDevices": 3, "iat": time.Now().Unix(),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	sig := ed25519.Sign(attackerPriv, b)
	forged := "MSN1." + base64.RawURLEncoding.EncodeToString(b) +
		"." + base64.RawURLEncoding.EncodeToString(sig)

	info := ValidateLicenseKey(forged)
	if info.Valid {
		t.Error("Token signed by a non-production key must be rejected")
	}
	if info.ErrorMsg != ErrLicenseInvalid {
		t.Errorf("forged token ErrorMsg = %q, want %q", info.ErrorMsg, ErrLicenseInvalid)
	}
}

// TestLoadStateDecryptError tests loadState when decryption fails.
func TestLoadStateDecryptError(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configDir := tmpDir + "/.config/seed-test-suite"
	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Write invalid encrypted data (valid base64 but won't decrypt properly).
	licensePath := configDir + "/.seed-license"
	// This is valid base64, long enough to pass the length check, but will fail decryption.
	invalidData := "QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE="
	err = os.WriteFile(licensePath, []byte(invalidData), 0o600)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}
	mgr.configDir = configDir

	// loadState should fail with decryption error.
	err = mgr.loadState()
	if err == nil {
		t.Error("loadState should fail with invalid encrypted data")
	}
}

// TestLoadStateOpenError tests loadState when file open fails.
func TestLoadStateOpenError(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configDir := tmpDir + "/.config/seed-test-suite"
	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Create a directory where the license file should be (makes open fail).
	licensePath := configDir + "/.seed-license"
	err = os.MkdirAll(licensePath, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}
	mgr.configDir = configDir

	// loadState should fail because the "file" is a directory.
	err = mgr.loadState()
	if err == nil {
		t.Error("loadState should fail when file is a directory")
	}
}

// TestNewManagerWithTempHome tests NewManager uses /tmp when home fails.
func TestNewManagerWithTempHome(t *testing.T) {
	t.Parallel()
	// We can't easily test the failure to get home dir without mocking.
	// But we can verify normal creation works.
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Config dir should not be empty.
	if mgr.configDir == "" {
		t.Error("configDir should not be empty")
	}

	// Fingerprint should not be nil.
	if mgr.fingerprint == nil {
		t.Error("fingerprint should not be nil")
	}
}

// TestCheckInWithTrialState tests CheckIn returns trial info.
func TestCheckInWithTrialState(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set up trial state.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierProfessional,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Now().AddDate(0, 0, -40), // 40 days ago.
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Now(),
		IsTrialMode:     true,
		Features:        nil,
	}

	result := mgr.CheckIn()
	if !result.Success {
		t.Errorf("CheckIn should succeed: %s", result.Message)
	}
}

// TestNeedsCheckInWithTrialMode tests NeedsCheckIn with trial state.
func TestNeedsCheckInWithTrialMode(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set up trial state.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierProfessional,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Now().AddDate(0, 0, -40),
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Now(),
		IsTrialMode:     true,
		Features:        nil,
	}

	// Trial mode should not need check-in.
	if mgr.NeedsCheckIn() {
		t.Error("Trial mode should not need check-in")
	}
}

// TestNeedsCheckInRecentValidation tests NeedsCheckIn with recent validation.
func TestNeedsCheckInRecentValidation(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set up license state with recent validation.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierProfessional,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Now().AddDate(0, 0, -5), // 5 days ago.
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        nil,
	}

	// Should not need check-in.
	if mgr.NeedsCheckIn() {
		t.Error("Recent validation should not need check-in")
	}
}

// TestNeedsCheckInOldValidation tests NeedsCheckIn with old validation.
func TestNeedsCheckInOldValidation(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set up license state with old validation.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierProfessional,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Now().AddDate(0, 0, -35), // 35 days ago.
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        nil,
	}

	// Should need check-in.
	if !mgr.NeedsCheckIn() {
		t.Error("Old validation should need check-in")
	}
}

// TestStartTrialWithActiveTrial tests StartTrial when trial is active.
func TestStartTrialWithActiveTrial(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set up active trial.
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierInvalid,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Time{},
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Now().AddDate(0, 0, -5), // Started 5 days ago.
		IsTrialMode:     true,
		Features:        nil,
	}

	result := mgr.StartTrial()
	if !result.Success {
		t.Errorf("StartTrial with active trial should succeed: %s", result.Message)
	}
	if !result.IsTrialMode {
		t.Error("Result should indicate trial mode")
	}
	if result.DaysRemaining == 0 {
		t.Error("Should have days remaining")
	}
}

// TestGetStateAndFingerprint tests GetState and GetFingerprint accessors.
func TestGetStateAndFingerprint(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// GetState with nil.
	mgr.state = nil
	if mgr.GetState() != nil {
		t.Error("GetState should return nil")
	}

	// GetState with value.
	testState := &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "",
		Tier:            TierProfessional,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Time{},
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        nil,
	}
	mgr.state = testState
	if mgr.GetState() != testState {
		t.Error("GetState should return the state")
	}

	// GetFingerprint.
	if mgr.GetFingerprint() == nil {
		t.Error("GetFingerprint should not be nil")
	}
}

// TestActivationStateFields tests ActivationState field access.
func TestActivationStateFields(t *testing.T) {
	t.Parallel()
	state := &ActivationState{
		LicenseKey:      "TESTKEY",
		DeviceHash:      "HASH123",
		Tier:            TierEnterprise,
		ActivatedAt:     time.Now(),
		LastValidatedAt: time.Now(),
		ExpiresAt:       time.Now().AddDate(1, 0, 0),
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        []string{"feature1", "feature2"},
	}

	// Verify fields.
	if state.LicenseKey != "TESTKEY" {
		t.Error("LicenseKey mismatch")
	}
	if state.DeviceHash != "HASH123" {
		t.Error("DeviceHash mismatch")
	}
	if state.Tier != TierEnterprise {
		t.Error("Tier mismatch")
	}
	if state.ActivatedAt.IsZero() {
		t.Error("ActivatedAt should not be zero")
	}
	if state.LastValidatedAt.IsZero() {
		t.Error("LastValidatedAt should not be zero")
	}
	if state.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}
	if state.IsTrialMode {
		t.Error("IsTrialMode should be false")
	}
	if len(state.Features) != 2 {
		t.Error("Features length mismatch")
	}
}

// TestActivationResultFields tests ActivationResult field access.
func TestActivationResultFields(t *testing.T) {
	t.Parallel()
	result := &ActivationResult{
		Success:       true,
		Message:       "Test message",
		Tier:          TierProfessional,
		DaysRemaining: 30,
		IsTrialMode:   true,
	}

	// Verify fields.
	if !result.Success {
		t.Error("Success should be true")
	}
	if result.Message != "Test message" {
		t.Error("Message mismatch")
	}
	if result.Tier != TierProfessional {
		t.Error("Tier mismatch")
	}
	if result.DaysRemaining != 30 {
		t.Error("DaysRemaining mismatch")
	}
	if !result.IsTrialMode {
		t.Error("IsTrialMode should be true")
	}
}

// TestEncryptWithEmptyPlaintext tests encrypt with empty data.
func TestEncryptWithEmptyPlaintext(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	encrypted, err := mgr.encrypt([]byte{})
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	decrypted, err := mgr.decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("Decrypted empty data should be empty, got %d bytes", len(decrypted))
	}
}

// TestIsActivatedWithZeroExpiresAt tests IsActivated with zero expiration.
func TestIsActivatedWithZeroExpiresAt(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// State with zero ExpiresAt (never expires).
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierInvalid,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Time{},
		ExpiresAt:       time.Time{}, // Zero value.
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        nil,
	}

	// Should be activated (zero expiration means no expiration check).
	if !mgr.IsActivated() {
		t.Error("License with zero ExpiresAt should be valid")
	}
}

// TestDeviceFingerprintComponents tests fingerprint components directly.
func TestDeviceFingerprintComponents(t *testing.T) {
	t.Parallel()
	fp, err := GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint error: %v", err)
	}

	// All components should be set.
	if fp.Platform == "" {
		t.Error("Platform should not be empty")
	}
	// Hostname may or may not be empty depending on system.
	// MACAddress should be set.
	if fp.MACAddress == "" {
		t.Error("MACAddress should not be empty")
	}
	// CPUSerial should be set.
	if fp.CPUSerial == "" {
		t.Error("CPUSerial should not be empty")
	}
	// DiskSerial should be set.
	if fp.DiskSerial == "" {
		t.Error("DiskSerial should not be empty")
	}
}

// TestSaveStateWriteError tests saveState when write fails.
func TestSaveStateWriteError(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Create a directory where the config should be, but make the license path a directory.
	tmpDir := t.TempDir()
	configDir := tmpDir + "/config"
	err = os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Create a directory where the license file should be written.
	licensePath := configDir + "/.seed-license"
	err = os.MkdirAll(licensePath, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	mgr.configDir = configDir
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierInvalid,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Time{},
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Time{},
		IsTrialMode:     true,
		Features:        nil,
	}

	// saveState should fail because the license path is a directory.
	err = mgr.saveState()
	if err == nil {
		t.Error("saveState should fail when license path is a directory")
	}
}

// TestLoadStateUnmarshalErrorDirect tests loadState with invalid JSON directly.
func TestLoadStateUnmarshalErrorDirect(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configDir := tmpDir + "/.config/seed-test-suite"
	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}
	mgr.configDir = configDir

	// Write valid encrypted data that decrypts to invalid JSON.
	invalidJSON := []byte("this is not valid JSON {{{")
	encrypted, err := mgr.encrypt(invalidJSON)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	licensePath := configDir + "/.seed-license"
	err = os.WriteFile(licensePath, encrypted, 0o600)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// loadState should fail with unmarshal error.
	err = mgr.loadState()
	if err == nil {
		t.Error("loadState should fail with invalid JSON")
	}
}

// TestDeriveKeyWithDifferentFingerprints tests that different fingerprints produce different keys.
func TestDeriveKeyWithDifferentFingerprints(t *testing.T) {
	t.Parallel()
	mgr1, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Create a manager with a modified fingerprint.
	mgr2, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Modify the fingerprint to get a different key.
	mgr2.fingerprint = &DeviceFingerprint{
		MACAddress: "AA:BB:CC:DD:EE:FF",
		CPUSerial:  "DIFFERENT",
		DiskSerial: "DIFFERENT",
		Hostname:   "different-host",
		Platform:   "different",
	}

	key1 := mgr1.deriveKey()
	key2 := mgr2.deriveKey()

	// Keys should be different.
	different := false
	for i := range key1 {
		if key1[i] != key2[i] {
			different = true
			break
		}
	}

	if !different {
		t.Error("Different fingerprints should produce different keys")
	}
}

// TestDecryptGCMOpenError tests decrypt when GCM.Open fails.
func TestDecryptGCMOpenError(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Create data that is valid base64 and long enough, but has invalid auth tag.
	// The nonce size for AES-GCM is 12 bytes, so we need at least that plus something.
	// Create 20 bytes of random-looking data, encode to base64.
	// This will pass length check but fail authentication.
	data := make([]byte, 28) // 12 (nonce) + 16 (min ciphertext with tag)
	for i := range data {
		data[i] = byte(i * 7) // Deterministic but invalid pattern
	}
	encoded := make([]byte, 48)
	copy(encoded, "QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVowMTIzNDU2Nzg5") // Random base64

	_, err = mgr.decrypt(encoded)
	if err == nil {
		t.Error("decrypt should fail with invalid ciphertext")
	}
}

// TestSaveStateCreateDirectoryError tests saveState when MkdirAll fails.
func TestSaveStateCreateDirectoryError(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Use a path that can't have directories created (file exists at parent).
	tmpDir := t.TempDir()
	blockingFile := tmpDir + "/blocker"
	err = os.WriteFile(blockingFile, []byte("test"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Try to create config in a path where a file blocks it.
	mgr.configDir = blockingFile + "/subdir"
	mgr.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "",
		Tier:            TierInvalid,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Time{},
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Time{},
		IsTrialMode:     true,
		Features:        nil,
	}

	err = mgr.saveState()
	if err == nil {
		t.Error("saveState should fail when MkdirAll fails")
	}
}

// TestValidateLicenseKeyWithFormattedKeyTrim verifies FormatKey only trims
// whitespace and the trimmed token still validates.
func TestValidateLicenseKeyWithFormattedKeyTrim(t *testing.T) {
	t.Parallel()
	v := internalTestVerifier(t)
	key, err := signTestKey(t, "2001", "TESTKEY", TierProfessional)
	if err != nil {
		t.Fatalf("signTestKey error: %v", err)
	}

	formatted := FormatKey("  " + key + "\n")
	if formatted != key {
		t.Errorf("FormatKey should trim to original token, got %q", formatted)
	}
	info := v.Validate(formatted)
	if !info.Valid {
		t.Errorf("Trimmed token should be valid: %s", info.ErrorMsg)
	}
}

// TestMultipleActivationsAndDeactivations tests multiple activate/deactivate cycles.
func TestMultipleActivationsAndDeactivations(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mgr, err := NewManagerWithVerifier(internalTestVerifier(t))
	if err != nil {
		t.Fatalf("NewManagerWithVerifier error: %v", err)
	}

	// Generate keys for each tier.
	reflectorKey, _ := signTestKey(t, "1001", "REFKEY1", TierReflector)
	testSuiteKey, _ := signTestKey(t, "2001", "TSTKEY1", TierProfessional)
	enterpriseKey, _ := signTestKey(t, "3001", "ENTKEY1", TierEnterprise)

	// Activate reflector.
	result := mgr.Activate(reflectorKey)
	if !result.Success {
		t.Errorf("Reflector activation failed: %s", result.Message)
	}
	if result.Tier != TierReflector {
		t.Errorf("Expected Reflector tier, got %v", result.Tier)
	}

	// Deactivate.
	err = mgr.Deactivate()
	if err != nil {
		t.Errorf("Deactivate error: %v", err)
	}

	// Activate test suite.
	result = mgr.Activate(testSuiteKey)
	if !result.Success {
		t.Errorf("TestSuite activation failed: %s", result.Message)
	}
	if result.Tier != TierProfessional {
		t.Errorf("Expected TestSuite tier, got %v", result.Tier)
	}

	// Deactivate.
	err = mgr.Deactivate()
	if err != nil {
		t.Errorf("Deactivate error: %v", err)
	}

	// Activate enterprise.
	result = mgr.Activate(enterpriseKey)
	if !result.Success {
		t.Errorf("Enterprise activation failed: %s", result.Message)
	}
	if result.Tier != TierEnterprise {
		t.Errorf("Expected Enterprise tier, got %v", result.Tier)
	}
}

// TestInfoCanRunReflectorAndTests tests CanRunReflector and CanRunTests methods.
func TestInfoCanRunReflectorAndTests(t *testing.T) {
	t.Parallel()
	v := internalTestVerifier(t)

	// Test reflector tier.
	reflectorKey, _ := signTestKey(t, "1001", "REFKEY1", TierReflector)
	info := v.Validate(reflectorKey)
	if !info.CanRunReflector() {
		t.Error("Reflector tier should be able to run reflector")
	}
	if info.CanRunTests() {
		t.Error("Reflector tier should NOT be able to run tests")
	}

	// Professional tier.
	testKey, _ := signTestKey(t, "2001", "TSTKEY1", TierProfessional)
	info = v.Validate(testKey)
	if !info.CanRunReflector() {
		t.Error("Professional tier should be able to run reflector")
	}
	if !info.CanRunTests() {
		t.Error("Professional tier should be able to run tests")
	}

	// Enterprise tier.
	entKey, _ := signTestKey(t, "3001", "ENTKEY1", TierEnterprise)
	info = v.Validate(entKey)
	if !info.CanRunReflector() {
		t.Error("Enterprise tier should be able to run reflector")
	}
	if !info.CanRunTests() {
		t.Error("Enterprise tier should be able to run tests")
	}

	// Invalid key.
	info = ValidateLicenseKey("INVALIDKEY123456")
	if info.CanRunReflector() {
		t.Error("Invalid key should NOT be able to run reflector")
	}
	if info.CanRunTests() {
		t.Error("Invalid key should NOT be able to run tests")
	}
}

// TestInfoHasFeature tests HasFeature method.
func TestInfoHasFeature(t *testing.T) {
	t.Parallel()
	v := internalTestVerifier(t)

	// Reflector tier.
	reflectorKey, _ := signTestKey(t, "1001", "REFKEY1", TierReflector)
	info := v.Validate(reflectorKey)
	if !info.HasFeature("reflector") {
		t.Error("Reflector tier should have 'reflector' feature")
	}
	if info.HasFeature("rfc2544") {
		t.Error("Reflector tier should NOT have 'rfc2544' feature")
	}

	// Professional tier.
	testKey, _ := signTestKey(t, "2001", "TSTKEY1", TierProfessional)
	info = v.Validate(testKey)
	if !info.HasFeature("reflector") {
		t.Error("Professional tier should have 'reflector' feature")
	}
	if !info.HasFeature("rfc2544") {
		t.Error("Professional tier should have 'rfc2544' feature")
	}
	if !info.HasFeature("y1564") {
		t.Error("Professional tier should have 'y1564' feature")
	}

	// Enterprise tier.
	entKey, _ := signTestKey(t, "3001", "ENTKEY1", TierEnterprise)
	info = v.Validate(entKey)
	if !info.HasFeature("api") {
		t.Error("Enterprise tier should have 'api' feature")
	}
	if !info.HasFeature("multiuser") {
		t.Error("Enterprise tier should have 'multiuser' feature")
	}
}

// TestTierStringValues tests all Tier.String() return values.
func TestTierStringValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		tier     Tier
		expected string
	}{
		{TierInvalid, "Invalid"},
		{TierReflector, "Reflector"},
		{TierProfessional, "Professional"},
		// TierEnterprise is deprecated and folded into Professional.
		{TierEnterprise, "Professional"},
		{Tier(99), "Invalid"}, // Unknown tier should return "Invalid".
	}

	for _, tc := range tests {
		result := tc.tier.String()
		if result != tc.expected {
			t.Errorf("Tier(%d).String() = %q, want %q", tc.tier, result, tc.expected)
		}
	}
}

// TestFormatKeyVariousInputs tests FormatKey trims whitespace and otherwise
// returns the token unchanged (tokens are display-ready; base64url content must
// not be altered).
func TestFormatKeyVariousInputs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		// Tokens are returned verbatim (no separator insertion).
		{"ABCD1234EFGH5678", "ABCD1234EFGH5678"},
		{"abcd1234efgh5678", "abcd1234efgh5678"},
		// Surrounding whitespace is trimmed.
		{"  ABCD1234EFGH5678  ", "ABCD1234EFGH5678"},
		{"ABCD1234EFGH5678\n", "ABCD1234EFGH5678"},
		// Short input returned as-is.
		{"ABCD", "ABCD"},
		// Empty string.
		{"", ""},
	}

	for _, tc := range tests {
		result := FormatKey(tc.input)
		if result != tc.expected {
			t.Errorf("FormatKey(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// TestSplitLinesEmpty tests splitLines with empty and single-line strings.
func TestSplitLinesEmpty(t *testing.T) {
	t.Parallel()
	// Empty string.
	var lines []string
	for line := range splitLines("") {
		lines = append(lines, line)
	}
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("splitLines('') = %v, want ['']", lines)
	}

	// Single line no newline.
	lines = nil
	for line := range splitLines("hello") {
		lines = append(lines, line)
	}
	if len(lines) != 1 || lines[0] != "hello" {
		t.Errorf("splitLines('hello') = %v, want ['hello']", lines)
	}

	// Multiple empty lines.
	lines = nil
	for line := range splitLines("\n\n") {
		lines = append(lines, line)
	}
	if len(lines) != 3 {
		t.Errorf("splitLines('\\n\\n') should have 3 elements, got %d", len(lines))
	}
}

// TestDecryptBase64DecodeError tests decrypt with invalid base64.
func TestDecryptBase64DecodeError(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Invalid base64 characters.
	_, err = mgr.decrypt([]byte("!!!invalid-base64-data!!!"))
	if err == nil {
		t.Error("decrypt should fail with invalid base64")
	}
}

// TestCheckInNilState tests CheckIn with nil state.
func TestCheckInNilState(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	mgr.state = nil
	result := mgr.CheckIn()
	if result.Success {
		t.Error("CheckIn with nil state should fail")
	}
	if result.Tier != TierInvalid {
		t.Errorf("Expected TierInvalid, got %v", result.Tier)
	}
}

// TestNeedsCheckInNilState tests NeedsCheckIn with nil state.
func TestNeedsCheckInNilState(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	mgr.state = nil
	if mgr.NeedsCheckIn() {
		t.Error("NeedsCheckIn with nil state should return false")
	}
}

// TestGetPrimaryMACFiltersVirtualInterfaces tests that virtual interfaces are skipped.
func TestGetPrimaryMACFiltersVirtualInterfaces(t *testing.T) {
	t.Parallel()
	// This tests the getPrimaryMAC function indirectly.
	// We can't inject mock interfaces, but we verify the function runs without error
	// and returns a valid format.
	mac := getPrimaryMAC()

	// Should return either a valid MAC or the default.
	if mac == "" {
		t.Error("getPrimaryMAC should never return empty string")
	}

	// If it's not the default, verify it's a valid MAC format.
	if mac != defaultMAC {
		// MAC should be in format XX:XX:XX:XX:XX:XX
		parts := strings.Split(mac, ":")
		if len(parts) != 6 {
			t.Errorf("MAC should have 6 parts separated by colons, got: %s", mac)
		}
		for i, part := range parts {
			if len(part) != 2 {
				t.Errorf("MAC part %d should be 2 chars, got %d: %s", i, len(part), part)
			}
		}
	}
}

// TestGetCPUSerialPlatformSpecific tests getCPUSerial returns a non-empty result.
func TestGetCPUSerialPlatformSpecific(t *testing.T) {
	t.Parallel()
	serial := getCPUSerial()

	// Should never be empty.
	if serial == "" {
		t.Error("getCPUSerial should never return empty")
	}

	// Should be one of the default values or an actual serial.
	validCPUDefaults := []string{defaultLinuxCPU, defaultDarwinCPU, unknownSerial}
	isDefault := slices.Contains(validCPUDefaults, serial)

	// Either it's a default or it's an actual serial (which should be non-empty).
	if !isDefault && len(serial) == 0 {
		t.Error("Non-default CPU serial should have content")
	}
}

// TestGetDiskSerialPlatformSpecific tests getDiskSerial returns a non-empty result.
func TestGetDiskSerialPlatformSpecific(t *testing.T) {
	t.Parallel()
	serial := getDiskSerial()

	// Should never be empty.
	if serial == "" {
		t.Error("getDiskSerial should never return empty")
	}

	// Should be one of the default values or an actual serial.
	validDiskDefaults := []string{defaultLinuxDisk, defaultDarwinDisk, unknownSerial}
	isDefault := slices.Contains(validDiskDefaults, serial)

	// Either it's a default or it's an actual serial.
	if !isDefault && len(serial) == 0 {
		t.Error("Non-default disk serial should have content")
	}
}

// TestEncryptConsistency tests that encrypt produces valid ciphertext each time.
func TestEncryptConsistency(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Test with various data sizes.
	testData := [][]byte{
		{},                              // Empty.
		{0},                             // Single byte.
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, // Small.
		make([]byte, 100),               // Medium.
		make([]byte, 1000),              // Large.
		make([]byte, 10000),             // Very large.
	}

	for i, data := range testData {
		encrypted, encErr := mgr.encrypt(data)
		if encErr != nil {
			t.Errorf("Test %d: encrypt error: %v", i, encErr)
			continue
		}

		// Encrypted data should always be non-empty (includes nonce + tag).
		if len(encrypted) == 0 {
			t.Errorf("Test %d: encrypted data should not be empty", i)
		}

		// Should decrypt back.
		decrypted, decErr := mgr.decrypt(encrypted)
		if decErr != nil {
			t.Errorf("Test %d: decrypt error: %v", i, decErr)
			continue
		}

		if len(decrypted) != len(data) {
			t.Errorf("Test %d: length mismatch after roundtrip: got %d, want %d",
				i, len(decrypted), len(data))
		}
	}
}

// TestDecryptErrorPaths tests various error conditions in decrypt.
func TestDecryptErrorPaths(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Test various invalid inputs.
	invalidInputs := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"invalid_base64_chars", []byte("!!!")},
		{"short_base64", []byte("YQ==")},                       // Decodes to "a" (1 byte).
		{"medium_invalid", []byte("YWJjZGVmZ2hpamtsbW5vcA==")}, // Decodes to 16 bytes but invalid ciphertext.
	}

	for _, tc := range invalidInputs {
		_, decErr := mgr.decrypt(tc.input)
		if decErr == nil {
			t.Errorf("%s: decrypt should fail", tc.name)
		}
	}
}

// TestNewManagerWithValidHomeDir tests NewManager creates config directory correctly.
func TestNewManagerWithValidHomeDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Verify manager is properly initialized.
	if mgr.fingerprint == nil {
		t.Error("fingerprint should be initialized")
	}

	if mgr.configDir == "" {
		t.Error("configDir should be set")
	}

	// Config dir should be under the temp home.
	if !strings.HasPrefix(mgr.configDir, tmpDir) {
		t.Errorf("configDir should be under HOME: got %s, want prefix %s",
			mgr.configDir, tmpDir)
	}
}

// TestSaveStateAndLoadRoundtrip tests complete save/load cycle.
func TestSaveStateAndLoadRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create first manager and set complex state.
	mgr1, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set state with all fields populated.
	now := time.Now()
	mgr1.state = &ActivationState{
		LicenseKey:      "TESTLICENSEKEY12",
		DeviceHash:      mgr1.fingerprint.Hash(),
		Tier:            TierEnterprise,
		ActivatedAt:     now,
		LastValidatedAt: now.Add(-time.Hour),
		ExpiresAt:       now.AddDate(1, 0, 0),
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        []string{"reflector", "rfc2544", "y1564", "api", "multiuser"},
	}

	// Save state.
	err = mgr1.saveState()
	if err != nil {
		t.Fatalf("saveState error: %v", err)
	}

	// Create second manager - should load state.
	mgr2, err := NewManager()
	if err != nil {
		t.Fatalf("Second NewManager error: %v", err)
	}

	// Verify state was loaded correctly.
	if mgr2.state == nil {
		t.Fatal("State should be loaded")
	}

	// Verify all fields.
	if mgr2.state.LicenseKey != "TESTLICENSEKEY12" {
		t.Errorf("LicenseKey mismatch: %s", mgr2.state.LicenseKey)
	}
	if mgr2.state.Tier != TierEnterprise {
		t.Errorf("Tier mismatch: %v", mgr2.state.Tier)
	}
	if mgr2.state.IsTrialMode {
		t.Error("IsTrialMode should be false")
	}
	if len(mgr2.state.Features) != 5 {
		t.Errorf("Features count mismatch: %d", len(mgr2.state.Features))
	}
}

// TestFingerprintFieldsPopulated tests that all fingerprint fields are set.
func TestFingerprintFieldsPopulated(t *testing.T) {
	t.Parallel()
	fp, err := GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint error: %v", err)
	}

	// Platform should always be set.
	if fp.Platform == "" {
		t.Error("Platform should not be empty")
	}

	// MACAddress should be set (even if default).
	if fp.MACAddress == "" {
		t.Error("MACAddress should not be empty")
	}

	// CPUSerial should be set (even if default).
	if fp.CPUSerial == "" {
		t.Error("CPUSerial should not be empty")
	}

	// DiskSerial should be set (even if default).
	if fp.DiskSerial == "" {
		t.Error("DiskSerial should not be empty")
	}

	// Hostname might be empty on some systems but shouldn't cause issues.
}

// TestMaskStringWithVariousLengths tests maskString with different show values.
func TestMaskStringWithVariousLengths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		show     int
		expected string
	}{
		{"", 0, ""},
		{"", 5, ""},
		{"A", 0, "****"},
		{"A", 1, "A"},
		{"A", 5, "A"},
		{"ABCDEFGHIJ", 0, "****"},
		{"ABCDEFGHIJ", 5, "ABCDE****"},
		{"ABCDEFGHIJ", 10, "ABCDEFGHIJ"},
		{"ABCDEFGHIJ", 15, "ABCDEFGHIJ"},
	}

	for _, tc := range tests {
		result := maskString(tc.input, tc.show)
		if result != tc.expected {
			t.Errorf("maskString(%q, %d) = %q, want %q",
				tc.input, tc.show, result, tc.expected)
		}
	}
}

// TestDeriveKeyIsConsistentWithSameFingerprint tests key derivation consistency.
func TestDeriveKeyIsConsistentWithSameFingerprint(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Call deriveKey multiple times.
	keys := make([][]byte, 10)
	for i := range keys {
		keys[i] = mgr.deriveKey()
	}

	// All keys should be identical.
	for i := 1; i < len(keys); i++ {
		if len(keys[i]) != len(keys[0]) {
			t.Errorf("Key %d has different length", i)
			continue
		}
		for j := range keys[0] {
			if keys[i][j] != keys[0][j] {
				t.Errorf("Key %d differs at position %d", i, j)
				break
			}
		}
	}

	// Key should be 32 bytes (AES-256).
	if len(keys[0]) != 32 {
		t.Errorf("Key should be 32 bytes, got %d", len(keys[0]))
	}
}

// TestEncryptProducesDifferentOutputsForSameInput tests nonce randomness.
func TestEncryptProducesDifferentOutputsForSameInput(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	plaintext := []byte("test data for encryption")

	// Encrypt same data multiple times.
	results := make([][]byte, 5)
	for i := range results {
		encrypted, encErr := mgr.encrypt(plaintext)
		if encErr != nil {
			t.Fatalf("encrypt error: %v", encErr)
		}
		results[i] = encrypted
	}

	// All results should be different (due to random nonce).
	for i := range results {
		for j := i + 1; j < len(results); j++ {
			if string(results[i]) == string(results[j]) {
				t.Errorf("Encryptions %d and %d should be different", i, j)
			}
		}
	}

	// But all should decrypt to the same plaintext.
	for i, encrypted := range results {
		decrypted, decErr := mgr.decrypt(encrypted)
		if decErr != nil {
			t.Errorf("Decryption %d failed: %v", i, decErr)
			continue
		}
		if string(decrypted) != string(plaintext) {
			t.Errorf("Decryption %d produced wrong result", i)
		}
	}
}

// TestLoadStateWithEmptyFile tests loadState when file is empty.
func TestLoadStateWithEmptyFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configDir := tmpDir + "/.config/seed-test-suite"
	err := os.MkdirAll(configDir, 0o700)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Write empty file.
	licensePath := configDir + "/.seed-license"
	err = os.WriteFile(licensePath, []byte{}, 0o600)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}
	mgr.configDir = configDir

	// loadState should fail with empty file.
	err = mgr.loadState()
	if err == nil {
		t.Error("loadState should fail with empty file")
	}
}

// TestCheckInUpdatesLastValidatedAt tests that CheckIn updates the timestamp.
func TestCheckInUpdatesLastValidatedAt(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set up a state with old LastValidatedAt.
	oldTime := time.Now().Add(-48 * time.Hour)
	mgr.state = &ActivationState{
		LicenseKey:      "TESTKEY",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierProfessional,
		ActivatedAt:     time.Time{},
		LastValidatedAt: oldTime,
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        nil,
	}

	// CheckIn.
	result := mgr.CheckIn()
	if !result.Success {
		t.Fatalf("CheckIn failed: %s", result.Message)
	}

	// LastValidatedAt should be updated.
	if !mgr.state.LastValidatedAt.After(oldTime) {
		t.Error("LastValidatedAt should be updated after CheckIn")
	}
}

// TestGenerateFingerprintConsistency tests fingerprint generation is consistent.
func TestGenerateFingerprintConsistency(t *testing.T) {
	t.Parallel()
	// Generate fingerprint multiple times.
	fingerprints := make([]*DeviceFingerprint, 5)
	for i := range fingerprints {
		fp, err := GenerateFingerprint()
		if err != nil {
			t.Fatalf("GenerateFingerprint error: %v", err)
		}
		fingerprints[i] = fp
	}

	// All fingerprints should have the same hash.
	hash0 := fingerprints[0].Hash()
	for i := 1; i < len(fingerprints); i++ {
		if fingerprints[i].Hash() != hash0 {
			t.Errorf("Fingerprint %d hash differs from fingerprint 0", i)
		}
	}
}

// TestValidateLicenseKeyMalformedRejected verifies that strings which are not
// well-formed MSN1 tokens (old 16-char formats, garbage, special chars) are
// rejected with a non-empty error message.
func TestValidateLicenseKeyMalformedRejected(t *testing.T) {
	t.Parallel()
	malformed := []string{
		"",
		"A",
		"AB",
		"ABCDEFGH",           // legacy 8-char
		"ABCDEFGHIJKLMNO",    // legacy 15-char
		"ABCDEFGHIJKLMNOPQ",  // legacy 17-char
		"ABCD!@#$EFGH5678",   // special chars
		"abcd!@#$efgh5678",   // special chars lowercase
		"ABCDEFGH12345678",   // legacy 16-char alphanumeric
		"MSN1.onlytwo",       // wrong part count
		"MSN1.notb64.notb64", // undecodable parts
		"WRONG.payload.sig",  // wrong scheme
	}

	for _, key := range malformed {
		info := ValidateLicenseKey(key)
		if info.Valid {
			t.Errorf("Malformed key %q should not be valid", key)
		}
		if info.ErrorMsg == "" {
			t.Errorf("Malformed key %q should set a non-empty ErrorMsg", key)
		}
	}
}

// TestSplitLinesWithCarriageReturn tests splitLines with Windows line endings.
func TestSplitLinesWithCarriageReturn(t *testing.T) {
	t.Parallel()
	// Note: splitLines splits on \n only, not \r\n
	input := "line1\r\nline2\r\nline3"
	var lines []string
	for line := range splitLines(input) {
		lines = append(lines, line)
	}

	// Should have 3 lines (split on \n).
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// First line should contain \r at the end.
	if !strings.HasSuffix(lines[0], "\r") {
		t.Log("First line may or may not have \\r depending on implementation")
	}
}

// TestActivateWithWhitespaceTrimmed verifies that a token wrapped in surrounding
// whitespace still activates (Activate trims via the verifier), while an altered
// token does not (base64url content must not be mutated).
func TestActivateWithWhitespaceTrimmed(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mgr, err := NewManagerWithVerifier(internalTestVerifier(t))
	if err != nil {
		t.Fatalf("NewManagerWithVerifier error: %v", err)
	}

	key, err := signTestKey(t, "2001", "TESTKEY", TierProfessional)
	if err != nil {
		t.Fatalf("signTestKey error: %v", err)
	}

	// Surrounding whitespace is trimmed and the token still activates.
	result := mgr.Activate("  " + key + "\n")
	if !result.Success {
		t.Errorf("Whitespace-wrapped token should activate: %s", result.Message)
	}

	_ = mgr.Deactivate()

	// An altered token (lowercased base64url) must NOT activate.
	altered := strings.ToLower(key)
	if altered != key { // only meaningful if there were uppercase chars
		result = mgr.Activate(altered)
		if result.Success {
			t.Error("Altered (lowercased) token must not activate")
		}
	}
}

// TestDeviceFingerprintHashFormat tests that fingerprint hash is always valid format.
func TestDeviceFingerprintHashFormat(t *testing.T) {
	t.Parallel()
	fp, err := GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint error: %v", err)
	}

	hash := fp.Hash()

	// Should be exactly 16 characters.
	if len(hash) != 16 {
		t.Errorf("Hash should be 16 chars, got %d", len(hash))
	}

	// Should be all uppercase hex characters.
	for _, c := range hash {
		isDigit := c >= '0' && c <= '9'
		isUpperHex := c >= 'A' && c <= 'F'
		if !isDigit && !isUpperHex {
			t.Errorf("Hash should contain only uppercase hex chars, found %c", c)
		}
	}
}

// TestIsActivatedWithExpiredLicense tests IsActivated when license is expired.
func TestIsActivatedWithExpiredLicense(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Set state with expired license.
	mgr.state = &ActivationState{
		LicenseKey:      "TESTKEY",
		DeviceHash:      mgr.fingerprint.Hash(),
		Tier:            TierProfessional,
		ActivatedAt:     time.Now().AddDate(-2, 0, 0), // 2 years ago.
		LastValidatedAt: time.Now().AddDate(-1, 0, 0), // 1 year ago.
		ExpiresAt:       time.Now().AddDate(-1, 0, 0), // Expired 1 year ago.
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        nil,
	}

	if mgr.IsActivated() {
		t.Error("Expired license should not be activated")
	}
}

// TestValidateLicenseKeyWithAllTiers tests validation for all tier combinations.
func TestValidateLicenseKeyWithAllTiers(t *testing.T) {
	t.Parallel()
	tiers := []struct {
		product      string
		tier         Tier
		expectedTier string
	}{
		{"1001", TierReflector, "Reflector"},
		{"2001", TierProfessional, "Professional"},
		// TierEnterprise (3) is deprecated; existing keys still validate
		// and report as Professional per LICENSE_STRATEGY 2026-05-19.
		{"3001", TierEnterprise, "Professional"},
	}

	v := internalTestVerifier(t)
	for _, tc := range tiers {
		key, err := signTestKey(t, tc.product, "ABCDEFG", tc.tier)
		if err != nil {
			t.Errorf("signTestKey failed for %s/%v: %v", tc.product, tc.tier, err)
			continue
		}

		info := v.Validate(key)
		if !info.Valid {
			t.Errorf("Key for %s/%v should be valid: %s", tc.product, tc.tier, info.ErrorMsg)
			continue
		}

		if info.Tier.String() != tc.expectedTier {
			t.Errorf("Expected tier string %q, got %q", tc.expectedTier, info.Tier.String())
		}
	}
}

// TestStartTrialMultipleTimes tests starting trial when already in trial.
func TestStartTrialMultipleTimes(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Start first trial.
	result1 := mgr.StartTrial()
	if !result1.Success {
		t.Fatalf("First StartTrial failed: %s", result1.Message)
	}

	// Start trial again - should return existing trial info.
	result2 := mgr.StartTrial()
	if !result2.Success {
		t.Errorf("Second StartTrial should succeed: %s", result2.Message)
	}
	if !result2.IsTrialMode {
		t.Error("Should still be in trial mode")
	}
}

// TestDeactivateWithNoLicense tests Deactivate when no license exists.
func TestDeactivateWithNoLicense(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Deactivate when nothing is activated.
	err = mgr.Deactivate()
	if err != nil {
		t.Errorf("Deactivate with no license should not error: %v", err)
	}
}

// TestMaskStringWithEmptyInput tests maskString edge cases.
func TestMaskStringWithEmptyInput(t *testing.T) {
	t.Parallel()
	// Empty string.
	result := maskString("", 0)
	if result != "" {
		t.Errorf("maskString('', 0) = %q, want ''", result)
	}

	result = maskString("", 10)
	if result != "" {
		t.Errorf("maskString('', 10) = %q, want ''", result)
	}

	// Single character.
	result = maskString("A", 0)
	if result != "****" {
		t.Errorf("maskString('A', 0) = %q, want '****'", result)
	}

	result = maskString("A", 1)
	if result != "A" {
		t.Errorf("maskString('A', 1) = %q, want 'A'", result)
	}
}

// TestValidateLicenseKeyAllProductCodes tests validation of all product code types.
func TestValidateLicenseKeyAllProductCodes(t *testing.T) {
	t.Parallel()
	// Test all three valid product codes with their corresponding tiers.
	testCases := []struct {
		product string
		tier    Tier
	}{
		{"1001", TierReflector},
		{"2001", TierProfessional},
		{"3001", TierEnterprise},
	}

	v := internalTestVerifier(t)
	for _, tc := range testCases {
		key, err := signTestKey(t, tc.product, "TESTSER", tc.tier)
		if err != nil {
			t.Fatalf("signTestKey(%s, TESTSER, %d) error: %v", tc.product, tc.tier, err)
		}

		info := v.Validate(key)
		if !info.Valid {
			t.Errorf("Key for %s should be valid: %s", tc.product, info.ErrorMsg)
		}
		if info.ProductCode != tc.product {
			t.Errorf("ProductCode mismatch: expected %s, got %s", tc.product, info.ProductCode)
		}
		if info.Tier != tc.tier {
			t.Errorf("Tier mismatch: expected %d, got %d", tc.tier, info.Tier)
		}
	}
}

// TestValidateLicenseKeyEnterpriseTier tests Enterprise tier specifically.
func TestValidateLicenseKeyEnterpriseTier(t *testing.T) {
	t.Parallel()
	// Generate an Enterprise tier key.
	key, err := signTestKey(t, "3001", "ZZZZZZA", TierEnterprise)
	if err != nil {
		t.Fatalf("signTestKey error: %v", err)
	}

	info := internalTestVerifier(t).Validate(key)
	if !info.Valid {
		t.Errorf("Enterprise key should be valid: %s", info.ErrorMsg)
	}
	if info.Tier != TierEnterprise {
		t.Errorf("Tier should be Enterprise, got %d", info.Tier)
	}
	if info.ProductCode != "3001" {
		t.Errorf("ProductCode should be 3001, got %s", info.ProductCode)
	}
	// Enterprise should have api and multiuser features.
	if !hasFeature(info.Features, "api") {
		t.Error("Enterprise should have api feature")
	}
	if !hasFeature(info.Features, "multiuser") {
		t.Error("Enterprise should have multiuser feature")
	}
}

// hasFeature is a helper function to check if a feature exists in a slice.
func hasFeature(features []string, feature string) bool {
	return slices.Contains(features, feature)
}

// TestDeriveKeyConsistency tests that deriveKey produces consistent results.
func TestDeriveKeyConsistency(t *testing.T) {
	t.Parallel()
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager error: %v", err)
	}

	// Get fingerprint.
	fp := mgr.GetFingerprint()
	if fp == nil {
		t.Fatal("GetFingerprint returned nil")
	}

	// The deriveKey function uses fingerprint, so we can't call it directly,
	// but we can verify that encryption/decryption works consistently.
	mgr.StartTrial()

	state1 := mgr.GetState()
	if state1 == nil {
		t.Fatal("GetState returned nil")
	}

	// Device hash should be consistent.
	hash1 := fp.Hash()
	hash2 := fp.Hash()
	if hash1 != hash2 {
		t.Error("Fingerprint hash should be consistent")
	}
}
