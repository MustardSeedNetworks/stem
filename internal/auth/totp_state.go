// SPDX-License-Identifier: BUSL-1.1

package auth

// TOTP state lives on Manager alongside the password hash. Stem ships
// a single-user auth surface (env-driven STEM_AUTH_USERNAME /
// STEM_AUTH_PASSWORD); the second-factor secret is held in memory and
// is the operator's responsibility to back up out-of-band — same model
// as the JWT secret and the rest of the bootstrap state.
//
// JUDGMENT CALL: storage. Stem currently has no key-derivation
// infrastructure (the bcrypt → Argon2id work landed in Wave 2 but only
// touches password hashes, not secret encryption). Persisting the TOTP
// secret at rest in plain form would be a regression over the in-
// memory model. Until a Wave-4 secret-management layer arrives, we
// keep the secret in memory only and document that operators who
// restart the binary must re-enroll. The JSON setup endpoint returns
// the base32 secret to the UI so the user can re-add it to their
// authenticator if needed — same UX as Google's recovery codes.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TOTP/MFA errors exposed to the HTTP layer.
var (
	// ErrTOTPNotEnabled indicates a verify-time check was attempted on
	// an account that has not enrolled TOTP.
	ErrTOTPNotEnabled = errors.New("TOTP is not enabled for this account")
	// ErrTOTPAlreadyEnabled indicates an enroll attempt for an account
	// that already has TOTP enabled.
	ErrTOTPAlreadyEnabled = errors.New("TOTP is already enabled for this account")
	// ErrMFATokenInvalid indicates the `mfa_token` presented to the
	// TOTP login finisher was not a valid pending-MFA JWT.
	ErrMFATokenInvalid = errors.New("invalid MFA token")
	// ErrMFATokenExpired indicates the pending-MFA JWT exceeded its
	// short TTL.
	ErrMFATokenExpired = errors.New("MFA token expired")
)

// MFAPendingTokenDuration is the lifetime of the short-lived JWT
// issued when a password check succeeds but a second factor is still
// required. Five minutes is enough for a user to fish their phone out
// of their pocket and short enough that a stolen token is useless in
// practice.
const MFAPendingTokenDuration = 5 * time.Minute

// mfaPendingClaimKey is the custom claim name on the pending-MFA JWT.
const mfaPendingClaimKey = "mfa_pending"

// totpState holds the per-account TOTP configuration. Today there is
// exactly one account (the env-driven admin); the field lives on
// Manager directly rather than in a map for that reason.
type totpState struct {
	secret  string
	enabled bool
}

// SetTOTPSecret records a freshly generated TOTP secret without
// enabling MFA. Used by /api/v1/auth/totp/setup — the secret is
// staged so VerifyAndEnableTOTP can promote it to enabled once the
// user proves they have configured their authenticator. The setup
// flow is atomic: a second SetTOTPSecret call before
// VerifyAndEnableTOTP overwrites the prior staged secret.
func (m *Manager) SetTOTPSecret(secret string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totp.secret = secret
	// SetTOTPSecret only ever stages; enabling happens via the
	// dedicated VerifyAndEnableTOTP path so the caller cannot
	// accidentally enable MFA with an unverified secret.
	m.totp.enabled = false
}

// GetTOTPSecret returns the currently staged-or-enabled TOTP secret.
// Returns ("", false) if no secret has been set. The boolean is
// strictly the existence check; enabled-state is queried separately
// via TOTPEnabled to keep the contract orthogonal.
func (m *Manager) GetTOTPSecret() (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.totp.secret == "" {
		return "", false
	}
	return m.totp.secret, true
}

// TOTPEnabled reports whether a verified TOTP secret is active for
// the configured account. The login flow consults this AFTER a
// successful password check to decide whether to issue a final JWT
// (false) or a pending-MFA JWT (true).
func (m *Manager) TOTPEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totp.enabled
}

