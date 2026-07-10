// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MustardSeedNetworks/foundation/pkg/csrf"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

// TestCSRFExemptList_Golden pins the CSRF exempt-list (#341). It is the
// regression gate: every currently-exempt path must stay exempt, and a
// representative set of state-changing routes must stay PROTECTED. Changing
// isCSRFExemptPath without updating this table fails CI — forcing a reviewer to
// confirm a new exemption is a pre-session endpoint that cannot carry a CSRF
// token, never a normal mutating route. Do NOT paste a new value in blind.
func TestCSRFExemptList_Golden(t *testing.T) {
	t.Parallel()
	exempt := []string{
		"/api/v1/auth/login",                 // pre-session
		"/api/v1/harvest/logs/client",        // fire-and-forget client log
		"/api/v1/auth/login/totp",            // MFA finisher, pre-session (mfa_token proves intent)
		"/api/v1/auth/webauthn/login/begin",  // MFA finisher, pre-session
		"/api/v1/auth/webauthn/login/finish", // MFA finisher, pre-session
		"/api/v1/sso/google/callback",        // prefix match
		"/api/v1/sso/",
	}
	for _, p := range exempt {
		if !auth.ExportIsCSRFExemptPath(p) {
			t.Errorf("expected %q to be CSRF-exempt, but it is NOT — did the exempt-list shrink?", p)
		}
	}

	// State-changing routes that MUST require a CSRF token. Note stem does NOT
	// exempt refresh/logout (unlike a session-cookie-refresh model) — they are
	// protected. If any becomes exempt, that is a security regression.
	protected := []string{
		"/api/v1/test/start",
		"/api/v1/config",
		"/api/v1/mode",
		"/api/v1/auth/refresh",
		"/api/v1/auth/logout",
		"/api/v1/sso",                   // prefix needs the trailing slash
		"/api/v1/harvest/logs/client/x", // exact entry, not a prefix
	}
	for _, p := range protected {
		if auth.ExportIsCSRFExemptPath(p) {
			t.Errorf("SECURITY: %q is CSRF-exempt but must be protected (#341)", p)
		}
	}
}

// TestCSRFMiddleware_BlocksAuthenticatedMutation is the integration test (#341).
// stem's middleware lets a session-less request through (the auth layer 401s it
// downstream) and only enforces CSRF for an *authenticated* session, so this
// injects a fake bearer token (unverified — GetSessionIDFromRequest just splits
// it) to exercise the real 403 path: an authenticated mutating request with no /
// wrong CSRF token is blocked, while the correct token passes.
func TestCSRFMiddleware_BlocksAuthenticatedMutation(t *testing.T) {
	t.Parallel()
	m := auth.NewCSRFManager(newTestLogger())
	defer m.Stop()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	h := m.CSRFMiddleware(next)

	// Fake bearer token → GetSessionIDFromRequest returns sha256(bearer) as
	// the session key (foundation keying). The signature is never verified
	// here; the middleware only hashes the extracted token.
	const bearer = "hdr.sess.sig"
	sessionID := csrf.SessionKey(bearer)

	post := func(withToken bool) *httptest.ResponseRecorder {
		called = false
		req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", nil)
		req.Header.Set("Authorization", "Bearer "+bearer)
		if withToken {
			tok, err := m.GenerateToken(sessionID)
			if err != nil {
				t.Fatalf("GenerateToken: %v", err)
			}
			req.Header.Set(auth.CSRFHeaderName, tok)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	// Authenticated POST, no CSRF token → 403, handler not called.
	rec := post(false)
	if called {
		t.Error("CSRF middleware let an authenticated POST without a token reach the handler")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("authenticated POST without CSRF token: status = %d, want 403", rec.Code)
	}

	// Authenticated POST with the matching token → passes.
	if post(true); !called {
		t.Error("CSRF middleware blocked an authenticated POST carrying a valid token")
	}

	// Exempt route passes through even with no session.
	called = false
	exReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	h.ServeHTTP(httptest.NewRecorder(), exReq)
	if !called {
		t.Error("CSRF middleware blocked an exempt POST /api/v1/auth/login")
	}
}
