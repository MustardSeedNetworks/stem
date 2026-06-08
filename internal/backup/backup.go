// SPDX-License-Identifier: BUSL-1.1

// Package backup provides database backup and restore functionality for The Stem.
//
// This package handles:
//   - Exporting test results, audit logs, and sessions to JSON
//   - Importing backup data to restore a system
//   - Creating timestamped backup archives
//
// Note: License data is device-bound and cannot be backed up/restored across devices.
package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/database"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

var (
	// ErrInvalidBackup indicates the backup file format is invalid.
	ErrInvalidBackup = errors.New("invalid backup file")
	// ErrBackupVersionMismatch indicates the backup version is incompatible.
	ErrBackupVersionMismatch = errors.New("backup version mismatch")
	// ErrBackupFailed indicates the backup operation failed.
	ErrBackupFailed = errors.New("backup failed")
	// ErrRestoreFailed indicates the restore operation failed.
	ErrRestoreFailed = errors.New("restore failed")
)

// BackupVersion is the current backup format version.
const BackupVersion = "1.0"

// Metadata contains information about the backup.
type Metadata struct {
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy,omitempty"`
	Description string    `json:"description,omitempty"`
	StemVersion string    `json:"stemVersion,omitempty"`
}

// Data contains all data that can be backed up.
type Data struct {
	Metadata    Metadata                 `json:"metadata"`
	TestRuns    []database.TestRun       `json:"testRuns,omitempty"`
	TestResults []database.TestResult    `json:"testResults,omitempty"`
	AuditLogs   []database.AuditLogEntry `json:"auditLogs,omitempty"`
	Sessions    []database.Session       `json:"sessions,omitempty"`
	LicenseInfo *LicenseBackupInfo       `json:"licenseInfo,omitempty"`
}

// LicenseBackupInfo contains non-sensitive license information for reference.
// Note: The actual license is device-bound and cannot be transferred.
type LicenseBackupInfo struct {
	Tier            string    `json:"tier,omitempty"`
	ActivatedAt     time.Time `json:"activatedAt,omitzero"`
	ExpiresAt       time.Time `json:"expiresAt,omitzero"`
	IsTrialMode     bool      `json:"isTrialMode"`
	TrialDaysUsed   int       `json:"trialDaysUsed,omitempty"`
	DeviceReference string    `json:"deviceReference,omitempty"` // Last 8 chars of device hash
}

// ExportOptions configures what to include in the backup.
type ExportOptions struct {
	IncludeTestResults bool
	IncludeAuditLogs   bool
	IncludeSessions    bool
	IncludeLicenseInfo bool
	Description        string
	CreatedBy          string
	StemVersion        string
}

// DefaultExportOptions returns options that include everything.
func DefaultExportOptions() *ExportOptions {
	return &ExportOptions{
		IncludeTestResults: true,
		IncludeAuditLogs:   true,
		IncludeSessions:    true,
		IncludeLicenseInfo: true,
		Description:        "",
		CreatedBy:          "",
		StemVersion:        "",
	}
}

