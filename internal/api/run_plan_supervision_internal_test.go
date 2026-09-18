// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/instance"

	"github.com/MustardSeedNetworks/stem/internal/logging"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// The run plan is the one production goroutine that executes the dataplane
// (D-STEM-19, stem#1336). Before this test the launch site was a bare
// `go s.runTestPlan(...)`, so a panic inside an executor took the daemon's
// whole process down with it and every other in-flight run with it. The
// supervisor has to turn that panic into a failed run the operator can read.
func TestDataplanePanicFailsTheRunAndLeavesTheDaemonUp(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "panicrunuser")
	t.Setenv("STEM_AUTH_PASSWORD", "panicrunpass123")
	s := newTestServer(t)

	plan, err := newRunPlan("", TestStartRequest{
		Peer:  "198.51.100.9",
		Tests: []TestStepRequest{{TestType: "rfc2544_throughput"}},
	})
	if err != nil {
		t.Fatalf("newRunPlan: %v", err)
	}
	s.executorResolver = func(string) (executorFactory, bool) {
		return func(string) (testExecutor, error) { return &panicExecutor{}, nil }, true
	}

	log := captureLog(t)
	runID, beginErr := s.beginRunPlan(plan)
	if beginErr != nil {
		t.Fatalf("beginRunPlan: %v", beginErr)
	}
	s.startRunPlan(runID, "lo0", plan.ID)

	// The daemon is still the process running this assertion: a panic that
	// escaped would have taken the test binary with it.
	waitForTestStatus(t, s, statusError)

	s.statsMu.RLock()
	cause, stepStatus := s.testError, s.runPlan.Steps[0].Status
	s.statsMu.RUnlock()
	if stepStatus != stepFailed {
		t.Errorf("step status = %q, want %q", stepStatus, stepFailed)
	}
	if cause != causeInternalFault {
		t.Errorf("testError = %q, want %q", cause, causeInternalFault)
	}

	// One fault, one line: the supervisor's, carrying the run it belongs to.
	// A second line from the bookkeeping would make one fault read as two.
	lines := errorLines(log.String())
	if len(lines) != 1 {
		t.Fatalf("the fault produced %d error lines, want 1:\n%s", len(lines), log.String())
	}
	if !strings.Contains(lines[0], plan.ID) {
		t.Errorf("the fault line does not name the run %q:\n%s", plan.ID, lines[0])
	}
	if !strings.Contains(lines[0], "forced dataplane fault") {
		t.Errorf("the fault line does not name the cause:\n%s", lines[0])
	}
}

// captureLog points the process logger at a buffer for the duration of the
// test. The "logged once" clause is a property of the log, so it has to be
// read, not inferred from the call sites.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	cfg := logging.DefaultConfig()
	cfg.Level = "debug"
	if err := logging.InitWithWriter(cfg, buf); err != nil {
		t.Fatalf("InitWithWriter: %v", err)
	}
	t.Cleanup(logging.Reset)
	return buf
}

// errorLines picks the error-level records out of the capture. The daemon's
// default handler is JSON; the text form is matched too so the assertion does
// not silently pass by matching nothing if that default changes.
func errorLines(output string) []string {
	var lines []string
	for line := range strings.SplitSeq(output, "\n") {
		if strings.Contains(line, `"level":"error"`) || strings.Contains(line, "level=ERROR") {
			lines = append(lines, line)
		}
	}
	return lines
}

// A panic must not leave the run slot held: the operator has to be able to
// start the next run without restarting the daemon.
func TestDataplanePanicReleasesTheRunSlot(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "panicslotuser")
	t.Setenv("STEM_AUTH_PASSWORD", "panicslotpass123")
	s := newTestServer(t)

	request := TestStartRequest{
		Peer:  "198.51.100.9",
		Tests: []TestStepRequest{{TestType: "rfc2544_throughput"}},
	}
	plan, err := newRunPlan("", request)
	if err != nil {
		t.Fatalf("newRunPlan: %v", err)
	}
	s.executorResolver = func(string) (executorFactory, bool) {
		return func(string) (testExecutor, error) { return &panicExecutor{}, nil }, true
	}
	runID, beginErr := s.beginRunPlan(plan)
	if beginErr != nil {
		t.Fatalf("beginRunPlan: %v", beginErr)
	}
	s.startRunPlan(runID, "lo0", plan.ID)
	waitForTestStatus(t, s, statusError)

	next, planErr := newRunPlan("", request)
	if planErr != nil {
		t.Fatalf("newRunPlan (second): %v", planErr)
	}
	if _, secondErr := s.beginRunPlan(next); secondErr != nil {
		t.Fatalf("beginRunPlan after a panicking run: %v", secondErr)
	}
}

func waitForTestStatus(t *testing.T, s *Server, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s.statsMu.RLock()
		got := s.testStatus
		s.statsMu.RUnlock()
		if got == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	s.statsMu.RLock()
	got := s.testStatus
	s.statsMu.RUnlock()
	t.Fatalf("testStatus = %q after 5s, want %q", got, want)
}

// panicExecutor is the forced dataplane fault: a module whose Execute panics
// the way a nil dereference in the cgo boundary would.
type panicExecutor struct {
	closed sync.Once
}

func (e *panicExecutor) Close() { e.closed.Do(func() {}) }

func (*panicExecutor) Execute(string, *modtypes.TestConfig) (*modtypes.Result, error) {
	panic("forced dataplane fault")
}

// Shutdown hands the data directory back. Process exit would release the
// flock anyway, but a daemon that stops and starts inside one process — the
// E2E harness does exactly that — must not refuse its own restart.
func TestShutdownReleasesTheInstanceLock(t *testing.T) {
	dataDir := t.TempDir()

	lock, err := instance.Acquire(dataDir)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	s := &Server{dataDir: dataDir, instanceLock: lock}
	if shutdownErr := s.Shutdown(); shutdownErr != nil {
		t.Fatalf("Shutdown: %v", shutdownErr)
	}

	again, reacquireErr := instance.Acquire(dataDir)
	if reacquireErr != nil {
		t.Fatalf("the data directory is still locked after Shutdown: %v", reacquireErr)
	}
	_ = again.Release()
}
