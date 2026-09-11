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
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
