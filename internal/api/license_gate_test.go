// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/license"
)

// proLicense returns a manager in a temp directory with a Professional trial
// started. Setup helpers install it so a test's entitlement is what the test
// says, not whatever activation state the developer's ~/.config/stem holds.
func proLicense(t testing.TB) *license.Manager {
	t.Helper()
	mgr, _, err := license.LoadFromDir(t.TempDir())
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}
	if result := mgr.StartTrial(); !result.Success {
		t.Fatalf("StartTrial() failed: %s", result.Message)
	}
	return mgr
}

// unlicensed returns a manager in an empty temp directory: no key, no trial,
// which is what a fresh install answers with.
func unlicensed(t testing.TB) *license.Manager {
	t.Helper()
	mgr, _, err := license.LoadFromDir(t.TempDir())
	if err != nil {
		t.Fatalf("LoadFromDir() error: %v", err)
	}
	return mgr
}

// startTest posts a test-start request for testType and returns the recorder.
// The token is passed in rather than minted per call: the auth limiter is 5
// logins a minute, which a table of eight standards would exhaust.
func startTest(t *testing.T, s *api.Server, token, testType string) *httptest.ResponseRecorder {
	t.Helper()
	body := bytes.NewBufferString(
		fmt.Sprintf(`{"tests":[{"testType":%q}],"interface":"nonexistent_iface_xyz123"}`, testType),
	)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	return w
}

// TestUnlicensedTestStartIsPaymentRequired asserts the entitlement boundary:
// an unlicensed deployment cannot run a paid standard through the API. The
// interface is deliberately bogus — the gate must answer before anything is
// resolved, so a 400 here would mean the run got past the gate.
func TestUnlicensedTestStartIsPaymentRequired(t *testing.T) {
	s := setupTestingTestServer(t)
	s.UseLicenseForTest(unlicensed(t))
	token := getTestingAuthToken(t, s)

	for testType, feature := range map[string]string{
		"rfc2544_throughput": license.FeatureRFC2544,
		"y1564":              license.FeatureY1564,
		"mef":                license.FeatureMEF,
		"y1731_delay":        license.FeatureY1731,
		"rfc2889_forwarding": license.FeatureRFC2889,
		"rfc6349_throughput": license.FeatureRFC6349,
		"tsn_timing":         license.FeatureTSN,
		"custom_stream":      license.FeatureTrafficGen,
	} {
		t.Run(testType, func(t *testing.T) {
			w := startTest(t, s, token, testType)
			if w.Code != http.StatusPaymentRequired {
				t.Fatalf("status = %d, want 402: %s", w.Code, w.Body.String())
			}

			var resp api.FeatureGateResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode 402 body: %v", err)
			}
			if resp.RequiredFeature != feature {
				t.Errorf("requiredFeature = %q, want %q", resp.RequiredFeature, feature)
			}
			if resp.Code != "TIER_TOO_LOW" {
				t.Errorf("code = %q, want TIER_TOO_LOW", resp.Code)
			}
			if resp.CurrentTier != license.TierInvalid.String() {
				t.Errorf("currentTier = %q, want %q", resp.CurrentTier, license.TierInvalid.String())
			}
		})
	}
}

// TestTrialTestStartPassesTheGate asserts the trial grants the Pro catalog:
// the same request now reaches the interface check, so the 400 proves the gate
// let it through rather than that it never ran.
func TestTrialTestStartPassesTheGate(t *testing.T) {
	s := setupTestingTestServer(t)
	s.UseLicenseForTest(proLicense(t))

	w := startTest(t, s, getTestingAuthToken(t, s), "rfc2544_throughput")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (bogus interface, gate passed): %s", w.Code, w.Body.String())
	}
}

// TestUnlicensedTestStopIsNotGated pins the shape of the boundary: stopping a
// run and reading a result are not sold, only starting a paid standard is.
func TestUnlicensedTestStopIsNotGated(t *testing.T) {
	s := setupTestingTestServer(t)
	s.UseLicenseForTest(unlicensed(t))

	token := getTestingAuthToken(t, s)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/stop", nil)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code == http.StatusPaymentRequired {
		t.Error("test/stop answered 402; stopping a run is not a paid capability")
	}
}

// TestUnwiredLicenseDeniesPaidStandards is the API half of #1068: a manager
// that could not be built at all is not a licence. The paid standards are
// refused exactly as they are for an unlicensed host, so a deployment whose
// fingerprint cannot be computed runs as Free rather than as Pro.
func TestUnwiredLicenseDeniesPaidStandards(t *testing.T) {
	s := setupTestingTestServer(t)
	s.UseLicenseForTest(nil)

	w := startTest(t, s, getTestingAuthToken(t, s), "rfc2544_throughput")
	if w.Code != http.StatusPaymentRequired {
		t.Fatalf("status = %d, want 402: %s", w.Code, w.Body.String())
	}
}