// Export creates a backup of the database to the specified file.
// Returns the number of records exported.
func Export(ctx context.Context, db *database.Database, outputPath string, opts *ExportOptions) (int, error) {
	if opts == nil {
		opts = DefaultExportOptions()
	}

	logging.Info("Starting backup export", "path", outputPath)

	backup := Data{
		Metadata: Metadata{
			Version:     BackupVersion,
			CreatedAt:   time.Now().UTC(),
			CreatedBy:   opts.CreatedBy,
			Description: opts.Description,
			StemVersion: opts.StemVersion,
		},
		TestResults: nil,
		AuditLogs:   nil,
		Sessions:    nil,
		LicenseInfo: nil,
	}

	recordCount := 0

	// Export test runs and results (runs must be included for results to work)
	if opts.IncludeTestResults {
		runs, runsErr := db.GetTestRuns(ctx)
		if runsErr != nil {
			return 0, fmt.Errorf("%w: failed to get test runs: %w", ErrBackupFailed, runsErr)
		}
		backup.TestRuns = runs
		recordCount += len(runs)
		logging.Debug("Exported test runs", "count", len(runs))

		results, resultsErr := db.GetTestResults(ctx, nil) // Get all results
		if resultsErr != nil {
			return 0, fmt.Errorf("%w: failed to get test results: %w", ErrBackupFailed, resultsErr)
		}
		backup.TestResults = results
		recordCount += len(results)
		logging.Debug("Exported test results", "count", len(results))
	}

	// Export audit logs
	if opts.IncludeAuditLogs {
		logs, logsErr := db.GetAuditLogs(ctx, 0) // Get all logs (no limit)
		if logsErr != nil {
			return 0, fmt.Errorf("%w: failed to get audit logs: %w", ErrBackupFailed, logsErr)
		}
		backup.AuditLogs = logs
		recordCount += len(logs)
		logging.Debug("Exported audit logs", "count", len(logs))
	}

	// Export sessions (blacklisted tokens)
	if opts.IncludeSessions {
		sessions, sessionsErr := db.GetAllBlacklistedSessions(ctx)
		if sessionsErr != nil {
			return 0, fmt.Errorf("%w: failed to get sessions: %w", ErrBackupFailed, sessionsErr)
		}
		backup.Sessions = sessions
		recordCount += len(sessions)
		logging.Debug("Exported sessions", "count", len(sessions))
	}

	// Serialize to JSON with indentation for readability
	jsonData, marshalErr := json.MarshalIndent(backup, "", "  ")
	if marshalErr != nil {
		return 0, fmt.Errorf("%w: failed to marshal backup: %w", ErrBackupFailed, marshalErr)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		mkdirErr := os.MkdirAll(dir, 0o750)
		if mkdirErr != nil {
			return 0, fmt.Errorf("%w: failed to create directory: %w", ErrBackupFailed, mkdirErr)
		}
	}

	// Write to file
	writeErr := os.WriteFile(outputPath, jsonData, 0o600)
	if writeErr != nil {
		return 0, fmt.Errorf("%w: failed to write file: %w", ErrBackupFailed, writeErr)
	}

	logging.Info("Backup export completed", "path", outputPath, "records", recordCount)
	return recordCount, nil
}

// ExportToWriter writes backup data to an [io.Writer].
func ExportToWriter(ctx context.Context, db *database.Database, w io.Writer, opts *ExportOptions) (int, error) {
	if opts == nil {
		opts = DefaultExportOptions()
	}

	logging.Info("Starting backup export to writer")

	backup := Data{
		Metadata: Metadata{
			Version:     BackupVersion,
			CreatedAt:   time.Now().UTC(),
			CreatedBy:   opts.CreatedBy,
			Description: opts.Description,
			StemVersion: opts.StemVersion,
		},
		TestResults: nil,
		AuditLogs:   nil,
		Sessions:    nil,
		LicenseInfo: nil,
	}

	recordCount := 0

	// Export test runs and results (runs must be included for results to work)
	if opts.IncludeTestResults {
		runs, runsErr := db.GetTestRuns(ctx)
		if runsErr != nil {
			return 0, fmt.Errorf("%w: failed to get test runs: %w", ErrBackupFailed, runsErr)
		}
		backup.TestRuns = runs
		recordCount += len(runs)

		results, resultsErr := db.GetTestResults(ctx, nil)
		if resultsErr != nil {
			return 0, fmt.Errorf("%w: failed to get test results: %w", ErrBackupFailed, resultsErr)
		}
		backup.TestResults = results
		recordCount += len(results)
	}

	// Export audit logs
	if opts.IncludeAuditLogs {
		logs, logsErr := db.GetAuditLogs(ctx, 0)
		if logsErr != nil {
			return 0, fmt.Errorf("%w: failed to get audit logs: %w", ErrBackupFailed, logsErr)
		}
		backup.AuditLogs = logs
		recordCount += len(logs)
	}

	// Export sessions
	if opts.IncludeSessions {
		sessions, sessionsErr := db.GetAllBlacklistedSessions(ctx)
		if sessionsErr != nil {
			return 0, fmt.Errorf("%w: failed to get sessions: %w", ErrBackupFailed, sessionsErr)
		}
		backup.Sessions = sessions
		recordCount += len(sessions)
	}

	// Encode to writer
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encodeErr := encoder.Encode(backup)
	if encodeErr != nil {
		return 0, fmt.Errorf("%w: failed to encode backup: %w", ErrBackupFailed, encodeErr)
	}

	logging.Info("Backup export to writer completed", "records", recordCount)
	return recordCount, nil
}

