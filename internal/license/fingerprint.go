// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package license provides licensing functionality for Seed Test Suite.
//
// Implements device fingerprinting for hardware-bound licenses using
// MAC address, CPU serial, and disk serial identifiers.
package license

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// DeviceFingerprint contains hardware identifiers for device binding
type DeviceFingerprint struct {
	MACAddress string `json:"mac"`
	CPUSerial  string `json:"cpu"`
	DiskSerial string `json:"disk"`
	Hostname   string `json:"hostname"`
	Platform   string `json:"platform"`
}

// GenerateFingerprint creates a unique device fingerprint
func GenerateFingerprint() (*DeviceFingerprint, error) {
	fp := &DeviceFingerprint{
		Platform: runtime.GOOS,
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err == nil {
		fp.Hostname = hostname
	}

	// Get primary MAC address
	fp.MACAddress = getPrimaryMAC()

	// Get CPU serial (platform-specific)
	fp.CPUSerial = getCPUSerial()

	// Get disk serial (platform-specific)
	fp.DiskSerial = getDiskSerial()

	return fp, nil
}

// Hash returns a 16-character hash of the fingerprint
func (fp *DeviceFingerprint) Hash() string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s",
		fp.MACAddress,
		fp.CPUSerial,
		fp.DiskSerial,
		fp.Hostname,
		fp.Platform,
	)

	hash := sha256.Sum256([]byte(data))
	hexHash := hex.EncodeToString(hash[:])

	// Return first 16 characters (64 bits of entropy)
	return strings.ToUpper(hexHash[:16])
}

// String returns a human-readable representation
func (fp *DeviceFingerprint) String() string {
	return fmt.Sprintf("MAC=%s CPU=%s DISK=%s HOST=%s",
		maskString(fp.MACAddress, 8),
		maskString(fp.CPUSerial, 4),
		maskString(fp.DiskSerial, 4),
		fp.Hostname,
	)
}

// getPrimaryMAC returns the MAC address of the first non-loopback interface
func getPrimaryMAC() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "00:00:00:00:00:00"
	}

	for _, iface := range interfaces {
		// Skip loopback and interfaces without MAC
		if iface.Flags&net.FlagLoopback != 0 || len(iface.HardwareAddr) == 0 {
			continue
		}
		// Skip virtual interfaces (common patterns)
		name := strings.ToLower(iface.Name)
		if strings.HasPrefix(name, "veth") ||
			strings.HasPrefix(name, "docker") ||
			strings.HasPrefix(name, "br-") ||
			strings.HasPrefix(name, "virbr") {
			continue
		}
		return iface.HardwareAddr.String()
	}

	return "00:00:00:00:00:00"
}

// getCPUSerial returns CPU serial number (platform-specific)
func getCPUSerial() string {
	switch runtime.GOOS {
	case "linux":
		return getLinuxCPUSerial()
	case "darwin":
		return getDarwinCPUSerial()
	default:
		return "UNKNOWN"
	}
}

// getLinuxCPUSerial reads CPU serial from /proc/cpuinfo or dmidecode
func getLinuxCPUSerial() string {
	// Try /proc/cpuinfo first (works on ARM)
	data, err := os.ReadFile("/proc/cpuinfo")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Serial") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Try dmidecode (requires root)
	out, err := exec.Command("dmidecode", "-s", "processor-id").Output()
	if err == nil {
		serial := strings.TrimSpace(string(out))
		if serial != "" {
			return serial
		}
	}

	// Try /sys/class/dmi/id/product_serial
	data, err = os.ReadFile("/sys/class/dmi/id/product_serial")
	if err == nil {
		return strings.TrimSpace(string(data))
	}

	return "LINUX-DEFAULT"
}

// getDarwinCPUSerial reads hardware UUID on macOS
func getDarwinCPUSerial() string {
	out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return "DARWIN-DEFAULT"
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				uuid := strings.TrimSpace(parts[1])
				uuid = strings.Trim(uuid, "\"")
				return uuid
			}
		}
	}

	return "DARWIN-DEFAULT"
}

// getDiskSerial returns root disk serial number
func getDiskSerial() string {
	switch runtime.GOOS {
	case "linux":
		return getLinuxDiskSerial()
	case "darwin":
		return getDarwinDiskSerial()
	default:
		return "UNKNOWN"
	}
}

// getLinuxDiskSerial reads disk serial from /sys or udevadm
func getLinuxDiskSerial() string {
	// Try common disk paths
	diskPaths := []string{
		"/sys/block/sda/device/serial",
		"/sys/block/nvme0n1/device/serial",
		"/sys/block/vda/serial",
	}

	for _, path := range diskPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			serial := strings.TrimSpace(string(data))
			if serial != "" {
				return serial
			}
		}
	}

	// Try udevadm
	out, err := exec.Command("udevadm", "info", "--query=property", "--name=/dev/sda").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ID_SERIAL=") {
				return strings.TrimPrefix(line, "ID_SERIAL=")
			}
		}
	}

	return "LINUX-DISK"
}

// getDarwinDiskSerial reads disk serial on macOS
func getDarwinDiskSerial() string {
	out, err := exec.Command("system_profiler", "SPSerialATADataType").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Serial Number") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Try NVMe
	out, err = exec.Command("system_profiler", "SPNVMeDataType").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Serial Number") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	return "DARWIN-DISK"
}

// maskString masks a string for display, showing only first N chars
func maskString(s string, show int) string {
	if len(s) <= show {
		return s
	}
	return s[:show] + "****"
}
