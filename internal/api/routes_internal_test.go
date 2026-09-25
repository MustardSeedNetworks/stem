// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// TestSPAFallback: an asset is served as itself, and any other path — a
// client-side route the browser refreshed on — gets index.html so the SPA can
// route it. With no index.html there is nothing to fall back to.
func TestSPAFallback(t *testing.T) {
	ui := fstest.MapFS{
		"index.html":    {Data: []byte("<!doctype html><title>stem</title>")},
		"assets/app.js": {Data: []byte("console.info('app')")},
	}
	tests := []struct {
		name       string
		fs         fstest.MapFS
		path       string
		wantStatus int
		wantBody   string
		wantType   string
	}{
		{name: "asset", fs: ui, path: "/assets/app.js", wantStatus: http.StatusOK, wantBody: "console.info"},
		{
			name:       "client route",
			fs:         ui,
			path:       "/history/run-7",
			wantStatus: http.StatusOK,
			wantBody:   "<title>stem</title>",
			wantType:   "text/html; charset=utf-8",
		},
		{name: "root", fs: ui, path: "/", wantStatus: http.StatusOK, wantBody: "<title>stem</title>"},
		{name: "no index", fs: fstest.MapFS{}, path: "/history", wantStatus: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			spaFallbackHandler(tt.fs)(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil))
			if w.Code != tt.wantStatus {
				t.Fatalf("status %d, want %d", w.Code, tt.wantStatus)
			}
			if !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body %q, want it to contain %q", w.Body, tt.wantBody)
			}
			if tt.wantType != "" && w.Header().Get("Content-Type") != tt.wantType {
				t.Errorf("Content-Type %q, want %q", w.Header().Get("Content-Type"), tt.wantType)
			}
		})
	}
}

// TestRegistrarErrorCode: the Registrar's refusals land in stem's code
// vocabulary, and each CSRF cause keeps a code of its own.
func TestRegistrarErrorCode(t *testing.T) {
	tests := map[string]ErrorCode{
		"method_not_allowed":    ErrCodeMethodNotAllowed,
		"internal_server_error": ErrCodeInternalError,
		"csrf_unavailable":      ErrCodeInternalError,
		"csrf_token_missing":    "CSRF_TOKEN_MISSING",
		"csrf_token_expired":    "CSRF_TOKEN_EXPIRED",
		"csrf_token_invalid":    "CSRF_TOKEN_INVALID",
	}
	for code, want := range tests {
		if got := registrarErrorCode(code); got != want {
			t.Errorf("registrarErrorCode(%q) = %q, want %q", code, got, want)
		}
	}
}
