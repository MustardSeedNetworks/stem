// SPDX-License-Identifier: BUSL-1.1

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SettingsRepository handles settings persistence.
type SettingsRepository struct {
	db *DB
}

// Get retrieves a setting by key.
func (r *SettingsRepository) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRow(ctx, `
		SELECT value FROM settings WHERE key = ?
	`, key).Scan(&value)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("querying setting: %w", err)
	}

	return value, nil
}

// GetWithDefault retrieves a setting, returning defaultVal if not found.
func (r *SettingsRepository) GetWithDefault(ctx context.Context, key, defaultVal string) (string, error) {
	value, err := r.Get(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return defaultVal, nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// Set creates or updates a setting.
func (r *SettingsRepository) Set(ctx context.Context, key, value string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := r.db.Exec(ctx, `
		INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value, now)
	if err != nil {
		return fmt.Errorf("setting value: %w", err)
	}

	return nil
}

// Delete removes a setting.
func (r *SettingsRepository) Delete(ctx context.Context, key string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM settings WHERE key = ?`, key)
	if err != nil {
		return fmt.Errorf("deleting setting: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// List returns all settings.
func (r *SettingsRepository) List(ctx context.Context) ([]Setting, error) {
	var settings []Setting
	err := r.db.Query(ctx, `
		SELECT key, value, updated_at FROM settings ORDER BY key
	`, func(rows *sql.Rows) error {
		for rows.Next() {
			var s Setting
			var updatedAt string

			if scanErr := rows.Scan(&s.Key, &s.Value, &updatedAt); scanErr != nil {
				return fmt.Errorf("scanning setting row: %w", scanErr)
			}

			if t, parseErr := time.Parse(time.RFC3339, updatedAt); parseErr == nil {
				s.UpdatedAt = t
			}

			settings = append(settings, s)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("querying settings: %w", err)
	}

	return settings, nil
}

// GetMultiple retrieves multiple settings by keys.
func (r *SettingsRepository) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}

	// Build query with placeholders
	query := "SELECT key, value FROM settings WHERE key IN (?" + repeatString(",?", len(keys)-1) + ")"
	args := make([]any, len(keys))
	for i, k := range keys {
		args[i] = k
	}

	result := make(map[string]string)
	err := r.db.Query(ctx, query, func(rows *sql.Rows) error {
		for rows.Next() {
			var key, value string
			if scanErr := rows.Scan(&key, &value); scanErr != nil {
				return fmt.Errorf("scanning setting row: %w", scanErr)
			}
			result[key] = value
		}
		return rows.Err()
	}, args...)
	if err != nil {
		return nil, fmt.Errorf("querying settings: %w", err)
	}

	return result, nil
}

// SetMultiple sets multiple settings in a single transaction.
func (r *SettingsRepository) SetMultiple(ctx context.Context, settings map[string]string) error {
	if len(settings) == 0 {
		return nil
	}

	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		now := time.Now().UTC().Format(time.RFC3339)

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO settings (key, value, updated_at)
			VALUES (?, ?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
		`)
		if err != nil {
			return fmt.Errorf("preparing statement: %w", err)
		}
		defer func() { _ = stmt.Close() }()

		for key, value := range settings {
			if _, execErr := stmt.ExecContext(ctx, key, value, now); execErr != nil {
				return fmt.Errorf("setting %s: %w", key, execErr)
			}
		}

		return nil
	})
}

// repeatString repeats a string n times.
func repeatString(s string, n int) string {
	var builder strings.Builder
	builder.Grow(len(s) * n)
	for range n {
		builder.WriteString(s)
	}
	return builder.String()
}
