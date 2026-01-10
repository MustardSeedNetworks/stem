// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

//go:build linux

package license

import (
	"strings"
	"testing"
)

// TestGetLinuxCPUSerialReturnsNonEmpty tests that getLinuxCPUSerial returns a value.
func TestGetLinuxCPUSerialReturnsNonEmpty(t *testing.T) {
	serial := getLinuxCPUSerial()

	// Should not be empty.
	if serial == "" {
		t.Error("getLinuxCPUSerial should not return empty string")
	}

	// Should be either the default or an actual serial.
	if serial != defaultLinuxCPU {
		// If it's not default, it should be a non-empty string.
		if len(serial) == 0 {
			t.Error("Non-default Linux CPU serial should have content")
		}
	}
}

// TestGetLinuxDiskSerialReturnsNonEmpty tests that getLinuxDiskSerial returns a value.
func TestGetLinuxDiskSerialReturnsNonEmpty(t *testing.T) {
	serial := getLinuxDiskSerial()

	// Should not be empty.
	if serial == "" {
		t.Error("getLinuxDiskSerial should not return empty string")
	}

	// Should be either the default or an actual serial.
	if serial != defaultLinuxDisk {
		// If it's not default, it should be a non-empty string.
		if len(serial) == 0 {
			t.Error("Non-default Linux disk serial should have content")
		}
	}
}

// TestGetCPUSerialOnLinux tests getCPUSerial returns Linux-specific result.
func TestGetCPUSerialOnLinux(t *testing.T) {
	serial := getCPUSerial()

	// Should not be empty.
	if serial == "" {
		t.Error("getCPUSerial should not return empty on Linux")
	}

	// Should not be the Darwin default.
	if serial == defaultDarwinCPU {
		t.Error("getCPUSerial should not return Darwin default on Linux")
	}
}

// TestGetDiskSerialOnLinux tests getDiskSerial returns Linux-specific result.
func TestGetDiskSerialOnLinux(t *testing.T) {
	serial := getDiskSerial()

	// Should not be empty.
	if serial == "" {
		t.Error("getDiskSerial should not return empty on Linux")
	}

	// Should not be the Darwin default.
	if serial == defaultDarwinDisk {
		t.Error("getDiskSerial should not return Darwin default on Linux")
	}
}

// TestLinuxCPUSerialFormats tests various paths in getLinuxCPUSerial.
func TestLinuxCPUSerialFormats(t *testing.T) {
	// This test exercises the function and verifies it handles various system configurations.
	serial := getLinuxCPUSerial()

	// The result should be a string that doesn't contain common error indicators.
	if strings.Contains(serial, "error") || strings.Contains(serial, "Error") {
		t.Errorf("CPU serial should not contain error text: %s", serial)
	}
}

// TestLinuxDiskSerialFormats tests various paths in getLinuxDiskSerial.
func TestLinuxDiskSerialFormats(t *testing.T) {
	// This test exercises the function and verifies it handles various system configurations.
	serial := getLinuxDiskSerial()

	// The result should be a string that doesn't contain common error indicators.
	if strings.Contains(serial, "error") || strings.Contains(serial, "Error") {
		t.Errorf("Disk serial should not contain error text: %s", serial)
	}
}
