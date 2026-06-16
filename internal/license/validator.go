// SPDX-License-Identifier: BUSL-1.1

package license

import (
	"slices"
	"strings"
	"time"
)

/*
Stem licenses are Ed25519-signed tokens (see signing.go). The previous 16-char
rotor-cipher key format was removed: its generator shipped inside the binary, so
any copy of Stem could mint a valid key. Tokens are now signed by the keygen
tool's private key and verified here with an embedded public key — offline and
un-forgeable.

Product codes:
  1001: Stem Reflector  (tier 1)
  2001: Stem Professional (tier 2). Was "Test Suite"; renamed per
        LICENSE_STRATEGY 2026-05-19. Wire value preserved.

  Enterprise (code 3001 / tier 3) was retired entirely on 2026-06-16. It is
  no longer a SKU, the keygen no longer mints it, and a token presenting
  tier 3 or product code 3001 is now REJECTED — no grandfathering (pre-v1).
*/

const (
	defaultMaxDevices = 3 // default activations per license

	// productName identifies this binary in a signed payload. A token issued
	// for another product (niac/seed) is rejected even if correctly signed.
	productName = "stem"
)

// Product codes accepted by Stem.
const (
	codeReflector    = "1001"
	codeProfessional = "2001"
)

// licensePublicKeyB64 is the standard-base64 Ed25519 public key that verifies
// production license tokens. The matching private key lives only in the keygen
// tool (msn-internal-tools/keygen) and never ships. See ADR-0007.
//
// Pre-launch signing key — rotate via keygen before GA.
const licensePublicKeyB64 = "O+o8n4qHHp/X//JrRXSdgGSWa2Fqz79OtgUkcylNxZg="

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
)

// Error messages.
const (
	errProductCodeMismatch = "Product code mismatch for tier"
	// ErrLicenseInvalid is the generic rejection message. Validation failures
	// deliberately do not distinguish "bad signature" from "tampered payload"
	// to a caller — both mean the same thing: not a genuine license.
	ErrLicenseInvalid = "License key is not valid"
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

// ValidateLicenseKey performs offline validation of a license token against the
// embedded production key. The verifier wraps a 32-byte key, so it is rebuilt
// per call rather than held as a package global; validation is not on a hot
// path (feature checks read cached Info, they do not re-validate).
func ValidateLicenseKey(key string) *Info {
	return mustVerifierFromB64(licensePublicKeyB64).Validate(key)
}

// Validate verifies a signed token and maps it to product feature data. The
// signature is checked first (in parseAndVerify); only a genuinely signed,
// current-version payload reaches the product-specific interpretation below.
func (v *Verifier) Validate(key string) *Info {
	info := &Info{
		Key:        strings.TrimSpace(key),
		Valid:      false,
		Tier:       TierInvalid,
		MaxDevices: defaultMaxDevices,
	}

	payload, err := v.parseAndVerify(key)
	if err != nil {
		info.ErrorMsg = ErrLicenseInvalid
		return info
	}

	// A correctly signed token for a different product must not validate here.
	if payload.Product != productName {
		info.ErrorMsg = ErrLicenseInvalid
		return info
	}

	info.ProductCode = payload.Code
	info.Serial = payload.Serial

	// Tier and feature set are authoritative in-binary: the payload's tier is
	// mapped to the feature list defined here, so a signed token can only grant
	// what this build knows about.
	switch payload.Tier {
	case int(TierReflector):
		info.Tier = TierReflector
		info.Features = []string{"reflector"}
	case int(TierProfessional):
		info.Tier = TierProfessional
		info.Features = proFeatures()
	default:
		info.ErrorMsg = "Invalid license tier"
		return info
	}

	if !productCodeMatchesTier(payload.Code, info.Tier) {
		info.ErrorMsg = errProductCodeMismatch
		return info
	}

	if payload.MaxDevices > 0 {
		info.MaxDevices = payload.MaxDevices
	}
	if payload.ExpiresAt > 0 {
		info.ExpiresAt = time.Unix(payload.ExpiresAt, 0).UTC()
		if time.Now().After(info.ExpiresAt) {
			info.ErrorMsg = "License has expired"
			return info
		}
	}

	info.Valid = true
	return info
}

// productCodeMatchesTier enforces that the product code embedded in the payload
// is the one expected for the tier, so a token cannot claim a code/tier pairing
// the catalog never issued.
func productCodeMatchesTier(code string, tier Tier) bool {
	switch code {
	case codeReflector:
		return tier == TierReflector
	case codeProfessional:
		return tier == TierProfessional
	default:
		return false
	}
}

// FormatKey returns a signed token for display. Tokens are already
// display-ready (single line, copy/paste); only surrounding whitespace is
// trimmed. Unlike the old 16-char format, tokens must NOT have characters
// stripped — base64url uses '-' and '_'.
func FormatKey(key string) string {
	return strings.TrimSpace(key)
}

// HasFeature checks if the license includes a specific feature.
func (li *Info) HasFeature(feature string) bool {
	return slices.Contains(li.Features, feature)
}

// CanRunReflector returns true if the license allows reflector mode.
func (li *Info) CanRunReflector() bool {
	return li.Valid && li.Tier >= TierReflector
}

// CanRunTests returns true if the license allows the Professional test suite
// (or higher).
func (li *Info) CanRunTests() bool {
	return li.Valid && li.Tier >= TierProfessional
}

// proFeatures returns the feature list granted to TierProfessional.
// Listed alphabetically after reflector.
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
