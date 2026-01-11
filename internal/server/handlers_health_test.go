// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/server"
)

// setupHealthTestServer creates a server for health endpoint tests.
func setupHealthTestServer(t testing.TB) *server.Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", "healthtest")
	t.Setenv("STEM_AUTH_PASSWORD", "healthpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// TestHandleHealthLive tests the GET /health/live endpoint.
func TestHandleHealthLive(t *testing.T) {
	s := setupHealthTestServer(t)

	t.Run("successful liveness check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp["status"] != "ok" {
			t.Errorf("Expected status 'ok', got '%v'", resp["status"])
		}
	})

	t.Run("liveness check method not allowed POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health/live", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("liveness check method not allowed PUT", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/health/live", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("liveness check method not allowed DELETE", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/health/live", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("liveness check content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})
}

// TestHandleHealthReady_Success tests successful readiness check.
func TestHandleHealthReady_Success(t *testing.T) {
	s := setupHealthTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["status"] != "ready" {
		t.Errorf("Expected status 'ready', got '%v'", resp["status"])
	}
}

// TestHandleHealthReady_ChecksPresent tests that required checks are present.
func TestHandleHealthReady_ChecksPresent(t *testing.T) {
	s := setupHealthTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	checks, ok := resp["checks"].(map[string]any)
	if !ok {
		t.Fatal("Expected 'checks' map in response")
	}

	requiredChecks := []string{"auth", "server", "license"}
	for _, check := range requiredChecks {
		if _, has := checks[check]; !has {
			t.Errorf("Expected '%s' check in response", check)
		}
	}
}

// TestHandleHealthReady_MethodNotAllowed tests that non-GET methods are rejected.
func TestHandleHealthReady_MethodNotAllowed(t *testing.T) {
	s := setupHealthTestServer(t)
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health/ready", nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleHealthReady_ContentType tests that response has correct content type.
func TestHandleHealthReady_ContentType(t *testing.T) {
	s := setupHealthTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

// TestHandleHealthReady_CheckStatuses tests that individual checks have ok status.
func TestHandleHealthReady_CheckStatuses(t *testing.T) {
	s := setupHealthTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	checks, ok := resp["checks"].(map[string]any)
	if !ok {
		t.Fatal("Expected 'checks' map in response")
	}

	// Verify auth check structure.
	if authCheck, hasAuth := checks["auth"].(map[string]any); hasAuth {
		if authCheck["status"] != "ok" {
			t.Errorf("Expected auth check status 'ok', got '%v'", authCheck["status"])
		}
	}

	// Verify server check structure.
	if serverCheck, hasServer := checks["server"].(map[string]any); hasServer {
		if serverCheck["status"] != "ok" {
			t.Errorf("Expected server check status 'ok', got '%v'", serverCheck["status"])
		}
	}
}

// TestHealthEndpointsNotVersioned tests that health endpoints are not under /api/v1/.
func TestHealthEndpointsNotVersioned(t *testing.T) {
	s := setupHealthTestServer(t)

	t.Run("live endpoint at root path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected /health/live to return 200, got %d", w.Code)
		}
	})

	t.Run("ready endpoint at root path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected /health/ready to return 200, got %d", w.Code)
		}
	})
}

// TestHealthEndpointsNoAuth tests that health endpoints don't require authentication.
func TestHealthEndpointsNoAuth(t *testing.T) {
	s := setupHealthTestServer(t)

	endpoints := []string{"/health/live", "/health/ready"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			// Request without any auth headers.
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			// Should succeed without auth.
			if w.Code == http.StatusUnauthorized {
				t.Errorf("Health endpoint %s should not require authentication", endpoint)
			}
		})
	}
}

// BenchmarkHandleHealthLive benchmarks the liveness endpoint.
func BenchmarkHandleHealthLive(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}

// BenchmarkHandleHealthReady benchmarks the readiness endpoint.
func BenchmarkHandleHealthReady(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
