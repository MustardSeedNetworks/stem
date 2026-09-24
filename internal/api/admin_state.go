// SPDX-License-Identifier: BUSL-1.1

package api

// The administrator credential outlives the process (#1282). A shipped
// service has no STEM_AUTH_* in its environment, so the daemon has to come
// up unclaimed, let an operator set the password through first-run setup,
// and still hold that password after a restart. It is a file next to
// reflector.json for the same reason that one is: no daemon code opens the
// database.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// credentialStateFile holds the persisted administrator credential.
const credentialStateFile = "credentials.json"

// storedCredential is the administrator as the store records it. Only the
// hash is kept; the password never reaches the disk.
type storedCredential struct {
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
}

// newAuthManager resolves where the administrator credential comes from: a
// stored credential first, so a password set through setup or recovery is
// not reverted by a stale environment; then STEM_AUTH_USERNAME /
// STEM_AUTH_PASSWORD; otherwise the daemon starts unclaimed for first-run
// setup. A store that exists but cannot be used is fatal: falling back to
// unclaimed would hand the daemon to whoever reaches the setup page first.
func newAuthManager(dataDir string) (*auth.Manager, error) {
	secret := os.Getenv("STEM_JWT_SECRET")

	stored, found, err := loadCredential(dataDir)
	if err != nil {
		return nil, err
	}
	if found {
		logging.Info("administrator credential loaded", "event", "auth.credential.source", "source", "store")
		return auth.NewManagerFromHash(secret, defaultAuthSessionTimeout, stored.Username, stored.PasswordHash)
	}

	username, password := os.Getenv("STEM_AUTH_USERNAME"), os.Getenv("STEM_AUTH_PASSWORD")
	if username == "" && password == "" {
		logging.Warn("no administrator credential configured; complete first-run setup in the web UI",
			"event", "auth.credential.source", "source", "unclaimed")
		return auth.NewManagerFromHash(secret, defaultAuthSessionTimeout, auth.FirstRunUsername, "")
	}
	logging.Info("administrator credential loaded", "event", "auth.credential.source", "source", "environment")
	return auth.NewManager(secret, defaultAuthSessionTimeout, username, password)
}

// loadCredential reads the stored credential. An absent file is a daemon
// that has never been claimed, not a failure.
func loadCredential(dataDir string) (storedCredential, bool, error) {
	var stored storedCredential
	data, err := os.ReadFile(filepath.Join(dataDir, credentialStateFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return stored, false, nil
		}
		return stored, false, fmt.Errorf("read credential store: %w", err)
	}
	if unmarshalErr := json.Unmarshal(data, &stored); unmarshalErr != nil {
		return stored, false, fmt.Errorf("decode credential store: %w", unmarshalErr)
	}
	if stored.Username == "" || auth.IsDefaultPasswordHash(stored.PasswordHash) {
		return stored, false, fmt.Errorf("credential store %s holds no usable credential",
			filepath.Join(dataDir, credentialStateFile))
	}
	return stored, true, nil
}

// saveCredential records a new password hash for the administrator, on disk
// first: a hash that only reached memory would be lost at the next restart,
// which is the defect this store exists to close.
func (s *Server) saveCredential(ctx context.Context, passwordHash string) error {
	encoded, err := json.Marshal(storedCredential{
		Username:     s.authManager.GetUsername(),
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("encode credential store: %w", err)
	}
	if writeErr := writeStateFile(s.dataDir, credentialStateFile, encoded); writeErr != nil {
		return writeErr
	}
	s.authManager.UpdatePasswordHash(ctx, passwordHash)
	return nil
}
