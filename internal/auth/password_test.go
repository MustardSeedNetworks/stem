// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/foundation/pkg/password"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

func TestGenerateSecurePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		length         int
		expectedLength int
	}{
		{
			name:           "default length",
			length:         0,
			expectedLength: auth.GeneratedPasswordLength,
		},
		{
			name:           "custom length",
			length:         16,
			expectedLength: 16,
		},
		{
			name:           "negative length uses default",
			length:         -1,
			expectedLength: auth.GeneratedPasswordLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			password, err := auth.GenerateSecurePassword(tt.length)
			if err != nil {
				t.Fatalf("GenerateSecurePassword(%d) error = %v", tt.length, err)
			}
			if len(password) != tt.expectedLength {
				t.Errorf("GenerateSecurePassword(%d) length = %d, want %d", tt.length, len(password), tt.expectedLength)
			}
		})
	}
}

func TestGenerateSecurePassword_Unique(t *testing.T) {
	t.Parallel()

	passwords := make(map[string]bool)
	for range 100 {
		password, err := auth.GenerateSecurePassword(24)
		if err != nil {
			t.Fatalf("GenerateSecurePassword failed: %v", err)
		}
		if passwords[password] {
			t.Error("GenerateSecurePassword produced duplicate password")
		}
		passwords[password] = true
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "securePassword123!",
			wantErr:  false,
		},
		{
			name:     "exactly minimum length",
			password: "123456789012",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "12345678901",
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
		{
			name:     "at the foundation input bound",
			password: strings.Repeat("a", password.MaxBytes),
			wantErr:  false,
		},
		{
			name:     "over the foundation input bound",
			password: strings.Repeat("a", password.MaxBytes+1),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := auth.ValidatePasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordStrength(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

// TestAuthenticate_FoundationPassword pins login to the foundation
// Argon2id record: the right password authenticates, and a wrong,
// oversize or non-Argon2id credential is refused the same way.
func TestAuthenticate_FoundationPassword(t *testing.T) {
	t.Parallel()

	const plaintext = "Correct-Horse-Battery-9"
	record, err := password.Hash(plaintext)
	if err != nil {
		t.Fatalf("password.Hash: %v", err)
	}
	if !strings.HasPrefix(record, "$argon2id$v=19$m=65536,t=3,p=4$") {
		t.Fatalf("record %q does not carry the foundation parameters", record)
	}

	tests := []struct {
		name    string
		hash    string
		guess   string
		wantErr error
	}{
		{name: "right password", hash: record, guess: plaintext},
		{name: "wrong password", hash: record, guess: "Wrong-Horse-Battery-9", wantErr: auth.ErrInvalidCredentials},
		{
			name:    "oversize password",
			hash:    record,
			guess:   strings.Repeat("a", password.MaxBytes+1),
			wantErr: auth.ErrInvalidCredentials,
		},
		{
			name:    "bcrypt record",
			hash:    "$2a$10$N9qo8uLOickgx2ZMRZoMy.MN4fJR7M/Cq5sRmz8R0KfRqDzzXcC9a",
			guess:   plaintext,
			wantErr: auth.ErrInvalidCredentials,
		},
		{
			name:    "weaker argon2id parameters",
			hash:    strings.Replace(record, "m=65536,t=3,p=4", "m=64,t=1,p=1", 1),
			guess:   plaintext,
			wantErr: auth.ErrInvalidCredentials,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mgr, mgrErr := auth.NewManagerFromHash("test-secret-key-minimum-32chars!", 0, "operator", tt.hash)
			if mgrErr != nil {
				t.Fatalf("NewManagerFromHash: %v", mgrErr)
			}
			t.Cleanup(mgr.Stop)

			_, authErr := mgr.Authenticate("operator", tt.guess)
			if !errors.Is(authErr, tt.wantErr) {
				t.Fatalf("Authenticate() error = %v, want %v", authErr, tt.wantErr)
			}
		})
	}
}

func TestUpdatePasswordHash(t *testing.T) {
	t.Parallel()

	manager, err := auth.NewManager(
		"test-secret-key-minimum-32chars!",
		0,
		"admin",
		"originalPassword123",
	)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	newHash := "$argon2id$v=19$m=8192,t=1,p=1$YWJjZGVmZ2hpamtsbW5vcA$" +
		"aGFzaC1zdHViLWZvci10ZXN0aW5nLW9ubHkAAAAAAA"
	manager.UpdatePasswordHash(context.Background(), newHash)

	if got := manager.GetPasswordHash(); got != newHash {
		t.Errorf("GetPasswordHash() = %q, want %q", got, newHash)
	}
}

func TestGetUsername(t *testing.T) {
	t.Parallel()

	manager, err := auth.NewManager(
		"test-secret-key-minimum-32chars!",
		0,
		"testadmin",
		"testPassword123!",
	)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if got := manager.GetUsername(); got != "testadmin" {
		t.Errorf("GetUsername() = %q, want %q", got, "testadmin")
	}
}

// TestFirstRun_UnclaimedDaemonRejectsAllPasswords locks the cross-repo
// first-run invariant (#356): an unclaimed daemon has no stored hash, and
// no password authenticates against it, so a fresh install forces the
// operator through setup. Mirrors the same guarantee seed/niac make.
func TestFirstRun_UnclaimedDaemonRejectsAllPasswords(t *testing.T) {
	t.Parallel()

	mgr, err := auth.NewManagerFromHash("test-secret-key-minimum-32chars!", 0, "admin", "")
	if err != nil {
		t.Fatalf("NewManagerFromHash: %v", err)
	}
	t.Cleanup(mgr.Stop)

	for _, guess := range []string{"", "admin", "password", "changeme", "default", "stem"} {
		if _, authErr := mgr.Authenticate("admin", guess); !errors.Is(authErr, auth.ErrInvalidCredentials) {
			t.Errorf("unclaimed daemon authenticated password %q (err=%v): usable default leaked", guess, authErr)
		}
	}
}
