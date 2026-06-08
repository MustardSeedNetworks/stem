// SPDX-License-Identifier: BUSL-1.1

//go:build linux

package license_test

import (
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

func TestGenerateFingerprintLinux(t *testing.T) {
	fp, err := license.GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint() error: %v", err)
	}

	if fp.CPUSerial == "" {
		t.Error("Expected Linux CPU serial to be populated")
	}
	if fp.DiskSerial == "" {
		t.Error("Expected Linux disk serial to be populated")
	}
	if fp.Platform != "linux" {
		t.Errorf("Expected platform 'linux', got %q", fp.Platform)
	}
	if strings.Contains(strings.ToLower(fp.CPUSerial), "error") {
		t.Errorf("CPU serial should not contain error text: %s", fp.CPUSerial)
	}
	if strings.Contains(strings.ToLower(fp.DiskSerial), "error") {
		t.Errorf("Disk serial should not contain error text: %s", fp.DiskSerial)
	}
}
