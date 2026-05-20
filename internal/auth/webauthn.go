// SPDX-License-Identifier: BUSL-1.1

package auth

// WebAuthn (passkeys) integration. Wraps go-webauthn so the HTTP layer
// only deals with the two ceremonies (register, login) and the auth
// manager stores credentials in memory alongside the TOTP state.
//
// The library's webauthn.User interface is satisfied by [webauthnUser]
// — a small adapter that pulls the username and credential list off
// the Manager. The Manager itself does not implement the interface
// because the library wants per-user instances and stem only ever has
// one account; the adapter pattern keeps the door open for a future
// multi-user store.
//
// JUDGMENT CALL: configuration. RPID defaults to "localhost" so the
// dev-mode loopback build works out of the box; production deployments
// MUST set STEM_WEBAUTHN_RPID and STEM_WEBAUTHN_ORIGINS to the served
// hostname and the full origin URL respectively. The library refuses
// to mint credentials when these mismatch the request Origin, so a
// misconfiguration fails closed rather than silently disabling the
// second factor.

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// WebAuthn configuration env vars + defaults.
const (
	// WebAuthnRPIDEnv is the relying-party ID. Browsers require this
	// to be a registrable suffix of the served hostname; we default
	// to "localhost" so the dev build works without configuration.
	WebAuthnRPIDEnv = "STEM_WEBAUTHN_RPID"
	// WebAuthnOriginsEnv is the comma-separated list of allowed
	// origins (scheme + host + port). The library rejects assertions
	// whose Origin header is not on the list.
	WebAuthnOriginsEnv = "STEM_WEBAUTHN_ORIGINS"
	// WebAuthnRPNameEnv overrides the relying-party display name shown
	// to the user during the browser ceremony. Defaults to "Stem".
	WebAuthnRPNameEnv = "STEM_WEBAUTHN_RPNAME"

	defaultWebAuthnRPID    = "localhost"
	defaultWebAuthnRPName  = "Stem"
	defaultWebAuthnOrigins = "http://localhost:8080,https://localhost:8443"

	// webauthnSessionTTL is how long a pending registration / login
	// ceremony remains valid between begin and finish. The browser
	// completes the ceremony in seconds; five minutes is a generous
	// upper bound that still bounds an attacker's replay window.
	webauthnSessionTTL = 5 * time.Minute
)

// WebAuthn errors.
var (
	// ErrWebAuthnNotConfigured indicates the relying-party config could
	// not be parsed (likely a missing/garbled origin list).
	ErrWebAuthnNotConfigured = errors.New("WebAuthn not configured")
	// ErrWebAuthnNoSession indicates a finish-ceremony call arrived
	// with no matching begin-ceremony state.
	ErrWebAuthnNoSession = errors.New("no pending WebAuthn ceremony")
	// ErrWebAuthnSessionExpired indicates a begin-ceremony state was
	// found but had exceeded webauthnSessionTTL.
	ErrWebAuthnSessionExpired = errors.New("WebAuthn ceremony expired")
	// ErrWebAuthnNoCredentials indicates a login ceremony was started
	// for a user with no registered credentials.
	ErrWebAuthnNoCredentials = errors.New("no WebAuthn credentials registered")
)

// WebAuthnStoredCredential is the persisted form of a registered
// passkey. The library's webauthn.Credential struct contains pointers
// to byte slices that we want to copy out before storing.
type WebAuthnStoredCredential struct {
	// ID is the credential identifier returned by the authenticator.
	ID []byte
	// PublicKey is the COSE-encoded public key. Used to verify
	// assertion signatures.
	PublicKey []byte
	// AttestationType is the verification model used during
	// registration (e.g. "none", "packed"). Preserved for audit.
	AttestationType string
	// SignCount is the authenticator's monotonic counter. Re-used as
	// a clone-detection hint per WebAuthn §6.1.
	SignCount uint32
	// CreatedAt is when the credential was registered. Useful for the
	// security page UI ("Added 3 days ago").
	CreatedAt time.Time
}

// WebAuthnSessionState holds the challenge + library session data
// between a begin and finish call. Per-user keyed; concurrent begins
// from the same user overwrite each other.
type WebAuthnSessionState struct {
	// SessionData is the opaque library state; pass back into Finish*.
	SessionData webauthn.SessionData
	// CreatedAt is used to expire the session after webauthnSessionTTL.
	CreatedAt time.Time
}

// webauthnUser adapts a username + stored-credential slice to the
// library's webauthn.User interface. Created on demand by the Manager
// for each ceremony.
type webauthnUser struct {
	username    string
	credentials []webauthn.Credential
}

