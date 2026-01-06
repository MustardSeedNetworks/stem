// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/auth"
)

// setupClaimsTestServer creates a server for claims tests.
func setupClaimsTestServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", "claimstest")
	t.Setenv("STEM_AUTH_PASSWORD", "claimspass123")

	s, err := NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	return s
}

// TestExtractClaims_ValidToken tests extractClaims with a valid token.
func TestExtractClaims_ValidToken(t *testing.T) {
	s := setupClaimsTestServer(t)
	token, _, authErr := s.authManager.AuthenticateWithRefresh(context.TODO(), "claimstest", "claimspass123")
	if authErr != nil {
		t.Fatalf("Failed to get token: %v", authErr)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	claims, claimsErr := s.extractClaims(req)
	if claimsErr != nil {
		t.Errorf("extractClaims() error = %v", claimsErr)
	}
	if claims == nil {
		t.Error("extractClaims() returned nil claims")
	}
	if claims != nil && claims.Username != "claimstest" {
		t.Errorf("Expected username 'claimstest', got '%s'", claims.Username)
	}
}

// TestExtractClaims_InvalidHeaders tests extractClaims with various invalid headers.
func TestExtractClaims_InvalidHeaders(t *testing.T) {
	s := setupClaimsTestServer(t)

	tests := []struct {
		name   string
		header string
	}{
		{"missing header", ""},
		{"non-bearer auth", "Basic sometoken"},
		{"invalid token", "Bearer invalid.token.string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			claims, claimsErr := s.extractClaims(req)
			if claimsErr == nil {
				t.Errorf("extractClaims() should return error for %s", tt.name)
			}
			if claims != nil {
				t.Errorf("extractClaims() should return nil claims for %s", tt.name)
			}
		})
	}
}

// TestExtractClaims_BearerWithSpaces tests that extra spaces are handled.
func TestExtractClaims_BearerWithSpaces(t *testing.T) {
	s := setupClaimsTestServer(t)
	token, _, authErr := s.authManager.AuthenticateWithRefresh(context.TODO(), "claimstest", "claimspass123")
	if authErr != nil {
		t.Fatalf("Failed to get token: %v", authErr)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "  Bearer  "+token)

	claims, claimsErr := s.extractClaims(req)
	if claimsErr != nil {
		t.Errorf("extractClaims() error = %v", claimsErr)
	}
	if claims == nil {
		t.Error("extractClaims() returned nil claims")
	}
}

// TestAuditAuthFailure tests the auditAuthFailure method.
func TestAuditAuthFailure(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "audituser")
	t.Setenv("STEM_AUTH_PASSWORD", "auditpass123")

	s, err := NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}

	// These tests verify that auditAuthFailure doesn't panic for various error types.
	// The actual logging is verified by observing that no panic occurs.

	t.Run("token expired error", func(_ *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		// Should not panic.
		s.auditAuthFailure(req, auth.ErrTokenExpired)
	})

	t.Run("token revoked error", func(_ *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"

		// Should not panic.
		s.auditAuthFailure(req, auth.ErrTokenRevoked)
	})

	t.Run("invalid token error", func(_ *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.3:12345"

		// Should not panic.
		s.auditAuthFailure(req, auth.ErrInvalidToken)
	})

	t.Run("missing auth token error", func(_ *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.4:12345"

		// Should not panic.
		s.auditAuthFailure(req, errMissingAuthToken)
	})

	t.Run("invalid auth header error", func(_ *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.5:12345"

		// Should not panic.
		s.auditAuthFailure(req, errInvalidAuthHeader)
	})

	t.Run("generic error", func(_ *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.6:12345"

		// Should not panic.
		s.auditAuthFailure(req, auth.ErrInvalidCredentials)
	})
}

// setupCorsTestHandler creates a handler and wraps it with corsMiddleware.
func setupCorsTestHandler() http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return corsMiddleware(handler)
}

