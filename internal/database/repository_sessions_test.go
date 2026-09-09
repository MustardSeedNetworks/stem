// SPDX-License-Identifier: BUSL-1.1

package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/database"
)

// blacklist inserts a session with explicit timestamps so cleanup tests do not
// depend on wall-clock timing.
func blacklist(t *testing.T, db *database.DB, tokenID, username, reason string, expiresAt time.Time) int64 {
	t.Helper()

	session := &database.Session{
		TokenID:       tokenID,
		Username:      username,
		Reason:        reason,
		BlacklistedAt: expiresAt.Add(-time.Hour).UTC(),
		ExpiresAt:     expiresAt.UTC(),
	}

	id, err := db.Sessions().Blacklist(context.Background(), session)
	if err != nil {
		t.Fatalf("Blacklist(%q) failed: %v", tokenID, err)
	}
	if session.ID != id {
		t.Errorf("Blacklist did not write the id back: session.ID = %d, returned %d", session.ID, id)
	}
	return id
}

func TestSessionRepositoryBlacklistAndGet(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	expires := time.Now().UTC().Add(time.Hour).Truncate(time.Second)

	id := blacklist(t, db, "jti-1", "operator", database.SessionReasonLogout, expires)
	if id == 0 {
		t.Fatal("Blacklist returned id 0")
	}

	got, err := db.Sessions().Get(ctx, "jti-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != id {
		t.Errorf("Get().ID = %d, want %d", got.ID, id)
	}
	if got.Username != "operator" {
		t.Errorf("Get().Username = %q, want %q", got.Username, "operator")
	}
	if got.Reason != database.SessionReasonLogout {
		t.Errorf("Get().Reason = %q, want %q", got.Reason, database.SessionReasonLogout)
	}
	if !got.ExpiresAt.Equal(expires) {
		t.Errorf("Get().ExpiresAt = %v, want %v", got.ExpiresAt, expires)
	}

	if _, err = db.Sessions().Get(ctx, "absent"); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("Get on an absent token = %v, want ErrNotFound", err)
	}
}

func TestSessionRepositoryBlacklistDefaultsTimestamp(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	before := time.Now().UTC().Add(-time.Second)
	session := &database.Session{
		TokenID:   "jti-default",
		Username:  "operator",
		Reason:    database.SessionReasonForcedLogout,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	if _, err := db.Sessions().Blacklist(ctx, session); err != nil {
		t.Fatalf("Blacklist failed: %v", err)
	}

	if session.BlacklistedAt.Before(before) {
		t.Errorf("BlacklistedAt = %v, want at or after %v (a zero value must be filled in)",
			session.BlacklistedAt, before)
	}

	got, err := db.Sessions().Get(ctx, "jti-default")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.BlacklistedAt.IsZero() {
		t.Error("stored BlacklistedAt is zero")
	}
}

func TestSessionRepositoryIsBlacklisted(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	blacklist(t, db, "jti-known", "operator", database.SessionReasonLogout, time.Now().UTC().Add(time.Hour))

	known, err := db.Sessions().IsBlacklisted(ctx, "jti-known")
	if err != nil {
		t.Fatalf("IsBlacklisted failed: %v", err)
	}
	if !known {
		t.Error("IsBlacklisted on a blacklisted token = false, want true")
	}

	unknown, err := db.Sessions().IsBlacklisted(ctx, "jti-unknown")
	if err != nil {
		t.Fatalf("IsBlacklisted failed: %v", err)
	}
	if unknown {
		t.Error("IsBlacklisted on an unknown token = true, want false")
	}
}

func TestSessionRepositoryListing(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	base := time.Now().UTC().Add(time.Hour).Truncate(time.Second)

	// BlacklistedAt is expires-1h, so a later expiry means a later blacklisting.
	blacklist(t, db, "jti-old", "alice", database.SessionReasonLogout, base)
	blacklist(t, db, "jti-new", "alice", database.SessionReasonPasswordChange, base.Add(time.Hour))
	blacklist(t, db, "jti-bob", "bob", database.SessionReasonLogout, base.Add(2*time.Hour))

	all, err := db.Sessions().ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	wantOrder := []string{"jti-bob", "jti-new", "jti-old"}
	if len(all) != len(wantOrder) {
		t.Fatalf("ListAll returned %d sessions, want %d", len(all), len(wantOrder))
	}
	for i, tokenID := range wantOrder {
		if all[i].TokenID != tokenID {
			t.Errorf("ListAll[%d].TokenID = %q, want %q (newest blacklisting first)",
				i, all[i].TokenID, tokenID)
		}
	}

	alice, err := db.Sessions().ListByUser(ctx, "alice")
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(alice) != 2 {
		t.Fatalf("ListByUser(alice) returned %d sessions, want 2", len(alice))
	}
	if alice[0].TokenID != "jti-new" || alice[1].TokenID != "jti-old" {
		t.Errorf("ListByUser(alice) = [%q, %q], want [jti-new, jti-old]",
			alice[0].TokenID, alice[1].TokenID)
	}
	if alice[0].Reason != database.SessionReasonPasswordChange {
		t.Errorf("ListByUser(alice)[0].Reason = %q, want %q",
			alice[0].Reason, database.SessionReasonPasswordChange)
	}

	count, err := db.Sessions().Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}
}

func TestSessionRepositoryDelete(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	blacklist(t, db, "jti-doomed", "operator", database.SessionReasonLogout, time.Now().UTC().Add(time.Hour))

	if err := db.Sessions().Delete(ctx, "jti-doomed"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := db.Sessions().Get(ctx, "jti-doomed"); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("Get after Delete = %v, want ErrNotFound", err)
	}
	if err := db.Sessions().Delete(ctx, "jti-doomed"); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("Delete of an absent token = %v, want ErrNotFound", err)
	}
}

func TestSessionRepositoryCleanupExpired(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	blacklist(t, db, "jti-expired", "alice", database.SessionReasonLogout, now.Add(-time.Hour))
	blacklist(t, db, "jti-live", "alice", database.SessionReasonLogout, now.Add(time.Hour))

	deleted, err := db.Sessions().CleanupExpired(ctx)
	if err != nil {
		t.Fatalf("CleanupExpired failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("CleanupExpired removed %d sessions, want 1", deleted)
	}

	if _, err = db.Sessions().Get(ctx, "jti-expired"); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("the expired session survived cleanup: Get = %v, want ErrNotFound", err)
	}
	if _, err = db.Sessions().Get(ctx, "jti-live"); err != nil {
		t.Errorf("cleanup removed an unexpired session: Get = %v", err)
	}
}
