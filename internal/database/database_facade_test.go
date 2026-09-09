// SPDX-License-Identifier: BUSL-1.1

package database_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/database"
)

func TestDefaultPagination(t *testing.T) {
	page := database.DefaultPagination()

	if page.Offset != 0 {
		t.Errorf("Offset = %d, want 0", page.Offset)
	}
	if page.Limit != 100 {
		t.Errorf("Limit = %d, want 100", page.Limit)
	}
}

func TestDBRepositoriesAreCached(t *testing.T) {
	db := newTestDB(t)

	// Each accessor lazily builds its repository once; a second call must hand
	// back the same value rather than a fresh struct.
	runs, results := db.TestRuns(), db.TestResults()
	settings, auditLog, sessions := db.Settings(), db.AuditLog(), db.Sessions()

	if got := db.TestRuns(); got != runs {
		t.Error("TestRuns() returned two different repositories")
	}
	if got := db.TestResults(); got != results {
		t.Error("TestResults() returned two different repositories")
	}
	if got := db.Settings(); got != settings {
		t.Error("Settings() returned two different repositories")
	}
	if got := db.AuditLog(); got != auditLog {
		t.Error("AuditLog() returned two different repositories")
	}
	if got := db.Sessions(); got != sessions {
		t.Error("Sessions() returned two different repositories")
	}
}

func TestDBStats(t *testing.T) {
	db := newTestDB(t)

	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	stats := db.Stats()
	if stats.MaxOpenConnections <= 0 {
		t.Errorf("MaxOpenConnections = %d, want the configured positive limit", stats.MaxOpenConnections)
	}
	if stats.OpenConnections < 1 {
		t.Errorf("OpenConnections = %d, want at least the connection Ping just used", stats.OpenConnections)
	}
}

// TestDBBackupConveniences covers the Get*/Save* wrappers the backup and
// restore paths call, which delegate to the repositories.
func TestDBBackupConveniences(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	run := &database.TestRun{
		ID:            "run-restored",
		Module:        "benchmark",
		TestType:      "rfc2544_throughput",
		Status:        database.TestRunStatusCompleted,
		InterfaceName: "eth0",
		TargetAddress: "10.44.10.31",
		StartedAt:     time.Now().UTC().Add(-time.Hour).Truncate(time.Second),
	}
	if err := db.SaveTestRun(ctx, run); err != nil {
		t.Fatalf("SaveTestRun failed: %v", err)
	}

	if err := db.CreateTestResult(ctx, &database.TestResult{
		RunID:      run.ID,
		MetricType: database.MetricTypeThroughput,
		Value:      940.5,
		Unit:       "Mbps",
	}); err != nil {
		t.Fatalf("CreateTestResult failed: %v", err)
	}

	if err := db.CreateAuditLog(ctx, &database.AuditLogEntry{
		Action: database.AuditActionTestComplete,
		User:   "operator",
	}); err != nil {
		t.Fatalf("CreateAuditLog failed: %v", err)
	}

	if err := db.BlacklistSession(ctx, &database.Session{
		TokenID:   "jti-restored",
		Username:  "operator",
		Reason:    database.SessionReasonLogout,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}); err != nil {
		t.Fatalf("BlacklistSession failed: %v", err)
	}

	runs, err := db.GetTestRuns(ctx)
	if err != nil {
		t.Fatalf("GetTestRuns failed: %v", err)
	}
	if len(runs) != 1 || runs[0].ID != run.ID {
		t.Fatalf("GetTestRuns = %v, want the single restored run %q", runs, run.ID)
	}
	if runs[0].TargetAddress != "10.44.10.31" {
		t.Errorf("TargetAddress = %q, want %q", runs[0].TargetAddress, "10.44.10.31")
	}
	if !runs[0].StartedAt.Equal(run.StartedAt) {
		t.Errorf("StartedAt = %v, want %v (SaveTestRun must not overwrite a supplied time)",
			runs[0].StartedAt, run.StartedAt)
	}

	results, err := db.GetTestResults(ctx, nil)
	if err != nil {
		t.Fatalf("GetTestResults failed: %v", err)
	}
	if len(results) != 1 || results[0].Value != 940.5 {
		t.Fatalf("GetTestResults = %v, want the single 940.5 Mbps result", results)
	}

	logs, err := db.GetAuditLogs(ctx, 10)
	if err != nil {
		t.Fatalf("GetAuditLogs failed: %v", err)
	}
	if len(logs) != 1 || logs[0].Action != database.AuditActionTestComplete {
		t.Fatalf("GetAuditLogs = %v, want the single test_completed entry", logs)
	}

	sessions, err := db.GetAllBlacklistedSessions(ctx)
	if err != nil {
		t.Fatalf("GetAllBlacklistedSessions failed: %v", err)
	}
	if len(sessions) != 1 || sessions[0].TokenID != "jti-restored" {
		t.Fatalf("GetAllBlacklistedSessions = %v, want the single restored session", sessions)
	}
}

