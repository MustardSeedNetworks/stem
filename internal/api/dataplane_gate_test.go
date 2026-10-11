// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

const noDataplaneReason = "CGO + Linux required"

func setupNoDataplaneServer(t *testing.T) *api.Server {
	t.Helper()
	s := setupTestServer(t)
	s.UseDataplaneAvailabilityForTest(func() (bool, string) { return false, noDataplaneReason })
	return s
}

// TestDataplaneGate_RefusesRequestsThatNeedTheDataplane pins stem#1650: a
// binary without a dataplane (the darwin and Windows builds) refuses a
// reflector or test start up front with the capability's reason, as a client
// error rather than the 500 the dataplane failure used to surface as.
func TestDataplaneGate_RefusesRequestsThatNeedTheDataplane(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
	}{
		{"reflector config", "/api/v1/reflector/config", `{"profile":"all","interface":"en0","autostart":true}`},
		{"reflector start", "/api/v1/test/start", `{"tests":[{"testType":"reflect"}]}`},
		{
			"test start", "/api/v1/test/start",
			`{"peer":"192.0.2.1","tests":[{"testType":"rfc2544_throughput"}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupNoDataplaneServer(t)
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			stampAuth(t, s, req, loginToken(t, s))
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body=%q", w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), noDataplaneReason) {
				t.Errorf("body = %q, want the capability reason %q", w.Body.String(), noDataplaneReason)
			}
		})
	}
}

// TestDataplaneGate_RefusedReflectorConfigPersistsNothing covers the second
// half of stem#1650: `stem reflect --at-boot` configures before it starts, and
// the saved autostart made every later daemon start log a failure.
func TestDataplaneGate_RefusedReflectorConfigPersistsNothing(t *testing.T) {
	s := setupNoDataplaneServer(t)
	jwt := loginToken(t, s)

	post := httptest.NewRequest(http.MethodPost, "/api/v1/reflector/config",
		bytes.NewBufferString(`{"profile":"all","interface":"en0","autostart":true}`))
	post.Header.Set("Content-Type", "application/json")
	stampAuth(t, s, post, jwt)
	s.ServeHTTP(httptest.NewRecorder(), post)

	get := httptest.NewRequest(http.MethodGet, "/api/v1/reflector/config", nil)
	stampAuth(t, s, get, jwt)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, get)
	if w.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200; body=%q", w.Code, w.Body.String())
	}
	var cfg api.ReflectorConfig
	if err := json.Unmarshal(w.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if cfg.Autostart || cfg.Interface != "" {
		t.Errorf("refused config was applied: autostart=%v interface=%q", cfg.Autostart, cfg.Interface)
	}
}
