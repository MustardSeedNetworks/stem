// SPDX-License-Identifier: BUSL-1.1

//go:build !windows

package main

import (
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

func fileOwner(t *testing.T, path string) int {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("no owner information for %s", path)
	}
	return int(stat.Uid)
}

// With no daemon, `stem license --trial` writes the file itself, and it must
// end up owned like the data directory it lands in.
func TestLicenseOfflineWriteMatchesTheDataDirectoryOwner(t *testing.T) {
	licenseHome(t) // STEM_DATA_DIR is a temp directory owned by this user
	dataDir := os.Getenv("STEM_DATA_DIR")

	out := captureStdout(t, func() {
		if err := licenseCmd([]string{"--trial"}); err != nil {
			t.Errorf("licenseCmd --trial: %v", err)
		}
	})
	if !strings.Contains(out, "Success") {
		t.Fatalf("the trial did not start:\n%s", out)
	}
	if got, want := fileOwner(t, license.DefaultLicensePath()), fileOwner(t, dataDir); got != want {
		t.Errorf("license owner = %d, want the data directory's %d", got, want)
	}
}

func TestPrepareLicenseWriteForTheOwnerIsANoOp(t *testing.T) {
	dir := t.TempDir()
	if err := prepareLicenseWrite(dir+"/.license", dir); err != nil {
		t.Fatalf("prepareLicenseWrite as the directory's owner = %v", err)
	}
}

func TestPrepareLicenseWriteWithNoPackagedDirectoryIsANoOp(t *testing.T) {
	if err := prepareLicenseWrite("/nonexistent/.license", "/nonexistent"); err != nil {
		t.Fatalf("prepareLicenseWrite with no data directory = %v", err)
	}
}

// A link planted at the license path before an activation must be refused, and
// what it points at must be untouched.
func TestLicenseWriteRefusesASymlinkAtTheLicensePath(t *testing.T) {
	licenseHome(t)
	dataDir := os.Getenv("STEM_DATA_DIR")
	target := t.TempDir() + "/victim"
	if err := os.WriteFile(target, []byte("untouched"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, license.DefaultLicensePath()); err != nil {
		t.Fatal(err)
	}
	before := fileOwner(t, target)

	var err error
	captureStdout(t, func() { err = licenseCmd([]string{"--trial"}) })
	if err == nil {
		t.Fatal("licenseCmd --trial wrote through a symbolic link")
	}
	if got, readErr := os.ReadFile(target); readErr != nil || string(got) != "untouched" {
		t.Errorf("the link target = %q, %v, want it untouched", got, readErr)
	}
	if got := fileOwner(t, target); got != before {
		t.Errorf("the link target's owner changed from %d to %d", before, got)
	}
	if _, statErr := os.Lstat(dataDir + "/.license"); statErr != nil {
		t.Errorf("the planted link was removed: %v", statErr)
	}
}
