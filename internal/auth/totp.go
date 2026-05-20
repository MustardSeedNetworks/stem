// SPDX-License-Identifier: BUSL-1.1

package auth

// TOTP (RFC 6238) enrollment and verification helpers. Wave 3 (#85)
// introduces a second factor for the existing single-user auth surface:
// after a successful password check, if TOTP is enrolled, the login flow
// issues a short-lived `mfa_pending` JWT and requires a second POST to
// /api/v1/auth/login/totp with a valid 6-digit code.
//
// The package is intentionally narrow — only the operations the HTTP
// layer needs (generate, verify, provisioning URI, QR PNG). Storage of
// the secret + the `enabled` flag is the auth manager's responsibility
// (see totp_state.go).

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTP issuer + key parameters. The issuer is what authenticator apps
// display alongside the account name; "Stem" mirrors the brand used in
// the rest of the UI.
const (
	// DefaultTOTPIssuer is the issuer displayed by authenticator apps.
	DefaultTOTPIssuer = "Stem"

	// totpDigits is the number of digits in a TOTP code. RFC 6238 §5.3
	// recommends six; we pin to that for compatibility with every
	// mainstream authenticator.
	totpDigits = otp.DigitsSix

	// totpPeriodSeconds is the time-step window in seconds. RFC 6238
	// §5.2 mandates 30s as the default.
	totpPeriodSeconds = 30

	// totpAlgorithm is the HMAC primitive. SHA1 is the RFC-mandated
	// default and the only algorithm Google Authenticator and most
	// hardware tokens support.
	totpAlgorithm = otp.AlgorithmSHA1

	// totpQRSize is the side length in pixels of the rendered QR PNG.
	// 256 is the smallest size that scans reliably on retina displays.
	totpQRSize = 256

	// totpSkewPeriods is the ± period tolerance accepted by VerifyTOTP.
	// Allowing ±1 period (i.e. ±30s) absorbs clock drift between the
	// authenticator device and the server while still keeping the
	// effective code lifetime under two minutes.
	totpSkewPeriods = 1

	// totpCodeLength is the expected length of a TOTP code string
	// (six digits). VerifyTOTP rejects anything else up front to keep
	// the constant-time path purely numeric.
	totpCodeLength = 6
)

// TOTP-related errors.
var (
	// ErrTOTPInvalidCode indicates the supplied code did not match.
	ErrTOTPInvalidCode = errors.New("invalid TOTP code")
	// ErrTOTPEmptyCode indicates the caller supplied an empty code.
	ErrTOTPEmptyCode = errors.New("TOTP code is required")
	// ErrTOTPEmptySecret indicates the caller supplied an empty secret.
	ErrTOTPEmptySecret = errors.New("TOTP secret is required")
	// ErrTOTPInvalidAccount indicates the account name was empty.
	ErrTOTPInvalidAccount = errors.New("TOTP account name is required")
)

// TOTPSetup bundles the artefacts a UI needs to enrol a new TOTP
// secret: the base32 secret (for manual entry), the otpauth:// URI
// (the canonical provisioning format), and a PNG QR code rendering of
// the URI.
type TOTPSetup struct {
	// Secret is the base32-encoded TOTP shared secret.
	Secret string
	// ProvisioningURI is the otpauth://totp/... URL embedded in the QR
	// code. UI can render this manually if the QR cannot be displayed.
	ProvisioningURI string
	// QRCodePNG is a PNG-encoded QR rendering of ProvisioningURI.
	// Callers MUST base64-encode this for transport in JSON.
	QRCodePNG []byte
}

// GenerateTOTPSecret mints a fresh TOTP secret for accountName and
// returns the bundle required to display a setup screen. The secret is
// random, base32-encoded, and ready to be persisted via
// Manager.SetTOTPSecret once the caller has verified the user can
// produce a valid code (see Manager.EnableTOTP).
//
// The issuer parameter is what authenticator apps show next to the
// account; pass [DefaultTOTPIssuer] to use the canonical "Stem" label.
func GenerateTOTPSecret(accountName, issuer string) (*TOTPSetup, error) {
	if strings.TrimSpace(accountName) == "" {
		return nil, ErrTOTPInvalidAccount
	}
	if strings.TrimSpace(issuer) == "" {
		issuer = DefaultTOTPIssuer
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
		Period:      totpPeriodSeconds,
		Digits:      totpDigits,
		Algorithm:   totpAlgorithm,
	})
	if err != nil {
		return nil, fmt.Errorf("generate totp secret: %w", err)
	}

	img, err := key.Image(totpQRSize, totpQRSize)
	if err != nil {
		return nil, fmt.Errorf("render totp qr: %w", err)
	}

	var buf bytes.Buffer
	if encErr := png.Encode(&buf, img); encErr != nil {
		return nil, fmt.Errorf("encode totp qr png: %w", encErr)
	}

	return &TOTPSetup{
		Secret:          key.Secret(),
		ProvisioningURI: key.URL(),
		QRCodePNG:       buf.Bytes(),
	}, nil
}

// VerifyTOTP checks `code` against `secret` with ±1 period skew. The
// constant-time semantics live in the underlying library (it hashes
// candidate codes and compares the resulting HMAC slices with
// crypto/subtle), so callers do not need an extra timing guard.
//
// The early length-prefix rejection here is intentionally NOT
// constant-time — a malformed code (length != 6 or non-digits) carries
// no secret-derived material, so timing-leaking the rejection does
// not help an attacker.
//
// Returns (true, nil) on a valid code, (false, ErrTOTPInvalidCode) on a
// wrong code, and a non-nil non-sentinel error only for malformed input
// the caller should treat as a 400 (empty secret, etc.).
func VerifyTOTP(secret, code string) (bool, error) {
	if strings.TrimSpace(secret) == "" {
		return false, ErrTOTPEmptySecret
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return false, ErrTOTPEmptyCode
	}
	// Reject obviously malformed codes early — anything other than six
	// ASCII digits cannot be a TOTP code, so we do not waste a hash on
	// it. The length check is a fast rejection that does not leak
	// timing about the secret; only the underlying HMAC compare needs
	// to be constant-time, and the library handles that.
	if len(code) != totpCodeLength {
		return false, ErrTOTPInvalidCode
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return false, ErrTOTPInvalidCode
		}
	}

	ok, err := totp.ValidateCustom(code, secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    totpPeriodSeconds,
		Skew:      totpSkewPeriods,
		Digits:    totpDigits,
		Algorithm: totpAlgorithm,
	})
	if err != nil {
		// ValidateCustom returns errors only for malformed secrets
		// (bad base32) — surface as ErrTOTPInvalidCode so callers
		// treat it as a normal verification failure.
		return false, fmt.Errorf("%w: %w", ErrTOTPInvalidCode, err)
	}
	if !ok {
		return false, ErrTOTPInvalidCode
	}
	return true, nil
}
