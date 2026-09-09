// SPDX-License-Identifier: BUSL-1.1

package database_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/database"
)

// newRun creates a test run and returns its id, so result tests have a parent
// row to hang off.
func newRun(t *testing.T, db *database.DB) string {
	t.Helper()

	id, err := db.TestRuns().Create(context.Background(), &database.TestRun{
		Module:   "benchmark",
		TestType: "rfc2544_throughput",
	})
	if err != nil {
		t.Fatalf("Create test run failed: %v", err)
	}
	return id
}

func TestTestResultRepositoryCreateBatch(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	runID := newRun(t, db)

	if err := db.TestResults().CreateBatch(ctx, nil); err != nil {
		t.Fatalf("CreateBatch(nil) failed: %v", err)
	}

	base := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	results := []database.TestResult{
		{
			RunID: runID, MetricType: database.MetricTypeThroughput, FrameSize: new(64),
			Value: 940.25, Unit: "Mbps", Timestamp: base,
		},
		{
			RunID: runID, MetricType: database.MetricTypeThroughput, FrameSize: new(1518),
			Value: 987.5, Unit: "Mbps", Timestamp: base.Add(time.Minute),
		},
	}

	if err := db.TestResults().CreateBatch(ctx, results); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	for i := range results {
		if results[i].ID == 0 {
			t.Errorf("CreateBatch did not write back an id for result %d", i)
		}
	}

	stored, err := db.TestResults().Get(ctx, results[0].ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if stored.Value != 940.25 {
		t.Errorf("Value = %v, want 940.25", stored.Value)
	}
	if stored.Unit != "Mbps" {
		t.Errorf("Unit = %q, want %q", stored.Unit, "Mbps")
	}
	if stored.FrameSize == nil || *stored.FrameSize != 64 {
		t.Errorf("FrameSize = %v, want 64", stored.FrameSize)
	}
	if !stored.Timestamp.Equal(base) {
		t.Errorf("Timestamp = %v, want %v", stored.Timestamp, base)
	}

	if _, err = db.TestResults().Get(ctx, results[1].ID+1000); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("Get on an absent id = %v, want ErrNotFound", err)
	}
}

func TestTestResultRepositoryCreateBatchFillsTimestamp(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	runID := newRun(t, db)

	before := time.Now().UTC().Add(-time.Second)
	results := []database.TestResult{
		{RunID: runID, MetricType: database.MetricTypeLatencyAvg, Value: 42, Unit: "us"},
	}
	if err := db.TestResults().CreateBatch(ctx, results); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	if results[0].Timestamp.Before(before) {
		t.Errorf("Timestamp = %v, want at or after %v (a zero timestamp must be filled in)",
			results[0].Timestamp, before)
	}
}

