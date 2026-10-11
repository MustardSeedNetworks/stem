// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
	"github.com/MustardSeedNetworks/stem/internal/license"
	"github.com/MustardSeedNetworks/stem/internal/netif"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
)

// sizeRecorder stands in for the dataplane and reports each frame size the
// daemon asks it to measure.
type sizeRecorder struct {
	sizes chan uint32
}

func (*sizeRecorder) Close() {}

func (r *sizeRecorder) Execute(testType string, cfg *modtypes.TestConfig) (*modtypes.Result, error) {
	r.sizes <- cfg.FrameSize
	return &modtypes.Result{TestType: testType, Success: true}, nil
}

// realDaemonClient serves a real api.Server — its handlers, its validation,
// its auth and CSRF — over TLS, with only the dataplane replaced, and returns
// a CLI client bound to it.
func realDaemonClient(t *testing.T, exec api.TestExecutor) (*daemonclient.Client, string) {
	t.Helper()
	const username, password = "cli-test", "cli-test-password-1412"
	t.Setenv("STEM_TEST_MODE", "1")
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	t.Setenv("STEM_AUTH_USERNAME", username)
	t.Setenv("STEM_AUTH_PASSWORD", password)

	ifaces, err := netif.DetectInterfaces()
	if err != nil || len(ifaces) == 0 {
		t.Skip("no network interface available")
	}

	s, err := api.NewServer(api.ListenAddr{Port: 8444})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	mgr, err := license.LoadFromDir(t.TempDir())
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if result := mgr.StartTrial(); !result.Success {
		t.Fatalf("StartTrial: %s", result.Message)
	}
	s.UseLicenseForTest(mgr)
	s.UseDataplaneAvailabilityForTest(func() (bool, string) { return true, "" })
	s.UseTestExecutorResolver(func(string) (api.TestExecutorFactory, bool) {
		return func(string) (api.TestExecutor, error) { return exec, nil }, true
	})

	login, _ := json.Marshal(map[string]string{"username": username, "password": password})
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(login)))
	var session struct {
		Token string `json:"token"`
	}
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &session) != nil || session.Token == "" {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}

	srv := httptest.NewTLSServer(s)
	t.Cleanup(srv.Close)
	dir := t.TempDir()
	caFile := filepath.Join(dir, "server.crt")
	encoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if writeErr := os.WriteFile(caFile, encoded, 0o600); writeErr != nil {
		t.Fatalf("write certificate: %v", writeErr)
	}
	if publishErr := daemonconn.Publish(dir, daemonconn.Descriptor{
		URL: srv.URL, Token: session.Token, CAFile: caFile,
	}); publishErr != nil {
		t.Fatalf("Publish: %v", publishErr)
	}
	client, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return client, ifaces[0].Name
}

// `stem test -t rfc2544_throughput` with no other options is the first thing
// an operator runs. The daemon refused it outright because the CLI's request
// carried a Y.1564 block holding the 64-byte RFC 2544 size (stem#1412); the
// request now has to clear the daemon's own validation and the run has to
// measure that size.
func TestDefaultThroughputRunPassesTheDaemonsValidation(t *testing.T) {
	exec := &sizeRecorder{sizes: make(chan uint32, 16)}
	client, iface := realDaemonClient(t, exec)

	flags := parseTestFlagsOrFail(t, "-i", iface, "--peer", "192.0.2.1", "-t", "rfc2544_throughput")
	seconds, err := validateDuration(flags.duration)
	if err != nil {
		t.Fatalf("validateDuration: %v", err)
	}
	req := buildStartRequest(flags, []string{flags.testTypes}, parseFrameSizes(flags.frameSizes), seconds)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, startErr := client.Start(ctx, req); startErr != nil {
		t.Fatalf("the daemon refused the CLI's default throughput run: %v", startErr)
	}

	select {
	case size := <-exec.sizes:
		if size != 64 {
			t.Errorf("first measured frame size = %d, want the 64-byte default", size)
		}
	case <-ctx.Done():
		t.Fatal("the run started but never reached the dataplane")
	}
}
