// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

// routePolicy is the part of a /__capabilities entry these tests read.
type routePolicy struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
	Auth    bool     `json:"auth"`
	CSRF    bool     `json:"csrf"`
	Hidden  bool     `json:"hidden"`
}

// capabilities reads the capability registry the way an auditor would: from
// the manifest the daemon serves.
func capabilities(t *testing.T) []routePolicy {
	t.Helper()
	s := setupAuthTestServer(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/__capabilities", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /__capabilities: status %d", w.Code)
	}
	var policies []routePolicy
	if err := json.Unmarshal(w.Body.Bytes(), &policies); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	return policies
}

func takesMutation(methods []string) bool {
	return slices.ContainsFunc(methods, func(m string) bool {
		return m == http.MethodPost || m == http.MethodPut || m == http.MethodPatch || m == http.MethodDelete
	})
}

// TestEveryRouteDeclaresItsMethods: a route with no method set is gated by
// nothing and documented as taking every method.
func TestEveryRouteDeclaresItsMethods(t *testing.T) {
	for _, p := range capabilities(t) {
		if !p.Hidden && len(p.Methods) == 0 {
			t.Errorf("%s declares no methods, so neither the gate nor the OpenAPI document knows what it takes", p.Path)
		}
	}
}

// TestCSRFRoutePolicy reads the capability registry, not a list kept beside
// it: every route that takes a state-changing method declares CSRF unless it
// is a pre-session sign-in step, and those steps take no session either.
func TestCSRFRoutePolicy(t *testing.T) {
	// The sign-in steps run before a session exists and so cannot carry a
	// token; the credential, mfa_token or WebAuthn assertion is the proof of
	// intent, and the auth limiter caps brute force. Adding a route here is a
	// security decision (#341): it must be a pre-session endpoint, never a
	// normal mutating route.
	preSession := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/login/totp",
		"/api/v1/auth/webauthn/login/begin",
		"/api/v1/auth/webauthn/login/finish",
	}

	seen := map[string]bool{}
	for _, p := range capabilities(t) {
		if p.Hidden || !takesMutation(p.Methods) {
			continue
		}
		switch {
		case slices.Contains(preSession, p.Path):
			seen[p.Path] = true
			if p.CSRF || p.Auth {
				t.Errorf("%s is a pre-session sign-in step but declares csrf=%v auth=%v", p.Path, p.CSRF, p.Auth)
			}
		case !p.CSRF:
			t.Errorf("SECURITY: %s takes %v without CSRF (#341)", p.Path, p.Methods)
		}
	}
	for _, path := range preSession {
		if !seen[path] {
			t.Errorf("%s is listed as a pre-session mutation but the registry does not serve it", path)
		}
	}
}

// TestCookieSessionMutationNeedsItsCSRFToken drives the browser's shape of
// request: the session is the access-token cookie and there is no
// Authorization header. The Registrar must key CSRF on that cookie, the key
// /api/v1/auth/csrf-token mints under, or every mutation from the UI is
// refused (and one keyed on the absent header would refuse them all).
func TestCookieSessionMutationNeedsItsCSRFToken(t *testing.T) {
	s := setupAuthTestServer(t)

	login := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/login",
		bytes.NewBufferString(fmt.Sprintf(`{"username":%q,"password":%q}`, authTestUsername, authTestPassword)))
	lw := httptest.NewRecorder()
	s.ServeHTTP(lw, login)
	if lw.Code != http.StatusOK {
		t.Fatalf("login: status %d: %s", lw.Code, lw.Body)
	}
	cookies := lw.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login set no cookies")
	}
	withSession := func(req *http.Request) *http.Request {
		for _, c := range cookies {
			req.AddCookie(c)
		}
		return req
	}

	tokReq := withSession(httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/auth/csrf-token", nil))
	tw := httptest.NewRecorder()
	s.ServeHTTP(tw, tokReq)
	if tw.Code != http.StatusOK {
		t.Fatalf("csrf-token over the cookie: status %d: %s", tw.Code, tw.Body)
	}
	var tok struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(tw.Body.Bytes(), &tok); err != nil || tok.Token == "" {
		t.Fatalf("csrf-token response %s: %v", tw.Body, err)
	}

	mutate := func(csrfToken string) *httptest.ResponseRecorder {
		req := withSession(httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/settings",
			bytes.NewBufferString(`{"interface":"no-such-interface"}`)))
		if csrfToken != "" {
			req.Header.Set("X-Csrf-Token", csrfToken)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		return w
	}

	refused := mutate("")
	if refused.Code != http.StatusForbidden {
		t.Fatalf("cookie session POST without a token: status %d, want 403: %s", refused.Code, refused.Body)
	}
	var envelope struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(refused.Body.Bytes(), &envelope); err != nil || envelope.Code != "CSRF_TOKEN_MISSING" {
		t.Errorf("refusal body %s, want stem's envelope with code CSRF_TOKEN_MISSING", refused.Body)
	}

	// The handler answers for itself once CSRF passes: an unknown interface is
	// its 400, not the CSRF layer's 403.
	if admitted := mutate(tok.Token); admitted.Code != http.StatusBadRequest {
		t.Errorf("cookie session POST with its token: status %d, want the handler's 400: %s",
			admitted.Code, admitted.Body)
	}
}

// TestRouteManifestIsTheServedTable: docs/openapi.yaml documents
// api.RouteManifest, built without a daemon, so it must be exactly the table a
// running daemon serves on /__capabilities, policy included.
func TestRouteManifestIsTheServedTable(t *testing.T) {
	served := capabilities(t)
	offline := api.RouteManifest()
	if len(offline) != len(served) {
		t.Fatalf("RouteManifest has %d routes, the daemon serves %d", len(offline), len(served))
	}
	for i, p := range offline {
		s := served[i]
		if p.Path != s.Path || !slices.Equal(p.Methods, s.Methods) || p.Auth != s.Auth || p.CSRF != s.CSRF ||
			p.Hidden != s.Hidden {
			t.Errorf("route %d: RouteManifest %+v, served %+v", i, p, s)
		}
	}
}
