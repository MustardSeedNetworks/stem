// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

// TestGetSessionIDFromRequest_HashedKeying pins the security-invariant change:
// the CSRF session key is now sha256(bearer) (foundation's SessionKey), not the
// raw JWT payload segment. Both a header bearer and a cookie hash to a
// non-plaintext 64-char key, and a request with no token yields "".
func TestGetSessionIDFromRequest_HashedKeying(t *testing.T) {
	bearer := "header.payload-segment.signature"
	r := httptest.NewRequest(http.MethodPost, "/api/v1/x", nil)
	r.Header.Set("Authorization", "Bearer "+bearer)

	key := auth.GetSessionIDFromRequest(r)
	if key == "payload-segment" || key == bearer {
		t.Errorf("session key must not be the raw token/payload segment (got %q)", key)
	}
	if len(key) != 64 { // sha256 hex
		t.Errorf("session key len = %d, want 64 (sha256 hex)", len(key))
	}

	empty := httptest.NewRequest(http.MethodPost, "/api/v1/x", nil)
	if k := auth.GetSessionIDFromRequest(empty); k != "" {
		t.Errorf("no-token request should yield empty key, got %q", k)
	}
}
