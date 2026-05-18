// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

// UserStore interface errors.
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserLocked      = errors.New("user account is locked")
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid password")
)

// User authentication security defaults.
const (
	// DefaultMaxLoginAttempts is the number of failed login attempts before account lockout.
	DefaultMaxLoginAttempts = 5

	// DefaultLockDuration is the account lockout duration.
	DefaultLockDuration = 15 * time.Minute
)

// UserStore defines the interface for user management operations.
// Ported from Seed project to allow pluggable user backends.
//
// This interface allows the auth package to use different storage backends
// (environment variables, SQLite, PostgreSQL, etc.) without changing the
// authentication logic.
type UserStore interface {
	// GetPasswordHash returns the password hash for a user.
	// Returns ErrUserNotFound if the user doesn't exist.
	GetPasswordHash(ctx context.Context, username string) (string, error)

	// GetTokenVersion returns the current token version for a user.
	// Token version is incremented when password changes or tokens are revoked.
	GetTokenVersion(ctx context.Context, username string) (int, error)

	// UpdatePassword updates a user's password hash.
	UpdatePassword(ctx context.Context, username, hash string) error

	// RecordLoginSuccess records a successful login attempt.
	// Resets failed login counter and unlocks the account.
	RecordLoginSuccess(ctx context.Context, username string) error

	// RecordLoginFailure records a failed login attempt.
	// May lock the account if max attempts exceeded.
	RecordLoginFailure(ctx context.Context, username string) error

	// IsLocked checks if a user account is locked due to failed login attempts.
	IsLocked(ctx context.Context, username string) (bool, error)

	// CreateUser creates a new user with the given credentials.
	// Returns ErrUserExists if the username is already taken.
	CreateUser(ctx context.Context, username, passwordHash, role string) error

	// GetUserCount returns the total number of users.
	GetUserCount(ctx context.Context) (int, error)
}

// MemoryUserStore implements UserStore using in-memory storage.
// Suitable for single-user deployments using environment variables.
type MemoryUserStore struct {
	mu               sync.RWMutex
	users            map[string]*memoryUser
	maxLoginAttempts int
	lockDuration     time.Duration
}

// memoryUser represents a user stored in memory.
type memoryUser struct {
	passwordHash   string
	tokenVersion   int
	failedAttempts int
	lockedUntil    time.Time
	role           string
}

// NewMemoryUserStore creates a new in-memory user store.
func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		mu:               sync.RWMutex{},
		users:            make(map[string]*memoryUser),
		maxLoginAttempts: DefaultMaxLoginAttempts,
		lockDuration:     DefaultLockDuration,
	}
}

// NewMemoryUserStoreWithConfig creates a memory store with custom settings.
func NewMemoryUserStoreWithConfig(maxAttempts int, lockDuration time.Duration) *MemoryUserStore {
	store := NewMemoryUserStore()
	store.maxLoginAttempts = maxAttempts
	store.lockDuration = lockDuration
	return store
}

// AddUser adds a user to the memory store (for initialization).
func (s *MemoryUserStore) AddUser(username, passwordHash, role string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[username] = &memoryUser{
		passwordHash:   passwordHash,
		tokenVersion:   1,
		failedAttempts: 0,
		lockedUntil:    time.Time{},
		role:           role,
	}
}

// GetPasswordHash returns the password hash for a user.
func (s *MemoryUserStore) GetPasswordHash(_ context.Context, username string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[username]
	if !ok {
		return "", ErrUserNotFound
	}
	return user.passwordHash, nil
}

// GetTokenVersion returns the current token version for a user.
func (s *MemoryUserStore) GetTokenVersion(_ context.Context, username string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[username]
	if !ok {
		return 0, ErrUserNotFound
	}
	return user.tokenVersion, nil
}

// UpdatePassword updates a user's password hash and increments token version.
func (s *MemoryUserStore) UpdatePassword(_ context.Context, username, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[username]
	if !ok {
		return ErrUserNotFound
	}
	user.passwordHash = hash
	user.tokenVersion++
	return nil
}

// RecordLoginSuccess records a successful login attempt.
func (s *MemoryUserStore) RecordLoginSuccess(_ context.Context, username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[username]
	if !ok {
		return ErrUserNotFound
	}
	user.failedAttempts = 0
	user.lockedUntil = time.Time{}
	return nil
}

// MaxLoginAttempts returns the configured max login attempts (mainly for tests).
func (s *MemoryUserStore) MaxLoginAttempts() int {
	return s.maxLoginAttempts
}

// LockDuration returns the configured lockout duration (mainly for tests).
func (s *MemoryUserStore) LockDuration() time.Duration {
	return s.lockDuration
}

// RecordLoginFailure records a failed login attempt.
func (s *MemoryUserStore) RecordLoginFailure(_ context.Context, username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[username]
	if !ok {
		return ErrUserNotFound
	}

	user.failedAttempts++
	if user.failedAttempts >= s.maxLoginAttempts {
		user.lockedUntil = time.Now().Add(s.lockDuration)
	}
	return nil
}

// IsLocked checks if a user account is locked.
func (s *MemoryUserStore) IsLocked(_ context.Context, username string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[username]
	if !ok {
		return false, ErrUserNotFound
	}

	if user.lockedUntil.IsZero() {
		return false, nil
	}

	// Check if lock has expired
	if time.Now().After(user.lockedUntil) {
		return false, nil
	}

	return true, nil
}

// CreateUser creates a new user.
func (s *MemoryUserStore) CreateUser(_ context.Context, username, passwordHash, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		return ErrUserExists
	}

	s.users[username] = &memoryUser{
		passwordHash:   passwordHash,
		tokenVersion:   1,
		failedAttempts: 0,
		lockedUntil:    time.Time{},
		role:           role,
	}
	return nil
}

// GetUserCount returns the total number of users.
func (s *MemoryUserStore) GetUserCount(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users), nil
}
