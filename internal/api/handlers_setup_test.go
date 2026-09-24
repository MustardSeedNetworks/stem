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
)

// Test constants for setup handler tests.
const (
	setupTestUsername = "setupadmin"
	setupTestPassword = "setuppass123"
)

// setupSetupTestServer creates a server for setup handler tests.
func setupSetupTestServer(t testing.TB) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", setupTestUsername)
	t.Setenv("STEM_AUTH_PASSWORD", setupTestPassword)

	s, err := api.NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// setupServerClaimed creates a server that is unclaimed (no credential in
// the environment or the store) or claimed from the environment.
func setupServerClaimed(t testing.TB, claimed bool) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	if claimed {
		t.Setenv("STEM_AUTH_USERNAME", setupTestUsername)
		t.Setenv("STEM_AUTH_PASSWORD", setupTestPassword)
	} else {
		t.Setenv("STEM_AUTH_USERNAME", "")
		t.Setenv("STEM_AUTH_PASSWORD", "")
	}

	s, err := api.NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// assertStatusCode checks that the response has the expected status code.
func assertStatusCode(t testing.TB, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("Expected status %d, got %d", expected, w.Code)
	}
}

// TestHandleSetupStatus tests the GET /api/v1/setup/status endpoint.
func TestHandleSetupStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := setupSetupTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Should have 'needsSetup' field.
		if _, ok := resp["needsSetup"]; !ok {
			t.Error("Expected 'needsSetup' field in response")
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		s := setupSetupTestServer(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("response structure", func(t *testing.T) {
		s := setupSetupTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check needsSetup is boolean.
		_, ok := resp["needsSetup"].(bool)
		if !ok {
			t.Error("Expected 'needsSetup' to be a boolean")
		}
	})

	t.Run("content type", func(t *testing.T) {
		s := setupSetupTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})

	t.Run("PUT method not allowed", func(t *testing.T) {
		s := setupSetupTestServer(t)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for PUT, got %d", w.Code)
		}
	})

	t.Run("DELETE method not allowed", func(t *testing.T) {
		s := setupSetupTestServer(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for DELETE, got %d", w.Code)
		}
	})
}

// TestHandleSetupComplete tests the POST /api/v1/setup/complete endpoint.
func TestHandleSetupComplete(t *testing.T) {
	t.Run("method not allowed for GET", func(t *testing.T) {
		s := setupSetupTestServer(t)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/complete", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusMethodNotAllowed)
	})

	t.Run("invalid JSON when setup needed", func(t *testing.T) {
		s := setupServerClaimed(t, false)
		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", body)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusBadRequest)
	})

	t.Run("empty body when setup needed", func(t *testing.T) {
		s := setupServerClaimed(t, false)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusBadRequest)
	})

	t.Run("invalid setup token", func(t *testing.T) {
		s := setupServerClaimed(t, false)
		body := bytes.NewBufferString(`{"password":"SecurePassword123!","setupToken":"invalid-token"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", body)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusForbidden)
	})

	t.Run("setup already complete", func(t *testing.T) {
		s := setupServerClaimed(t, true)
		body := bytes.NewBufferString(`{"password":"SecurePassword123!","setupToken":"any-token"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", body)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusForbidden)
	})

	t.Run("PUT method not allowed", func(t *testing.T) {
		s := setupSetupTestServer(t)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/setup/complete", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusMethodNotAllowed)
	})

	t.Run("DELETE method not allowed", func(t *testing.T) {
		s := setupSetupTestServer(t)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/setup/complete", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusMethodNotAllowed)
	})
}

// TestSetupEndpointsNoAuth tests that setup endpoints don't require auth.
func TestSetupEndpointsNoAuth(t *testing.T) {
	s := setupSetupTestServer(t)

	endpoint := "/api/v1/setup/status"

	req := httptest.NewRequest(http.MethodGet, endpoint, nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	// Should not require auth for GET requests.
	if w.Code == http.StatusUnauthorized {
		t.Errorf("GET %s should not require authentication", endpoint)
	}
}

// TestSetupRoutesRegistered verifies setup endpoint routes are registered.
func TestSetupRoutesRegistered(t *testing.T) {
	s := setupSetupTestServer(t)

	routes := []struct {
		path   string
		method string
	}{
		{"/api/v1/setup/status", http.MethodGet},
		{"/api/v1/setup/complete", http.MethodPost},
	}

	for _, route := range routes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			// Should not be 404.
			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s returned 404", route.method, route.path)
			}
		})
	}
}

// getSetupStatusResponse makes a request to setup status and returns the response.
func getSetupStatusResponse(t *testing.T, s *api.Server) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assertStatusCode(t, w, http.StatusOK)

	var resp map[string]any
	unmarshalErr := json.Unmarshal(w.Body.Bytes(), &resp)
	if unmarshalErr != nil {
		t.Fatalf("Failed to decode response: %v", unmarshalErr)
	}
	return resp
}

// assertNeedsSetupBool checks that needsSetup is a boolean and returns its value.
func assertNeedsSetupBool(t *testing.T, resp map[string]any) bool {
	t.Helper()
	needsSetup, ok := resp["needsSetup"].(bool)
	if !ok {
		t.Error("Expected 'needsSetup' to be a boolean")
	}
	return needsSetup
}

// TestSetupStatusFollowsCredential: setup is offered exactly while no
// administrator credential exists.
func TestSetupStatusFollowsCredential(t *testing.T) {
	t.Run("claimed", func(t *testing.T) {
		s := setupServerClaimed(t, true)
		resp := getSetupStatusResponse(t, s)
		if assertNeedsSetupBool(t, resp) {
			t.Error("needsSetup = true for a daemon with a configured credential")
		}
		if _, leaked := resp["setupToken"]; leaked {
			t.Error("a claimed daemon issued a setup token")
		}
	})

	t.Run("unclaimed", func(t *testing.T) {
		s := setupServerClaimed(t, false)
		resp := getSetupStatusResponse(t, s)
		if !assertNeedsSetupBool(t, resp) {
			t.Error("needsSetup = false for a daemon with no credential")
		}
		if token, _ := resp["setupToken"].(string); token == "" {
			t.Error("an unclaimed daemon issued no setup token")
		}
	})
}

// TestSetupCompleteAlreadyDone: a claimed daemon refuses a second claim.
func TestSetupCompleteAlreadyDone(t *testing.T) {
	s := setupServerClaimed(t, true)

	body := bytes.NewBufferString(`{"password":"NewPassword123!","setupToken":"any-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	assertStatusCode(t, w, http.StatusForbidden)
}

// BenchmarkHandleSetupStatus benchmarks the setup status endpoint.
func BenchmarkHandleSetupStatus(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := api.NewServer(8444)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
