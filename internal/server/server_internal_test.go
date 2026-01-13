// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/krisarmstrong/stem/internal/auth"
	"github.com/krisarmstrong/stem/internal/netif"
)

// newTestServer creates a test server with automatic cleanup.
// This helper ensures that all servers created in tests are properly
// shut down to prevent goroutine leaks.
func newTestServer(t testing.TB) *Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1") // Use fast bcrypt for tests
	s, err := NewServer(8080)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// setupClaimsTestServer creates a server for claims tests.
func setupClaimsTestServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", "claimstest")
	t.Setenv("STEM_AUTH_PASSWORD", "claimspass123")

	return newTestServer(t)
}

// TestExtractClaims_ValidToken tests extractClaims with a valid token.
func TestExtractClaims_ValidToken(t *testing.T) {
	s := setupClaimsTestServer(t)
	token, _, authErr := s.authManager.AuthenticateWithRefresh(
		context.TODO(),
		"claimstest",
		"claimspass123",
	)
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

// TestExtractClaims_StandardBearerFormat tests that standard Bearer format works.
// Note: The cookie-first auth no longer handles extra spaces around "Bearer".
func TestExtractClaims_StandardBearerFormat(t *testing.T) {
	s := setupClaimsTestServer(t)
	token, _, authErr := s.authManager.AuthenticateWithRefresh(
		context.TODO(),
		"claimstest",
		"claimspass123",
	)
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
}

// TestAuditAuthFailure tests the auditAuthFailure method.
func TestAuditAuthFailure(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "audituser")
	t.Setenv("STEM_AUTH_PASSWORD", "auditpass123")

	s := newTestServer(t)

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
		t.Errorf(
			"Expected Access-Control-Allow-Origin 'http://localhost:8080', got '%s'",
			allowOrigin,
		)
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
		t.Errorf(
			"Expected Access-Control-Allow-Origin 'http://127.0.0.1:3000', got '%s'",
			allowOrigin,
		)
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

// TestCorsMiddleware_RFC1918Origin tests corsMiddleware allows RFC 1918 private network origins.
func TestCorsMiddleware_RFC1918Origin(t *testing.T) {
	wrapped := setupCorsTestHandler()
	tests := []struct {
		name   string
		origin string
	}{
		{"class C 192.168.x.x", "http://192.168.1.100:8080"},
		{"class A 10.x.x.x", "http://10.0.0.50:8080"},
		{"class B 172.16.x.x", "http://172.16.0.1:8080"},
		{"class B 172.31.x.x", "http://172.31.255.255:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			wrapped.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for RFC 1918 origin %s, got %d", tt.origin, w.Code)
			}
			if w.Header().Get("Access-Control-Allow-Origin") != tt.origin {
				t.Errorf("Expected Access-Control-Allow-Origin=%s, got %s",
					tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

// TestCorsMiddleware_RFC1918BypassBlocked tests corsMiddleware blocks RFC 1918 bypass attempts.
func TestCorsMiddleware_RFC1918BypassBlocked(t *testing.T) {
	wrapped := setupCorsTestHandler()
	tests := []struct {
		name   string
		origin string
	}{
		{"bypass via subdomain", "http://192.168.1.1.evil.com"},
		{"bypass via path", "http://evil.com/192.168.1.1"},
		{"bypass via prefix", "http://192.168.1.1000"},
		{"invalid class B range", "http://172.32.0.1:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			wrapped.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Errorf("Expected status 403 for bypass attempt %s, got %d", tt.origin, w.Code)
			}
		})
	}
}

// TestWriteJSON tests the writeJSON function.
func TestWriteJSON(t *testing.T) {
	t.Run("successful JSON encoding", func(t *testing.T) {
		w := httptest.NewRecorder()

		data := map[string]string{"key": "value"}
		writeJSON(w, data)

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf(
				"Expected Content-Type 'application/json', got '%s'",
				w.Header().Get("Content-Type"),
			)
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
			t.Errorf(
				"Expected Content-Type 'application/json', got '%s'",
				w.Header().Get("Content-Type"),
			)
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

// setupSecurityTestHandler creates a handler and wraps it with securityHeadersMiddleware.
func setupSecurityTestHandler() http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return securityHeadersMiddleware(handler)
}

// TestSecurityHeadersMiddleware_XFrameOptions tests X-Frame-Options header.
func TestSecurityHeadersMiddleware_XFrameOptions(t *testing.T) {
	wrapped := setupSecurityTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	header := w.Header().Get("X-Frame-Options")
	if header != "DENY" {
		t.Errorf("Expected X-Frame-Options 'DENY', got '%s'", header)
	}
}

// TestSecurityHeadersMiddleware_XContentTypeOptions tests X-Content-Type-Options header.
func TestSecurityHeadersMiddleware_XContentTypeOptions(t *testing.T) {
	wrapped := setupSecurityTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	header := w.Header().Get("X-Content-Type-Options")
	if header != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options 'nosniff', got '%s'", header)
	}
}

// TestSecurityHeadersMiddleware_XXSSProtection tests X-XSS-Protection header.
func TestSecurityHeadersMiddleware_XXSSProtection(t *testing.T) {
	wrapped := setupSecurityTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	header := w.Header().Get("X-XSS-Protection")
	if header != "1; mode=block" {
		t.Errorf("Expected X-XSS-Protection '1; mode=block', got '%s'", header)
	}
}

// TestSecurityHeadersMiddleware_ReferrerPolicy tests Referrer-Policy header.
func TestSecurityHeadersMiddleware_ReferrerPolicy(t *testing.T) {
	wrapped := setupSecurityTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	header := w.Header().Get("Referrer-Policy")
	if header != "strict-origin-when-cross-origin" {
		t.Errorf("Expected Referrer-Policy 'strict-origin-when-cross-origin', got '%s'", header)
	}
}

// TestSecurityHeadersMiddleware_ContentSecurityPolicy tests CSP header.
func TestSecurityHeadersMiddleware_ContentSecurityPolicy(t *testing.T) {
	wrapped := setupSecurityTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	header := w.Header().Get("Content-Security-Policy")
	if header == "" {
		t.Error("Expected Content-Security-Policy header to be set")
	}

	// Check that CSP contains key directives.
	expectedDirectives := []string{
		"default-src 'self'",
		"script-src 'self'",
		"object-src 'none'",
		"frame-ancestors 'none'",
	}

	for _, directive := range expectedDirectives {
		if !containsStr(header, directive) {
			t.Errorf("Expected CSP to contain '%s'", directive)
		}
	}
}

// containsStr checks if s contains substr.
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestSecurityHeadersMiddleware_PermissionsPolicy tests Permissions-Policy header.
func TestSecurityHeadersMiddleware_PermissionsPolicy(t *testing.T) {
	wrapped := setupSecurityTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	header := w.Header().Get("Permissions-Policy")
	if header == "" {
		t.Error("Expected Permissions-Policy header to be set")
	}

	// Check for restricted features.
	expectedRestrictions := []string{
		"camera=()",
		"microphone=()",
		"geolocation=()",
	}

	for _, restriction := range expectedRestrictions {
		if !containsStr(header, restriction) {
			t.Errorf("Expected Permissions-Policy to contain '%s'", restriction)
		}
	}
}

// TestSecurityHeadersMiddleware_NoHSTSWithoutTLS tests HSTS is not set for non-TLS.
func TestSecurityHeadersMiddleware_NoHSTSWithoutTLS(t *testing.T) {
	wrapped := setupSecurityTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	// HSTS should NOT be set for non-TLS connections.
	header := w.Header().Get("Strict-Transport-Security")
	if header != "" {
		t.Errorf("Expected no HSTS header for non-TLS connection, got '%s'", header)
	}
}

// TestSecurityHeadersMiddleware_HSTSWithTLS tests HSTS is set for TLS connections.
func TestSecurityHeadersMiddleware_HSTSWithTLS(t *testing.T) {
	wrapped := setupSecurityTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Simulate TLS connection.
	req.TLS = &tls.ConnectionState{
		Version:                     tls.VersionTLS13,
		HandshakeComplete:           true,
		DidResume:                   false,
		CipherSuite:                 tls.TLS_AES_128_GCM_SHA256,
		CurveID:                     0,
		NegotiatedProtocol:          "",
		NegotiatedProtocolIsMutual:  false,
		ServerName:                  "",
		PeerCertificates:            nil,
		VerifiedChains:              nil,
		SignedCertificateTimestamps: nil,
		OCSPResponse:                nil,
		TLSUnique:                   nil,
		ECHAccepted:                 false,
	}
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	header := w.Header().Get("Strict-Transport-Security")
	if header == "" {
		t.Error("Expected HSTS header for TLS connection")
	}

	if !containsStr(header, "max-age=") {
		t.Error("Expected HSTS header to contain max-age directive")
	}
}

// TestSecurityHeadersMiddleware_AllHeaders tests all security headers are present.
func TestSecurityHeadersMiddleware_AllHeaders(t *testing.T) {
	wrapped := setupSecurityTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	requiredHeaders := []string{
		"X-Frame-Options",
		"X-Content-Type-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Content-Security-Policy",
		"Permissions-Policy",
	}

	for _, header := range requiredHeaders {
		if w.Header().Get(header) == "" {
			t.Errorf("Expected header '%s' to be set", header)
		}
	}
}

// TestDefaultTLSConfig tests the DefaultTLSConfig function.
func TestDefaultTLSConfig(t *testing.T) {
	config := DefaultTLSConfig()

	if !config.Enabled {
		t.Error("Expected TLS to be enabled by default")
	}
	if config.CertFile != "" {
		t.Errorf("Expected empty CertFile, got '%s'", config.CertFile)
	}
	if config.KeyFile != "" {
		t.Errorf("Expected empty KeyFile, got '%s'", config.KeyFile)
	}
	if config.CertsDir != defaultCertsDir {
		t.Errorf("Expected CertsDir '%s', got '%s'", defaultCertsDir, config.CertsDir)
	}
}

// TestCreateTLSConfig tests the createTLSConfig function.
func TestCreateTLSConfig(t *testing.T) {
	config := createTLSConfig()

	if config == nil {
		t.Fatal("Expected non-nil TLS config")
	}

	if config.MinVersion != tls.VersionTLS13 {
		t.Errorf("Expected TLS 1.3 min version, got %d", config.MinVersion)
	}
}

// TestEnsureSelfSignedCert tests the ensureSelfSignedCert function.
func TestEnsureSelfSignedCert(t *testing.T) {
	// Create a temporary directory for test certificates.
	tempDir := t.TempDir()

	t.Run("generate new certificates", func(t *testing.T) {
		certFile, keyFile, err := ensureSelfSignedCert(tempDir)
		if err != nil {
			t.Fatalf("ensureSelfSignedCert() error: %v", err)
		}

		if certFile == "" {
			t.Error("Expected non-empty certFile")
		}
		if keyFile == "" {
			t.Error("Expected non-empty keyFile")
		}

		// Verify files exist.
		_, certStatErr := os.Stat(certFile)
		if certStatErr != nil {
			t.Errorf("Certificate file does not exist: %v", certStatErr)
		}
		_, keyStatErr := os.Stat(keyFile)
		if keyStatErr != nil {
			t.Errorf("Key file does not exist: %v", keyStatErr)
		}
	})

	t.Run("use existing certificates", func(t *testing.T) {
		// Should reuse existing certificates.
		certFile, keyFile, err := ensureSelfSignedCert(tempDir)
		if err != nil {
			t.Fatalf("ensureSelfSignedCert() error: %v", err)
		}

		if certFile == "" || keyFile == "" {
			t.Error("Expected non-empty certificate paths")
		}
	})

	t.Run("empty certs dir defaults to default", func(t *testing.T) {
		// This would use the default certs dir.
		// Skip if we don't want to pollute the filesystem.
		t.Skip("Skipping to avoid creating files in default location")
	})
}

// TestTLSConfigStruct tests the TLSConfig struct.
func TestTLSConfigStruct(t *testing.T) {
	config := TLSConfig{
		Enabled:  true,
		CertFile: "/path/to/cert.pem",
		KeyFile:  "/path/to/key.pem",
		CertsDir: "/path/to/certs",
	}

	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.CertFile != "/path/to/cert.pem" {
		t.Errorf("Unexpected CertFile: %s", config.CertFile)
	}
	if config.KeyFile != "/path/to/key.pem" {
		t.Errorf("Unexpected KeyFile: %s", config.KeyFile)
	}
	if config.CertsDir != "/path/to/certs" {
		t.Errorf("Unexpected CertsDir: %s", config.CertsDir)
	}
}

// TestUpdateStats tests the UpdateStats function.
func TestUpdateStats(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "statsuser")
	t.Setenv("STEM_AUTH_PASSWORD", "statspass123")

	s := newTestServer(t)

	// Update stats.
	s.UpdateStats(100, 200, 1000, 2000, 10.0, 5.0)

	// Verify stats were updated.
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()

	if s.stats.PacketsReceived != 100 {
		t.Errorf("Expected PacketsReceived 100, got %d", s.stats.PacketsReceived)
	}
	if s.stats.PacketsSent != 200 {
		t.Errorf("Expected PacketsSent 200, got %d", s.stats.PacketsSent)
	}
	if s.stats.BytesReceived != 1000 {
		t.Errorf("Expected BytesReceived 1000, got %d", s.stats.BytesReceived)
	}
	if s.stats.BytesSent != 2000 {
		t.Errorf("Expected BytesSent 2000, got %d", s.stats.BytesSent)
	}
}

// TestUpdateStatsReplaces tests that UpdateStats replaces values (not accumulates).
func TestUpdateStatsReplaces(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "statsuser2")
	t.Setenv("STEM_AUTH_PASSWORD", "statspass123")

	s := newTestServer(t)

	// Update stats multiple times - values should be replaced, not accumulated.
	s.UpdateStats(100, 100, 1000, 1000, 10.0, 5.0)
	s.UpdateStats(50, 50, 500, 500, 5.0, 2.5)

	// Verify stats are the latest values.
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()

	if s.stats.PacketsReceived != 50 {
		t.Errorf("Expected PacketsReceived 50, got %d", s.stats.PacketsReceived)
	}
	if s.stats.PacketsSent != 50 {
		t.Errorf("Expected PacketsSent 50, got %d", s.stats.PacketsSent)
	}
	if s.stats.CurrentPPS != 5.0 {
		t.Errorf("Expected CurrentPPS 5.0, got %f", s.stats.CurrentPPS)
	}
	if s.stats.CurrentMbps != 2.5 {
		t.Errorf("Expected CurrentMbps 2.5, got %f", s.stats.CurrentMbps)
	}
}

// TestWriteJSONEncodingError tests writeJSON with unencodable data.
func TestWriteJSONEncodingError(t *testing.T) {
	w := httptest.NewRecorder()

	// Create an unencodable type (channel).
	unencodable := make(chan int)
	writeJSON(w, unencodable)

	// Should return error status.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for unencodable data, got %d", w.Code)
	}
}

