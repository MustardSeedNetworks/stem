// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// The History page can never show a run while the daemon answers with no
// per-run timing (#1333). The run's clock belongs to the side that ran it, so
// every result the daemon stores for an ended run carries when it started,
// when it finished and how long it took.
//
// The cases below are the terminal transitions a run can take. Each drives the
// production path rather than assigning the fields, because what is at stake
// is that the daemon records the timing, not that the struct can hold it.

func TestCompletedRunCarriesItsTimingOnTheWire(t *testing.T) {
	s := newTimingTestServer(t)
	s.executorResolver = passingExecutorResolver(map[string]any{"throughputMbps": 942.4})

	plan := planFor(t, "rfc2544_throughput")
	runID, err := s.beginRunPlan(plan)
	if err != nil {
		t.Fatalf("beginRunPlan: %v", err)
	}
	before := time.Now()
	s.startRunPlan(runID, "lo0", plan.ID)
	waitForTestStatus(t, s, statusCompleted)

	recorder := httptest.NewRecorder()
	s.handleTestResult(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/test/result", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body struct {
		Status      string     `json:"status"`
		StartedAt   *time.Time `json:"startedAt"`
		CompletedAt *time.Time `json:"completedAt"`
		Duration    *int64     `json:"duration"`
	}
	if decodeErr := json.Unmarshal(recorder.Body.Bytes(), &body); decodeErr != nil {
		t.Fatalf("decode %s: %v", recorder.Body.String(), decodeErr)
	}
	if body.Status != statusCompleted {
		t.Fatalf("status = %q, want %q", body.Status, statusCompleted)
	}
	if body.StartedAt == nil || body.CompletedAt == nil || body.Duration == nil {
		t.Fatalf("the completed run carries no timing: %s", recorder.Body.String())
	}
	if body.StartedAt.After(*body.CompletedAt) {
		t.Errorf("startedAt %s is after completedAt %s", body.StartedAt, body.CompletedAt)
	}
	if body.StartedAt.Before(before.Add(-time.Second)) {
		t.Errorf("startedAt %s predates the run", body.StartedAt)
	}
	// The duration is the daemon's own measurement of the two stamps, in
	// milliseconds — the unit HistoryPage formats.
	want := body.CompletedAt.Sub(*body.StartedAt).Milliseconds()
	if *body.Duration != want {
		t.Errorf("duration = %d ms, want %d ms", *body.Duration, want)
	}
}

