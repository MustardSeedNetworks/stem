// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/MustardSeedNetworks/foundation/pkg/password"
)

// Password validation constants.
const (
	// MinPasswordLength is the minimum required password length.
	MinPasswordLength = 12
	// GeneratedPasswordLength is the length for auto-generated passwords.
	GeneratedPasswordLength = 24
)

// Password validation errors.
var (
	// ErrPasswordTooShort indicates the password doesn't meet minimum length.
	ErrPasswordTooShort = errors.New("password must be at least 12 characters")
	// ErrPasswordTooLong indicates the password exceeds what
	// [password.Hash] accepts.
	ErrPasswordTooLong = errors.New("password must be at most 1024 bytes")
)

// GenerateSecurePassword creates a cryptographically random password.
// The password is base64-encoded for easy copying and includes special characters.
func GenerateSecurePassword(length int) (string, error) {
	if length <= 0 {
		length = GeneratedPasswordLength
	}

	// Generate random bytes (more than needed due to base64 expansion).
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	// Encode to base64 and trim to requested length.
	generated := base64.URLEncoding.EncodeToString(bytes)
	if len(generated) > length {
		generated = generated[:length]
	}

	return generated, nil
}

// ValidatePasswordStrength checks if a password meets security requirements.
// This is the cheap length-only check; the zxcvbn-backed strength meter
// lives in [EvaluatePasswordStrength].
func ValidatePasswordStrength(plaintext string) error {
	if len(plaintext) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(plaintext) > password.MaxBytes {
		return ErrPasswordTooLong
	}
	return nil
}

// UpdatePasswordHash updates the stored password hash in the auth manager.
// This is used during initial setup or password changes.
func (m *Manager) UpdatePasswordHash(_ context.Context, newHash string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.passwordHash = []byte(newHash)
}

// GetPasswordHash returns the current password hash (for config saving).
func (m *Manager) GetPasswordHash() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return string(m.passwordHash)
}

// GetUsername returns the configured username.
func (m *Manager) GetUsername() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.username
}
