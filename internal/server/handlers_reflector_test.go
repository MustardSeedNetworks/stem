// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/server"
)

// setupReflectorTestServer creates a server for reflector handler tests.
func setupReflectorTestServer(t testing.TB) *server.Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", "reflectortest")
	t.Setenv("STEM_AUTH_PASSWORD", "reflectorpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// getReflectorStatsResponse is a helper that fetches and decodes stats response.
func getReflectorStatsResponse(t *testing.T, s *server.Server) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/stats", nil)
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
	return resp
}

// TestHandleReflectorStats_Structure tests the reflector stats response fields.
func TestHandleReflectorStats_Structure(t *testing.T) {
	s := setupReflectorTestServer(t)
	resp := getReflectorStatsResponse(t, s)

	expectedFields := []string{
		"running", "packetsReceived", "packetsReflected", "bytesReceived",
		"bytesReflected", "txErrors", "rxInvalid", "ratePps", "rateMbps",
		"signatures", "latency", "uptime",
	}

	for _, field := range expectedFields {
		if _, ok := resp[field]; !ok {
			t.Errorf("Expected field '%s' in response", field)
		}
	}
}

// TestHandleReflectorStats_Signatures tests the signatures sub-object.
func TestHandleReflectorStats_Signatures(t *testing.T) {
	s := setupReflectorTestServer(t)
	resp := getReflectorStatsResponse(t, s)

	signatures, ok := resp["signatures"].(map[string]any)
	if !ok {
		t.Fatal("Expected 'signatures' to be an object")
	}

	signatureFields := []string{"probeot", "dataot", "latency", "rfc2544", "y1564", "msn"}
	for _, field := range signatureFields {
		if _, found := signatures[field]; !found {
			t.Errorf("Expected field 'signatures.%s' in response", field)
		}
	}
}

// TestHandleReflectorStats_Latency tests the latency sub-object.
func TestHandleReflectorStats_Latency(t *testing.T) {
	s := setupReflectorTestServer(t)
	resp := getReflectorStatsResponse(t, s)

	latency, ok := resp["latency"].(map[string]any)
	if !ok {
		t.Fatal("Expected 'latency' to be an object")
	}

	latencyFields := []string{"minUs", "avgUs", "maxUs", "count"}
	for _, field := range latencyFields {
		if _, found := latency[field]; !found {
			t.Errorf("Expected field 'latency.%s' in response", field)
		}
	}
}

// TestHandleReflectorStats_RunningBoolean tests running field type.
func TestHandleReflectorStats_RunningBoolean(t *testing.T) {
	s := setupReflectorTestServer(t)
	resp := getReflectorStatsResponse(t, s)

	_, ok := resp["running"].(bool)
	if !ok {
		t.Error("Expected 'running' to be a boolean")
	}
}

// TestHandleReflectorStats_UptimeNonNegative tests uptime is non-negative.
func TestHandleReflectorStats_UptimeNonNegative(t *testing.T) {
	s := setupReflectorTestServer(t)
	resp := getReflectorStatsResponse(t, s)

	uptime, ok := resp["uptime"].(float64)
	if !ok {
		t.Fatal("Expected 'uptime' to be a number")
	}
	if uptime < 0 {
		t.Errorf("Expected non-negative uptime, got %f", uptime)
	}
}

// TestHandleReflectorStatsMethodNotAllowed tests method restrictions.
func TestHandleReflectorStatsMethodNotAllowed(t *testing.T) {
	s := setupReflectorTestServer(t)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/reflector/stats", nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleReflectorConfigMethodRouting tests config endpoint method routing.
func TestHandleReflectorConfigMethodRouting(t *testing.T) {
	s := setupReflectorTestServer(t)

	t.Run("GET config returns config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/config", nil)
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

		// Verify config fields.
		if _, ok := resp["profile"]; !ok {
			t.Error("Expected 'profile' field in config")
		}
		if _, ok := resp["portFilter"]; !ok {
			t.Error("Expected 'portFilter' field in config")
		}
	})

	t.Run("PUT config method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/reflector/config", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for PUT, got %d", w.Code)
		}
	})

	t.Run("DELETE config method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/reflector/config", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for DELETE, got %d", w.Code)
		}
	})
}

