// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

const unknownTestTypeMessage = "Unknown or unsupported test type"

// startTestRaw posts a request body verbatim.
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

func TestRunPlanRequiresAtLeastOneStep(t *testing.T) {
	s := setupTestingTestServer(t)
	w := startTestRaw(t, s, getTestingAuthToken(t, s), `{"interface":"nonexistent_iface_xyz123"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
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
