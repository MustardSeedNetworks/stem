// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

// The published token is only worth publishing if the daemon's own auth
// middleware accepts it: the CLI presents it on the Bearer path a browser
// session uses, so this is the whole contract of #1166's credential.
const testDaemonURL = "https://localhost:8444"

func TestPublishConnectionIsAcceptedByAuth(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_AUTH_USERNAME", "clitokentest")
	t.Setenv("STEM_AUTH_PASSWORD", "clitokenpass123")
	s := newTestServer(t)
	s.dataDir = dir

	if err := s.publishConnection(testDaemonURL); err != nil {
		t.Fatalf("publishConnection: %v", err)
	}

	descriptor, err := daemonconn.Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	req.Header.Set("Authorization", "Bearer "+descriptor.Token)
	if authErr := s.requireAuth(req); authErr != nil {
		t.Errorf("requireAuth with published CLI token: %v", authErr)
	}
}

// A token file surviving the daemon is a credential nothing can revoke, and
// the next thing to bind the port inherits a client that will present it.
func TestShutdownWithdrawsConnection(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_AUTH_USERNAME", "clitokentest")
	t.Setenv("STEM_AUTH_PASSWORD", "clitokenpass123")
	s := newTestServer(t)
	s.dataDir = dir

	if err := s.publishConnection(testDaemonURL); err != nil {
		t.Fatalf("publishConnection: %v", err)
	}
	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if _, err := os.Stat(daemonconn.Path(dir)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("token file after Shutdown: stat err = %v, want ErrNotExist", err)
	}
}

// Shutdown must not delete a token this server never published — the test
// suite and a real daemon can share a data directory.
func TestShutdownLeavesForeignConnection(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_AUTH_USERNAME", "clitokentest")
	t.Setenv("STEM_AUTH_PASSWORD", "clitokenpass123")
	if err := daemonconn.Publish(dir, daemonconn.Descriptor{URL: testDaemonURL, Token: "another-daemons-token"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	s := newTestServer(t)
	s.dataDir = dir
	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	got, err := daemonconn.Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Token != "another-daemons-token" {
		t.Errorf("token = %q, want the foreign token untouched", got.Token)
	}
}

// Republishing rotates the credential rather than reusing it, so the
// refresher bounds what a copy taken off the host is worth.
func TestPublishConnectionRotatesTheToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_AUTH_USERNAME", "clitokentest")
	t.Setenv("STEM_AUTH_PASSWORD", "clitokenpass123")
	s := newTestServer(t)
	s.dataDir = dir

	if err := s.publishConnection(testDaemonURL); err != nil {
		t.Fatalf("publishConnection: %v", err)
	}
	first, err := daemonconn.Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if rotateErr := s.publishConnection(testDaemonURL); rotateErr != nil {
		t.Fatalf("publishConnection (rotate): %v", rotateErr)
	}
	second, err := daemonconn.Read(dir)
	if err != nil {
		t.Fatalf("Read (rotate): %v", err)
	}

	if first.Token == second.Token {
		t.Error("republish reused the same token; the credential never rotates")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	req.Header.Set("Authorization", "Bearer "+second.Token)
	if authErr := s.requireAuth(req); authErr != nil {
		t.Errorf("requireAuth with rotated token: %v", authErr)
	}
}
