// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
	"github.com/MustardSeedNetworks/stem/internal/license"
)

// fakeLicenseDaemon records what the CLI asked the daemon for. The daemon
// owning the licence is the whole point of #1335, so what matters is that
// the request arrived at all — not what the CLI would have computed itself.
type fakeLicenseDaemon struct {
	mu       sync.Mutex
	calls    []string
	activate []string
}

func (d *fakeLicenseDaemon) record(method, path string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls = append(d.calls, method+" "+path)
}

func (d *fakeLicenseDaemon) sequence() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.calls...)
}

func (d *fakeLicenseDaemon) keys() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.activate...)
}

// publishFakeLicenseDaemon stands a daemon up on the search path, so
// `stem license` discovers it exactly as it would a real one.
func publishFakeLicenseDaemon(t *testing.T) *fakeLicenseDaemon {
	t.Helper()

	daemon := &fakeLicenseDaemon{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/auth/csrf-token":
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "csrf-value"})
		case "/api/v1/license":
			daemon.record(r.Method, r.URL.Path)
			switch r.Method {
			case http.MethodDelete:
				_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "message": "License deactivated"})
			default:
				_ = json.NewEncoder(w).Encode(map[string]any{
					"activated": true, "isTrialMode": true, "tier": 2, "tierName": "Trial",
					"daysRemaining": 9, "deviceHash": "daemon-device", "message": "Trial mode: 9 days remaining",
				})
			}
		case "/api/v1/license/activate":
			daemon.record(r.Method, r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			var req struct {
				LicenseKey string `json:"licenseKey"`
			}
			_ = json.Unmarshal(body, &req)
			daemon.mu.Lock()
			daemon.activate = append(daemon.activate, req.LicenseKey)
			daemon.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": false, "message": "Invalid license key format",
			})
		case "/api/v1/license/trial":
			daemon.record(r.Method, r.URL.Path)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true, "message": "Trial started", "daysRemaining": 14, "tier": 2,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	caFile := filepath.Join(dir, "server.crt")
	encoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(caFile, encoded, 0o600); err != nil {
		t.Fatalf("write certificate: %v", err)
	}
	if err := daemonconn.Publish(dir, daemonconn.Descriptor{
		URL: srv.URL, Token: "fake-token", CAFile: caFile,
	}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	t.Setenv("STEM_DATA_DIR", dir)
	return daemon
}

// emptyDataDir points discovery at a directory with no descriptor, which is
// how a host with no daemon running looks.
func emptyDataDir(t *testing.T) {
	t.Helper()
	t.Setenv("STEM_DATA_DIR", t.TempDir())
}

// TestLicenseTrialGoesThroughTheRunningDaemon is the defect: `stem license
// --trial` wrote the licence file in-process while the daemon held a cached
// manager, so the daemon never saw the change until it restarted.
func TestLicenseTrialGoesThroughTheRunningDaemon(t *testing.T) {
	licenseHome(t)
	daemon := publishFakeLicenseDaemon(t)

	out := captureStdout(t, func() {
		if err := licenseCmd([]string{"--trial"}); err != nil {
			t.Errorf("licenseCmd --trial: %v", err)
		}
	})

	if got := daemon.sequence(); len(got) != 1 || got[0] != "POST /api/v1/license/trial" {
		t.Fatalf("daemon calls = %v, want exactly POST /api/v1/license/trial", got)
	}
	if _, err := os.Stat(license.DefaultLicensePath()); err == nil {
		t.Errorf("the CLI wrote %s while a daemon was running; the daemon owns that file",
			license.DefaultLicensePath())
	}
	if !strings.Contains(out, "Trial started") {
		t.Errorf("the daemon's answer was not reported:\n%s", out)
	}
}

// TestLicenseActivateSendsTheKeyToTheDaemon proves the request reaches the
// daemon carrying the key. Stem has no test signing key, so the daemon's
// refusal is the answer here — what matters is whose decision it is.
func TestLicenseActivateSendsTheKeyToTheDaemon(t *testing.T) {
	licenseHome(t)
	daemon := publishFakeLicenseDaemon(t)

	out := captureStdout(t, func() {
		err := licenseCmd([]string{"--activate", "MSN1.payload.signature"})
		if err == nil {
			t.Error("licenseCmd --activate: want the daemon's refusal, got nil")
		}
	})

	if got := daemon.keys(); len(got) != 1 || got[0] != "MSN1.payload.signature" {
		t.Fatalf("keys the daemon received = %v, want the one the operator typed", got)
	}
	if !strings.Contains(out, "Invalid license key format") {
		t.Errorf("the daemon's message was not reported verbatim:\n%s", out)
	}
}

