// SPDX-License-Identifier: BUSL-1.1

package license

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

/*
MSN License Key Format (16 characters):
+------+--------+-------+------+----------+
| CC   | PPPP   |SSSSSSS| T    | XX       |
|Check |Product |Serial |Tier  | Checksum |
+------+--------+-------+------+----------+

Positions:
  0-1:  Checksum prefix (encoded validation).
  2-5:  Product code (1001=Reflector, 2001=Professional).
  6-12: Serial number (unique per license).
  13:   Tier (1=Reflector, 2=Professional, 3=Enterprise-deprecated).
  14-15: Checksum suffix.

Product Codes:
  1001: Stem Reflector (Tier 1).
  2001: Stem Professional (Tier 2). Was "Test Suite"; renamed per
        LICENSE_STRATEGY 2026-05-19. Wire value preserved.
  3001: Stem Enterprise (Tier 3, deprecated). Folded into Professional;
        existing keys continue to validate and grant Pro features.
*/

// License key format constants.
const (
	keyLength         = 16
	productCodeLength = 4
	serialLength      = 7
	checksumLength    = 2
	cipherStartPos    = 7 // MSN rotor cipher starting position.
	defaultMaxDevices = 3 // Default activations per license.
)

// Tier represents the license tier.
type Tier int

// License tier constants.
const (
	// TierInvalid represents an invalid or unrecognized license tier.
	TierInvalid Tier = 0
	// TierReflector provides reflector-only functionality.
	TierReflector Tier = 1
	// TierProfessional provides the full Stem test suite (RFC 2544 /
	// Y.1564 / Y.1731 / RFC 2889 / RFC 6349 / MEF / TSN) plus the
	// reflector and API access.
	TierProfessional Tier = 2
	// TierEnterprise is deprecated as a SKU per LICENSE_STRATEGY
	// 2026-05-19 (folded into Professional). The constant is retained
	// so previously issued Enterprise keys keep validating; they now
	// grant the same features as Professional.
	TierEnterprise Tier = 3

	// TierTestSuite is a deprecated alias for TierProfessional, kept
	// so external callers that referenced the old name still compile.
	// New code MUST use TierProfessional.
	//
	// Deprecated: use TierProfessional.
	TierTestSuite = TierProfessional
)

// Error messages.
const (
	errProductCodeMismatch = "Product code mismatch for tier"
	// ErrLicenseKeyLength indicates the key length validation error message.
	ErrLicenseKeyLength = "License key must be 16 characters"
)

// String returns the tier name.
func (t Tier) String() string {
	switch t {
	case TierInvalid:
		return "Invalid"
	case TierReflector:
		return "Reflector"
	case TierProfessional:
		return "Professional"
	case TierEnterprise:
		return "Professional"
	}
	return "Invalid"
}

// Info contains parsed license information.
type Info struct {
	Key         string    `json:"key"`
	Valid       bool      `json:"valid"`
	Tier        Tier      `json:"tier"`
	ProductCode string    `json:"productCode"`
	Serial      string    `json:"serial"`
	Activated   bool      `json:"activated"`
	ActivatedAt time.Time `json:"activatedAt,omitzero"`
	ExpiresAt   time.Time `json:"expiresAt,omitzero"`
	DeviceHash  string    `json:"deviceHash,omitempty"`
	MaxDevices  int       `json:"maxDevices"`
	Features    []string  `json:"features"`
	ErrorMsg    string    `json:"error,omitempty"`
}

