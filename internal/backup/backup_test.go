// SPDX-License-Identifier: BUSL-1.1

package backup_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/backup"
	"github.com/krisarmstrong/stem/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}

	return db, cleanup
}

func seedTestData(t *testing.T, db *database.DB) {
	t.Helper()

	ctx := context.Background()

	// First create a test run to hold results
	run := &database.TestRun{
		Module:   "benchmark",
		TestType: "throughput",
		Status:   database.TestRunStatusCompleted,
	}
	runID, runErr := db.TestRuns().Create(ctx, run)
	if runErr != nil {
		t.Fatalf("Failed to create test run: %v", runErr)
	}

	// Add test results
	for i := range 3 {
		result := &database.TestResult{
			RunID:      runID,
			MetricType: database.MetricTypeThroughput,
			Value:      float64(1000 + i*100),
			Unit:       "Mbps",
		}
		_, saveErr := db.TestResults().Create(ctx, result)
		if saveErr != nil {
			t.Fatalf("Failed to save test result: %v", saveErr)
		}
	}

	// Add audit logs
	for range 2 {
		entry := &database.AuditLogEntry{
			Action:    database.AuditActionLogin,
			User:      "admin",
			IPAddress: "192.168.1.1",
		}
		_, saveErr := db.AuditLog().Log(ctx, entry)
		if saveErr != nil {
			t.Fatalf("Failed to save audit log: %v", saveErr)
		}
	}

	// Add session (use UTC and add significant buffer to avoid timing issues)
	session := &database.Session{
		TokenID:   "test-token-123",
		Username:  "admin",
		Reason:    database.SessionReasonLogout,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	_, saveErr := db.Sessions().Blacklist(ctx, session)
	if saveErr != nil {
		t.Fatalf("Failed to save session: %v", saveErr)
	}
}

func TestExport(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestData(t, db)

	ctx := context.Background()
	outputPath := filepath.Join(t.TempDir(), "backup.json")

	opts := backup.DefaultExportOptions()
	opts.Description = "Test backup"
	opts.CreatedBy = "test"
	opts.StemVersion = "v0.2.3"

	recordCount, err := backup.Export(ctx, db, outputPath, opts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Should have 1 test run + 3 test results + 2 audit logs + 1 session = 7 records
	expectedCount := 7
	if recordCount != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, recordCount)
	}

	// Verify file exists
	_, statErr := os.Stat(outputPath)
	if statErr != nil {
		t.Errorf("Backup file not created: %v", statErr)
	}

	// Verify file content is valid JSON
	data, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatalf("Failed to read backup file: %v", readErr)
	}

	var backupData backup.Data
	unmarshalErr := json.Unmarshal(data, &backupData)
	if unmarshalErr != nil {
		t.Fatalf("Backup file is not valid JSON: %v", unmarshalErr)
	}

	if backupData.Metadata.Version != backup.BackupVersion {
		t.Errorf("Expected version %s, got %s", backup.BackupVersion, backupData.Metadata.Version)
	}

	if backupData.Metadata.Description != "Test backup" {
		t.Errorf("Expected description 'Test backup', got %s", backupData.Metadata.Description)
	}

	if len(backupData.TestResults) != 3 {
		t.Errorf("Expected 3 test results, got %d", len(backupData.TestResults))
	}

	if len(backupData.AuditLogs) != 2 {
		t.Errorf("Expected 2 audit logs, got %d", len(backupData.AuditLogs))
	}
}

func TestExportNilOptions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestData(t, db)

	ctx := context.Background()
	outputPath := filepath.Join(t.TempDir(), "backup.json")

	// Test with nil options (should use defaults)
	recordCount, err := backup.Export(ctx, db, outputPath, nil)
	if err != nil {
		t.Fatalf("Export with nil options failed: %v", err)
	}

	if recordCount != 7 {
		t.Errorf("Expected 7 records, got %d", recordCount)
	}
}

func TestExportToWriter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestData(t, db)

	ctx := context.Background()
	var buf bytes.Buffer

	recordCount, err := backup.ExportToWriter(ctx, db, &buf, nil)
	if err != nil {
		t.Fatalf("ExportToWriter failed: %v", err)
	}

	if recordCount != 7 {
		t.Errorf("Expected 7 records, got %d", recordCount)
	}

	// Verify content is valid JSON
	var backupData backup.Data
	unmarshalErr := json.Unmarshal(buf.Bytes(), &backupData)
	if unmarshalErr != nil {
		t.Fatalf("Output is not valid JSON: %v", unmarshalErr)
	}
}

