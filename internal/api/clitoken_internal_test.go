// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

// The published token is only worth publishing if the daemon's own auth
// middleware accepts it: the CLI presents it on the Bearer path a browser
// session uses, so this is the whole contract of #1166's credential.
func TestPublishCLITokenIsAcceptedByAuth(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_AUTH_USERNAME", "clitokentest")
	t.Setenv("STEM_AUTH_PASSWORD", "clitokenpass123")
	s := newTestServer(t)
	s.dataDir = dir

	if err := s.publishCLIToken(); err != nil {
		t.Fatalf("publishCLIToken: %v", err)
	}

	token, err := auth.ReadCLIToken(dir)
	if err != nil {
		t.Fatalf("ReadCLIToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if authErr := s.requireAuth(req); authErr != nil {
		t.Errorf("requireAuth with published CLI token: %v", authErr)
	}
}

// A token file surviving the daemon is a credential nothing can revoke, and
// the next thing to bind the port inherits a client that will present it.
func TestShutdownWithdrawsCLIToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_AUTH_USERNAME", "clitokentest")
	t.Setenv("STEM_AUTH_PASSWORD", "clitokenpass123")
	s := newTestServer(t)
	s.dataDir = dir

	if err := s.publishCLIToken(); err != nil {
		t.Fatalf("publishCLIToken: %v", err)
	}
	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if _, err := os.Stat(auth.CLITokenPath(dir)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("token file after Shutdown: stat err = %v, want ErrNotExist", err)
	}
}

// Shutdown must not delete a token this server never published — the test
// suite and a real daemon can share a data directory.
func TestShutdownLeavesForeignCLIToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_AUTH_USERNAME", "clitokentest")
	t.Setenv("STEM_AUTH_PASSWORD", "clitokenpass123")
	if err := auth.WriteCLIToken(dir, "another-daemons-token"); err != nil {
		t.Fatalf("WriteCLIToken: %v", err)
	}

	s := newTestServer(t)
	s.dataDir = dir
	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	got, err := auth.ReadCLIToken(dir)
	if err != nil {
		t.Fatalf("ReadCLIToken: %v", err)
	}
	if got != "another-daemons-token" {
		t.Errorf("token = %q, want the foreign token untouched", got)
	}
}

// Republishing rotates the credential rather than reusing it, so the
// refresher bounds what a copy taken off the host is worth.
func TestPublishCLITokenRotates(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	t.Setenv("STEM_AUTH_USERNAME", "clitokentest")
	t.Setenv("STEM_AUTH_PASSWORD", "clitokenpass123")
	s := newTestServer(t)
	s.dataDir = dir

	if err := s.publishCLIToken(); err != nil {
		t.Fatalf("publishCLIToken: %v", err)
	}
	first, err := auth.ReadCLIToken(dir)
	if err != nil {
		t.Fatalf("ReadCLIToken: %v", err)
	}
	if rotateErr := s.publishCLIToken(); rotateErr != nil {
		t.Fatalf("publishCLIToken (rotate): %v", rotateErr)
	}
	second, err := auth.ReadCLIToken(dir)
	if err != nil {
		t.Fatalf("ReadCLIToken (rotate): %v", err)
	}

	if first == second {
		t.Error("republish reused the same token; the credential never rotates")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/interfaces", nil)
	req.Header.Set("Authorization", "Bearer "+second)
	if authErr := s.requireAuth(req); authErr != nil {
		t.Errorf("requireAuth with rotated token: %v", authErr)
	}
}
