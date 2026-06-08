// SPDX-License-Identifier: BUSL-1.1

package license_test

import (
	"crypto/ed25519"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

// TestValidateLicenseKey validates a test-key-signed token through the matching
// verifier and confirms arbitrary garbage is rejected.
func TestValidateLicenseKey(t *testing.T) {
	t.Parallel()
	// Mint a valid token and validate it against the test verifier.
	key, err := signTestKey(t, "1001", "ABCDEFG", license.TierReflector)
	if err != nil {
		t.Fatalf("Failed to sign test key: %v", err)
	}

	info := testVerifier(t).Validate(key)
	if !info.Valid {
		t.Errorf("Validate should return valid for signed key: %s, error: %s", key, info.ErrorMsg)
	}

	// Garbage strings must be rejected by the production verifier.
	invalidKeys := []string{"", "SHORT", "INVALID-CHARS-@@", "MSN1.notbase64.notbase64"}
	for _, k := range invalidKeys {
		invalidInfo := license.ValidateLicenseKey(k)
		if invalidInfo.Valid {
			t.Errorf("ValidateLicenseKey(%q) should not be valid", k)
		}
		if invalidInfo.ErrorMsg == "" {
			t.Errorf("ValidateLicenseKey(%q) should set a non-empty ErrorMsg", k)
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

// TestFormatKey verifies FormatKey trims surrounding whitespace and otherwise
// returns the token unchanged (tokens are already display-ready; base64url
// content must not be altered).
func TestFormatKey(t *testing.T) {
	t.Parallel()
	key, err := signTestKey(t, "2001", "1234567", license.TierProfessional)
	if err != nil {
		t.Fatalf("sign test key: %v", err)
	}

	testCases := []struct {
		input    string
		expected string
	}{
		{key, key},
		{"  " + key + "\n", key},
		{"SHORT", "SHORT"},
		{"", ""},
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
				Tier:        license.TierProfessional,
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
				Tier:        license.TierProfessional,
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
				Tier:        license.TierProfessional,
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

// TestValidateLicenseKeyEnterprise tests validation of an Enterprise-tier token.
func TestValidateLicenseKeyEnterprise(t *testing.T) {
	t.Parallel()
	key, err := signTestKey(t, "3001", "1234567", license.TierEnterprise)
	if err != nil {
		t.Fatalf("signTestKey() error: %v", err)
	}

	info := testVerifier(t).Validate(key)
	if !info.Valid {
		t.Errorf("Enterprise key should be valid: %s", info.ErrorMsg)
	}
	if info.Tier != license.TierEnterprise {
		t.Errorf("Expected TierEnterprise, got %v", info.Tier)
	}

	// Enterprise grants the Pro feature set.
	if !info.HasFeature("api") {
		t.Error("Enterprise should have api feature")
	}
	if !info.HasFeature("multiuser") {
		t.Error("Enterprise should have multiuser feature")
	}

	// FormatKey only trims whitespace; the trimmed token still validates.
	formatted := license.FormatKey("  " + key + "\n")
	if formatted != key {
		t.Errorf("FormatKey should trim to %q, got %q", key, formatted)
	}
	infoFormatted := testVerifier(t).Validate(formatted)
	if !infoFormatted.Valid {
		t.Error("Trimmed token should validate")
	}
}

// TestValidateLicenseKeyProductCodeMismatch verifies a signed token whose
// product code does not match its tier is rejected.
func TestValidateLicenseKeyProductCodeMismatch(t *testing.T) {
	t.Parallel()
	// Sign a Reflector tier with the Professional product code — a pairing the
	// catalog never issues. The signature is valid but the code/tier check fails.
	mismatched := signLicenseToken(
		t, testSigningKey(t), "2001", "1234567", license.TierReflector, 0,
	)
	info := testVerifier(t).Validate(mismatched)
	if info.Valid {
		t.Error("Token with mismatched product code/tier should be rejected")
	}
	if info.ErrorMsg == "" {
		t.Error("Mismatched token should set a non-empty ErrorMsg")
	}
}

// TestValidateLicenseKeyProductVariants verifies each valid product-code/tier
// pairing validates and is mapped to the right tier/code/serial.
func TestValidateLicenseKeyProductVariants(t *testing.T) {
	t.Parallel()
	tiers := []struct {
		product string
		tier    license.Tier
	}{
		{"1001", license.TierReflector},
		{"2001", license.TierProfessional},
		{"3001", license.TierEnterprise},
	}

	for _, tc := range tiers {
		key, err := signTestKey(t, tc.product, "1234567", tc.tier)
		if err != nil {
			t.Errorf("signTestKey(%s, %v) error: %v", tc.product, tc.tier, err)
			continue
		}

		info := testVerifier(t).Validate(key)
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

// TestValidateLicenseKeySerialExtraction tests that the serial is preserved.
func TestValidateLicenseKeySerialExtraction(t *testing.T) {
	t.Parallel()
	serial := "ABCDEFG"
	key, err := signTestKey(t, "1001", serial, license.TierReflector)
	if err != nil {
		t.Fatalf("signTestKey error: %v", err)
	}

	info := testVerifier(t).Validate(key)
	if !info.Valid {
		t.Errorf("Key should be valid: %s", info.ErrorMsg)
	}
	if info.Serial != serial {
		t.Errorf("Serial should be %s, got %s", serial, info.Serial)
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

// TestInfoMethods tests Info struct methods across all tier combinations.
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
		{false, license.TierProfessional, false, false},
		{false, license.TierEnterprise, false, false},
		{true, license.TierInvalid, false, false},
		{true, license.TierReflector, true, false},
		{true, license.TierProfessional, true, true},
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
		Tier:        license.TierProfessional,
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

// TestForgeryRejected verifies a token signed by an attacker key (not the
// production key) is rejected by the production verifier. This is the core
// security property of the Ed25519 scheme: a copy of the binary cannot mint
// valid licenses.
func TestForgeryRejected(t *testing.T) {
	t.Parallel()
	_, attackerPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate attacker key: %v", err)
	}

	forged := signLicenseToken(t, attackerPriv, "2001", "1234567", license.TierProfessional, 0)
	info := license.ValidateLicenseKey(forged)
	if info.Valid {
		t.Error("Token signed by a non-production key must be rejected by the production verifier")
	}
	if info.ErrorMsg != license.ErrLicenseInvalid {
		t.Errorf("forged token ErrorMsg = %q, want %q", info.ErrorMsg, license.ErrLicenseInvalid)
	}
}

// prodVector is a (production-signed token, expected-validation) pair produced
// by the canonical keygen tool against the embedded production key.
type prodVector struct {
	name     string
	key      string
	tier     license.Tier
	product  string
	serial   string
	features []string
}

func assertProdVector(t *testing.T, v prodVector) {
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

// TestKeygenContract pins the cross-tool signing contract. Every token in this
// table was produced by the canonical keygen tool (msn-internal-tools/keygen)
// against the embedded production key and must validate identically in every
// product's license package (stem, seed, niac). If this test fails, the
// embedded public key has drifted from keygen's private key — DO NOT "fix" the
// assertions; regenerate keygen and update all three products in lockstep.
func TestKeygenContract(t *testing.T) {
	t.Parallel()
	proFeatures := []string{
		"reflector", "api", "mef", "multiuser",
		"rfc2544", "rfc2889", "rfc6349", "tsn",
		"y1564", "y1731",
	}
	vectors := []prodVector{
		{
			name:     "stem-reflector / serial 1234567",
			key:      prodStemReflectorVector,
			tier:     license.TierReflector,
			product:  "1001",
			serial:   "1234567",
			features: []string{"reflector"},
		},
		{
			name:     "stem-professional / serial 1234567",
			key:      prodStemProfessionalVector,
			tier:     license.TierProfessional,
			product:  "2001",
			serial:   "1234567",
			features: proFeatures,
		},
		{
			name:     "stem-enterprise / serial 1234567",
			key:      prodStemEnterpriseVector,
			tier:     license.TierEnterprise,
			product:  "3001",
			serial:   "1234567",
			features: proFeatures,
		},
	}

	for _, v := range vectors {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			assertProdVector(t, v)
		})
	}
}
