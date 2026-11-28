// SPDX-License-Identifier: BUSL-1.1

package api

// The administrator account outlives the process (#1282): the password hash,
// the TOTP secret, the passkeys and their user handle live in foundation's
// credstore, an encrypted file under <datadir>/auth. A shipped service has no
// STEM_AUTH_* in its environment, so the daemon comes up unclaimed, lets an
// operator set the password through first-run setup, and still holds it, and
// every second factor enrolled since, after a restart.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MustardSeedNetworks/foundation/pkg/credstore"
	"github.com/MustardSeedNetworks/foundation/pkg/password"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

const (
	// credentialStoreDir is the directory, inside the data directory, that
	// belongs to the credential store alone: securedir judges everything in a
	// directory it checks, so the store does not share one.
	credentialStoreDir = "auth"

	// legacyCredentialFile held only a username and a password hash, and never
	// a second factor. It is not read: an install that has one comes up
	// unclaimed and is set up again (foundation ADR 0005).
	legacyCredentialFile = "credentials.json"
)

// newAuthManager resolves where the administrator account comes from: the
// credential store first, so a password set through setup or recovery is not
// reverted by a stale environment; then, for a store with no account yet,
// STEM_AUTH_USERNAME / STEM_AUTH_PASSWORD, which seed the store once;
// otherwise the daemon starts unclaimed for first-run setup. A store that
// exists but cannot be used is fatal: falling back to unclaimed would hand the
// daemon to whoever reaches the setup page first.
func newAuthManager(dataDir string) (*auth.Manager, error) {
	secret := os.Getenv("STEM_JWT_SECRET")

	store, err := openCredentialStore(dataDir)
	if err != nil {
		return nil, err
	}
	if _, accountErr := store.Account(); accountErr == nil {
		logging.Info("administrator credential loaded", "event", "auth.credential.source", "source", "store")
		return auth.NewManagerFromStore(secret, defaultAuthSessionTimeout, store, "")
	} else if !errors.Is(accountErr, credstore.ErrUnclaimed) {
		return nil, fmt.Errorf("credential store: %w", accountErr)
	}

	if _, statErr := os.Stat(filepath.Join(dataDir, legacyCredentialFile)); statErr == nil {
		logging.Warn(
			"an old credentials.json is present and is not read; set the password again through first-run setup",
			"event",
			"auth.credential.legacy_file",
			"path",
			filepath.Join(dataDir, legacyCredentialFile),
		)
	}

	username, plaintext := os.Getenv("STEM_AUTH_USERNAME"), os.Getenv("STEM_AUTH_PASSWORD")
	if username == "" && plaintext == "" {
		logging.Warn("no administrator credential configured; complete first-run setup in the web UI",
			"event", "auth.credential.source", "source", "unclaimed")
		return auth.NewManagerFromStore(secret, defaultAuthSessionTimeout, store, auth.FirstRunUsername)
	}
	if username == "" || plaintext == "" {
		return nil, auth.ErrMissingCredentials
	}
	hash, err := password.Hash(plaintext)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", auth.ErrPasswordHashFailed, err)
	}
	mgr, err := auth.NewManagerFromStore(secret, defaultAuthSessionTimeout, store, username)
	if err != nil {
		return nil, err
	}
	if err = mgr.ReplaceCredential(hash, false); err != nil {
		return nil, err
	}
	removeLegacyCredential(dataDir)
	logging.Info("administrator credential loaded", "event", "auth.credential.source", "source", "environment")
	return mgr, nil
}

// openCredentialStore opens the store under the data directory with the trust
// that fits the identity this process runs as.
func openCredentialStore(dataDir string) (*credstore.Store, error) {
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("resolve the data directory %q: %w", dataDir, err)
	}
	trust, err := credstore.ProcessTrust()
	if err != nil {
		return nil, err
	}
	store, err := credstore.Open(filepath.Join(abs, credentialStoreDir), credstore.Options{Trust: trust})
	if err != nil {
		return nil, fmt.Errorf("open the credential store: %w", err)
	}
	return store, nil
}

// saveCredential records a new password hash for the administrator, in the
// store first: a hash that only reached memory would be lost at the next
// restart. recovery also clears the second factors, in the same write. The old
// credentials.json is removed once the store holds the account, so a superseded
// hash is not left on disk.
func (s *Server) saveCredential(passwordHash string, recovery bool) error {
	if err := s.authManager.ReplaceCredential(passwordHash, recovery); err != nil {
		return err
	}
	removeLegacyCredential(s.dataDir)
	return nil
}

func removeLegacyCredential(dataDir string) {
	legacy := filepath.Join(dataDir, legacyCredentialFile)
	if err := os.Remove(legacy); err != nil && !errors.Is(err, os.ErrNotExist) {
		logging.Warn("could not remove the old credentials.json", "path", legacy, "error", err)
	}
}
