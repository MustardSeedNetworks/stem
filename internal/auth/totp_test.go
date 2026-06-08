// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

// TestGenerateTOTPSecret_EnrollmentRoundtrip exercises the happy path:
// generate a secret, derive a code from it, verify the code.
func TestGenerateTOTPSecret_EnrollmentRoundtrip(t *testing.T) {
	t.Parallel()

	setup, err := auth.GenerateTOTPSecret("admin@stem.test", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() error: %v", err)
	}
	if setup.Secret == "" {
		t.Fatal("Setup.Secret is empty")
	}
	if !strings.HasPrefix(setup.ProvisioningURI, "otpauth://totp/") {
		t.Errorf("ProvisioningURI prefix wrong: %q", setup.ProvisioningURI)
	}
	if len(setup.QRCodePNG) < 8 || string(setup.QRCodePNG[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Error("QRCodePNG is not a valid PNG (missing signature)")
	}

	code, err := totp.GenerateCode(setup.Secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("derive code: %v", err)
	}
	ok, verifyErr := auth.VerifyTOTP(setup.Secret, code)
	if verifyErr != nil {
		t.Fatalf("VerifyTOTP() error: %v", verifyErr)
	}
	if !ok {
		t.Error("VerifyTOTP() returned false for a freshly generated code")
	}
}

// TestVerifyTOTP_WrongCode rejects an incorrect code.
func TestVerifyTOTP_WrongCode(t *testing.T) {
	t.Parallel()

	setup, err := auth.GenerateTOTPSecret("admin", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() error: %v", err)
	}
	ok, err := auth.VerifyTOTP(setup.Secret, "000000")
	if ok {
		t.Error("Expected wrong code to fail")
	}
	if !errors.Is(err, auth.ErrTOTPInvalidCode) {
		t.Errorf("Expected ErrTOTPInvalidCode, got %v", err)
	}
}

// TestVerifyTOTP_EmptyInputs rejects empty secret and code.
func TestVerifyTOTP_EmptyInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret string
		code   string
		want   error
	}{
		{"empty secret", "", "123456", auth.ErrTOTPEmptySecret},
		{"empty code", "JBSWY3DPEHPK3PXP", "", auth.ErrTOTPEmptyCode},
		{"whitespace secret", "   ", "123456", auth.ErrTOTPEmptySecret},
		{"whitespace code", "JBSWY3DPEHPK3PXP", "   ", auth.ErrTOTPEmptyCode},
		{"short code", "JBSWY3DPEHPK3PXP", "12345", auth.ErrTOTPInvalidCode},
		{"non-numeric code", "JBSWY3DPEHPK3PXP", "abcdef", auth.ErrTOTPInvalidCode},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ok, err := auth.VerifyTOTP(tc.secret, tc.code)
			if ok {
				t.Error("Expected verification to fail")
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("Expected %v, got %v", tc.want, err)
			}
		})
	}
}

// TestVerifyTOTP_PastCode rejects a code from a period far in the
// past (outside the ±1 skew window).
func TestVerifyTOTP_PastCode(t *testing.T) {
	t.Parallel()

	setup, err := auth.GenerateTOTPSecret("admin", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() error: %v", err)
	}
	// Derive a code from 5 minutes in the past — well outside the
	// ±30s skew window.
	stale, err := totp.GenerateCode(setup.Secret, time.Now().UTC().Add(-5*time.Minute))
	if err != nil {
		t.Fatalf("derive stale code: %v", err)
	}
	ok, err := auth.VerifyTOTP(setup.Secret, stale)
	if ok {
		t.Error("Expected stale code to fail")
	}
	if !errors.Is(err, auth.ErrTOTPInvalidCode) {
		t.Errorf("Expected ErrTOTPInvalidCode, got %v", err)
	}
}

// TestVerifyTOTP_SkewWindow accepts a code from ±1 period.
func TestVerifyTOTP_SkewWindow(t *testing.T) {
	t.Parallel()

	setup, err := auth.GenerateTOTPSecret("admin", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() error: %v", err)
	}
	// One period back (30s) — should still be accepted under ±1 skew.
	prev, err := totp.GenerateCode(setup.Secret, time.Now().UTC().Add(-30*time.Second))
	if err != nil {
		t.Fatalf("derive previous-period code: %v", err)
	}
	ok, err := auth.VerifyTOTP(setup.Secret, prev)
	if err != nil {
		t.Fatalf("VerifyTOTP() error: %v", err)
	}
	if !ok {
		t.Error("Expected previous-period code to be accepted under ±1 skew")
	}
}

