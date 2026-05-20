// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/api"
)

// Auth test constants.
const (
	authTestUsername = "testadmin"
	authTestPassword = "testpass123"
)

// setupAuthTestServer creates a server with test credentials configured.
func setupAuthTestServer(t testing.TB) *api.Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_AUTH_USERNAME", authTestUsername)
	t.Setenv("STEM_AUTH_PASSWORD", authTestPassword)

	s, err := api.NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// authorizeWithCSRF stamps both the Authorization: Bearer header and the
// X-Csrf-Token header on a test request that targets a CSRF-protected
// endpoint. The CSRF token is minted directly against the server's
// CSRFManager for the session ID embedded in jwt (the JWT payload
// segment). Tests covering state-changing endpoints under
// /api/v1/* must use this helper after the Wave 1 fail-closed change
// (#87) — naked Bearer requests now 403.
func authorizeWithCSRF(t testing.TB, s *api.Server, req *http.Request, jwt string) {
	t.Helper()
	req.Header.Set("Authorization", "Bearer "+jwt)
	sessionID := api.SessionIDFromJWTForTest(jwt)
	if sessionID == "" {
		t.Fatalf("authorizeWithCSRF: empty session ID from JWT")
	}
	csrfToken, err := s.CSRFManagerForTest().GetOrCreateToken(sessionID)
	if err != nil {
		t.Fatalf("authorizeWithCSRF: GetOrCreateToken: %v", err)
	}
	req.Header.Set("X-Csrf-Token", csrfToken)
}