// WebAuthnID returns the user's identifier. WebAuthn allows up to 64
// bytes; we pass the lower-cased username, which is what the existing
// auth manager already uses as the canonical identifier.
func (u *webauthnUser) WebAuthnID() []byte {
	return []byte(strings.ToLower(u.username))
}

// WebAuthnName returns the human-readable username.
func (u *webauthnUser) WebAuthnName() string {
	return u.username
}

// WebAuthnDisplayName returns the display name. Same as WebAuthnName
// because the single-user model has no separate display field.
func (u *webauthnUser) WebAuthnDisplayName() string {
	return u.username
}

// WebAuthnCredentials returns the list of currently registered
// credentials. The library uses this when generating allowCredentials
// for the login ceremony.
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// NewWebAuthn constructs a configured webauthn.WebAuthn instance. The
// configuration comes from environment variables — see the
// `WebAuthn*Env` constants. Returns ErrWebAuthnNotConfigured if the
// origin list is empty or malformed.
func NewWebAuthn() (*webauthn.WebAuthn, error) {
	rpid := os.Getenv(WebAuthnRPIDEnv)
	if rpid == "" {
		rpid = defaultWebAuthnRPID
	}
	rpName := os.Getenv(WebAuthnRPNameEnv)
	if rpName == "" {
		rpName = defaultWebAuthnRPName
	}
	originsRaw := os.Getenv(WebAuthnOriginsEnv)
	if originsRaw == "" {
		originsRaw = defaultWebAuthnOrigins
	}

	origins := strings.Split(originsRaw, ",")
	cleaned := make([]string, 0, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o != "" {
			cleaned = append(cleaned, o)
		}
	}
	if len(cleaned) == 0 {
		return nil, ErrWebAuthnNotConfigured
	}

	wa, err := webauthn.New(&webauthn.Config{
		RPID:          rpid,
		RPDisplayName: rpName,
		RPOrigins:     cleaned,
	})
	if err != nil {
		return nil, fmt.Errorf("create webauthn: %w", err)
	}
	return wa, nil
}

// libraryCredentials projects the manager's stored credentials into
// the library's expected slice shape. Holds the read lock for the
// duration to keep the slice consistent.
func (m *Manager) libraryCredentials() []webauthn.Credential {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]webauthn.Credential, 0, len(m.webauthnCredentials))
	for _, c := range m.webauthnCredentials {
		out = append(out, webauthn.Credential{
			ID:              append([]byte(nil), c.ID...),
			PublicKey:       append([]byte(nil), c.PublicKey...),
			AttestationType: c.AttestationType,
			Authenticator: webauthn.Authenticator{
				SignCount: c.SignCount,
			},
		})
	}
	return out
}

// addWebAuthnCredential appends a freshly registered credential to the
// in-memory store. Used by FinishWebAuthnRegistration.
func (m *Manager) addWebAuthnCredential(cred webauthn.Credential) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.webauthnCredentials = append(m.webauthnCredentials, WebAuthnStoredCredential{
		ID:              append([]byte(nil), cred.ID...),
		PublicKey:       append([]byte(nil), cred.PublicKey...),
		AttestationType: cred.AttestationType,
		SignCount:       cred.Authenticator.SignCount,
		CreatedAt:       time.Now().UTC(),
	})
}

// updateWebAuthnSignCount bumps the stored sign-count of the matching
// credential. Called from FinishWebAuthnLogin so the next assertion
// can detect a cloned authenticator (counter that doesn't increase).
func (m *Manager) updateWebAuthnSignCount(credentialID []byte, newCount uint32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.webauthnCredentials {
		if string(m.webauthnCredentials[i].ID) == string(credentialID) {
			m.webauthnCredentials[i].SignCount = newCount
			return
		}
	}
}

// HasWebAuthnCredentials reports whether at least one passkey is
// registered. Used by the login flow to decide whether to offer the
// WebAuthn second factor.
func (m *Manager) HasWebAuthnCredentials() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.webauthnCredentials) > 0
}

// WebAuthnCredentialCount returns the number of registered passkeys.
// Used by the security UI to show how many devices are enrolled.
func (m *Manager) WebAuthnCredentialCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.webauthnCredentials)
}

// userForCeremony returns a webauthnUser snapshot for the configured
// account. Caller passes username so callers that have not yet bound
// to the account (registration begin) can still construct one.
func (m *Manager) userForCeremony(username string) *webauthnUser {
	return &webauthnUser{
		username:    username,
		credentials: m.libraryCredentials(),
	}
}

