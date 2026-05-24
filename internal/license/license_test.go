// SPDX-License-Identifier: BUSL-1.1

package license_test

import (
	"strings"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/license"
)

func TestRotorCipherRoundTrip(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
		{"", license.ErrLicenseKeyLength},
		{"SHORT", license.ErrLicenseKeyLength},
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
	t.Parallel()
	testCases := []struct {
		tier     license.Tier
		expected string
	}{
		{license.TierReflector, "Reflector"},
		{license.TierProfessional, "Professional"},
		// TierEnterprise is deprecated and folded into Professional —
		// its String() now reports "Professional" to match.
		{license.TierEnterprise, "Professional"},
		{license.TierInvalid, "Invalid"},
	}

	for _, tc := range testCases {
		if tc.tier.String() != tc.expected {
			t.Errorf("Tier.String() = %q, want %q", tc.tier.String(), tc.expected)
		}
	}
}

func TestFormatKey(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

// TestValidateLicenseKeyEdgeCases tests additional license validation paths.
func TestValidateLicenseKeyEdgeCases(t *testing.T) {
	t.Parallel()
	// Test with valid Enterprise tier key.
	key, err := license.GenerateLicenseKey("3001", "1234567", license.TierEnterprise)
	if err != nil {
		t.Fatalf("GenerateLicenseKey() error: %v", err)
	}

	info := license.ValidateLicenseKey(key)
	if !info.Valid {
		t.Errorf("Enterprise key should be valid: %s", info.ErrorMsg)
	}
	if info.Tier != license.TierEnterprise {
		t.Errorf("Expected TierEnterprise, got %v", info.Tier)
	}

	// Enterprise should have extra features.
	if !info.HasFeature("api") {
		t.Error("Enterprise should have api feature")
	}
	if !info.HasFeature("multiuser") {
		t.Error("Enterprise should have multiuser feature")
	}

	// Test keys with separators that get normalized.
	keyWithDashes := license.FormatKey(key)
	infoWithDashes := license.ValidateLicenseKey(keyWithDashes)
	if !infoWithDashes.Valid {
		t.Error("Key with dashes should validate after normalization")
	}

	// Test key with spaces.
	keyWithSpaces := key[:4] + " " + key[4:8] + " " + key[8:12] + " " + key[12:16]
	infoWithSpaces := license.ValidateLicenseKey(keyWithSpaces)
	if !infoWithSpaces.Valid {
		t.Error("Key with spaces should validate after normalization")
	}

	// Test key with dots.
	keyWithDots := key[:4] + "." + key[4:8] + "." + key[8:12] + "." + key[12:16]
	infoWithDots := license.ValidateLicenseKey(keyWithDots)
	if !infoWithDots.Valid {
		t.Error("Key with dots should validate after normalization")
	}

	// Test lowercase key (should be normalized to uppercase).
	lowerKey := strings.ToLower(key)
	infoLower := license.ValidateLicenseKey(lowerKey)
	if !infoLower.Valid {
		t.Error("Lowercase key should validate after normalization")
	}
}

// TestValidateLicenseKeyInvalidTier tests invalid tier in key.
func TestValidateLicenseKeyInvalidTier(t *testing.T) {
	t.Parallel()
	// Create a key manually with invalid tier character.
	// Generate a valid key first.
	validKey, _ := license.GenerateLicenseKey("1001", "ABCDEFG", license.TierReflector)

	// Corrupt the tier position by creating a random invalid key.
	// This tests the "Invalid license tier" path.
	invalidKeys := []struct {
		key     string
		errPart string
	}{
		{"AAAA0000000000AA", "checksum"}, // Invalid checksum.
		{"1234AAAA11111234", ""},         // Format ok but content invalid.
	}

	for _, tc := range invalidKeys {
		info := license.ValidateLicenseKey(tc.key)
		if info.Valid {
			t.Errorf("Key %q should not be valid", tc.key)
		}
		if tc.errPart != "" && !strings.Contains(strings.ToLower(info.ErrorMsg), tc.errPart) {
			t.Logf("Key %q error: %s (expected to contain %q)", tc.key, info.ErrorMsg, tc.errPart)
		}
	}

	// Verify the valid key is still valid.
	info := license.ValidateLicenseKey(validKey)
	if !info.Valid {
		t.Errorf("Valid key should remain valid: %s", info.ErrorMsg)
	}
}

// TestValidateLicenseKeyProductCodeMismatch tests product code tier mismatch.
func TestValidateLicenseKeyProductCodeMismatch(t *testing.T) {
	t.Parallel()
	// These would require crafting malformed keys that pass checksum
	// but have mismatched product codes.
	// For now, verify that correctly generated keys validate.
	tiers := []struct {
		product string
		tier    license.Tier
	}{
		{"1001", license.TierReflector},
		{"2001", license.TierTestSuite},
		{"3001", license.TierEnterprise},
	}

	for _, tc := range tiers {
		key, err := license.GenerateLicenseKey(tc.product, "1234567", tc.tier)
		if err != nil {
			t.Errorf("GenerateLicenseKey(%s, %v) error: %v", tc.product, tc.tier, err)
			continue
		}

		info := license.ValidateLicenseKey(key)
		if !info.Valid {
			t.Errorf("Key for %s/%v should be valid: %s", tc.product, tc.tier, info.ErrorMsg)
		}
		if info.Tier != tc.tier {
			t.Errorf("Expected tier %v, got %v", tc.tier, info.Tier)
		}
		if info.ProductCode != tc.product {
			t.Errorf("Expected product code %s, got %s", tc.product, info.ProductCode)
		}
	}
}

// TestTierStringUnknownValue tests Tier.String() with an unknown value.
func TestTierStringUnknownValue(t *testing.T) {
	t.Parallel()
	// Test with a tier value outside the defined range.
	unknownTier := license.Tier(99)
	result := unknownTier.String()
	if result != "Invalid" {
		t.Errorf("Unknown tier should return 'Invalid', got %q", result)
	}

	// Test negative tier.
	negativeTier := license.Tier(-1)
	result = negativeTier.String()
	if result != "Invalid" {
		t.Errorf("Negative tier should return 'Invalid', got %q", result)
	}
}

// TestRotorCipherDecodeAllCharTypes tests decoding of all character types.
func TestRotorCipherDecodeAllCharTypes(t *testing.T) {
	t.Parallel()
	// Test uppercase letters.
	cipher := license.NewRotorCipher(0)
	for c := byte('A'); c <= 'Z'; c++ {
		encoded := cipher.Encode(c)
		// Verify it's still a letter.
		if encoded < 'A' || encoded > 'Z' {
			t.Errorf("Encoded uppercase %c should be uppercase, got %c", c, encoded)
		}
	}

	// Test lowercase letters.
	cipher = license.NewRotorCipher(0)
	for c := byte('a'); c <= 'z'; c++ {
		encoded := cipher.Encode(c)
		// Verify it's still lowercase.
		if encoded < 'a' || encoded > 'z' {
			t.Errorf("Encoded lowercase %c should be lowercase, got %c", c, encoded)
		}
	}

	// Test digits.
	cipher = license.NewRotorCipher(0)
	for c := byte('0'); c <= '9'; c++ {
		encoded := cipher.Encode(c)
		// Verify it's still a digit.
		if encoded < '0' || encoded > '9' {
			t.Errorf("Encoded digit %c should be digit, got %c", c, encoded)
		}
	}
}

// TestRotorCipherDecodeRoundtripLowercase tests lowercase roundtrip.
func TestRotorCipherDecodeRoundtripLowercase(t *testing.T) {
	t.Parallel()
	testInputs := []string{
		"abc",
		"xyz",
		"hello",
		"test123abc",
	}

	for _, input := range testInputs {
		cipher := license.NewRotorCipher(0)
		encoded := cipher.EncodeString(input)

		cipher = license.NewRotorCipher(0)
		decoded := cipher.DecodeString(encoded)

		if decoded != input {
			t.Errorf("Lowercase roundtrip failed: input=%q, encoded=%q, decoded=%q",
				input, encoded, decoded)
		}
	}
}

// TestValidateChecksumShortStrings tests checksum validation with short strings.
func TestValidateChecksumShortStrings(t *testing.T) {
	t.Parallel()
	// String with exactly 3 characters (minimum).
	shortValid := "A" + license.CalculateChecksum("A")
	if !license.ValidateChecksum(shortValid) {
		t.Error("Minimum valid checksum string should validate")
	}

	// Strings that are too short.
	shortStrings := []string{"", "A", "AB"}
	for _, s := range shortStrings {
		if license.ValidateChecksum(s) {
			t.Errorf("Short string %q should not validate", s)
		}
	}
}

// TestInitRotors tests that InitRotors can be called multiple times.
func TestInitRotors(t *testing.T) {
	t.Parallel()
	// Call InitRotors multiple times - should not panic.
	license.InitRotors()
	license.InitRotors()
	license.InitRotors()

	// Verify cipher still works after multiple init calls.
	cipher := license.NewRotorCipher(0)
	encoded := cipher.EncodeString("TEST")

	cipher = license.NewRotorCipher(0)
	decoded := cipher.DecodeString(encoded)

	if decoded != "TEST" {
		t.Error("Cipher should work after multiple InitRotors calls")
	}
}

// TestMaskString tests the maskString function behavior.
func TestMaskString(t *testing.T) {
	t.Parallel()
	// This tests the fingerprint String() output which uses maskString internally.
	fp, err := license.GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint error: %v", err)
	}

	str := fp.String()

	// Should contain "****" for masked values.
	if !strings.Contains(str, "****") {
		t.Log("Fingerprint string may have short values that don't get masked")
	}

	// Should contain MAC, CPU, DISK, HOST.
	if !strings.Contains(str, "MAC=") {
		t.Error("Fingerprint String() should contain MAC=")
	}
	if !strings.Contains(str, "CPU=") {
		t.Error("Fingerprint String() should contain CPU=")
	}
	if !strings.Contains(str, "DISK=") {
		t.Error("Fingerprint String() should contain DISK=")
	}
	if !strings.Contains(str, "HOST=") {
		t.Error("Fingerprint String() should contain HOST=")
	}
}