// TestHandleReflectorConfigDefaults tests default config values.
func TestHandleReflectorConfigDefaults(t *testing.T) {
	s := setupReflectorTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/config", nil)
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

	// Default profile should be "all".
	profile, ok := resp["profile"].(string)
	if !ok || profile != "all" {
		t.Errorf("Expected default profile 'all', got '%v'", resp["profile"])
	}

	// Default port filter should be 3842.
	portFilter, ok := resp["portFilter"].(float64)
	if !ok || int(portFilter) != 3842 {
		t.Errorf("Expected default portFilter 3842, got '%v'", resp["portFilter"])
	}

	// Default OUI filter should be "00:c0:17".
	ouiFilter, ok := resp["ouiFilter"].(string)
	if !ok || ouiFilter != "00:c0:17" {
		t.Errorf("Expected default ouiFilter '00:c0:17', got '%v'", resp["ouiFilter"])
	}
}

// TestReflectorEndpointsContentType tests content type headers.
func TestReflectorEndpointsContentType(t *testing.T) {
	s := setupReflectorTestServer(t)

	endpoints := []string{"/api/v1/reflector/config", "/api/v1/reflector/stats"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}
		})
	}
}

// TestReflectorEndpointsNoAuth tests that reflector endpoints don't require auth.
func TestReflectorEndpointsNoAuth(t *testing.T) {
	s := setupReflectorTestServer(t)

	endpoints := []string{"/api/v1/reflector/config", "/api/v1/reflector/stats"}

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

// BenchmarkHandleReflectorStats benchmarks the reflector stats endpoint.
func BenchmarkHandleReflectorStats(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/stats", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}

// BenchmarkHandleReflectorConfig benchmarks the reflector config endpoint.
func BenchmarkHandleReflectorConfig(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := server.NewServer(8080)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/config", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}

// getReflectorAuthToken returns an auth token for reflector config tests.
func getReflectorAuthToken(t *testing.T, s *server.Server) string {
	t.Helper()
	body := bytes.NewBufferString(`{"username":"reflectortest","password":"reflectorpass123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected login status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Fatalf("Login response missing token: %v", resp)
	}
	return token
}

// TestHandleReflectorConfigPost_Success tests POST /api/v1/reflector/config.
func TestHandleReflectorConfigPost_Success(t *testing.T) {
	s := setupReflectorTestServer(t)
	token := getReflectorAuthToken(t, s)

	body := bytes.NewBufferString(`{"profile":"netally","portFilter":9999}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
	req.Header.Set("Authorization", "Bearer "+token)
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

	if resp["status"] != "updated" {
		t.Errorf("Expected status 'updated', got '%v'", resp["status"])
	}
}

// TestHandleReflectorConfigPost_InvalidProfile tests POST with invalid profile.
func TestHandleReflectorConfigPost_InvalidProfile(t *testing.T) {
	s := setupReflectorTestServer(t)
	token := getReflectorAuthToken(t, s)

	body := bytes.NewBufferString(`{"profile":"invalid_profile"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid profile, got %d", w.Code)
	}
}

// TestHandleReflectorConfigPost_OUIFilter tests POST with OUI filter.
func TestHandleReflectorConfigPost_OUIFilter(t *testing.T) {
	s := setupReflectorTestServer(t)
	token := getReflectorAuthToken(t, s)

	body := bytes.NewBufferString(`{"ouiFilter":"aa:bb:cc"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleReflectorConfigPost_SignatureFilter tests POST with signature filter.
func TestHandleReflectorConfigPost_SignatureFilter(t *testing.T) {
	s := setupReflectorTestServer(t)
	token := getReflectorAuthToken(t, s)

	body := bytes.NewBufferString(`{"signatureFilter":["probeot","dataot"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleReflectorConfigPost_InvalidJSON tests POST with invalid JSON.
func TestHandleReflectorConfigPost_InvalidJSON(t *testing.T) {
	s := setupReflectorTestServer(t)
	token := getReflectorAuthToken(t, s)

	body := bytes.NewBufferString(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
	}
}

// TestHandleReflectorConfigPost_EmptyBody tests POST with empty body.
func TestHandleReflectorConfigPost_EmptyBody(t *testing.T) {
	s := setupReflectorTestServer(t)
	token := getReflectorAuthToken(t, s)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty body, got %d", w.Code)
	}
}

// TestHandleReflectorConfigPost_NoChanges tests POST with no actual changes.
func TestHandleReflectorConfigPost_NoChanges(t *testing.T) {
	s := setupReflectorTestServer(t)
	token := getReflectorAuthToken(t, s)

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for empty config, got %d: %s", w.Code, w.Body.String())
	}
}