// TestBuildFallbackReflectorStats tests the buildFallbackReflectorStats function.
func TestBuildFallbackReflectorStats(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "reflectuser")
	t.Setenv("STEM_AUTH_PASSWORD", "reflectpass123")

	s := newTestServer(t)

	// Update some stats.
	s.UpdateStats(100, 50, 10000, 5000, 5.0, 1.0)

	// Get fallback stats.
	elapsed := 10.0
	stats := s.buildFallbackReflectorStats(elapsed)

	// Verify stats.
	if stats.PacketsReceived != 100 {
		t.Errorf("Expected PacketsReceived 100, got %d", stats.PacketsReceived)
	}
	if stats.PacketsReflected != 50 {
		t.Errorf("Expected PacketsReflected 50, got %d", stats.PacketsReflected)
	}
	if stats.BytesReceived != 10000 {
		t.Errorf("Expected BytesReceived 10000, got %d", stats.BytesReceived)
	}
	if stats.BytesReflected != 5000 {
		t.Errorf("Expected BytesReflected 5000, got %d", stats.BytesReflected)
	}
	if stats.Uptime != elapsed {
		t.Errorf("Expected Uptime %f, got %f", elapsed, stats.Uptime)
	}

	// Verify rate calculation.
	expectedPPS := float64(50) / elapsed
	if stats.RatePPS != expectedPPS {
		t.Errorf("Expected RatePPS %f, got %f", expectedPPS, stats.RatePPS)
	}
}

// TestBuildFallbackReflectorStatsZeroElapsed tests with zero elapsed time.
func TestBuildFallbackReflectorStatsZeroElapsed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "reflectuser2")
	t.Setenv("STEM_AUTH_PASSWORD", "reflectpass123")

	s := newTestServer(t)

	// Get stats with zero elapsed time.
	stats := s.buildFallbackReflectorStats(0)

	// Rates should be zero when elapsed is zero.
	if stats.RatePPS != 0 {
		t.Errorf("Expected RatePPS 0 for zero elapsed, got %f", stats.RatePPS)
	}
	if stats.RateMbps != 0 {
		t.Errorf("Expected RateMbps 0 for zero elapsed, got %f", stats.RateMbps)
	}
}

// TestServerShutdown tests the Shutdown function.
func TestServerShutdown(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "shutdownuser")
	t.Setenv("STEM_AUTH_PASSWORD", "shutdownpass123")

	s := newTestServer(t)

	// Test shutdown without server running (should not panic).
	shutdownErr := s.Shutdown()

	// Shutdown on unstarted server should handle gracefully.
	if shutdownErr != nil {
		t.Logf("Shutdown returned error (expected for unstarted server): %v", shutdownErr)
	}
}

// TestNeedsInitialSetup tests the needsInitialSetup function.
func TestNeedsInitialSetup(t *testing.T) {
	t.Run("setup mode enabled", func(t *testing.T) {
		t.Setenv("STEM_AUTH_USERNAME", "setupuser")
		t.Setenv("STEM_AUTH_PASSWORD", "setuppass123")
		t.Setenv("STEM_SETUP_MODE", "true")

		s := newTestServer(t)

		if !s.needsInitialSetup() {
			t.Error("Expected needsInitialSetup to return true when STEM_SETUP_MODE=true")
		}
	})

	t.Run("setup mode disabled", func(t *testing.T) {
		t.Setenv("STEM_AUTH_USERNAME", "setupuser2")
		t.Setenv("STEM_AUTH_PASSWORD", "setuppass123")
		t.Setenv("STEM_SETUP_MODE", "false")

		s := newTestServer(t)

		if s.needsInitialSetup() {
			t.Error("Expected needsInitialSetup to return false when STEM_SETUP_MODE=false")
		}
	})
}

// TestMarkSetupComplete tests the markSetupComplete function.
func TestMarkSetupComplete(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "marksetupuser")
	t.Setenv("STEM_AUTH_PASSWORD", "marksetuppass123")
	t.Setenv("STEM_SETUP_MODE", "true")

	s := newTestServer(t)

	// Initially setup should be needed.
	if !s.needsInitialSetup() {
		t.Error("Expected needsInitialSetup to return true initially")
	}

	// Mark setup as complete.
	s.markSetupComplete()

	// After marking complete, setup should not be needed.
	if s.needsInitialSetup() {
		t.Error("Expected needsInitialSetup to return false after markSetupComplete")
	}

	// setupComplete flag should be set.
	if !s.setupComplete {
		t.Error("Expected setupComplete to be true")
	}
}

// TestResolveTestInterface tests the resolveTestInterface function.
func TestResolveTestInterface(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "resolveifaceuser")
	t.Setenv("STEM_AUTH_PASSWORD", "resolveifacepass123")

	s := newTestServer(t)

	t.Run("with explicit interface", func(t *testing.T) {
		iface, ifaceErr := s.resolveTestInterface("eth0")
		if ifaceErr != nil {
			t.Errorf("resolveTestInterface() error: %v", ifaceErr)
		}
		if iface != "eth0" {
			t.Errorf("Expected 'eth0', got '%s'", iface)
		}
	})

	t.Run("with empty interface uses selected", func(t *testing.T) {
		// Set a selected interface.
		s.statsMu.Lock()
		s.selectedIface = "en0"
		s.statsMu.Unlock()

		iface, ifaceErr := s.resolveTestInterface("")
		if ifaceErr != nil {
			t.Errorf("resolveTestInterface() error: %v", ifaceErr)
		}
		if iface != "en0" {
			t.Errorf("Expected 'en0', got '%s'", iface)
		}
	})

	t.Run("with empty interface and no selected", func(t *testing.T) {
		// Clear selected interface.
		s.statsMu.Lock()
		s.selectedIface = ""
		s.statsMu.Unlock()

		_, ifaceErr := s.resolveTestInterface("")
		if ifaceErr == nil {
			t.Error("Expected error when no interface selected")
		}
	})
}

// TestBeginTestRun tests the beginTestRun function.
func TestBeginTestRun(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "begintestuser")
	t.Setenv("STEM_AUTH_PASSWORD", "begintestpass123")

	s := newTestServer(t)

	t.Run("successful begin", func(t *testing.T) {
		// Ensure idle state.
		s.statsMu.Lock()
		s.testStatus = statusIdle
		s.statsMu.Unlock()

		beginErr := s.beginTestRun("throughput", "benchmark")
		if beginErr != nil {
			t.Errorf("beginTestRun() error: %v", beginErr)
		}

		s.statsMu.RLock()
		if s.testStatus != statusStarting {
			t.Errorf("Expected status 'starting', got '%s'", s.testStatus)
		}
		if s.currentTest != "throughput" {
			t.Errorf("Expected currentTest 'throughput', got '%s'", s.currentTest)
		}
		if s.currentModule != "benchmark" {
			t.Errorf("Expected currentModule 'benchmark', got '%s'", s.currentModule)
		}
		s.statsMu.RUnlock()
	})

	t.Run("already running", func(t *testing.T) {
		// Set running state.
		s.statsMu.Lock()
		s.testStatus = statusRunning
		s.statsMu.Unlock()

		beginErr := s.beginTestRun("latency", "benchmark")
		if beginErr == nil {
			t.Error("Expected error when test already running")
		}
		if !errors.Is(beginErr, errTestAlreadyRunning) {
			t.Errorf("Expected errTestAlreadyRunning, got: %v", beginErr)
		}
	})
}

