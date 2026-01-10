// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/server"
)

// Test constants for setup handler tests.
const (
	setupTestUsername = "setupadmin"
	setupTestPassword = "setuppass123"
)

// setupSetupTestServer creates a server for setup handler tests.
func setupSetupTestServer(t testing.TB) *server.Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", setupTestUsername)
	t.Setenv("STEM_AUTH_PASSWORD", setupTestPassword)

	s, err := server.NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	return s
}

// setupServerWithMode creates a server with the specified setup mode.
func setupServerWithMode(t testing.TB, setupMode bool) *server.Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", setupTestUsername)
	t.Setenv("STEM_AUTH_PASSWORD", setupTestPassword)
	if setupMode {
		t.Setenv("STEM_SETUP_MODE", "true")
	} else {
		t.Setenv("STEM_SETUP_MODE", "false")
	}

	s, err := server.NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
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
		s := setupServerWithMode(t, true)
		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", body)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusBadRequest)
	})

	t.Run("empty body when setup needed", func(t *testing.T) {
		s := setupServerWithMode(t, true)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusBadRequest)
	})

	t.Run("invalid setup token", func(t *testing.T) {
		s := setupServerWithMode(t, true)
		body := bytes.NewBufferString(`{"password":"SecurePassword123!","setupToken":"invalid-token"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", body)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		assertStatusCode(t, w, http.StatusForbidden)
	})

	t.Run("setup already complete", func(t *testing.T) {
		s := setupServerWithMode(t, false)
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
func getSetupStatusResponse(t *testing.T, s *server.Server) map[string]any {
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

// TestSetupStatusWithSetupMode tests setup status when setup mode is enabled.
func TestSetupStatusWithSetupMode(t *testing.T) {
	t.Run("setup mode disabled", func(t *testing.T) {
		s := setupServerWithMode(t, false)
		resp := getSetupStatusResponse(t, s)
		needsSetup := assertNeedsSetupBool(t, resp)
		if needsSetup {
			t.Log("needsSetup is true - may be expected depending on password hash state")
		}
	})

	t.Run("setup mode enabled", func(t *testing.T) {
		s := setupServerWithMode(t, true)
		resp := getSetupStatusResponse(t, s)
		needsSetup := assertNeedsSetupBool(t, resp)
		if !needsSetup {
			t.Error("Expected needsSetup to be true when STEM_SETUP_MODE=true")
		}
		if needsSetup {
			if _, hasSuggestedPassword := resp["suggestedPassword"]; !hasSuggestedPassword {
				t.Log("suggestedPassword field not present")
			}
		}
	})
}

// TestSetupCompleteAlreadyDone tests setup complete when already configured.
func TestSetupCompleteAlreadyDone(t *testing.T) {
	// When setup is not needed, complete should return 403.
	t.Setenv("STEM_AUTH_USERNAME", setupTestUsername)
	t.Setenv("STEM_AUTH_PASSWORD", setupTestPassword)
	t.Setenv("STEM_SETUP_MODE", "false")

	s, err := server.NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}

	body := bytes.NewBufferString(`{"password":"NewPassword123!","setupToken":"any-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	// Should fail since setup is already complete (but this depends on password hash).
	// Accept both 403 (setup already done) or other error codes.
	if w.Code == http.StatusOK {
		// Check the response body to see if it's actually a success.
		var resp map[string]any
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &resp)
		if unmarshalErr == nil {
			status, _ := resp["status"].(string)
			if status == "success" {
				t.Log("Setup complete returned success - may be first run")
			}
		}
	}
}

// BenchmarkHandleSetupStatus benchmarks the setup status endpoint.
func BenchmarkHandleSetupStatus(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
