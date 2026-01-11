// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// TestResultRepository handles test result persistence.
type TestResultRepository struct {
	db *DB
}

// Create stores a new test result.
func (r *TestResultRepository) Create(ctx context.Context, result *TestResult) (int64, error) {
	if result.Timestamp.IsZero() {
		result.Timestamp = time.Now().UTC()
	}

	res, err := r.db.Exec(ctx, `
		INSERT INTO test_results (run_id, metric_type, frame_size, value, unit, timestamp, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		result.RunID, result.MetricType, result.FrameSize, result.Value,
		result.Unit, result.Timestamp.Format(time.RFC3339), result.Metadata,
	)
	if err != nil {
		return 0, fmt.Errorf("creating test result: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}

	result.ID = id
	return id, nil
}

// CreateBatch stores multiple test results efficiently.
func (r *TestResultRepository) CreateBatch(ctx context.Context, results []TestResult) error {
	if len(results) == 0 {
		return nil
	}

	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO test_results (run_id, metric_type, frame_size, value, unit, timestamp, metadata_json)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("preparing statement: %w", err)
		}
		defer func() { _ = stmt.Close() }()

		for i := range results {
			result := &results[i]
			if result.Timestamp.IsZero() {
				result.Timestamp = time.Now().UTC()
			}

			res, execErr := stmt.ExecContext(ctx,
				result.RunID, result.MetricType, result.FrameSize, result.Value,
				result.Unit, result.Timestamp.Format(time.RFC3339), result.Metadata,
			)
			if execErr != nil {
				return fmt.Errorf("inserting result %d: %w", i, execErr)
			}

			if id, idErr := res.LastInsertId(); idErr == nil {
				result.ID = id
			}
		}

		return nil
	})
}

// Get retrieves a test result by ID.
func (r *TestResultRepository) Get(ctx context.Context, id int64) (*TestResult, error) {
	var result TestResult
	var timestamp string

	err := r.db.QueryRow(ctx, `
		SELECT id, run_id, metric_type, frame_size, value, unit, timestamp, metadata_json
		FROM test_results WHERE id = ?
	`, id).Scan(
		&result.ID, &result.RunID, &result.MetricType, &result.FrameSize,
		&result.Value, &result.Unit, &timestamp, &result.Metadata,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying test result: %w", err)
	}

	if t, parseErr := time.Parse(time.RFC3339, timestamp); parseErr == nil {
		result.Timestamp = t
	}

	return &result, nil
}

// ListByRun retrieves all results for a test run.
func (r *TestResultRepository) ListByRun(ctx context.Context, runID string) ([]TestResult, error) {
	return r.List(ctx, TestResultQueryOptions{RunID: runID})
}

// List retrieves test results based on query options.
func (r *TestResultRepository) List(ctx context.Context, opts TestResultQueryOptions) ([]TestResult, error) {
	query := `
		SELECT id, run_id, metric_type, frame_size, value, unit, timestamp, metadata_json
		FROM test_results
		WHERE 1=1
	`
	args := []any{}

	if opts.RunID != "" {
		query += " AND run_id = ?"
		args = append(args, opts.RunID)
	}
	if opts.MetricType != "" {
		query += " AND metric_type = ?"
		args = append(args, opts.MetricType)
	}
	if opts.FrameSize != nil {
		query += " AND frame_size = ?"
		args = append(args, *opts.FrameSize)
	}
	if !opts.TimeRange.Start.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, opts.TimeRange.Start.Format(time.RFC3339))
	}
	if !opts.TimeRange.End.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, opts.TimeRange.End.Format(time.RFC3339))
	}

	query += " ORDER BY timestamp ASC"

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
		return nil, fmt.Errorf("querying test results: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []TestResult
	for rows.Next() {
		var result TestResult
		var timestamp string

		if scanErr := rows.Scan(
			&result.ID, &result.RunID, &result.MetricType, &result.FrameSize,
			&result.Value, &result.Unit, &timestamp, &result.Metadata,
		); scanErr != nil {
			return nil, fmt.Errorf("scanning test result row: %w", scanErr)
		}

		if t, parseErr := time.Parse(time.RFC3339, timestamp); parseErr == nil {
			result.Timestamp = t
		}

		results = append(results, result)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterating test result rows: %w", rowsErr)
	}

	return results, nil
}

// DeleteByRun removes all results for a test run.
func (r *TestResultRepository) DeleteByRun(ctx context.Context, runID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM test_results WHERE run_id = ?`, runID)
	if err != nil {
		return fmt.Errorf("deleting test results: %w", err)
	}
	return nil
}

