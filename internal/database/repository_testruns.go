// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TestRunRepository handles test run persistence.
type TestRunRepository struct {
	db *DB
}

// Create creates a new test run and returns its ID.
func (r *TestRunRepository) Create(ctx context.Context, run *TestRun) (string, error) {
	if run.ID == "" {
		run.ID = uuid.New().String()
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.Status == "" {
		run.Status = TestRunStatusPending
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO test_runs (id, module, test_type, status, config_json, interface_name, 
		                       target_address, started_at, completed_at, duration_ms, 
		                       error_message, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		run.ID, run.Module, run.TestType, run.Status, run.ConfigJSON,
		run.InterfaceName, run.TargetAddress, run.StartedAt.Format(time.RFC3339),
		formatNullableTime(run.CompletedAt), run.DurationMs, run.ErrorMessage, run.Metadata,
	)
	if err != nil {
		return "", fmt.Errorf("creating test run: %w", err)
	}

	return run.ID, nil
}

// Get retrieves a test run by ID.
func (r *TestRunRepository) Get(ctx context.Context, id string) (*TestRun, error) {
	var run TestRun
	var startedAt, completedAt sql.NullString

	err := r.db.QueryRow(ctx, `
		SELECT id, module, test_type, status, config_json, interface_name,
		       target_address, started_at, completed_at, duration_ms, 
		       error_message, metadata_json
		FROM test_runs WHERE id = ?
	`, id).Scan(
		&run.ID, &run.Module, &run.TestType, &run.Status, &run.ConfigJSON,
		&run.InterfaceName, &run.TargetAddress, &startedAt, &completedAt,
		&run.DurationMs, &run.ErrorMessage, &run.Metadata,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying test run: %w", err)
	}

	if startedAt.Valid {
		if t, parseErr := time.Parse(time.RFC3339, startedAt.String); parseErr == nil {
			run.StartedAt = t
		}
	}
	if completedAt.Valid {
		if t, parseErr := time.Parse(time.RFC3339, completedAt.String); parseErr == nil {
			run.CompletedAt = &t
		}
	}

	return &run, nil
}

// UpdateStatus updates the status of a test run.
func (r *TestRunRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.Exec(ctx, `
		UPDATE test_runs SET status = ? WHERE id = ?
	`, status, id)
	if err != nil {
		return fmt.Errorf("updating test run status: %w", err)
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

// Complete marks a test run as completed.
func (r *TestRunRepository) Complete(ctx context.Context, id string, status string, errorMsg string) error {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// Calculate duration from start time
	run, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	durationMs := now.Sub(run.StartedAt).Milliseconds()

	_, err = r.db.Exec(ctx, `
		UPDATE test_runs 
		SET status = ?, completed_at = ?, duration_ms = ?, error_message = ?
		WHERE id = ?
	`, status, nowStr, durationMs, errorMsg, id)
	if err != nil {
		return fmt.Errorf("completing test run: %w", err)
	}

	return nil
}

// List retrieves test runs based on query options.
func (r *TestRunRepository) List(ctx context.Context, opts TestRunQueryOptions) ([]TestRun, error) {
	query := `
		SELECT id, module, test_type, status, config_json, interface_name,
		       target_address, started_at, completed_at, duration_ms, 
		       error_message, metadata_json
		FROM test_runs
		WHERE 1=1
	`
	args := []any{}

	if opts.Module != "" {
		query += " AND module = ?"
		args = append(args, opts.Module)
	}
	if opts.TestType != "" {
		query += " AND test_type = ?"
		args = append(args, opts.TestType)
	}
	if opts.Status != "" {
		query += " AND status = ?"
		args = append(args, opts.Status)
	}
	if !opts.TimeRange.Start.IsZero() {
		query += " AND started_at >= ?"
		args = append(args, opts.TimeRange.Start.Format(time.RFC3339))
	}
	if !opts.TimeRange.End.IsZero() {
		query += " AND started_at <= ?"
		args = append(args, opts.TimeRange.End.Format(time.RFC3339))
	}

	query += " ORDER BY started_at DESC"

	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying test runs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var runs []TestRun
	for rows.Next() {
		var run TestRun
		var startedAt, completedAt sql.NullString

		if scanErr := rows.Scan(
			&run.ID, &run.Module, &run.TestType, &run.Status, &run.ConfigJSON,
			&run.InterfaceName, &run.TargetAddress, &startedAt, &completedAt,
			&run.DurationMs, &run.ErrorMessage, &run.Metadata,
		); scanErr != nil {
			return nil, fmt.Errorf("scanning test run row: %w", scanErr)
		}

		if startedAt.Valid {
			if t, parseErr := time.Parse(time.RFC3339, startedAt.String); parseErr == nil {
				run.StartedAt = t
			}
		}
		if completedAt.Valid {
			if t, parseErr := time.Parse(time.RFC3339, completedAt.String); parseErr == nil {
				run.CompletedAt = &t
			}
		}

		runs = append(runs, run)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterating test run rows: %w", rowsErr)
	}

	return runs, nil
}

// Delete removes a test run and its associated results.
func (r *TestRunRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM test_runs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting test run: %w", err)
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

// Count returns the total number of test runs matching the filter.
func (r *TestRunRepository) Count(ctx context.Context, opts TestRunQueryOptions) (int, error) {
	query := `SELECT COUNT(*) FROM test_runs WHERE 1=1`
	args := []any{}

	if opts.Module != "" {
		query += " AND module = ?"
		args = append(args, opts.Module)
	}
	if opts.TestType != "" {
		query += " AND test_type = ?"
		args = append(args, opts.TestType)
	}
	if opts.Status != "" {
		query += " AND status = ?"
		args = append(args, opts.Status)
	}

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting test runs: %w", err)
	}

	return count, nil
}

// GetLatest returns the most recent test run for a module/type.
func (r *TestRunRepository) GetLatest(ctx context.Context, module, testType string) (*TestRun, error) {
	runs, err := r.List(ctx, TestRunQueryOptions{
		Module:   module,
		TestType: testType,
		Limit:    1,
	})
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return nil, ErrNotFound
	}
	return &runs[0], nil
}

// formatNullableTime formats a time pointer for database storage.
func formatNullableTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

// ErrNotFound is returned when a requested record doesn't exist.
var ErrNotFound = errors.New("record not found")