func TestTestResultRepositoryListFilters(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	runID := newRun(t, db)
	otherRunID := newRun(t, db)
	base := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)

	results := []database.TestResult{
		{
			RunID: runID, MetricType: database.MetricTypeThroughput, FrameSize: new(64),
			Value: 900, Unit: "Mbps", Timestamp: base,
		},
		{
			RunID: runID, MetricType: database.MetricTypeThroughput, FrameSize: new(1518),
			Value: 990, Unit: "Mbps", Timestamp: base.Add(time.Minute),
		},
		{
			RunID: runID, MetricType: database.MetricTypeLatencyAvg, FrameSize: new(64),
			Value: 55, Unit: "us", Timestamp: base.Add(2 * time.Minute),
		},
		{
			RunID: otherRunID, MetricType: database.MetricTypeThroughput, FrameSize: new(64),
			Value: 100, Unit: "Mbps", Timestamp: base,
		},
	}
	if err := db.TestResults().CreateBatch(ctx, results); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	byRun, err := db.TestResults().ListByRun(ctx, runID)
	if err != nil {
		t.Fatalf("ListByRun failed: %v", err)
	}
	if len(byRun) != 3 {
		t.Fatalf("ListByRun returned %d results, want 3", len(byRun))
	}
	// List orders by timestamp ascending.
	wantValues := []float64{900, 990, 55}
	for i, want := range wantValues {
		if byRun[i].Value != want {
			t.Errorf("ListByRun[%d].Value = %v, want %v (oldest first)", i, byRun[i].Value, want)
		}
	}

	byMetric, err := db.TestResults().List(ctx, database.TestResultQueryOptions{
		RunID:      runID,
		MetricType: database.MetricTypeThroughput,
	})
	if err != nil {
		t.Fatalf("List by metric failed: %v", err)
	}
	if len(byMetric) != 2 {
		t.Fatalf("List by metric returned %d results, want 2", len(byMetric))
	}

	bySize, err := db.TestResults().List(ctx, database.TestResultQueryOptions{
		RunID:     runID,
		FrameSize: new(1518),
	})
	if err != nil {
		t.Fatalf("List by frame size failed: %v", err)
	}
	if len(bySize) != 1 {
		t.Fatalf("List by frame size returned %d results, want 1", len(bySize))
	}
	if bySize[0].Value != 990 {
		t.Errorf("List by frame size returned value %v, want 990", bySize[0].Value)
	}

	windowed, err := db.TestResults().List(ctx, database.TestResultQueryOptions{
		RunID:     runID,
		TimeRange: database.TimeRange{Start: base.Add(30 * time.Second), End: base.Add(90 * time.Second)},
	})
	if err != nil {
		t.Fatalf("List with a time range failed: %v", err)
	}
	if len(windowed) != 1 || windowed[0].Value != 990 {
		t.Errorf("List with a time range = %v, want the single 990 Mbps result", windowed)
	}

	paged, err := db.TestResults().List(ctx, database.TestResultQueryOptions{
		RunID: runID, Limit: 1, Offset: 1,
	})
	if err != nil {
		t.Fatalf("List with limit/offset failed: %v", err)
	}
	if len(paged) != 1 || paged[0].Value != 990 {
		t.Errorf("List with Limit 1 Offset 1 = %v, want the second-oldest result (990)", paged)
	}
}

func TestTestResultRepositoryGetByFrameSize(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	runID := newRun(t, db)

	if err := db.TestResults().CreateBatch(ctx, []database.TestResult{
		{RunID: runID, MetricType: database.MetricTypeThroughput, FrameSize: new(64), Value: 900},
		{RunID: runID, MetricType: database.MetricTypeThroughput, FrameSize: new(64), Value: 910},
		{RunID: runID, MetricType: database.MetricTypeThroughput, FrameSize: new(1518), Value: 990},
		{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 1},
	}); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	grouped, err := db.TestResults().GetByFrameSize(ctx, runID, database.MetricTypeThroughput)
	if err != nil {
		t.Fatalf("GetByFrameSize failed: %v", err)
	}
	if len(grouped) != 3 {
		t.Fatalf("GetByFrameSize returned %d groups, want 3", len(grouped))
	}
	if len(grouped[64]) != 2 {
		t.Errorf("group 64 has %d results, want 2", len(grouped[64]))
	}
	if grouped[1518][0].Value != 990 {
		t.Errorf("group 1518 value = %v, want 990", grouped[1518][0].Value)
	}
	if len(grouped[0]) != 1 {
		t.Errorf("group 0 has %d results, want 1 (a NULL frame size groups under zero)", len(grouped[0]))
	}
}

func TestTestResultRepositoryGetAggregates(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	runID := newRun(t, db)

	if err := db.TestResults().CreateBatch(ctx, []database.TestResult{
		{RunID: runID, MetricType: database.MetricTypeLatencyAvg, Value: 10},
		{RunID: runID, MetricType: database.MetricTypeLatencyAvg, Value: 20},
		{RunID: runID, MetricType: database.MetricTypeLatencyAvg, Value: 60},
		{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 999},
	}); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	agg, err := db.TestResults().GetAggregates(ctx, runID, database.MetricTypeLatencyAvg)
	if err != nil {
		t.Fatalf("GetAggregates failed: %v", err)
	}
	if agg.Count != 3 {
		t.Errorf("Count = %d, want 3 (the throughput row must not be counted)", agg.Count)
	}
	if agg.Min != 10 {
		t.Errorf("Min = %v, want 10", agg.Min)
	}
	if agg.Max != 60 {
		t.Errorf("Max = %v, want 60", agg.Max)
	}
	if math.Abs(agg.Avg-30) > 1e-9 {
		t.Errorf("Avg = %v, want 30", agg.Avg)
	}
	if agg.Sum != 90 {
		t.Errorf("Sum = %v, want 90", agg.Sum)
	}
}

