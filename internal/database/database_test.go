// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package database_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/database"
)

// newTestDB creates a new database in a temp directory for testing.
// Migrations are run automatically during Open.
func newTestDB(t *testing.T) *database.DB {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestOpen(t *testing.T) {
	t.Run("creates database successfully", func(t *testing.T) {
		db := newTestDB(t)
		if db == nil {
			t.Fatal("expected non-nil database")
		}
	})

	t.Run("returns correct path", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		db, err := database.Open(dbPath)
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		defer db.Close()

		if db.Path() != dbPath {
			t.Errorf("Path() = %q, want %q", db.Path(), dbPath)
		}
	})

	t.Run("fails with empty path", func(t *testing.T) {
		_, err := database.Open("")
		if err == nil {
			t.Error("expected error for empty path")
		}
	})
}

func TestClose(t *testing.T) {
	t.Run("closes successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		db, err := database.Open(dbPath)
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}

		err = db.Close()
		if err != nil {
			t.Errorf("Close failed: %v", err)
		}
	})

	t.Run("close is idempotent", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		db, err := database.Open(dbPath)
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}

		err1 := db.Close()
		if err1 != nil {
			t.Errorf("first Close failed: %v", err1)
		}

		err2 := db.Close()
		if err2 != nil {
			t.Errorf("second Close failed: %v", err2)
		}
	})
}

