// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"slices"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/license"
	"github.com/MustardSeedNetworks/stem/internal/services"
)

// TestEveryProFeatureIsRequiredByATestType is the catalog gate: a string in
// ProFeatures that no test type requires is a claim the binary never enforces
// — the drift docs/EDITIONS.md exists to prevent. Adding a catalog entry
// without a capability behind it fails here.
func TestEveryProFeatureIsRequiredByATestType(t *testing.T) {
	t.Parallel()

	required := map[string]string{}
	for _, mod := range services.GetAllModules() {
		for _, testType := range mod.TestTypes() {
			if feature, gated := api.FeatureForTestTypeForTest(testType); gated {
				required[feature] = testType
			}
		}
	}

	free := license.FreeFeatures()
	for _, feature := range license.ProFeatures() {
		if slices.Contains(free, feature) {
			continue
		}
		if _, ok := required[feature]; !ok {
			t.Errorf("Pro feature %q is granted by no gate: no registered test type requires it", feature)
		}
	}
}

// TestEveryTestTypeIsPriced is the reverse direction: a registered test type
// missing from the map runs unlicensed, silently giving away a paid
// capability. The Free reflector is the one deliberate exception.
func TestEveryTestTypeIsPriced(t *testing.T) {
	t.Parallel()

	pro := license.ProFeatures()
	for _, mod := range services.GetAllModules() {
		for _, testType := range mod.TestTypes() {
			feature, gated := api.FeatureForTestTypeForTest(testType)
			if !gated {
				if testType != "reflect" {
					t.Errorf("test type %q (module %s) requires no feature; only the Free reflector may be ungated",
						testType, mod.Name())
				}
				continue
			}
			if !slices.Contains(pro, feature) {
				t.Errorf("test type %q requires %q, which is not in the Pro catalog", testType, feature)
			}
		}
	}
}

// TestReflectorIsUngated pins the Free grant: reflection must never require an
// entitlement, or an unlicensed install stops being a usable reflector.
func TestReflectorIsUngated(t *testing.T) {
	t.Parallel()

	if feature, gated := api.FeatureForTestTypeForTest("reflect"); gated {
		t.Errorf("reflect is gated on %q; the Free tier is reflector-only and must stay ungated", feature)
	}
	if !slices.Contains(license.FreeFeatures(), license.FeatureReflector) {
		t.Error("FreeFeatures() no longer grants the reflector feature")
	}
}

// TestUnknownTestTypeIsUngated documents the boundary: an unregistered name
// reports no feature. handleTestStart rejects it as an unknown test type
// before the entitlement check, so this must not be read as "unknown is free".
func TestUnknownTestTypeIsUngated(t *testing.T) {
	t.Parallel()

	if _, gated := api.FeatureForTestTypeForTest("no_such_test"); gated {
		t.Error("unknown test type reported a feature requirement")
	}
}
