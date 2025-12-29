// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package license

import (
	"testing"
)

func TestRotorCipherRoundTrip(t *testing.T) {
	testCases := []struct {
		input    string
		position int
	}{
		{"ABCD1234", 0},
		{"1234ABCD", 7},
		{"MSN12345TEST", 15},
		{"0000000000000000", 0},
		{"AAAAAAAAAAAAAAAA", 0},
	}

	for _, tc := range testCases {
		cipher := NewRotorCipher(tc.position)
		encoded := cipher.EncodeString(tc.input)

		// Reset position for decoding
		cipher = NewRotorCipher(tc.position)
		decoded := cipher.DecodeString(encoded)

		if decoded != tc.input {
			t.Errorf("RoundTrip failed: input=%q, encoded=%q, decoded=%q", tc.input, encoded, decoded)
		}
	}
}

func TestCalculateChecksum(t *testing.T) {
	testCases := []struct {
		input    string
		expected int // length of checksum
	}{
		{"ABC123", 2},
		{"1001ABCDEF12", 2},
		{"", 2},
	}

	for _, tc := range testCases {
		checksum := CalculateChecksum(tc.input)
		if len(checksum) != tc.expected {
			t.Errorf("Checksum length wrong: input=%q, got=%d, want=%d", tc.input, len(checksum), tc.expected)
		}
	}
}

func TestValidateChecksum(t *testing.T) {
	// Test with valid checksums
	payload := "ABC123"
	checksum := CalculateChecksum(payload)
	valid := ValidateChecksum(payload + checksum)

	if !valid {
		t.Errorf("ValidateChecksum should return true for valid checksum")
	}

	// Test with invalid checksum
	invalid := ValidateChecksum(payload + "XX")
	if invalid {
		t.Errorf("ValidateChecksum should return false for invalid checksum")
	}
}

func TestGenerateLicenseKey(t *testing.T) {
	testCases := []struct {
		productCode string
		serial      string
		tier        Tier
		wantErr     bool
	}{
		{"1001", "ABCDEFG", TierReflector, false},
		{"2001", "1234567", TierTestSuite, false},
		{"3001", "XYZXYZX", TierEnterprise, false},
		{"100", "ABCDEFG", TierReflector, true}, // Invalid product code length
		{"1001", "ABCDEF", TierReflector, true}, // Invalid serial length
		{"1001", "ABCDEFG", Tier(0), true},      // Invalid tier
	}

	for _, tc := range testCases {
		key, err := GenerateLicenseKey(tc.productCode, tc.serial, tc.tier)
		if tc.wantErr {
			if err == nil {
				t.Errorf("GenerateLicenseKey(%q, %q, %d) should return error", tc.productCode, tc.serial, tc.tier)
			}
			continue
		}

		if err != nil {
			t.Errorf("GenerateLicenseKey(%q, %q, %d) returned unexpected error: %v", tc.productCode, tc.serial, tc.tier, err)
			continue
		}

		if len(key) != 16 {
			t.Errorf("Generated key length wrong: got=%d, want=16", len(key))
		}
	}
}

func TestValidateLicenseKey(t *testing.T) {
	// Generate a valid key and test validation
	key, err := GenerateLicenseKey("1001", "ABCDEFG", TierReflector)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	info := ValidateLicenseKey(key)
	if !info.Valid {
		t.Errorf("ValidateLicenseKey should return valid for generated key: %s, error: %s", key, info.ErrorMsg)
	}

	// Test invalid keys
	invalidKeys := []struct {
		key     string
		wantErr string
	}{
		{"", "License key must be 16 characters"},
		{"SHORT", "License key must be 16 characters"},
		{"INVALID-CHARS-@@", "License key must contain only letters and numbers"},
	}

	for _, tc := range invalidKeys {
		info := ValidateLicenseKey(tc.key)
		if info.Valid {
			t.Errorf("ValidateLicenseKey(%q) should not be valid", tc.key)
		}
	}
}