// getAuthToken performs login and returns both access and refresh tokens.
func getAuthToken(t *testing.T, s *api.Server) (string, string) {
	t.Helper()
	body := bytes.NewBufferString(
		fmt.Sprintf(`{"username":"%s","password":"%s"}`, authTestUsername, authTestPassword),
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

	refreshToken, _ := resp["refreshToken"].(string)
	return token, refreshToken
}

// TestHandleAuthLogout tests the POST /api/v1/auth/logout endpoint.
func TestHandleAuthLogout(t *testing.T) {
	// Each subtest gets a fresh server to avoid rate limit interference.

	t.Run("successful logout", func(t *testing.T) {
		s := setupAuthTestServer(t)
		token, _ := getAuthToken(t, s)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		authorizeWithCSRF(t, s, req, token)
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

		if resp["success"] != true {
			t.Errorf("Expected success: true, got %v", resp["success"])
		}
		if resp["message"] == "" {
			t.Error("Expected message in response")
		}
	})

	t.Run("logout without token", func(t *testing.T) {
		s := setupAuthTestServer(t)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("logout with invalid token", func(t *testing.T) {
		s := setupAuthTestServer(t)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-string")
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("logout method not allowed", func(t *testing.T) {
		s := setupAuthTestServer(t)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("token revoked after logout", func(t *testing.T) {
		s := setupAuthTestServer(t)
		token, _ := getAuthToken(t, s)

		// Logout.
		logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		authorizeWithCSRF(t, s, logoutReq, token)
		logoutW := httptest.NewRecorder()
		s.ServeHTTP(logoutW, logoutReq)

		if logoutW.Code != http.StatusOK {
			t.Fatalf("Logout failed: %d", logoutW.Code)
		}

		// Try to use revoked token. CSRF token for this session is
		// still in the map (logout does not revoke CSRF — it revokes
		// the JWT), so we attach it and expect 401 from the auth layer
		// downstream of CSRF validation.
		testReq := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/test/start",
			bytes.NewBufferString(`{"testType":"throughput"}`),
		)
		authorizeWithCSRF(t, s, testReq, token)
		testW := httptest.NewRecorder()
		s.ServeHTTP(testW, testReq)

		if testW.Code != http.StatusUnauthorized {
			t.Errorf("Expected revoked token to return 401, got %d", testW.Code)
		}
	})
}

// TestHandleAuthRefresh_Success tests successful refresh token flow.
func TestHandleAuthRefresh_Success(t *testing.T) {
	s := setupAuthTestServer(t)
	_, refreshToken := getAuthToken(t, s)
	if refreshToken == "" {
		t.Skip("Refresh token not provided by login")
	}

	body := bytes.NewBufferString(fmt.Sprintf(`{"refreshToken":"%s"}`, refreshToken))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", body)
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

	newToken, ok := resp["token"].(string)
	if !ok || newToken == "" {
		t.Error("Expected new token in response")
	}
}

// TestHandleAuthRefresh_InvalidCases tests various invalid refresh scenarios.
func TestHandleAuthRefresh_InvalidCases(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{"invalid token", `{"refreshToken":"invalid-refresh-token"}`, http.StatusUnauthorized},
		{"empty token", `{"refreshToken":""}`, http.StatusBadRequest}, // Empty token is a bad request
		{"invalid JSON", `{invalid json}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupAuthTestServer(t)
			body := bytes.NewBufferString(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", body)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestHandleAuthRefresh_AccessTokenMisuse tests using access token for refresh.
func TestHandleAuthRefresh_AccessTokenMisuse(t *testing.T) {
	s := setupAuthTestServer(t)
	accessToken, _ := getAuthToken(t, s)

	body := bytes.NewBufferString(fmt.Sprintf(`{"refreshToken":"%s"}`, accessToken))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", body)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	// Should fail because access tokens can't be used for refresh.
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 when using access token for refresh, got %d", w.Code)
	}
}

// TestHandleAuthRefresh_MethodNotAllowed tests that GET is not allowed.
func TestHandleAuthRefresh_MethodNotAllowed(t *testing.T) {
	s := setupAuthTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/refresh", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleAuthRefresh_MissingBody tests refresh with no body.
func TestHandleAuthRefresh_MissingBody(t *testing.T) {
	s := setupAuthTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestAuthLoginResponse tests the login response structure.
func TestAuthLoginResponse(t *testing.T) {
	s := setupAuthTestServer(t)

	body := bytes.NewBufferString(
		fmt.Sprintf(`{"username":"%s","password":"%s"}`, authTestUsername, authTestPassword),
	)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
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

	// Check required fields.
	if _, ok := resp["token"]; !ok {
		t.Error("Response missing 'token' field")
	}
	if _, ok := resp["expiresAt"]; !ok {
		t.Error("Response missing 'expiresAt' field")
	}
}

// TestHandleAuthCSRF tests the GET /api/v1/auth/csrf endpoint.
func TestHandleAuthCSRF(t *testing.T) {
	t.Run("requires authentication", func(t *testing.T) {
		s := setupAuthTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		// CSRF endpoint requires authentication.
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 without auth, got %d", w.Code)
		}
	})

	t.Run("success with valid token", func(t *testing.T) {
		s := setupAuthTestServer(t)
		token, _ := getAuthToken(t, s)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
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

		// Should have token field.
		csrfToken, ok := resp["token"].(string)
		if !ok || csrfToken == "" {
			t.Error("Expected 'token' field in response")
		}
	})

	t.Run("method not allowed for POST", func(t *testing.T) {
		s := setupAuthTestServer(t)
		token, _ := getAuthToken(t, s)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/csrf", nil)
		authorizeWithCSRF(t, s, req, token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for POST, got %d", w.Code)
		}
	})

	t.Run("method not allowed for PUT", func(t *testing.T) {
		s := setupAuthTestServer(t)
		token, _ := getAuthToken(t, s)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/csrf", nil)
		authorizeWithCSRF(t, s, req, token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for PUT, got %d", w.Code)
		}
	})

	t.Run("method not allowed for DELETE", func(t *testing.T) {
		s := setupAuthTestServer(t)
		token, _ := getAuthToken(t, s)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/csrf", nil)
		authorizeWithCSRF(t, s, req, token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405 for DELETE, got %d", w.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		s := setupAuthTestServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-string")
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for invalid token, got %d", w.Code)
		}
	})

	t.Run("content type is JSON", func(t *testing.T) {
		s := setupAuthTestServer(t)
		token, _ := getAuthToken(t, s)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})
}

// fetchCSRFToken issues a GET to the canonical Wave 1 endpoint
// /api/v1/auth/csrf-token and returns the token string from the JSON
// response body. Fails the test if the call doesn't return 200.
func fetchCSRFToken(t *testing.T, s *api.Server, jwt string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf-token", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("fetch csrf-token: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode csrf-token: %v", err)
	}
	return resp["token"]
}

// TestHandleAuthCSRFToken_Unauthenticated asserts that the canonical
// Wave 1 endpoint (#87) returns 401 when called without a session.
func TestHandleAuthCSRFToken_Unauthenticated(t *testing.T) {
	s := setupAuthTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf-token", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", w.Code)
	}
}

// TestHandleAuthCSRFToken_Authenticated asserts that the canonical Wave
// 1 endpoint (#87) returns a non-empty token for an authenticated
// session.
func TestHandleAuthCSRFToken_Authenticated(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf-token", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	csrfToken, ok := resp["token"].(string)
	if !ok || csrfToken == "" {
		t.Errorf("expected non-empty 'token' field, got %v", resp)
	}
}

// TestHandleAuthCSRFToken_RepeatedCallsReturnSameToken asserts that the
// canonical Wave 1 endpoint (#87) returns the same CSRF token across
// multiple GETs within the same session — the UI is allowed to cache
// the token and reuse it until next login or expiry.
func TestHandleAuthCSRFToken_RepeatedCallsReturnSameToken(t *testing.T) {
	s := setupAuthTestServer(t)
	token, _ := getAuthToken(t, s)

	first := fetchCSRFToken(t, s, token)
	second := fetchCSRFToken(t, s, token)
	if first == "" {
		t.Fatal("expected non-empty token")
	}
	if first != second {
		t.Errorf(
			"expected same CSRF token across calls for the same session; got %q then %q",
			first, second,
		)
	}
}

// TestHandleAuthCSRFToken_LoginRotatesToken asserts that a fresh login
// revokes the CSRF token bound to whatever session ID the new JWT
// payload decodes to (#87). The new session must start without a CSRF
// token; the UI is expected to call /api/v1/auth/csrf-token after
// logging in.
func TestHandleAuthCSRFToken_LoginRotatesToken(t *testing.T) {
	s := setupAuthTestServer(t)

	first, _ := getAuthToken(t, s)
	_ = fetchCSRFToken(t, s, first)

	firstSessionID := api.SessionIDFromJWTForTest(first)
	if firstSessionID == "" {
		t.Fatal("expected non-empty session ID from JWT payload")
	}
	if !s.CSRFManagerForTest().HasToken(firstSessionID) {
		t.Fatal("first session should have a CSRF token after fetch")
	}

	// Log in a second time. handleAuthLogin revokes the CSRF token for
	// whichever session ID the new JWT decodes to. We verify the new
	// session does NOT inherit a pre-existing token.
	second, _ := getAuthToken(t, s)
	secondSessionID := api.SessionIDFromJWTForTest(second)
	if secondSessionID == "" {
		t.Fatal("expected non-empty session ID from second JWT payload")
	}
	if s.CSRFManagerForTest().HasToken(secondSessionID) {
		t.Errorf(
			"second session must start without a CSRF token (rotation broken); "+
				"session ID %q already has one",
			secondSessionID,
		)
	}
}

// TestAuthRoutesRegistered verifies auth endpoint routes are registered.
func TestAuthRoutesRegistered(t *testing.T) {
	s := setupAuthTestServer(t)

	routes := []struct {
		path   string
		method string
	}{
		{"/api/v1/auth/login", http.MethodPost},
		{"/api/v1/auth/logout", http.MethodPost},
		{"/api/v1/auth/refresh", http.MethodPost},
		{"/api/v1/auth/csrf", http.MethodGet},
		{"/api/v1/auth/csrf-token", http.MethodGet},
	}

	for _, route := range routes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
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

// BenchmarkHandleAuthCSRF benchmarks the CSRF token endpoint.
func BenchmarkHandleAuthCSRF(b *testing.B) {
	b.Setenv("STEM_AUTH_USERNAME", "benchuser")
	b.Setenv("STEM_AUTH_PASSWORD", "benchpass123")

	s, err := api.NewServer(8444)
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
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
	}
}
