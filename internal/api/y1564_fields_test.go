// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"fmt"
	"net/http"
	"testing"
)

// The Y.1564 generator has no VLAN tagging and no colour-aware metering, and
// without a tag there is no PCP to set (stem#1753). A run that asks for any of
// them must be refused, not sent untagged while the request reads as honoured.
func TestY1564RefusesSettingsTheGeneratorCannotApply(t *testing.T) {
	cases := []struct {
		field string
		value string
	}{
		{"vlanId", "100"},
		{"pcp", "5"},
		{"colorAware", "true"},
	}

	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			s := setupTestingTestServer(t)
			s.UseLicenseForTest(proLicense(t))

			body := fmt.Sprintf(
				`{"interface":"lo","peer":"192.0.2.1","tests":[{"testType":"y1564_config",`+
					`"config":{"y1564":{"cir":100,%q:%s}}}]}`,
				tc.field, tc.value)
			w := startTestRaw(t, s, getTestingAuthToken(t, s), body)

			// The strict decoder refuses the field before any other check runs.
			if w.Code != http.StatusBadRequest || responseMessage(t, w) != "Invalid JSON in request body" {
				t.Fatalf("status = %d, want the strict decoder's 400: %s", w.Code, w.Body.String())
			}
		})
	}
}
