// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

// setupInterfacesTestServer creates a server for interface handler tests.
func setupInterfacesTestServer(t testing.TB) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_AUTH_USERNAME", testUsername)
	t.Setenv("STEM_AUTH_PASSWORD", testPassword)

	s, err := api.NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// TestHandleInterfaces_Success tests the GET /api/v1/interfaces endpoint.
func TestHandleInterfaces_Success(t *testing.T) {
	s := setupInterfacesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	stampAuth(t, s, req, loginToken(t, s))
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Response should be valid JSON array.
	var interfaces []any
	err := json.Unmarshal(w.Body.Bytes(), &interfaces)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

// TestHandleInterfaces_MethodNotAllowed tests non-GET methods are rejected.
func TestHandleInterfaces_MethodNotAllowed(t *testing.T) {
	s := setupInterfacesTestServer(t)
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	// Log in once outside the loop — auth now runs before the method
	// check (#340), so these must authenticate to reach the 405 path.
	jwt := loginToken(t, s)

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/interfaces", nil)
			stampAuth(t, s, req, jwt)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleInterfaces_ContentType tests that response has correct content type.
func TestHandleInterfaces_ContentType(t *testing.T) {
	s := setupInterfacesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	stampAuth(t, s, req, loginToken(t, s))
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

// TestHandleInterfaces_ResponseStructure tests the response structure.
func TestHandleInterfaces_ResponseStructure(t *testing.T) {
	s := setupInterfacesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	stampAuth(t, s, req, loginToken(t, s))
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var interfaces []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &interfaces)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// If we have interfaces, check their structure.
	if len(interfaces) > 0 {
		iface := interfaces[0]
		if _, ok := iface["name"]; !ok {
			t.Error("Expected 'name' field in interface response")
		}
	}
}

// BenchmarkHandleInterfaces benchmarks the interfaces endpoint.
func BenchmarkHandleInterfaces(b *testing.B) {
	b.Setenv("STEM_TEST_MODE", "1")
	b.Setenv("STEM_AUTH_USERNAME", testUsername)
	b.Setenv("STEM_AUTH_PASSWORD", testPassword)

	s, err := api.NewServer(8444)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })
	// /interfaces is auth-gated (#340); log in once outside the loop.
	jwt := loginToken(b, s)

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