func TestImport(t *testing.T) {
	// Create source database with data
	sourceDB, sourceCleanup := setupTestDB(t)
	seedTestData(t, sourceDB)

	ctx := context.Background()
	backupPath := filepath.Join(t.TempDir(), "backup.json")

	// Export from source
	_, exportErr := backup.Export(ctx, sourceDB, backupPath, nil)
	if exportErr != nil {
		sourceCleanup()
		t.Fatalf("Export failed: %v", exportErr)
	}
	sourceCleanup()

	// Create target database (empty)
	targetDB, targetCleanup := setupTestDB(t)
	defer targetCleanup()

	// Import to target
	importCount, importErr := backup.Import(ctx, targetDB, backupPath, nil)
	if importErr != nil {
		t.Fatalf("Import failed: %v", importErr)
	}

	if importCount != 7 {
		t.Errorf("Expected 7 imported records, got %d", importCount)
	}

	// Verify data was imported
	results, resultsErr := targetDB.GetTestResults(ctx, nil)
	if resultsErr != nil {
		t.Fatalf("Failed to get test results: %v", resultsErr)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 test results after import, got %d", len(results))
	}

	logs, logsErr := targetDB.GetAuditLogs(ctx, 0)
	if logsErr != nil {
		t.Fatalf("Failed to get audit logs: %v", logsErr)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 audit logs after import, got %d", len(logs))
	}
}

func TestImportWithSkipOptions(t *testing.T) {
	sourceDB, sourceCleanup := setupTestDB(t)
	seedTestData(t, sourceDB)

	ctx := context.Background()
	backupPath := filepath.Join(t.TempDir(), "backup.json")

	_, exportErr := backup.Export(ctx, sourceDB, backupPath, nil)
	if exportErr != nil {
		sourceCleanup()
		t.Fatalf("Export failed: %v", exportErr)
	}
	sourceCleanup()

	targetDB, targetCleanup := setupTestDB(t)
	defer targetCleanup()

	// Import with skip options
	opts := &backup.ImportOptions{
		SkipTestResults: true,
		SkipAuditLogs:   false,
		SkipSessions:    true,
		ClearExisting:   false,
	}

	importCount, importErr := backup.Import(ctx, targetDB, backupPath, opts)
	if importErr != nil {
		t.Fatalf("Import failed: %v", importErr)
	}

	// Should only import audit logs (2)
	if importCount != 2 {
		t.Errorf("Expected 2 imported records (audit logs only), got %d", importCount)
	}

	// Verify only audit logs were imported
	results, resultsErr := targetDB.GetTestResults(ctx, nil)
	if resultsErr != nil {
		t.Fatalf("Failed to get test results: %v", resultsErr)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 test results (skipped), got %d", len(results))
	}
}

func TestImportFromReader(t *testing.T) {
	sourceDB, sourceCleanup := setupTestDB(t)
	seedTestData(t, sourceDB)

	ctx := context.Background()
	var buf bytes.Buffer

	_, exportErr := backup.ExportToWriter(ctx, sourceDB, &buf, nil)
	if exportErr != nil {
		sourceCleanup()
		t.Fatalf("Export failed: %v", exportErr)
	}
	sourceCleanup()

	targetDB, targetCleanup := setupTestDB(t)
	defer targetCleanup()

	reader := bytes.NewReader(buf.Bytes())
	importCount, importErr := backup.ImportFromReader(ctx, targetDB, reader, nil)
	if importErr != nil {
		t.Fatalf("ImportFromReader failed: %v", importErr)
	}

	if importCount != 7 {
		t.Errorf("Expected 7 imported records, got %d", importCount)
	}
}

func TestImportInvalidFile(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test non-existent file
	_, err := backup.Import(ctx, db, "/nonexistent/path/backup.json", nil)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test invalid JSON
	invalidPath := filepath.Join(t.TempDir(), "invalid.json")
	writeErr := os.WriteFile(invalidPath, []byte("not json"), 0o600)
	if writeErr != nil {
		t.Fatalf("Failed to write invalid file: %v", writeErr)
	}

	_, err = backup.Import(ctx, db, invalidPath, nil)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestImportVersionMismatch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create backup with wrong version
	wrongVersion := backup.Data{
		Metadata: backup.Metadata{
			Version:     "99.0", // Wrong version
			CreatedAt:   time.Now(),
			CreatedBy:   "",
			Description: "",
			StemVersion: "",
		},
		TestResults: nil,
		AuditLogs:   nil,
		Sessions:    nil,
		LicenseInfo: nil,
	}

	data, marshalErr := json.Marshal(wrongVersion)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal: %v", marshalErr)
	}

	backupPath := filepath.Join(t.TempDir(), "backup.json")
	writeErr := os.WriteFile(backupPath, data, 0o600)
	if writeErr != nil {
		t.Fatalf("Failed to write backup: %v", writeErr)
	}

	_, err := backup.Import(ctx, db, backupPath, nil)
	if err == nil {
		t.Error("Expected error for version mismatch")
	}
}