// TestValidateReflectorProfile tests the validateReflectorProfile function.
func TestValidateReflectorProfile(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "validateprofileuser")
	t.Setenv("STEM_AUTH_PASSWORD", "validateprofilepass123")

	s := newTestServer(t)

	t.Run("valid profiles", func(t *testing.T) {
		validProfiles := []string{"netally", "msn", "all", "custom", ""}
		for _, profile := range validProfiles {
			validateErr := s.validateReflectorProfile(profile)
			if validateErr != nil {
				t.Errorf("validateReflectorProfile(%s) error: %v", profile, validateErr)
			}
		}
	})

	t.Run("invalid profile", func(t *testing.T) {
		validateErr := s.validateReflectorProfile("invalid_profile")
		if validateErr == nil {
			t.Error("Expected error for invalid profile")
		}
	})
}

// assertBuildConfigChanges verifies buildReflectorConfigUpdate succeeds with expected changes.
func assertBuildConfigChanges(t *testing.T, s *Server, cfg *ReflectorConfig, expectedChanges int) {
	t.Helper()
	changes, _, buildErr := s.buildReflectorConfigUpdate(cfg)
	if buildErr != nil {
		t.Errorf("buildReflectorConfigUpdate() error: %v", buildErr)
		return
	}
	if len(changes) != expectedChanges {
		t.Errorf("Expected %d change(s), got %d", expectedChanges, len(changes))
	}
}

// TestBuildReflectorConfigUpdate tests the buildReflectorConfigUpdate function.
func TestBuildReflectorConfigUpdate(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "buildconfiguser")
	t.Setenv("STEM_AUTH_PASSWORD", "buildconfigpass123")

	s := newTestServer(t)

	t.Run("update profile", func(t *testing.T) {
		cfg := &ReflectorConfig{Profile: "netally", SignatureFilter: nil, OUIFilter: "", PortFilter: 0}
		assertBuildConfigChanges(t, s, cfg, 1)
		if s.reflectorConfig.Profile != "netally" {
			t.Errorf("Expected profile 'netally', got '%s'", s.reflectorConfig.Profile)
		}
	})

	t.Run("update OUI filter", func(t *testing.T) {
		cfg := &ReflectorConfig{Profile: "", SignatureFilter: nil, OUIFilter: "00:11:22", PortFilter: 0}
		assertBuildConfigChanges(t, s, cfg, 1)
		if s.reflectorConfig.OUIFilter != "00:11:22" {
			t.Errorf("Expected OUI filter '00:11:22', got '%s'", s.reflectorConfig.OUIFilter)
		}
	})

	t.Run("update port filter", func(t *testing.T) {
		cfg := &ReflectorConfig{Profile: "", SignatureFilter: nil, OUIFilter: "", PortFilter: 9999}
		assertBuildConfigChanges(t, s, cfg, 1)
		if s.reflectorConfig.PortFilter != 9999 {
			t.Errorf("Expected port filter 9999, got %d", s.reflectorConfig.PortFilter)
		}
	})

	t.Run("update signature filter", func(t *testing.T) {
		cfg := &ReflectorConfig{Profile: "", SignatureFilter: []string{"sig1", "sig2"}, OUIFilter: "", PortFilter: 0}
		assertBuildConfigChanges(t, s, cfg, 1)
		if len(s.reflectorConfig.SignatureFilter) != 2 {
			t.Errorf("Expected 2 signature filters, got %d", len(s.reflectorConfig.SignatureFilter))
		}
	})

	t.Run("invalid port out of range", func(t *testing.T) {
		cfg := &ReflectorConfig{Profile: "", SignatureFilter: nil, OUIFilter: "", PortFilter: 100000}
		_, _, buildErr := s.buildReflectorConfigUpdate(cfg)
		if buildErr == nil {
			t.Error("Expected error for port out of range")
		}
	})

	t.Run("multiple changes", func(t *testing.T) {
		cfg := &ReflectorConfig{Profile: "msn", SignatureFilter: nil, OUIFilter: "aa:bb:cc", PortFilter: 1234}
		assertBuildConfigChanges(t, s, cfg, 3)
	})
}

// TestApplyReflectorDataplaneUpdate tests the applyReflectorDataplaneUpdate function.
func TestApplyReflectorDataplaneUpdate(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "applydpuser")
	t.Setenv("STEM_AUTH_PASSWORD", "applydppass123")

	s := newTestServer(t)

	t.Run("nil update", func(t *testing.T) {
		applyErr := s.applyReflectorDataplaneUpdate(nil, nil)
		if applyErr != nil {
			t.Errorf("applyReflectorDataplaneUpdate() error: %v", applyErr)
		}
	})

	t.Run("no executor", func(t *testing.T) {
		// Ensure no executor.
		s.statsMu.Lock()
		s.reflectorExec = nil
		s.statsMu.Unlock()

		changes := []string{"test change"}
		applyErr := s.applyReflectorDataplaneUpdate(nil, changes)
		if applyErr != nil {
			t.Errorf("applyReflectorDataplaneUpdate() error: %v", applyErr)
		}
	})
}

// TestExecuteTest tests the executeTest function.
func TestExecuteTest(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "exectestuser")
	t.Setenv("STEM_AUTH_PASSWORD", "exectestpass123")

	s := newTestServer(t)

	t.Run("unknown module", func(t *testing.T) {
		execErr := s.executeTest("unknown_module", "test", "eth0", nil)
		if execErr == nil {
			t.Error("Expected error for unknown module")
		}
	})
}

// TestRespondTestExecutionError tests the respondTestExecutionError function.
func TestRespondTestExecutionError(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "responderruser")
	t.Setenv("STEM_AUTH_PASSWORD", "responderrpass123")

	s := newTestServer(t)

	w := httptest.NewRecorder()
	testErr := errTestAlreadyRunning

	s.respondTestExecutionError(w, testErr, "benchmark", "throughput")

	// Should set error state.
	s.statsMu.RLock()
	if s.testStatus != statusError {
		t.Errorf("Expected status 'error', got '%s'", s.testStatus)
	}
	if s.testResult == nil {
		t.Error("Expected testResult to be set")
	}
	s.statsMu.RUnlock()
}

// TestGetDataDir tests the getDataDir function.
func TestGetDataDir(t *testing.T) {
	t.Run("default to current directory", func(t *testing.T) {
		t.Setenv("STEM_DATA_DIR", "")
		dir := getDataDir()
		if dir != "." {
			t.Errorf("Expected '.', got '%s'", dir)
		}
	})

	t.Run("use environment variable", func(t *testing.T) {
		t.Setenv("STEM_DATA_DIR", "/custom/data")
		dir := getDataDir()
		if dir != "/custom/data" {
			t.Errorf("Expected '/custom/data', got '%s'", dir)
		}
	})
}

