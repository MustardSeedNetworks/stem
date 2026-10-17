// SPDX-License-Identifier: BUSL-1.1

package auth

// The first-run setup token proves the caller can read the host's data
// directory, the same possession proof password recovery uses. It is never
// sent over HTTP: setup/status answers any network peer, so a token served
// there would let whoever reached the port first claim admin. A cross-site
// form cannot read the file either, so the CSRF protection the token was
// first added for holds as before.

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	// setupTokenLength is the length of setup tokens in bytes before encoding.
	setupTokenLength = 32

	// setupTokenFile is the data-directory file the token is written to.
	setupTokenFile = ".setup-token"
)

// SetupTokenManager issues the first-run setup token to a 0600 file and
// checks the token the operator copies from it. The token lives until setup
// completes; a restart issues a new one.
type SetupTokenManager struct {
	mu      sync.Mutex
	token   string
	dataDir string
}

// NewSetupTokenManager creates a manager that writes its token under dataDir.
// The path is made absolute because the operator is shown it, and the
// default data directory is the service's working directory.
func NewSetupTokenManager(dataDir string) *SetupTokenManager {
	if abs, err := filepath.Abs(dataDir); err == nil {
		dataDir = abs
	}
	return &SetupTokenManager{mu: sync.Mutex{}, token: "", dataDir: dataDir}
}

// FilePath returns the path the token is written to.
func (m *SetupTokenManager) FilePath() string {
	return filepath.Join(m.dataDir, setupTokenFile)
}

// Issue generates a new token and writes it to FilePath, readable only by the
// daemon's user, replacing any earlier token.
func (m *SetupTokenManager) Issue() error {
	tokenBytes := make([]byte, setupTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generate setup token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.dataDir, 0o700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	// Remove first so the 0600 mode applies even when an earlier file was
	// created with a looser one.
	path := m.FilePath()
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove old setup token: %w", err)
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return fmt.Errorf("write setup token: %w", err)
	}
	m.token = token
	return nil
}

// Validate reports whether provided matches the issued token, in constant
// time. It does not consume the token, so a password the policy rejects can
// be retried without a restart.
func (m *SetupTokenManager) Validate(provided string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.token == "" || provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(m.token)) == 1
}

// Consume clears the token and deletes its file once setup has completed.
func (m *SetupTokenManager) Consume() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.token = ""
	if err := os.Remove(m.FilePath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove setup token: %w", err)
	}
	return nil
}
