// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/krisarmstrong/stem/internal/api"
)

// setupCapabilitiesTestServer mirrors setupHealthTestServer — same
// auth env vars, same Shutdown cleanup, no rate limiter pre-warming
// (the capabilities route is unrate-limited).
func setupCapabilitiesTestServer(t testing.TB) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_AUTH_USERNAME", "capstest")
	t.Setenv("STEM_AUTH_PASSWORD", "capspass123")

	s, err := api.NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// TestHandleCapabilities_GetReturnsBothFlags verifies the endpoint
// always returns reflector + testMaster blocks with a boolean
// Supported field, regardless of platform. Exact reflector.supported
// value is platform-dependent (linux+cgo = true, everything else =
// false), so the test asserts a shape invariant plus a per-platform
// expectation.
func TestHandleCapabilities_GetReturnsBothFlags(t *testing.T) {
	s := setupCapabilitiesTestServer(t)

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

	if !resp.TestMaster.Supported {
		t.Errorf("expected testMaster.supported=true on all platforms, got false")
	}
	if resp.TestMaster.Reason != "" {
		t.Errorf("expected empty testMaster.reason when supported, got %q", resp.TestMaster.Reason)
	}

	// Reflector dataplane requires CGO + Linux. On any other host the
	// stub build is linked in, Supported is false, and Reason must be
	// the operator-facing string the UI banner reads. We use runtime.GOOS
	// here (not cgo build tag, which we can't read at runtime) — when
	// running on Linux this test still passes whether cgo is on or off,
	// because Available() returns matching values for both Supported
	// and Reason.
	if runtime.GOOS != "linux" {
		if resp.Reflector.Supported {
			t.Errorf("expected reflector.supported=false on %s, got true", runtime.GOOS)
		}
		if resp.Reflector.Reason == "" {
			t.Errorf("expected reflector.reason populated when unsupported, got empty string")
		}
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
