// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/update"
)

const updateSettingsPath = "/api/v1/update/settings"

func updateSettingsRequest(
	t *testing.T, s *api.Server, jwt, method, body string,
) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, updateSettingsPath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	authorizeWithCSRF(t, s, req, jwt)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	return w
}

func decodeUpdateSettings(t *testing.T, w *httptest.ResponseRecorder) api.UpdateSettings {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, body %q", w.Code, w.Body.String())
	}
	var got api.UpdateSettings
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", w.Body.String(), err)
	}
	return got
}

// TestUpdateSettingsRoundTrip is ADR-0012's Settings toggle: a fresh install
// reads off, a POST changes the file the root-owned timer reads, and the next
// GET returns what was saved. The source the CLI set survives the toggle.
func TestUpdateSettingsRoundTrip(t *testing.T) {
	s := setupModeTestServer(t)
	dataDir := os.Getenv("STEM_DATA_DIR")
	jwt := loginToken(t, s)

	got := decodeUpdateSettings(t, updateSettingsRequest(t, s, jwt, http.MethodGet, ""))
	if got != (api.UpdateSettings{Automatic: false, Source: update.DefaultSource}) {
		t.Fatalf("fresh install = %+v, want off from %s", got, update.DefaultSource)
	}
	if _, err := os.Stat(filepath.Join(dataDir, update.SettingsName)); !os.IsNotExist(err) {
		t.Fatalf("a GET wrote the opt-in: %v", err)
	}

	const mirror = "/srv/stem-mirror"
	if err := update.SaveSettings(dataDir, update.Settings{Source: mirror}); err != nil {
		t.Fatal(err)
	}
	for _, automatic := range []bool{true, false, true} {
		body, _ := json.Marshal(map[string]bool{"automatic": automatic})
		want := api.UpdateSettings{Automatic: automatic, Source: mirror}
		if got = decodeUpdateSettings(t, updateSettingsRequest(t, s, jwt, http.MethodPost, string(body))); got != want {
			t.Fatalf("POST automatic=%v = %+v, want %+v", automatic, got, want)
		}
		saved, err := update.LoadSettings(dataDir)
		if err != nil {
			t.Fatal(err)
		}
		if saved != (update.Settings{Automatic: automatic, Source: mirror}) {
			t.Fatalf("the timer would read %+v after POST automatic=%v", saved, automatic)
		}
		if got = decodeUpdateSettings(t, updateSettingsRequest(t, s, jwt, http.MethodGet, "")); got != want {
			t.Fatalf("GET after POST automatic=%v = %+v, want %+v", automatic, got, want)
		}
	}
}

// TestUpdateSettingsRefusesAndLeavesTheOptInAlone: a body that does not
// say on or off changes nothing, and so does a request without a session or
// CSRF token.
func TestUpdateSettingsRefusesAndLeavesTheOptInAlone(t *testing.T) {
	tests := []struct {
		name string
		body string
		send func(t *testing.T, s *api.Server, body string) *httptest.ResponseRecorder
		want int
	}{
		{name: "missing automatic", body: `{}`, want: http.StatusBadRequest},
		{name: "unknown field", body: `{"automatic":true,"source":"https://evil"}`, want: http.StatusBadRequest},
		{name: "not a boolean", body: `{"automatic":"on"}`, want: http.StatusBadRequest},
		{
			name: "no session",
			body: `{"automatic":true}`,
			send: func(_ *testing.T, s *api.Server, body string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodPost, updateSettingsPath, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				s.ServeHTTP(w, req)
				return w
			},
			want: http.StatusUnauthorized,
		},
		{
			name: "no CSRF token",
			body: `{"automatic":true}`,
			send: func(t *testing.T, s *api.Server, body string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodPost, updateSettingsPath, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+loginToken(t, s))
				w := httptest.NewRecorder()
				s.ServeHTTP(w, req)
				return w
			},
			want: http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupModeTestServer(t)
			send := tt.send
			if send == nil {
				send = func(t *testing.T, s *api.Server, body string) *httptest.ResponseRecorder {
					return updateSettingsRequest(t, s, loginToken(t, s), http.MethodPost, body)
				}
			}
			if w := send(t, s, tt.body); w.Code != tt.want {
				t.Fatalf("status %d, want %d, body %q", w.Code, tt.want, w.Body.String())
			}
			_, err := os.Stat(filepath.Join(os.Getenv("STEM_DATA_DIR"), update.SettingsName))
			if !os.IsNotExist(err) {
				t.Fatalf("a refused request wrote the opt-in: %v", err)
			}
		})
	}
}
