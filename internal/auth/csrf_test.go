// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MustardSeedNetworks/foundation/pkg/csrf"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

func newTestLogger() *slog.Logger {
	opts := &slog.HandlerOptions{}
	opts.Level = slog.LevelDebug
	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

// cookieValue builds the JWT-shaped access-token cookie value the middleware
// reads. The CSRF session key is sha256(cookieValue) (foundation.SessionKey),
// so a test that wants the middleware to accept a token must mint it under that
// same derived key — see sessionKeyFor.
func cookieValue(payload string) string {
	return "header." + payload + ".signature"
}

// sessionKeyFor returns the CSRF session key the middleware derives for a
// request carrying cookieValue(payload) — i.e. what GetSessionIDFromRequest
// returns. Tests mint tokens under this key so validation matches.
func sessionKeyFor(payload string) string {
	return csrf.SessionKey(cookieValue(payload))
}

func TestNewCSRFManager(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	if manager == nil {
		t.Fatal("NewCSRFManager returned nil")
	}
	defer manager.Stop()

	// A fresh manager mints and validates a token — the behavioral proxy for
	// "initialized correctly" now that the store lives in foundation.
	tok, err := manager.GenerateToken("session")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if valErr := manager.ValidateToken("session", tok); valErr != nil {
		t.Errorf("fresh manager failed to validate its own token: %v", valErr)
	}
}

func TestCSRFManagerGenerateAndValidate(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	sessionID := "test-session"

	token, err := manager.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if validateErr := manager.ValidateToken(sessionID, token); validateErr != nil {
		t.Errorf("failed to validate token: %v", validateErr)
	}

	if wrongSessionErr := manager.ValidateToken("wrong-session", token); !errors.Is(
		wrongSessionErr, csrf.ErrTokenInvalid,
	) {
		t.Errorf("expected ErrTokenInvalid, got %v", wrongSessionErr)
	}

	if wrongTokenErr := manager.ValidateToken(sessionID, "wrong-token"); !errors.Is(
		wrongTokenErr, csrf.ErrTokenInvalid,
	) {
		t.Errorf("expected ErrTokenInvalid, got %v", wrongTokenErr)
	}
}

func TestCSRFManagerEmptyToken(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	sessionID := "test-session"
	if _, err := manager.GenerateToken(sessionID); err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if emptyTokenErr := manager.ValidateToken(sessionID, ""); !errors.Is(
		emptyTokenErr, csrf.ErrTokenMissing,
	) {
		t.Errorf("expected ErrTokenMissing, got %v", emptyTokenErr)
	}
}

func TestCSRFManagerRevokeToken(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	sessionID := "test-session"
	token, err := manager.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if validateErr := manager.ValidateToken(sessionID, token); validateErr != nil {
		t.Errorf("failed to validate token: %v", validateErr)
	}

	manager.RevokeToken(sessionID)

	if revokedErr := manager.ValidateToken(sessionID, token); !errors.Is(
		revokedErr, csrf.ErrTokenInvalid,
	) {
		t.Errorf("expected ErrTokenInvalid after revoke, got %v", revokedErr)
	}
}

// TestGetSessionIDFromRequest_HashedKeying pins the security-invariant change:
// the CSRF session key is now sha256(bearer) (foundation's SessionKey), not the
// raw JWT payload segment. Both a header bearer and a cookie hash to a
// non-plaintext 64-char key, and a request with no token yields "".
func TestGetSessionIDFromRequest_HashedKeying(t *testing.T) {
	bearer := "header.payload-segment.signature"
	r := httptest.NewRequest(http.MethodPost, "/api/v1/x", nil)
	r.Header.Set("Authorization", "Bearer "+bearer)

	key := auth.GetSessionIDFromRequest(r)
	if key == "payload-segment" || key == bearer {
		t.Errorf("session key must not be the raw token/payload segment (got %q)", key)
	}
	if len(key) != 64 { // sha256 hex
		t.Errorf("session key len = %d, want 64 (sha256 hex)", len(key))
	}

	empty := httptest.NewRequest(http.MethodPost, "/api/v1/x", nil)
	if k := auth.GetSessionIDFromRequest(empty); k != "" {
		t.Errorf("no-token request should yield empty key, got %q", k)
	}
}

func TestCSRFMiddlewareSafeMethods(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	safeMethods := []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace}

	for _, method := range safeMethods {
		called = false
		req := httptest.NewRequest(method, "/api/v1/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if !called {
			t.Errorf("%s request should skip CSRF check", method)
		}
		if w.Code != http.StatusOK {
			t.Errorf("%s request should return 200, got %d", method, w.Code)
		}
	}
}

