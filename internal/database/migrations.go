// SPDX-License-Identifier: BUSL-1.1

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Migration represents a database schema migration.
type Migration struct {
	Version     int
	Description string
	Up          string
}

// migrationDef is the definition without version (computed from index).
type migrationDef struct {
	Description string
	Up          string
}

// migrationCount is the number of migrations defined in getMigrationDefs.
// Update this constant when adding new migrations.
const migrationCount = 9

// getMigrationDefs returns migration definitions without versions.
// IMPORTANT: Never modify existing migrations, only add new ones.
// The version is computed as index + 1.
func getMigrationDefs() []migrationDef {
	defs := make([]migrationDef, 0, migrationCount)
	defs = append(defs, migrationSchemaMigrations())
	defs = append(defs, migrationUsers())
	defs = append(defs, migrationSettings())
	defs = append(defs, migrationTestRuns())
	defs = append(defs, migrationTestResults())
	defs = append(defs, migrationAuditLog())
	defs = append(defs, migrationTestSummaries())
	defs = append(defs, migrationProfiles())
	defs = append(defs, migrationSessions())
	return defs
}

func migrationSchemaMigrations() migrationDef {
	return migrationDef{
		Description: "Create schema version table",
		Up: `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				applied_at TEXT NOT NULL,
				description TEXT
			);
		`,
	}
}

func migrationUsers() migrationDef {
	return migrationDef{
		Description: "Create users table for authentication",
		Up: `
			CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT NOT NULL UNIQUE,
				password_hash TEXT NOT NULL,
				role TEXT NOT NULL DEFAULT 'admin',
				is_active INTEGER DEFAULT 1,
				last_login TEXT,
				failed_attempts INTEGER DEFAULT 0,
				locked_until TEXT,
				token_version INTEGER DEFAULT 1,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			);

			CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
			CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
		`,
	}
}

func migrationSettings() migrationDef {
	return migrationDef{
		Description: "Create settings table for key-value settings",
		Up: `
			CREATE TABLE IF NOT EXISTS settings (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				updated_at TEXT NOT NULL
			);
		`,
	}
}

func migrationTestRuns() migrationDef {
	return migrationDef{
		Description: "Create test_runs table for test execution tracking",
		Up: `
			CREATE TABLE IF NOT EXISTS test_runs (
				id TEXT PRIMARY KEY,
				module TEXT NOT NULL,
				test_type TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'pending',
				config_json TEXT,
				interface_name TEXT,
				target_address TEXT,
				started_at TEXT NOT NULL,
				completed_at TEXT,
				duration_ms INTEGER,
				error_message TEXT,
				metadata_json TEXT
			);

			CREATE INDEX IF NOT EXISTS idx_test_runs_module ON test_runs(module);
			CREATE INDEX IF NOT EXISTS idx_test_runs_type ON test_runs(test_type);
			CREATE INDEX IF NOT EXISTS idx_test_runs_status ON test_runs(status);
			CREATE INDEX IF NOT EXISTS idx_test_runs_started ON test_runs(started_at);
		`,
	}
}

func migrationTestResults() migrationDef {
	return migrationDef{
		Description: "Create test_results table for individual test metrics",
		Up: `
			CREATE TABLE IF NOT EXISTS test_results (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				run_id TEXT NOT NULL,
				metric_type TEXT NOT NULL,
				frame_size INTEGER,
				value REAL NOT NULL,
				unit TEXT,
				timestamp TEXT NOT NULL,
				metadata_json TEXT,
				FOREIGN KEY (run_id) REFERENCES test_runs(id) ON DELETE CASCADE
			);

			CREATE INDEX IF NOT EXISTS idx_test_results_run ON test_results(run_id);
			CREATE INDEX IF NOT EXISTS idx_test_results_type ON test_results(metric_type);
			CREATE INDEX IF NOT EXISTS idx_test_results_timestamp ON test_results(timestamp);
			CREATE INDEX IF NOT EXISTS idx_test_results_frame_size ON test_results(frame_size);
		`,
	}
}

func migrationAuditLog() migrationDef {
	return migrationDef{
		Description: "Create audit_log table for activity tracking",
		Up: `
			CREATE TABLE IF NOT EXISTS audit_log (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				action TEXT NOT NULL,
				user TEXT,
				resource_type TEXT,
				resource_id TEXT,
				old_value_json TEXT,
				new_value_json TEXT,
				ip_address TEXT,
				user_agent TEXT,
				timestamp TEXT NOT NULL
			);

			CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log(action);
			CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_log(user);
			CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp);
			CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_log(resource_type, resource_id);
		`,
	}
}

