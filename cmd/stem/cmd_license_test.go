// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

// TestDisplayLicenseStatusOnAFreshInstall: with nothing on disk the operator
// must be told they are unlicensed and given both ways out.
func TestDisplayLicenseStatusOnAFreshInstall(t *testing.T) {
	licenseHome(t)
	mgr, err := license.Load()
	if err != nil {
		t.Fatalf("license.Load: %v", err)
	}

	out := captureStdout(t, func() { displayLicenseStatus(mgr) })

	if !strings.Contains(out, "Status:    Not Activated") {
		t.Errorf("status output %q does not report an unactivated host", out)
	}
	for _, want := range []string{"stem license --trial", "stem license --activate"} {
		if !strings.Contains(out, want) {
			t.Errorf("status output is missing the %q hint:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "Device ID:") || !strings.Contains(out, "Platform:") {
		t.Errorf("status output omits the device fingerprint an activation needs:\n%s", out)
	}
}

// TestDisplayLicenseStatusInTrialMode reports the remaining days, which is the
// number the operator acts on.
func TestDisplayLicenseStatusInTrialMode(t *testing.T) {
	licenseHome(t)
	mgr, err := license.Load()
	if err != nil {
		t.Fatalf("license.Load: %v", err)
	}
	if result := mgr.StartTrial(); !result.Success {
		t.Fatalf("StartTrial: %s", result.Message)
	}

	out := captureStdout(t, func() { displayLicenseStatus(mgr) })

	if !strings.Contains(out, "Status:    Trial Mode") {
		t.Errorf("status output %q does not report the trial", out)
	}
	remaining := mgr.TrialDaysRemaining()
	if remaining <= 0 {
		t.Fatalf("a trial started just now has %d days remaining", remaining)
	}
	if !strings.Contains(out, "full access during trial") {
		t.Errorf("status output does not say the trial grants full access:\n%s", out)
	}
}

// TestLicenseCmdWarnsOnAnUnusableFile is the CLI half of #1312: when the state
// on disk is one Stem cannot stand behind, the operator asking about licensing
// is told so and told which file to replace. The warning is what makes a
// stripped or damaged licence visible — the entitlements are silently Free
// either way.
func TestLicenseCmdWarnsOnAnUnusableFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads a 0000 file; the case cannot be built")
	}
	licenseHome(t)
	mgr, err := license.Load()
	if err != nil {
		t.Fatalf("license.Load: %v", err)
	}
	if result := mgr.StartTrial(); !result.Success {
		t.Fatalf("StartTrial: %s", result.Message)
	}
	path := license.DefaultLicensePath()
	if chmodErr := os.Chmod(path, 0o000); chmodErr != nil {
		t.Fatalf("chmod: %v", chmodErr)
	}

	out := captureStdout(t, func() { licenseCmd([]string{"--status"}) })

	want := "Warning: license file " + path + " is unreadable"
	if !strings.Contains(out, want) {
		t.Errorf("status output does not warn that the licence file is unusable:\nwant %q in:\n%s", want, out)
	}
}
