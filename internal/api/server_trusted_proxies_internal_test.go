// SPDX-License-Identifier: BUSL-1.1

package api

// server_trusted_proxies_internal_test.go covers the startup wiring of
// STEM_TRUSTED_PROXIES (#962): a bad list must stop the daemon rather than be
// ignored, and a good one must reach the package that keys the auth rate
// limiter and the failed-login tracker. The `_internal_test.go` suffix is
// what the testpackage linter accepts for an internal-package test.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/api/ratelimit"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// trustedProxyEnv sets the credentials NewServer requires plus the list under
// test.
func trustedProxyEnv(t *testing.T, list string) {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_AUTH_USERNAME", "proxytest")
	t.Setenv("STEM_AUTH_PASSWORD", "proxypass123")
	t.Setenv(logging.TrustedProxiesEnv, list)
}

// proxyPeer is the address the reverse proxy connects from in these tests;
// it is inside the trusted CIDR one case configures and outside every CIDR in
// the other.
const proxyPeer = "10.0.0.5:44321"

// authLimitStatus sends one request through the server's auth rate limiter,
// arriving from proxyPeer and naming xff as the client, and reports the
// status the limiter produced.
func authLimitStatus(s *Server, xff string) int {
	handler := s.authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	r.RemoteAddr = proxyPeer
	r.Header.Set("X-Forwarded-For", xff)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	return w.Code
}

func TestNewServerRejectsWildcardTrustedProxy(t *testing.T) {
	for _, list := range []string{"0.0.0.0/0", "::/0", "10.0.0.0/24,0.0.0.0/0", "10.0.0.0"} {
		trustedProxyEnv(t, list)

		s, err := NewServer(8444)
		if err == nil {
			t.Errorf("NewServer() accepted %s=%q", logging.TrustedProxiesEnv, list)
			_ = s.Shutdown()
			continue
		}
		if !strings.Contains(err.Error(), logging.TrustedProxiesEnv) {
			t.Errorf("NewServer() error for %q = %v, want it to name %s",
				list, err, logging.TrustedProxiesEnv)
		}
	}
}

// The list must reach the auth rate limiter, not merely be stored: behind a
// configured proxy each client gets its own bucket, so one client cannot
// spend another's burst (#962). AuthBurstLimit is 5, so a sixth request from
// one client is the discriminating case.
func TestNewServerTrustedProxiesReachTheAuthLimiter(t *testing.T) {
	trustedProxyEnv(t, "10.0.0.0/24")

	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })

	// One client behind the trusted proxy spends its own burst.
	for i := range ratelimit.AuthBurstLimit {
		if code := authLimitStatus(s, "198.51.100.7"); code != http.StatusOK {
			t.Fatalf("request %d from the first client = %d, want 200", i+1, code)
		}
	}
	if code := authLimitStatus(s, "198.51.100.7"); code != http.StatusTooManyRequests {
		t.Errorf("the first client's burst was not spent: %d, want 429", code)
	}

	// A second client through the same proxy is a different bucket.
	if code := authLimitStatus(s, "203.0.113.9"); code != http.StatusOK {
		t.Errorf("second client behind the same proxy = %d, want 200 (its own bucket)", code)
	}
}

// An unset variable must leave the loopback-only behaviour of #807 exactly as
// it was: the setting is additive and cannot loosen an existing deployment.
// Without it, the same two clients share the proxy's bucket.
func TestNewServerWithoutTrustedProxiesSharesOneBucket(t *testing.T) {
	trustedProxyEnv(t, "")

	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })

	if len(s.trustedProxies) != 0 {
		t.Errorf("trustedProxies = %v, want empty", s.trustedProxies)
	}

	for i := range ratelimit.AuthBurstLimit {
		if code := authLimitStatus(s, "198.51.100.7"); code != http.StatusOK {
			t.Fatalf("request %d = %d, want 200", i+1, code)
		}
	}
	if code := authLimitStatus(s, "203.0.113.9"); code != http.StatusTooManyRequests {
		t.Errorf("a different forwarded client got its own bucket: %d, want 429", code)
	}
}
