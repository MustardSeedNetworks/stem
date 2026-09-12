// SPDX-License-Identifier: BUSL-1.1

package daemonclient_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

const testToken = "published-token-value"

// recorder captures what the client actually put on the wire, which is the
// only way to tell an authenticated request from one the daemon would reject.
type recorder struct {
	mu       sync.Mutex
	requests []*http.Request
}

func (r *recorder) add(req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := req.Clone(context.Background())
	r.requests = append(r.requests, clone)
}

func (r *recorder) all() []*http.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*http.Request(nil), r.requests...)
}

func (r *recorder) last() *http.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.requests) == 0 {
		return nil
	}
	return r.requests[len(r.requests)-1]
}

// newDaemon starts a TLS server standing in for the daemon and publishes a
// descriptor pointing at it, exactly as a running daemon would.
func newDaemon(t *testing.T, handler http.HandlerFunc) (string, *recorder) {
	t.Helper()

	rec := &recorder{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		if r.URL.Path == "/api/v1/auth/csrf-token" {
			writeJSON(t, w, map[string]string{"token": "csrf-value"})
			return
		}
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	caFile := filepath.Join(dir, "server.crt")
	writePEM(t, caFile, srv.Certificate().Raw)

	if err := daemonconn.Publish(dir, daemonconn.Descriptor{
		URL:    srv.URL,
		Token:  testToken,
		CAFile: caFile,
	}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	return dir, rec
}

// Every request must carry the published token; without it the daemon
// answers 401 and the CLI is not a client of anything.
func TestClientAuthenticatesWithThePublishedToken(t *testing.T) {
	dir, rec := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"status": "started", "suiteId": "stem-abc-1"})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, startErr := c.Start(context.Background(), api.TestStartRequest{
		Interface: "eth0",
		Tests:     []api.TestStepRequest{{TestType: "rfc2544_throughput"}},
	}); startErr != nil {
		t.Fatalf("Start: %v", startErr)
	}

	got := rec.last()
	if want := "Bearer " + testToken; got.Header.Get("Authorization") != want {
		t.Errorf("Authorization = %q, want %q", got.Header.Get("Authorization"), want)
	}
}

// Mutating routes are CSRF-protected, so a client that does not fetch and
// present the token can start nothing.
func TestClientPresentsACSRFTokenOnMutatingRequests(t *testing.T) {
	dir, rec := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"status": "started", "suiteId": "stem-abc-1"})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, startErr := c.Start(context.Background(), api.TestStartRequest{
		Interface: "eth0",
		Tests:     []api.TestStepRequest{{TestType: "rfc2544_throughput"}},
	}); startErr != nil {
		t.Fatalf("Start: %v", startErr)
	}

	if got := rec.last().Header.Get("X-Csrf-Token"); got != "csrf-value" {
		t.Errorf("X-Csrf-Token = %q, want %q", got, "csrf-value")
	}
}

