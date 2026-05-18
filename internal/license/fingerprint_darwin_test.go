// SPDX-License-Identifier: BUSL-1.1

//go:build darwin

package license_test

import (
	"testing"

	"github.com/krisarmstrong/stem/internal/license"
)

func TestGenerateFingerprintDarwin(t *testing.T) {
	fp, err := license.GenerateFingerprint()
	if err != nil {
		t.Fatalf("GenerateFingerprint() error: %v", err)
	}

	if fp.CPUSerial == "" {
		t.Error("Expected Darwin CPU serial to be populated")
	}
	if fp.DiskSerial == "" {
		t.Error("Expected Darwin disk serial to be populated")
	}
	if fp.Platform != "darwin" {
		t.Errorf("Expected platform 'darwin', got %q", fp.Platform)
	}
	if fp.MACAddress == "" {
		t.Error("Expected MAC address to be populated")
	}
}
