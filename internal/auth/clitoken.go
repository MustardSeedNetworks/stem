// SPDX-License-Identifier: BUSL-1.1

package auth

// Local CLI credential. `stem test` and `stem reflect` are clients of the
// running daemon, not a second orchestration path (#1166), so they need to
// authenticate without an operator at the keyboard — systemd units and
// scripts have no one to prompt. The daemon mints a short-lived JWT for its
// own operator identity and leaves it in the data directory readable only by
// the account it runs as; filesystem permissions are the grant. It is not a
// second credential type: the token goes through the one ValidateToken path a
// browser session uses, so revocation, expiry and the blacklist all apply.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// cliTokenFile is the filename the daemon writes the CLI token to.
	cliTokenFile = ".cli-token"

	// cliTokenMode is the only mode this credential is ever written or read
	// at: readable and writable by the daemon's account, nobody else.
	cliTokenMode os.FileMode = 0o600

	// cliTokenForbiddenMode is the group and world half of the mode. Any bit
	// set there means another account can reach the credential.
	cliTokenForbiddenMode os.FileMode = 0o077

	// TokenTypeCLI marks a token minted for the local command line.
	TokenTypeCLI = "cli"

	// CLITokenLifetime bounds what a leaked token file is worth. The daemon
	// rewrites the file well before this elapses, so the operator never sees
	// the expiry; a copy taken off the host stops working within a day.
	CLITokenLifetime = 24 * time.Hour
)

var (
	// ErrCLITokenNotFound reports that no usable token file is present —
	// normally because no daemon is running on this host.
	ErrCLITokenNotFound = errors.New("no daemon CLI token found")

	// ErrCLITokenPermissions reports a token file other accounts can read or
	// write, which makes it unfit to present as a credential.
	ErrCLITokenPermissions = errors.New("daemon CLI token file is not owner-only")

	// errEmptyCLIToken guards against writing a file that would read back as
	// a valid-looking empty credential.
	errEmptyCLIToken = errors.New("refusing to write an empty CLI token")
)

// CLITokenPath returns the file the daemon publishes its CLI token to.
func CLITokenPath(dataDir string) string {
	return filepath.Join(dataDir, cliTokenFile)
}

// WriteCLIToken publishes token for local CLI use, replacing any previous one.
// The replacement is atomic so a CLI reading concurrently sees either the old
// token or the new one, never a truncated file.
func WriteCLIToken(dataDir, token string) error {
	if strings.TrimSpace(token) == "" {
		return errEmptyCLIToken
	}

	tmp, err := os.CreateTemp(dataDir, cliTokenFile+"-*")
	if err != nil {
		return fmt.Errorf("create CLI token file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once the rename succeeds

	// CreateTemp is already 0600, but the mode is the whole security control
	// here, so it is asserted rather than assumed.
	if chmodErr := tmp.Chmod(cliTokenMode); chmodErr != nil {
		_ = tmp.Close()
		return fmt.Errorf("restrict CLI token file: %w", chmodErr)
	}
	if _, writeErr := tmp.WriteString(token); writeErr != nil {
		_ = tmp.Close()
		return fmt.Errorf("write CLI token: %w", writeErr)
	}
	if closeErr := tmp.Close(); closeErr != nil {
		return fmt.Errorf("close CLI token file: %w", closeErr)
	}
	if renameErr := os.Rename(tmpName, CLITokenPath(dataDir)); renameErr != nil {
		return fmt.Errorf("publish CLI token: %w", renameErr)
	}
	return nil
}

// ReadCLIToken returns the token the local daemon published, refusing a file
// any other account could have read or tampered with.
func ReadCLIToken(dataDir string) (string, error) {
	path := CLITokenPath(dataDir)

	info, statErr := os.Stat(path)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return "", ErrCLITokenNotFound
		}
		return "", fmt.Errorf("stat CLI token: %w", statErr)
	}
	// Windows does not carry POSIX mode bits, so the check there would only
	// reject every file; access control is the ACL on the data directory.
	if runtime.GOOS != "windows" && info.Mode().Perm()&cliTokenForbiddenMode != 0 {
		return "", fmt.Errorf("%w: mode %#o", ErrCLITokenPermissions, info.Mode().Perm())
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return "", fmt.Errorf("read CLI token: %w", readErr)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", ErrCLITokenNotFound
	}
	return token, nil
}

// RemoveCLIToken withdraws the published token. A daemon that has stopped
// leaves no credential behind pointing at whatever binds the port next.
func RemoveCLIToken(dataDir string) error {
	err := os.Remove(CLITokenPath(dataDir))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove CLI token: %w", err)
	}
	return nil
}

// GenerateCLIToken mints the credential published by WriteCLIToken. It
// carries the daemon's own operator identity, so the CLI can do exactly what
// that operator can do in the web UI and no more.
func (m *Manager) GenerateCLIToken() (string, error) {
	m.mu.RLock()
	username := m.username
	m.mu.RUnlock()

	return m.generateTokenWithType(username, TokenTypeCLI, CLITokenLifetime)
}
