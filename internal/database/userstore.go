// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
)

// UserStoreAdapter adapts the database to the auth.UserStore interface.
// It provides persistent user storage with account locking support.
type UserStoreAdapter struct {
	db               *DB
	maxLoginAttempts int
	lockDuration     time.Duration
}

// NewUserStoreAdapter creates a new database-backed UserStore.
func NewUserStoreAdapter(db *DB) *UserStoreAdapter {
	return &UserStoreAdapter{
		db:               db,
		maxLoginAttempts: auth.DefaultMaxLoginAttempts,
		lockDuration:     auth.DefaultLockDuration,
	}
}

// NewUserStoreAdapterWithConfig creates a UserStoreAdapter with custom settings.
func NewUserStoreAdapterWithConfig(db *DB, maxAttempts int, lockDuration time.Duration) *UserStoreAdapter {
	return &UserStoreAdapter{
		db:               db,
		maxLoginAttempts: maxAttempts,
		lockDuration:     lockDuration,
	}
}

// GetPasswordHash returns the password hash for a user.
func (a *UserStoreAdapter) GetPasswordHash(ctx context.Context, username string) (string, error) {
	var hash string
	err := a.db.QueryRow(ctx, `
		SELECT password_hash FROM users WHERE username = ? AND is_active = 1
	`, username).Scan(&hash)

	if errors.Is(err, sql.ErrNoRows) {
		return "", auth.ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("querying password hash: %w", err)
	}

	return hash, nil
}

// GetTokenVersion returns the current token version for a user.
func (a *UserStoreAdapter) GetTokenVersion(ctx context.Context, username string) (int, error) {
	var version int
	err := a.db.QueryRow(ctx, `
		SELECT token_version FROM users WHERE username = ? AND is_active = 1
	`, username).Scan(&version)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, auth.ErrUserNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("querying token version: %w", err)
	}

	return version, nil
}

