// SPDX-License-Identifier: BUSL-1.1

package auth

// Passkeys for the single account. foundation/pkg/passkey owns the
// ceremonies and the fleet's verification policy; the Manager owns the
// account's opaque user handle and its complete credential records, which
// the credential store keeps across restarts.

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"slices"

	"github.com/go-webauthn/webauthn/webauthn"
)

// userHandleBytes is the size of the account's WebAuthn user handle.
const userHandleBytes = 32

var (
	// ErrWebAuthnNotConfigured indicates no valid relying party was
	// configured, so passkeys are disabled.
	ErrWebAuthnNotConfigured = errors.New("WebAuthn not configured")
	// ErrWebAuthnNoCredentials indicates a login ceremony was started
	// with no passkey registered.
	ErrWebAuthnNoCredentials = errors.New("no WebAuthn credentials registered")
	// ErrPasskeyUnknown indicates an assertion named a credential or user
	// handle this account does not hold.
	ErrPasskeyUnknown = errors.New("unknown passkey")
)

// newPasskeyHandle returns a random user handle for a daemon that is not yet
// claimed. It is opaque — never the username — so the authenticator stores
// nothing that names the account. The store generates the handle that lasts.
func newPasskeyHandle() []byte {
	handle := make([]byte, userHandleBytes)
	_, _ = rand.Read(handle) // crypto/rand.Read never returns an error
	return handle
}

// passkeyAccount is the account as go-webauthn sees it.
type passkeyAccount struct {
	handle      []byte
	name        string
	credentials []webauthn.Credential
}

func (a *passkeyAccount) WebAuthnID() []byte                         { return a.handle }
func (a *passkeyAccount) WebAuthnName() string                       { return a.name }
func (a *passkeyAccount) WebAuthnDisplayName() string                { return a.name }
func (a *passkeyAccount) WebAuthnCredentials() []webauthn.Credential { return a.credentials }

// PasskeyAccount returns a snapshot of the account for a ceremony.
func (m *Manager) PasskeyAccount() webauthn.User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.passkeyAccountLocked()
}

func (m *Manager) passkeyAccountLocked() *passkeyAccount {
	return &passkeyAccount{
		handle:      m.passkeyHandle,
		name:        m.username,
		credentials: slices.Clone(m.passkeys),
	}
}

// ResolvePasskey is the discoverable-login handler: it admits an assertion
// only when both the user handle and the credential ID belong to this account.
func (m *Manager) ResolvePasskey(rawID, userHandle []byte) (webauthn.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if subtle.ConstantTimeCompare(userHandle, m.passkeyHandle) != 1 ||
		!slices.ContainsFunc(m.passkeys, func(c webauthn.Credential) bool { return bytes.Equal(c.ID, rawID) }) {
		return nil, ErrPasskeyUnknown
	}
	return m.passkeyAccountLocked(), nil
}

// SavePasskey stores the complete credential record, replacing the one with
// the same ID: a sign-in returns the record with its updated counter and
// flags. A registration the store refuses is an error, not a passkey that
// works until the next restart.
func (m *Manager) SavePasskey(credential webauthn.Credential) error {
	return m.recordPasskey(credential)
}

// HasWebAuthnCredentials reports whether at least one passkey is
// registered.
func (m *Manager) HasWebAuthnCredentials() bool {
	return m.WebAuthnCredentialCount() > 0
}

// WebAuthnCredentialCount returns the number of registered passkeys.
// Used by the security UI to show how many devices are enrolled.
func (m *Manager) WebAuthnCredentialCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.passkeys)
}
