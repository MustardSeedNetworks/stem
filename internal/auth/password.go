// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Password validation constants.
const (
	// MinPasswordLength is the minimum required password length.
	MinPasswordLength = 12
	// GeneratedPasswordLength is the length for auto-generated passwords.
	GeneratedPasswordLength = 24

	// defaultPasswordHashPrefix identifies the default/placeholder password.
	// If the stored hash starts with this, setup is required.
	defaultPasswordHashPrefix = "$2a$10$default"
)

// Password errors.
var (
	// ErrPasswordTooShort indicates the password doesn't meet minimum length.
	ErrPasswordTooShort = errors.New("password must be at least 12 characters")
)

// IsDefaultPasswordHash checks if the given hash is the default placeholder.
// Returns true if initial setup is required.
func IsDefaultPasswordHash(hash string) bool {
	// If empty or matches the default prefix, setup is needed.
	if hash == "" {
		return true
	}
	// Check for the special default marker.
	if len(hash) >= len(defaultPasswordHashPrefix) {
		return hash[:len(defaultPasswordHashPrefix)] == defaultPasswordHashPrefix
	}
	return false
}

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
	password := base64.URLEncoding.EncodeToString(bytes)
	if len(password) > length {
		password = password[:length]
	}

	return password, nil
}

// ValidatePasswordStrength checks if a password meets security requirements.
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	return nil
}

// HashPassword creates a bcrypt hash of the given password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
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
