// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package database

// migrations contains all database schema migrations.
// Migrations are applied in order and should be idempotent (use IF NOT EXISTS).
// Each migration is a separate SQL statement for atomic application.
//
// Migration naming convention:
//   - Use descriptive names in comments
//   - Keep migrations small and focused
//   - Never modify existing migrations; always add new ones
//
//nolint:gochecknoglobals // Migrations are intentionally global for package-level access.
var migrations = []string{
	// Migration 1: Create test_results table
	// Stores completed test results with JSON-encoded result data.
	`CREATE TABLE IF NOT EXISTS test_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		test_type TEXT NOT NULL,
		module TEXT NOT NULL,
		status TEXT NOT NULL,
		result_json TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Migration 2: Create index on test_results for common queries
	`CREATE INDEX IF NOT EXISTS idx_test_results_created_at ON test_results(created_at DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_test_results_module ON test_results(module)`,
	`CREATE INDEX IF NOT EXISTS idx_test_results_test_type ON test_results(test_type)`,

	// Migration 3: Create audit_log table
	// Tracks security-relevant events for compliance and debugging.
	`CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL,
		user TEXT,
		details TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Migration 4: Create index on audit_log for time-based queries
	`CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_audit_log_event_type ON audit_log(event_type)`,
	`CREATE INDEX IF NOT EXISTS idx_audit_log_user ON audit_log(user)`,

	// Migration 5: Create sessions table for token blacklist persistence
	// Stores revoked tokens until they naturally expire.
	`CREATE TABLE IF NOT EXISTS sessions (
		token_id TEXT PRIMARY KEY,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Migration 6: Create index on sessions for cleanup queries
	`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`,
}
