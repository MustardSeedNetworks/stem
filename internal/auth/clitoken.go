// SPDX-License-Identifier: BUSL-1.1

package auth

// Local CLI credential. `stem test` and `stem reflect` are clients of the
// running daemon, not a second orchestration path (#1166), so they need to
// authenticate without an operator at the keyboard — systemd units and
// scripts have no one to prompt. The daemon mints a short-lived JWT for its
// own operator identity and publishes it, with the URL it bound, through
// internal/daemonconn. It is not a second credential type: the token goes
// through the one ValidateToken path a browser session uses, so revocation,
// expiry and the blacklist all apply.

import "time"

const (
	// TokenTypeCLI marks a token minted for the local command line.
	TokenTypeCLI = "cli"

	// CLITokenLifetime bounds what a leaked descriptor is worth. The daemon
	// republishes well before this elapses, so the operator never sees the
	// expiry; a copy taken off the host stops working within a day.
	CLITokenLifetime = 24 * time.Hour
)

// GenerateCLIToken mints the credential the daemon publishes for the local
// command line. It carries the daemon's own operator identity, so the CLI
// can do exactly what that operator can do in the web UI and no more.
func (m *Manager) GenerateCLIToken() (string, error) {
	m.mu.RLock()
	username := m.username
	m.mu.RUnlock()

	return m.generateTokenWithType(username, TokenTypeCLI, CLITokenLifetime)
}
