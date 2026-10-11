// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

// setupCapabilitiesTestServer mirrors setupHealthTestServer — same
// auth env vars, same Shutdown cleanup, no rate limiter pre-warming
// (the capabilities route is unrate-limited).
func setupCapabilitiesTestServer(t testing.TB) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", "capstest")
	t.Setenv("STEM_AUTH_PASSWORD", "capspass123")

	s, err := api.NewServer(api.ListenAddr{Port: 8444})
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// TestHandleCapabilities_ReportsTheDataplaneForBothRoles pins stem#1650:
// test-master support used to be hard-coded true, so a darwin daemon claimed
// it while every test failed in the dataplane. Both roles now report the same
// probe, with its reason when unsupported.
func TestHandleCapabilities_ReportsTheDataplaneForBothRoles(t *testing.T) {
	tests := []struct {
		name      string
		available bool
		reason    string
		want      api.CapabilityInfo
	}{
		{"dataplane present", true, "", api.CapabilityInfo{Supported: true}},
		{
			"dataplane absent", false, "CGO + Linux required",
			api.CapabilityInfo{Supported: false, Reason: "CGO + Linux required"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupCapabilitiesTestServer(t)
			s.UseDataplaneAvailabilityForTest(func() (bool, string) { return tt.available, tt.reason })

			req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d body=%q", w.Code, w.Body.String())
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("expected Content-Type application/json, got %q", got)
			}
			var resp api.CapabilitiesResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode response: %v body=%q", err, w.Body.String())
			}
			if resp.Reflector != tt.want || resp.TestMaster != tt.want {
				t.Errorf("capabilities = %+v, want both roles %+v", resp, tt.want)
			}
		})
	}
}

// TestHandleCapabilities_MethodNotAllowed verifies non-GET requests
// are rejected, matching the /__version contract.
func TestHandleCapabilities_MethodNotAllowed(t *testing.T) {
	s := setupCapabilitiesTestServer(t)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			req := httptest.NewRequest(m, "/api/v1/capabilities", nil)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)
			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected 405 for %s, got %d", m, w.Code)
			}
			if got := w.Header().Get("Allow"); got != http.MethodGet {
				t.Errorf("expected Allow header GET, got %q", got)
			}
		})
	}
}

// TestHandleCapabilities_NoAuth verifies the endpoint is reachable
// without credentials — the UI calls it before login completes so it
// must not 401.
func TestHandleCapabilities_NoAuth(t *testing.T) {
	s := setupCapabilitiesTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Errorf("capabilities endpoint must not require auth, got 401")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%q", w.Code, w.Body.String())
	}
}
