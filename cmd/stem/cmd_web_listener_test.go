// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestWebListenerServesTLSAndRedirectsPlaintext drives the listener a real
// `stem web` binds (foundation pkg/httpserver, STM-FDN-1): one port that
// speaks TLS 1.3 with the certificate it generated, answers a plaintext
// request with a 308 to the same address over https and nothing else (#1296),
// and reports that certificate's fingerprint on /__version.
func TestWebListenerServesTLSAndRedirectsPlaintext(t *testing.T) {
	dataDir := t.TempDir()
	daemon := startWebDaemon(t, dataDir, "8644")
	port := strconv.Itoa(waitForLockRecord(t, dataDir).Port)
	origin := "127.0.0.1:" + port

	certPEM, err := os.ReadFile(filepath.Join(daemon.Dir, "certs", "server.crt"))
	if err != nil {
		t.Fatalf("the daemon generated no certificate: %v", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certPEM) {
		t.Fatal("certs/server.crt holds no PEM certificate")
	}

	plain := &http.Client{
		Timeout:       5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp := get(t, plain, "http://"+origin+"/__version?probe=1")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusPermanentRedirect {
		t.Fatalf("plaintext GET: status %d, want 308", resp.StatusCode)
	}
	if got, want := resp.Header.Get("Location"), "https://"+origin+"/__version?probe=1"; got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}
	if len(body) != 0 {
		t.Errorf("plaintext GET was served content: %q", body)
	}

	// Verified, not InsecureSkipVerify: the served certificate is the one on
	// disk and it covers the loopback address.
	secure := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
	}}
	resp = get(t, secure, "https://"+origin+"/__version")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("https /__version: status %d", resp.StatusCode)
	}
	if resp.TLS.Version != tls.VersionTLS13 {
		t.Errorf("negotiated TLS %#04x, want TLS 1.3", resp.TLS.Version)
	}
	var version struct {
		TLSFingerprint string `json:"tlsFingerprint"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&version); decodeErr != nil {
		t.Fatalf("decode /__version: %v", decodeErr)
	}
	if want := fingerprint(t, certPEM); version.TLSFingerprint != want {
		t.Errorf("/__version tlsFingerprint = %q, want the served certificate's %q", version.TLSFingerprint, want)
	}

	tls12 := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12},
	}}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+origin+"/__version", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp12, err12 := tls12.Do(req); err12 == nil {
		_ = resp12.Body.Close()
		t.Error("a TLS 1.2-only client completed a handshake; the floor is TLS 1.3")
	}
}

func get(t *testing.T, client *http.Client, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}

// fingerprint is the SHA-256 of the certificate's DER as colon-separated
// uppercase hex, the form /__version reports.
func fingerprint(t *testing.T, certPEM []byte) string {
	t.Helper()
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("no PEM block in the certificate")
	}
	sum := sha256.Sum256(block.Bytes)
	hexSum := strings.ToUpper(hex.EncodeToString(sum[:]))
	pairs := make([]string, 0, len(sum))
	for i := 0; i < len(hexSum); i += 2 {
		pairs = append(pairs, hexSum[i:i+2])
	}
	return strings.Join(pairs, ":")
}
