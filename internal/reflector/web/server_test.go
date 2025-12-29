// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s := New(nil, 8080)
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.port != 8080 {
		t.Errorf("Expected port 8080, got %d", s.port)
	}
	if s.mux == nil {
		t.Error("mux should not be nil")
	}
	if s.startTime.IsZero() {
		t.Error("startTime should be initialized")
	}
}

func TestHandleHealth(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check that response has a status field
	if _, ok := resp["status"]; !ok {
		t.Error("Expected 'status' field in response")
	}
}

func TestHandleStats(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	s.handleStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var resp StatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleResetStats(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodPost, "/api/stats/reset", nil)
	w := httptest.NewRecorder()

	s.handleResetStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleConfigOptions(t *testing.T) {
	s := New(nil, 8080)
	// Test CORS preflight request (doesn't require dataplane)
	req := httptest.NewRequest(http.MethodOptions, "/api/config", nil)
	w := httptest.NewRecorder()

	s.handleConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected Access-Control-Allow-Origin header")
	}
}

func TestHandleFallback(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	s.handleFallback(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type 'text/html', got '%s'", contentType)
	}
}

func TestStatsResponseStruct(t *testing.T) {
	resp := StatsResponse{
		Timestamp:        time.Now().Format(time.RFC3339),
		Uptime:           3600.5,
		Interface:        "eth0",
		Running:          true,
		PacketsReceived:  1000000,
		PacketsReflected: 999999,
		BytesReceived:    64000000,
		BytesReflected:   63999936,
		TxErrors:         1,
		RxInvalid:        5,
		RatePackets:      100000.5,
		RateMbps:         800.25,
	}
	resp.Signatures.RFC2544 = 500
	resp.Signatures.Y1564 = 200
	resp.Latency.AvgUs = 25.5
	resp.Latency.Enabled = true

	if resp.Interface != "eth0" {
		t.Errorf("Expected Interface 'eth0', got '%s'", resp.Interface)
	}
	if !resp.Running {
		t.Error("Expected Running true")
	}
	if resp.Signatures.RFC2544 != 500 {
		t.Errorf("Expected RFC2544 500, got %d", resp.Signatures.RFC2544)
	}
	if resp.Latency.AvgUs != 25.5 {
		t.Errorf("Expected Latency.AvgUs 25.5, got %f", resp.Latency.AvgUs)
	}
}

func TestConfigResponseStruct(t *testing.T) {
	resp := ConfigResponse{
		Interface:       "eth0",
		SignatureFilter: "all",
	}
	resp.Filtering.Port = 3842
	resp.Filtering.FilterOUI = true
	resp.Filtering.OUI = "00:c0:17"
	resp.Reflection.Mode = "all"
	resp.Platform.Type = "af_xdp"

	if resp.Interface != "eth0" {
		t.Errorf("Expected Interface 'eth0', got '%s'", resp.Interface)
	}
	if resp.Filtering.Port != 3842 {
		t.Errorf("Expected Filtering.Port 3842, got %d", resp.Filtering.Port)
	}
	if resp.Reflection.Mode != "all" {
		t.Errorf("Expected Reflection.Mode 'all', got '%s'", resp.Reflection.Mode)
	}
}

func TestServerRoutesRegistered(t *testing.T) {
	s := New(nil, 8080)

	// Only test routes that don't require a valid dataplane
	routes := []struct {
		path   string
		method string
	}{
		{"/api/health", http.MethodGet},
		{"/api/stats", http.MethodGet},
		// Note: /api/config requires valid dataplane, tested separately with OPTIONS
	}

	for _, route := range routes {
		t.Run(route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			s.mux.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s returned 404", route.method, route.path)
			}
		})
	}
}

// Benchmark tests
func BenchmarkHandleHealth(b *testing.B) {
	s := New(nil, 8080)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()
		s.handleHealth(w, req)
	}
}

func BenchmarkHandleStats(b *testing.B) {
	s := New(nil, 8080)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		w := httptest.NewRecorder()
		s.handleStats(w, req)
	}
}