func TestTierString(t *testing.T) {
	testCases := []struct {
		tier     Tier
		expected string
	}{
		{TierReflector, "Reflector"},
		{TierTestSuite, "Test Suite"},
		{TierEnterprise, "Enterprise"},
		{TierInvalid, "Invalid"},
	}

	for _, tc := range testCases {
		if tc.tier.String() != tc.expected {
			t.Errorf("Tier.String() = %q, want %q", tc.tier.String(), tc.expected)
		}
	}
}

func TestFormatKey(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"ABCD1234EFGH5678", "ABCD-1234-EFGH-5678"},
		{"abcd1234efgh5678", "ABCD-1234-EFGH-5678"},
		{"ABCD-1234-EFGH-5678", "ABCD-1234-EFGH-5678"},
		{"SHORT", "SHORT"}, // Invalid length, return as-is
	}

	for _, tc := range testCases {
		result := FormatKey(tc.input)
		if result != tc.expected {
			t.Errorf("FormatKey(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestNormalizeKey(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"abcd-1234-efgh-5678", "ABCD1234EFGH5678"},
		{"ABCD 1234 EFGH 5678", "ABCD1234EFGH5678"},
		{"abcd.1234.efgh.5678", "ABCD1234EFGH5678"},
	}

	for _, tc := range testCases {
		result := normalizeKey(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeKey(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestLicenseInfoHasFeature(t *testing.T) {
	info := &LicenseInfo{
		Features: []string{"reflector", "rfc2544", "y1564"},
	}

	if !info.HasFeature("reflector") {
		t.Error("HasFeature should return true for existing feature")
	}

	if info.HasFeature("nonexistent") {
		t.Error("HasFeature should return false for non-existing feature")
	}
}

func TestLicenseInfoCanRunReflector(t *testing.T) {
	tests := []struct {
		info *LicenseInfo
		want bool
	}{
		{&LicenseInfo{Valid: true, Tier: TierReflector}, true},
		{&LicenseInfo{Valid: true, Tier: TierTestSuite}, true},
		{&LicenseInfo{Valid: true, Tier: TierEnterprise}, true},
		{&LicenseInfo{Valid: false, Tier: TierReflector}, false},
		{&LicenseInfo{Valid: true, Tier: TierInvalid}, false},
	}

	for i, tc := range tests {
		if tc.info.CanRunReflector() != tc.want {
			t.Errorf("Test %d: CanRunReflector() = %v, want %v", i, tc.info.CanRunReflector(), tc.want)
		}
	}
}

func TestLicenseInfoCanRunTests(t *testing.T) {
	tests := []struct {
		info *LicenseInfo
		want bool
	}{
		{&LicenseInfo{Valid: true, Tier: TierReflector}, false},
		{&LicenseInfo{Valid: true, Tier: TierTestSuite}, true},
		{&LicenseInfo{Valid: true, Tier: TierEnterprise}, true},
		{&LicenseInfo{Valid: false, Tier: TierTestSuite}, false},
	}

	for i, tc := range tests {
		if tc.info.CanRunTests() != tc.want {
			t.Errorf("Test %d: CanRunTests() = %v, want %v", i, tc.info.CanRunTests(), tc.want)
		}
	}
}

func TestDeviceFingerprint(t *testing.T) {
	fp, err := GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint failed: %v", err)
	}

	// Verify hash is 16 characters
	hash := fp.Hash()
	if len(hash) != 16 {
		t.Errorf("Fingerprint hash length = %d, want 16", len(hash))
	}

	// Verify hash is consistent
	hash2 := fp.Hash()
	if hash != hash2 {
		t.Error("Fingerprint hash should be consistent")
	}

	// Verify String() doesn't panic
	str := fp.String()
	if str == "" {
		t.Error("Fingerprint String() should not be empty")
	}
}
