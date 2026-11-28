// SPDX-License-Identifier: BUSL-1.1

package auth

// TOTP state lives on Manager alongside the password hash. A secret is
// staged in memory while the operator enrols, and reaches the credential
// store only once a code from it has been verified, so an abandoned
// enrolment leaves nothing behind and a restart keeps an enabled factor.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/MustardSeedNetworks/foundation/pkg/credstore"
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

// SetTOTPSecret stages a freshly generated TOTP secret without enabling
// MFA. Used by /api/v1/auth/totp/setup — the secret waits in memory so
// VerifyAndEnableTOTP can promote it once the user proves they have
// configured their authenticator. A second SetTOTPSecret call before
// VerifyAndEnableTOTP overwrites the prior staged secret. Staging over an
// enabled factor is refused with ErrTOTPAlreadyEnabled: it would otherwise
// switch the second factor off for a caller who never presented a code.
func (m *Manager) SetTOTPSecret(secret string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.totp.enabled {
		return ErrTOTPAlreadyEnabled
	}
	m.totp.secret = secret
	return nil
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
// This is the only path that flips `enabled` from false → true, and the
// secret is stored before it does: if the store refuses it, MFA stays off.
func (m *Manager) VerifyAndEnableTOTP(code string) error {
	m.mu.RLock()
	secret := m.totp.secret
	m.mu.RUnlock()

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

	return m.commit(func(a *credstore.Account) error {
		a.TOTPSecret = secret
		return nil
	})
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
// before this is called; the method itself does not re-authenticate.
// The factor stays on if the store refuses the change.
func (m *Manager) DisableTOTP() error {
	return m.commit(func(a *credstore.Account) error {
		a.TOTPSecret = ""
		return nil
	})
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
// modes. A token revoked by [Manager.RevokeMFAPendingToken] is invalid.
func (m *Manager) ValidateMFAPendingToken(_ context.Context, tokenString string) (string, error) {
	username, _, _, err := m.parseMFAPendingToken(tokenString)
	return username, err
}

// RevokeMFAPendingToken spends an mfa_token. The login handler calls it after
// a wrong code, so one password check buys one guess, not a window of guesses.
func (m *Manager) RevokeMFAPendingToken(tokenString string) {
	if _, id, expires, err := m.parseMFAPendingToken(tokenString); err == nil {
		m.blacklist.Add(id, expires)
	}
}

func (m *Manager) parseMFAPendingToken(tokenString string) (string, string, time.Time, error) {
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
			return "", "", time.Time{}, ErrMFATokenExpired
		}
		return "", "", time.Time{}, ErrMFATokenInvalid
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !token.Valid || !ok {
		return "", "", time.Time{}, ErrMFATokenInvalid
	}
	pending, _ := claims[mfaPendingClaimKey].(bool)
	username, _ := claims["username"].(string)
	id, _ := claims["jti"].(string)
	exp, expErr := claims.GetExpirationTime()
	if !pending || username == "" || id == "" || expErr != nil || exp == nil {
		return "", "", time.Time{}, ErrMFATokenInvalid
	}
	if m.blacklist.IsBlacklisted(id) {
		return "", "", time.Time{}, ErrMFATokenInvalid
	}
	return username, id, exp.Time, nil
}

// IssueAccessToken is the public access-token issuer used by the
// TOTP login finisher after a successful second-factor verification.
// It mirrors the Authenticate happy path (generateToken) without
// repeating the password check.
func (m *Manager) IssueAccessToken(username string) (string, error) {
	return m.generateToken(username)
}
