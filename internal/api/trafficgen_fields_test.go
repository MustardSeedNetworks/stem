// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"fmt"
	"net/http"
	"testing"
)

// The generator has no burst pacing, VLAN tagging or MAC override (stem#1744).
// A Custom Stream that asks for one must be refused, not run untagged and
// unpaced while the request reads as though it was honoured.
func TestCustomStreamRefusesSettingsTheGeneratorCannotApply(t *testing.T) {
	cases := []struct {
		field string
		value string
	}{
		{"burstMode", "true"},
		{"burstSize", "500"},
		{"interBurstGapUs", "250"},
		{"srcMac", `"00:11:22:33:44:55"`},
		{"dstMac", `"66:77:88:99:aa:bb"`},
		{"vlanId", "100"},
		{"vlanPriority", "5"},
	}

	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			s := setupTestingTestServer(t)
			s.UseLicenseForTest(proLicense(t))

			body := fmt.Sprintf(
				`{"interface":"lo","peer":"192.0.2.1","tests":[{"testType":"custom_stream",`+
					`"config":{"trafficGen":{"frameSize":128,%q:%s}}}]}`,
				tc.field, tc.value)
			w := startTestRaw(t, s, getTestingAuthToken(t, s), body)

			// The strict decoder refuses the field before any other check runs.
			if w.Code != http.StatusBadRequest || responseMessage(t, w) != "Invalid JSON in request body" {
				t.Fatalf("status = %d, want the strict decoder's 400: %s", w.Code, w.Body.String())
			}
		})
	}
}