func TestPing(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	err := db.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestTestRunRepository(t *testing.T) {
	t.Run("create and get test run", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		run := &database.TestRun{
			Module:        "benchmark",
			TestType:      "throughput",
			Status:        database.TestRunStatusPending,
			InterfaceName: "eth0",
			TargetAddress: "192.168.1.1",
		}

		id, err := db.TestRuns().Create(ctx, run)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id == "" {
			t.Error("expected non-empty ID")
		}

		retrieved, err := db.TestRuns().Get(ctx, id)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if retrieved.Module != run.Module {
			t.Errorf("Module = %q, want %q", retrieved.Module, run.Module)
		}
		if retrieved.TestType != run.TestType {
			t.Errorf("TestType = %q, want %q", retrieved.TestType, run.TestType)
		}
	})

	t.Run("update status", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		run := &database.TestRun{
			Module:   "benchmark",
			TestType: "latency",
		}

		id, _ := db.TestRuns().Create(ctx, run)

		err := db.TestRuns().UpdateStatus(ctx, id, database.TestRunStatusRunning)
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		retrieved, _ := db.TestRuns().Get(ctx, id)
		if retrieved.Status != database.TestRunStatusRunning {
			t.Errorf("Status = %q, want %q", retrieved.Status, database.TestRunStatusRunning)
		}
	})

	t.Run("complete test run", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		run := &database.TestRun{
			Module:   "benchmark",
			TestType: "throughput",
		}

		id, _ := db.TestRuns().Create(ctx, run)

		err := db.TestRuns().Complete(ctx, id, database.TestRunStatusCompleted, "")
		if err != nil {
			t.Fatalf("Complete failed: %v", err)
		}

		retrieved, _ := db.TestRuns().Get(ctx, id)
		if retrieved.Status != database.TestRunStatusCompleted {
			t.Errorf("Status = %q, want %q", retrieved.Status, database.TestRunStatusCompleted)
		}
		if retrieved.CompletedAt == nil {
			t.Error("expected CompletedAt to be set")
		}
	})

	t.Run("list test runs", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		for i := 0; i < 3; i++ {
			run := &database.TestRun{
				Module:   "benchmark",
				TestType: "throughput",
			}
			_, _ = db.TestRuns().Create(ctx, run)
		}

		runs, err := db.TestRuns().List(ctx, database.TestRunQueryOptions{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(runs) != 3 {
			t.Errorf("got %d runs, want 3", len(runs))
		}
	})

	t.Run("get not found", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		_, err := db.TestRuns().Get(ctx, "nonexistent-id")
		if !errors.Is(err, database.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTestResultRepository(t *testing.T) {
	t.Run("create and list results", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		// First create a test run
		run := &database.TestRun{Module: "benchmark", TestType: "throughput"}
		runID, _ := db.TestRuns().Create(ctx, run)

		result := &database.TestResult{
			RunID:      runID,
			MetricType: database.MetricTypeThroughput,
			Value:      1000.5,
			Unit:       "Mbps",
		}

		id, err := db.TestResults().Create(ctx, result)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id == 0 {
			t.Error("expected non-zero ID")
		}

		results, err := db.TestResults().ListByRun(ctx, runID)
		if err != nil {
			t.Fatalf("ListByRun failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("got %d results, want 1", len(results))
		}
	})

	t.Run("create batch", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		run := &database.TestRun{Module: "benchmark", TestType: "throughput"}
		runID, _ := db.TestRuns().Create(ctx, run)

		results := []database.TestResult{
			{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 100},
			{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 200},
			{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 300},
		}

		err := db.TestResults().CreateBatch(ctx, results)
		if err != nil {
			t.Fatalf("CreateBatch failed: %v", err)
		}

		retrieved, _ := db.TestResults().ListByRun(ctx, runID)
		if len(retrieved) != 3 {
			t.Errorf("got %d results, want 3", len(retrieved))
		}
	})

	t.Run("get aggregates", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		run := &database.TestRun{Module: "benchmark", TestType: "throughput"}
		runID, _ := db.TestRuns().Create(ctx, run)

		results := []database.TestResult{
			{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 100},
			{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 200},
			{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 300},
		}
		_ = db.TestResults().CreateBatch(ctx, results)

		agg, err := db.TestResults().GetAggregates(ctx, runID, database.MetricTypeThroughput)
		if err != nil {
			t.Fatalf("GetAggregates failed: %v", err)
		}

		if agg.Count != 3 {
			t.Errorf("Count = %d, want 3", agg.Count)
		}
		if agg.Min != 100 {
			t.Errorf("Min = %f, want 100", agg.Min)
		}
		if agg.Max != 300 {
			t.Errorf("Max = %f, want 300", agg.Max)
		}
		if agg.Avg != 200 {
			t.Errorf("Avg = %f, want 200", agg.Avg)
		}
	})
}

func TestSettingsRepository(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		err := db.Settings().Set(ctx, "test_key", "test_value")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		value, err := db.Settings().Get(ctx, "test_key")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if value != "test_value" {
			t.Errorf("value = %q, want %q", value, "test_value")
		}
	})

	t.Run("get with default", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		value, err := db.Settings().GetWithDefault(ctx, "nonexistent", "default")
		if err != nil {
			t.Fatalf("GetWithDefault failed: %v", err)
		}

		if value != "default" {
			t.Errorf("value = %q, want %q", value, "default")
		}
	})

	t.Run("list settings", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		_ = db.Settings().Set(ctx, "key1", "value1")
		_ = db.Settings().Set(ctx, "key2", "value2")

		settings, err := db.Settings().List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(settings) != 2 {
			t.Errorf("got %d settings, want 2", len(settings))
		}
	})

	t.Run("delete setting", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		_ = db.Settings().Set(ctx, "to_delete", "value")

		err := db.Settings().Delete(ctx, "to_delete")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = db.Settings().Get(ctx, "to_delete")
		if !errors.Is(err, database.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestAuditLogRepository(t *testing.T) {
	t.Run("log and list", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		entry := &database.AuditLogEntry{
			Action:    database.AuditActionLogin,
			User:      "admin",
			IPAddress: "192.168.1.1",
		}

		id, err := db.AuditLog().Log(ctx, entry)
		if err != nil {
			t.Fatalf("Log failed: %v", err)
		}
		if id == 0 {
			t.Error("expected non-zero ID")
		}

		entries, err := db.AuditLog().List(ctx, database.AuditLogQueryOptions{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(entries) != 1 {
			t.Errorf("got %d entries, want 1", len(entries))
		}
	})

	t.Run("list by user", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		_ = db.AuditLog().LogAction(ctx, database.AuditActionLogin, "user1", "1.1.1.1")
		_ = db.AuditLog().LogAction(ctx, database.AuditActionLogin, "user2", "2.2.2.2")
		_ = db.AuditLog().LogAction(ctx, database.AuditActionLogout, "user1", "1.1.1.1")

		entries, err := db.AuditLog().ListByUser(ctx, "user1", 10)
		if err != nil {
			t.Fatalf("ListByUser failed: %v", err)
		}

		if len(entries) != 2 {
			t.Errorf("got %d entries, want 2", len(entries))
		}
	})

	t.Run("delete older than", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		_ = db.AuditLog().LogAction(ctx, "test", "user", "1.1.1.1")

		// Verify entry exists
		entries, _ := db.AuditLog().List(ctx, database.AuditLogQueryOptions{})
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		// Delete entries older than 1 hour ago (should delete nothing since entry is new)
		deleted, err := db.AuditLog().DeleteOlderThan(ctx, time.Now().Add(-time.Hour))
		if err != nil {
			t.Fatalf("DeleteOlderThan failed: %v", err)
		}
		if deleted != 0 {
			t.Errorf("deleted = %d, want 0", deleted)
		}

		// Verify entry still exists
		entries, _ = db.AuditLog().List(ctx, database.AuditLogQueryOptions{})
		if len(entries) != 1 {
			t.Errorf("entry should still exist, got %d entries", len(entries))
		}
	})
}

func TestSessionRepository(t *testing.T) {
	t.Run("blacklist and check", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		session := &database.Session{
			TokenID:   "token-123",
			Username:  "admin",
			Reason:    database.SessionReasonLogout,
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}

		_, err := db.Sessions().Blacklist(ctx, session)
		if err != nil {
			t.Fatalf("Blacklist failed: %v", err)
		}

		blacklisted, err := db.Sessions().IsBlacklisted(ctx, "token-123")
		if err != nil {
			t.Fatalf("IsBlacklisted failed: %v", err)
		}
		if !blacklisted {
			t.Error("expected token to be blacklisted")
		}
	})

	t.Run("not blacklisted", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		blacklisted, err := db.Sessions().IsBlacklisted(ctx, "unknown-token")
		if err != nil {
			t.Fatalf("IsBlacklisted failed: %v", err)
		}
		if blacklisted {
			t.Error("expected token to not be blacklisted")
		}
	})

	t.Run("cleanup expired", func(t *testing.T) {
		db := newTestDB(t)
		ctx := context.Background()

		// Create expired session
		expired := &database.Session{
			TokenID:   "expired-token",
			Username:  "admin",
			Reason:    database.SessionReasonLogout,
			ExpiresAt: time.Now().Add(-time.Hour).UTC(),
		}
		_, _ = db.Sessions().Blacklist(ctx, expired)

		// Create valid session
		valid := &database.Session{
			TokenID:   "valid-token",
			Username:  "admin",
			Reason:    database.SessionReasonLogout,
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}
		_, _ = db.Sessions().Blacklist(ctx, valid)

		deleted, err := db.Sessions().CleanupExpired(ctx)
		if err != nil {
			t.Fatalf("CleanupExpired failed: %v", err)
		}
		if deleted != 1 {
			t.Errorf("deleted = %d, want 1", deleted)
		}

		// Valid should still be blacklisted
		blacklisted, _ := db.Sessions().IsBlacklisted(ctx, "valid-token")
		if !blacklisted {
			t.Error("valid session should still be blacklisted")
		}
	})
}

func TestSchemaVersion(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	version, err := db.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion failed: %v", err)
	}

	// Should be >= 1 since migrations run automatically
	if version < 1 {
		t.Errorf("version = %d, expected >= 1", version)
	}
}

func TestMigrationStatus(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	migrations, err := db.MigrationStatus(ctx)
	if err != nil {
		t.Fatalf("MigrationStatus failed: %v", err)
	}

	if len(migrations) == 0 {
		t.Error("expected at least one migration")
	}

	// All migrations should be applied
	for _, m := range migrations {
		if !m.Applied {
			t.Errorf("migration %d (%s) not applied", m.Version, m.Description)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	done := make(chan bool)
	iterations := 10

	// Concurrent writers
	go func() {
		for i := 0; i < iterations; i++ {
			run := &database.TestRun{Module: "benchmark", TestType: "throughput"}
			_, _ = db.TestRuns().Create(ctx, run)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			_ = db.Settings().Set(ctx, "concurrent_key", "value")
		}
		done <- true
	}()

	// Concurrent readers
	go func() {
		for i := 0; i < iterations; i++ {
			_, _ = db.TestRuns().List(ctx, database.TestRunQueryOptions{})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			_, _ = db.Settings().List(ctx)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}
}
