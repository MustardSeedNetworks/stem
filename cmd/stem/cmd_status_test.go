// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bytes"
	"encoding/json"
	"encoding/pem"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

// daemonStub is a TLS server standing in for the daemon, with the certificate
// file a client would be given to trust it.
type daemonStub struct {
	url    string
	caFile string
}

type daemonBehavior struct {
	versionStatus int
	versionBody   string
	readyStatus   int
	readyBody     string
}

const stubVersion = `{"version":"v1.2.3","commit":"abc1234","buildTime":"2026-11-21T00:00:00Z",` +
	`"uiBuildHash":"cafe","releaseTrain":"2026.11","tlsFingerprint":"AA:BB"}`

func healthyDaemon() daemonBehavior {
	return daemonBehavior{
		versionStatus: http.StatusOK,
		versionBody:   stubVersion,
		readyStatus:   http.StatusOK,
		readyBody:     `{"status":"ready","checks":{"auth":{"status":"ok"}}}`,
	}
}

func notReadyDaemon() daemonBehavior {
	behavior := healthyDaemon()
	behavior.readyStatus = http.StatusServiceUnavailable
	behavior.readyBody = `{"status":"not_ready","checks":{"auth":{"status":"error","error":"auth manager not initialized"}}}`
	return behavior
}

func brokenDaemon() daemonBehavior {
	behavior := healthyDaemon()
	behavior.versionStatus = http.StatusInternalServerError
	behavior.versionBody = "boom"
	return behavior
}

func startStubDaemon(t *testing.T, behavior daemonBehavior) daemonStub {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/__version", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(behavior.versionStatus)
		_, _ = w.Write([]byte(behavior.versionBody))
	})
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(behavior.readyStatus)
		_, _ = w.Write([]byte(behavior.readyBody))
	})
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)

	caFile := filepath.Join(t.TempDir(), "server.crt")
	encoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(caFile, encoded, 0o600); err != nil {
		t.Fatalf("write certificate: %v", err)
	}
	return daemonStub{url: srv.URL, caFile: caFile}
}

// closedURL is an address nothing listens on: a daemon that has stopped.
func closedURL(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	address := listener.Addr().String()
	if closeErr := listener.Close(); closeErr != nil {
		t.Fatalf("close: %v", closeErr)
	}
	return "https://" + address
}

func publishDescriptor(t *testing.T, dir string, descriptor daemonconn.Descriptor) {
	t.Helper()

	if err := daemonconn.Publish(dir, descriptor); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	t.Setenv("STEM_DATA_DIR", dir)
}

func runStatus(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var out, errOut bytes.Buffer
	err := statusCmd(args, &out, &errOut)
	return out.String(), errOut.String(), err
}

type statusCase struct {
	name       string
	descriptor *daemonconn.Descriptor
	args       []string
	wantCode   int
	wantState  string
	wantOut    string
	wantErr    string
}

func statusCases(t *testing.T) []statusCase {
	t.Helper()

	healthy := startStubDaemon(t, healthyDaemon())
	notReady := startStubDaemon(t, notReadyDaemon())
	broken := startStubDaemon(t, brokenDaemon())
	closed := closedURL(t)

	return []statusCase{
		{
			name:       "running and ready, found through the descriptor",
			descriptor: &daemonconn.Descriptor{URL: healthy.url, Token: "token-value", CAFile: healthy.caFile},
			wantCode:   exitOK, wantState: stateRunning, wantOut: "Status:      running",
		},
		{
			name:     "running and ready, named with flags",
			args:     []string{"--api", healthy.url, "--cacert", healthy.caFile},
			wantCode: exitOK, wantState: stateRunning, wantOut: "Version:     v1.2.3 (commit abc1234",
		},
		{
			name:     "up but not ready names the failing check",
			args:     []string{"--api", notReady.url, "--cacert", notReady.caFile},
			wantCode: statusNotRunning, wantState: stateNotReady, wantOut: "auth: auth manager not initialized",
		},
		{
			name:     "nothing listening",
			args:     []string{"--api", closed},
			wantCode: statusNotRunning, wantState: stateNotRunning, wantOut: "Status: not running",
		},
		{
			name:       "a stale descriptor left by a crashed daemon is not running",
			descriptor: &daemonconn.Descriptor{URL: closed, Token: "token-value"},
			wantCode:   statusNotRunning, wantState: stateNotRunning, wantOut: "Status: not running",
		},
		{
			name:     "a certificate this host does not trust is an error, not not-running",
			args:     []string{"--api", healthy.url},
			wantCode: statusError, wantState: stateError, wantErr: "certificate",
		},
		{
			name:     "a daemon answering 500 is an error",
			args:     []string{"--api", broken.url, "--cacert", broken.caFile},
			wantCode: statusError, wantState: stateError, wantErr: "answered 500",
		},
		{
			name:     "an unreadable certificate file is an error",
			args:     []string{"--api", healthy.url, "--cacert", filepath.Join(t.TempDir(), "missing.crt")},
			wantCode: statusError, wantState: stateError, wantErr: "read daemon certificate",
		},
	}
}