func migrationTestSummaries() migrationDef {
	return migrationDef{
		Description: "Create test_summaries table for aggregated results",
		Up: `
			CREATE TABLE IF NOT EXISTS test_summaries (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				run_id TEXT NOT NULL UNIQUE,
				module TEXT NOT NULL,
				test_type TEXT NOT NULL,
				pass INTEGER NOT NULL DEFAULT 0,
				throughput_mbps REAL,
				latency_avg_us REAL,
				latency_min_us REAL,
				latency_max_us REAL,
				jitter_us REAL,
				frame_loss_pct REAL,
				frames_sent INTEGER,
				frames_received INTEGER,
				created_at TEXT NOT NULL,
				FOREIGN KEY (run_id) REFERENCES test_runs(id) ON DELETE CASCADE
			);

			CREATE INDEX IF NOT EXISTS idx_test_summaries_run ON test_summaries(run_id);
			CREATE INDEX IF NOT EXISTS idx_test_summaries_module ON test_summaries(module);
			CREATE INDEX IF NOT EXISTS idx_test_summaries_pass ON test_summaries(pass);
		`,
	}
}

func migrationProfiles() migrationDef {
	return migrationDef{
		Description: "Create profiles table for test configurations",
		Up: `
			CREATE TABLE IF NOT EXISTS profiles (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL UNIQUE,
				description TEXT,
				module TEXT NOT NULL,
				config_json TEXT NOT NULL,
				is_default INTEGER DEFAULT 0,
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			);

			CREATE INDEX IF NOT EXISTS idx_profiles_name ON profiles(name);
			CREATE INDEX IF NOT EXISTS idx_profiles_module ON profiles(module);
			CREATE INDEX IF NOT EXISTS idx_profiles_default ON profiles(is_default);
		`,
	}
}

func migrationSessions() migrationDef {
	return migrationDef{
		Description: "Create sessions table for token blacklisting",
		Up: `
			CREATE TABLE IF NOT EXISTS sessions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				token_id TEXT NOT NULL UNIQUE,
				username TEXT NOT NULL,
				reason TEXT NOT NULL,
				blacklisted_at TEXT NOT NULL,
				expires_at TEXT NOT NULL
			);

			CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token_id);
			CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(username);
			CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
		`,
	}
}

// getMigrations returns migrations with computed version numbers.
// Version = index + 1 (starting from 1).
func getMigrations() []Migration {
	defs := getMigrationDefs()
	migrations := make([]Migration, len(defs))
	for i, d := range defs {
		migrations[i] = Migration{
			Version:     i + 1,
			Description: d.Description,
			Up:          d.Up,
		}
	}
	return migrations
}

// migrate runs all pending migrations.
func (db *DB) migrate() error {
	ctx := context.Background()

	// Ensure schema_migrations table exists (migration 1)
	_, err := db.conn.ExecContext(ctx, getMigrations()[0].Up)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := db.getCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Run pending migrations
	for _, m := range getMigrations() {
		if m.Version <= currentVersion {
			continue
		}

		if runErr := db.runMigration(ctx, m); runErr != nil {
			return fmt.Errorf(
				"failed to run migration %d (%s): %w",
				m.Version,
				m.Description,
				runErr,
			)
		}
	}

	return nil
}

// getCurrentVersion returns the current schema version.
func (db *DB) getCurrentVersion(ctx context.Context) (int, error) {
	var version int
	err := db.conn.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM schema_migrations
	`).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// runMigration executes a single migration within a transaction.
func (db *DB) runMigration(ctx context.Context, m Migration) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				// Log rollback error but don't override original error
				_ = rbErr // Original error already being returned
			}
		}
	}()

	// Execute migration SQL
	if _, err = tx.ExecContext(ctx, m.Up); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO schema_migrations (version, applied_at, description)
		VALUES (?, ?, ?)
	`, m.Version, now, m.Description)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// MigrationStatus returns the status of all migrations.
func (db *DB) MigrationStatus(ctx context.Context) ([]MigrationInfo, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, errors.New("database is closed")
	}

	// Get applied migrations
	rows, err := db.conn.QueryContext(ctx, `
		SELECT version, applied_at, description FROM schema_migrations ORDER BY version
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[int]time.Time)
	for rows.Next() {
		var version int
		var appliedAt string
		var desc sql.NullString
		if scanErr := rows.Scan(&version, &appliedAt, &desc); scanErr != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", scanErr)
		}
		t, parseErr := time.Parse(time.RFC3339, appliedAt)
		if parseErr != nil {
			// Fallback to current time if stored timestamp is malformed
			t = time.Now().UTC()
		}
		applied[version] = t
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("failed to iterate migration rows: %w", rowsErr)
	}

	// Build status list
	migrations := getMigrations()
	result := make([]MigrationInfo, 0, len(migrations))
	for _, m := range migrations {
		info := MigrationInfo{
			Version:     m.Version,
			Description: m.Description,
			Applied:     false,
		}
		if t, ok := applied[m.Version]; ok {
			info.Applied = true
			info.AppliedAt = t
		}
		result = append(result, info)
	}

	return result, nil
}

// MigrationInfo represents the status of a migration.
type MigrationInfo struct {
	Version     int
	Description string
	Applied     bool
	AppliedAt   time.Time
}

// SchemaVersion returns the current schema version.
func (db *DB) SchemaVersion(ctx context.Context) (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return 0, errors.New("database is closed")
	}

	return db.getCurrentVersion(ctx)
}
