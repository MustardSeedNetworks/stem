// SPDX-License-Identifier: BUSL-1.1

package auth_test

// hibp_test.go uses the exported test seams in export_test.go
// (HibpEndpoint, HibpLogger, HibpPrefixLen, HibpUserAgent) so we can
// stay in the `_test` package without leaking those names into
// production builds.

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
)

// withTestEndpoint redirects HIBP API calls to a test URL for the
// duration of the test, restoring the prior value on cleanup.
func withTestEndpoint(t *testing.T, url string) {
	t.Helper()
	prev := auth.SetHibpEndpointForTest(url)
	t.Cleanup(func() { auth.SetHibpEndpointForTest(prev) })
}

// suffixOf returns the 35-char uppercase SHA-1 suffix for a password,
// matching what HIBP would expect us to look for.
func suffixOf(t *testing.T, password string) string {
	t.Helper()
	sum := sha1.Sum([]byte(password))
	return strings.ToUpper(hex.EncodeToString(sum[:]))[auth.HibpPrefixLen:]
}

// NOTE: tests that call withTestEndpoint must NOT run in parallel.
// SetHibpEndpointForTest swaps a package-level global; concurrent
// tests stomp each other's value and read the wrong server's response.
// The hibpMu RWMutex makes the writes data-race-clean, but the
// stomping itself is a logical race that no lock can fix. Keep these
// serial — each test is well under a second.
func TestCheckPasswordBreached_Found(t *testing.T) {
	password := "P@ssw0rd-known-breach"
	wantSuffix := suffixOf(t, password)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != auth.HibpUserAgent {
			t.Errorf("missing/wrong User-Agent header: %q", got)
		}
		fmt.Fprintf(w, "%s:42\r\nFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF:1\r\n", wantSuffix)
	}))
	t.Cleanup(srv.Close)
	withTestEndpoint(t, srv.URL+"/")

	breached, count, err := auth.CheckPasswordBreached(t.Context(), password)
	if err != nil {
		t.Fatalf("CheckPasswordBreached returned error: %v", err)
	}
	if !breached {
		t.Error("expected breached=true")
	}
	if count != 42 {
		t.Errorf("count = %d, want 42", count)
	}
}

func TestCheckPasswordBreached_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return some other suffix so this password isn't found.
		fmt.Fprintf(w, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA:1\r\n")
	}))
	t.Cleanup(srv.Close)
	withTestEndpoint(t, srv.URL+"/")

	breached, count, err := auth.CheckPasswordBreached(t.Context(), "fresh-unique-password-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if breached {
		t.Error("expected breached=false")
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

// TestCheckPasswordBreached_NetworkFailure_DoesNotBlock is the
// air-gapped escape hatch: when the HIBP API is unreachable we must
// degrade to (false, 0, nil) so password changes still go through.
func TestCheckPasswordBreached_NetworkFailure_DoesNotBlock(t *testing.T) {
	// Point at a closed port — Dial will fail immediately.
	withTestEndpoint(t, "http://127.0.0.1:1/")

	captured := struct {
		called bool
	}{}
	prev := auth.SetHibpLoggerForTest(func(_ string, _ error) { captured.called = true })
	t.Cleanup(func() { auth.SetHibpLoggerForTest(prev) })

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	t.Cleanup(cancel)

	breached, count, err := auth.CheckPasswordBreached(ctx, "doesnt-matter-net-is-down")
	if err != nil {
		t.Fatalf("network failure must not return an error, got %v", err)
	}
	if breached || count != 0 {
		t.Errorf("network failure must report (false, 0): got (%v, %d)", breached, count)
	}
	if !captured.called {
		t.Error("network failure should log via hibpLogger")
	}
}

// TestCheckPasswordBreached_Server5xx_DoesNotBlock — non-2xx is treated
// the same as a network failure (degrade-open).
func TestCheckPasswordBreached_Server5xx_DoesNotBlock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)
	withTestEndpoint(t, srv.URL+"/")

	breached, count, err := auth.CheckPasswordBreached(t.Context(), "irrelevant")
	if err != nil {
		t.Fatalf("5xx must not return an error, got %v", err)
	}
	if breached || count != 0 {
		t.Errorf("5xx must report (false, 0): got (%v, %d)", breached, count)
	}
}

// TestCheckPasswordBreached_OptOut verifies STEM_DISABLE_HIBP=1
// short-circuits before any network call.
func TestCheckPasswordBreached_OptOut(t *testing.T) {
	t.Setenv("STEM_DISABLE_HIBP", "1")

	// Endpoint should never be consulted; point at a server that fails
	// the test if hit.
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("HIBP endpoint was contacted despite opt-out")
	}))
	t.Cleanup(srv.Close)
	withTestEndpoint(t, srv.URL+"/")

	breached, count, err := auth.CheckPasswordBreached(t.Context(), "anything")
	if err != nil {
		t.Fatalf("opt-out must not error: %v", err)
	}
	if breached || count != 0 {
		t.Errorf("opt-out must return (false, 0): got (%v, %d)", breached, count)
	}
}
