// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/api"
)

// setupLicenseTestServer creates a server for license handler tests.
func setupLicenseTestServer(t testing.TB) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_AUTH_USERNAME", "licensetest")
	t.Setenv("STEM_AUTH_PASSWORD", "licensepass123")

	s, err := api.NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// getLicenseAuthToken returns an auth token for license tests.
func getLicenseAuthToken(t *testing.T, s *api.Server) string {
	t.Helper()
	body := bytes.NewBufferString(`{"username":"licensetest","password":"licensepass123"}`)
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

// TestHandleLicense_GetSuccess tests GET /api/v1/license.
func TestHandleLicense_GetSuccess(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
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

	// Check expected fields.
	expectedFields := []string{"activated", "isTrialMode", "tier", "tierName", "daysRemaining"}
	for _, field := range expectedFields {
		if _, ok := resp[field]; !ok {
			t.Errorf("Expected field '%s' in license response", field)
		}
	}
}

// TestHandleLicense_MethodNotAllowed tests non-GET methods.
func TestHandleLicense_MethodNotAllowed(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/license", nil)
			// These are unsafe methods on an authenticated endpoint;
			// the CSRF middleware would 403 a naked Bearer request
			// (Wave 1 fail-closed). Provide a valid CSRF token so the
			// handler-level 405 surfaces as the test expects.
			authorizeWithCSRF(t, s, req, token)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleLicenseActivate_MethodNotAllowed tests non-POST methods.
func TestHandleLicenseActivate_MethodNotAllowed(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/license/activate", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			// GET is safe (CSRF middleware bypasses); PUT/DELETE on an
			// authenticated endpoint need a CSRF token to reach the
			// handler's 405 path.
			if method != http.MethodGet {
				authorizeWithCSRF(t, s, req, token)
			}
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleLicenseActivate_EmptyKey tests activation with empty license key.
func TestHandleLicenseActivate_EmptyKey(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)

	body := bytes.NewBufferString(`{"licenseKey":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/activate", body)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should indicate failure.
	if resp["success"] == true {
		t.Error("Expected success=false for empty license key")
	}
}

// TestHandleLicenseActivate_InvalidKey tests activation with invalid license key.
func TestHandleLicenseActivate_InvalidKey(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)

	body := bytes.NewBufferString(`{"licenseKey":"INVALID-KEY-FORMAT"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/activate", body)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should indicate failure.
	if resp["success"] == true {
		t.Error("Expected success=false for invalid license key")
	}
}

// TestHandleLicenseTrial_GetStatus tests GET /api/v1/license/trial.
func TestHandleLicenseTrial_GetStatus(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/license/trial", nil)
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

	if _, ok := resp["active"]; !ok {
		t.Error("Expected 'active' field in trial status response")
	}
}

// TestHandleLicenseTrial_StartTrial tests POST /api/v1/license/trial.
func TestHandleLicenseTrial_StartTrial(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/trial", nil)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	// May succeed or fail depending on trial state.
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleLicenseTrial_MethodNotAllowed tests non-GET/POST methods.
func TestHandleLicenseTrial_MethodNotAllowed(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)
	methods := []string{http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/license/trial", nil)
			authorizeWithCSRF(t, s, req, token)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleLicense_ContentType tests response content type.
func TestHandleLicense_ContentType(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)

	endpoints := []string{"/api/v1/license", "/api/v1/license/trial"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}
		})
	}
}

// TestHandleLicenseNoAuth tests that license endpoints are public (no auth required).
func TestHandleLicenseNoAuth(t *testing.T) {
	s := setupLicenseTestServer(t)

	endpoints := []string{"/api/v1/license", "/api/v1/license/trial"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			// Request without any auth headers.
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			// Should succeed without auth.
			if w.Code == http.StatusUnauthorized {
				t.Errorf("License endpoint %s should not require authentication", endpoint)
			}
		})
	}
}