// TestCorsMiddleware_NoOrigin tests corsMiddleware without an Origin header.
func TestCorsMiddleware_NoOrigin(t *testing.T) {
	wrapped := setupCorsTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestCorsMiddleware_LocalhostOrigin tests corsMiddleware with localhost origin.
func TestCorsMiddleware_LocalhostOrigin(t *testing.T) {
	wrapped := setupCorsTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:8080" {
		t.Errorf("Expected Access-Control-Allow-Origin 'http://localhost:8080', got '%s'", allowOrigin)
	}
}

// TestCorsMiddleware_LoopbackOrigin tests corsMiddleware with 127.0.0.1 origin.
func TestCorsMiddleware_LoopbackOrigin(t *testing.T) {
	wrapped := setupCorsTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://127.0.0.1:3000" {
		t.Errorf("Expected Access-Control-Allow-Origin 'http://127.0.0.1:3000', got '%s'", allowOrigin)
	}
}

// TestCorsMiddleware_IPv6Origin tests corsMiddleware with IPv6 localhost origin.
func TestCorsMiddleware_IPv6Origin(t *testing.T) {
	wrapped := setupCorsTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://[::1]:8080")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestCorsMiddleware_ExternalOriginBlocked tests corsMiddleware blocks external origins.
func TestCorsMiddleware_ExternalOriginBlocked(t *testing.T) {
	wrapped := setupCorsTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for external origin, got %d", w.Code)
	}
}

// TestCorsMiddleware_PreflightOptions tests corsMiddleware handles OPTIONS preflight.
func TestCorsMiddleware_PreflightOptions(t *testing.T) {
	wrapped := setupCorsTestHandler()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", w.Code)
	}

	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Error("Expected Access-Control-Allow-Methods header")
	}

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders == "" {
		t.Error("Expected Access-Control-Allow-Headers header")
	}

	maxAge := w.Header().Get("Access-Control-Max-Age")
	if maxAge != "3600" {
		t.Errorf("Expected Access-Control-Max-Age '3600', got '%s'", maxAge)
	}
}

// TestCorsMiddleware_BypassAttemptBlocked tests corsMiddleware blocks bypass attempts.
func TestCorsMiddleware_BypassAttemptBlocked(t *testing.T) {
	wrapped := setupCorsTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost.evil.com")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for bypass attempt, got %d", w.Code)
	}
}

// TestHandleAPIRedirect tests the handleAPIRedirect function.
func TestHandleAPIRedirect(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "redirectuser")
	t.Setenv("STEM_AUTH_PASSWORD", "redirectpass123")

	s, err := NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}

	t.Run("redirect legacy health endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()

		s.handleAPIRedirect(w, req)

		if w.Code != http.StatusPermanentRedirect {
			t.Errorf("Expected status 308, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/api/v1/health" {
			t.Errorf("Expected redirect to '/api/v1/health', got '%s'", location)
		}
	})

	t.Run("redirect with query string", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/interfaces?filter=eth", nil)
		w := httptest.NewRecorder()

		s.handleAPIRedirect(w, req)

		if w.Code != http.StatusPermanentRedirect {
			t.Errorf("Expected status 308, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/api/v1/interfaces?filter=eth" {
			t.Errorf("Expected redirect with query string, got '%s'", location)
		}
	})

	t.Run("no redirect for already versioned path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		s.handleAPIRedirect(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 for already versioned path, got %d", w.Code)
		}
	})

	t.Run("no redirect for non-api path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/something/else", nil)
		w := httptest.NewRecorder()

		s.handleAPIRedirect(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 for non-api path, got %d", w.Code)
		}
	})

	t.Run("redirect preserves method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/test/start", nil)
		w := httptest.NewRecorder()

		s.handleAPIRedirect(w, req)

		// 308 Permanent Redirect preserves method.
		if w.Code != http.StatusPermanentRedirect {
			t.Errorf("Expected status 308, got %d", w.Code)
		}
	})
}