func (tc statusCase) run(t *testing.T, extra ...string) (string, string, error) {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	if tc.descriptor != nil {
		publishDescriptor(t, dir, *tc.descriptor)
	}
	return runStatus(t, append(extra, tc.args...)...)
}

func TestStatusCmd(t *testing.T) {
	for _, tc := range statusCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := tc.run(t)

			if got := exitCodeOf(err); got != tc.wantCode {
				t.Fatalf("exit status = %d, want %d (err %v)", got, tc.wantCode, err)
			}
			if tc.wantOut != "" && !strings.Contains(stdout, tc.wantOut) {
				t.Errorf("stdout %q lacks %q", stdout, tc.wantOut)
			}
			if tc.wantErr != "" && !strings.Contains(stderr, tc.wantErr) {
				t.Errorf("stderr %q lacks %q", stderr, tc.wantErr)
			}
		})
	}
}

// --json prints one object whatever happened, and nothing on stderr.
func TestStatusCmdJSON(t *testing.T) {
	for _, tc := range statusCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := tc.run(t, "--json")

			if got := exitCodeOf(err); got != tc.wantCode {
				t.Fatalf("exit status = %d, want %d (err %v)", got, tc.wantCode, err)
			}
			var report statusReport
			if decodeErr := json.Unmarshal([]byte(stdout), &report); decodeErr != nil {
				t.Fatalf("--json did not print an object: %v\n%s", decodeErr, stdout)
			}
			if report.State != tc.wantState {
				t.Errorf("state = %q, want %q", report.State, tc.wantState)
			}
			if wantRunning := tc.wantState == stateRunning ||
				tc.wantState == stateNotReady; report.Running != wantRunning {
				t.Errorf("running = %v for state %q", report.Running, report.State)
			}
			if stderr != "" {
				t.Errorf("--json wrote to stderr: %q", stderr)
			}
			if !strings.Contains(report.Error, tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", report.Error, tc.wantErr)
			}
		})
	}
}

// A descriptor this account may not use is an error. Reporting it as "not
// running" would send an operator to restart a service that is fine.
func TestStatusDescriptorThisAccountCannotUse(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows carries no POSIX mode bits")
	}
	dir := t.TempDir()
	publishDescriptor(t, dir, daemonconn.Descriptor{URL: closedURL(t), Token: "token-value"})
	if err := os.Chmod(daemonconn.Path(dir), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := runStatus(t, "--json")

	if got := exitCodeOf(err); got != statusError {
		t.Fatalf("exit status = %d, want %d", got, statusError)
	}
	var report statusReport
	if decodeErr := json.Unmarshal([]byte(stdout), &report); decodeErr != nil {
		t.Fatalf("--json did not print an object: %v", decodeErr)
	}
	if report.State != stateError || !strings.Contains(report.Error, "--cacert") {
		t.Errorf("report = %+v", report)
	}
	if stderr != "" {
		t.Errorf("stderr = %q", stderr)
	}
}

// A host with no descriptor has no daemon to ask. It must not guess a port: a
// sibling product's fallback port could answer.
func TestStatusWithoutADescriptorIsNotRunningAndDialsNothing(t *testing.T) {
	t.Setenv("STEM_DATA_DIR", t.TempDir())

	target, found, err := statusTarget("", "")
	if err != nil || found {
		t.Fatalf("statusTarget = %+v, %v, %v; want no target", target, found, err)
	}
	_, _, runErr := runStatus(t)
	if got := exitCodeOf(runErr); got != statusNotRunning {
		t.Fatalf("exit status = %d, want %d", got, statusNotRunning)
	}
}

func TestStatusCmdRejectsBadInvocations(t *testing.T) {
	for _, args := range [][]string{{"extra"}, {"--nope"}, {"--api"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			_, _, err := runStatus(t, args...)
			if got := exitCodeOf(err); got != statusError {
				t.Fatalf("exit status = %d, want %d (err %v)", got, statusError, err)
			}
		})
	}
}

func TestStatusCmdHelpIsNotAnError(t *testing.T) {
	_, stderr, err := runStatus(t, "--help")
	if err != nil {
		t.Fatalf("--help = %v", err)
	}
	if !strings.Contains(stderr, "Exit status:") {
		t.Errorf("help does not state the exit statuses: %q", stderr)
	}
}
