// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The validator, its DTO tags and validateStruct all existed before #268 —
// and validateStruct had zero callers outside its own unit test, so five DTOs
// carried `validate:` tags that never ran. These drive the HTTP handlers, not
// the helper, because "the rule is declared" and "the rule is enforced" are
// exactly the two things that had come apart.
func TestHandlersRejectInvalidDTOs(t *testing.T) {
	s := setupTestServer(t)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantIn     string
	}{
		{
			name:       "login with no username",
			method:     http.MethodPost,
			path:       "/api/v1/auth/login",
			body:       `{"username":"","password":"testpass123"}`,
			wantStatus: http.StatusBadRequest,
			wantIn:     "username: is required",
		},
		{
			name:       "login with no password",
			method:     http.MethodPost,
			path:       "/api/v1/auth/login",
			body:       `{"username":"testadmin","password":""}`,
			wantStatus: http.StatusBadRequest,
			wantIn:     "password: is required",
		},
		{
			name:       "login with neither reports both fields",
			method:     http.MethodPost,
			path:       "/api/v1/auth/login",
			body:       `{"username":"","password":""}`,
			wantStatus: http.StatusBadRequest,
			wantIn:     "username: is required; password: is required",
		},
		{
			name:       "refresh with an empty token",
			method:     http.MethodPost,
			path:       "/api/v1/auth/refresh",
			body:       `{"refreshToken":""}`,
			wantStatus: http.StatusBadRequest,
			wantIn:     "refreshToken: is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), tt.wantIn) {
				t.Errorf("body = %s\nwant it to contain %q", w.Body.String(), tt.wantIn)
			}
		})
	}
}

// An empty login must not reach the authenticator. Before the wiring it did:
// AuthenticateWithRefresh was called with two empty strings, failed, and the
// failure was recorded as an audit event and a failed-login attempt — so an
// unauthenticated caller could fill the suspicious-activity tracker with
// requests that were never credentials in the first place.
func TestEmptyLoginIsRejectedBeforeAuthentication(t *testing.T) {
	s := setupTestServer(t)

	req := httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"","password":""}`),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	// 400 (malformed request), not 401 (wrong credentials): there were no
	// credentials to be wrong.
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 — an empty body is a bad request, not a failed login", w.Code)
	}
}

// The accepted mode values are declared once, on the DTO, as
// `oneof=reflector test_master`. The handler used to repeat them in an if,
// which is how a tag and a check get to disagree.
func TestModeUpdateRejectsAnUnknownMode(t *testing.T) {
	s := setupTestServer(t)
	token := loginToken(t, s)

	// /api/v1/mode is a mutating route, so it is behind the CSRF wrapper.
	csrfReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf-token", nil)
	csrfReq.Header.Set("Authorization", "Bearer "+token)
	csrfRec := httptest.NewRecorder()
	s.ServeHTTP(csrfRec, csrfReq)
	if csrfRec.Code != http.StatusOK {
		t.Fatalf("csrf-token status = %d (body: %s)", csrfRec.Code, csrfRec.Body.String())
	}
	var csrfResp map[string]any
	if err := json.Unmarshal(csrfRec.Body.Bytes(), &csrfResp); err != nil {
		t.Fatalf("decoding csrf response: %v", err)
	}
	csrfToken, ok := csrfResp["token"].(string)
	if !ok || csrfToken == "" {
		t.Fatalf("csrf response carried no token: %v", csrfResp)
	}

	req := httptest.NewRequest(
		http.MethodPost, "/api/v1/mode", bytes.NewBufferString(`{"mode":"not_a_mode"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Csrf-Token", csrfToken)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
	// The message must name the accepted values, not the library's rule.
	// `mode: oneof` would be a downgrade from the hand-written check it
	// replaced, which is why describeValidationTag exists.
	body := w.Body.String()
	for _, want := range []string{"mode:", "reflector", "test_master"} {
		if !strings.Contains(body, want) {
			t.Errorf("body = %s\nwant it to contain %q", body, want)
		}
	}
}
