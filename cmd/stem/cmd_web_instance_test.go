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
	second := exec.CommandContext(ctx, os.Args[0])
	second.Env = webDaemonEnv(dataDir, "8545")
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

func webDaemonEnv(dataDir, port string) []string {
	return append(os.Environ(),
		reexecEnv+"="+port,
		"STEM_DATA_DIR="+dataDir,
		"STEM_AUTH_USERNAME=instanceuser",
		"STEM_AUTH_PASSWORD=instancepass123",
	)
}

func startWebDaemon(t *testing.T, dataDir, port string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0])
	cmd.Env = webDaemonEnv(dataDir, port)
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