func TestTestResultRepositoryDeleteByRun(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	runID := newRun(t, db)
	otherRunID := newRun(t, db)

	if err := db.TestResults().CreateBatch(ctx, []database.TestResult{
		{RunID: runID, MetricType: database.MetricTypeThroughput, Value: 900},
		{RunID: otherRunID, MetricType: database.MetricTypeThroughput, Value: 100},
	}); err != nil {
		t.Fatalf("CreateBatch failed: %v", err)
	}

	if err := db.TestResults().DeleteByRun(ctx, runID); err != nil {
		t.Fatalf("DeleteByRun failed: %v", err)
	}

	gone, err := db.TestResults().ListByRun(ctx, runID)
	if err != nil {
		t.Fatalf("ListByRun failed: %v", err)
	}
	if len(gone) != 0 {
		t.Errorf("ListByRun after DeleteByRun returned %d results, want 0", len(gone))
	}

	kept, err := db.TestResults().ListByRun(ctx, otherRunID)
	if err != nil {
		t.Fatalf("ListByRun failed: %v", err)
	}
	if len(kept) != 1 || kept[0].Value != 100 {
		t.Errorf("DeleteByRun removed another run's results: %v", kept)
	}
}

func TestTestResultRepositorySummary(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	runID := newRun(t, db)

	if _, err := db.TestResults().GetSummary(ctx, runID); !errors.Is(err, database.ErrNotFound) {
		t.Errorf("GetSummary before CreateSummary = %v, want ErrNotFound", err)
	}

	throughput, latency, loss := 987.5, 42.25, 0.001
	sent, received := int64(1_000_000), int64(999_990)
	createdAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)

	summary := &database.TestSummary{
		RunID:          runID,
		Module:         "benchmark",
		TestType:       "rfc2544_throughput",
		Pass:           true,
		ThroughputMbps: &throughput,
		LatencyAvgUs:   &latency,
		FrameLossPct:   &loss,
		FramesSent:     &sent,
		FramesReceived: &received,
		CreatedAt:      createdAt,
	}

	id, err := db.TestResults().CreateSummary(ctx, summary)
	if err != nil {
		t.Fatalf("CreateSummary failed: %v", err)
	}
	if summary.ID != id {
		t.Errorf("CreateSummary did not write the id back: %d vs %d", summary.ID, id)
	}

	got, err := db.TestResults().GetSummary(ctx, runID)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if !got.Pass {
		t.Error("Pass = false, want true")
	}
	if got.ThroughputMbps == nil || *got.ThroughputMbps != throughput {
		t.Errorf("ThroughputMbps = %v, want %v", got.ThroughputMbps, throughput)
	}
	if got.LatencyAvgUs == nil || *got.LatencyAvgUs != latency {
		t.Errorf("LatencyAvgUs = %v, want %v", got.LatencyAvgUs, latency)
	}
	if got.FrameLossPct == nil || *got.FrameLossPct != loss {
		t.Errorf("FrameLossPct = %v, want %v", got.FrameLossPct, loss)
	}
	if got.FramesSent == nil || *got.FramesSent != sent {
		t.Errorf("FramesSent = %v, want %d", got.FramesSent, sent)
	}
	if got.FramesReceived == nil || *got.FramesReceived != received {
		t.Errorf("FramesReceived = %v, want %d", got.FramesReceived, received)
	}
	if got.LatencyMinUs != nil || got.LatencyMaxUs != nil || got.JitterUs != nil {
		t.Errorf("unset metrics came back non-nil: min=%v max=%v jitter=%v",
			got.LatencyMinUs, got.LatencyMaxUs, got.JitterUs)
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, createdAt)
	}
}

func TestTestResultRepositoryCreateSummaryFillsCreatedAt(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	runID := newRun(t, db)

	before := time.Now().UTC().Add(-time.Second)
	summary := &database.TestSummary{RunID: runID, Module: "benchmark", TestType: "rfc2544_latency"}
	if _, err := db.TestResults().CreateSummary(ctx, summary); err != nil {
		t.Fatalf("CreateSummary failed: %v", err)
	}
	if summary.CreatedAt.Before(before) {
		t.Errorf("CreatedAt = %v, want at or after %v (a zero value must be filled in)",
			summary.CreatedAt, before)
	}
}