func TestGenerateBackupFilename(t *testing.T) {
	// Test with default prefix
	filename := backup.GenerateBackupFilename("")
	if filename == "" {
		t.Error("Expected non-empty filename")
	}
	if len(filename) < 20 { // stem-backup-YYYYMMDD-HHMMSS.json
		t.Errorf("Filename too short: %s", filename)
	}

	// Test with custom prefix
	filename = backup.GenerateBackupFilename("my-backup")
	if filename == "" {
		t.Error("Expected non-empty filename")
	}
	if len(filename) < 15 {
		t.Errorf("Filename too short: %s", filename)
	}
}

func TestValidateBackup(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestData(t, db)

	ctx := context.Background()
	backupPath := filepath.Join(t.TempDir(), "backup.json")

	opts := backup.DefaultExportOptions()
	opts.Description = "Validation test"

	_, exportErr := backup.Export(ctx, db, backupPath, opts)
	if exportErr != nil {
		t.Fatalf("Export failed: %v", exportErr)
	}

	// Validate the backup
	metadata, err := backup.ValidateBackup(backupPath)
	if err != nil {
		t.Fatalf("ValidateBackup failed: %v", err)
	}

	if metadata.Version != backup.BackupVersion {
		t.Errorf("Expected version %s, got %s", backup.BackupVersion, metadata.Version)
	}

	if metadata.Description != "Validation test" {
		t.Errorf("Expected description 'Validation test', got %s", metadata.Description)
	}

	// Test invalid file
	_, err = backup.ValidateBackup("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestGetBackupStats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestData(t, db)

	ctx := context.Background()
	backupPath := filepath.Join(t.TempDir(), "backup.json")

	_, exportErr := backup.Export(ctx, db, backupPath, nil)
	if exportErr != nil {
		t.Fatalf("Export failed: %v", exportErr)
	}

	stats, err := backup.GetBackupStats(backupPath)
	if err != nil {
		t.Fatalf("GetBackupStats failed: %v", err)
	}

	if stats.TestResultCount != 3 {
		t.Errorf("Expected 3 test results, got %d", stats.TestResultCount)
	}

	if stats.AuditLogCount != 2 {
		t.Errorf("Expected 2 audit logs, got %d", stats.AuditLogCount)
	}

	if stats.SessionCount != 1 {
		t.Errorf("Expected 1 session, got %d", stats.SessionCount)
	}

	if stats.FileSizeBytes <= 0 {
		t.Error("Expected positive file size")
	}

	// Test non-existent file
	_, err = backup.GetBackupStats("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestDefaultExportOptions(t *testing.T) {
	opts := backup.DefaultExportOptions()

	if !opts.IncludeTestResults {
		t.Error("Expected IncludeTestResults to be true")
	}

	if !opts.IncludeAuditLogs {
		t.Error("Expected IncludeAuditLogs to be true")
	}

	if !opts.IncludeSessions {
		t.Error("Expected IncludeSessions to be true")
	}

	if !opts.IncludeLicenseInfo {
		t.Error("Expected IncludeLicenseInfo to be true")
	}
}

func TestDefaultImportOptions(t *testing.T) {
	opts := backup.DefaultImportOptions()

	if opts.SkipTestResults {
		t.Error("Expected SkipTestResults to be false")
	}

	if opts.SkipAuditLogs {
		t.Error("Expected SkipAuditLogs to be false")
	}

	if opts.SkipSessions {
		t.Error("Expected SkipSessions to be false")
	}

	if opts.ClearExisting {
		t.Error("Expected ClearExisting to be false")
	}
}

func TestExportEmptyDatabase(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	outputPath := filepath.Join(t.TempDir(), "backup.json")

	// Export empty database
	recordCount, err := backup.Export(ctx, db, outputPath, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if recordCount != 0 {
		t.Errorf("Expected 0 records from empty database, got %d", recordCount)
	}

	// Verify file was still created
	_, statErr := os.Stat(outputPath)
	if statErr != nil {
		t.Error("Expected backup file to be created for empty database")
	}
}

func TestImportExpiredSessions(t *testing.T) {
	ctx := context.Background()

	// Create backup with expired session
	backupData := backup.Data{
		Metadata: backup.Metadata{
			Version:     backup.BackupVersion,
			CreatedAt:   time.Now(),
			CreatedBy:   "",
			Description: "",
			StemVersion: "",
		},
		TestResults: nil,
		AuditLogs:   nil,
		Sessions: []database.Session{
			{
				TokenID:       "expired-token",
				Username:      "admin",
				Reason:        database.SessionReasonLogout,
				BlacklistedAt: time.Now().Add(-2 * time.Hour),
				ExpiresAt:     time.Now().Add(-time.Hour), // Already expired
			},
		},
		LicenseInfo: nil,
	}

	data, marshalErr := json.Marshal(backupData)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal: %v", marshalErr)
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	reader := bytes.NewReader(data)
	importCount, importErr := backup.ImportFromReader(ctx, db, reader, nil)
	if importErr != nil {
		t.Fatalf("Import failed: %v", importErr)
	}

	// Expired session should be skipped
	if importCount != 0 {
		t.Errorf("Expected 0 imported records (expired session skipped), got %d", importCount)
	}
}
