// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/api"
	"github.com/krisarmstrong/stem/internal/netif"
	"github.com/krisarmstrong/stem/internal/services/modtypes"
)

// fakeAPIExecutor is a no-op testExecutor used to drive the Server's
// test-start path without invoking the real cgo dataplane.
type fakeAPIExecutor struct {
	closeCalls atomic.Uint32
}

func (f *fakeAPIExecutor) Close() {
	f.closeCalls.Add(1)
}

func (f *fakeAPIExecutor) Execute(
	testType string,
	_ *modtypes.TestConfig,
) (*modtypes.Result, error) {
	return &modtypes.Result{
		TestType:   testType,
		ModuleName: "",
		Success:    true,
		Error:      "",
		Data:       nil,
	}, nil
}

// resetServerTestState gives any in-flight runModuleTest goroutine from
// a prior test-start request a moment to finish, then forcibly clears
// the Server's transient test state so the next request sees a clean
// status. The fakeAPIExecutor returns immediately, so a short sleep is
// sufficient.
func resetServerTestState(t *testing.T, s *api.Server) {
	t.Helper()
	time.Sleep(20 * time.Millisecond)
	s.ResetTestStateForTest()
}

// Test constants for testing handlers.
const (
	testingTestUsername = "testingadmin"
	testingTestPassword = "testingpass123"
)

// setupTestingTestServer creates a server for testing handler tests.
func setupTestingTestServer(t testing.TB) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_AUTH_USERNAME", testingTestUsername)
	t.Setenv("STEM_AUTH_PASSWORD", testingTestPassword)

	s, err := api.NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// getTestingAuthToken returns an auth token for testing handlers.