// ValidateLicenseKey performs offline validation of a license key.
func ValidateLicenseKey(key string) *Info {
	info := &Info{
		Key:         key,
		Valid:       false,
		Tier:        TierInvalid,
		ProductCode: "",
		Serial:      "",
		Activated:   false,
		ActivatedAt: time.Time{},
		ExpiresAt:   time.Time{},
		DeviceHash:  "",
		MaxDevices:  defaultMaxDevices,
		Features:    nil,
		ErrorMsg:    "",
	}

	// Normalize key (remove spaces, dashes, uppercase).
	key = normalizeKey(key)
	info.Key = key

	// Check length.
	if len(key) != keyLength {
		info.ErrorMsg = ErrLicenseKeyLength
		return info
	}

	// Check format (alphanumeric only).
	if !regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(key) {
		info.ErrorMsg = "License key must contain only letters and numbers"
		return info
	}

	// Decode the key through rotor cipher first.
	cipher := NewRotorCipher(cipherStartPos)
	decoded := cipher.DecodeString(key)

	// Validate checksum on decoded key (uses positions 0-1 and 14-15).
	if !validateKeyChecksum(decoded) {
		info.ErrorMsg = "Invalid license key checksum"
		return info
	}

	// Extract components.
	info.ProductCode = decoded[2:6]
	info.Serial = decoded[6:13]
	tierChar := decoded[13]

	// Parse tier.
	switch tierChar {
	case '1':
		info.Tier = TierReflector
		info.Features = []string{"reflector"}
	case '2':
		info.Tier = TierProfessional
		info.Features = proFeatures()
	case '3':
		info.Tier = TierEnterprise
		info.Features = proFeatures()
	default:
		info.ErrorMsg = "Invalid license tier"
		return info
	}

	// Validate product code.
	switch info.ProductCode {
	case "1001":
		if info.Tier != TierReflector {
			info.ErrorMsg = errProductCodeMismatch
			return info
		}
	case "2001":
		if info.Tier != TierProfessional {
			info.ErrorMsg = errProductCodeMismatch
			return info
		}
	case "3001":
		if info.Tier != TierEnterprise {
			info.ErrorMsg = errProductCodeMismatch
			return info
		}
	default:
		info.ErrorMsg = "Unknown product code"
		return info
	}

	info.Valid = true
	return info
}

// validateKeyChecksum checks the embedded checksum.
func validateKeyChecksum(key string) bool {
	// Extract the core payload (positions 2-13).
	payload := key[2:14]

	// Calculate expected checksum.
	expected := CalculateChecksum(payload)

	// Compare with key prefix (0-1) and suffix (14-15).
	// Both checksum positions must match the expected value (AND logic).
	// This prevents bypass attacks where only one component matches.
	prefixMatch := key[0:2] == expected
	suffixMatch := key[14:16] == expected

	return prefixMatch && suffixMatch
}

// GenerateLicenseKey creates a new license key (for admin/generator tool).
func GenerateLicenseKey(productCode string, serial string, tier Tier) (string, error) {
	// Validate inputs.
	if len(productCode) != productCodeLength {
		return "", errors.New("product code must be 4 characters")
	}
	if len(serial) != serialLength {
		return "", errors.New("serial must be 7 characters")
	}
	if tier < TierReflector || tier > TierEnterprise {
		return "", errors.New("invalid tier")
	}

	// Build payload: PPPP + SSSSSSS + T.
	payload := productCode + serial + fmt.Sprintf("%d", tier)

	// Calculate checksum.
	checksum := CalculateChecksum(payload)

	// Build full key: CC + payload + XX.
	fullKey := checksum[0:checksumLength] + payload + checksum

	// Encode through rotor cipher.
	cipher := NewRotorCipher(cipherStartPos)
	encoded := cipher.EncodeString(fullKey)

	return strings.ToUpper(encoded), nil
}

// normalizeKey cleans up a license key for validation.
func normalizeKey(key string) string {
	// Remove common separators.
	key = strings.ReplaceAll(key, "-", "")
	key = strings.ReplaceAll(key, " ", "")
	key = strings.ReplaceAll(key, ".", "")

	// Uppercase.
	return strings.ToUpper(key)
}

// FormatKey formats a license key for display (adds dashes).
func FormatKey(key string) string {
	key = normalizeKey(key)
	if len(key) != keyLength {
		return key
	}
	// Format as XXXX-XXXX-XXXX-XXXX.
	return key[0:4] + "-" + key[4:8] + "-" + key[8:12] + "-" + key[12:16]
}

// HasFeature checks if the license includes a specific feature.
func (li *Info) HasFeature(feature string) bool {
	return slices.Contains(li.Features, feature)
}

// CanRunReflector returns true if the license allows reflector mode.
func (li *Info) CanRunReflector() bool {
	return li.Valid && li.Tier >= TierReflector
}

// CanRunTests returns true if the license allows the Professional
// test suite (or higher).
func (li *Info) CanRunTests() bool {
	return li.Valid && li.Tier >= TierProfessional
}

// proFeatures returns the feature list granted to TierProfessional.
// TierEnterprise (deprecated) now grants the same set per
// LICENSE_STRATEGY 2026-05-19. Listed alphabetically after reflector.
func proFeatures() []string {
	return []string{
		"reflector",
		"api",
		"mef",
		"multiuser",
		"rfc2544",
		"rfc2889",
		"rfc6349",
		"tsn",
		"y1564",
		"y1731",
	}
}
