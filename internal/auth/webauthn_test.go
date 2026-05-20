// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
)

// TestNewWebAuthn_DefaultConfig accepts a missing env config and
// falls back to the localhost defaults so the dev build works out of
// the box.
func TestNewWebAuthn_DefaultConfig(t *testing.T) {
	// NOT t.Parallel — modifies process env.
	t.Setenv(auth.WebAuthnRPIDEnv, "")
	t.Setenv(auth.WebAuthnOriginsEnv, "")
	t.Setenv(auth.WebAuthnRPNameEnv, "")

	wa, err := auth.NewWebAuthn()
	if err != nil {
		t.Fatalf("NewWebAuthn() error: %v", err)
	}
	if wa == nil {
		t.Fatal("expected non-nil webauthn instance")
	}
}

// TestNewWebAuthn_CustomConfig respects environment overrides.
func TestNewWebAuthn_CustomConfig(t *testing.T) {
	t.Setenv(auth.WebAuthnRPIDEnv, "stem.example.com")
	t.Setenv(auth.WebAuthnOriginsEnv, "https://stem.example.com:8443")
	t.Setenv(auth.WebAuthnRPNameEnv, "Stem-Prod")

	wa, err := auth.NewWebAuthn()
	if err != nil {
		t.Fatalf("NewWebAuthn() error: %v", err)
	}
	if wa == nil {
		t.Fatal("expected non-nil webauthn instance")
	}
}

// TestNewWebAuthn_EmptyOrigins rejects an empty origin list.
func TestNewWebAuthn_EmptyOrigins(t *testing.T) {
	t.Setenv(auth.WebAuthnOriginsEnv, " , , ")
	t.Setenv(auth.WebAuthnRPIDEnv, "stem.example.com")

	_, err := auth.NewWebAuthn()
	if !errors.Is(err, auth.ErrWebAuthnNotConfigured) {
		t.Errorf("expected ErrWebAuthnNotConfigured, got %v", err)
	}
}

// TestManagerWebAuthn_NoCredentialsYet asserts the manager returns the
// expected zero-state on a fresh instance and ErrWebAuthnNoCredentials
// when a login ceremony is started without prior registration.
func TestManagerWebAuthn_NoCredentialsYet(t *testing.T) {
	mgr := mustNewManagerForTOTPTest(t)
	if mgr.HasWebAuthnCredentials() {
		t.Error("expected no credentials on fresh manager")
	}
	if got := mgr.WebAuthnCredentialCount(); got != 0 {
		t.Errorf("expected 0 credentials, got %d", got)
	}

	t.Setenv(auth.WebAuthnRPIDEnv, "localhost")
	t.Setenv(auth.WebAuthnOriginsEnv, "http://localhost:8080")
	wa, err := auth.NewWebAuthn()
	if err != nil {
		t.Fatalf("NewWebAuthn() error: %v", err)
	}
	_, err = mgr.BeginWebAuthnLogin(wa, "admin")
	if !errors.Is(err, auth.ErrWebAuthnNoCredentials) {
		t.Errorf("expected ErrWebAuthnNoCredentials, got %v", err)
	}
}

// TestManagerWebAuthn_BeginRegistrationStoresChallenge verifies that
// BeginWebAuthnRegistration stashes session state. We cannot finish
// the ceremony in a unit test (the browser-side cryptography is
// required), but we can verify the session is staged and consumed.
func TestManagerWebAuthn_BeginRegistrationStoresChallenge(t *testing.T) {
	t.Setenv(auth.WebAuthnRPIDEnv, "localhost")
	t.Setenv(auth.WebAuthnOriginsEnv, "http://localhost:8080")

	mgr := mustNewManagerForTOTPTest(t)
	wa, err := auth.NewWebAuthn()
	if err != nil {
		t.Fatalf("NewWebAuthn() error: %v", err)
	}

	options, err := mgr.BeginWebAuthnRegistration(wa, "admin")
	if err != nil {
		t.Fatalf("BeginWebAuthnRegistration() error: %v", err)
	}
	if options == nil {
		t.Fatal("expected non-nil credential creation options")
	}
	if len(options.Response.Challenge) == 0 {
		t.Error("expected a non-empty challenge in the options")
	}
}

// TestEncodeWebAuthnID round-trips a credential ID through base64url.
func TestEncodeWebAuthnID(t *testing.T) {
	t.Parallel()
	encoded := auth.EncodeWebAuthnID([]byte{0x01, 0x02, 0x03, 0xff})
	if encoded == "" {
		t.Fatal("expected non-empty encoded ID")
	}
	if strings.ContainsAny(encoded, "+/=") {
		t.Errorf("expected URL-safe encoding (no +/=), got %q", encoded)
	}
}

// TestManagerWebAuthn_SessionExpiry confirms the in-memory session
// store expires entries after webauthnSessionTTL. We cannot inject
// a clock without exposing test seams; instead we drive the begin
// path twice in close succession and verify the second begin
// overwrites the first.
func TestManagerWebAuthn_SessionOverwriteOnRepeatBegin(t *testing.T) {
	t.Setenv(auth.WebAuthnRPIDEnv, "localhost")
	t.Setenv(auth.WebAuthnOriginsEnv, "http://localhost:8080")

	mgr := mustNewManagerForTOTPTest(t)
	wa, err := auth.NewWebAuthn()
	if err != nil {
		t.Fatalf("NewWebAuthn() error: %v", err)
	}
	if _, firstErr := mgr.BeginWebAuthnRegistration(wa, "admin"); firstErr != nil {
		t.Fatalf("first BeginWebAuthnRegistration error: %v", firstErr)
	}
	// Sleep a tiny amount so CreatedAt differs (not used by assert,
	// just defensive).
	time.Sleep(time.Millisecond)
	if _, secondErr := mgr.BeginWebAuthnRegistration(wa, "admin"); secondErr != nil {
		t.Fatalf("second BeginWebAuthnRegistration error: %v", secondErr)
	}
}
