// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/netif"
	"github.com/krisarmstrong/stem/internal/server"
)

// setupSettingsTestServer creates a server for settings handler tests.
func setupSettingsTestServer(t testing.TB) *server.Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", "settingstest")
	t.Setenv("STEM_AUTH_PASSWORD", "settingspass123")

	s, err := server.NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	return s
}

// getSettingsAuthToken returns an auth token for settings tests.
func getSettingsAuthToken(t *testing.T, s *server.Server) string {
	t.Helper()
	body := bytes.NewBufferString(`{"username":"settingstest","password":"settingspass123"}`)
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

// TestHandleSettings_GetSuccess tests GET /api/v1/settings.
func TestHandleSettings_GetSuccess(t *testing.T) {
	s := setupSettingsTestServer(t)
	token := getSettingsAuthToken(t, s)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
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

	if _, ok := resp["mode"]; !ok {
		t.Error("Expected 'mode' field in settings response")
	}
	if _, ok := resp["interface"]; !ok {
		t.Error("Expected 'interface' field in settings response")
	}
}

// TestHandleSettings_PostUpdateInterface tests POST /api/v1/settings with interface.
func TestHandleSettings_PostUpdateInterface(t *testing.T) {
	s := setupSettingsTestServer(t)

	// Get a real interface name.
	ifaces, err := netif.DetectInterfaces()
	if err != nil || len(ifaces) == 0 {
		t.Skip("No network interfaces available for testing")
	}
	testIface := ifaces[0].Name
	token := getSettingsAuthToken(t, s)

	body := bytes.NewBufferString(fmt.Sprintf(`{"interface":"%s"}`, testIface))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleSettings_PostInvalidInterface tests POST with nonexistent interface.
func TestHandleSettings_PostInvalidInterface(t *testing.T) {
	s := setupSettingsTestServer(t)
	token := getSettingsAuthToken(t, s)

	body := bytes.NewBufferString(`{"interface":"nonexistent_iface_xyz"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid interface, got %d", w.Code)
	}
}

// TestHandleSettings_MethodNotAllowed tests non-GET/POST methods.
func TestHandleSettings_MethodNotAllowed(t *testing.T) {
	s := setupSettingsTestServer(t)
	token := getSettingsAuthToken(t, s)
	methods := []string{http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/settings", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleMode_GetSuccess tests GET /api/v1/mode.
func TestHandleMode_GetSuccess(t *testing.T) {
	s := setupSettingsTestServer(t)
	token := getSettingsAuthToken(t, s)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mode", nil)
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

	if _, ok := resp["mode"]; !ok {
		t.Error("Expected 'mode' field in response")
	}
}

// TestHandleMode_PostSwitchMode tests POST /api/v1/mode to switch modes.
func TestHandleMode_PostSwitchMode(t *testing.T) {
	s := setupSettingsTestServer(t)
	token := getSettingsAuthToken(t, s)

	t.Run("switch to reflector", func(t *testing.T) {
		body := bytes.NewBufferString(`{"mode":"reflector"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mode", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("switch to test_master", func(t *testing.T) {
		body := bytes.NewBufferString(`{"mode":"test_master"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mode", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

// TestHandleMode_PostInvalidMode tests POST with invalid mode value.
func TestHandleMode_PostInvalidMode(t *testing.T) {
	s := setupSettingsTestServer(t)
	token := getSettingsAuthToken(t, s)

	body := bytes.NewBufferString(`{"mode":"invalid_mode"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mode", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid mode, got %d", w.Code)
	}
}

// TestHandleMode_MethodNotAllowed tests non-GET/POST methods.
func TestHandleMode_MethodNotAllowed(t *testing.T) {
	s := setupSettingsTestServer(t)
	token := getSettingsAuthToken(t, s)
	methods := []string{http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/mode", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

// TestHandleSettings_ContentType tests that response has correct content type.
func TestHandleSettings_ContentType(t *testing.T) {
	s := setupSettingsTestServer(t)
	token := getSettingsAuthToken(t, s)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}
