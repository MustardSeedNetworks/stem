//go:build linux

package netif

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadSysfsFileSuccess tests the successful path of readSysfsFile on Linux.
// This test only runs on Linux where /sys/class/net exists.
func TestReadSysfsFileSuccess(t *testing.T) {
	// Find an interface that exists on the system.
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		t.Skipf("Cannot read /sys/class/net: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ifaceName := entry.Name()

		// Try to read the address file (should exist for all interfaces).
		addressPath := filepath.Join("/sys/class/net", ifaceName, "address")
		if _, err := os.Stat(addressPath); os.IsNotExist(err) {
			continue
		}

		// Test readSysfsFile with this interface.
		data, err := readSysfsFile(ifaceName, "address")
		if err != nil {
			t.Logf("readSysfsFile(%s, address) failed: %v", ifaceName, err)
			continue
		}

		if len(data) == 0 {
			t.Errorf("readSysfsFile(%s, address) returned empty data", ifaceName)
		}
		return
	}

	t.Skip("No suitable interface found for testing")
}

// TestGetSpeedSuccess tests the successful path of getSpeed on Linux.
func TestGetSpeedSuccess(t *testing.T) {
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		t.Skipf("Cannot read /sys/class/net: %v", err)
	}

	for _, entry := range entries {
		ifaceName := entry.Name()
		if ifaceName == "lo" {
			continue
		}

		// Check if speed file exists.
		speedPath := filepath.Join("/sys/class/net", ifaceName, "speed")
		if _, err := os.Stat(speedPath); os.IsNotExist(err) {
			continue
		}

		speed := getSpeed(ifaceName)
		t.Logf("getSpeed(%s) = %d", ifaceName, speed)
		// Speed could be 0, -1, or positive.
		return
	}

	t.Skip("No interface with speed file found")
}

// TestGetDuplexSuccess tests the successful path of getDuplex on Linux.
func TestGetDuplexSuccess(t *testing.T) {
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		t.Skipf("Cannot read /sys/class/net: %v", err)
	}

	for _, entry := range entries {
		ifaceName := entry.Name()
		if ifaceName == "lo" {
			continue
		}

		// Check if duplex file exists.
		duplexPath := filepath.Join("/sys/class/net", ifaceName, "duplex")
		if _, err := os.Stat(duplexPath); os.IsNotExist(err) {
			continue
		}

		duplex := getDuplex(ifaceName)
		t.Logf("getDuplex(%s) = %s", ifaceName, duplex)
		return
	}

	t.Skip("No interface with duplex file found")
}

// TestGetDriverSuccess tests the successful path of getDriver on Linux.
func TestGetDriverSuccess(t *testing.T) {
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		t.Skipf("Cannot read /sys/class/net: %v", err)
	}

	for _, entry := range entries {
		ifaceName := entry.Name()
		if ifaceName == "lo" {
			continue
		}

		// Check if driver symlink exists.
		driverPath := filepath.Join("/sys/class/net", ifaceName, "device", "driver")
		if _, err := os.Lstat(driverPath); os.IsNotExist(err) {
			continue
		}

		driver := getDriver(ifaceName)
		t.Logf("getDriver(%s) = %s", ifaceName, driver)
		if driver == valueUnknown {
			t.Errorf("getDriver(%s) returned unknown for existing driver", ifaceName)
		}
		return
	}

	t.Skip("No interface with driver symlink found")
}

// TestIsPhysicalSuccess tests the successful path of isPhysical on Linux.
func TestIsPhysicalSuccess(t *testing.T) {
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		t.Skipf("Cannot read /sys/class/net: %v", err)
	}

	for _, entry := range entries {
		ifaceName := entry.Name()
		if ifaceName == "lo" {
			continue
		}

		// Check if device directory exists.
		devicePath := filepath.Join("/sys/class/net", ifaceName, "device")
		if _, err := os.Stat(devicePath); os.IsNotExist(err) {
			continue
		}

		result := isPhysical(ifaceName)
		if !result {
			t.Errorf("isPhysical(%s) = false, but device directory exists", ifaceName)
		}
		return
	}

	t.Skip("No physical interface found")
}