func TestEveryTerminalRunCarriesItsTiming(t *testing.T) {
	tests := []struct {
		name   string
		drive  func(t *testing.T, s *Server)
		status string
	}{
		{
			name: "the run plan completes",
			drive: func(t *testing.T, s *Server) {
				t.Helper()
				s.executorResolver = passingExecutorResolver(nil)
				plan := planFor(t, "rfc2544_throughput")
				runID, err := s.beginRunPlan(plan)
				if err != nil {
					t.Fatalf("beginRunPlan: %v", err)
				}
				s.startRunPlan(runID, "lo0", plan.ID)
				waitForTestStatus(t, s, statusCompleted)
			},
			status: statusCompleted,
		},
		{
			name: "a step fails",
			drive: func(t *testing.T, s *Server) {
				t.Helper()
				s.executorResolver = failingExecutorResolver()
				plan := planFor(t, "rfc2544_throughput")
				runID, err := s.beginRunPlan(plan)
				if err != nil {
					t.Fatalf("beginRunPlan: %v", err)
				}
				s.startRunPlan(runID, "lo0", plan.ID)
				waitForTestStatus(t, s, statusError)
			},
			status: statusError,
		},
		{
			name: "the operator stops the reflector",
			drive: func(t *testing.T, s *Server) {
				t.Helper()
				if _, err := s.beginTestRun(testTypeReflect, moduleReflector); err != nil {
					t.Fatalf("beginTestRun: %v", err)
				}
				s.statsMu.Lock()
				s.testStatus = statusRunning
				s.testResult = &TestResultResponse{
					Status:   statusRunning,
					TestType: testTypeReflect,
					Module:   moduleReflector,
				}
				s.stampRunTimingLocked(s.testResult)
				s.markStoppedLocked()
				s.statsMu.Unlock()
			},
			status: statusStopped,
		},
		{
			name: "the executor cannot start",
			drive: func(t *testing.T, s *Server) {
				t.Helper()
				if _, err := s.beginTestRun(testTypeReflect, moduleReflector); err != nil {
					t.Fatalf("beginTestRun: %v", err)
				}
				s.respondTestExecutionError(
					httptest.NewRecorder(),
					errStubDataplane,
					moduleReflector,
					testTypeReflect,
				)
			},
			status: statusError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newTimingTestServer(t)
			tc.drive(t, s)

			s.statsMu.RLock()
			result := s.testResult
			s.statsMu.RUnlock()
			if result == nil {
				t.Fatal("the ended run stored no result")
			}
			if result.Status != tc.status {
				t.Fatalf("status = %q, want %q", result.Status, tc.status)
			}
			if result.StartedAt == nil {
				t.Error("the ended run carries no startedAt")
			}
			if result.CompletedAt == nil {
				t.Fatal("the ended run carries no completedAt")
			}
			if result.DurationMs == nil {
				t.Fatal("the ended run carries no duration")
			}
			if *result.DurationMs < 0 {
				t.Errorf("duration = %d ms, want a non-negative measurement", *result.DurationMs)
			}
		})
	}
}

// A run still in flight has a start and no end: stamping a completion on it
// would put a finished run in the operator's history while it is still going.
func TestRunningReflectorHasNoCompletion(t *testing.T) {
	s := newTimingTestServer(t)
	if _, err := s.beginTestRun(testTypeReflect, moduleReflector); err != nil {
		t.Fatalf("beginTestRun: %v", err)
	}
	s.statsMu.Lock()
	result := &TestResultResponse{Status: statusRunning, TestType: testTypeReflect}
	s.stampRunTimingLocked(result)
	s.statsMu.Unlock()

	if result.StartedAt == nil {
		t.Error("a running reflector carries no startedAt")
	}
	if result.CompletedAt != nil {
		t.Errorf("a running reflector was stamped complete at %s", result.CompletedAt)
	}
	if result.DurationMs != nil {
		t.Errorf("duration = %d ms, want none while the run is in flight", *result.DurationMs)
	}
}

// newTimingTestServer is newTestServer with the credentials the auth manager
// requires; the package's other suites set them per test the same way.
func newTimingTestServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", "runtiminguser")
	t.Setenv("STEM_AUTH_PASSWORD", "runtimingpass123")
	return newTestServer(t)
}

func planFor(t *testing.T, testType string) *runPlan {
	t.Helper()
	plan, err := newRunPlan("", TestStartRequest{
		Peer:  "198.51.100.9",
		Tests: []TestStepRequest{{TestType: testType}},
	})
	if err != nil {
		t.Fatalf("newRunPlan: %v", err)
	}
	return plan
}

func passingExecutorResolver(data any) func(string) (executorFactory, bool) {
	return func(string) (executorFactory, bool) {
		return func(string) (testExecutor, error) {
			return &stubExecutor{result: &modtypes.Result{Success: true, Data: data}}, nil
		}, true
	}
}

func failingExecutorResolver() func(string) (executorFactory, bool) {
	return func(string) (executorFactory, bool) {
		return func(string) (testExecutor, error) {
			return &stubExecutor{err: errStubDataplane}, nil
		}, true
	}
}

var errStubDataplane = errors.New("no dataplane on this build")

type stubExecutor struct {
	result *modtypes.Result
	err    error
}

func (*stubExecutor) Close() {}

func (e *stubExecutor) Execute(string, *modtypes.TestConfig) (*modtypes.Result, error) {
	return e.result, e.err
}