// TestWriteJSON tests the writeJSON function.
func TestWriteJSON(t *testing.T) {
	t.Run("successful JSON encoding", func(t *testing.T) {
		w := httptest.NewRecorder()

		data := map[string]string{"key": "value"}
		writeJSON(w, data)

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
		}

		expected := `{"key":"value"}` + "\n"
		if w.Body.String() != expected {
			t.Errorf("Expected body '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("JSON encoding with struct", func(t *testing.T) {
		w := httptest.NewRecorder()

		data := StatusResponse{Status: "ok"}
		writeJSON(w, data)

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
		}

		expected := `{"status":"ok"}` + "\n"
		if w.Body.String() != expected {
			t.Errorf("Expected body '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("JSON encoding with nil", func(t *testing.T) {
		w := httptest.NewRecorder()

		writeJSON(w, nil)

		// nil should encode as "null".
		expected := "null\n"
		if w.Body.String() != expected {
			t.Errorf("Expected body '%s', got '%s'", expected, w.Body.String())
		}
	})
}

// TestSafeIntToUint16 tests the safeIntToUint16 function.
func TestSafeIntToUint16(t *testing.T) {
	tests := []struct {
		name   string
		input  int
		want   uint16
		wantOK bool
	}{
		{
			name:   "zero",
			input:  0,
			want:   0,
			wantOK: true,
		},
		{
			name:   "positive valid",
			input:  1000,
			want:   1000,
			wantOK: true,
		},
		{
			name:   "max uint16",
			input:  65535,
			want:   65535,
			wantOK: true,
		},
		{
			name:   "negative",
			input:  -1,
			want:   0,
			wantOK: false,
		},
		{
			name:   "over max uint16",
			input:  65536,
			want:   0,
			wantOK: false,
		},
		{
			name:   "large positive",
			input:  100000,
			want:   0,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := safeIntToUint16(tt.input)
			if ok != tt.wantOK {
				t.Errorf("safeIntToUint16(%d) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("safeIntToUint16(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestExtractBearerToken tests the extractBearerToken function.
func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{
			name:    "valid bearer token",
			header:  "Bearer mytoken123",
			want:    "mytoken123",
			wantErr: false,
		},
		{
			name:    "bearer lowercase",
			header:  "bearer mytoken123",
			want:    "mytoken123",
			wantErr: false,
		},
		{
			name:    "bearer mixed case",
			header:  "BEARER mytoken123",
			want:    "mytoken123",
			wantErr: false,
		},
		{
			name:    "empty header",
			header:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			header:  "   ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "basic auth",
			header:  "Basic sometoken",
			want:    "",
			wantErr: true,
		},
		{
			name:    "just bearer",
			header:  "Bearer",
			want:    "",
			wantErr: true,
		},
		{
			name:    "bearer with leading spaces",
			header:  "  Bearer mytoken123",
			want:    "mytoken123",
			wantErr: false,
		},
		{
			name:    "too many parts",
			header:  "Bearer token extra",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractBearerToken(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractBearerToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestApiVersionMiddleware tests the apiVersionMiddleware function.
func TestApiVersionMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := apiVersionMiddleware(handler)

	t.Run("adds version header for API path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		version := w.Header().Get(APIVersionHeader)
		if version != APIVersion {
			t.Errorf("Expected %s header '%s', got '%s'", APIVersionHeader, APIVersion, version)
		}
	})

	t.Run("adds version header for legacy API path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		version := w.Header().Get(APIVersionHeader)
		if version != APIVersion {
			t.Errorf("Expected %s header '%s', got '%s'", APIVersionHeader, APIVersion, version)
		}
	})

	t.Run("no version header for non-API path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		version := w.Header().Get(APIVersionHeader)
		if version != "" {
			t.Errorf("Expected no %s header, got '%s'", APIVersionHeader, version)
		}
	})

	t.Run("no version header for root path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		version := w.Header().Get(APIVersionHeader)
		if version != "" {
			t.Errorf("Expected no %s header, got '%s'", APIVersionHeader, version)
		}
	})
}

// TestDecodeJSONStrict_Success tests decodeJSONStrict with valid input.
func TestDecodeJSONStrict_Success(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "jsontest")
	t.Setenv("STEM_AUTH_PASSWORD", "jsonpass123")

	var data struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	body := `{"name":"test","value":42}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	ok := decodeJSONStrict(w, req, &data, maxRequestBodySize)
	if !ok {
		t.Error("decodeJSONStrict() returned false for valid JSON")
	}

	if data.Name != "test" || data.Value != 42 {
		t.Errorf("decodeJSONStrict() parsed incorrectly: %+v", data)
	}
}

// TestDecodeJSONStrict_InvalidJSON tests decodeJSONStrict with invalid JSON.
func TestDecodeJSONStrict_InvalidJSON(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "jsontest")
	t.Setenv("STEM_AUTH_PASSWORD", "jsonpass123")

	var data struct {
		Name string `json:"name"`
	}

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	ok := decodeJSONStrict(w, req, &data, maxRequestBodySize)
	if ok {
		t.Error("decodeJSONStrict() should return false for invalid JSON")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestDecodeJSONStrict_UnknownFields tests decodeJSONStrict rejects unknown fields.
func TestDecodeJSONStrict_UnknownFields(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "jsontest")
	t.Setenv("STEM_AUTH_PASSWORD", "jsonpass123")

	var data struct {
		Name string `json:"name"`
	}

	body := `{"name":"test","unknown":"field"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	ok := decodeJSONStrict(w, req, &data, maxRequestBodySize)
	if ok {
		t.Error("decodeJSONStrict() should return false for unknown fields")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestDecodeJSONStrict_EmptyBody tests decodeJSONStrict with empty body.
func TestDecodeJSONStrict_EmptyBody(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "jsontest")
	t.Setenv("STEM_AUTH_PASSWORD", "jsonpass123")

	var data struct {
		Name string `json:"name"`
	}

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	ok := decodeJSONStrict(w, req, &data, maxRequestBodySize)
	if ok {
		t.Error("decodeJSONStrict() should return false for empty body")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestNewServer_MissingCredentials tests NewServer with missing credentials.
func TestNewServer_MissingCredentials(t *testing.T) {
	// Clear any existing credentials.
	t.Setenv("STEM_AUTH_USERNAME", "")
	t.Setenv("STEM_AUTH_PASSWORD", "")

	_, err := NewServer(8080)
	if err == nil {
		t.Error("NewServer() should return error with missing credentials")
	}
}

// TestNewServer_PartialCredentials tests NewServer with partial credentials.
func TestNewServer_PartialCredentials(t *testing.T) {
	t.Run("username only", func(t *testing.T) {
		t.Setenv("STEM_AUTH_USERNAME", "user")
		t.Setenv("STEM_AUTH_PASSWORD", "")

		_, err := NewServer(8080)
		if err == nil {
			t.Error("NewServer() should return error with username only")
		}
	})

	t.Run("password only", func(t *testing.T) {
		t.Setenv("STEM_AUTH_USERNAME", "")
		t.Setenv("STEM_AUTH_PASSWORD", "pass123")

		_, err := NewServer(8080)
		if err == nil {
			t.Error("NewServer() should return error with password only")
		}
	})
}

// TestStatusResponse tests the StatusResponse struct.
func TestStatusResponse(t *testing.T) {
	resp := StatusResponse{Status: "ok"}

	if resp.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", resp.Status)
	}
}

// TestReflectorConfig tests the ReflectorConfig struct defaults.
func TestReflectorConfig(t *testing.T) {
	config := ReflectorConfig{
		Profile:         DefaultProfile,
		SignatureFilter: []string{"test"},
		PortFilter:      DefaultPortFilter,
		OUIFilter:       DefaultOUIFilter,
	}

	if config.Profile != "all" {
		t.Errorf("Expected default profile 'all', got '%s'", config.Profile)
	}
	if config.PortFilter != 3842 {
		t.Errorf("Expected default port filter 3842, got %d", config.PortFilter)
	}
	if config.OUIFilter != "00:c0:17" {
		t.Errorf("Expected default OUI filter '00:c0:17', got '%s'", config.OUIFilter)
	}
	if len(config.SignatureFilter) != 1 {
		t.Errorf("Expected 1 signature filter, got %d", len(config.SignatureFilter))
	}
}
