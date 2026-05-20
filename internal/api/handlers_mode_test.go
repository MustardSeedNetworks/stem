// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/krisarmstrong/stem/internal/api"
)

// setupModeTestServer builds a server tuned for /api/v1/mode tests.
// It uses [STEM_TEST_MODE]=1 for fast bcrypt and registers automatic
// shutdown so per-test rate limiters do not leak goroutines.
func setupModeTestServer(t *testing.T) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_AUTH_USERNAME", "modeuser")
	t.Setenv("STEM_AUTH_PASSWORD", "modepass123")

	s, err := api.NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// postMode is a small helper that issues POST /api/v1/mode and returns
// the decoded response. It always uses ServeHTTP so the middleware
// stack (api-version header, CSRF skip path, etc.) matches production.
func postMode(t *testing.T, s *api.Server, body string) (*httptest.ResponseRecorder, api.ModeUpdateResponse) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mode", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	var resp api.ModeUpdateResponse
	if w.Code == http.StatusOK {
		if decodeErr := json.Unmarshal(w.Body.Bytes(), &resp); decodeErr != nil {
			t.Fatalf("decode response: %v body=%q", decodeErr, w.Body.String())
		}
	}
	return w, resp
}

// TestHandleMode_PostUpdatesModeAndReturnsPrevious is the happy path:
// the server starts in test_master mode, we POST reflector, and we
// expect the body to echo both the new mode and the previous mode it
// replaced.
func TestHandleMode_PostUpdatesModeAndReturnsPrevious(t *testing.T) {
	s := setupModeTestServer(t)
	// Force the reflector probe to return supported so this test
	// does not depend on the runtime OS / cgo state.
	s.UseReflectorAvailabilityForTest(func() (bool, string) { return true, "" })

	w, resp := postMode(t, s, `{"mode":"reflector"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", w.Code, w.Body.String())
	}

	if resp.Status != "updated" {
		t.Errorf("expected status=updated, got %q", resp.Status)
	}
	if resp.Mode != "reflector" {
		t.Errorf("expected mode=reflector, got %q", resp.Mode)
	}
	if resp.Previous != "test_master" {
		t.Errorf("expected previous=test_master, got %q", resp.Previous)
	}
}

// TestHandleMode_PostSameModeNoOp asserts the "same mode" short-circuit:
// the response is 200 with status="unchanged" and previous == mode.
// No teardown side effects should fire — but here we only assert the
// API contract, since teardown is best observed in the executor test.
func TestHandleMode_PostSameModeNoOp(t *testing.T) {
	s := setupModeTestServer(t)
	s.UseReflectorAvailabilityForTest(func() (bool, string) { return true, "" })

	// Server boots in test_master mode. Asking for test_master again
	// must be a no-op.
	w, resp := postMode(t, s, `{"mode":"test_master"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", w.Code, w.Body.String())
	}
	if resp.Status != "unchanged" {
		t.Errorf("expected status=unchanged, got %q", resp.Status)
	}
	if resp.Mode != "test_master" {
		t.Errorf("expected mode=test_master, got %q", resp.Mode)
	}
	if resp.Previous != "test_master" {
		t.Errorf("expected previous=test_master, got %q", resp.Previous)
	}
}

// TestHandleMode_PostUnsupportedPlatformReturns403 asserts that asking
// for a mode the binary cannot support yields 403 with a reason. The
// availability probe is overridden to simulate the macOS / Windows
// pure-Go build behaviour without rebuilding.
func TestHandleMode_PostUnsupportedPlatformReturns403(t *testing.T) {
	s := setupModeTestServer(t)
	s.UseReflectorAvailabilityForTest(func() (bool, string) {
		return false, "CGO + Linux required"
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mode",
		bytes.NewBufferString(`{"mode":"reflector"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%q", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "CGO + Linux required") {
		t.Errorf("expected reason in response body, got %q", w.Body.String())
	}
}

// TestHandleMode_PostUnsupportedDoesNotBlockTestMaster verifies the
// platform gate only fires for reflector. Test Master must remain
// reachable on every platform the binary builds for today.
func TestHandleMode_PostUnsupportedDoesNotBlockTestMaster(t *testing.T) {
	s := setupModeTestServer(t)
	s.UseReflectorAvailabilityForTest(func() (bool, string) {
		return false, "CGO + Linux required"
	})

	// First flip to reflector via the test override below would
	// require bypassing the gate, so instead flip to test_master and
	// assert the response shape — the server boots in test_master so
	// this exercises the no-op branch, which is enough to prove the
	// gate did not run.
	w, resp := postMode(t, s, `{"mode":"test_master"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for test_master switch, got %d body=%q", w.Code, w.Body.String())
	}
	if resp.Mode != "test_master" {
		t.Errorf("expected mode=test_master, got %q", resp.Mode)
	}
}

// TestHandleMode_PostInvalidModeReturns400 keeps the original
// invalid-value contract: unknown mode strings must 400, not 403.
func TestHandleMode_PostInvalidModeReturns400(t *testing.T) {
	s := setupModeTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mode",
		bytes.NewBufferString(`{"mode":"banana"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%q", w.Code, w.Body.String())
	}
}

// TestHandleMode_GetReturnsCurrentMode is a thin sanity check that the
// GET branch is untouched by the new POST plumbing.
func TestHandleMode_GetReturnsCurrentMode(t *testing.T) {
	s := setupModeTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mode", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", w.Code, w.Body.String())
	}
	var resp api.ModeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%q", err, w.Body.String())
	}
	if resp.Mode != "test_master" && resp.Mode != "reflector" {
		t.Errorf("expected mode in {test_master,reflector}, got %q", resp.Mode)
	}
}
