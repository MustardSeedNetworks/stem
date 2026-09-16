// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
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

// runModuleTest's asynchronous failure branch keeps the same bookkeeping as
// the plan path. NOTE: no production caller reaches it today — executeTest is
// called only from startReflectorRequest, whose module is always the
// reflector, so this branch is exercised here and nowhere else (#1251).
func TestRunModuleTestFailureRecordsTheCause(t *testing.T) {
	t.Setenv("STEM_AUTH_USERNAME", "runmodcauseuser")
	t.Setenv("STEM_AUTH_PASSWORD", "runmodcausepass123")
	s := newTestServer(t)

	exec := &causeExecutor{err: errors.New("bind eth0: address already in use")}
	factory := func(string) (testExecutor, error) { return exec, nil }
	if err := s.runModuleTest(factory, "benchmark", "throughput", "lo0", nil); err != nil {
		t.Fatalf("runModuleTest() error: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		s.statsMu.RLock()
		status, cause := s.testStatus, s.testError
		s.statsMu.RUnlock()
		if status == statusError {
			if cause != causeInterfaceBusy {
				t.Errorf("testError = %q, want %q", cause, causeInterfaceBusy)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("testStatus = %q, want %q within the deadline", status, statusError)
		}
		time.Sleep(time.Millisecond)
	}
}

type causeExecutor struct{ err error }

func (*causeExecutor) Close() {}

func (e *causeExecutor) Execute(string, *modtypes.TestConfig) (*modtypes.Result, error) {
	return nil, e.err
}