// ImportOptions configures restore behavior.
type ImportOptions struct {
	SkipTestResults bool // Skip importing test results
	SkipAuditLogs   bool // Skip importing audit logs
	SkipSessions    bool // Skip importing sessions
	ClearExisting   bool // Clear existing data before import
}

// DefaultImportOptions returns default import options.
func DefaultImportOptions() *ImportOptions {
	return &ImportOptions{
		SkipTestResults: false,
		SkipAuditLogs:   false,
		SkipSessions:    false,
		ClearExisting:   false,
	}
}

// Import restores data from a backup file.
// Returns the number of records imported.
func Import(ctx context.Context, db *database.Database, inputPath string, opts *ImportOptions) (int, error) {
	if opts == nil {
		opts = DefaultImportOptions()
	}

	// Sanitize file path to prevent directory traversal.
	cleanPath := filepath.Clean(inputPath)

	logging.Info("Starting backup import", "path", cleanPath)

	// Read file
	data, readErr := os.ReadFile(cleanPath)
	if readErr != nil {
		return 0, fmt.Errorf("%w: failed to read file: %w", ErrRestoreFailed, readErr)
	}

	return importFromData(ctx, db, data, opts)
}

// ImportFromReader restores data from an [io.Reader].
func ImportFromReader(ctx context.Context, db *database.Database, r io.Reader, opts *ImportOptions) (int, error) {
	if opts == nil {
		opts = DefaultImportOptions()
	}

	logging.Info("Starting backup import from reader")

	data, readErr := io.ReadAll(r)
	if readErr != nil {
		return 0, fmt.Errorf("%w: failed to read data: %w", ErrRestoreFailed, readErr)
	}

	return importFromData(ctx, db, data, opts)
}

func importFromData(ctx context.Context, db *database.Database, data []byte, opts *ImportOptions) (int, error) {
	backup, parseErr := parseBackup(data)
	if parseErr != nil {
		return 0, parseErr
	}

	recordCount := 0
	// Import test runs first (test results have foreign key to runs)
	recordCount += importTestRuns(ctx, db, backup.TestRuns, opts.SkipTestResults)
	recordCount += importTestResults(ctx, db, backup.TestResults, opts.SkipTestResults)
	recordCount += importAuditLogs(ctx, db, backup.AuditLogs, opts.SkipAuditLogs)
	recordCount += importSessions(ctx, db, backup.Sessions, opts.SkipSessions)

	logging.Info("Backup import completed", "records", recordCount)
	return recordCount, nil
}

func parseBackup(data []byte) (*Data, error) {
	var backup Data
	unmarshalErr := json.Unmarshal(data, &backup)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("%w: failed to parse backup: %w", ErrInvalidBackup, unmarshalErr)
	}

	if backup.Metadata.Version != BackupVersion {
		return nil, fmt.Errorf("%w: expected %s, got %s",
			ErrBackupVersionMismatch, BackupVersion, backup.Metadata.Version)
	}

	return &backup, nil
}