// The run ID the daemon issued is what the operator and the web UI see, so
// the client reports it rather than inventing one.
func TestStartReturnsTheDaemonsRunID(t *testing.T) {
	dir, _ := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"status": "started", "suiteId": "stem-9f8e-3"})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	runID, err := c.Start(context.Background(), api.TestStartRequest{
		Interface: "eth0",
		Tests:     []api.TestStepRequest{{TestType: "rfc2544_throughput"}},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if runID != "stem-9f8e-3" {
		t.Errorf("run ID = %q, want %q", runID, "stem-9f8e-3")
	}
}

// A tier refusal is the operator's answer, not a generic failure: the CLI
// has to be able to name the feature the daemon refused.
func TestStartSurfacesTheFeatureGate(t *testing.T) {
	dir, _ := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		writeJSON(t, w, map[string]any{
			"error": map[string]any{"code": "TIER_TOO_LOW", "feature": "rfc2544", "message": "Professional required"},
		})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_, startErr := c.Start(context.Background(), api.TestStartRequest{
		Interface: "eth0",
		Tests:     []api.TestStepRequest{{TestType: "rfc2544_throughput"}},
	})

	var gate *daemonclient.FeatureGateError
	if !errors.As(startErr, &gate) {
		t.Fatalf("err = %v, want a FeatureGateError", startErr)
	}
	if gate.Feature != "rfc2544" {
		t.Errorf("feature = %q, want %q", gate.Feature, "rfc2544")
	}
}

// Another client already owns the interface. The CLI must say so rather
// than report a nondescript failure or, worse, appear to have started.
func TestStartSurfacesAConflict(t *testing.T) {
	dir, _ := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		writeJSON(t, w, map[string]any{"error": map[string]any{"message": "A test is already running"}})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_, startErr := c.Start(context.Background(), api.TestStartRequest{
		Interface: "eth0",
		Tests:     []api.TestStepRequest{{TestType: "rfc2544_throughput"}},
	})
	if !errors.Is(startErr, daemonclient.ErrRunInProgress) {
		t.Errorf("err = %v, want ErrRunInProgress", startErr)
	}
}

func TestStatusReportsProgress(t *testing.T) {
	dir, _ := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{
			"suiteId": "stem-9f8e-3", "testStatus": "running",
			"stepsTotal": 3, "stepsComplete": 1, "currentStep": 2,
			"phase": "Executing rfc2544_latency", "elapsedSeconds": 12,
		})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	status, err := c.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.SuiteID != "stem-9f8e-3" {
		t.Errorf("suiteId = %q, want %q", status.SuiteID, "stem-9f8e-3")
	}
	if status.TestStatus != "running" {
		t.Errorf("testStatus = %q, want %q", status.TestStatus, "running")
	}
	if status.StepsComplete != 1 || status.StepsTotal != 3 {
		t.Errorf("steps = %d/%d, want 1/3", status.StepsComplete, status.StepsTotal)
	}
	if status.Phase != "Executing rfc2544_latency" {
		t.Errorf("phase = %q", status.Phase)
	}
}

// Terminal states end a watch. Treating "completed" as still-running would
// hang the CLI forever.
func TestStatusTerminalStates(t *testing.T) {
	for state, terminal := range map[string]bool{
		"running":   false,
		"starting":  false,
		"completed": true,
		"error":     true,
		"cancelled": true,
		"stopped":   true,
		"idle":      true,
	} {
		t.Run(state, func(t *testing.T) {
			if got := daemonclient.Terminal(state); got != terminal {
				t.Errorf("Terminal(%q) = %v, want %v", state, got, terminal)
			}
		})
	}
}

// No daemon means no run: the CLI must say so plainly instead of silently
// falling back to its own dataplane (#1166).
func TestOpenWithoutADaemon(t *testing.T) {
	t.Parallel()

	_, err := daemonclient.Open(t.TempDir())
	if !errors.Is(err, daemonconn.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// The daemon serves a self-signed certificate. The client pins the one the
// daemon published and must reject anything else — never skip verification.
func TestClientRejectsAnUntrustedCertificate(t *testing.T) {
	dir, _ := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"status": "started"})
	})

	// Point the descriptor at a certificate the daemon does not serve.
	// A second httptest server would not do: net/http/httptest hands every
	// TLS server the same built-in certificate, so pinning to it would
	// succeed and the test would prove nothing.
	d, err := daemonconn.Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	writePEM(t, d.CAFile, unrelatedCertDER(t))

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_, statusErr := c.Status(context.Background())
	if statusErr == nil {
		t.Fatal("Status against an untrusted certificate succeeded; verification is not being enforced")
	}
	var unknown x509.UnknownAuthorityError
	var hostErr x509.HostnameError
	if !errors.As(statusErr, &unknown) && !errors.As(statusErr, &hostErr) {
		t.Errorf("err = %v, want a certificate verification failure", statusErr)
	}
}

// unrelatedCertDER returns a valid certificate that signs nothing this test
// connects to.
func unrelatedCertDER(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "not the stem daemon"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return der
}

