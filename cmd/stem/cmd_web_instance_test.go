// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/instance"
)

// reexecEnv makes the test binary run webCmd instead of the test suite, so
// both daemons below are real processes: an in-process check could not show
// that the refusal names another process's PID.
const reexecEnv = "STEM_TEST_REEXEC_WEB"

func TestMain(m *testing.M) {
	if port := os.Getenv(reexecEnv); port != "" {
		webCmd([]string{"--port", port})
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// D-STEM-19 (stem#1336): with the +1..+9 port fallback a second `stem web` on
// the same data directory does not collide on the port — it binds a
// neighbouring one and then shares the first daemon's state and descriptor.
// The single-instance lock has to refuse it, and name the daemon that already
// owns the directory so the operator can find it.
//
// Both daemons are subprocesses on purpose. A first daemon faked in-process
// would let the test pass with no SetPort call in the production start-up
// path: the port in the message would be one this test wrote itself.
func TestSecondWebInstanceIsRefusedNamingPIDAndPort(t *testing.T) {
	dataDir := t.TempDir()

	first := startWebDaemon(t, dataDir, "8544")
	held := waitForLockRecord(t, dataDir)
	if held.PID != first.Process.Pid {
		t.Fatalf("the lock names pid %d, the daemon is %d", held.PID, first.Process.Pid)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	second := exec.CommandContext(ctx, testBinary(t))
	second.Env = webDaemonEnv(t, dataDir, "8545")
	second.Dir = t.TempDir()
	output, runErr := second.CombinedOutput()

	if ctx.Err() != nil {
		t.Fatalf("the second `stem web` was still running at the deadline; output:\n%s", output)
	}
	if runErr == nil {
		t.Fatalf("the second `stem web` exited 0; output:\n%s", output)
	}
	var exitErr *exec.ExitError
	if !errors.As(runErr, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("exit = %v, want exit status 1; output:\n%s", runErr, output)
	}

	text := string(output)
	if !strings.Contains(text, strconv.Itoa(held.PID)) {
		t.Errorf("the refusal does not name the holder's pid %d:\n%s", held.PID, text)
	}
	if !strings.Contains(text, strconv.Itoa(held.Port)) {
		t.Errorf("the refusal does not name the holder's port %d:\n%s", held.Port, text)
	}
}

// webDaemonEnv gives the child daemon a HOME of its own as well as a data
// directory: the licence manager reads ~/.config/stem, and a test that runs a
// real daemon under the developer's HOME activates and deactivates their
// licence (the defect D-STEM-18 fixed in the licence handler tests).
func webDaemonEnv(t *testing.T, dataDir, port string) []string {
	t.Helper()
	return append(os.Environ(),
		reexecEnv+"="+port,
		"STEM_DATA_DIR="+dataDir,
		"HOME="+t.TempDir(),
		"STEM_AUTH_USERNAME=instanceuser",
		"STEM_AUTH_PASSWORD=instancepass123",
	)
}

// testBinary is this test binary's own absolute path. os.Args[0] can be
// relative, and the children below run from a working directory of their own.
func testBinary(t *testing.T) string {
	t.Helper()
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	return path
}

func startWebDaemon(t *testing.T, dataDir, port string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(testBinary(t))
	cmd.Env = webDaemonEnv(t, dataDir, port)
	// The daemon writes its self-signed certificate relative to its working
	// directory, so it gets one of its own rather than leaving a certs/ tree
	// in the package source.
	cmd.Dir = t.TempDir()
	if err := cmd.Start(); err != nil {
		t.Fatalf("start the first daemon: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})
	return cmd
}

// waitForLockRecord waits until the running daemon has taken the lock and
// recorded the port it bound. Both facts come from the daemon's own start-up,
// which is the point: the refusal below quotes what the daemon published.
func waitForLockRecord(t *testing.T, dataDir string) instance.Info {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		info, held, err := instance.Probe(dataDir)
		if err != nil {
			t.Fatalf("Probe: %v", err)
		}
		if held && info.PID != 0 && info.Port != 0 {
			return info
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("the first daemon never recorded its pid and port in the instance lock")
	return instance.Info{}
}
