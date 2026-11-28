// SPDX-License-Identifier: BUSL-1.1

package auth

// The administrator account outlives the process. The Manager keeps a working
// copy of it for the login path, and every change goes to foundation's
// credstore first: a password, a TOTP secret or a passkey that reached only
// memory would be gone at the next restart or upgrade, which is the defect
// this store exists to close. The copy is replaced by what the store
// committed, never by what the caller asked for.

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/MustardSeedNetworks/foundation/pkg/credstore"
)

// ErrNoCredentialStore is returned by a change to the account on a Manager
// built without a store. Only tests build one; the daemon always has a store.
var ErrNoCredentialStore = errors.New("manager has no credential store")

// NewManagerFromStore creates the daemon's auth manager over a credential
// store. An account in the store is the administrator, whatever
// unclaimedUsername says; an empty store is an unclaimed daemon, offered for
// first-run setup under unclaimedUsername.
func NewManagerFromStore(
	jwtSecret string,
	sessionTimeout time.Duration,
	store *credstore.Store,
	unclaimedUsername string,
) (*Manager, error) {
	m, err := NewManagerFromHash(jwtSecret, sessionTimeout, unclaimedUsername, "")
	if err != nil {
		return nil, err
	}
	m.store = store
	acct, err := store.Account()
	switch {
	case err == nil:
		m.apply(acct)
	case !errors.Is(err, credstore.ErrUnclaimed):
		return nil, fmt.Errorf("load the administrator account: %w", err)
	}
	return m, nil
}

// commit applies change to the stored account and then refreshes the working
// copy from the committed result. commitMu keeps two changes from refreshing
// the copy out of order.
func (m *Manager) commit(change func(*credstore.Account) error) error {
	if m.store == nil {
		return ErrNoCredentialStore
	}
	m.commitMu.Lock()
	defer m.commitMu.Unlock()
	if err := m.store.Update(change); err != nil {
		return fmt.Errorf("store the administrator account: %w", err)
	}
	acct, err := m.store.Account()
	if err != nil {
		return fmt.Errorf("reload the administrator account: %w", err)
	}
	m.apply(acct)
	return nil
}

// apply replaces the working copy with a committed account. A TOTP secret that
// is staged but not yet verified is not part of the account and survives, so a
// passkey sign-in during enrolment does not discard it; an enabled secret the
// store no longer holds is dropped.
func (m *Manager) apply(acct credstore.Account) {
	creds := make([]webauthn.Credential, len(acct.Passkeys))
	for i, p := range acct.Passkeys {
		creds[i] = p.Credential
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.username = acct.Username
	m.passwordHash = []byte(acct.PasswordHash)
	m.passkeyHandle = bytes.Clone(acct.UserHandle)
	m.passkeys = creds
	switch {
	case acct.TOTPSecret != "":
		m.totp = totpState{secret: acct.TOTPSecret, enabled: true}
	case m.totp.enabled:
		m.totp = totpState{secret: "", enabled: false}
	}
}

// ReplaceCredential stores a new password hash for the administrator, which is
// how setup claims the daemon and how recovery takes it back. Recovery passes
// recovery=true, and the TOTP secret and every passkey go in the same write:
// whoever holds the recovery token has the filesystem, and the second factors
// were enrolled by someone who may no longer hold them. A pending enrolment is
// dropped too. The failed-login count is cleared.
func (m *Manager) ReplaceCredential(passwordHash string, recovery bool) error {
	username := m.GetUsername()
	err := m.commit(func(a *credstore.Account) error {
		a.Username, a.PasswordHash = username, passwordHash
		if recovery {
			a.TOTPSecret, a.Passkeys = "", nil
		}
		return nil
	})
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if recovery {
		m.totp = totpState{secret: "", enabled: false}
	}
	m.lock = lockout{failures: 0, until: time.Time{}, inflight: m.lock.inflight}
	return nil
}

// recordPasskey adds a credential or replaces the one with the same ID: a
// sign-in returns the record with its updated counter and flags.
func (m *Manager) recordPasskey(credential webauthn.Credential) error {
	now := m.now().UTC()
	return m.commit(func(a *credstore.Account) error {
		i := slices.IndexFunc(a.Passkeys, func(p credstore.Passkey) bool {
			return bytes.Equal(p.Credential.ID, credential.ID)
		})
		if i >= 0 {
			a.Passkeys[i].Credential = credential
			return nil
		}
		a.Passkeys = append(a.Passkeys, credstore.Passkey{
			Name:       fmt.Sprintf("Passkey %d", len(a.Passkeys)+1),
			CreatedAt:  now,
			Credential: credential,
		})
		return nil
	})
}
