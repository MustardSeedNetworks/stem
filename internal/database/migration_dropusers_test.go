// SPDX-License-Identifier: BUSL-1.1

package database_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	// Registers the sqlite driver used to build the pre-upgrade fixture below.
	_ "modernc.org/sqlite"

	"github.com/MustardSeedNetworks/stem/internal/database"
)

// tableExists reports whether a table is present in the schema.
func tableExists(t *testing.T, db *database.DB, name string) bool {
	t.Helper()
	var found string
	err := db.QueryRow(context.Background(),
		`SELECT name FROM sqlite_master WHERE type='table' AND name = ?`, name).Scan(&found)
	switch {
	case err == nil:
		return true
	case errors.Is(err, sql.ErrNoRows):
		return false
	default:
		t.Fatalf("querying sqlite_master for %q: %v", name, err)
		return false
	}
}

// A fresh install must not end up with the table: migration 2 creates it and
// migration 10 drops it, and both run during the same Open.
func TestFreshDatabaseHasNoUsersTable(t *testing.T) {
	db := newTestDB(t)

	if tableExists(t, db, "users") {
		t.Error("users table present after a fresh migration run")
	}
	// A table the drop should not have touched, so this is not passing because
	// nothing was created at all.
	if !tableExists(t, db, "sessions") {
		t.Error("sessions table missing — the migration run did not complete")
	}
}

// The upgrade path is the one that matters: an existing database already has
// the table, with rows in it. Build a database that stops at version 9, then
// reopen it through the real migrator and require that migration 10 lands.
func TestUpgradeFromV9DropsPopulatedUsersTable(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "upgrade.db")

	// Build the v9 fixture with the raw driver so the migrator plays no part.
	raw, openErr := sql.Open("sqlite", dbPath)
	if openErr != nil {
		t.Fatalf("opening fixture: %v", openErr)
	}
	for _, stmt := range []string{
		`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL, description TEXT);`,
		`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'admin', is_active INTEGER DEFAULT 1,
			last_login TEXT, failed_attempts INTEGER DEFAULT 0, locked_until TEXT,
			token_version INTEGER DEFAULT 1, created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL);`,
		`CREATE INDEX idx_users_username ON users(username);`,
		`CREATE INDEX idx_users_active ON users(is_active);`,
		`INSERT INTO users (username, password_hash, created_at, updated_at)
			VALUES ('legacy-admin', 'not-a-real-hash', '2026-01-01', '2026-01-01');`,
	} {
		if _, execErr := raw.ExecContext(ctx, stmt); execErr != nil {
			t.Fatalf("building fixture (%.40s...): %v", stmt, execErr)
		}
	}
	for v := 1; v <= 9; v++ {
		if _, insErr := raw.ExecContext(ctx,
			`INSERT INTO schema_migrations (version, applied_at, description)
			 VALUES (?, '2026-01-01', 'fixture')`, v); insErr != nil {
			t.Fatalf("recording fixture version %d: %v", v, insErr)
		}
	}
	if closeErr := raw.Close(); closeErr != nil {
		t.Fatalf("closing fixture: %v", closeErr)
	}

	// Reopen through the real migrator.
	db, dbErr := database.Open(dbPath)
	if dbErr != nil {
		t.Fatalf("Open on the v9 fixture failed: %v", dbErr)
	}
	t.Cleanup(func() { _ = db.Close() })

	if tableExists(t, db, "users") {
		t.Error("users table survived the upgrade from version 9")
	}

	var version int
	if scanErr := db.QueryRow(ctx,
		`SELECT MAX(version) FROM schema_migrations`).Scan(&version); scanErr != nil {
		t.Fatalf("reading schema version: %v", scanErr)
	}
	if version != 10 {
		t.Errorf("schema version = %d, want 10", version)
	}
}
