// SPDX-License-Identifier: BUSL-1.1

package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/database"
)

func TestTestRunRepositoryDelete(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	runID := newRun(t, db)
	keptID := newRun(t, db)

	if err := db.TestResults().CreateBatch(ctx, []database.TestResult{
		{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 900},
	}); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	if err := db.TestRuns().Delete(ctx, runID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := db.TestRuns().Get(ctx, runID); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("Get after Delete = %v, want ErrNotFound", err)
	}
	if _, err := db.TestResults().ListByRun(ctx, runID); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("the deleted run's results survived (the CASCADE did not fire): %v", err)
	}
	if _, err := db.TestRuns().Get(ctx, keptID); err != nil {
		t.Errorf("Delete removed another run: %v", err)
	}

	if err := db.TestRuns().Delete(ctx, runID); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("Delete of an absent run = %v, want ErrNotFound", err)
	}
}

func TestTestRunRepositoryCount(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	seed := []database.TestRun{
		{Module: "benchmark", TestType: "rfc2544_throughput", Status: database.TestRunStatusCompleted},
		{Module: "benchmark", TestType: "rfc2544_latency", Status: database.TestRunStatusCompleted},
		{Module: "servicetest", TestType: "y1564", Status: database.TestRunStatusFailed},
	}
	for i := range seed {
		if _, err := db.TestRuns().Create(ctx, &seed[i]); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	cases := []struct {
		name string
		opts database.TestRunQueryOptions
		want int
	}{
		{"unfiltered", database.TestRunQueryOptions{}, 3},
		{"by module", database.TestRunQueryOptions{Module: "benchmark"}, 2},
		{"by test type", database.TestRunQueryOptions{TestType: "y1564"}, 1},
		{"by status", database.TestRunQueryOptions{Status: database.TestRunStatusCompleted}, 2},
		{
			"by module and status",
			database.TestRunQueryOptions{Module: "benchmark", Status: database.TestRunStatusFailed},
			0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := db.TestRuns().Count(ctx, tc.opts)
			if err != nil {
				t.Fatalf("Count failed: %v", err)
			}
			if got != tc.want {
				t.Errorf("Count = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTestRunRepositoryGetLatest(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	base := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)

	older := &database.TestRun{
		Module: "benchmark", TestType: "rfc2544_throughput",
		TargetAddress: "10.44.10.20", StartedAt: base,
	}
	newer := &database.TestRun{
		Module: "benchmark", TestType: "rfc2544_throughput",
		TargetAddress: "10.44.10.31", StartedAt: base.Add(time.Hour),
	}
	otherType := &database.TestRun{
		Module: "benchmark", TestType: "rfc2544_latency",
		TargetAddress: "10.44.10.99", StartedAt: base.Add(2 * time.Hour),
	}
	for _, run := range []*database.TestRun{older, newer, otherType} {
		if _, err := db.TestRuns().Create(ctx, run); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	latest, err := db.TestRuns().GetLatest(ctx, "benchmark", "rfc2544_throughput")
	if err != nil {
		t.Fatalf("GetLatest failed: %v", err)
	}
	if latest.ID != newer.ID {
		t.Errorf("GetLatest returned %q, want %q (the most recent run)", latest.ID, newer.ID)
	}
	if latest.TargetAddress != "10.44.10.31" {
		t.Errorf("GetLatest TargetAddress = %q, want %q", latest.TargetAddress, "10.44.10.31")
	}

	if _, err = db.TestRuns().GetLatest(ctx, "benchmark", "rfc6349"); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("GetLatest for a type with no runs = %v, want ErrNotFound", err)
	}
}

func TestTestRunRepositoryListFilters(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	base := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)

	first := &database.TestRun{
		Module: "benchmark", TestType: "rfc2544_throughput",
		Status: database.TestRunStatusCompleted, StartedAt: base,
	}
	second := &database.TestRun{
		Module: "benchmark", TestType: "rfc2544_latency",
		Status: database.TestRunStatusFailed, StartedAt: base.Add(time.Hour),
	}
	third := &database.TestRun{
		Module: "servicetest", TestType: "y1564",
		Status: database.TestRunStatusCompleted, StartedAt: base.Add(2 * time.Hour),
	}
	for _, run := range []*database.TestRun{first, second, third} {
		if _, err := db.TestRuns().Create(ctx, run); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	all, err := db.TestRuns().List(ctx, database.TestRunQueryOptions{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	wantOrder := []string{third.ID, second.ID, first.ID}
	if len(all) != len(wantOrder) {
		t.Fatalf("List returned %d runs, want %d", len(all), len(wantOrder))
	}
	for i, id := range wantOrder {
		if all[i].ID != id {
			t.Errorf("List[%d].ID = %q, want %q (newest first)", i, all[i].ID, id)
		}
	}

	byStatus, err := db.TestRuns().List(ctx, database.TestRunQueryOptions{
		Status: database.TestRunStatusFailed,
	})
	if err != nil {
		t.Fatalf("List by status failed: %v", err)
	}
	if len(byStatus) != 1 || byStatus[0].ID != second.ID {
		t.Fatalf("List by status = %v, want only %q", byStatus, second.ID)
	}

	windowed, err := db.TestRuns().List(ctx, database.TestRunQueryOptions{
		TimeRange: database.TimeRange{
			Start: base.Add(30 * time.Minute),
			End:   base.Add(90 * time.Minute),
		},
	})
	if err != nil {
		t.Fatalf("List with a time range failed: %v", err)
	}
	if len(windowed) != 1 || windowed[0].ID != second.ID {
		t.Fatalf("List with a time range = %v, want only %q", windowed, second.ID)
	}

	paged, err := db.TestRuns().List(ctx, database.TestRunQueryOptions{
		Module: "benchmark", Limit: 1, Offset: 1,
	})
	if err != nil {
		t.Fatalf("List with limit/offset failed: %v", err)
	}
	if len(paged) != 1 || paged[0].ID != first.ID {
		t.Fatalf("List with Limit 1 Offset 1 = %v, want only %q", paged, first.ID)
	}
}

func TestTestRunRepositoryCreateDefaults(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	before := time.Now().UTC().Add(-time.Second)
	run := &database.TestRun{Module: "benchmark", TestType: "rfc2544_throughput"}

	id, err := db.TestRuns().Create(ctx, run)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id == "" {
		t.Fatal("Create returned an empty id")
	}
	if run.ID != id {
		t.Errorf("Create did not write the id back: %q vs %q", run.ID, id)
	}
	if run.Status != database.TestRunStatusPending {
		t.Errorf("Status = %q, want %q by default", run.Status, database.TestRunStatusPending)
	}
	if run.StartedAt.Before(before) {
		t.Errorf("StartedAt = %v, want at or after %v (a zero value must be filled in)",
			run.StartedAt, before)
	}

	stored, err := db.TestRuns().Get(ctx, id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if stored.CompletedAt != nil {
		t.Errorf("CompletedAt = %v on a pending run, want nil", stored.CompletedAt)
	}
	if stored.DurationMs != nil {
		t.Errorf("DurationMs = %v on a pending run, want nil", stored.DurationMs)
	}
}

func TestTestRunRepositoryUpdateStatusAndCompleteOnMissingRun(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := db.TestRuns().UpdateStatus(ctx, "absent", database.TestRunStatusRunning); !errors.Is(
		err, database.ErrNotFound,
	) {
		t.Errorf("UpdateStatus on an absent run = %v, want ErrNotFound", err)
	}
	if err := db.TestRuns().Complete(ctx, "absent", database.TestRunStatusCompleted, ""); !errors.Is(
		err, database.ErrNotFound,
	) {
		t.Errorf("Complete on an absent run = %v, want ErrNotFound", err)
	}
}

func TestTestRunRepositoryCompleteRecordsDuration(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	startedAt := time.Now().UTC().Add(-90 * time.Second).Truncate(time.Second)
	run := &database.TestRun{
		Module: "benchmark", TestType: "rfc2544_throughput", StartedAt: startedAt,
	}
	id, err := db.TestRuns().Create(ctx, run)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err = db.TestRuns().Complete(ctx, id, database.TestRunStatusFailed, "link down"); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	stored, err := db.TestRuns().Get(ctx, id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if stored.Status != database.TestRunStatusFailed {
		t.Errorf("Status = %q, want %q", stored.Status, database.TestRunStatusFailed)
	}
	if stored.ErrorMessage != "link down" {
		t.Errorf("ErrorMessage = %q, want %q", stored.ErrorMessage, "link down")
	}
	if stored.CompletedAt == nil {
		t.Fatal("CompletedAt is nil after Complete")
	}
	if stored.DurationMs == nil {
		t.Fatal("DurationMs is nil after Complete")
	}
	// The run started 90s ago; the recorded duration must be that elapsed time,
	// not a placeholder.
	if *stored.DurationMs < 90_000 || *stored.DurationMs > 120_000 {
		t.Errorf("DurationMs = %d, want about 90000 (the time since StartedAt)", *stored.DurationMs)
	}
}
