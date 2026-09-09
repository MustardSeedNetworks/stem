// SPDX-License-Identifier: BUSL-1.1

package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/database"
)

// logAt writes an audit entry with an explicit timestamp so ordering and
// retention assertions do not depend on wall-clock timing.
func logAt(t *testing.T, db *database.DB, entry *database.AuditLogEntry, at time.Time) int64 {
	t.Helper()

	entry.Timestamp = at.UTC()
	id, err := db.AuditLog().Log(context.Background(), entry)
	if err != nil {
		t.Fatalf("Log(%q) failed: %v", entry.Action, err)
	}
	if entry.ID != id {
		t.Errorf("Log did not write the id back: entry.ID = %d, returned %d", entry.ID, id)
	}
	return id
}

func TestAuditLogRepositoryLogAndGet(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	at := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)

	id := logAt(t, db, &database.AuditLogEntry{
		Action:       database.AuditActionConfigChange,
		User:         "operator",
		ResourceType: "settings",
		ResourceID:   "reflector.port",
		OldValueJSON: `{"port":3842}`,
		NewValueJSON: `{"port":3843}`,
		IPAddress:    "10.44.10.9",
		UserAgent:    "stem-cli",
	}, at)

	got, err := db.AuditLog().Get(ctx, id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Action != database.AuditActionConfigChange {
		t.Errorf("Action = %q, want %q", got.Action, database.AuditActionConfigChange)
	}
	if got.User != "operator" {
		t.Errorf("User = %q, want %q", got.User, "operator")
	}
	if got.ResourceType != "settings" || got.ResourceID != "reflector.port" {
		t.Errorf("resource = (%q, %q), want (settings, reflector.port)", got.ResourceType, got.ResourceID)
	}
	if got.OldValueJSON != `{"port":3842}` || got.NewValueJSON != `{"port":3843}` {
		t.Errorf("values = (%q, %q), want the old and new JSON round-tripped",
			got.OldValueJSON, got.NewValueJSON)
	}
	if got.IPAddress != "10.44.10.9" {
		t.Errorf("IPAddress = %q, want %q", got.IPAddress, "10.44.10.9")
	}
	if got.UserAgent != "stem-cli" {
		t.Errorf("UserAgent = %q, want %q", got.UserAgent, "stem-cli")
	}
	if !got.Timestamp.Equal(at) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, at)
	}

	if _, err = db.AuditLog().Get(ctx, id+1000); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("Get on an absent id = %v, want ErrNotFound", err)
	}
}

func TestAuditLogRepositoryConvenienceLoggers(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	before := time.Now().UTC().Add(-time.Second)
	if err := db.AuditLog().LogAction(ctx, database.AuditActionLogin, "alice", "10.44.10.5"); err != nil {
		t.Fatalf("LogAction failed: %v", err)
	}
	if err := db.AuditLog().LogResourceChange(ctx,
		database.AuditActionUserUpdated, "alice", "user", "bob", `{"role":"viewer"}`, `{"role":"operator"}`,
		"10.44.10.5",
	); err != nil {
		t.Fatalf("LogResourceChange failed: %v", err)
	}

	entries, err := db.AuditLog().ListByUser(ctx, "alice", 10)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("ListByUser returned %d entries, want 2", len(entries))
	}

	byAction := make(map[string]database.AuditLogEntry, len(entries))
	for _, entry := range entries {
		byAction[entry.Action] = entry
	}

	login, ok := byAction[database.AuditActionLogin]
	if !ok {
		t.Fatalf("LogAction did not store a %q entry", database.AuditActionLogin)
	}
	if login.IPAddress != "10.44.10.5" {
		t.Errorf("LogAction IPAddress = %q, want %q", login.IPAddress, "10.44.10.5")
	}
	if login.ResourceType != "" {
		t.Errorf("LogAction ResourceType = %q, want empty", login.ResourceType)
	}
	if login.Timestamp.Before(before) {
		t.Errorf("LogAction Timestamp = %v, want at or after %v (a zero timestamp must be filled in)",
			login.Timestamp, before)
	}

	change, ok := byAction[database.AuditActionUserUpdated]
	if !ok {
		t.Fatalf("LogResourceChange did not store a %q entry", database.AuditActionUserUpdated)
	}
	if change.ResourceType != "user" || change.ResourceID != "bob" {
		t.Errorf("LogResourceChange resource = (%q, %q), want (user, bob)",
			change.ResourceType, change.ResourceID)
	}
	if change.OldValueJSON != `{"role":"viewer"}` || change.NewValueJSON != `{"role":"operator"}` {
		t.Errorf("LogResourceChange values = (%q, %q), want the viewer/operator JSON",
			change.OldValueJSON, change.NewValueJSON)
	}

	byResource, err := db.AuditLog().ListByResource(ctx, "user", "bob", 10)
	if err != nil {
		t.Fatalf("ListByResource failed: %v", err)
	}
	if len(byResource) != 1 {
		t.Fatalf("ListByResource returned %d entries, want 1", len(byResource))
	}
	if byResource[0].ID != change.ID {
		t.Errorf("ListByResource returned id %d, want %d", byResource[0].ID, change.ID)
	}
}

