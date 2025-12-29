// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	s := NewServer(8080)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.port != 8080 {
		t.Errorf("Expected port 8080, got %d", s.port)
	}
	if s.mux == nil {
		t.Error("Server mux is nil")
	}
	if s.stats == nil {
		t.Error("Server stats is nil")
	}
	if s.testStatus != "idle" {
		t.Errorf("Expected initial testStatus 'idle', got '%s'", s.testStatus)
	}
	if s.mode != "test_master" {
		t.Errorf("Expected initial mode 'test_master', got '%s'", s.mode)
	}
}

func TestHandleHealth(t *testing.T) {
	s := NewServer(8080)
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

	if resp["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", resp["status"])
	}
	if resp["product"] != "Seed Test Suite" {
		t.Errorf("Expected product 'Seed Test Suite', got '%v'", resp["product"])
	}
	if resp["company"] != "Mustard Seed Networks" {
		t.Errorf("Expected company 'Mustard Seed Networks', got '%v'", resp["company"])
	}
}

func TestHandleHealthMethodNotAllowed(t *testing.T) {
	s := NewServer(8080)
	req := httptest.NewRequest(http.MethodPost, "/api/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleInterfaces(t *testing.T) {
	s := NewServer(8080)
	req := httptest.NewRequest(http.MethodGet, "/api/interfaces", nil)
	w := httptest.NewRecorder()

	s.handleInterfaces(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Response should be valid JSON array
	var resp []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleStats(t *testing.T) {
	s := NewServer(8080)
	s.stats.PacketsReceived = 1000
	s.stats.PacketsSent = 900

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	s.handleStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp Stats
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.PacketsReceived != 1000 {
		t.Errorf("Expected PacketsReceived 1000, got %d", resp.PacketsReceived)
	}
	if resp.PacketsSent != 900 {
		t.Errorf("Expected PacketsSent 900, got %d", resp.PacketsSent)
	}
}

func TestHandleTestStart(t *testing.T) {
	s := NewServer(8080)
	req := httptest.NewRequest(http.MethodPost, "/api/test/start", nil)
	w := httptest.NewRecorder()

	s.handleTestStart(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if s.testStatus != "running" {
		t.Errorf("Expected testStatus 'running', got '%s'", s.testStatus)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["status"] != "started" {
		t.Errorf("Expected status 'started', got '%s'", resp["status"])
	}
}

func TestHandleTestStartConflict(t *testing.T) {
	s := NewServer(8080)
	s.testStatus = "running"

	req := httptest.NewRequest(http.MethodPost, "/api/test/start", nil)
	w := httptest.NewRecorder()

	s.handleTestStart(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 (conflict), got %d", w.Code)
	}
}

func TestHandleTestStop(t *testing.T) {
	s := NewServer(8080)
	s.testStatus = "running"
	s.currentTest = "throughput"

	req := httptest.NewRequest(http.MethodPost, "/api/test/stop", nil)
	w := httptest.NewRecorder()

	s.handleTestStop(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if s.testStatus != "completed" {
		t.Errorf("Expected testStatus 'completed', got '%s'", s.testStatus)
	}
	if s.currentTest != "" {
		t.Errorf("Expected currentTest '', got '%s'", s.currentTest)
	}
}

func TestHandleTestStopNoTestRunning(t *testing.T) {
	s := NewServer(8080)

	req := httptest.NewRequest(http.MethodPost, "/api/test/stop", nil)
	w := httptest.NewRecorder()

	s.handleTestStop(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 (bad request), got %d", w.Code)
	}
}

func TestHandleSettingsGet(t *testing.T) {
	s := NewServer(8080)
	s.selectedIface = "eth0"

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()

	s.handleSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["interface"] != "eth0" {
		t.Errorf("Expected interface 'eth0', got '%v'", resp["interface"])
	}
}

func TestHandleSettingsPost(t *testing.T) {
	s := NewServer(8080)

	body := bytes.NewBufferString(`{"interface": "enp1s0"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/settings", body)
	w := httptest.NewRecorder()

	s.handleSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if s.selectedIface != "enp1s0" {
		t.Errorf("Expected selectedIface 'enp1s0', got '%s'", s.selectedIface)
	}
}

func TestHandleSettingsInvalidJSON(t *testing.T) {
	s := NewServer(8080)

	body := bytes.NewBufferString(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/settings", body)
	w := httptest.NewRecorder()

	s.handleSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleModeGet(t *testing.T) {
	s := NewServer(8080)
	s.mode = "reflector"

	req := httptest.NewRequest(http.MethodGet, "/api/mode", nil)
	w := httptest.NewRecorder()

	s.handleMode(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["mode"] != "reflector" {
		t.Errorf("Expected mode 'reflector', got '%s'", resp["mode"])
	}
}

func TestHandleModePost(t *testing.T) {
	s := NewServer(8080)

	body := bytes.NewBufferString(`{"mode": "reflector"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/mode", body)
	w := httptest.NewRecorder()

	s.handleMode(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if s.mode != "reflector" {
		t.Errorf("Expected mode 'reflector', got '%s'", s.mode)
	}
}

func TestHandleModePostInvalid(t *testing.T) {
	s := NewServer(8080)

	body := bytes.NewBufferString(`{"mode": "invalid_mode"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/mode", body)
	w := httptest.NewRecorder()

	s.handleMode(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleReflectorConfigGet(t *testing.T) {
	s := NewServer(8080)

	req := httptest.NewRequest(http.MethodGet, "/api/reflector/config", nil)
	w := httptest.NewRecorder()

	s.handleReflectorConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp ReflectorConfig
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Profile != "all" {
		t.Errorf("Expected default profile 'all', got '%s'", resp.Profile)
	}
	if resp.PortFilter != 3842 {
		t.Errorf("Expected default port 3842, got %d", resp.PortFilter)
	}
}

func TestHandleReflectorConfigPost(t *testing.T) {
	s := NewServer(8080)

	body := bytes.NewBufferString(`{"profile": "netally", "portFilter": 9999}`)
	req := httptest.NewRequest(http.MethodPost, "/api/reflector/config", body)
	w := httptest.NewRecorder()

	s.handleReflectorConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if s.reflectorConfig.Profile != "netally" {
		t.Errorf("Expected profile 'netally', got '%s'", s.reflectorConfig.Profile)
	}
	if s.reflectorConfig.PortFilter != 9999 {
		t.Errorf("Expected port 9999, got %d", s.reflectorConfig.PortFilter)
	}
}

func TestHandleReflectorConfigPostInvalidProfile(t *testing.T) {
	s := NewServer(8080)

	body := bytes.NewBufferString(`{"profile": "invalid_profile"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/reflector/config", body)
	w := httptest.NewRecorder()

	s.handleReflectorConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleReflectorStats(t *testing.T) {
	s := NewServer(8080)
	s.mode = "reflector"
	s.testStatus = "running"
	s.stats.PacketsReceived = 5000
	s.stats.PacketsSent = 4900

	req := httptest.NewRequest(http.MethodGet, "/api/reflector/stats", nil)
	w := httptest.NewRecorder()

	s.handleReflectorStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp ReflectorStats
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Running {
		t.Error("Expected Running true")
	}
	if resp.PacketsReceived != 5000 {
		t.Errorf("Expected PacketsReceived 5000, got %d", resp.PacketsReceived)
	}
	if resp.PacketsReflected != 4900 {
		t.Errorf("Expected PacketsReflected 4900, got %d", resp.PacketsReflected)
	}
}

func TestHandleLicense(t *testing.T) {
	s := NewServer(8080)

	req := httptest.NewRequest(http.MethodGet, "/api/license", nil)
	w := httptest.NewRecorder()

	s.handleLicense(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp LicenseStatus
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have a device hash
	if resp.DeviceHash == "" && s.licenseManager != nil {
		t.Error("Expected non-empty device hash when license manager exists")
	}
}

func TestHandleLicenseActivate(t *testing.T) {
	s := NewServer(8080)

	body := bytes.NewBufferString(`{"licenseKey": "1001-TEST-1234-5678"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/license/activate", body)
	w := httptest.NewRecorder()

	s.handleLicenseActivate(w, req)

	// Response should be valid JSON (success or failure)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleLicenseActivateEmpty(t *testing.T) {
	s := NewServer(8080)

	body := bytes.NewBufferString(`{"licenseKey": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/license/activate", body)
	w := httptest.NewRecorder()

	s.handleLicenseActivate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["success"] != false {
		t.Error("Expected success: false for empty license key")
	}
}

func TestHandleLicenseTrialGet(t *testing.T) {
	s := NewServer(8080)

	req := httptest.NewRequest(http.MethodGet, "/api/license/trial", nil)
	w := httptest.NewRecorder()

	s.handleLicenseTrial(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have 'active' field
	if _, ok := resp["active"]; !ok {
		t.Error("Expected 'active' field in response")
	}
}

func TestHandleLicenseTrialPost(t *testing.T) {
	s := NewServer(8080)

	req := httptest.NewRequest(http.MethodPost, "/api/license/trial", nil)
	w := httptest.NewRecorder()

	s.handleLicenseTrial(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestUpdateStats(t *testing.T) {
	s := NewServer(8080)

	s.UpdateStats(1000, 900, 64000, 57600, 10000.0, 100.5)

	if s.stats.PacketsReceived != 1000 {
		t.Errorf("Expected PacketsReceived 1000, got %d", s.stats.PacketsReceived)
	}
	if s.stats.PacketsSent != 900 {
		t.Errorf("Expected PacketsSent 900, got %d", s.stats.PacketsSent)
	}
	if s.stats.BytesReceived != 64000 {
		t.Errorf("Expected BytesReceived 64000, got %d", s.stats.BytesReceived)
	}
	if s.stats.BytesSent != 57600 {
		t.Errorf("Expected BytesSent 57600, got %d", s.stats.BytesSent)
	}
	if s.stats.CurrentPPS != 10000.0 {
		t.Errorf("Expected CurrentPPS 10000.0, got %f", s.stats.CurrentPPS)
	}
	if s.stats.CurrentMbps != 100.5 {
		t.Errorf("Expected CurrentMbps 100.5, got %f", s.stats.CurrentMbps)
	}
}

func TestServerStructFields(t *testing.T) {
	s := &Server{
		port:       8080,
		mux:        http.NewServeMux(),
		stats:      &Stats{},
		testStatus: "idle",
		startTime:  time.Now(),
		mode:       "test_master",
	}

	if s.port != 8080 {
		t.Errorf("Expected port 8080, got %d", s.port)
	}
	if s.mode != "test_master" {
		t.Errorf("Expected mode 'test_master', got '%s'", s.mode)
	}
}

func TestReflectorConfigStruct(t *testing.T) {
	cfg := ReflectorConfig{
		Profile:         "netally",
		SignatureFilter: []string{"rfc2544", "y1564"},
		OUIFilter:       "00:c0:17",
		PortFilter:      3842,
	}

	if cfg.Profile != "netally" {
		t.Errorf("Expected profile 'netally', got '%s'", cfg.Profile)
	}
	if len(cfg.SignatureFilter) != 2 {
		t.Errorf("Expected 2 signature filters, got %d", len(cfg.SignatureFilter))
	}
}

func TestStatsStruct(t *testing.T) {
	testName := "throughput"
	stats := Stats{
		PacketsReceived: 1000,
		PacketsSent:     900,
		BytesReceived:   64000,
		BytesSent:       57600,
		CurrentPPS:      10000.0,
		CurrentMbps:     100.5,
		Uptime:          3600,
		TestStatus:      "running",
		CurrentTest:     &testName,
	}

	if stats.PacketsReceived != 1000 {
		t.Errorf("Expected PacketsReceived 1000, got %d", stats.PacketsReceived)
	}
	if *stats.CurrentTest != "throughput" {
		t.Errorf("Expected CurrentTest 'throughput', got '%s'", *stats.CurrentTest)
	}
}

func TestReflectorStatsStruct(t *testing.T) {
	stats := ReflectorStats{
		Running:          true,
		PacketsReceived:  5000,
		PacketsReflected: 4900,
		BytesReceived:    320000,
		BytesReflected:   313600,
		TxErrors:         5,
		RxInvalid:        10,
		RatePPS:          50000.0,
		RateMbps:         400.5,
		Uptime:           120.5,
	}
	stats.Signatures.RFC2544 = 100
	stats.Signatures.Y1564 = 50
	stats.Latency.AvgUs = 25.5

	if !stats.Running {
		t.Error("Expected Running true")
	}
	if stats.Signatures.RFC2544 != 100 {
		t.Errorf("Expected RFC2544 100, got %d", stats.Signatures.RFC2544)
	}
	if stats.Latency.AvgUs != 25.5 {
		t.Errorf("Expected Latency.AvgUs 25.5, got %f", stats.Latency.AvgUs)
	}
}

func TestLicenseStatusStruct(t *testing.T) {
	status := LicenseStatus{
		Activated:     true,
		IsTrialMode:   false,
		Tier:          2,
		TierName:      "TestSuite",
		DaysRemaining: 365,
		Features:      []string{"reflector", "rfc2544", "y1564"},
		DeviceHash:    "abc123def456",
		LicenseKey:    "2001-XXXX-XXXX-XXXX",
		Message:       "Licensed: TestSuite",
	}

	if !status.Activated {
		t.Error("Expected Activated true")
	}
	if status.Tier != 2 {
		t.Errorf("Expected Tier 2, got %d", status.Tier)
	}
	if len(status.Features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(status.Features))
	}
}

// Integration tests
func TestServerRoutesRegistered(t *testing.T) {
	s := NewServer(8080)

	routes := []struct {
		path   string
		method string
	}{
		{"/api/health", http.MethodGet},
		{"/api/interfaces", http.MethodGet},
		{"/api/stats", http.MethodGet},
		{"/api/test/start", http.MethodPost},
		{"/api/test/stop", http.MethodPost},
		{"/api/settings", http.MethodGet},
		{"/api/mode", http.MethodGet},
		{"/api/reflector/config", http.MethodGet},
		{"/api/reflector/stats", http.MethodGet},
		{"/api/license", http.MethodGet},
		{"/api/license/trial", http.MethodGet},
	}

	for _, route := range routes {
		t.Run(route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			s.mux.ServeHTTP(w, req)

			// Should not be 404
			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s returned 404", route.method, route.path)
			}
		})
	}
}

func TestConcurrentStatsAccess(t *testing.T) {
	s := NewServer(8080)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			s.UpdateStats(uint64(i), uint64(i), uint64(i*64), uint64(i*64), float64(i*10), float64(i))
		}
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
			w := httptest.NewRecorder()
			s.handleStats(w, req)
		}
	}()

	// Should not panic or deadlock
}

// Benchmark tests
func BenchmarkHandleHealth(b *testing.B) {
	s := NewServer(8080)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()
		s.handleHealth(w, req)
	}
}

func BenchmarkHandleStats(b *testing.B) {
	s := NewServer(8080)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		w := httptest.NewRecorder()
		s.handleStats(w, req)
	}
}

func BenchmarkUpdateStats(b *testing.B) {
	s := NewServer(8080)

	for i := 0; i < b.N; i++ {
		s.UpdateStats(1000, 900, 64000, 57600, 10000.0, 100.5)
	}
}

// Test JSON field tags
func TestJSONFieldTags(t *testing.T) {
	stats := Stats{
		PacketsReceived: 100,
		CurrentPPS:      1000.5,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal Stats: %v", err)
	}

	// Check that camelCase field names are in JSON
	jsonStr := string(data)
	if !strings.Contains(jsonStr, "packetsReceived") {
		t.Error("Expected 'packetsReceived' in JSON output")
	}
	if !strings.Contains(jsonStr, "currentPps") {
		t.Error("Expected 'currentPps' in JSON output")
	}
}
