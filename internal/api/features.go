// SPDX-License-Identifier: BUSL-1.1

package api

import "github.com/MustardSeedNetworks/stem/internal/license"

// featureForTestType returns the catalog feature a test type is sold under and
// whether running it requires an entitlement at all. Entitlement lives at the
// API layer, not inside the modules (depguard "service-layer-inward-only"), so
// the map that prices a capability lives here next to the gate that reads it.
//
// Reflection is the Free grant, so it reports ("", false) and runs unlicensed;
// every other registered test type maps to a Pro feature name from
// [license.ProFeatures].
//
// This map is the single place a capability is priced. features_test.go fails
// if a registered test type is missing from it, or if a Pro catalog string is
// not required by any test type — the two ways a catalog entry becomes a claim
// nobody enforces.
func featureForTestType(testType string) (string, bool) {
	feature, ok := featuresByTestType()[testType]
	if !ok || feature == "" {
		return "", false
	}
	return feature, true
}

// featuresByTestType maps every registered test type to the catalog feature it
// needs. The empty string means "no entitlement required" (the Free reflector).
func featuresByTestType() map[string]string {
	return map[string]string{
		// Reflector — Free, no entitlement required.
		"reflect": "",

		// RFC 2544 benchmarking.
		"rfc2544_throughput":      license.FeatureRFC2544,
		"rfc2544_latency":         license.FeatureRFC2544,
		"rfc2544_frame_loss":      license.FeatureRFC2544,
		"rfc2544_back_to_back":    license.FeatureRFC2544,
		"rfc2544_system_recovery": license.FeatureRFC2544,
		"rfc2544_reset":           license.FeatureRFC2544,

		// Y.1564 service activation and the MEF service tests that share
		// the servicetest module.
		"y1564_config": license.FeatureY1564,
		"y1564_perf":   license.FeatureY1564,
		"y1564":        license.FeatureY1564,
		"mef_config":   license.FeatureMEF,
		"mef_perf":     license.FeatureMEF,
		"mef":          license.FeatureMEF,

		// Y.1731 performance monitoring.
		"y1731_delay":    license.FeatureY1731,
		"y1731_loss":     license.FeatureY1731,
		"y1731_slm":      license.FeatureY1731,
		"y1731_loopback": license.FeatureY1731,

		// RFC 2889 / RFC 6349 / TSN, all three in the certify module.
		"rfc2889_forwarding": license.FeatureRFC2889,
		"rfc2889_caching":    license.FeatureRFC2889,
		"rfc2889_learning":   license.FeatureRFC2889,
		"rfc2889_broadcast":  license.FeatureRFC2889,
		"rfc2889_congestion": license.FeatureRFC2889,
		"rfc6349_throughput": license.FeatureRFC6349,
		"rfc6349_path":       license.FeatureRFC6349,
		"tsn_timing":         license.FeatureTSN,
		"tsn_isolation":      license.FeatureTSN,
		"tsn_latency":        license.FeatureTSN,
		"tsn":                license.FeatureTSN,

		// Custom traffic generation.
		"custom_stream": license.FeatureTrafficGen,
	}
}