func TestAuditLogRepositoryListFilters(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	base := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)

	oldest := logAt(t, db, &database.AuditLogEntry{
		Action: database.AuditActionLogin, User: "alice",
	}, base)
	middle := logAt(t, db, &database.AuditLogEntry{
		Action: database.AuditActionLoginFailed, User: "alice",
	}, base.Add(time.Hour))
	newest := logAt(t, db, &database.AuditLogEntry{
		Action: database.AuditActionLogin, User: "bob",
	}, base.Add(2*time.Hour))

	all, err := db.AuditLog().List(ctx, database.AuditLogQueryOptions{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	wantOrder := []int64{newest, middle, oldest}
	if len(all) != len(wantOrder) {
		t.Fatalf("List returned %d entries, want %d", len(all), len(wantOrder))
	}
	for i, id := range wantOrder {
		if all[i].ID != id {
			t.Errorf("List[%d].ID = %d, want %d (newest first)", i, all[i].ID, id)
		}
	}

	byAction, err := db.AuditLog().List(ctx, database.AuditLogQueryOptions{
		Action: database.AuditActionLogin,
	})
	if err != nil {
		t.Fatalf("List by action failed: %v", err)
	}
	if len(byAction) != 2 {
		t.Fatalf("List by action returned %d entries, want 2", len(byAction))
	}
	for _, entry := range byAction {
		if entry.Action != database.AuditActionLogin {
			t.Errorf("List by action returned action %q", entry.Action)
		}
	}

	limited, err := db.AuditLog().List(ctx, database.AuditLogQueryOptions{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("List with limit/offset failed: %v", err)
	}
	if len(limited) != 1 {
		t.Fatalf("List with Limit 1 returned %d entries, want 1", len(limited))
	}
	if limited[0].ID != middle {
		t.Errorf("List with Offset 1 returned id %d, want %d (the second newest)", limited[0].ID, middle)
	}

	windowed, err := db.AuditLog().List(ctx, database.AuditLogQueryOptions{
		TimeRange: database.TimeRange{
			Start: base.Add(30 * time.Minute),
			End:   base.Add(90 * time.Minute),
		},
	})
	if err != nil {
		t.Fatalf("List with a time range failed: %v", err)
	}
	if len(windowed) != 1 {
		t.Fatalf("List with a time range returned %d entries, want 1", len(windowed))
	}
	if windowed[0].ID != middle {
		t.Errorf("List with a time range returned id %d, want %d", windowed[0].ID, middle)
	}
}

func TestAuditLogRepositoryCount(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	base := time.Now().UTC().Add(-time.Hour)

	logAt(t, db, &database.AuditLogEntry{
		Action: database.AuditActionLogin, User: "alice", ResourceType: "session",
	}, base)
	logAt(t, db, &database.AuditLogEntry{
		Action: database.AuditActionLogin, User: "bob", ResourceType: "session",
	}, base)
	logAt(t, db, &database.AuditLogEntry{
		Action: database.AuditActionLoginFailed, User: "alice",
	}, base)

	cases := []struct {
		name string
		opts database.AuditLogQueryOptions
		want int
	}{
		{"unfiltered", database.AuditLogQueryOptions{}, 3},
		{"by action", database.AuditLogQueryOptions{Action: database.AuditActionLogin}, 2},
		{"by user", database.AuditLogQueryOptions{User: "alice"}, 2},
		{"by resource type", database.AuditLogQueryOptions{ResourceType: "session"}, 2},
		{
			"by action and user",
			database.AuditLogQueryOptions{Action: database.AuditActionLogin, User: "alice"},
			1,
		},
		{"no match", database.AuditLogQueryOptions{User: "carol"}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := db.AuditLog().Count(ctx, tc.opts)
			if err != nil {
				t.Fatalf("Count failed: %v", err)
			}
			if got != tc.want {
				t.Errorf("Count = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestAuditLogRepositoryDeleteOlderThan(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	stale := logAt(t, db, &database.AuditLogEntry{Action: database.AuditActionLogin, User: "alice"},
		now.Add(-48*time.Hour))
	kept := logAt(t, db, &database.AuditLogEntry{Action: database.AuditActionLogin, User: "bob"},
		now.Add(-time.Hour))

	deleted, err := db.AuditLog().DeleteOlderThan(ctx, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("DeleteOlderThan failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("DeleteOlderThan removed %d entries, want 1", deleted)
	}

	if _, err = db.AuditLog().Get(ctx, stale); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("the entry older than the cutoff survived: Get = %v, want ErrNotFound", err)
	}
	if _, err = db.AuditLog().Get(ctx, kept); err != nil {
		t.Errorf("an entry newer than the cutoff was deleted: Get = %v", err)
	}
}
