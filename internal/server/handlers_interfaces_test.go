// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/server"
)

// setupInterfacesTestServer creates a server for interface handler tests.
func setupInterfacesTestServer(t testing.TB) *server.Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", "interfacetest")
	t.Setenv("STEM_AUTH_PASSWORD", "interfacepass123")

	s, err := server.NewServer(8080)
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

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/interfaces", nil)
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
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