func TestStopRequestsCancellation(t *testing.T) {
	dir, rec := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"status": "stopped"})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if stopErr := c.Stop(context.Background()); stopErr != nil {
		t.Fatalf("Stop: %v", stopErr)
	}

	got := rec.last()
	if got.Method != http.MethodPost || got.URL.Path != "/api/v1/test/stop" {
		t.Errorf("request = %s %s, want POST /api/v1/test/stop", got.Method, got.URL.Path)
	}
}

// writePEM stores a DER certificate in the PEM form the client reads.
func writePEM(t *testing.T, path string, der []byte) {
	t.Helper()
	encoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write certificate: %v", err)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

// Starting the reflector is the same daemon-owned run as any test, so the
// CLI gets a run ID for it and the web UI sees the same one.
func TestStartReflectorReturnsARunID(t *testing.T) {
	dir, rec := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"status": "started", "suiteId": "stem-abc-4"})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	runID, err := c.StartReflector(context.Background(), "eth0", "netally", 3842)
	if err != nil {
		t.Fatalf("StartReflector: %v", err)
	}
	if runID != "stem-abc-4" {
		t.Errorf("run ID = %q, want %q", runID, "stem-abc-4")
	}
	if got := rec.last().URL.Path; got != "/api/v1/test/start" {
		t.Errorf("path = %q, want /api/v1/test/start", got)
	}
}

// The port and profile an operator asked for have to reach the daemon;
// dropping them would silently reflect on the wrong port.
func TestStartReflectorCarriesProfileAndPort(t *testing.T) {
	var body []byte
	dir, _ := newDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		writeJSON(t, w, map[string]any{"status": "started", "suiteId": "stem-abc-5"})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, startErr := c.StartReflector(context.Background(), "eth1", "netally", 3842); startErr != nil {
		t.Fatalf("StartReflector: %v", startErr)
	}

	var sent api.TestStartRequest
	if unmarshalErr := json.Unmarshal(body, &sent); unmarshalErr != nil {
		t.Fatalf("decode request: %v", unmarshalErr)
	}
	if sent.Interface != "eth1" {
		t.Errorf("interface = %q, want eth1", sent.Interface)
	}
	if sent.Profile != "netally" {
		t.Errorf("profile = %q, want netally", sent.Profile)
	}
	if sent.PeerPort != 3842 {
		t.Errorf("peerPort = %d, want 3842", sent.PeerPort)
	}
	if len(sent.Tests) != 1 || sent.Tests[0].TestType != "reflect" {
		t.Errorf("tests = %+v, want a single reflect step", sent.Tests)
	}
}

func TestReflectorStatsReportsCounters(t *testing.T) {
	dir, rec := newDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{
			"running": true, "packetsReceived": 12, "packetsReflected": 11, "bytesReceived": 792,
		})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	stats, err := c.ReflectorStats(context.Background())
	if err != nil {
		t.Fatalf("ReflectorStats: %v", err)
	}
	if !stats.Running || stats.PacketsReceived != 12 || stats.PacketsReflected != 11 {
		t.Errorf("stats = %+v, want a running reflector with 12/11 packets", stats)
	}
	if got := rec.last().URL.Path; got != "/api/v1/reflector/stats" {
		t.Errorf("path = %q, want /api/v1/reflector/stats", got)
	}
}

// --oui and --port are reflector configuration, not run parameters. If the
// client did not send them the flags would parse and then do nothing.
func TestConfigureReflectorSendsTheFilters(t *testing.T) {
	var body []byte
	dir, rec := newDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		writeJSON(t, w, map[string]any{"status": "updated"})
	})

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	want := api.ReflectorConfig{Profile: "netally", OUIFilter: "00:c0:17", PortFilter: 3842}
	if cfgErr := c.ConfigureReflector(context.Background(), want); cfgErr != nil {
		t.Fatalf("ConfigureReflector: %v", cfgErr)
	}

	if got := rec.last().URL.Path; got != "/api/v1/reflector/config" {
		t.Errorf("path = %q, want /api/v1/reflector/config", got)
	}
	var sent api.ReflectorConfig
	if unmarshalErr := json.Unmarshal(body, &sent); unmarshalErr != nil {
		t.Fatalf("decode request: %v", unmarshalErr)
	}
	if sent.Profile != want.Profile || sent.OUIFilter != want.OUIFilter || sent.PortFilter != want.PortFilter {
		t.Errorf("sent = %+v, want %+v", sent, want)
	}
}

