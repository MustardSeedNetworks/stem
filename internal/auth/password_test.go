// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"context"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

// bcryptCost10Fixture is a deterministic fixture used to assert that
// VerifyPassword flags legacy bcrypt hashes for re-hash. Generated with
// bcrypt.DefaultCost (10) so callers can assert the rehash-on-login
// migration path without running bcrypt at production cost.
//
// Plaintext for fixture: "legacyPassw0rd!123".
const (
	legacyBcryptPassword = "legacyPassw0rd!123"
	// Generated once via bcrypt.GenerateFromPassword at cost 10.
	legacyBcryptFixture = "$2a$10$.PLBMCDzHzrJWtvjJOq2tOztcj.oZVrB/gclaUFL1Qrgj2n6k4i32"
)

func TestIsDefaultPasswordHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{
			name:     "empty hash",
			hash:     "",
			expected: true,
		},
		{
			name:     "default prefix hash",
			hash:     "$2a$10$defaultSomeRandomStuff",
			expected: true,
		},
		{
			name:     "valid bcrypt hash",
			hash:     "$2a$10$N9qo8uLOickgx2ZMRZoMy.MN4fJR7M/Cq5sRmz8R0KfRqDzzXcC9a",
			expected: false,
		},
		{
			name:     "short hash",
			hash:     "$2a$10",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := auth.IsDefaultPasswordHash(tt.hash)
			if result != tt.expected {
				t.Errorf("IsDefaultPasswordHash(%q) = %v, want %v", tt.hash, result, tt.expected)
			}
		})
	}
}

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

// TestHashPassword_ProducesArgon2id verifies HashPassword now emits a
// PHC-formatted Argon2id hash (task #84, Wave 2 migration off bcrypt).
func TestHashPassword_ProducesArgon2id(t *testing.T) {
	t.Parallel()

	password := "testPassword123!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("HashPassword should produce $argon2id$-prefixed hash, got %q", hash)
	}
	if !strings.Contains(hash, "$v=19$") {
		t.Errorf("HashPassword should embed v=19, got %q", hash)
	}
}

// TestHashPassword_Unique ensures distinct salts are used per hash.
func TestHashPassword_Unique(t *testing.T) {
	t.Parallel()

	password := "samePassword123!"
	hash1, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	hash2, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Different salts should produce different hashes.
	if hash1 == hash2 {
		t.Error("HashPassword should produce different hashes for same password (different salts)")
	}
}

// TestVerifyPassword_Argon2id_Match verifies the happy path for
// Argon2id verification — matched=true, no rehash needed.
func TestVerifyPassword_Argon2id_Match(t *testing.T) {
	t.Parallel()

	password := "argon2VerifyMatch!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	matched, needsRehash, vErr := auth.VerifyPassword(hash, password)
	if vErr != nil {
		t.Fatalf("VerifyPassword returned error: %v", vErr)
	}
	if !matched {
		t.Error("VerifyPassword should match correct password")
	}
	if needsRehash {
		t.Error("Argon2id hash should not trigger rehash")
	}
}

// TestVerifyPassword_Argon2id_NoMatch verifies wrong-password rejection.
func TestVerifyPassword_Argon2id_NoMatch(t *testing.T) {
	t.Parallel()

	hash, err := auth.HashPassword("right-password")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	matched, needsRehash, vErr := auth.VerifyPassword(hash, "wrong-password")
	if vErr != nil {
		t.Fatalf("VerifyPassword returned error: %v", vErr)
	}
	if matched {
		t.Error("VerifyPassword should reject wrong password")
	}
	if needsRehash {
		t.Error("Failed verify should not flag for rehash")
	}
}

// TestVerifyPassword_Bcrypt_FlagsRehash is the centerpiece migration
// test: a legacy bcrypt hash must verify successfully AND set
// needsRehash=true so the caller upgrades the storage to Argon2id on
// the next login.
func TestVerifyPassword_Bcrypt_FlagsRehash(t *testing.T) {
	t.Parallel()

	// Sanity-check the fixture before the actual assertion so a
	// fixture typo doesn't masquerade as a regression.
	if err := bcrypt.CompareHashAndPassword(
		[]byte(legacyBcryptFixture), []byte(legacyBcryptPassword),
	); err != nil {
		t.Fatalf("fixture bcrypt hash does not verify against fixture password: %v", err)
	}

	matched, needsRehash, vErr := auth.VerifyPassword(legacyBcryptFixture, legacyBcryptPassword)
	if vErr != nil {
		t.Fatalf("VerifyPassword returned error: %v", vErr)
	}
	if !matched {
		t.Error("VerifyPassword should match legacy bcrypt hash")
	}
	if !needsRehash {
		t.Error("VerifyPassword must flag bcrypt hash for rehash (migration path)")
	}
}

// TestVerifyPassword_Bcrypt_WrongPassword verifies that a bcrypt-format
// hash with the wrong password reports no match and no rehash.
func TestVerifyPassword_Bcrypt_WrongPassword(t *testing.T) {
	t.Parallel()

	matched, needsRehash, vErr := auth.VerifyPassword(legacyBcryptFixture, "nope-not-it")
	if vErr != nil {
		t.Fatalf("VerifyPassword returned error: %v", vErr)
	}
	if matched {
		t.Error("VerifyPassword should not match wrong password")
	}
	if needsRehash {
		t.Error("Failed bcrypt verify must not request rehash")
	}
}

// TestVerifyPassword_UnsupportedHash verifies the unknown-prefix path.
func TestVerifyPassword_UnsupportedHash(t *testing.T) {
	t.Parallel()

	matched, needsRehash, vErr := auth.VerifyPassword("$pbkdf2$something", "whatever")
	if matched || needsRehash {
		t.Error("Unknown hash format must not match or trigger rehash")
	}
	if vErr == nil {
		t.Error("Unknown hash format must return an error")
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

// TestFirstRun_PlaceholderHashRejectsAllPasswords locks the cross-repo
// first-run invariant (#356): the shipped placeholder password hash is
// NOT a usable default — no password authenticates against it, so a
// fresh install forces the operator through setup. Mirrors the same
// guarantee seed/niac make in their first-run flows.
func TestFirstRun_PlaceholderHashRejectsAllPasswords(t *testing.T) {
	// The placeholder marker baked into a fresh install. Kept as a
	// literal here (the prefix const is unexported) — if this value
	// ever changes, IsDefaultPasswordHash must change in lockstep, and
	// this test will fail loudly if the two drift.
	const placeholder = "$2a$10$default"

	if !auth.IsDefaultPasswordHash(placeholder) {
		t.Fatal("placeholder hash must be flagged as default → setup required")
	}

	for _, guess := range []string{"", "admin", "password", "changeme", "default", "stem"} {
		ok, valid, _ := auth.VerifyPassword(placeholder, guess)
		if ok || valid {
			t.Errorf("placeholder hash authenticated password %q (ok=%v valid=%v) — usable default leaked",
				guess, ok, valid)
		}
	}
}
