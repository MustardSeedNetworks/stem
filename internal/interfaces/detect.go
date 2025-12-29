// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package interfaces provides network interface detection and scoring.
//
// This package detects available network interfaces, gathers detailed
// information about each (speed, driver, XDP/DPDK support), and calculates
// a suitability score for network testing purposes.
package interfaces

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// InterfaceInfo contains detailed information about a network interface
type InterfaceInfo struct {
	Name        string `json:"name"`
	MAC         string `json:"mac"`
	Speed       int    `json:"speed"`    // Mbps
	Duplex      string `json:"duplex"`   // full, half, unknown
	State       string `json:"state"`    // up, down
	Driver      string `json:"driver"`   // ixgbe, mlx5_core, etc.
	Physical    bool   `json:"physical"` // true for real NICs
	XDPSupport  bool   `json:"xdp"`      // AF_XDP capable
	DPDKSupport bool   `json:"dpdk"`     // DPDK capable
	Score       int    `json:"score"`    // Auto-selection score
	MTU         int    `json:"mtu"`      // Maximum transmission unit
	IPv4        string `json:"ipv4"`     // Primary IPv4 address
	IPv6        string `json:"ipv6"`     // Primary IPv6 address
}

// DetectInterfaces returns a list of all network interfaces with detailed info
func DetectInterfaces() ([]InterfaceInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	var result []InterfaceInfo
	for _, iface := range interfaces {
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		info := InterfaceInfo{
			Name:     iface.Name,
			MAC:      iface.HardwareAddr.String(),
			MTU:      iface.MTU,
			Physical: isPhysical(iface.Name),
			Duplex:   "unknown",
		}

		// Check interface state
		if iface.Flags&net.FlagUp != 0 {
			info.State = "up"
		} else {
			info.State = "down"
		}

		// Get IP addresses
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ip, _, _ := net.ParseCIDR(addr.String())
			if ip == nil {
				continue
			}
			if ip.To4() != nil && info.IPv4 == "" {
				info.IPv4 = ip.String()
			} else if ip.To16() != nil && info.IPv6 == "" && !ip.IsLinkLocalUnicast() {
				info.IPv6 = ip.String()
			}
		}

		// Get Linux-specific info
		info.Speed = getSpeed(iface.Name)
		info.Duplex = getDuplex(iface.Name)
		info.Driver = getDriver(iface.Name)
		info.XDPSupport = checkXDPSupport(iface.Name)
		info.DPDKSupport = checkDPDKSupport(iface.Name)

		// Calculate score
		info.Score = calculateScore(info)

		result = append(result, info)
	}

	return result, nil
}

// isPhysical checks if the interface is a physical device
func isPhysical(name string) bool {
	// Check if device exists in /sys/class/net/<name>/device
	devicePath := filepath.Join("/sys/class/net", name, "device")
	_, err := os.Stat(devicePath)
	return err == nil
}

// getSpeed returns the interface speed in Mbps
func getSpeed(name string) int {
	speedFile := filepath.Join("/sys/class/net", name, "speed")
	data, err := os.ReadFile(speedFile)
	if err != nil {
		return 0
	}
	speed, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	if speed < 0 {
		return 0 // Speed is -1 if not available
	}
	return speed
}

// getDuplex returns the interface duplex mode
func getDuplex(name string) string {
	duplexFile := filepath.Join("/sys/class/net", name, "duplex")
	data, err := os.ReadFile(duplexFile)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

// getDriver returns the driver name for the interface
func getDriver(name string) string {
	driverPath := filepath.Join("/sys/class/net", name, "device", "driver")
	link, err := os.Readlink(driverPath)
	if err != nil {
		return "unknown"
	}
	return filepath.Base(link)
}

// checkXDPSupport checks if the interface supports AF_XDP
func checkXDPSupport(name string) bool {
	// Check for known XDP-capable drivers
	driver := getDriver(name)
	xdpDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb", "igc",
	}
	for _, d := range xdpDrivers {
		if driver == d {
			return true
		}
	}
	return false
}

// checkDPDKSupport checks if the interface can be used with DPDK
func checkDPDKSupport(name string) bool {
	// DPDK support is primarily determined by driver
	driver := getDriver(name)
	dpdkDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb",
		"e1000", "e1000e", "fm10k",
	}
	for _, d := range dpdkDrivers {
		if driver == d {
			return true
		}
	}
	return false
}

// calculateScore calculates the auto-selection score for an interface
func calculateScore(info InterfaceInfo) int {
	score := 0

	// Must be up to be considered
	if info.State != "up" {
		return 0
	}

	// Physical NICs strongly preferred
	if info.Physical {
		score += 100
	}

	// Higher speed = higher score (10G = +100, 1G = +10)
	score += info.Speed / 100

	// AF_XDP support bonus
	if info.XDPSupport {
		score += 50
	}

	// DPDK support bonus
	if info.DPDKSupport {
		score += 30
	}

	// Has IPv4 address (configured)
	if info.IPv4 != "" {
		score += 10
	}

	// Full duplex preferred
	if info.Duplex == "full" {
		score += 5
	}

	return score
}

// GetBestInterface returns the interface with the highest score
func GetBestInterface() (*InterfaceInfo, error) {
	interfaces, err := DetectInterfaces()
	if err != nil {
		return nil, err
	}

	if len(interfaces) == 0 {
		return nil, fmt.Errorf("no suitable interfaces found")
	}

	best := &interfaces[0]
	for i := range interfaces {
		if interfaces[i].Score > best.Score {
			best = &interfaces[i]
		}
	}

	if best.Score == 0 {
		return nil, fmt.Errorf("no interfaces with valid score found (all down?)")
	}

	return best, nil
}