// The daemon rotates its token every 12 hours and mints a fresh one on
// restart. A client holding the old one must pick up the published
// replacement rather than making the operator re-run a command when a
// valid descriptor is sitting on disk (#1179).
func TestClientReloadsTheDescriptorAfterA401(t *testing.T) {
	var (
		mu       sync.Mutex
		accepted = "first-token"
	)
	rec := &recorder{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		if r.URL.Path == "/api/v1/auth/csrf-token" {
			writeJSON(t, w, map[string]string{"token": "csrf-value"})
			return
		}
		mu.Lock()
		want := accepted
		mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer "+want {
			w.WriteHeader(http.StatusUnauthorized)
			writeJSON(t, w, map[string]any{"error": map[string]any{"message": "expired"}})
			return
		}
		writeJSON(t, w, map[string]any{"suiteId": "stem-abc-1", "testStatus": "running"})
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	caFile := filepath.Join(dir, "server.crt")
	writePEM(t, caFile, srv.Certificate().Raw)
	publish := func(token string) {
		t.Helper()
		if err := daemonconn.Publish(dir, daemonconn.Descriptor{
			URL: srv.URL, Token: token, CAFile: caFile,
		}); err != nil {
			t.Fatalf("Publish: %v", err)
		}
	}
	publish("first-token")

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	// The daemon rotates: the old token stops working and a new descriptor
	// is published, exactly as the refresher does.
	mu.Lock()
	accepted = "rotated-token"
	mu.Unlock()
	publish("rotated-token")

	status, err := c.Status(context.Background())
	if err != nil {
		t.Fatalf("Status after rotation: %v", err)
	}
	if status.SuiteID != "stem-abc-1" {
		t.Errorf("suiteId = %q, want stem-abc-1", status.SuiteID)
	}
	if got := rec.last().Header.Get("Authorization"); got != "Bearer rotated-token" {
		t.Errorf("Authorization = %q, want the rotated token", got)
	}
}

