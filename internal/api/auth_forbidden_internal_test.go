// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

// captureLogs routes the default logger into a buffer for the test and
// returns the decoded records.
func captureLogs(t *testing.T) func() []map[string]any {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	return func() []map[string]any {
		var records []map[string]any
		for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
			if line == "" {
				continue
			}
			var rec map[string]any
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				t.Fatalf("log line is not JSON: %v: %q", err, line)
			}
			records = append(records, rec)
		}
		return records
	}
}

func findForbidden(records []map[string]any) map[string]any {
	for _, rec := range records {
		if rec["event"] == "auth.forbidden" {
			return rec
		}
	}
	return nil
}

func TestAuthMiddlewareLogsForbiddenOnDenial(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{name: "no credentials"},
		{name: "malformed authorization header", header: "Token abc"},
		{name: "unverifiable bearer token", header: "Bearer not-a-valid-jwt"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer(t)
			records := captureLogs(t)

			reached := false
			h := s.authMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached = true }))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/mode", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if reached || w.Code != http.StatusUnauthorized {
				t.Fatalf("reached=%v status=%d, want a 401 that stops the request", reached, w.Code)
			}
			rec := findForbidden(records())
			if rec == nil {
				t.Fatal("no event=auth.forbidden record was logged")
			}
			if rec["reason"] != "credentials" || rec["path"] != "/api/v1/mode" || rec["method"] != http.MethodPost {
				t.Errorf("record = %v, want reason=credentials path=/api/v1/mode method=POST", rec)
			}
			if _, ok := rec["client_ip"]; !ok {
				t.Error("record has no client_ip")
			}
		})
	}
}

func TestLoginFailureLogsForbidden(t *testing.T) {
	s := newTestServer(t)
	records := captureLogs(t)

	body := `{"username":"admin","password":"definitely-the-wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.loginWithMFAGate(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
	rec := findForbidden(records())
	if rec == nil {
		t.Fatal("no event=auth.forbidden record was logged")
	}
	if rec["reason"] != "credentials" || rec["username"] != "admin" || rec["path"] != "/api/v1/auth/login" {
		t.Errorf("record = %v, want reason=credentials username=admin path=/api/v1/auth/login", rec)
	}
}

const lockoutTestPassword = "kettle-orbit-granite-velvet-42"

func lockoutLogin(s *Server, password string) *httptest.ResponseRecorder {
	body := `{"username":"admin","password":"` + password + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.loginWithMFAGate(w, req)
	return w
}

// TestLoginLockoutRefusesCorrectPasswordAndLogsLocked drives the handler
// directly, so the per-address limiter in front of it does not stand in for
// the lockout: this is the distributed guess the lockout exists to stop.
func TestLoginLockoutRefusesCorrectPasswordAndLogsLocked(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "admin")
	t.Setenv("STEM_AUTH_PASSWORD", lockoutTestPassword)
	s := newTestServer(t)
	records := captureLogs(t)

	for range auth.MaxFailedLogins {
		if w := lockoutLogin(s, "definitely-the-wrong-password"); w.Code != http.StatusUnauthorized {
			t.Fatalf("wrong password status = %d, want 401", w.Code)
		}
	}
	wrong := lockoutLogin(s, "definitely-the-wrong-password")
	locked := lockoutLogin(s, lockoutTestPassword)
	if locked.Code != http.StatusUnauthorized {
		t.Fatalf("correct password while locked: status = %d, want 401", locked.Code)
	}
	if locked.Body.String() != wrong.Body.String() {
		t.Errorf("locked answer %q differs from wrong-password answer %q", locked.Body.String(), wrong.Body.String())
	}

	var sawLocked bool
	for _, rec := range records() {
		if rec["event"] == "auth.forbidden" && rec["reason"] == "locked" {
			sawLocked = true
			if rec["username"] != "admin" || rec["path"] != "/api/v1/auth/login" {
				t.Errorf("locked record = %v, want username=admin path=/api/v1/auth/login", rec)
			}
		}
	}
	if !sawLocked {
		t.Error("no event=auth.forbidden reason=locked record was logged")
	}

	s.authManager.ResetLoginFailures()
	if w := lockoutLogin(s, lockoutTestPassword); w.Code != http.StatusOK {
		t.Fatalf("login after reset: status = %d, want 200: %s", w.Code, w.Body.String())
	}
}

