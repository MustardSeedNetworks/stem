// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// MSN Rotor Cipher for license key validation.
// Uses Enigma-style rotor substitution for offline-capable validation.
package license

import (
	"strings"
)

// MSN-specific rotor tables (unique to Mustard Seed Networks products)
// These provide the substitution mapping for license key encoding/decoding
var (
	// Rotor for digits (0-9)
	msnRotor10 = [10]int{7, 2, 9, 0, 5, 8, 1, 6, 3, 4}

	// Rotor for uppercase letters (A-Z)
	msnRotor26 = [26]int{
		19, 3, 24, 7, 12, 0, 21, 15, 8, 25,
		2, 17, 10, 5, 22, 13, 1, 18, 6, 11,
		23, 4, 16, 9, 20, 14,
	}

	// Inverse rotors for decoding
	msnRotor10Inv [10]int
	msnRotor26Inv [26]int
)

func init() {
	// Generate inverse rotor tables
	for i, v := range msnRotor10 {
		msnRotor10Inv[v] = i
	}
	for i, v := range msnRotor26 {
		msnRotor26Inv[v] = i
	}
}

// RotorCipher provides Enigma-style encoding/decoding
type RotorCipher struct {
	position int // Current rotor position (advances with each character)
}

// NewRotorCipher creates a new cipher with the given starting position
func NewRotorCipher(startPosition int) *RotorCipher {
	return &RotorCipher{
		position: startPosition % 36, // 10 digits + 26 letters
	}
}

// Encode encodes a single character through the rotor
func (rc *RotorCipher) Encode(c byte) byte {
	if c >= '0' && c <= '9' {
		// Digit encoding
		idx := int(c - '0')
		idx = (idx + rc.position) % 10
		encoded := msnRotor10[idx]
		rc.position = (rc.position + 1) % 36
		return byte('0' + encoded)
	} else if c >= 'A' && c <= 'Z' {
		// Letter encoding
		idx := int(c - 'A')
		idx = (idx + rc.position) % 26
		encoded := msnRotor26[idx]
		rc.position = (rc.position + 1) % 36
		return byte('A' + encoded)
	} else if c >= 'a' && c <= 'z' {
		// Lowercase - convert to upper, encode, keep case
		idx := int(c - 'a')
		idx = (idx + rc.position) % 26
		encoded := msnRotor26[idx]
		rc.position = (rc.position + 1) % 36
		return byte('a' + encoded)
	}
	// Non-alphanumeric passes through unchanged
	return c
}

// Decode decodes a single character through the inverse rotor
func (rc *RotorCipher) Decode(c byte) byte {
	if c >= '0' && c <= '9' {
		// Digit decoding
		idx := msnRotor10Inv[int(c-'0')]
		idx = (idx - rc.position%10 + 10) % 10
		rc.position = (rc.position + 1) % 36
		return byte('0' + idx)
	} else if c >= 'A' && c <= 'Z' {
		// Letter decoding
		idx := msnRotor26Inv[int(c-'A')]
		idx = (idx - rc.position%26 + 26) % 26
		rc.position = (rc.position + 1) % 36
		return byte('A' + idx)
	} else if c >= 'a' && c <= 'z' {
		// Lowercase decoding
		idx := msnRotor26Inv[int(c-'a')]
		idx = (idx - rc.position%26 + 26) % 26
		rc.position = (rc.position + 1) % 36
		return byte('a' + idx)
	}
	return c
}

// EncodeString encodes an entire string
func (rc *RotorCipher) EncodeString(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		result[i] = rc.Encode(s[i])
	}
	return string(result)
}

// DecodeString decodes an entire string
func (rc *RotorCipher) DecodeString(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		result[i] = rc.Decode(s[i])
	}
	return string(result)
}

// CalculateChecksum generates a 2-character checksum for a string
// Uses a simple polynomial hash that's quick to compute offline
func CalculateChecksum(s string) string {
	s = strings.ToUpper(s)
	var sum1, sum2 int

	for i, c := range s {
		val := 0
		if c >= '0' && c <= '9' {
			val = int(c - '0')
		} else if c >= 'A' && c <= 'Z' {
			val = int(c-'A') + 10
		}

		// Polynomial accumulation with position weighting
		sum1 = (sum1 + val*(i+1)) % 36
		sum2 = (sum2 + val*val + i) % 36
	}

	// Convert to alphanumeric
	c1 := toAlphanumeric(sum1)
	c2 := toAlphanumeric(sum2)

	return string([]byte{c1, c2})
}

// ValidateChecksum checks if the last 2 characters are a valid checksum
func ValidateChecksum(s string) bool {
	if len(s) < 3 {
		return false
	}
	s = strings.ToUpper(s)
	payload := s[:len(s)-2]
	checksum := s[len(s)-2:]
	return CalculateChecksum(payload) == checksum
}

// toAlphanumeric converts a value 0-35 to 0-9 or A-Z
func toAlphanumeric(val int) byte {
	if val < 10 {
		return byte('0' + val)
	}
	return byte('A' + val - 10)
}

// fromAlphanumeric converts 0-9 or A-Z to a value 0-35.
// Currently unused but kept for future license parsing.
func fromAlphanumeric(c byte) int { //nolint:unused
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'A' && c <= 'Z' {
		return int(c-'A') + 10
	}
	if c >= 'a' && c <= 'z' {
		return int(c-'a') + 10
	}
	return 0
}
