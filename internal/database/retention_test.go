// SPDX-License-Identifier: BUSL-1.1

package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/database"
)

func TestDefaultRetentionConfig(t *testing.T) {
	cfg := database.DefaultRetentionConfig()

	if cfg.TestResultRetention != 90*24*time.Hour {
		t.Errorf("TestResultRetention = %v, want 90 days", cfg.TestResultRetention)
	}
	if cfg.AuditLogRetention != 365*24*time.Hour {
		t.Errorf("AuditLogRetention = %v, want 365 days", cfg.AuditLogRetention)
	}
	if cfg.CleanupInterval != 24*time.Hour {
		t.Errorf("CleanupInterval = %v, want 24h", cfg.CleanupInterval)
	}
}

// runStartedAt creates a test run whose started_at is an explicit past time, so
// retention cutoffs are exercised without waiting on the clock.
func runStartedAt(t *testing.T, db *database.DB, at time.Time) string {
	t.Helper()

	id, err := db.TestRuns().Create(context.Background(), &database.TestRun{
		Module:    "benchmark",
		TestType:  "rfc2544_throughput",
		StartedAt: at.UTC(),
	})
	if err != nil {
		t.Fatalf("Create test run failed: %v", err)
	}
	return id
}

func TestRetentionManagerRunCleanup(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	staleRun := runStartedAt(t, db, now.Add(-100*24*time.Hour))
	freshRun := runStartedAt(t, db, now.Add(-24*time.Hour))

	if err := db.TestResults().CreateBatch(ctx, []database.TestResult{
		{RunID: staleRun, MetricType: database.MetricTypeThroughput, Value: 1},
	}); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	staleAudit := logAt(t, db, &database.AuditLogEntry{
		Action: database.AuditActionLogin, User: "alice",
	}, now.Add(-400*24*time.Hour))
	freshAudit := logAt(t, db, &database.AuditLogEntry{
		Action: database.AuditActionLogin, User: "bob",
	}, now.Add(-24*time.Hour))

	manager := database.NewRetentionManager(db, database.DefaultRetentionConfig())
	if err := manager.RunCleanup(ctx); err != nil {
		t.Fatalf("RunCleanup failed: %v", err)
	}

	if _, err := db.TestRuns().Get(ctx, staleRun); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("a run older than the 90-day cutoff survived: %v", err)
	}
	if _, err := db.TestRuns().Get(ctx, freshRun); err != nil {
		t.Errorf("a run inside the retention window was deleted: %v", err)
	}
	if _, err := db.TestResults().ListByRun(ctx, staleRun); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("results of a deleted run survived (the CASCADE did not fire): %v", err)
	}

	if _, err := db.AuditLog().Get(ctx, staleAudit); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("an audit entry older than the 365-day cutoff survived: %v", err)
	}
	if _, err := db.AuditLog().Get(ctx, freshAudit); err != nil {
		t.Errorf("an audit entry inside the retention window was deleted: %v", err)
	}
}

func TestRetentionManagerZeroRetentionKeepsEverything(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	ancientRun := runStartedAt(t, db, now.Add(-10*365*24*time.Hour))
	ancientAudit := logAt(t, db, &database.AuditLogEntry{
		Action: database.AuditActionLogin, User: "alice",
	}, now.Add(-10*365*24*time.Hour))

	manager := database.NewRetentionManager(db, database.RetentionConfig{})
	if err := manager.RunCleanup(ctx); err != nil {
		t.Fatalf("RunCleanup failed: %v", err)
	}

	if _, err := db.TestRuns().Get(ctx, ancientRun); err != nil {
		t.Errorf("a zero TestResultRetention must keep every run, got %v", err)
	}
	if _, err := db.AuditLog().Get(ctx, ancientAudit); err != nil {
		t.Errorf("a zero AuditLogRetention must keep every audit entry, got %v", err)
	}
}

