// SPDX-License-Identifier: BUSL-1.1

package daemonconn_test

import (
	"errors"
	"path/filepath"
	"slices"
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

// An override names where the daemon is, so it is the only place searched:
// falling through to the packaged directory would reach a different daemon.
func TestSearchDirsUsesOnlyTheOverride(t *testing.T) {
	t.Setenv("STEM_DATA_DIR", "/tmp/explicit-dir")
	if dirs := daemonconn.SearchDirs(); !slices.Equal(dirs, []string{"/tmp/explicit-dir"}) {
		t.Fatalf("SearchDirs() = %v, want only the override", dirs)
	}
}

// Without an override the packaged location comes first, then the working
// directory the daemon falls back to.
func TestSearchDirsWithoutAnOverride(t *testing.T) {
	t.Setenv("STEM_DATA_DIR", "")
	dirs := daemonconn.SearchDirs()
	if len(dirs) != 2 || !filepath.IsAbs(dirs[0]) || dirs[1] != "." {
		t.Fatalf("SearchDirs() = %v, want [<packaged dir> .]", dirs)
	}
}

// The service's data directory is the override when set, else the packaged
// location, never the working directory: `stem update` copies and restores it.
func TestServiceDataDir(t *testing.T) {
	t.Setenv("STEM_DATA_DIR", "/tmp/explicit-dir")
	if got := daemonconn.ServiceDataDir(); got != "/tmp/explicit-dir" {
		t.Fatalf("ServiceDataDir() with the override = %q", got)
	}
	t.Setenv("STEM_DATA_DIR", "")
	if got := daemonconn.ServiceDataDir(); got == "." || !filepath.IsAbs(got) {
		t.Fatalf("ServiceDataDir() = %q; want the packaged absolute path", got)
	}
}
