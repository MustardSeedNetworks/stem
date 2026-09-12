// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"testing"
)

func newRunIDStatsServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", "runidstats")
	t.Setenv("STEM_AUTH_PASSWORD", "runidstats123")
	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// A reflector run gets a run ID at start, but nothing could read it back:
// runPlan.describe is the only writer of SuiteID and a single-step run has
// no plan, so /api/v1/stats reported an empty suiteId while the daemon's own
// log named one (#1193). The CLI and the web UI cannot agree on a run ID
// neither of them can observe.
func TestStatsReportsTheReflectorRunID(t *testing.T) {
	s := newRunIDStatsServer(t)

	runID, err := s.beginTestRun(testTypeReflect, "reflector")
	if err != nil {
		t.Fatalf("beginTestRun: %v", err)
	}
	if runID == "" {
		t.Fatal("beginTestRun returned an empty run ID")
	}

	stats := s.snapshotStats()
	if stats.SuiteID != runID {
		t.Errorf("stats suiteId = %q, want the ID the run was started with (%q)", stats.SuiteID, runID)
	}
}

// A multi-step plan already reported its own ID; the single-step path must
// not start overwriting it.
func TestStatsPrefersThePlanRunID(t *testing.T) {
	s := newRunIDStatsServer(t)

	plan := &runPlan{Steps: []RunPlanStep{{TestType: "rfc2544_throughput"}}, Current: -1}
	if _, err := s.beginRunPlan(plan); err != nil {
		t.Fatalf("beginRunPlan: %v", err)
	}

	stats := s.snapshotStats()
	if stats.SuiteID != plan.ID {
		t.Errorf("stats suiteId = %q, want the plan's ID %q", stats.SuiteID, plan.ID)
	}
}

// Once a run is finished and cleared, stats must not keep advertising it.
func TestStatsDropsTheRunIDWhenNothingIsRunning(t *testing.T) {
	s := newRunIDStatsServer(t)

	if _, err := s.beginTestRun(testTypeReflect, "reflector"); err != nil {
		t.Fatalf("beginTestRun: %v", err)
	}
	s.statsMu.Lock()
	s.testStatus = statusStopped
	s.currentTest = ""
	s.currentRunID = ""
	s.statsMu.Unlock()

	if got := s.snapshotStats().SuiteID; got != "" {
		t.Errorf("stats suiteId = %q after the run was cleared, want empty", got)
	}
}
