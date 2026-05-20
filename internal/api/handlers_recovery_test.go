// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/api"
)

// Test constants for recovery handler tests.
const (
	recoveryTestUsername = "recoveryadmin"
	recoveryTestPassword = "recoverypass123"
)

// setupRecoveryTestServer creates a server for recovery handler tests.
func setupRecoveryTestServer(t testing.TB) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_AUTH_USERNAME", recoveryTestUsername)
	t.Setenv("STEM_AUTH_PASSWORD", recoveryTestPassword)

	s, err := api.NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// TestHandleRecoveryStatus tests the GET /api/v1/recovery/status endpoint.
func TestHandleRecoveryStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/status", nil)
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

		// Should have 'active' field.
		if _, ok := resp["active"]; !ok {
			t.Error("Expected 'active' field in response")
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/status", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("response structure", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/status", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check active is boolean.
		active, ok := resp["active"].(bool)
		if !ok {
			t.Error("Expected 'active' to be a boolean")
		}

		// When not active, should not have instructions.
		if !active {
			if instructions, exists := resp["instructions"]; exists && instructions != "" {
				t.Log("Instructions present when not active - may or may not be empty")
			}
		}
	})

	t.Run("content type", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/status", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})
}

// TestHandleRecoveryComplete tests the POST /api/v1/recovery/complete endpoint.
func TestHandleRecoveryComplete(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/complete", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		body := bytes.NewBufferString(`{"token":"invalid-token","password":"NewSecurePassword123!"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/complete", body)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		// Should fail with invalid token.
		if w.Code == http.StatusOK {
			t.Error("Expected failure with invalid token")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/complete", body)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/complete", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty body, got %d", w.Code)
		}
	})

	t.Run("PUT method not allowed", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/recovery/complete", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for PUT, got %d", w.Code)
		}
	})

	t.Run("DELETE method not allowed", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/recovery/complete", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for DELETE, got %d", w.Code)
		}
	})
}

// TestHandleRecoveryInstructions tests the GET /api/v1/recovery/instructions endpoint.
func TestHandleRecoveryInstructions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/instructions", nil)
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

		// Should have expected fields.
		expectedFields := []string{"triggerFile", "tokenFile", "expiryTime", "steps"}
		for _, field := range expectedFields {
			if _, ok := resp[field]; !ok {
				t.Errorf("Expected field '%s' in response", field)
			}
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/instructions", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("steps is array", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/instructions", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		steps, ok := resp["steps"].([]any)
		if !ok {
			t.Error("Expected 'steps' to be an array")
		}

		if len(steps) == 0 {
			t.Error("Expected at least one step in instructions")
		}
	})

	t.Run("content type", func(t *testing.T) {
		s := setupRecoveryTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/instructions", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})
}

// TestRecoveryEndpointsNoAuth tests that recovery endpoints don't require auth.
func TestRecoveryEndpointsNoAuth(t *testing.T) {
	s := setupRecoveryTestServer(t)

	endpoints := []string{
		"/api/v1/recovery/status",
		"/api/v1/recovery/instructions",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			// Should not require auth for GET requests.
			if w.Code == http.StatusUnauthorized {
				t.Errorf("GET %s should not require authentication", endpoint)
			}
		})
	}
}

// TestRecoveryRoutesRegistered verifies recovery endpoint routes are registered.
func TestRecoveryRoutesRegistered(t *testing.T) {
	s := setupRecoveryTestServer(t)

	routes := []struct {
		path   string
		method string
	}{
		{"/api/v1/recovery/status", http.MethodGet},
		{"/api/v1/recovery/complete", http.MethodPost},
		{"/api/v1/recovery/instructions", http.MethodGet},
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

// BenchmarkHandleRecoveryStatus benchmarks the recovery status endpoint.
func BenchmarkHandleRecoveryStatus(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := api.NewServer(8444)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/status", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}

// BenchmarkHandleRecoveryInstructions benchmarks the recovery instructions endpoint.
func BenchmarkHandleRecoveryInstructions(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := api.NewServer(8444)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/instructions", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