func TestDBWithTxCommits(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	err := db.WithTx(ctx, func(tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx,
			`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)`,
			"committed", "yes", time.Now().UTC().Format(time.RFC3339))
		return execErr
	})
	if err != nil {
		t.Fatalf("WithTx failed: %v", err)
	}

	value, err := db.Settings().Get(ctx, "committed")
	if err != nil {
		t.Fatalf("Get after a committed transaction failed: %v", err)
	}
	if value != "yes" {
		t.Errorf("Get = %q, want %q", value, "yes")
	}
}

func TestDBWithTxRollsBackOnError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	sentinel := errors.New("callback failed")

	err := db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, execErr := tx.ExecContext(ctx,
			`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)`,
			"rolled-back", "no", time.Now().UTC().Format(time.RFC3339)); execErr != nil {
			return execErr
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithTx = %v, want the callback's own error", err)
	}

	if _, err = db.Settings().Get(ctx, "rolled-back"); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("the row written inside a failed transaction survived: Get = %v, want ErrNotFound", err)
	}
}

func TestDBBeginTxRejectsClosedDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := database.Open(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	if rbErr := tx.Rollback(); rbErr != nil {
		t.Fatalf("Rollback failed: %v", rbErr)
	}

	if err = db.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if _, err = db.BeginTx(ctx, nil); err == nil {
		t.Error("BeginTx on a closed database succeeded, want an error")
	}
	if err = db.WithTx(ctx, func(*sql.Tx) error { return nil }); err == nil {
		t.Error("WithTx on a closed database succeeded, want an error")
	}
}

func TestDBMigrationStatusAndSchemaVersion(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	status, err := db.MigrationStatus(ctx)
	if err != nil {
		t.Fatalf("MigrationStatus failed: %v", err)
	}
	if len(status) == 0 {
		t.Fatal("MigrationStatus returned no migrations")
	}

	for i, info := range status {
		if !info.Applied {
			t.Errorf("migration %d (%q) is not applied after Open", info.Version, info.Description)
			continue
		}
		if info.AppliedAt.IsZero() {
			t.Errorf("migration %d has a zero AppliedAt", info.Version)
		}
		if i > 0 && info.Version <= status[i-1].Version {
			t.Errorf("MigrationStatus is not ordered by version: %d after %d",
				info.Version, status[i-1].Version)
		}
	}

	version, err := db.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion failed: %v", err)
	}
	if want := status[len(status)-1].Version; version != want {
		t.Errorf("SchemaVersion = %d, want %d (the highest applied migration)", version, want)
	}

	if err = db.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if _, err = db.MigrationStatus(ctx); err == nil {
		t.Error("MigrationStatus on a closed database succeeded, want an error")
	}
	if _, err = db.SchemaVersion(ctx); err == nil {
		t.Error("SchemaVersion on a closed database succeeded, want an error")
	}
}

func TestOpenWithAutoRebuildOnFreshPath(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "fresh.db")

	db, err := database.OpenWithAutoRebuild(dbPath)
	if err != nil {
		t.Fatalf("OpenWithAutoRebuild failed: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err = db.Settings().Set(context.Background(), "key", "value"); err != nil {
		t.Fatalf("Set on the new database failed: %v", err)
	}
}

func TestOpenWithAutoRebuildBacksUpCorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "corrupt.db")

	// A file that is not a SQLite database at all: the driver reports "file is
	// not a database", which isDatabaseCorrupted must recognise.
	if err := os.WriteFile(dbPath, []byte("this is not a sqlite database"), 0o600); err != nil {
		t.Fatalf("writing the corrupted file failed: %v", err)
	}

	db, err := database.OpenWithAutoRebuild(dbPath)
	if err != nil {
		t.Fatalf("OpenWithAutoRebuild failed on a corrupted file: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	if err = db.Settings().Set(ctx, "rebuilt", "yes"); err != nil {
		t.Fatalf("the rebuilt database is not usable: %v", err)
	}
	value, err := db.Settings().Get(ctx, "rebuilt")
	if err != nil || value != "yes" {
		t.Fatalf("Get on the rebuilt database = (%q, %v), want (\"yes\", nil)", value, err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	var backups int
	for _, entry := range entries {
		if len(entry.Name()) > len("corrupt.db.corrupted.") &&
			entry.Name()[:len("corrupt.db.corrupted.")] == "corrupt.db.corrupted." {
			backups++
		}
	}
	if backups != 1 {
		t.Errorf("found %d backup files, want 1 (the corrupted file must be preserved)", backups)
	}
}

func TestOpenWithAutoRebuildPropagatesNonCorruptionErrors(t *testing.T) {
	if _, err := database.OpenWithAutoRebuild(""); err == nil {
		t.Error("OpenWithAutoRebuild(\"\") succeeded, want the empty-path error")
	}
}