// VerifyAndEnableTOTP completes TOTP enrolment: it validates `code`
// against the currently staged secret and, on success, marks TOTP as
// enabled. Returns ErrTOTPInvalidCode on a wrong code,
// ErrTOTPEmptySecret if SetTOTPSecret was never called.
//
// This is the only path that flips `enabled` from false → true.
func (m *Manager) VerifyAndEnableTOTP(code string) error {
	m.mu.Lock()
	secret := m.totp.secret
	m.mu.Unlock()

	if secret == "" {
		return ErrTOTPEmptySecret
	}

	ok, err := VerifyTOTP(secret, code)
	if err != nil {
		return err
	}
	if !ok {
		return ErrTOTPInvalidCode
	}

	m.mu.Lock()
	m.totp.enabled = true
	m.mu.Unlock()
	return nil
}

// VerifyTOTPCode checks a code against the active TOTP secret without
// changing the enabled flag. Used by the login finisher and by the
// disable endpoint (which re-verifies the code as part of the
// step-up requirement). Returns ErrTOTPNotEnabled if no secret is set
// or MFA is not enabled.
func (m *Manager) VerifyTOTPCode(code string) error {
	m.mu.RLock()
	secret := m.totp.secret
	enabled := m.totp.enabled
	m.mu.RUnlock()

	if !enabled || secret == "" {
		return ErrTOTPNotEnabled
	}
	ok, err := VerifyTOTP(secret, code)
	if err != nil {
		return err
	}
	if !ok {
		return ErrTOTPInvalidCode
	}
	return nil
}

// DisableTOTP wipes the stored secret and clears the enabled flag.
// Step-up checks (password + current code) happen in the HTTP layer
// before this is called; the method itself does not re-authenticate
// because there is no reasonable failure mode at this point — it is
// the final mutation in the disable flow.
func (m *Manager) DisableTOTP() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totp.secret = ""
	m.totp.enabled = false
}

// GenerateMFAPendingToken issues a short-lived JWT signalling that a
// password check succeeded but a second factor is still required. The
// token's `mfa_pending` custom claim is set to true and its TTL is
// [MFAPendingTokenDuration]. The token has a distinct TokenType so it
// cannot be accepted by ValidateToken-paths that expect a full access
// token.
func (m *Manager) GenerateMFAPendingToken(username string) (string, error) {
	m.mu.RLock()
	secret := m.jwtSecret
	issuer := m.issuer
	m.mu.RUnlock()

	now := time.Now()
	tokenID, err := m.generateTokenID()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"sub":              username,
		"username":         username,
		"token_type":       "mfa_pending",
		"iss":              issuer,
		"iat":              now.Unix(),
		"nbf":              now.Unix(),
		"exp":              now.Add(MFAPendingTokenDuration).Unix(),
		"jti":              tokenID,
		mfaPendingClaimKey: true,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign mfa pending token: %w", err)
	}
	return signed, nil
}

// ValidateMFAPendingToken parses an mfa_token returned by the login
// endpoint and returns the username it was minted for. Returns
// ErrMFATokenExpired / ErrMFATokenInvalid for the expected failure
// modes.
func (m *Manager) ValidateMFAPendingToken(_ context.Context, tokenString string) (string, error) {
	m.mu.RLock()
	secret := m.jwtSecret
	m.mu.RUnlock()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrMFATokenInvalid
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrMFATokenExpired
		}
		return "", ErrMFATokenInvalid
	}
	if !token.Valid {
		return "", ErrMFATokenInvalid
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrMFATokenInvalid
	}
	pending, _ := claims[mfaPendingClaimKey].(bool)
	if !pending {
		return "", ErrMFATokenInvalid
	}
	username, _ := claims["username"].(string)
	if username == "" {
		return "", ErrMFATokenInvalid
	}
	return username, nil
}

// IssueAccessToken is the public access-token issuer used by the
// TOTP login finisher after a successful second-factor verification.
// It mirrors the Authenticate happy path (generateToken) without
// repeating the password check.
func (m *Manager) IssueAccessToken(username string) (string, error) {
	return m.generateToken(username)
}
