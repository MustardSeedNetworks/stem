// SPDX-License-Identifier: BUSL-1.1

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SessionRepository handles session/token blacklist persistence.
type SessionRepository struct {
	db *DB
}

// Blacklist adds a token to the blacklist.
func (r *SessionRepository) Blacklist(ctx context.Context, session *Session) (int64, error) {
	if session.BlacklistedAt.IsZero() {
		session.BlacklistedAt = time.Now().UTC()
	}

	res, err := r.db.Exec(ctx, `
		INSERT INTO sessions (token_id, username, reason, blacklisted_at, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`,
		session.TokenID, session.Username, session.Reason,
		session.BlacklistedAt.Format(time.RFC3339),
		session.ExpiresAt.Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("blacklisting session: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}

	session.ID = id
	return id, nil
}

// IsBlacklisted checks if a token is blacklisted.
func (r *SessionRepository) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM sessions WHERE token_id = ?
	`, tokenID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking blacklist: %w", err)
	}

	return count > 0, nil
}

// Get retrieves a session by token ID.
func (r *SessionRepository) Get(ctx context.Context, tokenID string) (*Session, error) {
	var session Session
	var blacklistedAt, expiresAt string

	err := r.db.QueryRow(ctx, `
		SELECT id, token_id, username, reason, blacklisted_at, expires_at
		FROM sessions WHERE token_id = ?
	`, tokenID).Scan(
		&session.ID, &session.TokenID, &session.Username, &session.Reason,
		&blacklistedAt, &expiresAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying session: %w", err)
	}

	if t, parseErr := time.Parse(time.RFC3339, blacklistedAt); parseErr == nil {
		session.BlacklistedAt = t
	}
	if t, parseErr := time.Parse(time.RFC3339, expiresAt); parseErr == nil {
		session.ExpiresAt = t
	}

	return &session, nil
}

// ListAll returns all blacklisted sessions.
func (r *SessionRepository) ListAll(ctx context.Context) ([]Session, error) {
	var sessions []Session
	err := r.db.Query(ctx, `
		SELECT id, token_id, username, reason, blacklisted_at, expires_at
		FROM sessions ORDER BY blacklisted_at DESC
	`, func(rows *sql.Rows) error {
		for rows.Next() {
			var session Session
			var blacklistedAt, expiresAt string
			if scanErr := rows.Scan(
				&session.ID, &session.TokenID, &session.Username, &session.Reason,
				&blacklistedAt, &expiresAt,
			); scanErr != nil {
				return fmt.Errorf("scanning session row: %w", scanErr)
			}
			if t, parseErr := time.Parse(time.RFC3339, blacklistedAt); parseErr == nil {
				session.BlacklistedAt = t
			}
			if t, parseErr := time.Parse(time.RFC3339, expiresAt); parseErr == nil {
				session.ExpiresAt = t
			}
			sessions = append(sessions, session)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("querying sessions: %w", err)
	}
	return sessions, nil
}

// ListByUser returns all blacklisted sessions for a user.
func (r *SessionRepository) ListByUser(ctx context.Context, username string) ([]Session, error) {
	var sessions []Session
	err := r.db.Query(ctx, `
		SELECT id, token_id, username, reason, blacklisted_at, expires_at
		FROM sessions WHERE username = ? ORDER BY blacklisted_at DESC
	`, func(rows *sql.Rows) error {
		for rows.Next() {
			var session Session
			var blacklistedAt, expiresAt string
			if scanErr := rows.Scan(
				&session.ID, &session.TokenID, &session.Username, &session.Reason,
				&blacklistedAt, &expiresAt,
			); scanErr != nil {
				return fmt.Errorf("scanning session row: %w", scanErr)
			}
			if t, parseErr := time.Parse(time.RFC3339, blacklistedAt); parseErr == nil {
				session.BlacklistedAt = t
			}
			if t, parseErr := time.Parse(time.RFC3339, expiresAt); parseErr == nil {
				session.ExpiresAt = t
			}
			sessions = append(sessions, session)
		}
		return rows.Err()
	}, username)
	if err != nil {
		return nil, fmt.Errorf("querying sessions: %w", err)
	}
	return sessions, nil
}

// Delete removes a session from the blacklist.
func (r *SessionRepository) Delete(ctx context.Context, tokenID string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE token_id = ?`, tokenID)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// CleanupExpired removes sessions whose tokens have expired.
func (r *SessionRepository) CleanupExpired(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := r.db.Exec(ctx, `
		DELETE FROM sessions WHERE expires_at < ?
	`, now)
	if err != nil {
		return 0, fmt.Errorf("cleaning up expired sessions: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("checking rows affected: %w", err)
	}

	return deleted, nil
}

// Count returns the total number of blacklisted sessions.
func (r *SessionRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM sessions`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting sessions: %w", err)
	}
	return count, nil
}
