// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/krisarmstrong/stem/internal/auth"
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

func TestHashPassword(t *testing.T) {
	t.Parallel()

	password := "testPassword123!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword returned empty hash")
	}

	// Verify the hash is valid bcrypt.
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("bcrypt.CompareHashAndPassword failed: %v", err)
	}

	// Verify wrong password fails.
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrongPassword"))
	if err == nil {
		t.Error("bcrypt.CompareHashAndPassword should fail for wrong password")
	}
}

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

	newHash := "$2a$10$newHashValue12345678901234567890"
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
