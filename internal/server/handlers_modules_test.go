// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/server"
)

// setupModulesTestServer creates a server for modules handler tests.
func setupModulesTestServer(t testing.TB) *server.Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", "modulestest")
	t.Setenv("STEM_AUTH_PASSWORD", "modulespass123")

	s, err := server.NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// TestHandleModules_GetSuccess tests GET /api/v1/modules.
func TestHandleModules_GetSuccess(t *testing.T) {
	s := setupModulesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules", nil)
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

	if _, ok := resp["modules"]; !ok {
		t.Error("Expected 'modules' field in response")
	}
	if _, ok := resp["count"]; !ok {
		t.Error("Expected 'count' field in response")
	}
}

// TestHandleModules_MethodNotAllowed tests non-GET methods.
func TestHandleModules_MethodNotAllowed(t *testing.T) {
	s := setupModulesTestServer(t)
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/modules", nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleModules_ModulesList tests that modules list is returned.
func TestHandleModules_ModulesList(t *testing.T) {
	s := setupModulesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	modules, ok := resp["modules"].([]any)
	if !ok {
		t.Fatal("Expected 'modules' to be an array")
	}

	// Should have at least one module.
	if len(modules) == 0 {
		t.Error("Expected at least one module")
	}
}

// TestHandleModuleByName_Success tests GET /api/v1/modules/{name}.
func TestHandleModuleByName_Success(t *testing.T) {
	s := setupModulesTestServer(t)

	// Test known modules.
	knownModules := []string{"reflector", "benchmark", "servicetest", "trafficgen", "measure", "certify"}

	for _, moduleName := range knownModules {
		t.Run(moduleName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/"+moduleName, nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", moduleName, w.Code)
			}

			var resp map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			if err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if _, ok := resp["name"]; !ok {
				t.Error("Expected 'name' field in module response")
			}
		})
	}
}

// TestHandleModuleByName_NotFound tests GET /api/v1/modules/{name} with unknown module.
func TestHandleModuleByName_NotFound(t *testing.T) {
	s := setupModulesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/nonexistent_module", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for unknown module, got %d", w.Code)
	}
}

// TestHandleModuleByName_MethodNotAllowed tests non-GET methods.
func TestHandleModuleByName_MethodNotAllowed(t *testing.T) {
	s := setupModulesTestServer(t)
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/modules/benchmark", nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleModuleTests_Success tests GET /api/v1/modules/{name}/tests.
func TestHandleModuleTests_Success(t *testing.T) {
	s := setupModulesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/benchmark/tests", nil)
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

	if _, ok := resp["module"]; !ok {
		t.Error("Expected 'module' field in response")
	}
	if _, ok := resp["tests"]; !ok {
		t.Error("Expected 'tests' field in response")
	}
	if _, ok := resp["count"]; !ok {
		t.Error("Expected 'count' field in response")
	}
}

// TestHandleModuleTests_NotFound tests GET /api/v1/modules/{name}/tests with unknown module.
func TestHandleModuleTests_NotFound(t *testing.T) {
	s := setupModulesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/nonexistent/tests", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for unknown module tests, got %d", w.Code)
	}
}

// TestHandleModules_ContentType tests response content type.
func TestHandleModules_ContentType(t *testing.T) {
	s := setupModulesTestServer(t)

	endpoints := []string{"/api/v1/modules", "/api/v1/modules/benchmark", "/api/v1/modules/benchmark/tests"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
				}
			}
		})
	}
}

// TestHandleModules_ModuleStructure tests module response structure.
func TestHandleModules_ModuleStructure(t *testing.T) {
	s := setupModulesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/benchmark", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedFields := []string{"name", "displayName", "description", "color", "standard", "tests"}
	for _, field := range expectedFields {
		if _, ok := resp[field]; !ok {
			t.Errorf("Expected field '%s' in module response", field)
		}
	}
}

// TestHandleModulesNoAuth tests that modules endpoints don't require auth.
func TestHandleModulesNoAuth(t *testing.T) {
	s := setupModulesTestServer(t)

	endpoints := []string{"/api/v1/modules", "/api/v1/modules/benchmark", "/api/v1/modules/benchmark/tests"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			// Request without any auth headers.
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			// Should succeed without auth.
			if w.Code == http.StatusUnauthorized {
				t.Errorf("Modules endpoint %s should not require authentication", endpoint)
			}
		})
	}
}

// BenchmarkHandleModules benchmarks the modules list endpoint.
func BenchmarkHandleModules(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/modules", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}

// BenchmarkHandleModuleByName benchmarks the module by name endpoint.
func BenchmarkHandleModuleByName(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/benchmark", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