// BeginWebAuthnRegistration kicks off enrolment of a new passkey for
// `username`. Returns the credential creation options the browser
// should pass to navigator.credentials.create. The challenge + session
// data are stored in-memory for the matching finish call.
func (m *Manager) BeginWebAuthnRegistration(
	wa *webauthn.WebAuthn, username string,
) (*protocol.CredentialCreation, error) {
	user := m.userForCeremony(username)
	options, sessionData, err := wa.BeginRegistration(user)
	if err != nil {
		return nil, fmt.Errorf("begin webauthn registration: %w", err)
	}

	m.mu.Lock()
	m.webauthnSessions[m.sessionKey(username, "register")] = &WebAuthnSessionState{
		SessionData: *sessionData,
		CreatedAt:   time.Now().UTC(),
	}
	m.mu.Unlock()
	return options, nil
}

// FinishWebAuthnRegistration validates the browser-supplied attestation
// against the matching begin-ceremony state and, on success, persists
// the credential in the manager's in-memory store. Returns the stored
// credential snapshot for the caller's audit log.
func (m *Manager) FinishWebAuthnRegistration(
	wa *webauthn.WebAuthn, username string, response *protocol.ParsedCredentialCreationData,
) (*WebAuthnStoredCredential, error) {
	state, err := m.consumeWebAuthnSession(username, "register")
	if err != nil {
		return nil, err
	}
	user := m.userForCeremony(username)
	cred, err := wa.CreateCredential(user, state.SessionData, response)
	if err != nil {
		return nil, fmt.Errorf("create webauthn credential: %w", err)
	}
	m.addWebAuthnCredential(*cred)
	stored := WebAuthnStoredCredential{
		ID:              append([]byte(nil), cred.ID...),
		PublicKey:       append([]byte(nil), cred.PublicKey...),
		AttestationType: cred.AttestationType,
		SignCount:       cred.Authenticator.SignCount,
		CreatedAt:       time.Now().UTC(),
	}
	return &stored, nil
}

// BeginWebAuthnLogin issues an assertion challenge for an existing
// account. Returns ErrWebAuthnNoCredentials if no passkeys are
// registered (the UI should hide the passkey button in that case, but
// we re-check here to fail loudly on race conditions).
func (m *Manager) BeginWebAuthnLogin(
	wa *webauthn.WebAuthn, username string,
) (*protocol.CredentialAssertion, error) {
	if !m.HasWebAuthnCredentials() {
		return nil, ErrWebAuthnNoCredentials
	}
	user := m.userForCeremony(username)
	options, sessionData, err := wa.BeginLogin(user)
	if err != nil {
		return nil, fmt.Errorf("begin webauthn login: %w", err)
	}
	m.mu.Lock()
	m.webauthnSessions[m.sessionKey(username, "login")] = &WebAuthnSessionState{
		SessionData: *sessionData,
		CreatedAt:   time.Now().UTC(),
	}
	m.mu.Unlock()
	return options, nil
}

// FinishWebAuthnLogin validates an assertion against the matching
// begin-ceremony state. On success, the credential's sign-count is
// updated to enable clone detection on subsequent logins.
func (m *Manager) FinishWebAuthnLogin(
	wa *webauthn.WebAuthn, username string, response *protocol.ParsedCredentialAssertionData,
) error {
	state, err := m.consumeWebAuthnSession(username, "login")
	if err != nil {
		return err
	}
	user := m.userForCeremony(username)
	cred, err := wa.ValidateLogin(user, state.SessionData, response)
	if err != nil {
		return fmt.Errorf("validate webauthn login: %w", err)
	}
	m.updateWebAuthnSignCount(cred.ID, cred.Authenticator.SignCount)
	return nil
}

// consumeWebAuthnSession atomically reads-and-deletes the session
// state for the named ceremony. Returns ErrWebAuthnNoSession when no
// state exists; ErrWebAuthnSessionExpired when the TTL has lapsed.
func (m *Manager) consumeWebAuthnSession(username, kind string) (*WebAuthnSessionState, error) {
	key := m.sessionKey(username, kind)
	m.mu.Lock()
	defer m.mu.Unlock()
	state, ok := m.webauthnSessions[key]
	if !ok {
		return nil, ErrWebAuthnNoSession
	}
	delete(m.webauthnSessions, key)
	if time.Since(state.CreatedAt) > webauthnSessionTTL {
		return nil, ErrWebAuthnSessionExpired
	}
	return state, nil
}

// sessionKey builds the map key used by the in-memory ceremony store.
// Lower-casing the username matches the canonicalisation Authenticate
// already performs.
func (m *Manager) sessionKey(username, kind string) string {
	return strings.ToLower(username) + ":" + kind
}

// EncodeWebAuthnID returns the base64url-encoded form of a stored
// credential's ID. Used by the UI to enumerate registered passkeys.
func EncodeWebAuthnID(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}
