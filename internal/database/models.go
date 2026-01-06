// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// TestResult represents a completed test result stored in the database.
type TestResult struct {
	ID         int64           `json:"id"`
	TestType   string          `json:"testType"`
	Module     string          `json:"module"`
	Status     string          `json:"status"`
	ResultJSON json.RawMessage `json:"resultJson,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}

// AuditLogEntry represents a security audit log entry.
type AuditLogEntry struct {
	ID        int64     `json:"id"`
	EventType string    `json:"eventType"`
	User      string    `json:"user,omitempty"`
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// Session represents a revoked token session for blacklist persistence.
type Session struct {
	TokenID   string    `json:"tokenId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// TestResultFilter defines optional filters for querying test results.
type TestResultFilter struct {
	Module   string
	TestType string
	Status   string
	Limit    int
	Offset   int
}

// SaveTestResult persists a test result to the database.
func (d *Database) SaveTestResult(ctx context.Context, result *TestResult) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	openErr := d.ensureOpen()
	if openErr != nil {
		return openErr
	}

	query := `INSERT INTO test_results (test_type, module, status, result_json, created_at)
	          VALUES (?, ?, ?, ?, ?)`

	createdAt := result.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	res, err := d.db.ExecContext(ctx, query,
		result.TestType,
		result.Module,
		result.Status,
		string(result.ResultJSON),
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("insert test result: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	result.ID = id
	result.CreatedAt = createdAt

	return nil
}

// GetTestResult retrieves a single test result by ID.
func (d *Database) GetTestResult(ctx context.Context, id int64) (*TestResult, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	openErr := d.ensureOpen()
	if openErr != nil {
		return nil, openErr
	}

	query := `SELECT id, test_type, module, status, result_json, created_at
	          FROM test_results WHERE id = ?`

	var result TestResult
	var resultJSON sql.NullString

	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&result.ID,
		&result.TestType,
		&result.Module,
		&result.Status,
		&resultJSON,
		&result.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("query test result: %w", err)
	}

	if resultJSON.Valid {
		result.ResultJSON = json.RawMessage(resultJSON.String)
	}

	return &result, nil
}

// GetTestResults retrieves test results with optional filtering.
func (d *Database) GetTestResults(ctx context.Context, filter *TestResultFilter) ([]TestResult, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	openErr := d.ensureOpen()
	if openErr != nil {
		return nil, openErr
	}

	query := `SELECT id, test_type, module, status, result_json, created_at
	          FROM test_results WHERE 1=1`
	args := []any{}

	if filter != nil {
		if filter.Module != "" {
			query += " AND module = ?"
			args = append(args, filter.Module)
		}
		if filter.TestType != "" {
			query += " AND test_type = ?"
			args = append(args, filter.TestType)
		}
		if filter.Status != "" {
			query += " AND status = ?"
			args = append(args, filter.Status)
		}
	}

	query += " ORDER BY created_at DESC"

	if filter != nil && filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query test results: %w", err)
	}
	defer rows.Close()

	var results []TestResult
	for rows.Next() {
		var result TestResult
		var resultJSON sql.NullString

		scanErr := rows.Scan(
			&result.ID,
			&result.TestType,
			&result.Module,
			&result.Status,
			&resultJSON,
			&result.CreatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("scan test result: %w", scanErr)
		}

		if resultJSON.Valid {
			result.ResultJSON = json.RawMessage(resultJSON.String)
		}

		results = append(results, result)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate test results: %w", rowsErr)
	}

	return results, nil
}

// SaveAuditLog persists an audit log entry to the database.
func (d *Database) SaveAuditLog(ctx context.Context, entry *AuditLogEntry) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	openErr := d.ensureOpen()
	if openErr != nil {
		return openErr
	}

	query := `INSERT INTO audit_log (event_type, user, details, created_at)
	          VALUES (?, ?, ?, ?)`

	createdAt := entry.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	res, err := d.db.ExecContext(ctx, query,
		entry.EventType,
		entry.User,
		entry.Details,
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	entry.ID = id
	entry.CreatedAt = createdAt

	return nil
}

// GetAuditLogs retrieves audit log entries with optional limit.
func (d *Database) GetAuditLogs(ctx context.Context, limit int) ([]AuditLogEntry, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	openErr := d.ensureOpen()
	if openErr != nil {
		return nil, openErr
	}

	query := `SELECT id, event_type, user, details, created_at
	          FROM audit_log ORDER BY created_at DESC`
	args := []any{}

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var entries []AuditLogEntry
	for rows.Next() {
		var entry AuditLogEntry
		var user, details sql.NullString

		scanErr := rows.Scan(
			&entry.ID,
			&entry.EventType,
			&user,
			&details,
			&entry.CreatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("scan audit log: %w", scanErr)
		}

		if user.Valid {
			entry.User = user.String
		}
		if details.Valid {
			entry.Details = details.String
		}

		entries = append(entries, entry)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", rowsErr)
	}

	return entries, nil
}

// SaveSession persists a revoked session token to the database.
func (d *Database) SaveSession(ctx context.Context, session *Session) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	openErr := d.ensureOpen()
	if openErr != nil {
		return openErr
	}

	query := `INSERT OR REPLACE INTO sessions (token_id, expires_at, created_at)
	          VALUES (?, ?, ?)`

	createdAt := session.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	_, err := d.db.ExecContext(ctx, query,
		session.TokenID,
		session.ExpiresAt,
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	session.CreatedAt = createdAt
	return nil
}

// IsSessionBlacklisted checks if a token ID is in the session blacklist.
func (d *Database) IsSessionBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	openErr := d.ensureOpen()
	if openErr != nil {
		return false, openErr
	}

	query := `SELECT COUNT(*) FROM sessions WHERE token_id = ? AND expires_at > ?`

	var count int
	err := d.db.QueryRowContext(ctx, query, tokenID, time.Now().UTC()).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query session: %w", err)
	}

	return count > 0, nil
}

// DeleteExpiredSessions removes all expired sessions from the database.
// This should be called periodically to clean up the sessions table.
func (d *Database) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	openErr := d.ensureOpen()
	if openErr != nil {
		return 0, openErr
	}

	query := `DELETE FROM sessions WHERE expires_at <= ?`

	res, err := d.db.ExecContext(ctx, query, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("get rows affected: %w", err)
	}

	return count, nil
}

// GetAllBlacklistedSessions retrieves all non-expired blacklisted sessions.
// Useful for loading the blacklist into memory on startup.
func (d *Database) GetAllBlacklistedSessions(ctx context.Context) ([]Session, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	openErr := d.ensureOpen()
	if openErr != nil {
		return nil, openErr
	}

	query := `SELECT token_id, expires_at, created_at
	          FROM sessions WHERE expires_at > ?
	          ORDER BY expires_at ASC`

	rows, err := d.db.QueryContext(ctx, query, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var session Session
		scanErr := rows.Scan(
			&session.TokenID,
			&session.ExpiresAt,
			&session.CreatedAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("scan session: %w", scanErr)
		}
		sessions = append(sessions, session)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate sessions: %w", rowsErr)
	}

	return sessions, nil
}