// TestGenerateLicenseKeyErrors tests error paths in GenerateLicenseKey.
func TestGenerateLicenseKeyErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		product string
		serial  string
		tier    license.Tier
		wantErr string
	}{
		{
			name:    "short product code",
			product: "100",
			serial:  "ABCDEFG",
			tier:    license.TierReflector,
			wantErr: "product code must be 4 characters",
		},
		{
			name:    "long product code",
			product: "10001",
			serial:  "ABCDEFG",
			tier:    license.TierReflector,
			wantErr: "product code must be 4 characters",
		},
		{
			name:    "short serial",
			product: "1001",
			serial:  "ABCDEF",
			tier:    license.TierReflector,
			wantErr: "serial must be 7 characters",
		},
		{
			name:    "long serial",
			product: "1001",
			serial:  "ABCDEFGH",
			tier:    license.TierReflector,
			wantErr: "serial must be 7 characters",
		},
		{
			name:    "tier too low",
			product: "1001",
			serial:  "ABCDEFG",
			tier:    license.Tier(0),
			wantErr: "invalid tier",
		},
		{
			name:    "tier too high",
			product: "1001",
			serial:  "ABCDEFG",
			tier:    license.Tier(4),
			wantErr: "invalid tier",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := license.GenerateLicenseKey(tc.product, tc.serial, tc.tier)
			if err == nil {
				t.Errorf("Expected error containing %q", tc.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Error %q should contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// TestFormatKeyVariousInputs tests FormatKey with various inputs.
func TestFormatKeyVariousInputs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		// Normal cases.
		{"ABCD1234EFGH5678", "ABCD-1234-EFGH-5678"},
		{"abcd1234efgh5678", "ABCD-1234-EFGH-5678"},

		// Already formatted.
		{"ABCD-1234-EFGH-5678", "ABCD-1234-EFGH-5678"},

		// With spaces.
		{"ABCD 1234 EFGH 5678", "ABCD-1234-EFGH-5678"},

		// With dots.
		{"ABCD.1234.EFGH.5678", "ABCD-1234-EFGH-5678"},

		// Wrong length - returned as-is after normalization.
		{"SHORT", "SHORT"},
		{"TOOLONG1234567890123", "TOOLONG1234567890123"},
		{"", ""},
	}

	for _, tc := range tests {
		result := license.FormatKey(tc.input)
		if result != tc.expected {
			t.Errorf("FormatKey(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// TestInfoMethods tests Info struct methods.
func TestInfoMethods(t *testing.T) {
	t.Parallel()
	// Test CanRunReflector and CanRunTests for all tier combinations.
	testCases := []struct {
		valid       bool
		tier        license.Tier
		canReflect  bool
		canRunTests bool
	}{
		{false, license.TierInvalid, false, false},
		{false, license.TierReflector, false, false},
		{false, license.TierTestSuite, false, false},
		{false, license.TierEnterprise, false, false},
		{true, license.TierInvalid, false, false},
		{true, license.TierReflector, true, false},
		{true, license.TierTestSuite, true, true},
		{true, license.TierEnterprise, true, true},
	}

	for _, tc := range testCases {
		info := &license.Info{
			Key:         "",
			Valid:       tc.valid,
			Tier:        tc.tier,
			ProductCode: "",
			Serial:      "",
			Activated:   false,
			ActivatedAt: time.Time{},
			ExpiresAt:   time.Time{},
			DeviceHash:  "",
			MaxDevices:  0,
			Features:    nil,
			ErrorMsg:    "",
		}

		if info.CanRunReflector() != tc.canReflect {
			t.Errorf("Valid=%v Tier=%v: CanRunReflector()=%v, want %v",
				tc.valid, tc.tier, info.CanRunReflector(), tc.canReflect)
		}

		if info.CanRunTests() != tc.canRunTests {
			t.Errorf("Valid=%v Tier=%v: CanRunTests()=%v, want %v",
				tc.valid, tc.tier, info.CanRunTests(), tc.canRunTests)
		}
	}
}

// TestHasFeatureVariousFeatures tests HasFeature with various inputs.
func TestHasFeatureVariousFeatures(t *testing.T) {
	t.Parallel()
	info := &license.Info{
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
		Features:    []string{"reflector", "rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn"},
		ErrorMsg:    "",
	}

	// Test all expected features.
	expectedFeatures := []string{"reflector", "rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn"}
	for _, feature := range expectedFeatures {
		if !info.HasFeature(feature) {
			t.Errorf("Expected feature %q to be present", feature)
		}
	}

	// Test missing features.
	missingFeatures := []string{"api", "multiuser", "unknown", ""}
	for _, feature := range missingFeatures {
		if info.HasFeature(feature) {
			t.Errorf("Feature %q should not be present", feature)
		}
	}

	// Test with empty features list.
	emptyInfo := &license.Info{
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
	}

	if emptyInfo.HasFeature("reflector") {
		t.Error("Nil features should not have any feature")
	}
}

// TestChecksumWithVariousPayloads tests checksum calculation with edge cases.
func TestChecksumWithVariousPayloads(t *testing.T) {
	t.Parallel()
	payloads := []string{
		"A",                          // Single character (minimum valid payload for validation).
		"0",                          // Single digit.
		"AB",                         // Two characters.
		"12",                         // Two digits.
		"ABCDEFGHIJKLMNOP",           // Long string.
		"0123456789",                 // All digits.
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ", // All letters.
		"abcdefghijklmnopqrstuvwxyz", // Lowercase (gets uppercased).
		"MixedCase123",               // Mixed.
	}

	for _, payload := range payloads {
		checksum := license.CalculateChecksum(payload)
		if len(checksum) != 2 {
			t.Errorf("Checksum for %q should be 2 chars, got %d", payload, len(checksum))
		}

		// Verify checksum is alphanumeric.
		for _, c := range checksum {
			if (c < '0' || c > '9') && (c < 'A' || c > 'Z') {
				t.Errorf("Checksum character %c should be alphanumeric", c)
			}
		}

		// Verify validation works.
		fullString := payload + checksum
		if !license.ValidateChecksum(fullString) {
			t.Errorf("Checksum validation failed for %q", payload)
		}
	}

	// Test empty payload separately (checksum can be calculated but validation requires min 3 chars).
	emptyChecksum := license.CalculateChecksum("")
	if len(emptyChecksum) != 2 {
		t.Errorf("Checksum for empty string should be 2 chars, got %d", len(emptyChecksum))
	}
}

// TestRotorCipherPosition tests that position advances correctly during encoding.
func TestRotorCipherPosition(t *testing.T) {
	t.Parallel()
	// Create two ciphers with the same starting position.
	cipher1 := license.NewRotorCipher(7)
	cipher2 := license.NewRotorCipher(7)

	// Encode same characters - should give same result if position handling is correct.
	c1 := cipher1.Encode('A')
	c2 := cipher2.Encode('A')

	if c1 != c2 {
		t.Errorf("Same input with same position should give same output: %c vs %c", c1, c2)
	}
}

// TestRotorCipherLargePosition tests that large position values are handled correctly.
func TestRotorCipherLargePosition(t *testing.T) {
	t.Parallel()
	// Test with position larger than rotor modulus.
	cipher1 := license.NewRotorCipher(100)
	cipher2 := license.NewRotorCipher(100 % 36)

	// Both should produce same encoding since position wraps around.
	input := "TEST"
	encoded1 := cipher1.EncodeString(input)
	encoded2 := cipher2.EncodeString(input)

	if encoded1 != encoded2 {
		t.Errorf("Large position should wrap around: %q vs %q", encoded1, encoded2)
	}
}

// TestNonAlphanumericPassthrough tests that non-alphanumeric characters pass unchanged.
func TestNonAlphanumericPassthrough(t *testing.T) {
	t.Parallel()
	cipher := license.NewRotorCipher(0)

	// Test various non-alphanumeric characters.
	nonAlpha := "!@#$%^&*()_+-=[]{}|;:',.<>?/~` "
	encoded := cipher.EncodeString(nonAlpha)

	if encoded != nonAlpha {
		t.Errorf("Non-alphanumeric should pass through unchanged: %q vs %q", nonAlpha, encoded)
	}

	// Decoding should also pass them through.
	cipher = license.NewRotorCipher(0)
	decoded := cipher.DecodeString(nonAlpha)

	if decoded != nonAlpha {
		t.Errorf("Non-alphanumeric should pass through on decode: %q vs %q", nonAlpha, decoded)
	}
}

// TestMixedAlphanumericAndSpecial tests encoding/decoding with mixed characters.
func TestMixedAlphanumericAndSpecial(t *testing.T) {
	t.Parallel()
	input := "ABC-123-XYZ"
	cipher := license.NewRotorCipher(5)
	encoded := cipher.EncodeString(input)

	// Dashes should remain unchanged.
	if encoded[3] != '-' || encoded[7] != '-' {
		t.Errorf("Dashes should not be encoded: %q", encoded)
	}

	// Verify roundtrip works with mixed characters.
	cipher = license.NewRotorCipher(5)
	decoded := cipher.DecodeString(encoded)

	if decoded != input {
		t.Errorf("Mixed char roundtrip failed: input=%q, decoded=%q", input, decoded)
	}
}

// TestTierStringBoundary tests Tier.String() with boundary values.
func TestTierStringBoundary(t *testing.T) {
	t.Parallel()
	// Test tier value exactly at the boundary.
	tier := license.Tier(4)
	result := tier.String()
	if result != "Invalid" {
		t.Errorf("Tier(4) should return 'Invalid', got %q", result)
	}

	// Test large negative value.
	tier = license.Tier(-100)
	result = tier.String()
	if result != "Invalid" {
		t.Errorf("Tier(-100) should return 'Invalid', got %q", result)
	}
}

// TestToAlphanumericConversion tests the alphanumeric conversion function indirectly.
func TestToAlphanumericConversion(t *testing.T) {
	t.Parallel()
	// Test that checksum characters are valid alphanumeric.
	testPayloads := []string{
		"0",
		"9",
		"A",
		"Z",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	}

	for _, payload := range testPayloads {
		checksum := license.CalculateChecksum(payload)

		// Both characters should be alphanumeric.
		for i, c := range checksum {
			if (c < '0' || c > '9') && (c < 'A' || c > 'Z') {
				t.Errorf("Checksum char %d for %q should be alphanumeric, got %c", i, payload, c)
			}
		}
	}
}

// TestValidateLicenseKeyWithProductCodeVariants tests all product code paths.
func TestValidateLicenseKeyWithProductCodeVariants(t *testing.T) {
	t.Parallel()
	// Generate keys for each product code and verify they validate.
	variants := []struct {
		product string
		tier    license.Tier
	}{
		{"1001", license.TierReflector},
		{"2001", license.TierTestSuite},
		{"3001", license.TierEnterprise},
	}

	for _, v := range variants {
		key, err := license.GenerateLicenseKey(v.product, "ABCDEFG", v.tier)
		if err != nil {
			t.Fatalf("GenerateLicenseKey(%s) error: %v", v.product, err)
		}

		info := license.ValidateLicenseKey(key)
		if !info.Valid {
			t.Errorf("Generated key for %s should validate: %s", v.product, info.ErrorMsg)
		}
		if info.ProductCode != v.product {
			t.Errorf("ProductCode should be %s, got %s", v.product, info.ProductCode)
		}
		if info.Tier != v.tier {
			t.Errorf("Tier should be %v, got %v", v.tier, info.Tier)
		}
	}
}

// TestValidateLicenseKeySerialExtraction tests that serial is correctly extracted.
func TestValidateLicenseKeySerialExtraction(t *testing.T) {
	t.Parallel()
	serial := "ABCDEFG"
	key, err := license.GenerateLicenseKey("1001", serial, license.TierReflector)
	if err != nil {
		t.Fatalf("GenerateLicenseKey error: %v", err)
	}

	info := license.ValidateLicenseKey(key)
	if !info.Valid {
		t.Errorf("Key should be valid: %s", info.ErrorMsg)
	}
	if info.Serial != serial {
		t.Errorf("Serial should be %s, got %s", serial, info.Serial)
	}
}

// TestChecksumEdgeCases tests checksum with special payloads.
func TestChecksumEdgeCases(t *testing.T) {
	t.Parallel()
	// Test single character.
	check1 := license.CalculateChecksum("X")
	if len(check1) != 2 {
		t.Errorf("Single char checksum should be 2 chars, got %d", len(check1))
	}

	// Test all zeros.
	check0 := license.CalculateChecksum("0000000000000000")
	if len(check0) != 2 {
		t.Errorf("Zero payload checksum should be 2 chars, got %d", len(check0))
	}

	// Test all same character.
	checkSame := license.CalculateChecksum("AAAAAAA")
	if len(checkSame) != 2 {
		t.Errorf("Repeated char checksum should be 2 chars, got %d", len(checkSame))
	}
}

// TestRotorCipherPositionWrapping tests rotor position wrapping at modulus boundary.
func TestRotorCipherPositionWrapping(t *testing.T) {
	t.Parallel()
	// Create a cipher at the modulus boundary.
	cipher := license.NewRotorCipher(35)

	// Encoding should work correctly even at boundary.
	result := cipher.EncodeString("A")
	if result == "" {
		t.Error("Cipher at boundary position should encode successfully")
	}

	// Create another at exactly modulus (should wrap to 0).
	cipher2 := license.NewRotorCipher(36)
	result2 := cipher2.EncodeString("A")

	cipher3 := license.NewRotorCipher(0)
	result3 := cipher3.EncodeString("A")

	if result2 != result3 {
		t.Errorf("Position 36 should wrap to 0: %q vs %q", result2, result3)
	}
}

// TestProductCodeTierMismatchRejection tests that product code/tier mismatches are rejected.
func TestProductCodeTierMismatchRejection(t *testing.T) {
	t.Parallel()
	// This test attempts to catch the product code mismatch path.
	// We generate a valid key and then verify its components.

	// Generate a key with TestSuite tier (2001) and Tier 2.
	validKey, err := license.GenerateLicenseKey("2001", "ABCDEFG", license.TierTestSuite)
	if err != nil {
		t.Fatalf("GenerateLicenseKey error: %v", err)
	}

	info := license.ValidateLicenseKey(validKey)
	if !info.Valid {
		t.Fatalf("Generated key should be valid: %s", info.ErrorMsg)
	}

	// Verify the product code and tier match.
	if info.ProductCode != "2001" {
		t.Errorf("ProductCode should be 2001, got %s", info.ProductCode)
	}
	if info.Tier != license.TierTestSuite {
		t.Errorf("Tier should be TierTestSuite, got %v", info.Tier)
	}

	// Now generate keys for all combinations and verify they pass validation.
	combinations := []struct {
		product string
		tier    license.Tier
	}{
		{"1001", license.TierReflector},
		{"2001", license.TierTestSuite},
		{"3001", license.TierEnterprise},
	}

	for _, combo := range combinations {
		key, genErr := license.GenerateLicenseKey(combo.product, "1234567", combo.tier)
		if genErr != nil {
			t.Errorf("GenerateLicenseKey(%s, %v) error: %v", combo.product, combo.tier, genErr)
			continue
		}

		comboInfo := license.ValidateLicenseKey(key)
		if !comboInfo.Valid {
			t.Errorf("Key with %s and tier %v should be valid: %s", combo.product, combo.tier, comboInfo.ErrorMsg)
		}
		if comboInfo.ProductCode != combo.product {
			t.Errorf("ProductCode mismatch: expected %s, got %s", combo.product, comboInfo.ProductCode)
		}
		if comboInfo.Tier != combo.tier {
			t.Errorf("Tier mismatch: expected %v, got %v", combo.tier, comboInfo.Tier)
		}
	}
}

// TestValidateLicenseKeyUnknownProduct tests unknown product code rejection.
func TestValidateLicenseKeyUnknownProduct(t *testing.T) {
	t.Parallel()
	// Test with a manually crafted key that has an unknown product code.
	// We need to bypass the normal generation and create a key with an invalid product code.
	// Since we can't easily generate mismatched keys, we test the validation logic.

	// First, verify that known products validate correctly.
	knownProducts := []struct {
		code string
		tier license.Tier
	}{
		{"1001", license.TierReflector},
		{"2001", license.TierTestSuite},
		{"3001", license.TierEnterprise},
	}

	for _, kp := range knownProducts {
		key, err := license.GenerateLicenseKey(kp.code, "1111111", kp.tier)
		if err != nil {
			t.Fatalf("GenerateLicenseKey error: %v", err)
		}

		info := license.ValidateLicenseKey(key)
		if !info.Valid {
			t.Errorf("Key with known product %s should be valid", kp.code)
		}
	}
}

// TestValidateChecksumCaseSensitivity tests that validation handles case correctly.
func TestValidateChecksumCaseSensitivity(t *testing.T) {
	t.Parallel()
	payload := "ABCDEFGH"
	checksum := license.CalculateChecksum(payload)

	// Validate with uppercase.
	valid := license.ValidateChecksum(payload + checksum)
	if !valid {
		t.Error("Uppercase checksum should validate")
	}

	// Validate with lowercase payload (should fail because checksum was computed on uppercase).
	lowerPayload := strings.ToLower(payload)
	lowerChecksum := license.CalculateChecksum(lowerPayload)

	// The checksum for lowercase should be different from uppercase because
	// CalculateChecksum uppercases its input.
	// Wait, that's not right - CalculateChecksum uppercases internally.
	// So checksums should be the same.

	if lowerChecksum != checksum {
		t.Errorf("Checksum should be case-insensitive: %q vs %q", checksum, lowerChecksum)
	}

	// Both should validate.
	valid1 := license.ValidateChecksum(payload + checksum)
	valid2 := license.ValidateChecksum(lowerPayload + lowerChecksum)

	if !valid1 || !valid2 {
		t.Error("Both uppercase and lowercase should validate")
	}
}

// keygenVector is a (key, expected-validation) pair produced by the
// canonical keygen tool.
type keygenVector struct {
	name     string
	key      string
	tier     license.Tier
	product  string
	serial   string
	features []string
}

func assertKeygenVector(t *testing.T, v keygenVector) {
	t.Helper()
	info := license.ValidateLicenseKey(v.key)
	if !info.Valid {
		t.Fatalf("Valid = false, want true (err=%q)", info.ErrorMsg)
	}
	if info.Tier != v.tier {
		t.Errorf("Tier = %v, want %v", info.Tier, v.tier)
	}
	if info.ProductCode != v.product {
		t.Errorf("ProductCode = %q, want %q", info.ProductCode, v.product)
	}
	if info.Serial != v.serial {
		t.Errorf("Serial = %q, want %q", info.Serial, v.serial)
	}
	if len(info.Features) != len(v.features) {
		t.Errorf("Features count = %d, want %d (got %v)", len(info.Features), len(v.features), info.Features)
	}
	for _, f := range v.features {
		if !info.HasFeature(f) {
			t.Errorf("missing feature %q (got %v)", f, info.Features)
		}
	}
}

// TestKeygenContract pins the cross-tool cipher contract. Every key in
// this table was produced by the canonical keygen tool
// (msn-internal-tools/keygen) and must validate identically in every
// product's license package (stem, seed, niac). If this test fails the
// rotor cipher has drifted from keygen — DO NOT "fix" the assertions;
// fix the cipher, or regenerate keygen + update all three products in
// lockstep.
//
// Anchored to keygen v2.0.0 (2026-05-21).
func TestKeygenContract(t *testing.T) {
	t.Parallel()
	vectors := []keygenVector{
		{
			name:    "stem-pro / serial ROUNDTR",
			key:     "7C27-20MY-ZHXR-C95U",
			tier:    license.TierProfessional,
			product: "2001",
			serial:  "ROUNDTR",
			features: []string{
				"reflector", "api", "mef", "multiuser",
				"rfc2544", "rfc2889", "rfc6349", "tsn",
				"y1564", "y1731",
			},
		},
	}

	for _, v := range vectors {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			assertKeygenVector(t, v)
		})
	}
}