// TestLoginLockoutCountsWrongTOTPCodes: a caller who knows the password must
// not be able to guess second-factor codes without limit.
func TestLoginLockoutCountsWrongTOTPCodes(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "admin")
	t.Setenv("STEM_AUTH_PASSWORD", lockoutTestPassword)
	s := newTestServer(t)

	setup, err := auth.GenerateTOTPSecret("admin", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatal(err)
	}
	s.authManager.SetTOTPSecret(setup.Secret)
	code, err := totp.GenerateCode(setup.Secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.authManager.VerifyAndEnableTOTP(code); err != nil {
		t.Fatal(err)
	}

	for range auth.MaxFailedLogins {
		w := lockoutLogin(s, lockoutTestPassword)
		if w.Code != http.StatusOK {
			t.Fatalf("password stage status = %d: %s", w.Code, w.Body.String())
		}
		var resp mfaRequiredResponse
		if err = json.Unmarshal(w.Body.Bytes(), &resp); err != nil || !resp.MFARequired {
			t.Fatalf("expected an mfa challenge, got %s", w.Body.String())
		}
		body := `{"mfaToken":"` + resp.MFAToken + `","code":"000000"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/totp", strings.NewReader(body))
		fw := httptest.NewRecorder()
		s.handleLoginTOTP(fw, req)

		// The token is spent by the wrong code: even the right code is refused.
		right, genErr := totp.GenerateCode(setup.Secret, time.Now().UTC())
		if genErr != nil {
			t.Fatal(genErr)
		}
		again := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/totp",
			strings.NewReader(`{"mfaToken":"`+resp.MFAToken+`","code":"`+right+`"}`))
		aw := httptest.NewRecorder()
		s.handleLoginTOTP(aw, again)
		if aw.Code != http.StatusUnauthorized {
			t.Fatalf("spent mfa token accepted: status %d", aw.Code)
		}
	}
	if !s.authManager.LoginLocked() {
		t.Fatal("repeated wrong TOTP codes did not lock the account")
	}
}

// TestLoginAnswersAreIdenticalForLockedUnknownAndWrong: nothing in the
// response may say which of the three happened; only the server log does.
func TestLoginAnswersAreIdenticalForLockedUnknownAndWrong(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "admin")
	t.Setenv("STEM_AUTH_PASSWORD", lockoutTestPassword)
	s := newTestServer(t)

	login := func(username, password string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
			strings.NewReader(`{"username":"`+username+`","password":"`+password+`"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.loginWithMFAGate(w, req)
		return w
	}
	wrong := login("admin", "definitely-the-wrong-password")
	unknown := login("nobody", "definitely-the-wrong-password")
	for range auth.MaxFailedLogins {
		login("admin", "definitely-the-wrong-password")
	}
	locked := login("admin", lockoutTestPassword)

	for name, w := range map[string]*httptest.ResponseRecorder{"unknown": unknown, "locked": locked} {
		if w.Code != wrong.Code || w.Body.String() != wrong.Body.String() {
			t.Errorf("%s answer (%d %q) differs from wrong password (%d %q)",
				name, w.Code, w.Body.String(), wrong.Code, wrong.Body.String())
		}
	}
}

func TestTooManyAttemptsAnswers429WithRetryAfter(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	s.writeAuthError(w, auth.ErrTooManyAttempts)
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") == "" {
		t.Fatalf("status %d Retry-After %q, want 429 with Retry-After", w.Code, w.Header().Get("Retry-After"))
	}
}
