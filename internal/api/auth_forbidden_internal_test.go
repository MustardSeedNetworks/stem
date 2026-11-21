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
