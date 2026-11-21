// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

// A real daemon, not a stub: its own descriptor, certificate and readiness
// body. status finds it, backup and restore are refused while it runs, and
// once it is gone status says so.
func TestOperationsAgainstARealDaemon(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dataDir)

	daemon := startWebDaemon(t, dataDir, "--port", "8546")
	waitForLockRecord(t, dataDir)
	eventually(t, "the daemon to publish its descriptor and answer ready", func() bool {
		_, _, err := runStatus(t)
		return err == nil
	})

	stdout, _, err := runStatus(t, "--json")
	if err != nil || !strings.Contains(stdout, `"state": "running"`) {
		t.Fatalf("status = %v\n%s", err, stdout)
	}

	archive := filepath.Join(t.TempDir(), "stem.tar.gz")
	var stderr bytes.Buffer
	if err = backupCmd([]string{"-o", archive}, io.Discard, &stderr); err == nil ||
		!strings.Contains(stderr.String(), "the daemon is running") {
		t.Fatalf("backup under a running daemon = %v, %q", err, stderr.String())
	}
	if _, statErr := os.Stat(archive); statErr == nil {
		t.Error("a refused backup left an archive")
	}
	stderr.Reset()
	if err = restoreCmd([]string{archive}, io.Discard, &stderr); err == nil {
		t.Fatal("restore under a running daemon succeeded")
	}

	if killErr := daemon.Process.Kill(); killErr != nil {
		t.Fatal(killErr)
	}
	_, _ = daemon.Process.Wait()
	if _, statErr := os.Stat(daemonconn.Path(dataDir)); statErr != nil {
		t.Fatalf("the killed daemon's descriptor should remain, stale: %v", statErr)
	}
	eventually(t, "status to report the stopped daemon", func() bool {
		_, _, runErr := runStatus(t)
		return exitCodeOf(runErr) == statusNotRunning
	})
}
