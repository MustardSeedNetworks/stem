// SPDX-License-Identifier: BUSL-1.1

package database_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

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
	t.Run("create and get test run", testCreateAndGetTestRun)
	t.Run("update status", testUpdateTestRunStatus)
	t.Run("complete test run", testCompleteTestRun)
	t.Run("list test runs", testListTestRuns)
	t.Run("get not found", testGetTestRunNotFound)
}

func testCreateAndGetTestRun(t *testing.T) {
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
}

func testUpdateTestRunStatus(t *testing.T) {
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
}

func testCompleteTestRun(t *testing.T) {
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
}

func testListTestRuns(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	for range 3 {
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
}

func testGetTestRunNotFound(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.TestRuns().Get(ctx, "nonexistent-id")
	if !errors.Is(err, database.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTestResultRepository(t *testing.T) {
	t.Run("create and list results", testCreateAndListResults)
	t.Run("list results for missing run", testListResultsMissingRun)
	t.Run("create result with missing run", testCreateResultMissingRun)
	t.Run("get results by invalid run", testGetResultsByInvalidRun)
}

func testCreateAndListResults(t *testing.T) {
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
}

func testListResultsMissingRun(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	run := &database.TestRun{Module: "benchmark", TestType: "throughput"}
	runID, _ := db.TestRuns().Create(ctx, run)

	_, err := db.TestResults().ListByRun(ctx, runID)
	if err != nil {
		t.Fatalf("ListByRun failed: %v", err)
	}
}

func testCreateResultMissingRun(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	result := &database.TestResult{
		RunID:      "missing-run",
		MetricType: database.MetricTypeThroughput,
		Value:      1000.5,
		Unit:       "Mbps",
	}

	_, err := db.TestResults().Create(ctx, result)
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func testGetResultsByInvalidRun(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.TestResults().ListByRun(ctx, "invalid-id")
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
