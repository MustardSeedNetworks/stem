// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package license

import (
	"strings"
)

// Rotor constants for cipher operations.
const (
	rotorModulus      = 36 // 10 digits + 26 letters.
	digitModulus      = 10
	letterModulus     = 26
	alphanumericSplit = 10 // Values 0-9 are digits, 10-35 are letters.
)

// MSN-specific rotor tables (unique to Mustard Seed Networks products).
// These provide the substitution mapping for license key encoding/decoding.
//
//nolint:gochecknoglobals // Static cipher lookup tables required for license encoding.
var (
	// Rotor for digits (0-9).
	msnRotor10 = [10]int{7, 2, 9, 0, 5, 8, 1, 6, 3, 4}

	// Rotor for uppercase letters (A-Z).
	msnRotor26 = [26]int{
		19, 3, 24, 7, 12, 0, 21, 15, 8, 25,
		2, 17, 10, 5, 22, 13, 1, 18, 6, 11,
		23, 4, 16, 9, 20, 14,
	}

	// Inverse rotors for decoding.
	msnRotor10Inv [10]int
	msnRotor26Inv [26]int
)

// InitRotors initializes the inverse rotor tables for decoding.
// This must be called before using the cipher.
func InitRotors() {
	// Generate inverse rotor tables.
	for i, v := range msnRotor10 {
		msnRotor10Inv[v] = i
	}
	for i, v := range msnRotor26 {
		msnRotor26Inv[v] = i
	}
}

//nolint:gochecknoinits // Required to initialize cipher lookup tables at startup.
func init() {
	InitRotors()
}

// RotorCipher provides Enigma-style encoding/decoding.
type RotorCipher struct {
	position int // Current rotor position (advances with each character).
}

// NewRotorCipher creates a new cipher with the given starting position.
func NewRotorCipher(startPosition int) *RotorCipher {
	return &RotorCipher{
		position: startPosition % rotorModulus,
	}
}

// Encode encodes a single character through the rotor.
func (rc *RotorCipher) Encode(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		// Digit encoding.
		idx := int(c - '0')
		idx = (idx + rc.position) % digitModulus
		encoded := msnRotor10[idx]
		rc.position = (rc.position + 1) % rotorModulus
		return byte('0' + encoded)
	case c >= 'A' && c <= 'Z':
		// Letter encoding.
		idx := int(c - 'A')
		idx = (idx + rc.position) % letterModulus
		encoded := msnRotor26[idx]
		rc.position = (rc.position + 1) % rotorModulus
		return byte('A' + encoded)
	case c >= 'a' && c <= 'z':
		// Lowercase - convert to upper, encode, keep case.
		idx := int(c - 'a')
		idx = (idx + rc.position) % letterModulus
		encoded := msnRotor26[idx]
		rc.position = (rc.position + 1) % rotorModulus
		return byte('a' + encoded)
	default:
		// Non-alphanumeric passes through unchanged.
		return c
	}
}

// Decode decodes a single character through the inverse rotor.
func (rc *RotorCipher) Decode(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		// Digit decoding.
		idx := msnRotor10Inv[int(c-'0')]
		idx = (idx - rc.position%digitModulus + digitModulus) % digitModulus
		rc.position = (rc.position + 1) % rotorModulus
		return byte('0' + idx)
	case c >= 'A' && c <= 'Z':
		// Letter decoding.
		idx := msnRotor26Inv[int(c-'A')]
		idx = (idx - rc.position%letterModulus + letterModulus) % letterModulus
		rc.position = (rc.position + 1) % rotorModulus
		return byte('A' + idx)
	case c >= 'a' && c <= 'z':
		// Lowercase decoding.
		idx := msnRotor26Inv[int(c-'a')]
		idx = (idx - rc.position%letterModulus + letterModulus) % letterModulus
		rc.position = (rc.position + 1) % rotorModulus
		return byte('a' + idx)
	default:
		return c
	}
}

// EncodeString encodes an entire string.
func (rc *RotorCipher) EncodeString(s string) string {
	result := make([]byte, len(s))
	for i := range len(s) {
		result[i] = rc.Encode(s[i])
	}
	return string(result)
}

// DecodeString decodes an entire string.
func (rc *RotorCipher) DecodeString(s string) string {
	result := make([]byte, len(s))
	for i := range len(s) {
		result[i] = rc.Decode(s[i])
	}
	return string(result)
}

// CalculateChecksum generates a 2-character checksum for a string.
// Uses a simple polynomial hash that's quick to compute offline.
func CalculateChecksum(s string) string {
	s = strings.ToUpper(s)
	var sum1, sum2 int

	for i, c := range s {
		val := 0
		if c >= '0' && c <= '9' {
			val = int(c - '0')
		} else if c >= 'A' && c <= 'Z' {
			val = int(c-'A') + alphanumericSplit
		}

		// Polynomial accumulation with position weighting.
		sum1 = (sum1 + val*(i+1)) % rotorModulus
		sum2 = (sum2 + val*val + i) % rotorModulus
	}

	// Convert to alphanumeric.
	c1 := toAlphanumeric(sum1)
	c2 := toAlphanumeric(sum2)

	return string([]byte{c1, c2})
}

// ValidateChecksum checks if the last 2 characters are a valid checksum.
func ValidateChecksum(s string) bool {
	const minLength = 3
	if len(s) < minLength {
		return false
	}
	s = strings.ToUpper(s)
	payload := s[:len(s)-2]
	checksum := s[len(s)-2:]
	return CalculateChecksum(payload) == checksum
}

// toAlphanumeric converts a value 0-35 to 0-9 or A-Z.
func toAlphanumeric(val int) byte {
	if val < alphanumericSplit {
		return byte('0' + val)
	}
	return byte('A' + val - alphanumericSplit)
}

// fromAlphanumeric converts 0-9 or A-Z to a value 0-35.
// Currently unused but kept for future license parsing.
//
//nolint:unused // Reserved for future license parsing functionality.
func fromAlphanumeric(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'A' && c <= 'Z' {
		return int(c-'A') + alphanumericSplit
	}
	if c >= 'a' && c <= 'z' {
		return int(c-'a') + alphanumericSplit
	}
	return 0
}
