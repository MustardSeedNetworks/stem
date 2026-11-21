// SPDX-License-Identifier: BUSL-1.1

package daemonclient_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
)

const (
	readyBody    = `{"status":"ready","checks":{"auth":{"status":"ok"}}}`
	notReadyBody = `{"status":"not_ready","checks":{"auth":{"status":"error","error":"auth manager not initialized"}}}`
	buildBody    = `{"version":"v1.2.3","commit":"abc1234","buildTime":"2026-11-01T00:00:00Z",` +
		`"uiBuildHash":"deadbeef","releaseTrain":"2026.11","tlsFingerprint":"AA:BB"}`
)

// probeServer serves the two unauthenticated endpoints with the given
// statuses and bodies, and returns a target that trusts its certificate.
func probeServer(
	t *testing.T,
	versionStatus int,
	versionBody string,
	readyStatus int,
	readyBody string,
) daemonclient.Target {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/__version", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("probe sent a credential to /__version")
		}
		w.WriteHeader(versionStatus)
		_, _ = w.Write([]byte(versionBody))
	})
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("probe sent a credential to /health/ready")
		}
		w.WriteHeader(readyStatus)
		_, _ = w.Write([]byte(readyBody))
	})
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)

	caFile := filepath.Join(t.TempDir(), "server.crt")
	writePEM(t, caFile, srv.Certificate().Raw)
	return daemonclient.Target{URL: srv.URL, CAFile: caFile}
}

func TestProbeReadsBothEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		readyStatus int
		readyBody   string
		wantHealthy bool
		wantStatus  string
		wantCheck   string
	}{
		{"ready", http.StatusOK, readyBody, true, "ready", ""},
		{
			"not ready keeps the failing check",
			http.StatusServiceUnavailable,
			notReadyBody,
			false,
			"not_ready",
			"auth manager not initialized",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := probeServer(t, http.StatusOK, buildBody, tc.readyStatus, tc.readyBody)
			health, err := daemonclient.Probe(context.Background(), target)
			if err != nil {
				t.Fatalf("Probe: %v", err)
			}
			if health.Healthy() != tc.wantHealthy {
				t.Errorf("Healthy() = %v, want %v", health.Healthy(), tc.wantHealthy)
			}
			if health.Ready.Status != tc.wantStatus {
				t.Errorf("ready status = %q, want %q", health.Ready.Status, tc.wantStatus)
			}
			if got := health.Ready.Checks["auth"].Error; got != tc.wantCheck {
				t.Errorf("auth check error = %q, want %q", got, tc.wantCheck)
			}
			if health.Build["version"] != "v1.2.3" || health.Build["tlsFingerprint"] != "AA:BB" {
				t.Errorf("Build = %v", health.Build)
			}
		})
	}
}

func TestProbeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		versionStatus int
		versionBody   string
		readyStatus   int
		readyBody     string
	}{
		{"version 500 is not unreachable", http.StatusInternalServerError, "boom", http.StatusOK, readyBody},
		{"version not JSON", http.StatusOK, "<html>", http.StatusOK, readyBody},
		{"ready 404", http.StatusOK, buildBody, http.StatusNotFound, ""},
		{"ready not JSON", http.StatusOK, buildBody, http.StatusOK, "<html>"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := probeServer(t, tc.versionStatus, tc.versionBody, tc.readyStatus, tc.readyBody)
			_, err := daemonclient.Probe(context.Background(), target)
			if err == nil {
				t.Fatal("Probe succeeded, want an error")
			}
			if errors.Is(err, daemonclient.ErrUnreachable) {
				t.Errorf("a daemon that answered was reported unreachable: %v", err)
			}
		})
	}
}

// A listener that has gone away is the one case ErrUnreachable names.
func TestProbeReportsAClosedPortAsUnreachable(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	url := "https://" + listener.Addr().String()
	if closeErr := listener.Close(); closeErr != nil {
		t.Fatalf("close: %v", closeErr)
	}

	_, err = daemonclient.Probe(context.Background(), daemonclient.Target{URL: url})
	if !errors.Is(err, daemonclient.ErrUnreachable) {
		t.Fatalf("Probe = %v, want ErrUnreachable", err)
	}
}

// A daemon whose certificate the caller does not trust IS running: calling it
// unreachable would send an operator to restart a healthy service.
func TestProbeDoesNotMistakeAnUntrustedCertificateForADownDaemon(t *testing.T) {
	t.Parallel()

	target := probeServer(t, http.StatusOK, buildBody, http.StatusOK, readyBody)
	target.CAFile = ""

	_, err := daemonclient.Probe(context.Background(), target)
	if err == nil {
		t.Fatal("Probe trusted a self-signed certificate it was not given")
	}
	if errors.Is(err, daemonclient.ErrUnreachable) {
		t.Fatalf("an untrusted certificate was reported as unreachable: %v", err)
	}
}

func TestProbeRejectsAnUnreadableCertificate(t *testing.T) {
	t.Parallel()

	_, err := daemonclient.Probe(context.Background(), daemonclient.Target{
		URL:    "https://127.0.0.1:1",
		CAFile: filepath.Join(t.TempDir(), "missing.crt"),
	})
	if err == nil || errors.Is(err, daemonclient.ErrUnreachable) {
		t.Fatalf("Probe = %v, want a certificate error", err)
	}
}
