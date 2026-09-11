// SPDX-License-Identifier: BUSL-1.1

package daemonconn_test

import (
	"errors"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

// An operator runs `stem test` from their home directory while the daemon
// keeps its state under the packaged data directory. Discovery has to look
// where the daemon actually is, not where the CLI happens to be.
func TestDiscoverFindsADescriptorOutsideTheWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STEM_DATA_DIR", dir)
	want := daemonconn.Descriptor{URL: "https://localhost:8444", Token: "sentinel"}
	if err := daemonconn.Publish(dir, want); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	got, err := daemonconn.Discover()
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got != want {
		t.Errorf("descriptor = %+v, want %+v", got, want)
	}
}

// No daemon anywhere is a clear refusal, never a silent fallback (#1166).
func TestDiscoverWithoutADaemon(t *testing.T) {
	t.Setenv("STEM_DATA_DIR", t.TempDir())
	if _, err := daemonconn.Discover(); !errors.Is(err, daemonconn.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// The search order has to be deterministic and start with the override, so
// an operator can point the CLI at a daemon in a non-standard location.
func TestSearchDirsHonoursTheOverrideFirst(t *testing.T) {
	t.Setenv("STEM_DATA_DIR", "/tmp/explicit-dir")
	dirs := daemonconn.SearchDirs()
	if len(dirs) == 0 || dirs[0] != "/tmp/explicit-dir" {
		t.Fatalf("SearchDirs() = %v, want the override first", dirs)
	}
	if len(dirs) < 2 {
		t.Error("SearchDirs() offers no fallback beyond the override")
	}
}
