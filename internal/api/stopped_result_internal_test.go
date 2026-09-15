// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newStoppedResultServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", "stoppedresult")
	t.Setenv("STEM_AUTH_PASSWORD", "stoppedresult123")
	s, err := NewServer(8444)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	return s
}

// runningReflectorState reproduces what startReflector leaves behind: a
// result stored at start so the run is observable while it runs.
func runningReflectorState(s *Server) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.testStatus = statusRunning
	s.currentTest = testTypeReflect
	s.currentRunID = "run-1"
	s.testResult = &TestResultResponse{
		Status:   statusRunning,
		TestType: testTypeReflect,
		Module:   moduleReflector,
		Success:  true,
		Data:     map[string]any{"framesReflected": float64(12)},
	}
}

func testResultBody(t *testing.T, s *Server) TestResultResponse {
	t.Helper()
	rec := httptest.NewRecorder()
	s.handleTestResult(rec, httptest.NewRequest(http.MethodGet, "/api/v1/test/result", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/test/result = %d, want 200", rec.Code)
	}
	var got TestResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode result: %v (body %s)", err, rec.Body.String())
	}
	return got
}

// A stopped reflector kept advertising itself as running: the stop path moved
// testStatus to "stopped" but left the stored result at "running", so
// /api/v1/test/result described a run that had ended. The web UI only keeps a
// terminal result, so it kept none and the Results card said no test had run
// immediately after a real one (#1248).
func TestStoppingARunMarksTheStoredResultStopped(t *testing.T) {
	s := newStoppedResultServer(t)
	runningReflectorState(s)

	s.statsMu.Lock()
	s.markStoppedLocked()
	s.statsMu.Unlock()

	got := testResultBody(t, s)
	if got.Status != statusStopped {
		t.Errorf("result status = %q after the run was stopped, want %q", got.Status, statusStopped)
	}
	if got.Data == nil {
		t.Error("stopping the run threw away what it had captured; data must survive")
	}
	if got.TestType != testTypeReflect || got.Module != moduleReflector {
		t.Errorf("result identifies %q/%q, want %q/%q",
			got.TestType, got.Module, testTypeReflect, moduleReflector)
	}
}

// Stopping must clear the run the daemon advertises, or stats keep naming a
// run that is over (the invariant TestStatsDropsTheRunIDWhenNothingIsRunning
// asserts from the other side).
func TestStoppingClearsTheAdvertisedRun(t *testing.T) {
	s := newStoppedResultServer(t)
	runningReflectorState(s)

	s.statsMu.Lock()
	s.markStoppedLocked()
	s.statsMu.Unlock()

	stats := s.snapshotStats()
	if stats.TestStatus != statusStopped {
		t.Errorf("stats testStatus = %q, want %q", stats.TestStatus, statusStopped)
	}
	if stats.SuiteID != "" || stats.CurrentTest != nil {
		t.Errorf("stats still name a run: suiteId=%q currentTest=%v", stats.SuiteID, stats.CurrentTest)
	}
}

// Nothing to carry forward is not a crash: a stop before any result was
// stored leaves the endpoint on its no-result answer.
func TestStoppingWithNoStoredResultIsNotAFailure(t *testing.T) {
	s := newStoppedResultServer(t)

	s.statsMu.Lock()
	s.testStatus = statusRunning
	s.testResult = nil
	s.markStoppedLocked()
	s.statsMu.Unlock()

	got := testResultBody(t, s)
	if got.Status != statusStopped {
		t.Errorf("result status = %q, want the daemon's own %q", got.Status, statusStopped)
	}
}
