// SPDX-License-Identifier: BUSL-1.1

package api

// Publication of the local CLI credential (#1166). `stem test` and
// `stem reflect` control the daemon rather than opening their own dataplane,
// and a systemd unit has no operator to prompt, so the daemon leaves a
// short-lived token for its own identity in the data directory where only
// the account it runs as can read it. See internal/auth/clitoken.go.

import (
	"context"
	"fmt"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// cliTokenRefreshInterval republishes well inside the token's lifetime, so a
// CLI never reads a credential that is about to expire mid-run.
const cliTokenRefreshInterval = auth.CLITokenLifetime / 2

// publishCLIToken mints a fresh CLI token and writes it to the data
// directory, replacing any token this daemon published earlier.
func (s *Server) publishCLIToken() error {
	token, err := s.authManager.GenerateCLIToken()
	if err != nil {
		return fmt.Errorf("mint CLI token: %w", err)
	}
	if writeErr := auth.WriteCLIToken(s.dataDir, token); writeErr != nil {
		return fmt.Errorf("publish CLI token: %w", writeErr)
	}
	s.cliTokenPublished.Store(true)
	return nil
}

// withdrawCLIToken removes a token this daemon published. A token published
// by something else is left alone: the test suite and a running daemon can
// share a data directory, and deleting a live daemon's credential on the way
// out of an unrelated Shutdown would break it.
func (s *Server) withdrawCLIToken() {
	if !s.cliTokenPublished.Load() {
		return
	}
	if err := auth.RemoveCLIToken(s.dataDir); err != nil {
		logging.Warn("failed to withdraw CLI token", "error", err)
		return
	}
	s.cliTokenPublished.Store(false)
}

// runCLITokenRefresher rotates the published token until ctx is cancelled.
func (s *Server) runCLITokenRefresher(ctx context.Context) {
	ticker := time.NewTicker(cliTokenRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.publishCLIToken(); err != nil {
				logging.Error("failed to rotate CLI token", "error", err)
			}
		}
	}
}
