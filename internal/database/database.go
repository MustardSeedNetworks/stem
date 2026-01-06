// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package database provides SQLite database persistence for The Stem.
//
// This package handles:
//   - Database connection management
//   - Schema migrations
//   - CRUD operations for test results, audit logs, and sessions
//
// The database stores test results for historical analysis, audit logs for
// security compliance, and session data for token blacklist persistence.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	// SQLite driver.
	_ "github.com/mattn/go-sqlite3"

	"github.com/krisarmstrong/stem/internal/logging"
)

var (
	// ErrDatabaseClosed indicates an operation was attempted on a closed database.
	ErrDatabaseClosed = errors.New("database is closed")
	// ErrMigrationFailed indicates a migration could not be applied.
	ErrMigrationFailed = errors.New("migration failed")
	// ErrRecordNotFound indicates the requested record does not exist.
	ErrRecordNotFound = errors.New("record not found")
)

// Database wraps a SQLite connection with thread-safe access.
type Database struct {
	db     *sql.DB
	mu     sync.RWMutex
	closed bool
	path   string
}

// NewDatabase opens or creates a SQLite database at the given path.
// The parent directory is created if it does not exist.
// Returns an initialized Database ready for use.
func NewDatabase(path string) (*Database, error) {
	// Ensure parent directory exists.
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		mkdirErr := os.MkdirAll(dir, 0o750)
		if mkdirErr != nil {
			return nil, fmt.Errorf("create database directory: %w", mkdirErr)
		}
	}

	// Open SQLite database with recommended pragmas for performance and safety.
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000&_foreign_keys=ON", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Verify connection.
	pingErr := db.PingContext(context.Background())
	if pingErr != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", pingErr)
	}

	// Configure connection pool for embedded use.
	db.SetMaxOpenConns(1) // SQLite only supports one writer.
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	logging.Info("Database opened", "path", path)

	d := &Database{
		db:     db,
		mu:     sync.RWMutex{},
		closed: false,
		path:   path,
	}

	return d, nil
}

// Close closes the database connection.
// After Close, all operations will return ErrDatabaseClosed.
func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true
	closeErr := d.db.Close()
	if closeErr != nil {
		return fmt.Errorf("close database: %w", closeErr)
	}

	logging.Info("Database closed", "path", d.path)
	return nil
}

// RunMigrations applies all pending database migrations.
// Migrations are idempotent and can be run multiple times safely.
func (d *Database) RunMigrations() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrDatabaseClosed
	}

	logging.Info("Running database migrations")

	for i, migration := range migrations {
		_, execErr := d.db.ExecContext(context.Background(), migration)
		if execErr != nil {
			return fmt.Errorf("%w: migration %d: %w", ErrMigrationFailed, i+1, execErr)
		}
	}

	logging.Info("Database migrations completed", "count", len(migrations))
	return nil
}

// Path returns the filesystem path to the database file.
func (d *Database) Path() string {
	return d.path
}

// DB returns the underlying sql.DB for advanced operations.
// Use with caution; prefer the typed methods when possible.
func (d *Database) DB() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.db
}

// ensureOpen returns an error if the database is closed.
func (d *Database) ensureOpen() error {
	if d.closed {
		return ErrDatabaseClosed
	}
	return nil
}