// A credential that is genuinely revoked must still surface as a refusal.
// When the descriptor has not moved there is nothing to retry with, so the
// client does not spend a second round-trip presenting the same token.
func TestClientSurfacesA401TheReloadCannotFix(t *testing.T) {
	rec := &recorder{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		if r.URL.Path == "/api/v1/auth/csrf-token" {
			writeJSON(t, w, map[string]string{"token": "csrf-value"})
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(t, w, map[string]any{"error": map[string]any{"message": "revoked"}})
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	caFile := filepath.Join(dir, "server.crt")
	writePEM(t, caFile, srv.Certificate().Raw)
	if err := daemonconn.Publish(dir, daemonconn.Descriptor{
		URL: srv.URL, Token: "stale-token", CAFile: caFile,
	}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if _, statusErr := c.Status(context.Background()); statusErr == nil {
		t.Fatal("Status = nil error against a daemon that refuses the credential")
	}

	if got := countPath(rec, "/api/v1/stats"); got != 1 {
		t.Errorf("stats was requested %d times, want 1 — the descriptor had not moved", got)
	}
}

// A rotated descriptor that the daemon still refuses must surface the
// refusal after exactly one retry, never loop.
func TestClientRetriesOnceThenSurfacesTheRefusal(t *testing.T) {
	rec := &recorder{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		if r.URL.Path == "/api/v1/auth/csrf-token" {
			writeJSON(t, w, map[string]string{"token": "csrf-value"})
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(t, w, map[string]any{"error": map[string]any{"message": "revoked"}})
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	caFile := filepath.Join(dir, "server.crt")
	writePEM(t, caFile, srv.Certificate().Raw)
	if err := daemonconn.Publish(dir, daemonconn.Descriptor{
		URL: srv.URL, Token: "first-token", CAFile: caFile,
	}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	// The descriptor moves, so a retry is warranted — but the daemon still
	// refuses, and that answer has to reach the caller.
	if rotateErr := daemonconn.Publish(dir, daemonconn.Descriptor{
		URL: srv.URL, Token: "rotated-token", CAFile: caFile,
	}); rotateErr != nil {
		t.Fatalf("Publish: %v", rotateErr)
	}

	if _, statusErr := c.Status(context.Background()); statusErr == nil {
		t.Fatal("Status = nil error against a daemon that refuses every credential")
	}
	if got := countPath(rec, "/api/v1/stats"); got != 2 {
		t.Errorf("stats was requested %d times, want 2 (the original and one retry)", got)
	}
}

// A client given an explicit URL and token has no descriptor to reload, so
// its 401 must surface immediately rather than reading some unrelated
// daemon's file off this host.
func TestRemoteClientDoesNotReload(t *testing.T) {
	rec := &recorder{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		if r.URL.Path == "/api/v1/auth/csrf-token" {
			writeJSON(t, w, map[string]string{"token": "csrf-value"})
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(t, w, map[string]any{"error": map[string]any{"message": "no"}})
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	caFile := filepath.Join(dir, "server.crt")
	writePEM(t, caFile, srv.Certificate().Raw)

	c, err := daemonclient.OpenRemote(srv.URL, "explicit-token", caFile)
	if err != nil {
		t.Fatalf("OpenRemote: %v", err)
	}
	if _, statusErr := c.Status(context.Background()); statusErr == nil {
		t.Fatal("Status = nil error against a daemon that refuses the credential")
	}

	if got := countPath(rec, "/api/v1/stats"); got != 1 {
		t.Errorf("stats was requested %d times, want 1 (no descriptor to reload)", got)
	}
}

// countPath reports how many times the client asked for one path.
func countPath(rec *recorder, path string) int {
	count := 0
	for _, req := range rec.all() {
		if req.URL.Path == path {
			count++
		}
	}
	return count
}

// The daemon keys CSRF by the bearer it was issued for, so a token rotation
// invalidates the CSRF token with it. A mutating request after rotation has
// to fetch a fresh one or the retry trades a 401 for a 403.
func TestClientRefreshesCSRFAfterRotation(t *testing.T) {
	var (
		mu       sync.Mutex
		accepted = "first-token"
		csrfFor  = map[string]string{"first-token": "csrf-first", "rotated-token": "csrf-rotated"}
	)
	rec := &recorder{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r)
		bearer := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		if r.URL.Path == "/api/v1/auth/csrf-token" {
			writeJSON(t, w, map[string]string{"token": csrfFor[bearer]})
			return
		}

		mu.Lock()
		want := accepted
		mu.Unlock()
		if bearer != want {
			w.WriteHeader(http.StatusUnauthorized)
			writeJSON(t, w, map[string]any{"error": map[string]any{"message": "expired"}})
			return
		}
		// The CSRF token must be the one issued for THIS bearer.
		if r.Header.Get("X-Csrf-Token") != csrfFor[bearer] {
			w.WriteHeader(http.StatusForbidden)
			writeJSON(t, w, map[string]any{"error": map[string]any{"message": "stale CSRF token"}})
			return
		}
		writeJSON(t, w, map[string]any{"status": "stopped"})
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	caFile := filepath.Join(dir, "server.crt")
	writePEM(t, caFile, srv.Certificate().Raw)
	publish := func(token string) {
		t.Helper()
		if err := daemonconn.Publish(dir, daemonconn.Descriptor{
			URL: srv.URL, Token: token, CAFile: caFile,
		}); err != nil {
			t.Fatalf("Publish: %v", err)
		}
	}
	publish("first-token")

	c, err := daemonclient.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	// Prime the CSRF token against the first bearer.
	if stopErr := c.Stop(context.Background()); stopErr != nil {
		t.Fatalf("Stop before rotation: %v", stopErr)
	}

	mu.Lock()
	accepted = "rotated-token"
	mu.Unlock()
	publish("rotated-token")

	if stopErr := c.Stop(context.Background()); stopErr != nil {
		t.Errorf("Stop after rotation: %v", stopErr)
	}
}
