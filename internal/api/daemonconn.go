// SPDX-License-Identifier: BUSL-1.1

package api

// Publication of the local connection descriptor (#1166). `stem test` and
// `stem reflect` control the daemon rather than opening their own dataplane,
// and a systemd unit has no operator to prompt — or any way to know which
// port the daemon ended up on once fallback (#69) has walked it. The daemon
// therefore publishes both facts, owner-only, in its data directory. See
// internal/daemonconn.

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// connRefreshInterval republishes well inside the token's lifetime, so a CLI
// never reads a credential that is about to expire mid-run.
const connRefreshInterval = auth.CLITokenLifetime / 2

// publishConnection mints a fresh CLI token and publishes it alongside the
// URL this daemon bound, replacing any descriptor it published earlier.
func (s *Server) publishConnection(baseURL string) error {
	token, err := s.authManager.GenerateCLIToken()
	if err != nil {
		return fmt.Errorf("mint CLI token: %w", err)
	}
	descriptor := daemonconn.Descriptor{
		URL:    baseURL,
		Token:  token,
		CAFile: s.publishedCertPath(),
	}
	if publishErr := daemonconn.Publish(s.dataDir, descriptor); publishErr != nil {
		return fmt.Errorf("publish daemon descriptor: %w", publishErr)
	}
	s.publishedURL.Store(&baseURL)
	return nil
}

// withdrawConnection removes a descriptor this daemon published. One
// published by something else is left alone: the test suite and a running
// daemon can share a data directory, and deleting a live daemon's credential
// on the way out of an unrelated Shutdown would break it.
func (s *Server) withdrawConnection() {
	if s.publishedURL.Load() == nil {
		return
	}
	if err := daemonconn.Withdraw(s.dataDir); err != nil {
		logging.Warn("failed to withdraw daemon descriptor", "error", err)
		return
	}
	s.publishedURL.Store(nil)
}

// runConnectionRefresher rotates the published token until ctx is cancelled.
func (s *Server) runConnectionRefresher(ctx context.Context) {
	ticker := time.NewTicker(connRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			baseURL := s.publishedURL.Load()
			if baseURL == nil {
				continue
			}
			if err := s.publishConnection(*baseURL); err != nil {
				logging.Error("failed to rotate daemon descriptor", "error", err)
			}
		}
	}
}

// publishedCertPath is the certificate a client should trust, as an
// absolute path. activeCertPath returns what the daemon serves, which for
// the self-signed default is relative to the daemon's working directory —
// unusable to a CLI invoked from anywhere else, which is every CLI.
func (s *Server) publishedCertPath() string {
	path := s.activeCertPath()
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		logging.Warn("could not resolve the TLS certificate path for publication",
			"path", path, "error", err)
		return path
	}
	return absolute
}