func TestCSRFMiddlewareNonAPIPath(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/static/file.js", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("non-API path should skip CSRF check")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCSRFMiddlewareExemptEndpoints(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// Pre-session / IdP-callback endpoints that cannot carry a CSRF token
	// (see isCSRFExemptPath in csrf.go).
	exemptPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/harvest/logs/client",
		"/api/v1/auth/login/totp",
		"/api/v1/auth/webauthn/login/begin",
		"/api/v1/auth/webauthn/login/finish",
		"/api/v1/sso/google/callback",
		"/api/v1/sso/saml/acs",
	}

	for _, path := range exemptPaths {
		called = false
		req := httptest.NewRequest(http.MethodPost, path, nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if !called {
			t.Errorf("exempt endpoint %s should skip CSRF check", path)
		}
		if w.Code != http.StatusOK {
			t.Errorf("exempt endpoint %s should return 200, got %d", path, w.Code)
		}
	}
}

// TestCSRFMiddlewareFailClosedPreviouslyExempt asserts that endpoints
// removed from the exempt list (#87) now require a valid CSRF token on
// state-changing methods: an authenticated session with an existing CSRF
// token submitting POST/PUT/PATCH/DELETE without the X-Csrf-Token header → 403.
func TestCSRFMiddlewareFailClosedPreviouslyExempt(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	// Mint a token under the key the middleware will derive for this cookie,
	// so the request reaches the missing-header branch (not "never minted").
	payload := "eyJ1c2VybmFtZSI6ImFkbWluIn0"
	if _, err := manager.GenerateToken(sessionKeyFor(payload)); err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	formerlyExemptPaths := []string{
		"/api/v1/auth/refresh",
		"/api/v1/auth/logout",
		"/api/v1/setup/complete",
	}
	methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, path := range formerlyExemptPaths {
		for _, method := range methods {
			req := httptest.NewRequest(method, path, nil)
			req.AddCookie(&http.Cookie{Name: auth.CookieNameAccess, Value: cookieValue(payload)})
			// Intentionally no X-Csrf-Token header.
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Errorf("%s %s without CSRF token: expected 403, got %d", method, path, w.Code)
			}
		}
	}
}

func TestCSRFMiddlewareMissingSessionID(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// No session cookie: CSRF middleware passes through — auth middleware
	// handles the authentication check.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called when no session (auth handles authentication)")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCSRFMiddlewareMissingToken(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	payload := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	if _, err := manager.GenerateToken(sessionKeyFor(payload)); err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieNameAccess, Value: cookieValue(payload)})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("handler should not be called without CSRF token")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCSRFMiddlewareValidToken(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	payload := "eyJ1c2VybmFtZSI6ImFkbWluIn0"
	// Mint under the middleware-derived key so validation matches.
	csrfToken, err := manager.GenerateToken(sessionKeyFor(payload))
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieNameAccess, Value: cookieValue(payload)})
	req.Header.Set(auth.CSRFHeaderName, csrfToken)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called with valid CSRF token")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCSRFMiddlewareInvalidToken(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	payload := "eyJ1c2VybmFtZSI6ImFkbWluIn0"
	if _, err := manager.GenerateToken(sessionKeyFor(payload)); err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieNameAccess, Value: cookieValue(payload)})
	req.Header.Set(auth.CSRFHeaderName, "wrong-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Error("handler should not be called with invalid CSRF token")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestCSRFManagerStop(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	if _, err := manager.GenerateToken("session1"); err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	// Stop delegates to foundation's cleanup goroutine; it must not hang.
	manager.Stop()
}
