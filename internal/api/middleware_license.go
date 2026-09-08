// SPDX-License-Identifier: BUSL-1.1

package api

// middleware_license.go is Stem's entitlement boundary. Stem prices whole test
// standards rather than whole routes — every test starts at POST
// /api/v1/test/start and names its type in the body — so the check lives at the
// point the test type is resolved rather than at route registration. The
// decision itself is one predicate here so the API cannot disagree with the
// catalog in internal/license.

import (
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/license"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// errCodeTierTooLow is the error code a client keys off to render an upgrade
// hint rather than a generic failure.
const errCodeTierTooLow = "TIER_TOO_LOW"

// FeatureGateResponse is the 402 body returned when a request asks for a
// capability the active license does not cover.
type FeatureGateResponse struct {
	Error           string `json:"error"`
	Code            string `json:"code"`
	RequiredFeature string `json:"requiredFeature"`
	CurrentTier     string `json:"currentTier"`
	UpgradeMessage  string `json:"upgradeMessage"`
}

// hasFeature reports whether the active license covers feature.
//
// A nil license manager means licensing is not wired at all (a developer build
// whose manager failed to load, or a test that has not installed one) and
// permits everything, so a broken licence file degrades to an unenforced
// binary rather than an unusable one.
func (s *Server) hasFeature(feature string) bool {
	return s.licenseManager == nil || s.licenseManager.HasFeature(feature)
}

// sendFeatureGate writes the 402 an unlicensed capability answers with, and
// logs the denial under a single event so an operator can see which feature a
// deployment is being asked for.
func (s *Server) sendFeatureGate(w http.ResponseWriter, feature string) {
	tierName := license.TierInvalid.String()
	if s.licenseManager != nil {
		if state := s.licenseManager.GetState(); state != nil {
			tierName = license.Tier(state.Tier).String()
		}
	}

	logging.Warn("feature gate denied request",
		"event", "license.forbidden",
		"feature", feature,
		"tier", tierName,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	writeJSON(w, FeatureGateResponse{
		Error:           "Feature requires a higher tier",
		Code:            errCodeTierTooLow,
		RequiredFeature: feature,
		CurrentTier:     tierName,
		UpgradeMessage: "Start a 14-day Pro trial with `stem license --trial` " +
			"or activate a Pro key with `stem license --activate <KEY>`.",
	})
}
