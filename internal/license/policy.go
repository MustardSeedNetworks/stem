// SPDX-License-Identifier: BUSL-1.1

// Package license is Stem's thin product layer over the fleet-shared
// github.com/MustardSeedNetworks/foundation/pkg/license core. The generic
// signing format, device fingerprint, activation state machine and encrypted
// on-disk state all live in foundation; this package supplies only Stem's
// product policy (tier vocabulary, product codes, feature catalog, trial terms)
// and re-exports the manager surface its callers use. See ADR-0007.
package license

import fnd "github.com/MustardSeedNetworks/foundation/pkg/license"

// Stem licenses are Ed25519-signed tokens verified offline against an embedded
// public key (the matching private key lives only in the keygen tool). Product
// codes: 1001 = Reflector (tier 1), 2001 = Professional (tier 2). Enterprise
// (code 3001 / tier 3) was retired 2026-06-16 — a token presenting it is
// rejected, no grandfathering (pre-v1).
const (
	// productName identifies this binary in a signed payload. A token issued
	// for another product (niac/seed) is rejected even if correctly signed.
	productName = "stem"

	// Product codes accepted by Stem.
	codeReflector    = "1001"
	codeProfessional = "2001"

	// defaultMaxDevices is the activation cap assumed until a validated token
	// specifies its own MaxDevices.
	defaultMaxDevices = 3

	// TrialDays is the trial-period length before a paid key is required.
	// Exported for the CLI's "days left of N" display.
	TrialDays = 14

	// encryptionSalt is product-distinct so a license file from another
	// product can't be reused by renaming it into Stem's config directory.
	// (Corrects the pre-migration "MSN-SEED-2024-LICENSE" fork artifact.)
	encryptionSalt = "MSN-STEM-2026-LICENSE"

	// configSubdir is the directory under ~/.config where activation state is
	// persisted (~/.config/stem); licenseFileName is the file within it.
	// (Corrects the pre-migration "seed-test-suite"/".seed-license" fork
	// artifacts.)
	configSubdir    = "stem"
	licenseFileName = ".license"
)

// Tier represents the license tier.
type Tier int

// License tier constants. Wire values match the tier field in the signed
// payload; tier 0 is the implicit unlicensed/invalid tier.
const (
	// TierInvalid represents an invalid or unrecognized (or unlicensed) tier.
	TierInvalid Tier = 0
	// TierReflector provides reflector-only functionality. Wire value 1.
	TierReflector Tier = 1
	// TierProfessional provides the full Stem test suite (RFC 2544 / Y.1564 /
	// Y.1731 / RFC 2889 / RFC 6349 / MEF / TSN) plus the reflector and API
	// access. Wire value 2.
	TierProfessional Tier = 2
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

// featuresForTier maps a signed wire-tier to the features Stem grants and the
// product code expected for that tier. Only Reflector/Professional carry a
// token; every other tier (including the retired Enterprise tier 3 and the
// unlicensed tier 0) is rejected so a signed token can only grant what this
// build knows about. Passed to foundation as ProductPolicy.FeaturesForTier.
func featuresForTier(wireTier int) ([]string, string, bool) {
	switch Tier(wireTier) {
	case TierReflector:
		return []string{"reflector"}, codeReflector, true
	case TierProfessional:
		return proFeatures(), codeProfessional, true
	case TierInvalid:
		return nil, "", false
	default:
		return nil, "", false
	}
}

// Policy is Stem's product configuration for the foundation license core.
func Policy() fnd.ProductPolicy {
	return fnd.ProductPolicy{
		ProductName:       productName,
		FeaturesForTier:   featuresForTier,
		EncryptionSalt:    encryptionSalt,
		ConfigSubdir:      configSubdir,
		LicenseFileName:   licenseFileName,
		DefaultMaxDevices: defaultMaxDevices,
		TrialDays:         TrialDays,
		TrialTier:         int(TierProfessional),
	}
}