func importTestRuns(ctx context.Context, db *database.Database, runs []database.TestRun, skip bool) int {
	if skip || len(runs) == 0 {
		return 0
	}

	count := 0
	for i := range runs {
		run := &runs[i]
		saveErr := db.SaveTestRun(ctx, run)
		if saveErr != nil {
			logging.Warn("Failed to import test run", "error", saveErr)
			continue
		}
		count++
	}
	logging.Debug("Imported test runs", "count", len(runs))
	return count
}

func importTestResults(ctx context.Context, db *database.Database, results []database.TestResult, skip bool) int {
	if skip || len(results) == 0 {
		return 0
	}

	count := 0
	for i := range results {
		result := &results[i]
		result.ID = 0
		saveErr := db.SaveTestResult(ctx, result)
		if saveErr != nil {
			logging.Warn("Failed to import test result", "error", saveErr)
			continue
		}
		count++
	}
	logging.Debug("Imported test results", "count", len(results))
	return count
}

func importAuditLogs(ctx context.Context, db *database.Database, logs []database.AuditLogEntry, skip bool) int {
	if skip || len(logs) == 0 {
		return 0
	}

	count := 0
	for i := range logs {
		entry := &logs[i]
		entry.ID = 0
		saveErr := db.SaveAuditLog(ctx, entry)
		if saveErr != nil {
			logging.Warn("Failed to import audit log", "error", saveErr)
			continue
		}
		count++
	}
	logging.Debug("Imported audit logs", "count", len(logs))
	return count
}

func importSessions(ctx context.Context, db *database.Database, sessions []database.Session, skip bool) int {
	if skip || len(sessions) == 0 {
		return 0
	}

	count := 0
	now := time.Now()
	for i := range sessions {
		session := &sessions[i]
		if !session.ExpiresAt.After(now) {
			continue
		}
		saveErr := db.SaveSession(ctx, session)
		if saveErr != nil {
			logging.Warn("Failed to import session", "error", saveErr)
			continue
		}
		count++
	}
	logging.Debug("Imported sessions", "count", len(sessions))
	return count
}

// GenerateBackupFilename creates a timestamped backup filename.
func GenerateBackupFilename(prefix string) string {
	if prefix == "" {
		prefix = "stem-backup"
	}
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s.json", prefix, timestamp)
}

// ValidateBackup checks if a backup file is valid without importing.
func ValidateBackup(inputPath string) (*Metadata, error) {
	// Sanitize file path to prevent directory traversal.
	cleanPath := filepath.Clean(inputPath)

	data, readErr := os.ReadFile(cleanPath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read file: %w", readErr)
	}

	var backup Data
	unmarshalErr := json.Unmarshal(data, &backup)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("%w: failed to parse backup: %w", ErrInvalidBackup, unmarshalErr)
	}

	if backup.Metadata.Version == "" {
		return nil, ErrInvalidBackup
	}

	return &backup.Metadata, nil
}

// Stats contains statistics about a backup file.
type Stats struct {
	Metadata        Metadata
	TestResultCount int
	AuditLogCount   int
	SessionCount    int
	FileSizeBytes   int64
}

// GetBackupStats returns statistics about a backup file.
func GetBackupStats(inputPath string) (*Stats, error) {
	// Sanitize file path to prevent directory traversal.
	cleanPath := filepath.Clean(inputPath)

	// Get file size
	fileInfo, statErr := os.Stat(cleanPath)
	if statErr != nil {
		return nil, fmt.Errorf("failed to stat file: %w", statErr)
	}

	data, readErr := os.ReadFile(cleanPath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read file: %w", readErr)
	}

	var backup Data
	unmarshalErr := json.Unmarshal(data, &backup)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("%w: failed to parse backup: %w", ErrInvalidBackup, unmarshalErr)
	}

	return &Stats{
		Metadata:        backup.Metadata,
		TestResultCount: len(backup.TestResults),
		AuditLogCount:   len(backup.AuditLogs),
		SessionCount:    len(backup.Sessions),
		FileSizeBytes:   fileInfo.Size(),
	}, nil
}
