// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"strings"
	"testing"
)

func newRunIDTestServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", "runidtest")
	t.Setenv("STEM_AUTH_PASSWORD", "runidpass123")
	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// The CLI and the web UI follow a run by its ID (#1166). A daemon restart
// resets the counter, so a plain "stem-1" names a different run before and
// after — a CLI reconnecting would report a stranger's result as its own.
func TestRunIDIsUniqueAcrossDaemonRestarts(t *testing.T) {
	first := newRunIDTestServer(t)
	second := newRunIDTestServer(t)

	plan := &runPlan{Steps: []RunPlanStep{{TestType: "rfc2544_throughput"}}, Current: -1}
	if _, err := first.beginRunPlan(plan); err != nil {
		t.Fatalf("beginRunPlan: %v", err)
	}
	restarted := &runPlan{Steps: []RunPlanStep{{TestType: "rfc2544_throughput"}}, Current: -1}
	if _, err := second.beginRunPlan(restarted); err != nil {
		t.Fatalf("beginRunPlan (restarted): %v", err)
	}

	if plan.ID == restarted.ID {
		t.Errorf("both daemons issued run ID %q; a restart must not reuse it", plan.ID)
	}
}

// Within one daemon the ID must still distinguish consecutive runs.
func TestRunIDIsUniquePerRun(t *testing.T) {
	s := newRunIDTestServer(t)

	seen := make(map[string]bool)
	for range 3 {
		plan := &runPlan{Steps: []RunPlanStep{{TestType: "rfc2544_throughput"}}, Current: -1}
		if _, err := s.beginRunPlan(plan); err != nil {
			t.Fatalf("beginRunPlan: %v", err)
		}
		if seen[plan.ID] {
			t.Fatalf("run ID %q issued twice", plan.ID)
		}
		seen[plan.ID] = true
		s.statsMu.Lock()
		s.testStatus = statusCompleted
		s.statsMu.Unlock()
	}
}

// Whatever the shape, the ID stays a single opaque token: the CLI passes it
// through to logs and comparisons, and a space or newline would split it.
func TestRunIDIsASingleToken(t *testing.T) {
	s := newRunIDTestServer(t)
	plan := &runPlan{Steps: []RunPlanStep{{TestType: "rfc2544_throughput"}}, Current: -1}
	if _, err := s.beginRunPlan(plan); err != nil {
		t.Fatalf("beginRunPlan: %v", err)
	}

	if strings.ContainsAny(plan.ID, " \t\r\n\"") {
		t.Errorf("run ID %q contains whitespace or a quote", plan.ID)
	}
	if !strings.HasPrefix(plan.ID, "stem-") {
		t.Errorf("run ID %q does not carry the stem- prefix", plan.ID)
	}
}

// The reflector start path built its own ID string by hand from testRunID,
// read without holding statsMu. Returning the ID from beginTestRun removes
// both the duplicated format and the unsynchronised read.
func TestBeginTestRunReturnsTheSameRunIDFormat(t *testing.T) {
	s := newRunIDTestServer(t)

	reflectID, err := s.beginTestRun(testTypeReflect, "reflector")
	if err != nil {
		t.Fatalf("beginTestRun: %v", err)
	}
	s.statsMu.Lock()
	s.testStatus = statusCompleted
	s.statsMu.Unlock()

	plan := &runPlan{Steps: []RunPlanStep{{TestType: "rfc2544_throughput"}}, Current: -1}
	if _, planErr := s.beginRunPlan(plan); planErr != nil {
		t.Fatalf("beginRunPlan: %v", planErr)
	}

	if reflectID == plan.ID {
		t.Errorf("reflector and plan runs share ID %q", reflectID)
	}
	for _, id := range []string{reflectID, plan.ID} {
		if !strings.HasPrefix(id, "stem-") || strings.ContainsAny(id, " \t\r\n\"") {
			t.Errorf("run ID %q is not a single stem- prefixed token", id)
		}
	}
	// Same instance, so the two IDs differ only in their trailing counter.
	if reflectPrefix, planPrefix := idInstance(reflectID), idInstance(plan.ID); reflectPrefix != planPrefix {
		t.Errorf("instance segment differs within one daemon: %q vs %q", reflectPrefix, planPrefix)
	}
}

// idInstance returns everything but the trailing counter segment.
func idInstance(id string) string {
	cut := strings.LastIndex(id, "-")
	if cut < 0 {
		return id
	}
	return id[:cut]
}
