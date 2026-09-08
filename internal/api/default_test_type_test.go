// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/license"
)

const unknownTestTypeMessage = "Unknown or unsupported test type"

// startTestRaw posts body verbatim so a test can omit testType entirely,
// which the typed helpers cannot express.
func startTestRaw(t *testing.T, s *api.Server, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", bytes.NewBufferString(body))
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	return w
}

// responseMessage returns the message field of an error body, or "".
func responseMessage(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		return ""
	}
	msg, _ := resp["message"].(string)
	return msg
}

// TestOmittedTestTypeResolvesToARegisteredModule is #1069: the default the
// handler substitutes for an omitted testType must name a test type some
// module actually registers, or the default is a guaranteed 400. The
// entitlement answer is the proof: an unlicensed host asked for the default
// gets the 402 for the standard the default belongs to, which can only happen
// once the type resolved.
func TestOmittedTestTypeResolvesToARegisteredModule(t *testing.T) {
	s := setupTestingTestServer(t)
	s.UseLicenseForTest(unlicensed(t))

	w := startTestRaw(t, s, getTestingAuthToken(t, s), `{"interface":"nonexistent_iface_xyz123"}`)
	if w.Code != http.StatusPaymentRequired {
		t.Fatalf("status = %d, want 402 (%s): %s", w.Code, unknownTestTypeMessage, w.Body.String())
	}

	var resp api.FeatureGateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode 402 body: %v", err)
	}
	if resp.RequiredFeature != license.FeatureRFC2544 {
		t.Errorf("requiredFeature = %q, want %q", resp.RequiredFeature, license.FeatureRFC2544)
	}
}

// TestOmittedTestTypeStartsUnderAProLicence is the other half: with the
// standard licensed, the default gets past the gate and fails only on the
// bogus interface — never on its own name.
func TestOmittedTestTypeStartsUnderAProLicence(t *testing.T) {
	s := setupTestingTestServer(t)
	s.UseLicenseForTest(proLicense(t))

	w := startTestRaw(t, s, getTestingAuthToken(t, s), `{"interface":"nonexistent_iface_xyz123"}`)
	if msg := responseMessage(t, w); msg == unknownTestTypeMessage {
		t.Fatalf("default test type is not registered by any module: %s", w.Body.String())
	}
}

// TestBareThroughputIsNotAnAPITestType pins the vocabulary the API speaks.
// "throughput" is the CLI's name for the RFC 2544 throughput test; the module
// registry indexes the standard-qualified names, and nothing registers the
// bare one. A request that sends it gets told so rather than silently
// resolving to something else.
func TestBareThroughputIsNotAnAPITestType(t *testing.T) {
	s := setupTestingTestServer(t)
	s.UseLicenseForTest(proLicense(t))

	w := startTest(t, s, getTestingAuthToken(t, s), "throughput")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
	}
	if msg := responseMessage(t, w); msg != unknownTestTypeMessage {
		t.Errorf("message = %q, want %q", msg, unknownTestTypeMessage)
	}
}
