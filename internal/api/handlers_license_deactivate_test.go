// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api"
)

// licenseHome points the server's licence manager at a temp HOME so a case
// decides what is on disk rather than reading the developer's own licence.
func licenseHome(t testing.TB) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".config", "stem"), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}

func licenseStatusNow(t *testing.T, s *api.Server, token string) api.LicenseStatus {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/license", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/license = %d: %s", w.Code, w.Body.String())
	}
	var status api.LicenseStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode license status: %v", err)
	}
	return status
}

// TestLicenseChangesAreVisibleWithoutARestart is the property #1335 exists
// for: the daemon caches its licence manager, so a change it did not make
// itself is invisible until it restarts. Every change therefore goes through
// the daemon — and when it does, the next read sees it in the same process.
func TestLicenseChangesAreVisibleWithoutARestart(t *testing.T) {
	s := setupLicenseTestServer(t)
	token := getLicenseAuthToken(t, s)

	if before := licenseStatusNow(t, s, token); before.Activated {
		t.Fatalf("a fresh host reports %+v, want no activation", before)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/license/trial", nil)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/license/trial = %d: %s", w.Code, w.Body.String())
	}

	during := licenseStatusNow(t, s, token)
	if !during.IsTrialMode || during.DaysRemaining <= 0 {
		t.Fatalf("after starting a trial the same process reports %+v, want a live trial", during)
	}

	del := httptest.NewRequest(http.MethodDelete, "/api/v1/license", nil)
	authorizeWithCSRF(t, s, del, token)
	dw := httptest.NewRecorder()
	s.ServeHTTP(dw, del)
	if dw.Code != http.StatusOK {
		t.Fatalf("DELETE /api/v1/license = %d: %s", dw.Code, dw.Body.String())
	}
	var removed api.ErrorResponse
	if err := json.Unmarshal(dw.Body.Bytes(), &removed); err != nil {
		t.Fatalf("decode deactivation: %v", err)
	}
	if !removed.Success {
		t.Fatalf("deactivation reported %+v", removed)
	}

	if after := licenseStatusNow(t, s, token); after.IsTrialMode || after.Activated {
		t.Errorf("after deactivation the same process reports %+v, want nothing in force", after)
	}
}

// TestLicenseDeleteRequiresAuth keeps the route under the same gate as the
// rest of /api/v1/license (#1317): removing entitlement is a write.
func TestLicenseDeleteRequiresAuth(t *testing.T) {
	s := setupLicenseTestServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/license", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated DELETE /api/v1/license = %d, want 401: %s", w.Code, w.Body.String())
	}
}