func TestRetentionManagerGetStats(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	oldest := now.Add(-48 * time.Hour)
	newest := now.Add(-time.Hour)
	runStartedAt(t, db, oldest)
	runStartedAt(t, db, newest)
	logAt(t, db, &database.AuditLogEntry{Action: database.AuditActionLogin, User: "alice"}, oldest)
	logAt(t, db, &database.AuditLogEntry{Action: database.AuditActionLogin, User: "bob"}, newest)

	cfg := database.DefaultRetentionConfig()
	stats, err := database.NewRetentionManager(db, cfg).GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.TotalTestRuns != 2 {
		t.Errorf("TotalTestRuns = %d, want 2", stats.TotalTestRuns)
	}
	if stats.TotalAuditLogs != 2 {
		t.Errorf("TotalAuditLogs = %d, want 2", stats.TotalAuditLogs)
	}

	if stats.OldestTestRun == nil || !stats.OldestTestRun.Equal(oldest) {
		t.Errorf("OldestTestRun = %v, want %v", stats.OldestTestRun, oldest)
	}
	if stats.OldestAuditLog == nil || !stats.OldestAuditLog.Equal(oldest) {
		t.Errorf("OldestAuditLog = %v, want %v", stats.OldestAuditLog, oldest)
	}

	if stats.TestResultCutoff == nil {
		t.Fatal("TestResultCutoff is nil")
	}
	wantCutoff := now.Add(-cfg.TestResultRetention)
	if stats.TestResultCutoff.Sub(wantCutoff).Abs() > time.Minute {
		t.Errorf("TestResultCutoff = %v, want about %v", stats.TestResultCutoff, wantCutoff)
	}
	if stats.AuditLogCutoff == nil {
		t.Fatal("AuditLogCutoff is nil")
	}
	wantAuditCutoff := now.Add(-cfg.AuditLogRetention)
	if stats.AuditLogCutoff.Sub(wantAuditCutoff).Abs() > time.Minute {
		t.Errorf("AuditLogCutoff = %v, want about %v", stats.AuditLogCutoff, wantAuditCutoff)
	}
}

func TestRetentionManagerGetStatsWithoutCutoffs(t *testing.T) {
	db := newTestDB(t)

	stats, err := database.NewRetentionManager(db, database.RetentionConfig{}).GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats.TestResultCutoff != nil || stats.AuditLogCutoff != nil {
		t.Errorf("zero retention must report no cutoffs, got %v and %v",
			stats.TestResultCutoff, stats.AuditLogCutoff)
	}
	if stats.OldestTestRun != nil || stats.OldestAuditLog != nil {
		t.Errorf("an empty database must report no oldest rows, got %v and %v",
			stats.OldestTestRun, stats.OldestAuditLog)
	}
}

func TestRetentionManagerStartRunsCleanupAndStops(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	staleRun := runStartedAt(t, db, time.Now().UTC().Add(-100*24*time.Hour))

	manager := database.NewRetentionManager(db, database.DefaultRetentionConfig())
	manager.Start()
	// Start is idempotent: a second call must not launch a second loop.
	manager.Start()
	t.Cleanup(manager.Stop)

	// The loop runs one cleanup pass immediately; poll for its effect rather
	// than sleeping for a tick.
	deadline := time.Now().Add(5 * time.Second)
	for {
		_, err := db.TestRuns().Get(ctx, staleRun)
		if errors.Is(err, database.ErrNotFound) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("Start did not run a cleanup pass within 5s (Get = %v)", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	manager.Stop()
	// Stop is idempotent: a second call must not close the channel twice.
	manager.Stop()
}

func TestRetentionManagerVacuumAndAnalyze(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	manager := database.NewRetentionManager(db, database.DefaultRetentionConfig())

	runID := runStartedAt(t, db, time.Now().UTC())

	if err := manager.Analyze(ctx); err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if err := manager.Vacuum(ctx); err != nil {
		t.Fatalf("Vacuum failed: %v", err)
	}

	if _, err := db.TestRuns().Get(ctx, runID); err != nil {
		t.Errorf("data did not survive VACUUM: %v", err)
	}
}
