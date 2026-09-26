// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

func TestClassifyRunCause(t *testing.T) {
	tests := []struct {
		name  string
		cause string
		want  string
	}{
		{"bind conflict", "bind /var/run/stem.sock: address already in use", causeInterfaceBusy},
		{"busy device", "pcap open eth0: device or resource busy", causeInterfaceBusy},
		{"raw socket denied", "open /dev/bpf0: operation not permitted", causeNotPermitted},
		{"capability denied", "socket(AF_PACKET): permission denied", causeNotPermitted},
		{"interface gone", "send on eth3: no such device", causeInterfaceMissing},
		{"address gone", "cannot assign requested address", causeInterfaceMissing},
		{"peer silent", "dial 10.44.40.23:862: i/o timeout", causeUnreachable},
		{"peer refusing", "connection refused", causeUnreachable},
		{
			"throughput search with nothing reflected",
			"benchmark test rfc2544_throughput failed: " + dataplane.ErrThroughputNoFramesReturned.Error(),
			causeUnreachable,
		},
		{
			"throughput search lossy at every rate",
			"benchmark test rfc2544_throughput failed: " + dataplane.ErrThroughputNoPassingRate.Error(),
			causeCriteriaNotMet,
		},
		{"unrecognised", "unexpected EOF while reading frame 9", causeGeneric},
		{"no cause at all", "", causeGeneric},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyRunCause(tc.cause); got != tc.want {
				t.Errorf("classifyRunCause(%q) = %q, want %q", tc.cause, got, tc.want)
			}
		})
	}
}

// Whatever the cause text carries, the classification must not repeat it: the
// raw error names interfaces, socket paths and peer addresses.
func TestClassifyRunCauseNeverEchoesItsInput(t *testing.T) {
	secrets := []string{"/var/run/stem.sock", "10.44.40.23", "eth3", "/dev/bpf0"}
	for _, secret := range secrets {
		got := classifyRunCause("failed on " + secret + ": address already in use")
		if strings.Contains(got, secret) {
			t.Errorf("classifyRunCause echoed %q in %q", secret, got)
		}
	}
}

// respondTestExecutionError is the synchronous failure site — the one a
// reflector start reaches, which is the surface the UI headline belongs to.
func TestRespondTestExecutionErrorRecordsTheCause(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "causeuser")
	t.Setenv("STEM_AUTH_PASSWORD", "causepass123456")
	s := newTestServer(t)
	s.respondTestExecutionError(
		httptest.NewRecorder(),
		errors.New("open /dev/bpf0: operation not permitted"),
		"reflector",
		testTypeReflect,
	)
	if got := s.snapshotStats().ErrorMessage; got != causeNotPermitted {
		t.Errorf("stats.errorMessage = %q, want %q", got, causeNotPermitted)
	}
}

// The reflector's boot path records the cause too: an autostart that fails is
// precisely the case where nobody is watching the daemon log, and the operator
// meets the failure later through the UI.
func TestAutostartReflectorRecordsTheCause(t *testing.T) {
	s, _ := newReflectorStateServer(t)
	s.reflectorConfig = ReflectorConfig{
		Profile:   "netally",
		Interface: "stem-nonexistent-if0",
		Autostart: true,
	}

	s.autostartReflector()

	s.statsMu.RLock()
	defer s.statsMu.RUnlock()
	if s.testStatus != statusError {
		t.Skipf("autostart reached %q on this host; the failure path is what this test covers", s.testStatus)
	}
	if s.testError == "" {
		t.Error("testError is empty after a failed autostart")
	}
	if strings.Contains(s.testError, "stem-nonexistent-if0") {
		t.Errorf("testError = %q names the raw interface", s.testError)
	}
}

// The asynchronous failure site is the run plan: a step whose executor
// returns an error must classify the cause the same way the synchronous
// reflector start does (#1251). Nothing drove this path before: the
// bookkeeping that WAS covered lived in the module-executor path deleted
// in #1332, which no production caller reached.
func TestPlanStepFailureRecordsTheCause(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "runmodcauseuser")
	t.Setenv("STEM_AUTH_PASSWORD", "runmodcausepass123")
	s := newTestServer(t)

	plan, err := newRunPlan("", TestStartRequest{
		Peer:  "198.51.100.7",
		Tests: []TestStepRequest{{TestType: "rfc2544_throughput"}},
	})
	if err != nil {
		t.Fatalf("newRunPlan: %v", err)
	}
	exec := &causeExecutor{err: errors.New("bind eth0: address already in use")}
	s.executorResolver = func(string) (executorFactory, bool) {
		return func(string) (testExecutor, error) { return exec, nil }, true
	}

	runID, beginErr := s.beginRunPlan(plan)
	if beginErr != nil {
		t.Fatalf("beginRunPlan: %v", beginErr)
	}
	s.runTestPlan(runID, "lo0")

	s.statsMu.RLock()
	status, cause, stepStatus := s.testStatus, s.testError, s.runPlan.Steps[0].Status
	s.statsMu.RUnlock()
	if status != statusError {
		t.Errorf("testStatus = %q, want %q", status, statusError)
	}
	if cause != causeInterfaceBusy {
		t.Errorf("testError = %q, want %q", cause, causeInterfaceBusy)
	}
	if stepStatus != stepFailed {
		t.Errorf("step status = %q, want %q", stepStatus, stepFailed)
	}
}

type causeExecutor struct{ err error }

func (*causeExecutor) Close() {}

func (e *causeExecutor) Execute(string, *modtypes.TestConfig) (*modtypes.Result, error) {
	return nil, e.err
}
