// SPDX-License-Identifier: BUSL-1.1

package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/database"
)

func TestSettingsRepositoryGetSet(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	settings := db.Settings()

	if _, err := settings.Get(ctx, "absent"); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("Get on an absent key = %v, want ErrNotFound", err)
	}

	if err := settings.Set(ctx, "reflector.port", "3842"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := settings.Get(ctx, "reflector.port")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if value != "3842" {
		t.Errorf("Get = %q, want %q", value, "3842")
	}

	// Set is an upsert: a second Set replaces rather than duplicating.
	if err = settings.Set(ctx, "reflector.port", "3843"); err != nil {
		t.Fatalf("second Set failed: %v", err)
	}
	value, err = settings.Get(ctx, "reflector.port")
	if err != nil {
		t.Fatalf("Get after upsert failed: %v", err)
	}
	if value != "3843" {
		t.Errorf("Get after upsert = %q, want %q", value, "3843")
	}

	all, err := settings.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("List returned %d settings, want 1 after an upsert of the same key", len(all))
	}
	if all[0].Key != "reflector.port" || all[0].Value != "3843" {
		t.Errorf("List[0] = {%q, %q}, want {%q, %q}", all[0].Key, all[0].Value, "reflector.port", "3843")
	}
	if all[0].UpdatedAt.IsZero() {
		t.Error("List[0].UpdatedAt is zero; Set must stamp updated_at")
	}
}

func TestSettingsRepositoryGetWithDefault(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	settings := db.Settings()

	value, err := settings.GetWithDefault(ctx, "absent", "fallback")
	if err != nil {
		t.Fatalf("GetWithDefault failed: %v", err)
	}
	if value != "fallback" {
		t.Errorf("GetWithDefault on an absent key = %q, want %q", value, "fallback")
	}

	if err = settings.Set(ctx, "absent", "stored"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	value, err = settings.GetWithDefault(ctx, "absent", "fallback")
	if err != nil {
		t.Fatalf("GetWithDefault failed: %v", err)
	}
	if value != "stored" {
		t.Errorf("GetWithDefault on a stored key = %q, want %q", value, "stored")
	}
}

func TestSettingsRepositoryDelete(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	settings := db.Settings()

	if err := settings.Set(ctx, "doomed", "value"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := settings.Delete(ctx, "doomed"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := settings.Get(ctx, "doomed"); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("Get after Delete = %v, want ErrNotFound", err)
	}

	// Deleting a key that is not there is reported, not silently accepted.
	if err := settings.Delete(ctx, "doomed"); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("Delete of an absent key = %v, want ErrNotFound", err)
	}
}

func TestSettingsRepositoryListIsKeyOrdered(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	settings := db.Settings()

	for _, key := range []string{"gamma", "alpha", "beta"} {
		if err := settings.Set(ctx, key, "v-"+key); err != nil {
			t.Fatalf("Set(%q) failed: %v", key, err)
		}
	}

	all, err := settings.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	want := []string{"alpha", "beta", "gamma"}
	if len(all) != len(want) {
		t.Fatalf("List returned %d settings, want %d", len(all), len(want))
	}
	for i, key := range want {
		if all[i].Key != key {
			t.Errorf("List[%d].Key = %q, want %q (List must be ordered by key)", i, all[i].Key, key)
		}
		if all[i].Value != "v-"+key {
			t.Errorf("List[%d].Value = %q, want %q", i, all[i].Value, "v-"+key)
		}
	}
}

func TestSettingsRepositoryGetMultiple(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	settings := db.Settings()

	stored := map[string]string{"a": "1", "b": "2", "c": "3"}
	for key, value := range stored {
		if err := settings.Set(ctx, key, value); err != nil {
			t.Fatalf("Set(%q) failed: %v", key, err)
		}
	}

	empty, err := settings.GetMultiple(ctx, nil)
	if err != nil {
		t.Fatalf("GetMultiple(nil) failed: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("GetMultiple(nil) = %v, want an empty map", empty)
	}

	// Three keys exercise the placeholder builder past its first ",?".
	got, err := settings.GetMultiple(ctx, []string{"a", "c", "absent"})
	if err != nil {
		t.Fatalf("GetMultiple failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("GetMultiple returned %d keys, want 2 (the absent key must be omitted)", len(got))
	}
	if got["a"] != "1" {
		t.Errorf(`GetMultiple["a"] = %q, want "1"`, got["a"])
	}
	if got["c"] != "3" {
		t.Errorf(`GetMultiple["c"] = %q, want "3"`, got["c"])
	}
	if _, ok := got["b"]; ok {
		t.Error(`GetMultiple returned "b", which was not requested`)
	}
}

func TestSettingsRepositorySetMultiple(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	settings := db.Settings()

	if err := settings.SetMultiple(ctx, nil); err != nil {
		t.Fatalf("SetMultiple(nil) failed: %v", err)
	}
	all, err := settings.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("SetMultiple(nil) wrote %d settings, want 0", len(all))
	}

	if err = settings.Set(ctx, "existing", "old"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err = settings.SetMultiple(ctx, map[string]string{
		"existing": "new",
		"fresh":    "value",
	}); err != nil {
		t.Fatalf("SetMultiple failed: %v", err)
	}

	got, err := settings.GetMultiple(ctx, []string{"existing", "fresh"})
	if err != nil {
		t.Fatalf("GetMultiple failed: %v", err)
	}
	if got["existing"] != "new" {
		t.Errorf(`SetMultiple left "existing" = %q, want "new" (it must upsert)`, got["existing"])
	}
	if got["fresh"] != "value" {
		t.Errorf(`SetMultiple left "fresh" = %q, want "value"`, got["fresh"])
	}
}
