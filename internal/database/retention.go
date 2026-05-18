// SPDX-License-Identifier: BUSL-1.1

package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/krisarmstrong/stem/internal/logging"
)

// RetentionConfig configures data retention policies.
type RetentionConfig struct {
	// TestResultRetention is how long to keep test results (0 = forever)
	TestResultRetention time.Duration

	// AuditLogRetention is how long to keep audit logs (0 = forever)
	AuditLogRetention time.Duration

	// CleanupInterval is how often to run the cleanup job
	CleanupInterval time.Duration
}

const defaultCleanupInterval = 24 * time.Hour

// DefaultRetentionConfig returns sensible default retention settings.
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		TestResultRetention: 90 * 24 * time.Hour,    // 90 days
		AuditLogRetention:   365 * 24 * time.Hour,   // 1 year
		CleanupInterval:     defaultCleanupInterval, // Daily
	}
}

// RetentionManager handles automatic cleanup of old data.
type RetentionManager struct {
	db     *DB
	config RetentionConfig

	mu       sync.Mutex
	stopChan chan struct{}
	running  bool
}

// NewRetentionManager creates a new retention manager.
func NewRetentionManager(db *DB, config RetentionConfig) *RetentionManager {
	return &RetentionManager{
		db:     db,
		config: config,
	}
}

// Start begins the background retention cleanup job.
func (m *RetentionManager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.stopChan = make(chan struct{})
	m.running = true

	go m.runLoop()
}

// Stop halts the background retention cleanup job.
func (m *RetentionManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	close(m.stopChan)
	m.running = false
}

// runLoop is the main cleanup loop.
func (m *RetentionManager) runLoop() {
	// Run immediately on start
	_ = m.RunCleanup(context.Background())

	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = m.RunCleanup(context.Background())
		case <-m.stopChan:
			return
		}
	}
}

// RunCleanup performs a single cleanup pass.
func (m *RetentionManager) RunCleanup(ctx context.Context) error {
	var errs []error

	if m.config.TestResultRetention > 0 {
		if err := m.cleanupTestResults(ctx); err != nil {
			errs = append(errs, fmt.Errorf("cleaning test results: %w", err))
		}
	}

	if m.config.AuditLogRetention > 0 {
		if err := m.cleanupAuditLogs(ctx); err != nil {
			errs = append(errs, fmt.Errorf("cleaning audit logs: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// cleanupTestResults removes old test runs and their results.
func (m *RetentionManager) cleanupTestResults(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-m.config.TestResultRetention)

	result, err := m.db.Exec(ctx, `
		DELETE FROM test_runs WHERE started_at < ?
	`, cutoff.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("deleting old test runs: %w", err)
	}

	deleted, _ := result.RowsAffected()
	if deleted > 0 {
		// Test results are deleted via CASCADE
		logging.InfoContext(
			ctx,
			"Retention cleanup deleted test runs",
			"deleted",
			deleted,
			"cutoff",
			cutoff.Format(time.RFC3339),
		)
	}

	return nil
}

// cleanupAuditLogs removes old audit log entries.
func (m *RetentionManager) cleanupAuditLogs(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-m.config.AuditLogRetention)

	deleted, err := m.db.AuditLog().DeleteOlderThan(ctx, cutoff)
	if err != nil {
		return err
	}

	if deleted > 0 {
		logging.InfoContext(
			ctx,
			"Retention cleanup deleted audit log entries",
			"deleted",
			deleted,
			"cutoff",
			cutoff.Format(time.RFC3339),
		)
	}

	return nil
}

// GetStats returns statistics about data retention.
func (m *RetentionManager) GetStats(ctx context.Context) (*RetentionStats, error) {
	stats := &RetentionStats{}

	// Count test runs
	testRunCount, err := m.db.TestRuns().Count(ctx, TestRunQueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("counting test runs: %w", err)
	}
	stats.TotalTestRuns = testRunCount

	// Count audit logs
	auditCount, err := m.db.AuditLog().Count(ctx, AuditLogQueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("counting audit logs: %w", err)
	}
	stats.TotalAuditLogs = auditCount

	// Get oldest test run
	runs, err := m.db.TestRuns().List(ctx, TestRunQueryOptions{Limit: 1})
	if err == nil && len(runs) > 0 {
		stats.OldestTestRun = &runs[0].StartedAt
	}

	// Get oldest audit log
	audits, err := m.db.AuditLog().List(ctx, AuditLogQueryOptions{Limit: 1})
	if err == nil && len(audits) > 0 {
		stats.OldestAuditLog = &audits[0].Timestamp
	}

	// Calculate retention thresholds
	now := time.Now().UTC()
	if m.config.TestResultRetention > 0 {
		t := now.Add(-m.config.TestResultRetention)
		stats.TestResultCutoff = &t
	}
	if m.config.AuditLogRetention > 0 {
		t := now.Add(-m.config.AuditLogRetention)
		stats.AuditLogCutoff = &t
	}

	return stats, nil
}

// RetentionStats contains statistics about stored data.
type RetentionStats struct {
	TotalTestRuns    int        `json:"totalTestRuns"`
	TotalAuditLogs   int        `json:"totalAuditLogs"`
	OldestTestRun    *time.Time `json:"oldestTestRun,omitempty"`
	OldestAuditLog   *time.Time `json:"oldestAuditLog,omitempty"`
	TestResultCutoff *time.Time `json:"testResultCutoff,omitempty"`
	AuditLogCutoff   *time.Time `json:"auditLogCutoff,omitempty"`
}

// Vacuum runs SQLite VACUUM to reclaim disk space.
func (m *RetentionManager) Vacuum(ctx context.Context) error {
	_, err := m.db.Exec(ctx, "VACUUM")
	if err != nil {
		return fmt.Errorf("running vacuum: %w", err)
	}
	return nil
}

// Analyze updates SQLite statistics for query optimization.
func (m *RetentionManager) Analyze(ctx context.Context) error {
	_, err := m.db.Exec(ctx, "ANALYZE")
	if err != nil {
		return fmt.Errorf("running analyze: %w", err)
	}
	return nil
}