// Additional tests for better coverage
func TestHandleStatsMethodNotAllowed(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodPost, "/api/stats", nil)
	w := httptest.NewRecorder()

	s.handleStats(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleResetStatsMethodNotAllowed(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/api/stats/reset", nil)
	w := httptest.NewRecorder()

	s.handleResetStats(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleResetStatsOptions(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodOptions, "/api/stats/reset", nil)
	w := httptest.NewRecorder()

	s.handleResetStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}
}

func TestHandleConfigGet(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	s.handleConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should return JSON with platform type "none" since no dataplane
	var resp ConfigResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Platform.Type != "none" {
		t.Errorf("Expected platform type 'none', got '%s'", resp.Platform.Type)
	}
}

func TestHandleConfigNoDataplaneAnyMethod(t *testing.T) {
	// With nil dataplane, handleConfig returns 200 with platform "none" for any method
	s := New(nil, 8080)
	methods := []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/config", nil)
		w := httptest.NewRecorder()

		s.handleConfig(w, req)

		// All should return 200 since dp is nil (returns early with "none" platform)
		if w.Code != http.StatusOK {
			t.Errorf("Method %s: Expected status 200, got %d", method, w.Code)
		}

		var resp ConfigResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Errorf("Method %s: Failed to parse response: %v", method, err)
		}
		if resp.Platform.Type != "none" {
			t.Errorf("Method %s: Expected platform type 'none', got '%s'", method, resp.Platform.Type)
		}
	}
}

func TestHandleFallbackNotFoundPath(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	s.handleFallback(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestServerPort(t *testing.T) {
	testCases := []int{80, 8080, 9000, 3000}

	for _, port := range testCases {
		s := New(nil, port)
		if s.port != port {
			t.Errorf("Expected port %d, got %d", port, s.port)
		}
	}
}

func TestServerStartTime(t *testing.T) {
	before := time.Now()
	s := New(nil, 8080)
	after := time.Now()

	if s.startTime.Before(before) {
		t.Error("startTime should be after test start")
	}
	if s.startTime.After(after) {
		t.Error("startTime should be before test end")
	}
}

func TestHandleHealthCORS(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	// Check CORS header
	cors := w.Header().Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", cors)
	}
}

func TestHandleStatsCORS(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	s.handleStats(w, req)

	// Check CORS header
	cors := w.Header().Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", cors)
	}
}

func TestStatsResponseTimestamp(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	s.handleStats(w, req)

	var resp StatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Timestamp == "" {
		t.Error("Expected non-empty timestamp")
	}

	// Timestamp should be valid RFC3339
	_, err := time.Parse(time.RFC3339, resp.Timestamp)
	if err != nil {
		t.Errorf("Invalid timestamp format: %v", err)
	}
}

func TestStatsResponseUptime(t *testing.T) {
	s := New(nil, 8080)
	time.Sleep(10 * time.Millisecond) // Small delay to ensure uptime > 0

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	s.handleStats(w, req)

	var resp StatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Uptime <= 0 {
		t.Error("Expected positive uptime")
	}
}

func TestHealthResponseFields(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check required fields
	if _, ok := resp["status"]; !ok {
		t.Error("Expected 'status' field")
	}
	if _, ok := resp["uptime"]; !ok {
		t.Error("Expected 'uptime' field")
	}
	if _, ok := resp["version"]; !ok {
		t.Error("Expected 'version' field")
	}
	if _, ok := resp["running"]; !ok {
		t.Error("Expected 'running' field")
	}
}

func TestResetStatsResetsUptime(t *testing.T) {
	s := New(nil, 8080)
	originalTime := s.startTime

	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest(http.MethodPost, "/api/stats/reset", nil)
	w := httptest.NewRecorder()

	s.handleResetStats(w, req)

	if !s.startTime.After(originalTime) {
		t.Error("Expected startTime to be reset to later time")
	}
}

func TestConfigOptionsMethodCORS(t *testing.T) {
	s := New(nil, 8080)
	req := httptest.NewRequest(http.MethodOptions, "/api/config", nil)
	w := httptest.NewRecorder()

	s.handleConfig(w, req)

	// Check CORS headers for preflight
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS Allow-Origin header")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected CORS Allow-Methods header")
	}
}
