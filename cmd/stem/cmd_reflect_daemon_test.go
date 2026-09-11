// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

// fakeDaemon records the sequence of calls the CLI makes, which is the only
// way to prove the reflector was configured before it was started.
type fakeDaemon struct {
	mu       sync.Mutex
	calls    []string
	configs  []api.ReflectorConfig
	startReq api.TestStartRequest

	// started closes once the start request has been recorded, so a test
	// ends the watch loop after the run began rather than during it.
	started chan struct{}
}

func (d *fakeDaemon) record(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls = append(d.calls, path)
}

func (d *fakeDaemon) sequence() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.calls...)
}

func newFakeDaemonClient(t *testing.T) (*daemonclient.Client, *fakeDaemon) {
	t.Helper()

	daemon := &fakeDaemon{started: make(chan struct{})}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/auth/csrf-token":
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "csrf-value"})
		case "/api/v1/reflector/config":
			daemon.record(r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			var cfg api.ReflectorConfig
			_ = json.Unmarshal(body, &cfg)
			daemon.mu.Lock()
			daemon.configs = append(daemon.configs, cfg)
			daemon.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
		case "/api/v1/test/start":
			daemon.record(r.URL.Path)
			body, _ := io.ReadAll(r.Body)
			daemon.mu.Lock()
			_ = json.Unmarshal(body, &daemon.startReq)
			daemon.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "started", "suiteId": "stem-fake-1"})
			close(daemon.started)
		case "/api/v1/test/stop":
			daemon.record(r.URL.Path)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
		case "/api/v1/reflector/stats":
			daemon.record(r.URL.Path)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"running": true, "packetsReceived": 5, "packetsReflected": 5,
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

	client, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return client, daemon
}

// --oui and --port are reflector configuration. The daemon refuses a
// configuration change while the reflector runs, so they must be applied
// before the start — otherwise both flags parse and then do nothing.
func TestRunReflectorConfiguresBeforeStarting(t *testing.T) {
	client, daemon := newFakeDaemonClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cancelOnceStarted(daemon, cancel)

	out := captureStdout(t, func() {
		if err := runReflector(ctx, client, &reflectCmdArgs{
			iface: "eth0", profile: "netally", oui: "00:c0:17",
		}); err != nil {
			t.Errorf("runReflector: %v", err)
		}
	})

	got := daemon.sequence()
	if len(got) < 2 || got[0] != "/api/v1/reflector/config" || got[1] != "/api/v1/test/start" {
		t.Fatalf("call sequence = %v, want config then start", got)
	}

	daemon.mu.Lock()
	defer daemon.mu.Unlock()
	if len(daemon.configs) != 1 {
		t.Fatalf("configs = %+v, want exactly one", daemon.configs)
	}
	cfg := daemon.configs[0]
	if cfg.OUIFilter != "00:c0:17" {
		t.Errorf("ouiFilter = %q, want 00:c0:17", cfg.OUIFilter)
	}
	if cfg.Profile != "netally" {
		t.Errorf("profile = %q, want netally", cfg.Profile)
	}
	// The NetAlly profile's own port, which CT307 reflects on.
	if cfg.PortFilter != 3842 {
		t.Errorf("portFilter = %d, want 3842", cfg.PortFilter)
	}
	if daemon.startReq.Interface != "eth0" {
		t.Errorf("start interface = %q, want eth0", daemon.startReq.Interface)
	}
	if !strings.Contains(out, "stem-fake-1") {
		t.Errorf("the run ID the daemon issued was not reported:\n%s", out)
	}
}

// Interrupting stops the daemon-owned reflector rather than just abandoning
// it, and reports the counters the daemon ends with.
func TestRunReflectorStopsTheDaemonRunOnInterrupt(t *testing.T) {
	client, daemon := newFakeDaemonClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cancelOnceStarted(daemon, cancel)

	out := captureStdout(t, func() {
		if err := runReflector(ctx, client, &reflectCmdArgs{iface: "eth0", profile: "netally"}); err != nil {
			t.Errorf("runReflector: %v", err)
		}
	})

	if !slices.Contains(daemon.sequence(), "/api/v1/test/stop") {
		t.Errorf("the reflector was never stopped in the daemon: %v", daemon.sequence())
	}
	if !strings.Contains(out, "Final Statistics") || !strings.Contains(out, "Packets Reflected: 5") {
		t.Errorf("final daemon counters were not reported:\n%s", out)
	}
}

// cancelOnceStarted ends the watch after the start response has reached the
// client, standing in for the operator's Ctrl+C.
func cancelOnceStarted(daemon *fakeDaemon, cancel context.CancelFunc) {
	<-daemon.started
	time.Sleep(50 * time.Millisecond)
	cancel()
}
