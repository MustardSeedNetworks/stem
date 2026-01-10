// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package auth_test

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
)

//nolint:exhaustruct // Test helper doesn't need all fields
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestNewCSRFManager(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	if manager == nil {
		t.Fatal("NewCSRFManager returned nil")
	}

	if manager.TokenCount() != 0 {
		t.Fatal("tokens map should be empty")
	}

	// Clean up
	manager.Stop()
}

func TestCSRFManagerCleanupStopsOnContextCancel(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())

	// Generate a token to ensure the manager is working
	token, err := manager.GenerateToken("test-session")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Stop the manager
	manager.Stop()

	// Give the goroutine a moment to stop
	time.Sleep(100 * time.Millisecond)

	// Verify context is canceled
	select {
	case <-manager.CtxDone():
		// Expected - context is canceled
	default:
		t.Fatal("context should be canceled after Stop()")
	}
}

func TestCSRFManagerGenerateAndValidate(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	sessionID := "test-session"

	// Generate a token
	token, err := manager.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate the token
	validateErr := manager.ValidateToken(sessionID, token)
	if validateErr != nil {
		t.Errorf("failed to validate token: %v", validateErr)
	}

	// Validate with wrong session ID
	wrongSessionErr := manager.ValidateToken("wrong-session", token)
	if !errors.Is(wrongSessionErr, auth.ErrCSRFTokenInvalid) {
		t.Errorf("expected ErrCSRFTokenInvalid, got %v", wrongSessionErr)
	}

	// Validate with wrong token
	wrongTokenErr := manager.ValidateToken(sessionID, "wrong-token")
	if !errors.Is(wrongTokenErr, auth.ErrCSRFTokenInvalid) {
		t.Errorf("expected ErrCSRFTokenInvalid, got %v", wrongTokenErr)
	}
}

func TestCSRFManagerEmptyToken(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	sessionID := "test-session"

	// Generate a token
	_, err := manager.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate with empty token
	emptyTokenErr := manager.ValidateToken(sessionID, "")
	if !errors.Is(emptyTokenErr, auth.ErrCSRFTokenMissing) {
		t.Errorf("expected ErrCSRFTokenMissing, got %v", emptyTokenErr)
	}
}

func TestCSRFManagerRevokeToken(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	sessionID := "test-session"

	// Generate a token
	token, err := manager.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate should succeed
	validateErr := manager.ValidateToken(sessionID, token)
	if validateErr != nil {
		t.Errorf("failed to validate token: %v", validateErr)
	}

	// Revoke the token
	manager.RevokeToken(sessionID)

	// Validate should fail
	revokedErr := manager.ValidateToken(sessionID, token)
	if !errors.Is(revokedErr, auth.ErrCSRFTokenInvalid) {
		t.Errorf("expected ErrCSRFTokenInvalid after revoke, got %v", revokedErr)
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

	// Non-API paths should skip CSRF
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

func TestCSRFMiddlewareAuthEndpoints(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	exemptPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/auth/logout",
	}

	for _, path := range exemptPaths {
		called = false
		req := httptest.NewRequest(http.MethodPost, path, nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if !called {
			t.Errorf("auth endpoint %s should skip CSRF check", path)
		}
		if w.Code != http.StatusOK {
			t.Errorf("auth endpoint %s should return 200, got %d", path, w.Code)
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

	// POST to protected endpoint without session cookie
	// CSRF middleware should pass through - let auth middleware handle authentication
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Request should pass through since no session = no CSRF protection needed
	// Auth middleware will handle authentication check
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

	// Create a mock JWT cookie (header.payload.signature format)
	sessionID := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"

	// Generate a CSRF token for this session
	_, err := manager.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// POST with session cookie but no CSRF token header
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameAccess,
		Value: "header." + sessionID + ".signature",
	})
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

	// Create a mock JWT payload (this will be used as session ID)
	sessionID := "eyJ1c2VybmFtZSI6ImFkbWluIn0"

	// Generate a CSRF token for this session
	csrfToken, err := manager.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// POST with both session cookie and CSRF token header
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameAccess,
		Value: "header." + sessionID + ".signature",
	})
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

	sessionID := "eyJ1c2VybmFtZSI6ImFkbWluIn0"

	// Generate a CSRF token for this session
	_, err := manager.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	called := false
	handler := manager.CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// POST with session cookie but WRONG CSRF token
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.CookieNameAccess,
		Value: "header." + sessionID + ".signature",
	})
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

func TestCSRFManagerTokenCount(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())
	defer manager.Stop()

	if manager.TokenCount() != 0 {
		t.Errorf("expected 0 tokens, got %d", manager.TokenCount())
	}

	_, err := manager.GenerateToken("session1")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if manager.TokenCount() != 1 {
		t.Errorf("expected 1 token, got %d", manager.TokenCount())
	}

	_, err = manager.GenerateToken("session2")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if manager.TokenCount() != 2 {
		t.Errorf("expected 2 tokens, got %d", manager.TokenCount())
	}

	manager.RevokeToken("session1")

	if manager.TokenCount() != 1 {
		t.Errorf("expected 1 token after revoke, got %d", manager.TokenCount())
	}
}

func TestCSRFManagerStop(t *testing.T) {
	manager := auth.NewCSRFManager(newTestLogger())

	// Generate some tokens
	_, err := manager.GenerateToken("session1")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Verify context is not canceled initially
	select {
	case <-manager.CtxDone():
		t.Fatal("context should not be canceled initially")
	default:
		// Expected - context is still active
	}

	// Stop the manager
	manager.Stop()

	// Give goroutine time to exit
	time.Sleep(50 * time.Millisecond)

	// Verify context is canceled
	select {
	case <-manager.CtxDone():
		// Expected - context is canceled
	default:
		t.Fatal("context should be canceled after Stop()")
	}
}