func getTestingAuthToken(t *testing.T, s *api.Server) string {
	t.Helper()
	body := bytes.NewBufferString(
		fmt.Sprintf(`{"username":"%s","password":"%s"}`, testingTestUsername, testingTestPassword),
	)
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

// TestHandleTestStop tests the POST /api/v1/test/stop endpoint.
func TestHandleTestStop(t *testing.T) {
	s := setupTestingTestServer(t)

	t.Run("stop with no test running", func(t *testing.T) {
		token := getTestingAuthToken(t, s)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/stop", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		// Should return error when no test is running.
		if w.Code != http.StatusBadRequest {
			t.Errorf(
				"Expected status 400 when no test running, got %d: %s",
				w.Code,
				w.Body.String(),
			)
		}
	})

	t.Run("stop unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/stop", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 without auth, got %d", w.Code)
		}
	})

	t.Run("stop method not allowed", func(t *testing.T) {
		token := getTestingAuthToken(t, s)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/test/stop", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("stop with invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/stop", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

// TestHandleTestResult tests the GET /api/v1/test/result endpoint.
func TestHandleTestResult(t *testing.T) {
	s := setupTestingTestServer(t)

	t.Run("get result with no test run", func(t *testing.T) {
		token := getTestingAuthToken(t, s)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/test/result", nil)
		req.Header.Set("Authorization", "Bearer "+token)
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

		// Should have status field.
		if _, ok := resp["status"]; !ok {
			t.Error("Expected 'status' field in response")
		}

		// Should have message indicating no result.
		if resp["message"] == nil {
			t.Error("Expected 'message' field in response")
		}
	})

	t.Run("get result unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test/result", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 without auth, got %d", w.Code)
		}
	})

	t.Run("get result method not allowed", func(t *testing.T) {
		token := getTestingAuthToken(t, s)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/result", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("get result content type", func(t *testing.T) {
		token := getTestingAuthToken(t, s)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/test/result", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})
}

// TestHandleTestStartWithInterface tests test start with interface parameter.
func TestHandleTestStartWithInterface(t *testing.T) {
	s := setupTestingTestServer(t)

	t.Run("start with explicit interface", func(t *testing.T) {
		// Get a real interface name from the system.
		ifaces, err := netif.DetectInterfaces()
		if err != nil || len(ifaces) == 0 {
			t.Skip("No network interfaces available for testing")
		}
		testIface := ifaces[0].Name
		token := getTestingAuthToken(t, s)

		body := bytes.NewBufferString(
			fmt.Sprintf(`{"testType":"rfc2544_throughput","interface":"%s"}`, testIface),
		)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		// May succeed or fail depending on platform, but should not be 400 for missing interface.
		if w.Code == http.StatusBadRequest {
			var resp map[string]any
			unmarshalErr := json.Unmarshal(w.Body.Bytes(), &resp)
			if unmarshalErr == nil {
				msg, _ := resp["message"].(string)
				if msg == "No network interface specified" {
					t.Error("Interface was specified but error says no interface")
				}
			}
		}
	})

	t.Run("start with nonexistent interface", func(t *testing.T) {
		token := getTestingAuthToken(t, s)

		body := bytes.NewBufferString(
			`{"testType":"rfc2544_throughput","interface":"nonexistent_iface_xyz123"}`,
		)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		// Should fail because interface doesn't exist.
		if w.Code != http.StatusBadRequest {
			t.Errorf(
				"Expected status 400 for nonexistent interface, got %d: %s",
				w.Code,
				w.Body.String(),
			)
		}
	})
}

// TestHandleTestStartValidation tests various test start validation scenarios.
func TestHandleTestStartValidation(t *testing.T) {
	s := setupTestingTestServer(t)

	t.Run("start with invalid JSON", func(t *testing.T) {
		token := getTestingAuthToken(t, s)

		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
		}
	})

	t.Run("start with unknown fields in JSON", func(t *testing.T) {
		token := getTestingAuthToken(t, s)

		body := bytes.NewBufferString(`{"testType":"rfc2544_throughput","unknownField":"value"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		// Should reject unknown fields (strict JSON parsing).
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for unknown fields, got %d", w.Code)
		}
	})

	t.Run("start with empty body", func(t *testing.T) {
		token := getTestingAuthToken(t, s)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		// Empty body should result in 400.
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty body, got %d", w.Code)
		}
	})

	t.Run("start with supported test types", func(t *testing.T) {
		// This subtest verifies the API's test-start path accepts the
		// supported RFC 2544 test types. It uses a fake executor (via
		// Server.UseTestExecutorResolver) so the request never reaches
		// the real cgo dataplane — earlier this test was gated under
		// -short because it SIGSEGV'd in CI runners without raw-socket
		// capabilities.
		token := getTestingAuthToken(t, s)

		// Get a real interface.
		ifaces, err := netif.DetectInterfaces()
		if err != nil || len(ifaces) == 0 {
			t.Skip("No network interfaces available for testing")
		}
		testIface := ifaces[0].Name

		s.UseTestExecutorResolver(func(_ string) (api.TestExecutorFactory, bool) {
			return func(_ string) (api.TestExecutor, error) {
				return &fakeAPIExecutor{}, nil
			}, true
		})
		t.Cleanup(func() { s.UseTestExecutorResolver(nil) })

		testTypes := []string{
			"rfc2544_throughput",
			"rfc2544_latency",
			"rfc2544_frame_loss",
			"rfc2544_back_to_back",
		}

		for _, testType := range testTypes {
			t.Run(testType, func(t *testing.T) {
				// Reset server test state between subtests so each
				// request starts from a clean status.
				resetServerTestState(t, s)

				body := bytes.NewBufferString(
					fmt.Sprintf(`{"testType":"%s","interface":"%s"}`, testType, testIface),
				)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
				req.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()

				s.ServeHTTP(w, req)

				// Should not return 400 for unknown test type.
				if w.Code == http.StatusBadRequest {
					var resp map[string]any
					_ = json.Unmarshal(w.Body.Bytes(), &resp)
					msg, _ := resp["message"].(string)
					if msg == "Unknown or unsupported test type" {
						t.Errorf("Test type '%s' should be supported", testType)
					}
				}
			})
		}
	})
}

// TestHandleTestStartOptionalParameters tests optional parameters.
func TestHandleTestStartOptionalParameters(t *testing.T) {
	s := setupTestingTestServer(t)

	t.Run("reject legacy frameSize field", func(t *testing.T) {
		ifaces, err := netif.DetectInterfaces()
		if err != nil || len(ifaces) == 0 {
			t.Skip("No network interfaces available for testing")
		}
		token := getTestingAuthToken(t, s)
		testIface := ifaces[0].Name

		body := bytes.NewBufferString(
			fmt.Sprintf(`{"testType":"throughput","interface":"%s","frameSize":1518}`, testIface),
		)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		msg, _ := resp["message"].(string)
		if msg != "Invalid JSON in request body" {
			t.Fatalf("expected invalid JSON error, got %q", msg)
		}
	})

	t.Run("reject legacy duration field", func(t *testing.T) {
		ifaces, err := netif.DetectInterfaces()
		if err != nil || len(ifaces) == 0 {
			t.Skip("No network interfaces available for testing")
		}
		token := getTestingAuthToken(t, s)
		testIface := ifaces[0].Name

		body := bytes.NewBufferString(
			fmt.Sprintf(`{"testType":"throughput","interface":"%s","duration":60}`, testIface),
		)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		msg, _ := resp["message"].(string)
		if msg != "Invalid JSON in request body" {
			t.Fatalf("expected invalid JSON error, got %q", msg)
		}
	})
}

// TestTestEndpointsRoutes verifies test endpoint routes are registered.
func TestTestEndpointsRoutes(t *testing.T) {
	s := setupTestingTestServer(t)

	routes := []struct {
		path   string
		method string
	}{
		{"/api/v1/test/start", http.MethodPost},
		{"/api/v1/test/stop", http.MethodPost},
		{"/api/v1/test/result", http.MethodGet},
	}

	for _, route := range routes {
		t.Run(route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			// Should not be 404 (route exists).
			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s returned 404", route.method, route.path)
			}
		})
	}
}

// BenchmarkHandleTestResult benchmarks the test result endpoint.
func BenchmarkHandleTestResult(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := api.NewServer(8080)
	if err != nil {
		b.Fatalf("NewServer() error: %v", err)
	}
	b.Cleanup(func() { _ = s.Shutdown() })

	// Get token once.
	body := bytes.NewBufferString(`{"username":"benchuser","password":"benchpass123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	loginW := httptest.NewRecorder()
	s.ServeHTTP(loginW, loginReq)

	var resp map[string]any
	_ = json.Unmarshal(loginW.Body.Bytes(), &resp)
	token, _ := resp["token"].(string)

	b.ResetTimer()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test/result", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
