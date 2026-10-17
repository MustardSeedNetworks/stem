// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/update"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

// newRunLockServer is newTestServer with credentials, plus a second guard on
// the daemon's data directory standing in for `stem update`.
func newRunLockServer(t *testing.T) (*Server, *update.RunGuard) {
	t.Helper()
	t.Setenv("STEM_AUTH_USERNAME", "runlockuser")
	t.Setenv("STEM_AUTH_PASSWORD", "runlockpass123")
	s := newTestServer(t)
	return s, update.NewRunGuard(s.dataDir)
}

// While an update holds the run lock, no run may start: it would measure a
// binary that is being replaced. The refusal leaves the daemon idle, so the
// operator can start the run once the update is done.
func TestRunStartRefusedWhileAnUpdateHoldsTheRunLock(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"run plan", `{"peer":"192.0.2.1","tests":[{"testType":"rfc2544_latency"}]}`},
		{"reflector", `{"tests":[{"testType":"reflect"}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, updater := newRunLockServer(t)
			release, err := updater.Acquire()
			if err != nil {
				t.Fatalf("updater Acquire: %v", err)
			}
			defer release()

			s.statsMu.RLock()
			runID := s.testRunID
			s.statsMu.RUnlock()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			s.handleTestStart(w, req)

			if w.Code != http.StatusConflict {
				t.Fatalf("status = %d, want 409 while an update holds the run lock: %s", w.Code, w.Body)
			}
			if !strings.Contains(w.Body.String(), "update") {
				t.Errorf("body %q does not say an update is being installed", w.Body)
			}
			s.statsMu.RLock()
			defer s.statsMu.RUnlock()
			if s.testRunID != runID || s.testStatus == statusStarting || s.testStatus == statusRunning {
				t.Errorf("a run was reserved (runID %d -> %d, status %q)", runID, s.testRunID, s.testStatus)
			}
		})
	}
}

// heldExecutor stays in Execute until released, so the test can look at the
// run lock while a step is on the wire.
type heldExecutor struct{ release chan struct{} }

func (*heldExecutor) Close() {}

func (e *heldExecutor) Execute(testType string, _ *modtypes.TestConfig) (*modtypes.Result, error) {
	<-e.release
	return &modtypes.Result{TestType: testType, Success: true}, nil
}

// A run plan holds the run lock from its start until its worker returns, and
// a finished plan lets go so the update can proceed.
func TestRunPlanHoldsTheRunLockUntilItEnds(t *testing.T) {
	s, updater := newRunLockServer(t)
	exec := &heldExecutor{release: make(chan struct{})}
	s.executorResolver = func(string) (executorFactory, bool) {
		return func(string) (testExecutor, error) { return exec, nil }, true
	}

	body := bytes.NewBufferString(`{"peer":"192.0.2.1","tests":[{"testType":"rfc2544_throughput"}]}`)
	w := httptest.NewRecorder()
	s.handleTestStart(w, httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body))
	if w.Code != http.StatusOK {
		t.Fatalf("start status = %d: %s", w.Code, w.Body)
	}
	waitForTestStatus(t, s, statusRunning)

	if _, err := updater.Acquire(); !errors.Is(err, update.ErrRunLockHeld) {
		t.Fatalf("updater Acquire during a run = %v; want ErrRunLockHeld", err)
	}

	close(exec.release)
	waitForTestStatus(t, s, statusCompleted)
	waitForRunLock(t, updater)
}

// A reflector that fails to start drops the hold it was given; one that
// never released would block every later update until the daemon restarts.
func TestFailedReflectorStartReleasesTheRunLock(t *testing.T) {
	s, updater := newRunLockServer(t)
	release, err := s.runGuard.Acquire()
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	if execErr := s.executeReflector("stem-nonexistent-if0", "", release); execErr == nil {
		t.Fatal("executeReflector on a missing interface succeeded")
	}
	waitForRunLock(t, updater)
}

// A mode switch stops the reflector, and the hold goes with it.
func TestModeSwitchReleasesTheReflectorHold(t *testing.T) {
	s, updater := newRunLockServer(t)
	release, err := s.runGuard.Acquire()
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	s.statsMu.Lock()
	s.reflectorRelease = release
	s.statsMu.Unlock()

	s.teardownForModeSwitch()
	waitForRunLock(t, updater)
}

// `stem update` keeps the run lock through the restart it causes, so the new
// daemon finds it held at boot. The reflector must wait for it rather than
// give up (the operator asked for it at boot) or start under the update.
func TestAutostartWaitsForTheUpdateToFinish(t *testing.T) {
	s, updater := newRunLockServer(t)
	s.reflectorConfig = ReflectorConfig{
		Profile:   "netally",
		Interface: "stem-nonexistent-if0",
		Autostart: true,
	}
	releaseUpdate, err := updater.Acquire()
	if err != nil {
		t.Fatalf("updater Acquire: %v", err)
	}

	s.autostartReflector(t.Context())

	s.statsMu.RLock()
	before := s.testRunID
	s.statsMu.RUnlock()
	time.Sleep(2 * runLockRetry)
	s.statsMu.RLock()
	during := s.testRunID
	s.statsMu.RUnlock()
	if during != before {
		t.Fatalf("autostart began a run (runID %d -> %d) while an update held the run lock", before, during)
	}

	releaseUpdate()
	deadline := time.Now().Add(5 * runLockRetry)
	for {
		s.statsMu.RLock()
		after := s.testRunID
		s.statsMu.RUnlock()
		if after != before {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("autostart never tried the reflector after the update let go")
		}
		time.Sleep(10 * time.Millisecond)
	}
	// The interface does not exist, so the start fails and must let go too.
	waitForRunLock(t, updater)
}

// Every result names the build that produced it: results from either side of
// an update must be distinguishable.
func TestResultNamesTheBuildThatProducedIt(t *testing.T) {
	t.Cleanup(version.SetForTesting("v0.27.15", "1cae4f9", "2026-10-17T00:00:00Z"))
	s, _ := newRunLockServer(t)

	s.statsMu.Lock()
	result := &TestResultResponse{Status: statusCompleted}
	s.stampRunRecordLocked(result)
	s.statsMu.Unlock()

	if result.StemVersion != "v0.27.15" || result.StemCommit != "1cae4f9" {
		t.Errorf("result names stem %q (%q), want v0.27.15 (1cae4f9)", result.StemVersion, result.StemCommit)
	}
}

// waitForRunLock takes and drops the run lock as the updater. A plan's hold is
// released by its worker goroutine just after the status turns terminal, so
// the lock is polled rather than read once.
func waitForRunLock(t *testing.T, updater *update.RunGuard) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		release, err := updater.Acquire()
		if err == nil {
			release()
			return
		}
		if !errors.Is(err, update.ErrRunLockHeld) || time.Now().After(deadline) {
			t.Fatalf("updater Acquire = %v; the daemon kept the run lock", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
