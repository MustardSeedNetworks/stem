// SPDX-License-Identifier: BUSL-1.1

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// AuditLogRepository handles audit log persistence.
type AuditLogRepository struct {
	db *DB
}

// Log creates a new audit log entry.
func (r *AuditLogRepository) Log(ctx context.Context, entry *AuditLogEntry) (int64, error) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	res, err := r.db.Exec(ctx, `
		INSERT INTO audit_log (action, user, resource_type, resource_id, 
		                       old_value_json, new_value_json, ip_address, user_agent, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		entry.Action, entry.User, entry.ResourceType, entry.ResourceID,
		entry.OldValueJSON, entry.NewValueJSON, entry.IPAddress, entry.UserAgent,
		entry.Timestamp.Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("creating audit log entry: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}

	entry.ID = id
	return id, nil
}

// LogAction is a convenience method for logging simple actions.
func (r *AuditLogRepository) LogAction(ctx context.Context, action, user, ipAddress string) error {
	_, err := r.Log(ctx, &AuditLogEntry{
		Action:    action,
		User:      user,
		IPAddress: ipAddress,
	})
	return err
}

// LogResourceChange logs a change to a resource.
func (r *AuditLogRepository) LogResourceChange(
	ctx context.Context,
	action, user, resourceType, resourceID, oldValue, newValue, ipAddress string,
) error {
	_, err := r.Log(ctx, &AuditLogEntry{
		Action:       action,
		User:         user,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValueJSON: oldValue,
		NewValueJSON: newValue,
		IPAddress:    ipAddress,
	})
	return err
}

// Get retrieves an audit log entry by ID.
func (r *AuditLogRepository) Get(ctx context.Context, id int64) (*AuditLogEntry, error) {
	var entry AuditLogEntry
	var timestamp string

	err := r.db.QueryRow(ctx, `
		SELECT id, action, user, resource_type, resource_id,
		       old_value_json, new_value_json, ip_address, user_agent, timestamp
		FROM audit_log WHERE id = ?
	`, id).Scan(
		&entry.ID, &entry.Action, &entry.User, &entry.ResourceType, &entry.ResourceID,
		&entry.OldValueJSON, &entry.NewValueJSON, &entry.IPAddress, &entry.UserAgent, &timestamp,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying audit log entry: %w", err)
	}

	if t, parseErr := time.Parse(time.RFC3339, timestamp); parseErr == nil {
		entry.Timestamp = t
	}

	return &entry, nil
}

// AuditLogQueryOptions specifies criteria for querying audit logs.
type AuditLogQueryOptions struct {
	Action       string
	User         string
	ResourceType string
	ResourceID   string
	TimeRange    TimeRange
	Limit        int
	Offset       int
}

// List retrieves audit log entries based on query options.
func (r *AuditLogRepository) List(ctx context.Context, opts AuditLogQueryOptions) ([]AuditLogEntry, error) {
	query := `
		SELECT id, action, user, resource_type, resource_id,
		       old_value_json, new_value_json, ip_address, user_agent, timestamp
		FROM audit_log
		WHERE 1=1
	`
	args := []any{}

	if opts.Action != "" {
		query += " AND action = ?"
		args = append(args, opts.Action)
	}
	if opts.User != "" {
		query += " AND user = ?"
		args = append(args, opts.User)
	}
	if opts.ResourceType != "" {
		query += " AND resource_type = ?"
		args = append(args, opts.ResourceType)
	}
	if opts.ResourceID != "" {
		query += " AND resource_id = ?"
		args = append(args, opts.ResourceID)
	}
	if !opts.TimeRange.Start.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, opts.TimeRange.Start.Format(time.RFC3339))
	}
	if !opts.TimeRange.End.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, opts.TimeRange.End.Format(time.RFC3339))
	}

	query += " ORDER BY timestamp DESC"

	if opts.Limit > 0 {
		query += limitClause
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		query += offsetClause
		args = append(args, opts.Offset)
	}

	var entries []AuditLogEntry
	err := r.db.Query(ctx, query, func(rows *sql.Rows) error {
		for rows.Next() {
			var entry AuditLogEntry
			var timestamp string

			if scanErr := rows.Scan(
				&entry.ID, &entry.Action, &entry.User, &entry.ResourceType, &entry.ResourceID,
				&entry.OldValueJSON, &entry.NewValueJSON, &entry.IPAddress, &entry.UserAgent, &timestamp,
			); scanErr != nil {
				return fmt.Errorf("scanning audit log row: %w", scanErr)
			}

			if t, parseErr := time.Parse(time.RFC3339, timestamp); parseErr == nil {
				entry.Timestamp = t
			}

			entries = append(entries, entry)
		}
		return rows.Err()
	}, args...)
	if err != nil {
		return nil, fmt.Errorf("querying audit log: %w", err)
	}

	return entries, nil
}

// ListByUser returns all audit log entries for a specific user.
func (r *AuditLogRepository) ListByUser(ctx context.Context, user string, limit int) ([]AuditLogEntry, error) {
	return r.List(ctx, AuditLogQueryOptions{
		User:  user,
		Limit: limit,
	})
}

// ListByResource returns all audit log entries for a specific resource.
func (r *AuditLogRepository) ListByResource(
	ctx context.Context,
	resourceType, resourceID string,
	limit int,
) ([]AuditLogEntry, error) {
	return r.List(ctx, AuditLogQueryOptions{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Limit:        limit,
	})
}

// Count returns the total number of audit log entries matching the filter.
func (r *AuditLogRepository) Count(ctx context.Context, opts AuditLogQueryOptions) (int, error) {
	query := `SELECT COUNT(*) FROM audit_log WHERE 1=1`
	args := []any{}

	if opts.Action != "" {
		query += " AND action = ?"
		args = append(args, opts.Action)
	}
	if opts.User != "" {
		query += " AND user = ?"
		args = append(args, opts.User)
	}
	if opts.ResourceType != "" {
		query += " AND resource_type = ?"
		args = append(args, opts.ResourceType)
	}

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting audit log entries: %w", err)
	}

	return count, nil
}

// DeleteOlderThan removes audit log entries older than the given time.
func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.Exec(ctx, `
		DELETE FROM audit_log WHERE timestamp < ?
	`, before.Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("deleting old audit log entries: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("checking rows affected: %w", err)
	}

	return deleted, nil
}
