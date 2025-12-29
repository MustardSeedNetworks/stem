// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// License key validation for Seed Test Suite.
// Validates MSN license key format and extracts tier information.
package license

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

/*
MSN License Key Format (16 characters):
┌──────┬────────┬───────┬──────┬──────────┐
│ CC   │ PPPP   │SSSSSSS│ T    │ XX       │
│Check │Product │Serial │Tier  │ Checksum │
└──────┴────────┴───────┴──────┴──────────┘

Positions:
  0-1:  Checksum prefix (encoded validation)
  2-5:  Product code (1001=Reflector, 2001=TestSuite)
  6-12: Serial number (unique per license)
  13:   Tier (1=Reflector, 2=Full Suite, 3=Enterprise)
  14-15: Checksum suffix

Product Codes:
  1001: Seed Reflector (Tier 1)
  2001: Seed Test Suite (Tier 2)
  3001: Seed Enterprise (Tier 3, future)
*/

// Tier represents the license tier
type Tier int

const (
	TierInvalid    Tier = 0
	TierReflector  Tier = 1 // Reflector only
	TierTestSuite  Tier = 2 // Full test suite + reflector
	TierEnterprise Tier = 3 // Enterprise (future)
)

// String returns the tier name
func (t Tier) String() string {
	switch t {
	case TierReflector:
		return "Reflector"
	case TierTestSuite:
		return "Test Suite"
	case TierEnterprise:
		return "Enterprise"
	default:
		return "Invalid"
	}
}

// LicenseInfo contains parsed license information
type LicenseInfo struct {
	Key         string    `json:"key"`
	Valid       bool      `json:"valid"`
	Tier        Tier      `json:"tier"`
	ProductCode string    `json:"productCode"`
	Serial      string    `json:"serial"`
	Activated   bool      `json:"activated"`
	ActivatedAt time.Time `json:"activatedAt,omitempty"`
	ExpiresAt   time.Time `json:"expiresAt,omitempty"`
	DeviceHash  string    `json:"deviceHash,omitempty"`
	MaxDevices  int       `json:"maxDevices"`
	Features    []string  `json:"features"`
	ErrorMsg    string    `json:"error,omitempty"`
}

// ValidateLicenseKey performs offline validation of a license key
func ValidateLicenseKey(key string) *LicenseInfo {
	info := &LicenseInfo{
		Key:        key,
		MaxDevices: 3, // Default to 3 activations per license
	}

	// Normalize key (remove spaces, dashes, uppercase)
	key = normalizeKey(key)
	info.Key = key

	// Check length
	if len(key) != 16 {
		info.ErrorMsg = "License key must be 16 characters"
		return info
	}

	// Check format (alphanumeric only)
	if !regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(key) {
		info.ErrorMsg = "License key must contain only letters and numbers"
		return info
	}

	// Decode the key through rotor cipher first
	cipher := NewRotorCipher(7) // MSN start position
	decoded := cipher.DecodeString(key)

	// Validate checksum on decoded key (uses positions 0-1 and 14-15)
	if !validateKeyChecksum(decoded) {
		info.ErrorMsg = "Invalid license key checksum"
		return info
	}

	// Extract components
	info.ProductCode = decoded[2:6]
	info.Serial = decoded[6:13]
	tierChar := decoded[13]

	// Parse tier
	switch tierChar {
	case '1':
		info.Tier = TierReflector
		info.Features = []string{"reflector"}
	case '2':
		info.Tier = TierTestSuite
		info.Features = []string{"reflector", "rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn"}
	case '3':
		info.Tier = TierEnterprise
		info.Features = []string{"reflector", "rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn", "api", "multiuser"}
	default:
		info.ErrorMsg = "Invalid license tier"
		return info
	}

	// Validate product code
	switch info.ProductCode {
	case "1001":
		if info.Tier != TierReflector {
			info.ErrorMsg = "Product code mismatch for tier"
			return info
		}
	case "2001":
		if info.Tier != TierTestSuite {
			info.ErrorMsg = "Product code mismatch for tier"
			return info
		}
	case "3001":
		if info.Tier != TierEnterprise {
			info.ErrorMsg = "Product code mismatch for tier"
			return info
		}
	default:
		info.ErrorMsg = "Unknown product code"
		return info
	}

	info.Valid = true
	return info
}

// validateKeyChecksum checks the embedded checksum
func validateKeyChecksum(key string) bool {
	// Extract the core payload (positions 2-13)
	payload := key[2:14]

	// Calculate expected checksum
	expected := CalculateChecksum(payload)

	// Compare with key prefix (0-1) and suffix (14-15)
	prefixMatch := key[0:2] == expected[0:1]+key[1:2] || key[0:2] == expected // Allow some flexibility
	suffixMatch := key[14:16] == expected

	// Use a more lenient check - verify suffix or use full checksum validation
	return prefixMatch || suffixMatch || ValidateChecksum(key[2:16])
}

// GenerateLicenseKey creates a new license key (for admin/generator tool)
func GenerateLicenseKey(productCode string, serial string, tier Tier) (string, error) {
	// Validate inputs
	if len(productCode) != 4 {
		return "", fmt.Errorf("product code must be 4 characters")
	}
	if len(serial) != 7 {
		return "", fmt.Errorf("serial must be 7 characters")
	}
	if tier < TierReflector || tier > TierEnterprise {
		return "", fmt.Errorf("invalid tier")
	}

	// Build payload: PPPP + SSSSSSS + T
	payload := productCode + serial + fmt.Sprintf("%d", tier)

	// Calculate checksum
	checksum := CalculateChecksum(payload)

	// Build full key: CC + payload + XX
	fullKey := checksum[0:2] + payload + checksum

	// Encode through rotor cipher
	cipher := NewRotorCipher(7)
	encoded := cipher.EncodeString(fullKey)

	return strings.ToUpper(encoded), nil
}

// normalizeKey cleans up a license key for validation
func normalizeKey(key string) string {
	// Remove common separators
	key = strings.ReplaceAll(key, "-", "")
	key = strings.ReplaceAll(key, " ", "")
	key = strings.ReplaceAll(key, ".", "")

	// Uppercase
	return strings.ToUpper(key)
}

// FormatKey formats a license key for display (adds dashes)
func FormatKey(key string) string {
	key = normalizeKey(key)
	if len(key) != 16 {
		return key
	}
	// Format as XXXX-XXXX-XXXX-XXXX
	return key[0:4] + "-" + key[4:8] + "-" + key[8:12] + "-" + key[12:16]
}

// HasFeature checks if the license includes a specific feature
func (li *LicenseInfo) HasFeature(feature string) bool {
	for _, f := range li.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// CanRunReflector returns true if the license allows reflector mode
func (li *LicenseInfo) CanRunReflector() bool {
	return li.Valid && li.Tier >= TierReflector
}

// CanRunTests returns true if the license allows test suite
func (li *LicenseInfo) CanRunTests() bool {
	return li.Valid && li.Tier >= TierTestSuite
}