// TestLicenseStatusReadsTheDaemonNotTheFile: a CLI reading the file while the
// daemon holds a newer state in memory reports a licence that is not in force.
func TestLicenseStatusReadsTheDaemonNotTheFile(t *testing.T) {
	licenseHome(t)
	daemon := publishFakeLicenseDaemon(t)

	out := captureStdout(t, func() {
		if err := licenseCmd([]string{"--status"}); err != nil {
			t.Errorf("licenseCmd --status: %v", err)
		}
	})

	if got := daemon.sequence(); len(got) != 1 || got[0] != "GET /api/v1/license" {
		t.Fatalf("daemon calls = %v, want exactly GET /api/v1/license", got)
	}
	if !strings.Contains(out, "daemon-device") {
		t.Errorf("status came from somewhere other than the daemon:\n%s", out)
	}
}

// TestLicenseDeactivateGoesThroughTheDaemon closes the same hole for the one
// verb that removes entitlement: leaving it offline would let the CLI delete
// the file under a daemon that goes on serving the cached licence.
func TestLicenseDeactivateGoesThroughTheDaemon(t *testing.T) {
	licenseHome(t)
	daemon := publishFakeLicenseDaemon(t)

	out := captureStdout(t, func() {
		if err := licenseCmd([]string{"--deactivate"}); err != nil {
			t.Errorf("licenseCmd --deactivate: %v", err)
		}
	})

	if got := daemon.sequence(); len(got) != 1 || got[0] != "DELETE /api/v1/license" {
		t.Fatalf("daemon calls = %v, want exactly DELETE /api/v1/license", got)
	}
	if !strings.Contains(out, "License deactivated") {
		t.Errorf("the daemon's answer was not reported:\n%s", out)
	}
}

// The row's table: a descriptor that is found, absent, or present but
// unusable. Only the middle case writes locally; an unusable descriptor is an
// error, never a quiet fallback to the file the daemon is holding open.

// TestLicenseDescriptorFoundLeavesTheFileAlone: with a daemon on the search
// path the CLI must not write the licence the daemon is serving from memory.
func TestLicenseDescriptorFoundLeavesTheFileAlone(t *testing.T) {
	licenseHome(t)
	daemon := publishFakeLicenseDaemon(t)

	captureStdout(t, func() {
		if err := licenseCmd([]string{"--trial"}); err != nil {
			t.Errorf("licenseCmd: %v", err)
		}
	})

	if got := daemon.sequence(); len(got) != 1 {
		t.Errorf("daemon calls = %v, want one", got)
	}
	if _, err := os.Stat(license.DefaultLicensePath()); err == nil {
		t.Error("the CLI wrote the licence file although a daemon was running")
	}
}

// TestLicenseDescriptorNotFoundWritesOffline: with no daemon anywhere on the
// search path this process is the only writer, so it writes the file itself.
func TestLicenseDescriptorNotFoundWritesOffline(t *testing.T) {
	licenseHome(t)
	emptyDataDir(t)

	captureStdout(t, func() {
		if err := licenseCmd([]string{"--trial"}); err != nil {
			t.Errorf("licenseCmd: %v", err)
		}
	})

	if _, err := os.Stat(license.DefaultLicensePath()); err != nil {
		t.Errorf("with no daemon running the CLI must write the licence itself: %v", err)
	}
}

// TestLicenseDescriptorUnreadableIsAnError: a descriptor that exists but
// cannot be read means a daemon IS running, so falling back to a local write
// would be the very race this fix removes.
func TestLicenseDescriptorUnreadableIsAnError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads a 0000 descriptor; the case cannot be built")
	}
	licenseHome(t)
	dir := t.TempDir()
	if err := daemonconn.Publish(dir, daemonconn.Descriptor{
		URL: "https://127.0.0.1:9", Token: "tok",
	}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if err := os.Chmod(daemonconn.Path(dir), 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Setenv("STEM_DATA_DIR", dir)

	captureStdout(t, func() {
		if err := licenseCmd([]string{"--trial"}); err == nil {
			t.Error("an unreadable descriptor must be an error, not a silent local write")
		}
	})

	if _, err := os.Stat(license.DefaultLicensePath()); err == nil {
		t.Error("the CLI wrote the licence file after failing to reach the daemon")
	}
}
