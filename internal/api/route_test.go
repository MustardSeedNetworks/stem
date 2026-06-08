// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/api"
)

// TestRoutePolicyManifest verifies the capability registry exposes every route
// via /__capabilities and records each route's policy correctly — the registry
// is the single source of truth for route policy (#398 regression class).
func TestRoutePolicyManifest(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "manifesttest")
	t.Setenv("STEM_AUTH_PASSWORD", "manifestpass123")

	s, err := api.NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })

	req := httptest.NewRequest(http.MethodGet, "/__capabilities", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /__capabilities: expected 200, got %d", w.Code)
	}

	var views []struct {
		Path        string `json:"path"`
		Auth        bool   `json:"auth"`
		RateLimited bool   `json:"rateLimited"`
	}
	if decErr := json.Unmarshal(w.Body.Bytes(), &views); decErr != nil {
		t.Fatalf("decode manifest: %v", decErr)
	}
	if len(views) == 0 {
		t.Fatal("expected a non-empty route manifest")
	}

	authByPath := make(map[string]bool, len(views))
	for _, v := range views {
		authByPath[v.Path] = v.Auth
	}

	// Spot-check the recorded policy: the reflector reconfig endpoint requires
	// auth (#398), the pre-session login endpoint does not.
	gotConfig, okConfig := authByPath["/api/v1/reflector/config"]
	if !okConfig || !gotConfig {
		t.Errorf("/api/v1/reflector/config should be in the manifest with auth=true, got auth=%v present=%v",
			gotConfig, okConfig)
	}
	if authByPath["/api/v1/auth/login"] {
		t.Error("/api/v1/auth/login should NOT require auth in the manifest")
	}
}
