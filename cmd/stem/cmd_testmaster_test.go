// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// licenseHome points the license manager at a temp HOME so a case decides what
// is on disk, and returns the config directory it will read.
func licenseHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "stem")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return dir
}

// TestCheckTestLicenseRefusesAMalformedFile is the CLI half of #1068: a file
// the manager cannot parse used to be indistinguishable from a fresh install,
// so `stem test` started a Professional trial and overwrote the operator's
// license. It must refuse instead.
func TestCheckTestLicenseRefusesAMalformedFile(t *testing.T) {
	dir := licenseHome(t)
	if err := os.WriteFile(filepath.Join(dir, ".license"), []byte("not a licence"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if checkTestLicense() {
		t.Error("checkTestLicense() = true on a malformed license file")
	}
	if _, err := os.ReadFile(filepath.Join(dir, ".license")); err != nil {
		t.Errorf("license file no longer readable after the check: %v", err)
	}
}

// TestCheckTestLicenseStartsTheTrialOnAFreshInstall is the control: refusing
// the damaged file must not have made every unlicensed host unable to test.
func TestCheckTestLicenseStartsTheTrialOnAFreshInstall(t *testing.T) {
	dir := licenseHome(t)

	if !checkTestLicense() {
		t.Fatal("checkTestLicense() = false on a fresh install; the trial should start")
	}
	if _, err := os.Stat(filepath.Join(dir, ".license")); err != nil {
		t.Errorf("no state written after starting the trial: %v", err)
	}
}