// UpdatePassword updates a user's password hash and increments token version.
func (a *UserStoreAdapter) UpdatePassword(ctx context.Context, username, hash string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := a.db.Exec(ctx, `
		UPDATE users 
		SET password_hash = ?, token_version = token_version + 1, updated_at = ?
		WHERE username = ? AND is_active = 1
	`, hash, now, username)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// RecordLoginSuccess records a successful login attempt.
func (a *UserStoreAdapter) RecordLoginSuccess(ctx context.Context, username string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := a.db.Exec(ctx, `
		UPDATE users 
		SET failed_attempts = 0, locked_until = NULL, last_login = ?, updated_at = ?
		WHERE username = ? AND is_active = 1
	`, now, now, username)
	if err != nil {
		return fmt.Errorf("recording login success: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// RecordLoginFailure records a failed login attempt.
func (a *UserStoreAdapter) RecordLoginFailure(ctx context.Context, username string) error {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// First get current failed attempts
	var failedAttempts int
	err := a.db.QueryRow(ctx, `
		SELECT failed_attempts FROM users WHERE username = ? AND is_active = 1
	`, username).Scan(&failedAttempts)

	if errors.Is(err, sql.ErrNoRows) {
		return auth.ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("querying failed attempts: %w", err)
	}

	failedAttempts++

	// Lock if max attempts exceeded
	var lockedUntil *string
	if failedAttempts >= a.maxLoginAttempts {
		lockTime := now.Add(a.lockDuration).Format(time.RFC3339)
		lockedUntil = &lockTime
	}

	_, err = a.db.Exec(ctx, `
		UPDATE users 
		SET failed_attempts = ?, locked_until = ?, updated_at = ?
		WHERE username = ? AND is_active = 1
	`, failedAttempts, lockedUntil, nowStr, username)
	if err != nil {
		return fmt.Errorf("recording login failure: %w", err)
	}

	return nil
}

// IsLocked checks if a user account is locked.
func (a *UserStoreAdapter) IsLocked(ctx context.Context, username string) (bool, error) {
	var lockedUntil sql.NullString
	err := a.db.QueryRow(ctx, `
		SELECT locked_until FROM users WHERE username = ? AND is_active = 1
	`, username).Scan(&lockedUntil)

	if errors.Is(err, sql.ErrNoRows) {
		return false, auth.ErrUserNotFound
	}
	if err != nil {
		return false, fmt.Errorf("querying lock status: %w", err)
	}

	if !lockedUntil.Valid || lockedUntil.String == "" {
		return false, nil
	}

	lockTime, err := time.Parse(time.RFC3339, lockedUntil.String)
	if err != nil {
		// Invalid timestamp format - treat as not locked
		return false, nil
	}

	return time.Now().UTC().Before(lockTime), nil
}

// CreateUser creates a new user with the given credentials.
func (a *UserStoreAdapter) CreateUser(ctx context.Context, username, passwordHash, role string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := a.db.Exec(ctx, `
		INSERT INTO users (username, password_hash, role, is_active, token_version, created_at, updated_at)
		VALUES (?, ?, ?, 1, 1, ?, ?)
	`, username, passwordHash, role, now, now)

	if err != nil {
		// Check for unique constraint violation
		if isUniqueConstraintError(err) {
			return auth.ErrUserExists
		}
		return fmt.Errorf("creating user: %w", err)
	}

	return nil
}

// GetUserCount returns the total number of active users.
func (a *UserStoreAdapter) GetUserCount(ctx context.Context) (int, error) {
	var count int
	err := a.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE is_active = 1
	`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}

	return count, nil
}

// GetUser returns a user by username.
func (a *UserStoreAdapter) GetUser(ctx context.Context, username string) (*User, error) {
	var user User
	var lastLogin, lockedUntil sql.NullString

	err := a.db.QueryRow(ctx, `
		SELECT id, username, password_hash, role, is_active, last_login, 
		       failed_attempts, locked_until, token_version, created_at, updated_at
		FROM users WHERE username = ?
	`, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.IsActive,
		&lastLogin, &user.FailedAttempts, &lockedUntil, &user.TokenVersion,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}

	if lastLogin.Valid {
		if t, parseErr := time.Parse(time.RFC3339, lastLogin.String); parseErr == nil {
			user.LastLogin = &t
		}
	}

	if lockedUntil.Valid {
		if t, parseErr := time.Parse(time.RFC3339, lockedUntil.String); parseErr == nil {
			user.LockedUntil = &t
		}
	}

	return &user, nil
}

// ListUsers returns all active users.
func (a *UserStoreAdapter) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := a.db.Query(ctx, `
		SELECT id, username, password_hash, role, is_active, last_login,
		       failed_attempts, locked_until, token_version, created_at, updated_at
		FROM users WHERE is_active = 1
		ORDER BY username
	`)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []User
	for rows.Next() {
		var user User
		var lastLogin, lockedUntil sql.NullString

		if scanErr := rows.Scan(
			&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.IsActive,
			&lastLogin, &user.FailedAttempts, &lockedUntil, &user.TokenVersion,
			&user.CreatedAt, &user.UpdatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("scanning user row: %w", scanErr)
		}

		if lastLogin.Valid {
			if t, parseErr := time.Parse(time.RFC3339, lastLogin.String); parseErr == nil {
				user.LastLogin = &t
			}
		}

		if lockedUntil.Valid {
			if t, parseErr := time.Parse(time.RFC3339, lockedUntil.String); parseErr == nil {
				user.LockedUntil = &t
			}
		}

		users = append(users, user)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterating user rows: %w", rowsErr)
	}

	return users, nil
}

// DeleteUser soft-deletes a user by setting is_active to false.
func (a *UserStoreAdapter) DeleteUser(ctx context.Context, username string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := a.db.Exec(ctx, `
		UPDATE users SET is_active = 0, updated_at = ? WHERE username = ?
	`, now, username)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// isUniqueConstraintError checks if an error is a unique constraint violation.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsIgnoreCase(errStr, "unique constraint") ||
		containsIgnoreCase(errStr, "UNIQUE constraint failed")
}

// Compile-time interface check.
var _ auth.UserStore = (*UserStoreAdapter)(nil)
