// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package license_test

import (
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/license"
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
		cipher := license.NewRotorCipher(tc.position)
		encoded := cipher.EncodeString(tc.input)

		// Reset position for decoding.
		cipher = license.NewRotorCipher(tc.position)
		decoded := cipher.DecodeString(encoded)

		if decoded != tc.input {
			t.Errorf("RoundTrip failed: input=%q, encoded=%q, decoded=%q", tc.input, encoded, decoded)
		}
	}
}

func TestCalculateChecksum(t *testing.T) {
	testCases := []struct {
		input    string
		expected int // length of checksum.
	}{
		{"ABC123", 2},
		{"1001ABCDEF12", 2},
		{"", 2},
	}

	for _, tc := range testCases {
		checksum := license.CalculateChecksum(tc.input)
		if len(checksum) != tc.expected {
			t.Errorf("Checksum length wrong: input=%q, got=%d, want=%d", tc.input, len(checksum), tc.expected)
		}
	}
}

func TestValidateChecksum(t *testing.T) {
	// Test with valid checksums.
	payload := "ABC123"
	checksum := license.CalculateChecksum(payload)
	valid := license.ValidateChecksum(payload + checksum)

	if !valid {
		t.Errorf("ValidateChecksum should return true for valid checksum")
	}

	// Test with invalid checksum.
	invalid := license.ValidateChecksum(payload + "XX")
	if invalid {
		t.Errorf("ValidateChecksum should return false for invalid checksum")
	}
}

func TestGenerateLicenseKey(t *testing.T) {
	testCases := []struct {
		productCode string
		serial      string
		tier        license.Tier
		wantErr     bool
	}{
		{"1001", "ABCDEFG", license.TierReflector, false},
		{"2001", "1234567", license.TierTestSuite, false},
		{"3001", "XYZXYZX", license.TierEnterprise, false},
		{"100", "ABCDEFG", license.TierReflector, true}, // Invalid product code length.
		{"1001", "ABCDEF", license.TierReflector, true}, // Invalid serial length.
		{"1001", "ABCDEFG", license.Tier(0), true},      // Invalid tier.
	}

	for _, tc := range testCases {
		key, err := license.GenerateLicenseKey(tc.productCode, tc.serial, tc.tier)
		if tc.wantErr {
			if err == nil {
				t.Errorf("GenerateLicenseKey(%q, %q, %d) should return error", tc.productCode, tc.serial, tc.tier)
			}
			continue
		}

		if err != nil {
			t.Errorf(
				"GenerateLicenseKey(%q, %q, %d) returned unexpected error: %v",
				tc.productCode, tc.serial, tc.tier, err,
			)
			continue
		}

		const expectedKeyLen = 16
		if len(key) != expectedKeyLen {
			t.Errorf("Generated key length wrong: got=%d, want=16", len(key))
		}
	}
}

func TestValidateLicenseKey(t *testing.T) {
	// Generate a valid key and test validation.
	key, err := license.GenerateLicenseKey("1001", "ABCDEFG", license.TierReflector)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	info := license.ValidateLicenseKey(key)
	if !info.Valid {
		t.Errorf("ValidateLicenseKey should return valid for generated key: %s, error: %s", key, info.ErrorMsg)
	}

	// Test invalid keys.
	invalidKeys := []struct {
		key     string
		wantErr string
	}{
		{"", "License key must be 16 characters"},
		{"SHORT", "License key must be 16 characters"},
		{"INVALID-CHARS-@@", "License key must contain only letters and numbers"},
	}

	for _, tc := range invalidKeys {
		invalidInfo := license.ValidateLicenseKey(tc.key)
		if invalidInfo.Valid {
			t.Errorf("ValidateLicenseKey(%q) should not be valid", tc.key)
		}
	}
}

