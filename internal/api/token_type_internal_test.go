// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The exposure in #1169 was at the HTTP edge, not in the token library: a
// refresh token presented as a Bearer credential authenticated every API
// request. This pins the edge, since that is what an attacker reaches.
func TestRequireAuthRejectsARefreshToken(t *testing.T) {
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", "tokentypetest")
	t.Setenv("STEM_AUTH_PASSWORD", "tokentypepass123")
	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })

	refresh, err := s.authManager.GenerateRefreshToken("tokentypetest")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	req.Header.Set("Authorization", "Bearer "+refresh)
	if authErr := s.requireAuth(req); authErr == nil {
		t.Error("requireAuth accepted a refresh token as an API credential")
	}

	// The credentials that should still work, so the fix is not a blanket
	// refusal: a browser session and the local CLI.
	cli, err := s.authManager.GenerateCLIToken()
	if err != nil {
		t.Fatalf("GenerateCLIToken: %v", err)
	}
	cliReq := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	cliReq.Header.Set("Authorization", "Bearer "+cli)
	if authErr := s.requireAuth(cliReq); authErr != nil {
		t.Errorf("requireAuth rejected the CLI credential: %v", authErr)
	}
}
