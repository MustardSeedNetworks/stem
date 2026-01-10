// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

//go:build darwin

package license

import (
	"testing"
)

// TestGetDarwinCPUSerialReturnsNonEmpty tests that getDarwinCPUSerial returns a value.
func TestGetDarwinCPUSerialReturnsNonEmpty(t *testing.T) {
	serial := getDarwinCPUSerial()

	// Should not be empty.
	if serial == "" {
		t.Error("getDarwinCPUSerial should not return empty string")
	}
}

// TestGetDarwinDiskSerialReturnsNonEmpty tests that getDarwinDiskSerial returns a value.
func TestGetDarwinDiskSerialReturnsNonEmpty(t *testing.T) {
	serial := getDarwinDiskSerial()

	// Should not be empty.
	if serial == "" {
		t.Error("getDarwinDiskSerial should not return empty string")
	}
}

// TestGetCPUSerialOnDarwin tests getCPUSerial returns Darwin-specific result.
func TestGetCPUSerialOnDarwin(t *testing.T) {
	serial := getCPUSerial()

	// Should not be empty.
	if serial == "" {
		t.Error("getCPUSerial should not return empty on Darwin")
	}

	// Should not be the Linux default.
	if serial == defaultLinuxCPU {
		t.Error("getCPUSerial should not return Linux default on Darwin")
	}
}

// TestGetDiskSerialOnDarwin tests getDiskSerial returns Darwin-specific result.
func TestGetDiskSerialOnDarwin(t *testing.T) {
	serial := getDiskSerial()

	// Should not be empty.
	if serial == "" {
		t.Error("getDiskSerial should not return empty on Darwin")
	}

	// Should not be the Linux default.
	if serial == defaultLinuxDisk {
		t.Error("getDiskSerial should not return Linux default on Darwin")
	}
}
