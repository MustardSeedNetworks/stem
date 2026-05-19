// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/krisarmstrong/stem/internal/api"
	"github.com/krisarmstrong/stem/internal/netif"
)

// Test constants for repeated strings.
const (
	testModeTestMaster  = "test_master"
	testModeReflector   = "reflector"
	testIfaceEth0       = "eth0"
	testStatusError     = "error"
	testStatusRunning   = "running"
	testStatusStarting  = "starting"
	testStatusCancelled = "cancelled"
	testModuleBenchmark = "benchmark"
	testProfileNetally  = "netally"
	testTypeThroughput  = "throughput"
	testUsername        = "testadmin"
	testPassword        = "testpass123"
)

// setupTestServer creates a server with test credentials configured.
func setupTestServer(t testing.TB) *api.Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", testUsername)
	t.Setenv("STEM_AUTH_PASSWORD", testPassword)

	s, err := api.NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

func loginToken(t *testing.T, s *api.Server) string {
	t.Helper()
	body := bytes.NewBufferString(fmt.Sprintf(`{"username":"%s","password":"%s"}`, testUsername, testPassword))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected login status 200, got %d", w.Code)
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

func TestHandleAuthLogin(t *testing.T) {
	s := setupTestServer(t)
	token := loginToken(t, s)
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestHandleAuthLoginFailure(t *testing.T) {
	s := setupTestServer(t)
	body := bytes.NewBufferString(`{"username":"admin","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 for invalid login, got %d", w.Code)
	}
}

func TestNewServer(t *testing.T) {
	s := setupTestServer(t)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestHandleHealth(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", resp["status"])
	}
	if resp["product"] != "The Stem" {
		t.Errorf("Expected product 'The Stem', got '%v'", resp["product"])
	}
	if resp["company"] != "Mustard Seed Networks" {
		t.Errorf("Expected company 'Mustard Seed Networks', got '%v'", resp["company"])
	}
}

func TestHandleHealthMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleInterfaces(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Response should be valid JSON array.
	var resp []any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleStats(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleTestStartUnknownType(t *testing.T) {
	s := setupTestServer(t)

	body := strings.NewReader(`{"testType": "nonexistent_test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
	token := loginToken(t, s)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleTestStartUnauthorized(t *testing.T) {
	s := setupTestServer(t)

	body := strings.NewReader(`{"testType": "throughput"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 without auth, got %d", w.Code)
	}
}

func TestHandleTestStartNoInterface(t *testing.T) {
	// This test relies on no interface being selected.
	// In environments with auto-selected interface, this may not apply.
	t.Skip("This test requires explicit interface clearing which is internal state")
}

func TestHandleSettingsGet(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleSettingsPost(t *testing.T) {
	s := setupTestServer(t)

	// Get a real interface name from the system.
	ifaces, err := netif.DetectInterfaces()
	if err != nil || len(ifaces) == 0 {
		t.Skip("No network interfaces available for testing")
	}
	testIface := ifaces[0].Name

	body := bytes.NewBufferString(fmt.Sprintf(`{"interface": "%s"}`, testIface))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSettingsInvalidInterface(t *testing.T) {
	s := setupTestServer(t)

	// Use an interface name that definitely doesn't exist.
	body := bytes.NewBufferString(`{"interface": "nonexistent_interface_xyz123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid interface, got %d", w.Code)
	}
}

func TestHandleSettingsInvalidJSON(t *testing.T) {
	s := setupTestServer(t)

	body := bytes.NewBufferString(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleModeGet(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mode", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleModePost(t *testing.T) {
	s := setupTestServer(t)
	// Force the platform probe to "supported" so this test does not
	// depend on the host's CGO / Linux availability. The
	// unsupported-platform 403 path is covered in
	// handlers_mode_test.go.
	s.UseReflectorAvailabilityForTest(func() (bool, string) { return true, "" })
	t.Cleanup(func() { s.UseReflectorAvailabilityForTest(nil) })

	body := bytes.NewBufferString(`{"mode": "` + testModeReflector + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mode", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleModePostInvalid(t *testing.T) {
	s := setupTestServer(t)

	body := bytes.NewBufferString(`{"mode": "invalid_mode"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mode", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleReflectorConfigGet(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/config", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleReflectorConfigPost(t *testing.T) {
	s := setupTestServer(t)

	body := bytes.NewBufferString(`{"profile": "` + testProfileNetally + `", "portFilter": 9999}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleReflectorConfigPostInvalidProfile(t *testing.T) {
	s := setupTestServer(t)

	body := bytes.NewBufferString(`{"profile": "invalid_profile"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleReflectorStats(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/stats", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleLicense(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleLicenseActivate(t *testing.T) {
	s := setupTestServer(t)

	body := bytes.NewBufferString(`{"licenseKey": "1001-TEST-1234-5678"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/activate", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	// Response should be valid JSON (success or failure).
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestHandleLicenseActivateEmpty(t *testing.T) {
	s := setupTestServer(t)

	body := bytes.NewBufferString(`{"licenseKey": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/activate", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["success"] != false {
		t.Error("Expected success: false for empty license key")
	}
}

func TestHandleLicenseTrialGet(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/license/trial", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have 'active' field.
	if _, ok := resp["active"]; !ok {
		t.Error("Expected 'active' field in response")
	}
}

func TestHandleLicenseTrialPost(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/trial", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

// Integration tests.
func TestServerRoutesRegistered(t *testing.T) {
	s := setupTestServer(t)

	routes := []struct {
		path   string
		method string
	}{
		{"/api/v1/health", http.MethodGet},
		{"/api/v1/interfaces", http.MethodGet},
		{"/api/v1/stats", http.MethodGet},
		{"/api/v1/settings", http.MethodGet},
		{"/api/v1/mode", http.MethodGet},
		{"/api/v1/reflector/config", http.MethodGet},
		{"/api/v1/reflector/stats", http.MethodGet},
		{"/api/v1/license", http.MethodGet},
		{"/api/v1/license/trial", http.MethodGet},
	}

	for _, route := range routes {
		t.Run(route.path, func(t *testing.T) {
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

// Module API tests.
func TestHandleModules(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have modules array.
	modules, ok := resp["modules"].([]any)
	if !ok {
		t.Fatal("Expected 'modules' array in response")
	}

	// Should have all 6 modules (including reflector).
	if len(modules) != 6 {
		t.Errorf("Expected 6 modules, got %d", len(modules))
	}

	// Should have count field.
	count, ok := resp["count"].(float64)
	if !ok {
		t.Fatal("Expected 'count' field in response")
	}
	if int(count) != 6 {
		t.Errorf("Expected count 6, got %d", int(count))
	}
}

func TestHandleModulesMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/modules", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleModuleByName(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/benchmark", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check module fields.
	if resp["name"] != "benchmark" {
		t.Errorf("Expected name 'benchmark', got '%v'", resp["name"])
	}
	if resp["displayName"] != "Benchmark" {
		t.Errorf("Expected displayName 'Benchmark', got '%v'", resp["displayName"])
	}
	if resp["color"] != "#dc2626" {
		t.Errorf("Expected color '#dc2626', got '%v'", resp["color"])
	}
	if resp["standard"] != "RFC 2544" {
		t.Errorf("Expected standard 'RFC 2544', got '%v'", resp["standard"])
	}

	// Should have tests array.
	tests, ok := resp["tests"].([]any)
	if !ok {
		t.Fatal("Expected 'tests' array in response")
	}
	if len(tests) == 0 {
		t.Error("Expected at least one test type")
	}
}

func TestHandleModuleByNameNotFound(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/nonexistent", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleModuleByNameTests(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/benchmark/tests", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have module name.
	if resp["module"] != "benchmark" {
		t.Errorf("Expected module 'benchmark', got '%v'", resp["module"])
	}

	// Should have tests array.
	tests, ok := resp["tests"].([]any)
	if !ok {
		t.Fatal("Expected 'tests' array in response")
	}

	// Benchmark should have RFC 2544 tests.
	expectedTests := []string{"throughput", "latency", "frame_loss", "back_to_back", "system_recovery", "reset"}
	if len(tests) != len(expectedTests) {
		t.Errorf("Expected %d tests, got %d", len(expectedTests), len(tests))
	}

	// Should have count.
	count, ok := resp["count"].(float64)
	if !ok {
		t.Fatal("Expected 'count' field in response")
	}
	if int(count) != len(expectedTests) {
		t.Errorf("Expected count %d, got %d", len(expectedTests), int(count))
	}
}

func TestHandleModuleByNameMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/modules/benchmark", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestModuleRoutesRegistered(t *testing.T) {
	s := setupTestServer(t)

	routes := []struct {
		path   string
		method string
	}{
		{"/api/v1/modules", http.MethodGet},
		{"/api/v1/modules/benchmark", http.MethodGet},
		{"/api/v1/modules/benchmark/tests", http.MethodGet},
	}

	for _, route := range routes {
		t.Run(route.path, func(t *testing.T) {
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

// Benchmark tests.
func BenchmarkHandleHealth(b *testing.B) {
	s := setupTestServer(b)

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}

func BenchmarkHandleStats(b *testing.B) {
	s := setupTestServer(b)

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