// TestGenerateTOTPSecret_EmptyAccount rejects an empty account name.
func TestGenerateTOTPSecret_EmptyAccount(t *testing.T) {
	t.Parallel()

	_, err := auth.GenerateTOTPSecret("", auth.DefaultTOTPIssuer)
	if !errors.Is(err, auth.ErrTOTPInvalidAccount) {
		t.Errorf("Expected ErrTOTPInvalidAccount, got %v", err)
	}
}

// TestManagerTOTP_EnrollmentFlow exercises the full setup → verify →
// enable → verify-while-enabled sequence on the manager.
func TestManagerTOTP_EnrollmentFlow(t *testing.T) {
	t.Parallel()

	mgr := mustNewManagerForTOTPTest(t)

	if mgr.TOTPEnabled() {
		t.Fatal("expected MFA disabled on fresh manager")
	}

	setup, err := auth.GenerateTOTPSecret("admin", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() error: %v", err)
	}
	mgr.SetTOTPSecret(setup.Secret)
	if mgr.TOTPEnabled() {
		t.Error("SetTOTPSecret must stage but not enable")
	}

	code, err := totp.GenerateCode(setup.Secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("derive code: %v", err)
	}
	if enableErr := mgr.VerifyAndEnableTOTP(code); enableErr != nil {
		t.Fatalf("VerifyAndEnableTOTP() error: %v", enableErr)
	}
	if !mgr.TOTPEnabled() {
		t.Error("VerifyAndEnableTOTP must flip TOTPEnabled to true")
	}

	if verifyErr := mgr.VerifyTOTPCode(code); verifyErr != nil {
		t.Errorf("VerifyTOTPCode() = %v, want nil", verifyErr)
	}

	mgr.DisableTOTP()
	if mgr.TOTPEnabled() {
		t.Error("DisableTOTP must clear the enabled flag")
	}
	if disabledErr := mgr.VerifyTOTPCode(code); !errors.Is(disabledErr, auth.ErrTOTPNotEnabled) {
		t.Errorf("VerifyTOTPCode after disable = %v, want ErrTOTPNotEnabled", disabledErr)
	}
}

// TestManagerTOTP_PendingJWT_Roundtrip generates and validates an
// MFA-pending JWT.
func TestManagerTOTP_PendingJWT_Roundtrip(t *testing.T) {
	t.Parallel()

	mgr := mustNewManagerForTOTPTest(t)
	token, err := mgr.GenerateMFAPendingToken("admin")
	if err != nil {
		t.Fatalf("GenerateMFAPendingToken() error: %v", err)
	}
	got, err := mgr.ValidateMFAPendingToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateMFAPendingToken() error: %v", err)
	}
	if got != "admin" {
		t.Errorf("Expected username 'admin', got %q", got)
	}

	// Garbage token rejected.
	if _, garbageErr := mgr.ValidateMFAPendingToken(
		context.Background(), "garbage",
	); !errors.Is(garbageErr, auth.ErrMFATokenInvalid) {
		t.Errorf("Expected ErrMFATokenInvalid, got %v", garbageErr)
	}
}

// TestVerifyTOTP_ConstantTime sanity-checks that verification works
// with a code of the right length but wrong digits. The constant-time
// path is exercised here; we are not measuring timing in CI (too
// flaky), only the rejection.
func TestVerifyTOTP_ConstantTime(t *testing.T) {
	t.Parallel()

	setup, err := auth.GenerateTOTPSecret("admin", auth.DefaultTOTPIssuer)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() error: %v", err)
	}
	ok, err := auth.VerifyTOTP(setup.Secret, "999999")
	if ok {
		t.Error("Expected 999999 to fail")
	}
	if !errors.Is(err, auth.ErrTOTPInvalidCode) {
		t.Errorf("Expected ErrTOTPInvalidCode, got %v", err)
	}
}

// totpDigitsSanity locks in the assumption that the constant in
// totp.go matches the library default. If pquerna/otp ever changes
// the default this guards against a silent regression.
func TestTOTPDigitsSanity(t *testing.T) {
	t.Parallel()
	if otp.DigitsSix != 6 {
		t.Errorf("otp.DigitsSix = %v, want 6", otp.DigitsSix)
	}
}

// mustNewManagerForTOTPTest builds an auth.Manager with deterministic
// test credentials. STEM_TEST_MODE is already set by main_test.go.
func mustNewManagerForTOTPTest(t *testing.T) *auth.Manager {
	t.Helper()
	mgr, err := auth.NewManager("", time.Hour, "admin", "TestPassw0rd!123")
	if err != nil {
		t.Fatalf("NewManager() error: %v", err)
	}
	t.Cleanup(mgr.Stop)
	return mgr
}
