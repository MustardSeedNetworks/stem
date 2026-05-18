// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"sync"
	"time"
)

const cleanupInterval = 5 * time.Minute

// TokenBlacklist stores revoked tokens until they expire.
// Uses [sync.Map] for concurrent access without locking overhead.
type TokenBlacklist struct {
	// tokens maps token ID (jti) to expiration time.
	tokens sync.Map
	// stopCh signals the cleanup goroutine to stop.
	stopCh chan struct{}
	// stopOnce ensures Stop is called only once.
	stopOnce sync.Once
}

// NewTokenBlacklist creates a new token blacklist.
func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens: sync.Map{},
		stopCh: make(chan struct{}),
	}
	// Start cleanup goroutine to remove expired tokens.
	go bl.cleanupLoop()
	return bl
}

// Stop stops the cleanup goroutine. Safe to call multiple times.
func (bl *TokenBlacklist) Stop() {
	bl.stopOnce.Do(func() {
		close(bl.stopCh)
	})
}

// Add adds a token to the blacklist.
// The token remains blacklisted until its expiration time.
func (bl *TokenBlacklist) Add(tokenID string, expiresAt time.Time) {
	if tokenID == "" {
		return
	}
	bl.tokens.Store(tokenID, expiresAt)
}

// IsBlacklisted checks if a token ID is in the blacklist.
func (bl *TokenBlacklist) IsBlacklisted(tokenID string) bool {
	if tokenID == "" {
		return false
	}
	_, exists := bl.tokens.Load(tokenID)
	return exists
}

// Remove removes a token from the blacklist (for testing).
func (bl *TokenBlacklist) Remove(tokenID string) {
	bl.tokens.Delete(tokenID)
}

// Cleanup removes all expired tokens from the blacklist.
// This is exposed for testing purposes.
func (bl *TokenBlacklist) Cleanup() {
	bl.cleanup()
}

// cleanupLoop periodically removes expired tokens from the blacklist.
func (bl *TokenBlacklist) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bl.stopCh:
			return
		case <-ticker.C:
			bl.cleanup()
		}
	}
}

// cleanup removes all expired tokens.
func (bl *TokenBlacklist) cleanup() {
	now := time.Now()
	bl.tokens.Range(func(key, value any) bool {
		if expiresAt, ok := value.(time.Time); ok {
			if now.After(expiresAt) {
				bl.tokens.Delete(key)
			}
		}
		return true
	})
}
