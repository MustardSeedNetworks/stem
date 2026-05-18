// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// Setup token configuration.
const (
	// setupTokenLength is the length of setup tokens in bytes before encoding.
	setupTokenLength = 32

	// SetupTokenExpiry is how long a setup token remains valid.
	// Short expiry prevents token reuse after the setup wizard is closed.
	SetupTokenExpiry = 15 * time.Minute
)

// SetupToken represents a one-time setup token with metadata.
type SetupToken struct {
	Token     string    // The actual token string (base64-encoded)
	ExpiresAt time.Time // When the token expires
	Used      bool      // Whether the token has been used
}

// SetupTokenManager manages setup token generation and validation.
// Implements a single-use token pattern for setup completion security.
// Prevents CSRF and unauthenticated password reset during first-time setup.
type SetupTokenManager struct {
	mu    sync.RWMutex
	token *SetupToken
}

// NewSetupTokenManager creates a new setup token manager.
func NewSetupTokenManager() *SetupTokenManager {
	return &SetupTokenManager{
		mu:    sync.RWMutex{},
		token: nil,
	}
}

// GenerateToken creates a new cryptographically secure setup token.
// Each call generates a new token, invalidating any previous token.
// Returns the token string to be sent to the client.
func (m *SetupTokenManager) GenerateToken() (string, error) {
	// Generate cryptographically secure random bytes.
	tokenBytes := make([]byte, setupTokenLength)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate setup token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Store the new token, replacing any existing one.
	m.token = &SetupToken{
		Token:     token,
		ExpiresAt: time.Now().Add(SetupTokenExpiry),
		Used:      false,
	}

	return token, nil
}

// ValidateToken checks if the provided token is valid.
// The token is invalidated after successful validation (single-use).
// Uses constant-time comparison to prevent timing attacks.
func (m *SetupTokenManager) ValidateToken(providedToken string) bool {
	if providedToken == "" {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if token exists.
	if m.token == nil {
		return false
	}

	// Check if already used.
	if m.token.Used {
		return false
	}

	// Check expiry.
	if time.Now().After(m.token.ExpiresAt) {
		m.token = nil // Clean up expired token.
		return false
	}

	// Constant-time comparison to prevent timing attacks.
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(m.token.Token)) != 1 {
		return false
	}

	// Mark as used (single-use).
	m.token.Used = true

	return true
}

// Invalidate removes any existing token.
// Called after successful setup completion to ensure token cannot be reused.
func (m *SetupTokenManager) Invalidate() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.token = nil
}

// HasValidToken returns true if there's a valid unexpired token.
// Used for debugging/logging purposes only.
func (m *SetupTokenManager) HasValidToken() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.token == nil || m.token.Used {
		return false
	}

	return time.Now().Before(m.token.ExpiresAt)
}
