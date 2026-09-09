// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

// TestDisplayLicenseStatusOnAFreshInstall: with nothing on disk the operator
// must be told they are unlicensed and given both ways out.
func TestDisplayLicenseStatusOnAFreshInstall(t *testing.T) {
	licenseHome(t)
	mgr, _, err := license.Load()
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
	mgr, _, err := license.Load()
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
