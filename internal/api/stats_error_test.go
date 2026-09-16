// SPDX-License-Identifier: BUSL-1.1

package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/netif"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// failingExecutor fails every test with an error carrying the kind of detail a
// real bind failure does: a device name and a socket path the wire must not
// repeat back.
type failingExecutor struct{ err error }

func (*failingExecutor) Close() {}

func (e *failingExecutor) Execute(string, *modtypes.TestConfig) (*modtypes.Result, error) {
	return nil, e.err
}

func statsSnapshot(t *testing.T, s *api.Server, token string) api.Stats {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	var stats api.Stats
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	return stats
}

// awaitStatsStatus polls /api/v1/stats until testStatus matches, and returns
// that snapshot. The condition is evaluated before the deadline is tested, so a
// run that finishes inside the first tick is still seen.
func awaitStatsStatus(t *testing.T, s *api.Server, token, want string) api.Stats {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		stats := statsSnapshot(t, s, token)
		if stats.TestStatus == want {
			return stats
		}
		if time.Now().After(deadline) {
			t.Fatalf("testStatus = %q, want %q within the deadline", stats.TestStatus, want)
		}
		time.Sleep(time.Millisecond)
	}
}

func startSingleTest(t *testing.T, s *api.Server, token, iface string) {
	t.Helper()
	body := bytes.NewBufferString(
		`{"peer":"192.0.2.1","tests":[{"testType":"rfc2544_throughput"}],"interface":"` + iface + `"}`,
	)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("start status = %d: %s", w.Code, w.Body.String())
	}
}

func testInterface(t *testing.T) string {
	t.Helper()
	ifaces, err := netif.DetectInterfaces()
	if err != nil || len(ifaces) == 0 {
		t.Skip("no network interface available")
	}
	return ifaces[0].Name
}

// A failed run must tell the operator WHY on the wire. The UI headline is
// stats.errorMessage (ReflectorPage), and before #1251 the field existed only
// in TypeScript: the daemon logged the cause and the operator saw the generic
// "stopped with an error" for a bind failure they could have fixed in seconds.
func TestStatsCarriesTheFailureCause(t *testing.T) {
	s := setupTestingTestServer(t)
	iface := testInterface(t)
	s.UseTestExecutorResolver(func(string) (api.TestExecutorFactory, bool) {
		return func(string) (api.TestExecutor, error) {
			return &failingExecutor{err: errors.New("bind /var/run/stem-42.sock: address already in use")}, nil
		}, true
	})
	t.Cleanup(func() { s.UseTestExecutorResolver(nil) })
	token := getTestingAuthToken(t, s)

	startSingleTest(t, s, token, iface)
	stats := awaitStatsStatus(t, s, token, "error")

	if stats.ErrorMessage == "" {
		t.Fatal("stats.errorMessage is empty after a failed run; the UI can only show its generic headline")
	}
	if !strings.Contains(strings.ToLower(stats.ErrorMessage), "already in use") {
		t.Errorf("errorMessage = %q, want the classified in-use cause", stats.ErrorMessage)
	}
	// Sanitised, not echoed: the raw error names a socket path, and a path on
	// the daemon host is not the operator's business and not stable API.
	if strings.Contains(stats.ErrorMessage, "/var/run/stem-42.sock") {
		t.Errorf("errorMessage = %q leaks the raw error's path", stats.ErrorMessage)
	}
}

// An unrecognised failure must still say something truthful rather than
// echoing arbitrary text onto the page.
func TestStatsFailureCauseIsAClosedSet(t *testing.T) {
	s := setupTestingTestServer(t)
	iface := testInterface(t)
	s.UseTestExecutorResolver(func(string) (api.TestExecutorFactory, bool) {
		return func(string) (api.TestExecutor, error) {
			return &failingExecutor{err: errors.New("unexpected EOF from 10.44.40.23 while reading frame 9")}, nil
		}, true
	})
	t.Cleanup(func() { s.UseTestExecutorResolver(nil) })
	token := getTestingAuthToken(t, s)

	startSingleTest(t, s, token, iface)
	stats := awaitStatsStatus(t, s, token, "error")

	if stats.ErrorMessage == "" {
		t.Fatal("stats.errorMessage is empty for an unclassified failure")
	}
	if strings.Contains(stats.ErrorMessage, "10.44.40.23") {
		t.Errorf("errorMessage = %q leaks the raw error's address", stats.ErrorMessage)
	}
}

// The cause belongs to the run that produced it. A stale message on the next
// run would be worse than none: it describes a failure that is not happening.
func TestStatsFailureCauseClearsOnTheNextRun(t *testing.T) {
	s := setupTestingTestServer(t)
	iface := testInterface(t)
	fail := &failingExecutor{err: errors.New("operation not permitted")}
	s.UseTestExecutorResolver(func(string) (api.TestExecutorFactory, bool) {
		return func(string) (api.TestExecutor, error) { return fail, nil }, true
	})
	t.Cleanup(func() { s.UseTestExecutorResolver(nil) })
	token := getTestingAuthToken(t, s)

	startSingleTest(t, s, token, iface)
	if stats := awaitStatsStatus(t, s, token, "error"); stats.ErrorMessage == "" {
		t.Fatal("stats.errorMessage is empty after the first failure")
	}

	blocked := make(chan struct{})
	s.UseTestExecutorResolver(func(string) (api.TestExecutorFactory, bool) {
		return func(string) (api.TestExecutor, error) { return &blockingExecutor{release: blocked}, nil }, true
	})
	startSingleTest(t, s, token, iface)
	stats := awaitStatsStatus(t, s, token, "running")
	close(blocked)
	if stats.ErrorMessage != "" {
		t.Errorf("errorMessage = %q on a run that is still running", stats.ErrorMessage)
	}
}

// A multi-step run reports the cause of the step that stopped it: the run-plan
// path keeps its own failure bookkeeping (run_plan.go) and would otherwise
// leave the field empty on exactly the runs that take longest to repeat.
func TestStatsCarriesTheFailureCauseForARunPlan(t *testing.T) {
	s := setupTestingTestServer(t)
	iface := testInterface(t)
	s.UseTestExecutorResolver(func(string) (api.TestExecutorFactory, bool) {
		return func(string) (api.TestExecutor, error) {
			return &failingExecutor{err: errors.New("send on eth3: no such device")}, nil
		}, true
	})
	t.Cleanup(func() { s.UseTestExecutorResolver(nil) })
	token := getTestingAuthToken(t, s)

	body := bytes.NewBufferString(`{"peer":"192.0.2.1","tests":[
		{"testType":"rfc2544_throughput"},
		{"testType":"rfc2544_latency"}
	],"interface":"` + iface + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/start", body)
	authorizeWithCSRF(t, s, req, token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("start status = %d: %s", w.Code, w.Body.String())
	}

	stats := awaitStatsStatus(t, s, token, "error")
	if !strings.Contains(stats.ErrorMessage, "not available") {
		t.Errorf("errorMessage = %q, want the missing-interface cause", stats.ErrorMessage)
	}
	if strings.Contains(stats.ErrorMessage, "eth3") {
		t.Errorf("errorMessage = %q leaks the raw error's device name", stats.ErrorMessage)
	}
}

// blockingExecutor stays in Execute until released, so the test can observe the
// running state rather than racing the result.
type blockingExecutor struct{ release chan struct{} }

func (*blockingExecutor) Close() {}

func (e *blockingExecutor) Execute(testType string, _ *modtypes.TestConfig) (*modtypes.Result, error) {
	<-e.release
	return &modtypes.Result{TestType: testType, Success: true}, nil
}
