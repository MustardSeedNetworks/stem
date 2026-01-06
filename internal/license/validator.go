// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

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
  2-5:  Product code (1001=Reflector, 2001=TestSuite).
  6-12: Serial number (unique per license).
  13:   Tier (1=Reflector, 2=Full Suite, 3=Enterprise).
  14-15: Checksum suffix.

Product Codes:
  1001: Seed Reflector (Tier 1).
  2001: Seed Test Suite (Tier 2).
  3001: Seed Enterprise (Tier 3, future).
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
	// TierTestSuite provides full test suite plus reflector.
	TierTestSuite Tier = 2
	// TierEnterprise provides enterprise features (future).
	TierEnterprise Tier = 3
)

// Error messages.
const errProductCodeMismatch = "Product code mismatch for tier"

// String returns the tier name.
func (t Tier) String() string {
	switch t {
	case TierInvalid:
		return "Invalid"
	case TierReflector:
		return "Reflector"
	case TierTestSuite:
		return "Test Suite"
	case TierEnterprise:
		return "Enterprise"
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
		info.ErrorMsg = "License key must be 16 characters"
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
		info.Tier = TierTestSuite
		info.Features = []string{"reflector", "rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn"}
	case '3':
		info.Tier = TierEnterprise
		info.Features = []string{
			"reflector", "rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn", "api", "multiuser",
		}
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
		if info.Tier != TierTestSuite {
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

// CanRunTests returns true if the license allows test suite.
func (li *Info) CanRunTests() bool {
	return li.Valid && li.Tier >= TierTestSuite
}