func TestTierString(t *testing.T) {
	testCases := []struct {
		tier     license.Tier
		expected string
	}{
		{license.TierReflector, "Reflector"},
		{license.TierTestSuite, "Test Suite"},
		{license.TierEnterprise, "Enterprise"},
		{license.TierInvalid, "Invalid"},
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
		{"SHORT", "SHORT"}, // Invalid length, return as-is.
	}

	for _, tc := range testCases {
		result := license.FormatKey(tc.input)
		if result != tc.expected {
			t.Errorf("FormatKey(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestInfoHasFeature(t *testing.T) {
	info := &license.Info{
		Key:         "",
		Valid:       false,
		Tier:        license.TierInvalid,
		ProductCode: "",
		Serial:      "",
		Activated:   false,
		ActivatedAt: time.Time{},
		ExpiresAt:   time.Time{},
		DeviceHash:  "",
		MaxDevices:  0,
		Features:    []string{"reflector", "rfc2544", "y1564"},
		ErrorMsg:    "",
	}

	if !info.HasFeature("reflector") {
		t.Error("HasFeature should return true for existing feature")
	}

	if info.HasFeature("nonexistent") {
		t.Error("HasFeature should return false for non-existing feature")
	}
}

func TestInfoCanRunReflector(t *testing.T) {
	tests := []struct {
		info *license.Info
		want bool
	}{
		{
			&license.Info{
				Key:         "",
				Valid:       true,
				Tier:        license.TierReflector,
				ProductCode: "",
				Serial:      "",
				Activated:   false,
				ActivatedAt: time.Time{},
				ExpiresAt:   time.Time{},
				DeviceHash:  "",
				MaxDevices:  0,
				Features:    nil,
				ErrorMsg:    "",
			},
			true,
		},
		{
			&license.Info{
				Key:         "",
				Valid:       true,
				Tier:        license.TierTestSuite,
				ProductCode: "",
				Serial:      "",
				Activated:   false,
				ActivatedAt: time.Time{},
				ExpiresAt:   time.Time{},
				DeviceHash:  "",
				MaxDevices:  0,
				Features:    nil,
				ErrorMsg:    "",
			},
			true,
		},
		{
			&license.Info{
				Key:         "",
				Valid:       true,
				Tier:        license.TierEnterprise,
				ProductCode: "",
				Serial:      "",
				Activated:   false,
				ActivatedAt: time.Time{},
				ExpiresAt:   time.Time{},
				DeviceHash:  "",
				MaxDevices:  0,
				Features:    nil,
				ErrorMsg:    "",
			},
			true,
		},
		{
			&license.Info{
				Key:         "",
				Valid:       false,
				Tier:        license.TierReflector,
				ProductCode: "",
				Serial:      "",
				Activated:   false,
				ActivatedAt: time.Time{},
				ExpiresAt:   time.Time{},
				DeviceHash:  "",
				MaxDevices:  0,
				Features:    nil,
				ErrorMsg:    "",
			},
			false,
		},
		{
			&license.Info{
				Key:         "",
				Valid:       true,
				Tier:        license.TierInvalid,
				ProductCode: "",
				Serial:      "",
				Activated:   false,
				ActivatedAt: time.Time{},
				ExpiresAt:   time.Time{},
				DeviceHash:  "",
				MaxDevices:  0,
				Features:    nil,
				ErrorMsg:    "",
			},
			false,
		},
	}

	for i, tc := range tests {
		if tc.info.CanRunReflector() != tc.want {
			t.Errorf("Test %d: CanRunReflector() = %v, want %v", i, tc.info.CanRunReflector(), tc.want)
		}
	}
}

func TestInfoCanRunTests(t *testing.T) {
	tests := []struct {
		info *license.Info
		want bool
	}{
		{
			&license.Info{
				Key:         "",
				Valid:       true,
				Tier:        license.TierReflector,
				ProductCode: "",
				Serial:      "",
				Activated:   false,
				ActivatedAt: time.Time{},
				ExpiresAt:   time.Time{},
				DeviceHash:  "",
				MaxDevices:  0,
				Features:    nil,
				ErrorMsg:    "",
			},
			false,
		},
		{
			&license.Info{
				Key:         "",
				Valid:       true,
				Tier:        license.TierTestSuite,
				ProductCode: "",
				Serial:      "",
				Activated:   false,
				ActivatedAt: time.Time{},
				ExpiresAt:   time.Time{},
				DeviceHash:  "",
				MaxDevices:  0,
				Features:    nil,
				ErrorMsg:    "",
			},
			true,
		},
		{
			&license.Info{
				Key:         "",
				Valid:       true,
				Tier:        license.TierEnterprise,
				ProductCode: "",
				Serial:      "",
				Activated:   false,
				ActivatedAt: time.Time{},
				ExpiresAt:   time.Time{},
				DeviceHash:  "",
				MaxDevices:  0,
				Features:    nil,
				ErrorMsg:    "",
			},
			true,
		},
		{
			&license.Info{
				Key:         "",
				Valid:       false,
				Tier:        license.TierTestSuite,
				ProductCode: "",
				Serial:      "",
				Activated:   false,
				ActivatedAt: time.Time{},
				ExpiresAt:   time.Time{},
				DeviceHash:  "",
				MaxDevices:  0,
				Features:    nil,
				ErrorMsg:    "",
			},
			false,
		},
	}

	for i, tc := range tests {
		if tc.info.CanRunTests() != tc.want {
			t.Errorf("Test %d: CanRunTests() = %v, want %v", i, tc.info.CanRunTests(), tc.want)
		}
	}
}

func TestDeviceFingerprint(t *testing.T) {
	fp, err := license.GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint failed: %v", err)
	}

	// Verify hash is 16 characters.
	hash := fp.Hash()
	const expectedHashLen = 16
	if len(hash) != expectedHashLen {
		t.Errorf("Fingerprint hash length = %d, want 16", len(hash))
	}

	// Verify hash is consistent.
	hash2 := fp.Hash()
	if hash != hash2 {
		t.Error("Fingerprint hash should be consistent")
	}

	// Verify String() doesn't panic.
	str := fp.String()
	if str == "" {
		t.Error("Fingerprint String() should not be empty")
	}
}
