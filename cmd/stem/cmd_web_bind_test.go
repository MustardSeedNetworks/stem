// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

// stem#1437: `stem web --host 127.0.0.1` keeps the daemon off the network. The
// flag used to be printed and then ignored, so the listener took every
// address. The daemon must refuse a connection on any other address of the
// host, and the descriptor it publishes for the CLI must name the address it
// bound, since localhost is no longer guaranteed to reach it.
func TestWebHostBindsOnlyThatAddress(t *testing.T) {
	other := nonLoopbackIPv4(t)
	dataDir := t.TempDir()
	startWebDaemon(t, dataDir, "--host", "127.0.0.1", "--port", "8744")
	port := strconv.Itoa(waitForLockRecord(t, dataDir).Port)

	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(t.Context(), "tcp", net.JoinHostPort("127.0.0.1", port))
	if err != nil {
		t.Fatalf("the bound address refused a connection: %v", err)
	}
	_ = conn.Close()

	if conn, err = dialer.DialContext(t.Context(), "tcp", net.JoinHostPort(other, port)); err == nil {
		_ = conn.Close()
		t.Errorf("`--host 127.0.0.1` accepted a connection on %s:%s", other, port)
	}

	descriptor := waitForDescriptor(t, dataDir)
	if want := "https://127.0.0.1:" + port; descriptor.URL != want {
		t.Fatalf("descriptor URL = %q, want %q", descriptor.URL, want)
	}
	caPEM, err := os.ReadFile(descriptor.CAFile)
	if err != nil {
		t.Fatalf("read the published certificate: %v", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		t.Fatal("the published certificate holds no PEM certificate")
	}
	client := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
	}}
	resp := get(t, client, descriptor.URL+"/__version")
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET %s/__version: status %d", descriptor.URL, resp.StatusCode)
	}
}

// An operator who names a port is asking for that port. When it is taken the
// daemon has to say so and exit, not walk to a neighbour the operator never
// asked for and may have firewalled.
func TestExplicitPortInUseIsRefused(t *testing.T) {
	var listenConfig net.ListenConfig
	busy, err := listenConfig.Listen(t.Context(), "tcp", ":0")
	if err != nil {
		t.Fatalf("occupy a port: %v", err)
	}
	t.Cleanup(func() { _ = busy.Close() })
	tcpAddr, isTCP := busy.Addr().(*net.TCPAddr)
	if !isTCP {
		t.Fatalf("listener address %s is not TCP", busy.Addr())
	}
	port := strconv.Itoa(tcpAddr.Port)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	daemon := exec.CommandContext(ctx, testBinary(t))
	daemon.Env = webDaemonEnv(t, t.TempDir(), "--port", port)
	daemon.Dir = t.TempDir()
	output, runErr := daemon.CombinedOutput()

	if ctx.Err() != nil {
		t.Fatalf("`stem web --port %s` was still running at the deadline with the port taken; output:\n%s",
			port, output)
	}
	var exitErr *exec.ExitError
	if !errors.As(runErr, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("exit = %v, want exit status 1; output:\n%s", runErr, output)
	}
	if !strings.Contains(string(output), port) {
		t.Errorf("the refusal does not name port %s:\n%s", port, output)
	}
}

// nonLoopbackIPv4 is one of this host's own addresses that is not loopback,
// the address a daemon bound to 127.0.0.1 must refuse.
func nonLoopbackIPv4(t *testing.T) string {
	t.Helper()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatalf("InterfaceAddrs: %v", err)
	}
	for _, addr := range addrs {
		if ipNet, isIPNet := addr.(*net.IPNet); isIPNet && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	t.Skip("no non-loopback IPv4 address on this host")
	return ""
}

// waitForDescriptor waits for the descriptor the daemon publishes after it
// binds, which comes after the lock records the port.
func waitForDescriptor(t *testing.T, dataDir string) daemonconn.Descriptor {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		descriptor, err := daemonconn.Read(dataDir)
		if err == nil {
			return descriptor
		}
		if !daemonconn.IsNotFound(err) {
			t.Fatalf("read the descriptor: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("the daemon never published its descriptor")
	return daemonconn.Descriptor{}
}
