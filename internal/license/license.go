// SPDX-License-Identifier: BUSL-1.1

package license

import fnd "github.com/MustardSeedNetworks/foundation/pkg/license"

// The manager, activation state and parsed-license types are the foundation
// core's, aliased so Stem callers keep referring to them as license.Manager
// etc. Their Tier fields are plain ints on the wire; convert with license.Tier
// when a Stem tier name or comparison is needed (see Tier.String).
type (
	// Manager owns Stem's activation state, trial and device binding.
	Manager = fnd.Manager
	// ActivationState is the persisted activation snapshot GetState returns.
	ActivationState = fnd.ActivationState
	// ActivationResult is returned by Activate/StartTrial/CheckIn.
	ActivationResult = fnd.ActivationResult
	// Info is a parsed, validated license token.
	Info = fnd.Info
	// DeviceFingerprint identifies the host a license is bound to.
	DeviceFingerprint = fnd.DeviceFingerprint
)

// FormatKey returns a signed token trimmed for display.
func FormatKey(key string) string {
	return fnd.FormatKey(key)
}