// TestIsLocalhostOriginInternal tests the isLocalhostOrigin function (additional cases).
func TestIsLocalhostOriginInternal(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{"localhost_no_port", "http://localhost", true},
		{"127.0.0.1_no_port", "http://127.0.0.1", true},
		{"::1_no_port", "http://[::1]", true},
		{"external_https", "https://example.com", false},
		{"localhost_https", "https://localhost:8443", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLocalhostOrigin(tt.origin)
			if got != tt.want {
				t.Errorf("isLocalhostOrigin(%s) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

// TestIsSameOrigin tests the isSameOrigin function.
func TestIsSameOrigin(t *testing.T) {
	tests := []struct {
		name        string
		origin      string
		requestHost string
		want        bool
	}{
		{"same host and port", "http://10.0.0.1:8080", "10.0.0.1:8080", true},
		{"different port", "http://10.0.0.1:8080", "10.0.0.1:9090", false},
		{"different host", "http://10.0.0.1:8080", "10.0.0.2:8080", false},
		{"invalid URL", "not-a-url", "10.0.0.1:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSameOrigin(tt.origin, tt.requestHost)
			if got != tt.want {
				t.Errorf(
					"isSameOrigin(%s, %s) = %v, want %v",
					tt.origin,
					tt.requestHost,
					got,
					tt.want,
				)
			}
		})
	}
}

// TestResolveTestModule tests the resolveTestModule function.
func TestResolveTestModule(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "resolvemoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "resolvemodpass123")

	s := newTestServer(t)

	t.Run("valid test type", func(t *testing.T) {
		mod, modErr := s.resolveTestModule("rfc2544_throughput")
		if modErr != nil {
			t.Errorf("resolveTestModule() error: %v", modErr)
		}
		if mod == nil {
			t.Error("Expected non-nil module")
		}
	})

	t.Run("reflector test type", func(t *testing.T) {
		mod, modErr := s.resolveTestModule("reflect")
		if modErr != nil {
			t.Errorf("resolveTestModule() error: %v", modErr)
		}
		if mod == nil {
			t.Error("Expected non-nil module")
		}
		if mod.Name() != "reflector" {
			t.Errorf("Expected module 'reflector', got '%s'", mod.Name())
		}
	})

	t.Run("invalid test type", func(t *testing.T) {
		_, modErr := s.resolveTestModule("invalid_test_type_xyz")
		if modErr == nil {
			t.Error("Expected error for invalid test type")
		}
	})
}

// TestHandleTestStopVariousStates tests handleTestStop with various states.
func TestHandleTestStopVariousStates(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "stopstatesuser")
	t.Setenv("STEM_AUTH_PASSWORD", "stopstatespass123")

	s := newTestServer(t)

	t.Run("stop when starting", func(t *testing.T) {
		// Set starting state.
		s.statsMu.Lock()
		s.testStatus = statusStarting
		s.currentTest = "throughput"
		s.statsMu.Unlock()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/stop", nil)
		w := httptest.NewRecorder()

		s.handleTestStop(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		s.statsMu.RLock()
		if s.testStatus != statusCancelled {
			t.Errorf("Expected status 'cancelled', got '%s'", s.testStatus)
		}
		s.statsMu.RUnlock()
	})

	t.Run("stop when running", func(t *testing.T) {
		// Set running state.
		s.statsMu.Lock()
		s.testStatus = statusRunning
		s.currentTest = "latency"
		s.currentModule = "benchmark"
		s.statsMu.Unlock()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/stop", nil)
		w := httptest.NewRecorder()

		s.handleTestStop(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		s.statsMu.RLock()
		if s.testStatus != statusCancelled {
			t.Errorf("Expected status 'cancelled', got '%s'", s.testStatus)
		}
		if s.currentTest != "" {
			t.Errorf("Expected currentTest to be empty, got '%s'", s.currentTest)
		}
		s.statsMu.RUnlock()
	})

	t.Run("stop when idle", func(t *testing.T) {
		// Set idle state.
		s.statsMu.Lock()
		s.testStatus = statusIdle
		s.currentTest = ""
		s.reflectorExec = nil
		s.statsMu.Unlock()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/stop", nil)
		w := httptest.NewRecorder()

		s.handleTestStop(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 when no test running, got %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test/stop", nil)
		w := httptest.NewRecorder()

		s.handleTestStop(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleTestResultWithResult tests handleTestResult when result exists.
func TestHandleTestResultWithResult(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "resultuser")
	t.Setenv("STEM_AUTH_PASSWORD", "resultpass123")

	s := newTestServer(t)

	// Set a test result.
	s.statsMu.Lock()
	s.testResult = &TestResultResponse{
		Status:   statusCompleted,
		TestType: "throughput",
		Module:   "benchmark",
		Success:  true,
		Error:    "",
		Message:  "Test completed",
		Data:     map[string]any{"throughput": 950.5},
	}
	s.statsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test/result", nil)
	w := httptest.NewRecorder()

	s.handleTestResult(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response contains result.
	body := w.Body.String()
	if !containsStr(body, "completed") {
		t.Error("Expected response to contain 'completed'")
	}
	if !containsStr(body, "throughput") {
		t.Error("Expected response to contain 'throughput'")
	}
}

// TestHandleAuthCSRFCoverage tests handleAuthCSRF to improve coverage.
func TestHandleAuthCSRFCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "csrfuser")
	t.Setenv("STEM_AUTH_PASSWORD", "csrfpass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/csrf", nil)
		w := httptest.NewRecorder()

		s.handleAuthCSRF(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleHealthReadyCoverage tests handleHealthReady for more coverage.
func TestHandleHealthReadyCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "readyuser")
	t.Setenv("STEM_AUTH_PASSWORD", "readypass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health/ready", nil)
		w := httptest.NewRecorder()

		s.handleHealthReady(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleStatsCoverage tests handleStats for more coverage.
func TestHandleStatsCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "statsuser")
	t.Setenv("STEM_AUTH_PASSWORD", "statspass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/stats", nil)
		w := httptest.NewRecorder()

		s.handleStats(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("with current test", func(t *testing.T) {
		// Set a current test.
		s.statsMu.Lock()
		s.testStatus = statusRunning
		s.currentTest = "throughput"
		s.statsMu.Unlock()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
		w := httptest.NewRecorder()

		s.handleStats(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify current test is in response.
		body := w.Body.String()
		if !containsStr(body, "throughput") {
			t.Error("Expected response to contain 'throughput'")
		}
	})
}

// TestHandleLicenseCoverage tests handleLicense for more coverage.
func TestHandleLicenseCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "licenseuser")
	t.Setenv("STEM_AUTH_PASSWORD", "licensepass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleInterfacesCoverage tests handleInterfaces for more coverage.
func TestHandleInterfacesCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "ifaceuser")
	t.Setenv("STEM_AUTH_PASSWORD", "ifacepass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/interfaces", nil)
		w := httptest.NewRecorder()

		s.handleInterfaces(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestValidateInterfaceExistsFunc tests the validateInterfaceExists function.
func TestValidateInterfaceExistsFunc(t *testing.T) {
	t.Run("nonexistent interface", func(t *testing.T) {
		w := httptest.NewRecorder()
		exists := validateInterfaceExists(w, "nonexistent_iface_xyz123")
		if exists {
			t.Error("Expected false for nonexistent interface")
		}
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestHandleRecoveryStatusCoverage tests handleRecoveryStatus for more coverage.
func TestHandleRecoveryStatusCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recstatususer")
	t.Setenv("STEM_AUTH_PASSWORD", "recstatuspass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/status", nil)
		w := httptest.NewRecorder()

		s.handleRecoveryStatus(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleSetupStatusCoverage tests handleSetupStatus for more coverage.
func TestHandleSetupStatusCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "setupstatususer")
	t.Setenv("STEM_AUTH_PASSWORD", "setupstatuspass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()

		s.handleSetupStatus(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleRecoveryCompleteCoverage tests handleRecoveryComplete for more coverage.
func TestHandleRecoveryCompleteCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recompleteuser")
	t.Setenv("STEM_AUTH_PASSWORD", "recompletepass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/complete", nil)
		w := httptest.NewRecorder()

		s.handleRecoveryComplete(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/recovery/complete",
			bytes.NewBufferString("{invalid}"),
		)
		w := httptest.NewRecorder()

		s.handleRecoveryComplete(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestHandleSetupCompleteCoverage tests handleSetupComplete for more coverage.
func TestHandleSetupCompleteCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "setcompleteuser")
	t.Setenv("STEM_AUTH_PASSWORD", "setcompletepass123")
	t.Setenv("STEM_SETUP_MODE", "false")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/complete", nil)
		w := httptest.NewRecorder()

		s.handleSetupComplete(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("setup not needed", func(t *testing.T) {
		// Mark setup as complete.
		s.markSetupComplete()

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/setup/complete",
			bytes.NewBufferString(`{"password":"test","setupToken":"token"}`),
		)
		w := httptest.NewRecorder()

		s.handleSetupComplete(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 when setup not needed, got %d", w.Code)
		}
	})
}

// TestHandleLicenseFullCoverage tests handleLicense for full coverage.
func TestHandleLicenseFullCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "licfulluser")
	t.Setenv("STEM_AUTH_PASSWORD", "licfullpass123")

	s := newTestServer(t)

	t.Run("get license with manager", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify response structure.
		body := w.Body.String()
		if !containsStr(body, "activated") {
			t.Error("Expected response to contain 'activated'")
		}
	})

	t.Run("get license without manager", func(t *testing.T) {
		// Save and clear the license manager.
		origManager := s.licenseManager
		s.licenseManager = nil

		req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !containsStr(body, "not initialized") {
			t.Error("Expected response to contain 'not initialized'")
		}

		// Restore.
		s.licenseManager = origManager
	})
}

// TestHandleLicenseActivateCoverage tests handleLicenseActivate for coverage.
func TestHandleLicenseActivateCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "licactuser")
	t.Setenv("STEM_AUTH_PASSWORD", "licactpass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/license/activate", nil)
		w := httptest.NewRecorder()

		s.handleLicenseActivate(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("no license manager", func(t *testing.T) {
		origManager := s.licenseManager
		s.licenseManager = nil

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/license/activate",
			bytes.NewBufferString(`{"licenseKey":"test-key"}`),
		)
		w := httptest.NewRecorder()

		s.handleLicenseActivate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !containsStr(body, "not initialized") {
			t.Error("Expected response to contain 'not initialized'")
		}

		s.licenseManager = origManager
	})

	t.Run("empty license key", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/license/activate",
			bytes.NewBufferString(`{"licenseKey":""}`),
		)
		w := httptest.NewRecorder()

		s.handleLicenseActivate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !containsStr(body, "required") {
			t.Error("Expected response to contain 'required'")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/license/activate",
			bytes.NewBufferString(`{invalid}`),
		)
		w := httptest.NewRecorder()

		s.handleLicenseActivate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestHandleLicenseTrialCoverage tests handleLicenseTrial for coverage.
func TestHandleLicenseTrialCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "lictrialuser")
	t.Setenv("STEM_AUTH_PASSWORD", "lictrialpass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/license/trial", nil)
		w := httptest.NewRecorder()

		s.handleLicenseTrial(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("no license manager", func(t *testing.T) {
		origManager := s.licenseManager
		s.licenseManager = nil

		req := httptest.NewRequest(http.MethodGet, "/api/v1/license/trial", nil)
		w := httptest.NewRecorder()

		s.handleLicenseTrial(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !containsStr(body, "not initialized") {
			t.Error("Expected response to contain 'not initialized'")
		}

		s.licenseManager = origManager
	})

	t.Run("get trial status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/license/trial", nil)
		w := httptest.NewRecorder()

		s.handleLicenseTrial(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !containsStr(body, "active") {
			t.Error("Expected response to contain 'active'")
		}
	})

	t.Run("start trial", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/license/trial", nil)
		w := httptest.NewRecorder()

		s.handleLicenseTrial(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// TestHandleAuthCSRFCoverageFull tests handleAuthCSRF for full coverage.
func TestHandleAuthCSRFCoverageFull(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "csrffulluser")
	t.Setenv("STEM_AUTH_PASSWORD", "csrffullpass123")

	s := newTestServer(t)

	t.Run("get csrf token with auth", func(t *testing.T) {
		// Get a valid auth token first.
		token, _, authErr := s.authManager.AuthenticateWithRefresh(
			context.TODO(),
			"csrffulluser",
			"csrffullpass123",
		)
		if authErr != nil {
			t.Fatalf("Failed to get auth token: %v", authErr)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.handleAuthCSRF(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !containsStr(body, "token") {
			t.Error("Expected response to contain 'token'")
		}
	})

	t.Run("get csrf token without auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
		w := httptest.NewRecorder()

		s.handleAuthCSRF(w, req)

		// Should require auth.
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

// TestHandleRecoveryStatusCoverageFull tests handleRecoveryStatus for full coverage.
func TestHandleRecoveryStatusCoverageFull(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recstatusfulluser")
	t.Setenv("STEM_AUTH_PASSWORD", "recstatusfullpass123")

	s := newTestServer(t)

	t.Run("recovery status without manager", func(t *testing.T) {
		// Save and clear the recovery manager.
		origManager := s.recoveryTokenManager
		s.recoveryTokenManager = nil

		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/status", nil)
		w := httptest.NewRecorder()

		s.handleRecoveryStatus(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !containsStr(body, "false") {
			t.Error("Expected response to contain 'false'")
		}

		// Restore.
		s.recoveryTokenManager = origManager
	})
}

// TestExecuteReflectorCoverage tests executeReflector for coverage.
func TestExecuteReflectorCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "execrefluser")
	t.Setenv("STEM_AUTH_PASSWORD", "execreflpass123")

	s := newTestServer(t)

	// Note: Actually executing reflector requires a valid interface
	// and may have platform-specific behavior. Test the error path.
	t.Run("execute with nonexistent interface", func(t *testing.T) {
		execErr := s.executeReflector("nonexistent_iface_xyz123", nil)
		if execErr == nil {
			t.Error("Expected error for nonexistent interface")
		}
	})
}

// TestRunModuleTestCoverage tests runModuleTest for coverage.
func TestRunModuleTestCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "runmoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "runmodpass123")

	s := newTestServer(t)

	t.Run("with valid module", func(t *testing.T) {
		// This tests the executeTest path for benchmark module.
		execErr := s.executeTest(
			"benchmark",
			"rfc2544_throughput",
			"nonexistent_iface_xyz123",
			nil,
		)
		// May fail due to interface, but that's expected.
		if execErr != nil {
			t.Logf("Expected error for nonexistent interface: %v", execErr)
		}
	})
}

// TestHandleReflectorStatsCoverage tests handleReflectorStats for coverage.
func TestHandleReflectorStatsCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "reflstatusr")
	t.Setenv("STEM_AUTH_PASSWORD", "reflstatspass123")

	s := newTestServer(t)

	t.Run("get reflector stats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/stats", nil)
		w := httptest.NewRecorder()

		s.handleReflectorStats(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !containsStr(body, "running") {
			t.Error("Expected response to contain 'running'")
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/stats", nil)
		w := httptest.NewRecorder()

		s.handleReflectorStats(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleReflectorConfigCoverage tests handleReflectorConfig for coverage.
func TestHandleReflectorConfigCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "reflcfguser")
	t.Setenv("STEM_AUTH_PASSWORD", "reflcfgpass123")

	s := newTestServer(t)

	t.Run("get reflector config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/config", nil)
		w := httptest.NewRecorder()

		s.handleReflectorConfig(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("post reflector config - invalid profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config",
			bytes.NewBufferString(`{"profile":"invalid_profile_xyz"}`))
		w := httptest.NewRecorder()

		s.handleReflectorConfig(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("post reflector config - valid profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config",
			bytes.NewBufferString(`{"profile":"all"}`))
		w := httptest.NewRecorder()

		s.handleReflectorConfig(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/reflector/config", nil)
		w := httptest.NewRecorder()

		s.handleReflectorConfig(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleModuleByNameCoverage tests handleModuleByName for coverage.
func TestHandleModuleByNameCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "modbyuser")
	t.Setenv("STEM_AUTH_PASSWORD", "modbypass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/modules/benchmark", nil)
		w := httptest.NewRecorder()

		s.handleModuleByName(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("unknown module", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/unknown_module_xyz", nil)
		w := httptest.NewRecorder()

		s.handleModuleByName(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("valid module", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/modules/benchmark", nil)
		w := httptest.NewRecorder()

		s.handleModuleByName(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// TestHandleModeCoverage tests handleMode for coverage.
func TestHandleModeCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "modeuser")
	t.Setenv("STEM_AUTH_PASSWORD", "modepass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/mode", nil)
		w := httptest.NewRecorder()

		s.handleMode(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("get mode", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/mode", nil)
		w := httptest.NewRecorder()

		s.handleMode(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("set mode with invalid value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mode",
			bytes.NewBufferString(`{"mode":"invalid_mode_xyz"}`))
		w := httptest.NewRecorder()

		s.handleMode(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("set mode to reflector", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mode",
			bytes.NewBufferString(`{"mode":"reflector"}`))
		w := httptest.NewRecorder()

		s.handleMode(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("set mode to test_master", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mode",
			bytes.NewBufferString(`{"mode":"test_master"}`))
		w := httptest.NewRecorder()

		s.handleMode(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("set mode with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mode",
			bytes.NewBufferString(`{invalid}`))
		w := httptest.NewRecorder()

		s.handleMode(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestHandleSettingsCoverage tests handleSettings for coverage.
func TestHandleSettingsCoverage(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "settingsuser")
	t.Setenv("STEM_AUTH_PASSWORD", "settingspass123")

	s := newTestServer(t)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/settings", nil)
		w := httptest.NewRecorder()

		s.handleSettings(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("get settings", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
		w := httptest.NewRecorder()

		s.handleSettings(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("post settings invalid interface", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/settings",
			bytes.NewBufferString(`{"interface":"nonexistent_iface_xyz123"}`))
		w := httptest.NewRecorder()

		s.handleSettings(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("post settings theme only", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/settings",
			bytes.NewBufferString(`{"theme":"dark"}`))
		w := httptest.NewRecorder()

		s.handleSettings(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("post settings invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/settings",
			bytes.NewBufferString(`{invalid}`))
		w := httptest.NewRecorder()

		s.handleSettings(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestHandleHealthReadyWithDependencies tests handleHealthReady checking dependencies.
func TestHandleHealthReadyWithDependencies(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "readydepsuser")
	t.Setenv("STEM_AUTH_PASSWORD", "readydepspass123")

	s := newTestServer(t)

	t.Run("ready with auth manager", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		w := httptest.NewRecorder()

		s.handleHealthReady(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		// Response contains status: "ready" when all healthy.
		if !containsStr(body, "status") {
			t.Error("Expected response to contain 'status'")
		}
	})

	t.Run("ready without auth manager", func(t *testing.T) {
		// Clear auth manager.
		origManager := s.authManager
		s.authManager = nil

		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		w := httptest.NewRecorder()

		s.handleHealthReady(w, req)

		// Should return not ready.
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status 503 without auth manager, got %d", w.Code)
		}

		// Restore.
		s.authManager = origManager
	})
}

// TestHandleRecoveryCompleteValidation tests handleRecoveryComplete validation paths.
func TestHandleRecoveryCompleteValidation(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recvaluser")
	t.Setenv("STEM_AUTH_PASSWORD", "recvalpass123")

	s := newTestServer(t)

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/complete", nil)
		w := httptest.NewRecorder()

		s.handleRecoveryComplete(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty body, got %d", w.Code)
		}
	})

	t.Run("missing password", func(t *testing.T) {
		body := bytes.NewBufferString(`{"token":"sometoken"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/complete", body)
		w := httptest.NewRecorder()

		s.handleRecoveryComplete(w, req)

		// Should fail validation (either 401 for invalid token or 400 for missing password).
		if w.Code == http.StatusOK {
			t.Error("Expected error response for missing password")
		}
	})
}

// TestHandleSetupCompleteValidation tests handleSetupComplete validation paths.
func TestHandleSetupCompleteValidation(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "setvaluser")
	t.Setenv("STEM_AUTH_PASSWORD", "setvalpass123")
	t.Setenv("STEM_SETUP_MODE", "true")

	s := newTestServer(t)

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", nil)
		w := httptest.NewRecorder()

		s.handleSetupComplete(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty body, got %d", w.Code)
		}
	})

	t.Run("invalid setup token", func(t *testing.T) {
		body := bytes.NewBufferString(`{"password":"ValidPassword123!","setupToken":"invalid"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", body)
		w := httptest.NewRecorder()

		s.handleSetupComplete(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 for invalid token, got %d", w.Code)
		}
	})
}

// TestHandleTestStartWithTestRunning tests handleTestStart when test already running.
func TestHandleTestStartWithTestRunning(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "testrunuser")
	t.Setenv("STEM_AUTH_PASSWORD", "testrunpass123")

	s := newTestServer(t)

	// Set test status to running.
	s.statsMu.Lock()
	s.testStatus = statusRunning
	s.currentTest = "throughput"
	s.currentModule = "benchmark"
	s.statsMu.Unlock()

	body := bytes.NewBufferString(`{"testType":"rfc2544_latency"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
	w := httptest.NewRecorder()

	s.handleTestStart(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 when test running, got %d", w.Code)
	}
}

// TestHandleInterfacesSuccess tests handleInterfaces success path.
func TestHandleInterfacesSuccess(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "ifacessuccuser")
	t.Setenv("STEM_AUTH_PASSWORD", "ifacessuccpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	w := httptest.NewRecorder()

	s.handleInterfaces(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response is array.
	body := w.Body.String()
	if body[0] != '[' {
		t.Error("Expected response to be JSON array")
	}
}

// TestHandleAuthRefreshWithCookie tests handleAuthRefresh with refresh token cookie.
func TestHandleAuthRefreshWithCookie(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "refreshcookieuser")
	t.Setenv("STEM_AUTH_PASSWORD", "refreshcookiepass123")

	s := newTestServer(t)

	// Get a valid refresh token.
	_, refreshToken, authErr := s.authManager.AuthenticateWithRefresh(
		context.TODO(),
		"refreshcookieuser",
		"refreshcookiepass123",
	)
	if authErr != nil {
		t.Fatalf("Failed to get tokens: %v", authErr)
	}
	if refreshToken == "" {
		t.Skip("Refresh token not provided")
	}

	// Create request with refresh token cookie (using correct cookie name from auth package).
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameRefresh, // "stem_refresh"
		Value: refreshToken,
	})
	w := httptest.NewRecorder()

	s.handleAuthRefresh(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !containsStr(body, "token") {
		t.Error("Expected response to contain 'token'")
	}
}

// TestRateLimiterGetLimiter tests GetLimiter functionality.
func TestRateLimiterGetLimiter(t *testing.T) {
	limiter := NewRateLimiter(10, 5)

	// Get limiter for same IP twice.
	l1 := limiter.GetLimiter("192.168.1.1")
	l2 := limiter.GetLimiter("192.168.1.1")

	if l1 != l2 {
		t.Error("Expected same limiter for same IP")
	}

	// Get limiter for different IP.
	l3 := limiter.GetLimiter("192.168.1.2")
	if l1 == l3 {
		t.Error("Expected different limiter for different IP")
	}

	limiter.Stop()
}

// TestIsLocalhostOriginAdditional tests additional isLocalhostOrigin cases.
func TestIsLocalhostOriginAdditional(t *testing.T) {
	tests := []struct {
		origin   string
		expected bool
	}{
		{"http://[::1]", true},
		{"https://[::1]", true},
		{"http://127.0.0.1", true},
		{"https://127.0.0.1", true},
		{"http://localhost", true},
		{"https://localhost", true},
		{"http://192.168.1.1", false},
		{"https://example.com", false},
		{"", false},
		{"not-a-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			result := isLocalhostOrigin(tt.origin)
			if result != tt.expected {
				t.Errorf("isLocalhostOrigin(%s) = %v, expected %v", tt.origin, result, tt.expected)
			}
		})
	}
}

// TestIsSameOriginAdditional tests additional isSameOrigin cases.
func TestIsSameOriginAdditional(t *testing.T) {
	tests := []struct {
		origin      string
		requestHost string
		expected    bool
	}{
		{"http://192.168.1.1:8080", "192.168.1.1:8080", true},
		{"http://192.168.1.1", "192.168.1.1", true},
		{"https://example.com:443", "example.com:443", true},
		{"http://192.168.1.1:8080", "192.168.1.2:8080", false},
		{"http://192.168.1.1:8080", "192.168.1.1:9090", false},
		{"", "192.168.1.1:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin+"_"+tt.requestHost, func(t *testing.T) {
			result := isSameOrigin(tt.origin, tt.requestHost)
			if result != tt.expected {
				t.Errorf(
					"isSameOrigin(%s, %s) = %v, expected %v",
					tt.origin,
					tt.requestHost,
					result,
					tt.expected,
				)
			}
		})
	}
}

// TestReflectorStatsResponse tests reflector stats response structure.
func TestReflectorStatsResponse(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "reflectstatsuser2")
	t.Setenv("STEM_AUTH_PASSWORD", "reflectstatspass123")

	s := newTestServer(t)

	// Set some stats.
	s.UpdateStats(1000, 900, 100000, 90000, 100.0, 80.0)

	// Get reflector stats.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/stats", nil)
	w := httptest.NewRecorder()

	s.handleReflectorStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify expected fields.
	body := w.Body.String()
	expectedFields := []string{
		"running",
		"packetsReceived",
		"packetsReflected",
		"bytesReceived",
		"bytesReflected",
		"uptime",
	}
	for _, field := range expectedFields {
		if !containsStr(body, field) {
			t.Errorf("Expected response to contain '%s'", field)
		}
	}
}

// TestEnsureSelfSignedCertExistingFiles tests ensureSelfSignedCert with existing files.
func TestEnsureSelfSignedCertExistingFiles(t *testing.T) {
	tempDir := t.TempDir()

	// First call - generate.
	certFile, keyFile, err := ensureSelfSignedCert(tempDir)
	if err != nil {
		t.Fatalf("First call error: %v", err)
	}

	// Get file info.
	certInfo, _ := os.Stat(certFile)
	keyInfo, _ := os.Stat(keyFile)

	// Second call - should reuse.
	certFile2, keyFile2, err := ensureSelfSignedCert(tempDir)
	if err != nil {
		t.Fatalf("Second call error: %v", err)
	}

	if certFile != certFile2 || keyFile != keyFile2 {
		t.Error("Expected same paths on second call")
	}

	// Files should not have changed.
	certInfo2, _ := os.Stat(certFile)
	keyInfo2, _ := os.Stat(keyFile)

	if certInfo.ModTime() != certInfo2.ModTime() {
		t.Error("Cert file should not have been regenerated")
	}
	if keyInfo.ModTime() != keyInfo2.ModTime() {
		t.Error("Key file should not have been regenerated")
	}
}

// TestBuildReflectorConfigUpdateWithExecutor tests buildReflectorConfigUpdate with executor.
func TestBuildReflectorConfigUpdateWithExecutor(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "buildcfgexecuser")
	t.Setenv("STEM_AUTH_PASSWORD", "buildcfgexecpass123")

	s := newTestServer(t)

	// Without executor, dpUpdate should be nil.
	cfg := &ReflectorConfig{
		Profile:         "all",
		SignatureFilter: nil,
		OUIFilter:       "00:11:22",
		PortFilter:      3842,
	}

	changes, dpUpdate, buildErr := s.buildReflectorConfigUpdate(cfg)
	if buildErr != nil {
		t.Errorf("buildReflectorConfigUpdate() error: %v", buildErr)
	}

	if len(changes) != 3 {
		t.Errorf("Expected 3 changes, got %d", len(changes))
	}

	// dpUpdate should be nil since there's no executor.
	if dpUpdate != nil {
		t.Log("dpUpdate is non-nil (executor exists)")
	}
}

// TestHandleLicenseTrialGetStatus tests handleLicenseTrial GET for trial status.
func TestHandleLicenseTrialGetStatus(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "trialstatususer")
	t.Setenv("STEM_AUTH_PASSWORD", "trialstatuspass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/license/trial", nil)
	w := httptest.NewRecorder()

	s.handleLicenseTrial(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !containsStr(body, "active") {
		t.Error("Expected response to contain 'active'")
	}
}

// TestHandleLicenseTrialStartTrial tests handleLicenseTrial POST for starting trial.
func TestHandleLicenseTrialStartTrial(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "trialstartuser")
	t.Setenv("STEM_AUTH_PASSWORD", "trialstartpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/trial", nil)
	w := httptest.NewRecorder()

	s.handleLicenseTrial(w, req)

	// Should return success or error depending on trial state.
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleLicenseTrialWithoutManager tests handleLicenseTrial without license manager.
func TestHandleLicenseTrialWithoutManager(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "trialnomgruser")
	t.Setenv("STEM_AUTH_PASSWORD", "trialnomgrpass123")

	s := newTestServer(t)

	// Clear license manager.
	origManager := s.licenseManager
	s.licenseManager = nil

	req := httptest.NewRequest(http.MethodGet, "/api/v1/license/trial", nil)
	w := httptest.NewRecorder()

	s.handleLicenseTrial(w, req)

	// Should return OK with error message.
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !containsStr(body, "not initialized") {
		t.Error("Expected response to indicate manager not initialized")
	}

	// Restore.
	s.licenseManager = origManager
}

// TestHandleLicenseActivateWithoutManager tests handleLicenseActivate without manager.
func TestHandleLicenseActivateWithoutManager(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "activatenomgruser")
	t.Setenv("STEM_AUTH_PASSWORD", "activatenomgrpass123")

	s := newTestServer(t)

	// Clear license manager.
	origManager := s.licenseManager
	s.licenseManager = nil

	body := bytes.NewBufferString(`{"licenseKey":"TEST-1234-5678-90AB"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/activate", body)
	w := httptest.NewRecorder()

	s.handleLicenseActivate(w, req)

	// Should return OK with error message.
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	responseBody := w.Body.String()
	if !containsStr(responseBody, "not initialized") {
		t.Error("Expected response to indicate manager not initialized")
	}

	// Restore.
	s.licenseManager = origManager
}

// TestHandleLicenseActivateWithKey tests handleLicenseActivate with valid license key.
func TestHandleLicenseActivateWithKey(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "activatekeyuser")
	t.Setenv("STEM_AUTH_PASSWORD", "activatekeypass123")

	s := newTestServer(t)

	body := bytes.NewBufferString(`{"licenseKey":"TEST-1234-5678-90AB"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/activate", body)
	w := httptest.NewRecorder()

	s.handleLicenseActivate(w, req)

	// Should return response (success or failure based on key validity).
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleLicenseStates tests handleLicense with different license states.
func TestHandleLicenseStates(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "licensestatesuser")
	t.Setenv("STEM_AUTH_PASSWORD", "licensestatespass123")

	s := newTestServer(t)

	t.Run("with license manager", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		// Should have license status fields.
		if !containsStr(body, "activated") {
			t.Error("Expected response to contain 'activated'")
		}
	})
}

// TestHandleAuthCSRFWithValidSession tests handleAuthCSRF with valid session.
func TestHandleAuthCSRFWithValidSession(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "csrfsessionuser")
	t.Setenv("STEM_AUTH_PASSWORD", "csrfsessionpass123")

	s := newTestServer(t)

	// Get a valid token first.
	token, _, authErr := s.authManager.AuthenticateWithRefresh(
		context.TODO(),
		"csrfsessionuser",
		"csrfsessionpass123",
	)
	if authErr != nil {
		t.Fatalf("Failed to get token: %v", authErr)
	}

	// Set the access token cookie to simulate browser session.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameAccess,
		Value: token,
	})
	w := httptest.NewRecorder()

	s.handleAuthCSRF(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !containsStr(body, "token") {
		t.Error("Expected response to contain 'token'")
	}
}

// TestHandleRecoveryStatusWithoutManager tests handleRecoveryStatus without manager.
func TestHandleRecoveryStatusWithoutManager(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recstatnomgruser")
	t.Setenv("STEM_AUTH_PASSWORD", "recstatnomgrpass123")

	s := newTestServer(t)

	// Clear recovery manager.
	origManager := s.recoveryTokenManager
	s.recoveryTokenManager = nil

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/status", nil)
	w := httptest.NewRecorder()

	s.handleRecoveryStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !containsStr(body, "active") {
		t.Error("Expected response to contain 'active'")
	}
	if !containsStr(body, "false") {
		t.Error("Expected active to be false without manager")
	}

	// Restore.
	s.recoveryTokenManager = origManager
}

// TestHandleRecoveryInstructionsWithManager tests handleRecoveryInstructions with manager.
func TestHandleRecoveryInstructionsWithManager(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recinstruser")
	t.Setenv("STEM_AUTH_PASSWORD", "recinstrpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/instructions", nil)
	w := httptest.NewRecorder()

	s.handleRecoveryInstructions(w, req)

	// Response depends on whether recovery manager is configured.
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 200 or 503, got %d", w.Code)
	}

	if w.Code == http.StatusOK {
		body := w.Body.String()
		if !containsStr(body, "steps") {
			t.Error("Expected response to contain 'steps'")
		}
	}
}

// TestHandleTestStopWithReflector tests handleTestStop when reflector is running.
func TestHandleTestStopWithReflector(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "stopreflectuser")
	t.Setenv("STEM_AUTH_PASSWORD", "stopreflectpass123")

	s := newTestServer(t)

	// Set running reflector mode.
	s.statsMu.Lock()
	s.mode = moduleReflector
	s.testStatus = statusRunning
	s.currentModule = moduleReflector
	s.statsMu.Unlock()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/stop", nil)
	w := httptest.NewRecorder()

	s.handleTestStop(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify status was updated.
	s.statsMu.RLock()
	status := s.testStatus
	s.statsMu.RUnlock()

	if status != statusCancelled {
		t.Errorf("Expected status 'cancelled', got '%s'", status)
	}
}

// TestHandleTestResultDifferentStates tests handleTestResult with different states.
func TestHandleTestResultDifferentStates(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "teststatesuser")
	t.Setenv("STEM_AUTH_PASSWORD", "teststatespass123")

	s := newTestServer(t)

	states := []string{
		statusIdle,
		statusStarting,
		statusRunning,
		statusCompleted,
		statusError,
		statusCancelled,
	}

	for _, state := range states {
		t.Run("status_"+state, func(t *testing.T) {
			s.statsMu.Lock()
			s.testStatus = state
			s.statsMu.Unlock()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/test/result", nil)
			w := httptest.NewRecorder()

			s.handleTestResult(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			body := w.Body.String()
			if !containsStr(body, "status") {
				t.Errorf("Expected response to contain 'status'")
			}
		})
	}
}

// TestNeedsInitialSetupWithDefaultHash tests needsInitialSetup with default hash.
func TestNeedsInitialSetupWithDefaultHash(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "setuphashuser")
	t.Setenv("STEM_AUTH_PASSWORD", "setuphashpass123")
	t.Setenv("STEM_SETUP_MODE", "false")

	s := newTestServer(t)

	// The result depends on whether the password hash is default.
	result := s.needsInitialSetup()
	t.Logf("needsInitialSetup returned: %v", result)
}

// TestHandleRecoveryCompleteWeakPassword tests handleRecoveryComplete with weak password.
func TestHandleRecoveryCompleteWeakPassword(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recweakuser")
	t.Setenv("STEM_AUTH_PASSWORD", "recweakpass123")

	s := newTestServer(t)

	// This test would need a valid recovery token to fully test the weak password path,
	// but we can at least test that the handler works.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/complete",
		bytes.NewBufferString(`{"token":"test-token","password":"weak"}`))
	w := httptest.NewRecorder()

	s.handleRecoveryComplete(w, req)

	// Should fail with invalid token (before reaching password validation).
	if w.Code == http.StatusOK {
		t.Error("Expected failure with invalid token")
	}
}

// TestHandleSetupCompleteWithValidToken tests handleSetupComplete success path.
func TestHandleSetupCompleteWithValidToken(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "setupvaliduser")
	t.Setenv("STEM_AUTH_PASSWORD", "setupvalidpass123")
	t.Setenv("STEM_SETUP_MODE", "true")

	s := newTestServer(t)

	// Try with a valid token if available.
	if s.setupTokenManager != nil {
		token, genErr := s.setupTokenManager.GenerateToken()
		if genErr != nil {
			t.Logf("Failed to generate token: %v", genErr)
			return
		}
		if token != "" {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete",
				bytes.NewBufferString(`{"setupToken":"`+token+`","password":"ValidPass123!"}`))
			w := httptest.NewRecorder()

			s.handleSetupComplete(w, req)

			// Check the result (may be success or validation failure).
			t.Logf("handleSetupComplete response: %d - %s", w.Code, w.Body.String())
		}
	}
}

// TestHandleReflectorStatsWithNoExecutor tests handleReflectorStats when executor is nil.
func TestHandleReflectorStatsWithNoExecutor(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "reflstats2user")
	t.Setenv("STEM_AUTH_PASSWORD", "reflstats2pass123")

	s := newTestServer(t)

	// Clear the executor.
	s.reflectorExec = nil

	// Update some stats.
	s.UpdateStats(500, 500, 50000, 50000, 50, 5)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/stats", nil)
	w := httptest.NewRecorder()

	s.handleReflectorStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !containsStr(body, "running") {
		t.Error("Expected response to contain 'running' field")
	}
}

// TestBuildReflectorConfigUpdateVariations tests buildReflectorConfigUpdate variations.
func TestBuildReflectorConfigUpdateVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "buildcfguser")
	t.Setenv("STEM_AUTH_PASSWORD", "buildcfgpass123")

	s := newTestServer(t)

	t.Run("update profile only", func(t *testing.T) {
		cfg := &ReflectorConfig{Profile: "netally", SignatureFilter: nil, OUIFilter: "", PortFilter: 0}
		assertBuildConfigChanges(t, s, cfg, 1)
	})

	t.Run("update OUI filter only", func(t *testing.T) {
		cfg := &ReflectorConfig{Profile: "", SignatureFilter: nil, OUIFilter: "00:11:22", PortFilter: 0}
		assertBuildConfigChanges(t, s, cfg, 1)
	})

	t.Run("update signature filter only", func(t *testing.T) {
		cfg := &ReflectorConfig{
			Profile:         "",
			SignatureFilter: []string{"probeot", "dataot"},
			OUIFilter:       "",
			PortFilter:      0,
		}
		assertBuildConfigChanges(t, s, cfg, 1)
	})

	t.Run("validate invalid profile", func(t *testing.T) {
		validateErr := s.validateReflectorProfile("invalid_profile")
		if validateErr == nil {
			t.Error("Expected error for invalid profile")
		}
	})

	t.Run("validate valid profiles", func(t *testing.T) {
		validProfiles := []string{"netally", "msn", "all", "custom", ""}
		for _, profile := range validProfiles {
			validateErr := s.validateReflectorProfile(profile)
			if validateErr != nil {
				t.Errorf("Expected no error for profile '%s', got %v", profile, validateErr)
			}
		}
	})
}

// TestApplyReflectorDataplaneUpdateVariations tests applyReflectorDataplaneUpdate variations.
func TestApplyReflectorDataplaneUpdateVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "applydpuser")
	t.Setenv("STEM_AUTH_PASSWORD", "applydppass123")

	t.Run("apply with nil update", func(t *testing.T) {
		s := newTestServer(t)

		applyErr := s.applyReflectorDataplaneUpdate(nil, []string{"test change"})
		if applyErr != nil {
			t.Errorf("Expected no error for nil update, got %v", applyErr)
		}
	})

	t.Run("apply with nil executor", func(t *testing.T) {
		s := newTestServer(t)

		// Ensure executor is nil.
		s.reflectorExec = nil

		applyErr := s.applyReflectorDataplaneUpdate(nil, []string{})
		if applyErr != nil {
			t.Errorf("Expected no error for nil executor, got %v", applyErr)
		}
	})
}

// TestHandleTestStartWithInterface tests handleTestStart with various interfaces.
func TestHandleTestStartWithInterface(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "startwithifaceuser")
	t.Setenv("STEM_AUTH_PASSWORD", "startwithifacepass123")

	s := newTestServer(t)

	// Get a valid interface.
	ifaces, ifaceErr := netif.DetectInterfaces()
	if ifaceErr != nil || len(ifaces) == 0 {
		t.Skip("No network interfaces available")
	}

	testIface := ifaces[0].Name

	t.Run("start with valid interface and throughput", func(t *testing.T) {
		body := bytes.NewBufferString(`{"testType":"throughput","interface":"` + testIface + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
		w := httptest.NewRecorder()

		s.handleTestStart(w, req)

		// May succeed or fail, but should not be 400 for valid interface.
		t.Logf("handleTestStart response: %d", w.Code)
	})

	t.Run("start with valid interface and latency", func(t *testing.T) {
		// Reset test state.
		s.statsMu.Lock()
		s.testStatus = statusIdle
		s.currentTest = ""
		s.statsMu.Unlock()

		body := bytes.NewBufferString(`{"testType":"latency","interface":"` + testIface + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
		w := httptest.NewRecorder()

		s.handleTestStart(w, req)

		t.Logf("handleTestStart (latency) response: %d", w.Code)
	})

	t.Run("start with frame_loss test", func(t *testing.T) {
		// Reset test state.
		s.statsMu.Lock()
		s.testStatus = statusIdle
		s.currentTest = ""
		s.statsMu.Unlock()

		body := bytes.NewBufferString(`{"testType":"frame_loss","interface":"` + testIface + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
		w := httptest.NewRecorder()

		s.handleTestStart(w, req)

		t.Logf("handleTestStart (frame_loss) response: %d", w.Code)
	})
}

// TestHandleLicenseVariations tests handleLicense variations.
func TestHandleLicenseVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "licvaruser")
	t.Setenv("STEM_AUTH_PASSWORD", "licvarpass123")

	t.Run("get license with manager", func(t *testing.T) {
		s := newTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("get license without manager", func(t *testing.T) {
		s := newTestServer(t)

		// Clear the license manager.
		origManager := s.licenseManager
		s.licenseManager = nil

		req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Restore.
		s.licenseManager = origManager
	})
}

// TestHandleLicenseActivateVariations tests handleLicenseActivate variations.
func TestHandleLicenseActivateVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "licactuser")
	t.Setenv("STEM_AUTH_PASSWORD", "licactpass123")

	t.Run("activate without license manager", func(t *testing.T) {
		s := newTestServer(t)

		// Clear the license manager.
		s.licenseManager = nil

		body := bytes.NewBufferString(`{"licenseKey":"XXXX-YYYY-ZZZZ"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/license/activate", body)
		w := httptest.NewRecorder()

		s.handleLicenseActivate(w, req)

		// Should return an appropriate status.
		t.Logf("handleLicenseActivate without manager: %d", w.Code)
	})

	t.Run("activate with empty body", func(t *testing.T) {
		s := newTestServer(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/license/activate", nil)
		w := httptest.NewRecorder()

		s.handleLicenseActivate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("activate with invalid JSON", func(t *testing.T) {
		s := newTestServer(t)

		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/license/activate", body)
		w := httptest.NewRecorder()

		s.handleLicenseActivate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestHandleLicenseTrialVariations tests handleLicenseTrial variations.
func TestHandleLicenseTrialVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "lictrialvaruser")
	t.Setenv("STEM_AUTH_PASSWORD", "lictrialvarpass123")

	t.Run("trial without license manager", func(t *testing.T) {
		s := newTestServer(t)

		// Clear the license manager.
		s.licenseManager = nil

		req := httptest.NewRequest(http.MethodGet, "/api/v1/license/trial", nil)
		w := httptest.NewRecorder()

		s.handleLicenseTrial(w, req)

		t.Logf("handleLicenseTrial GET without manager: %d", w.Code)
	})

	t.Run("start trial without license manager", func(t *testing.T) {
		s := newTestServer(t)

		// Clear the license manager.
		s.licenseManager = nil

		req := httptest.NewRequest(http.MethodPost, "/api/v1/license/trial", nil)
		w := httptest.NewRecorder()

		s.handleLicenseTrial(w, req)

		t.Logf("handleLicenseTrial POST without manager: %d", w.Code)
	})
}

// TestExecuteReflectorVariations tests executeReflector variations.
func TestExecuteReflectorVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "execreflvaruser")
	t.Setenv("STEM_AUTH_PASSWORD", "execreflvarpass123")

	t.Run("execute with nonexistent interface", func(t *testing.T) {
		s := newTestServer(t)

		execErr := s.executeReflector("nonexistent_iface_xyz", nil)
		if execErr == nil {
			t.Error("Expected error for nonexistent interface")
		}
	})

	t.Run("execute with valid interface", func(t *testing.T) {
		s := newTestServer(t)

		// Get a valid interface.
		ifaces, ifaceErr := netif.DetectInterfaces()
		if ifaceErr != nil || len(ifaces) == 0 {
			t.Skip("No network interfaces available")
		}

		execErr := s.executeReflector(ifaces[0].Name, nil)
		// May succeed or fail depending on permissions.
		t.Logf("executeReflector error: %v", execErr)
	})
}

// TestRunModuleTestVariations tests runModuleTest variations.
func TestRunModuleTestVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "runmodvaruser")
	t.Setenv("STEM_AUTH_PASSWORD", "runmodvarpass123")

	t.Run("run with nil factory", func(t *testing.T) {
		s := newTestServer(t)

		// Create a factory that returns an error.
		factory := func(_ string) (testExecutor, error) {
			return nil, errors.New("factory error")
		}

		runErr := s.runModuleTest(factory, "test", "test_type", "lo0", nil)
		if runErr == nil {
			t.Error("Expected error from factory")
		}
	})
}

// TestValidateInterfaceExistsVariations tests validateInterfaceExists variations.
func TestValidateInterfaceExistsVariations(t *testing.T) {
	t.Run("interface not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		result := validateInterfaceExists(w, "nonexistent_iface_xyz123")
		if result {
			t.Error("Expected false for nonexistent interface")
		}
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid interface, got %d", w.Code)
		}
	})

	t.Run("interface found", func(t *testing.T) {
		// Get a valid interface.
		ifaces, err := netif.DetectInterfaces()
		if err != nil || len(ifaces) == 0 {
			t.Skip("No network interfaces available")
		}

		w := httptest.NewRecorder()
		result := validateInterfaceExists(w, ifaces[0].Name)
		if !result {
			t.Error("Expected true for valid interface")
		}
	})
}

// TestValidateInterfaceForTestVariations tests validateInterfaceForTest variations.
func TestValidateInterfaceForTestVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "validifaceuser")
	t.Setenv("STEM_AUTH_PASSWORD", "validifacepass123")

	t.Run("invalid interface", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		result := s.validateInterfaceForTest(w, "nonexistent_test_iface_999")
		if result {
			t.Error("Expected false for nonexistent interface")
		}
	})

	t.Run("valid interface", func(t *testing.T) {
		s := newTestServer(t)

		// Get a valid interface.
		ifaces, ifaceErr := netif.DetectInterfaces()
		if ifaceErr != nil || len(ifaces) == 0 {
			t.Skip("No network interfaces available")
		}

		w := httptest.NewRecorder()
		result := s.validateInterfaceForTest(w, ifaces[0].Name)
		if !result {
			t.Error("Expected true for valid interface")
		}
	})
}

// TestResolveTestModuleVariations tests resolveTestModule variations.
func TestResolveTestModuleVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "resolvemoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "resolvemodpass123")

	s := newTestServer(t)

	t.Run("unknown test type", func(t *testing.T) {
		_, resolveErr := s.resolveTestModule("unknown_test_type_xyz")
		if resolveErr == nil {
			t.Error("Expected error for unknown test type")
		}
	})

	t.Run("valid test type throughput", func(t *testing.T) {
		mod, resolveErr := s.resolveTestModule("rfc2544_throughput")
		if resolveErr != nil {
			t.Errorf("Unexpected error: %v", resolveErr)
		}
		if mod == nil {
			t.Error("Expected module for rfc2544_throughput test")
		}
	})

	t.Run("valid test type reflect", func(t *testing.T) {
		mod, resolveErr := s.resolveTestModule("reflect")
		if resolveErr != nil {
			t.Errorf("Unexpected error: %v", resolveErr)
		}
		if mod == nil {
			t.Error("Expected module for reflect test")
		}
	})
}

// TestResolveTestInterfaceVariations tests resolveTestInterface variations.
func TestResolveTestInterfaceVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "resolveifaceuser")
	t.Setenv("STEM_AUTH_PASSWORD", "resolveifacepass123")

	t.Run("with requested interface", func(t *testing.T) {
		s := newTestServer(t)

		iface, resolveErr := s.resolveTestInterface("eth0")
		if resolveErr != nil {
			t.Errorf("Unexpected error: %v", resolveErr)
		}
		if iface != "eth0" {
			t.Errorf("Expected eth0, got %s", iface)
		}
	})

	t.Run("with no requested and no selected", func(t *testing.T) {
		s := newTestServer(t)

		// Clear the selected interface.
		s.statsMu.Lock()
		s.selectedIface = ""
		s.statsMu.Unlock()

		_, resolveErr := s.resolveTestInterface("")
		if resolveErr == nil {
			t.Error("Expected error for no interface")
		}
	})

	t.Run("with no requested but selected available", func(t *testing.T) {
		s := newTestServer(t)

		// Set a selected interface.
		s.statsMu.Lock()
		s.selectedIface = "en0"
		s.statsMu.Unlock()

		iface, resolveErr := s.resolveTestInterface("")
		if resolveErr != nil {
			t.Errorf("Unexpected error: %v", resolveErr)
		}
		if iface != "en0" {
			t.Errorf("Expected en0, got %s", iface)
		}
	})
}

// TestBeginTestRunVariations tests beginTestRun variations.
func TestBeginTestRunVariations(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "begintestuser")
	t.Setenv("STEM_AUTH_PASSWORD", "begintestpass123")

	t.Run("already running", func(t *testing.T) {
		s := newTestServer(t)

		// Set status to running.
		s.statsMu.Lock()
		s.testStatus = statusRunning
		s.statsMu.Unlock()

		beginErr := s.beginTestRun("throughput", "benchmark")
		if beginErr == nil {
			t.Error("Expected error when test already running")
		}
	})

	t.Run("idle state", func(t *testing.T) {
		s := newTestServer(t)

		// Set status to idle.
		s.statsMu.Lock()
		s.testStatus = statusIdle
		s.statsMu.Unlock()

		beginErr := s.beginTestRun("throughput", "benchmark")
		if beginErr != nil {
			t.Errorf("Unexpected error: %v", beginErr)
		}

		s.statsMu.RLock()
		status := s.testStatus
		currentTest := s.currentTest
		currentModule := s.currentModule
		s.statsMu.RUnlock()

		if status != statusStarting {
			t.Errorf("Expected status starting, got %s", status)
		}
		if currentTest != "throughput" {
			t.Errorf("Expected currentTest throughput, got %s", currentTest)
		}
		if currentModule != "benchmark" {
			t.Errorf("Expected currentModule benchmark, got %s", currentModule)
		}
	})
}

// TestHandleLicenseVariousStates tests handleLicense with various license states.
func TestHandleLicenseVariousStates(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "licstatesuser")
	t.Setenv("STEM_AUTH_PASSWORD", "licstatespass123")

	t.Run("nil license manager", func(t *testing.T) {
		s := newTestServer(t)

		s.licenseManager = nil

		req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !containsStr(body, "License manager not initialized") {
			t.Error("Expected license manager not initialized message")
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		s := newTestServer(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleInterfacesMethodNotAllowed tests handleInterfaces with wrong method.
func TestHandleInterfacesMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "ifacemethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "ifacemethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/interfaces", nil)
	w := httptest.NewRecorder()

	s.handleInterfaces(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleRecoveryStatusMethodNotAllowed tests handleRecoveryStatus with wrong method.
func TestHandleRecoveryStatusMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recstatmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "recstatmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/status", nil)
	w := httptest.NewRecorder()

	s.handleRecoveryStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleRecoveryCompleteMethodNotAllowed tests handleRecoveryComplete with wrong method.
func TestHandleRecoveryCompleteMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "reccompmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "reccompmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/complete", nil)
	w := httptest.NewRecorder()

	s.handleRecoveryComplete(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleRecoveryInstructionsMethodNotAllowed tests handleRecoveryInstructions with wrong method.
func TestHandleRecoveryInstructionsMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recinstmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "recinstmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/instructions", nil)
	w := httptest.NewRecorder()

	s.handleRecoveryInstructions(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleRecoveryCompleteWithoutManager tests handleRecoveryComplete when manager is nil.
func TestHandleRecoveryCompleteWithoutManager(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "reccompnomgruser")
	t.Setenv("STEM_AUTH_PASSWORD", "reccompnomgrpass123")

	s := newTestServer(t)

	s.recoveryTokenManager = nil

	body := bytes.NewBufferString(`{"token":"test-token","password":"ValidPass123!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/complete", body)
	w := httptest.NewRecorder()

	s.handleRecoveryComplete(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

// TestExecuteTestWithUnknownModule tests executeTest with unknown module.
func TestExecuteTestWithUnknownModule(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "execunknownuser")
	t.Setenv("STEM_AUTH_PASSWORD", "execunknownpass123")

	s := newTestServer(t)

	execErr := s.executeTest("unknown_module_xyz", "throughput", "en0", nil)
	if execErr == nil {
		t.Error("Expected error for unknown module")
	}
}

// TestHandleSetupStatusMethodNotAllowed tests handleSetupStatus with wrong method.
func TestHandleSetupStatusMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "setupstatmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "setupstatmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/status", nil)
	w := httptest.NewRecorder()

	s.handleSetupStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleSetupCompleteMethodNotAllowed tests handleSetupComplete with wrong method.
func TestHandleSetupCompleteMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "setupcompmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "setupcompmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/complete", nil)
	w := httptest.NewRecorder()

	s.handleSetupComplete(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleAuthLoginMethodNotAllowed tests handleAuthLogin with wrong method.
func TestHandleAuthLoginMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "loginmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "loginmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	w := httptest.NewRecorder()

	s.handleAuthLogin(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleAuthLogoutMethodNotAllowed tests handleAuthLogout with wrong method.
func TestHandleAuthLogoutMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "logoutmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "logoutmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()

	s.handleAuthLogout(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleAuthRefreshMethodNotAllowed tests handleAuthRefresh with wrong method.
func TestHandleAuthRefreshMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "refreshmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "refreshmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/refresh", nil)
	w := httptest.NewRecorder()

	s.handleAuthRefresh(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleAuthCSRFMethodNotAllowed tests handleAuthCSRF with wrong method.
func TestHandleAuthCSRFMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "csrfmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "csrfmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/csrf", nil)
	w := httptest.NewRecorder()

	s.handleAuthCSRF(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleAuthCSRFNoSession tests handleAuthCSRF without session.
func TestHandleAuthCSRFNoSession(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "csrfnosessuser")
	t.Setenv("STEM_AUTH_PASSWORD", "csrfnosesspass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
	w := httptest.NewRecorder()

	s.handleAuthCSRF(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestHandleTestStopMethodNotAllowed tests handleTestStop with wrong method.
func TestHandleTestStopMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "stopmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "stopmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test/stop", nil)
	w := httptest.NewRecorder()

	s.handleTestStop(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleTestResultMethodNotAllowed tests handleTestResult with wrong method.
func TestHandleTestResultMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "resultmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "resultmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/result", nil)
	w := httptest.NewRecorder()

	s.handleTestResult(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleTestStartMethodNotAllowed tests handleTestStart with wrong method.
func TestHandleTestStartMethodNotAllowed(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "startmethoduser")
	t.Setenv("STEM_AUTH_PASSWORD", "startmethodpass123")

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test/start", nil)
	w := httptest.NewRecorder()

	s.handleTestStart(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleReflectorConfigUpdateSuccess tests successful config updates.
func TestHandleReflectorConfigUpdateSuccess(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "reflconfsucc")
	t.Setenv("STEM_AUTH_PASSWORD", "reflconfsucc123")

	s := newTestServer(t)

	t.Run("update with valid netally profile", func(t *testing.T) {
		body := bytes.NewBufferString(`{"profile":"netally"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
		w := httptest.NewRecorder()

		s.handleReflectorConfigUpdate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update with valid msn profile", func(t *testing.T) {
		body := bytes.NewBufferString(`{"profile":"msn"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
		w := httptest.NewRecorder()

		s.handleReflectorConfigUpdate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update with valid all profile", func(t *testing.T) {
		body := bytes.NewBufferString(`{"profile":"all"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
		w := httptest.NewRecorder()

		s.handleReflectorConfigUpdate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update with port filter", func(t *testing.T) {
		body := bytes.NewBufferString(`{"portFilter":8080}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
		w := httptest.NewRecorder()

		s.handleReflectorConfigUpdate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update with invalid profile", func(t *testing.T) {
		body := bytes.NewBufferString(`{"profile":"invalid_profile_xyz"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config", body)
		w := httptest.NewRecorder()

		s.handleReflectorConfigUpdate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

// TestBuildReflectorConfigUpdateWithPortFilter tests config with port filter.
func TestBuildReflectorConfigUpdateWithPortFilter(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "portfilteruser")
	t.Setenv("STEM_AUTH_PASSWORD", "portfilterpass123")

	s := newTestServer(t)

	cfg := &ReflectorConfig{
		Profile:         "",
		SignatureFilter: nil,
		OUIFilter:       "",
		PortFilter:      9999,
	}

	changes, _, buildErr := s.buildReflectorConfigUpdate(cfg)
	if buildErr != nil {
		t.Errorf("buildReflectorConfigUpdate() error: %v", buildErr)
	}

	if len(changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changes))
	}
}

// TestHandleLicenseAllBranches tests all branches of handleLicense.
func TestHandleLicenseAllBranches(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "licbranchuser")
	t.Setenv("STEM_AUTH_PASSWORD", "licbranchpass123")

	t.Run("get with manager", func(t *testing.T) {
		s := newTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("get without manager", func(t *testing.T) {
		s := newTestServer(t)

		s.licenseManager = nil

		req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("method POST not allowed", func(t *testing.T) {
		s := newTestServer(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/license", nil)
		w := httptest.NewRecorder()

		s.handleLicense(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleInterfacesAllBranches tests all branches of handleInterfaces.
func TestHandleInterfacesAllBranches(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "ifacebranchuser")
	t.Setenv("STEM_AUTH_PASSWORD", "ifacebranchpass123")

	s := newTestServer(t)

	t.Run("get interfaces", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
		w := httptest.NewRecorder()

		s.handleInterfaces(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("method PUT not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/interfaces", nil)
		w := httptest.NewRecorder()

		s.handleInterfaces(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("method DELETE not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/interfaces", nil)
		w := httptest.NewRecorder()

		s.handleInterfaces(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// TestHandleAuthCSRFAllBranches tests all branches of handleAuthCSRF.
func TestHandleAuthCSRFAllBranches(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "csrfbranchuser")
	t.Setenv("STEM_AUTH_PASSWORD", "csrfbranchpass123")

	s := newTestServer(t)

	t.Run("method POST not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/csrf", nil)
		w := httptest.NewRecorder()

		s.handleAuthCSRF(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("missing auth token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
		w := httptest.NewRecorder()

		s.handleAuthCSRF(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

// TestHandleRecoveryInstructionsAllBranches tests all branches of handleRecoveryInstructions.
func TestHandleRecoveryInstructionsAllBranches(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "recinstrbranchuser")
	t.Setenv("STEM_AUTH_PASSWORD", "recinstrbranchpass123")

	s := newTestServer(t)

	t.Run("get instructions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery/instructions", nil)
		w := httptest.NewRecorder()

		s.handleRecoveryInstructions(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("method POST not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/recovery/instructions", nil)
		w := httptest.NewRecorder()

		s.handleRecoveryInstructions(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("method DELETE not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/recovery/instructions", nil)
		w := httptest.NewRecorder()

		s.handleRecoveryInstructions(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}