// GetAggregates returns aggregated metrics for a test run.
func (r *TestResultRepository) GetAggregates(ctx context.Context, runID, metricType string) (*MetricAggregates, error) {
	var agg MetricAggregates
	err := r.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as count,
			MIN(value) as min_value,
			MAX(value) as max_value,
			AVG(value) as avg_value,
			SUM(value) as sum_value
		FROM test_results 
		WHERE run_id = ? AND metric_type = ?
	`, runID, metricType).Scan(
		&agg.Count, &agg.Min, &agg.Max, &agg.Avg, &agg.Sum,
	)
	if err != nil {
		return nil, fmt.Errorf("calculating aggregates: %w", err)
	}

	return &agg, nil
}

// MetricAggregates holds aggregated metric values.
type MetricAggregates struct {
	Count int     `json:"count"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Avg   float64 `json:"avg"`
	Sum   float64 `json:"sum"`
}

// GetByFrameSize returns results grouped by frame size for a run.
func (r *TestResultRepository) GetByFrameSize(ctx context.Context, runID, metricType string) (map[int][]TestResult, error) {
	results, err := r.List(ctx, TestResultQueryOptions{
		RunID:      runID,
		MetricType: metricType,
	})
	if err != nil {
		return nil, err
	}

	grouped := make(map[int][]TestResult)
	for _, result := range results {
		frameSize := 0
		if result.FrameSize != nil {
			frameSize = *result.FrameSize
		}
		grouped[frameSize] = append(grouped[frameSize], result)
	}

	return grouped, nil
}

// CreateSummary creates a test summary from results.
func (r *TestResultRepository) CreateSummary(ctx context.Context, summary *TestSummary) (int64, error) {
	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now().UTC()
	}

	res, err := r.db.Exec(ctx, `
		INSERT INTO test_summaries (run_id, module, test_type, pass, throughput_mbps,
		                           latency_avg_us, latency_min_us, latency_max_us,
		                           jitter_us, frame_loss_pct, frames_sent, frames_received, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		summary.RunID, summary.Module, summary.TestType, summary.Pass,
		summary.ThroughputMbps, summary.LatencyAvgUs, summary.LatencyMinUs,
		summary.LatencyMaxUs, summary.JitterUs, summary.FrameLossPct,
		summary.FramesSent, summary.FramesReceived, summary.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("creating test summary: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}

	summary.ID = id
	return id, nil
}

// GetSummary retrieves the test summary for a run.
func (r *TestResultRepository) GetSummary(ctx context.Context, runID string) (*TestSummary, error) {
	var summary TestSummary
	var createdAt string

	err := r.db.QueryRow(ctx, `
		SELECT id, run_id, module, test_type, pass, throughput_mbps,
		       latency_avg_us, latency_min_us, latency_max_us, jitter_us,
		       frame_loss_pct, frames_sent, frames_received, created_at
		FROM test_summaries WHERE run_id = ?
	`, runID).Scan(
		&summary.ID, &summary.RunID, &summary.Module, &summary.TestType, &summary.Pass,
		&summary.ThroughputMbps, &summary.LatencyAvgUs, &summary.LatencyMinUs,
		&summary.LatencyMaxUs, &summary.JitterUs, &summary.FrameLossPct,
		&summary.FramesSent, &summary.FramesReceived, &createdAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying test summary: %w", err)
	}

	if t, parseErr := time.Parse(time.RFC3339, createdAt); parseErr == nil {
		summary.CreatedAt = t
	}

	return &summary, nil
}
