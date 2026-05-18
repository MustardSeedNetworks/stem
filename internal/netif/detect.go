// SPDX-License-Identifier: BUSL-1.1

// Package netif provides network interface detection and scoring.
//
// This package detects available network interfaces, gathers detailed
// information about each (speed, driver, XDP/DPDK support), and calculates
// a suitability score for network testing purposes.
//
// # Platform Dependencies
//
// This package relies on the Linux sysfs virtual filesystem for interface
// metadata. On non-Linux platforms, most detection functions return
// fallback values. The specific sysfs paths used are:
//
//   - /sys/class/net/<iface>/speed       - Link speed in Mbps
//   - /sys/class/net/<iface>/duplex      - Duplex mode (full/half)
//   - /sys/class/net/<iface>/device      - Physical device presence
//   - /sys/class/net/<iface>/device/driver - Driver symlink
//
// These paths require the interface to be up and linked for speed/duplex.
// Virtual interfaces (bridges, bonds, vlans) lack the device directory.
//
// # Capability Detection
//
// XDP and DPDK support are detected using driver name heuristics. This is
// because there is no reliable runtime API to query these capabilities.
// The detection uses known lists of drivers that support each technology:
//
// XDP-capable drivers:
//
//	ixgbe, i40e, ice, mlx5_core, mlx4_en, bnxt_en, nfp, virtio_net, igb, igc
//
// DPDK-capable drivers:
//
//	ixgbe, i40e, ice, mlx5_core, mlx4_en, bnxt_en, nfp, virtio_net, igb, e1000, e1000e, fm10k
//
// Limitations of driver heuristics:
//   - May report false positives if driver is present but XDP/DPDK unavailable
//   - May report false negatives for newer drivers not in the list
//   - Does not verify kernel XDP support or DPDK library availability
//   - Operators should verify capabilities before relying on them
//
// # Interface Scoring
//
// The [GetBestInterface] function selects interfaces using a scoring system:
//
//	+100 points:     Physical interface that is UP
//	+speed/100:      Speed bonus (10G = +100, 1G = +10, 100G = +1000)
//	+50 points:      XDP support (driver-based detection)
//	+30 points:      DPDK support (driver-based detection)
//	+10 points:      Has IPv4 address assigned
//	+5 points:       Full duplex mode
//
// The highest-scoring interface is selected. In case of ties, the first
// interface (as returned by the OS) wins. Operators can override the
// automatic selection by specifying an interface explicitly.
//
// # Usage Notes
//
// For best results:
//   - Ensure target interface is UP before detection
//   - On non-Linux systems, expect reduced functionality
//   - Use explicit interface selection for production deployments
//   - Verify XDP/DPDK capabilities independently if critical
package netif

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Scoring constants for interface auto-selection.
// These values are used by calculateScore to rank interfaces.
const (
	// scorePhysical is awarded to physical NICs (vs virtual interfaces).
	scorePhysical = 100
	// scoreXDPSupport is awarded to interfaces with AF_XDP capability.
	scoreXDPSupport = 50
	// scoreDPDKSupport is awarded to interfaces with DPDK capability.
	scoreDPDKSupport = 30
	// scoreHasIPv4 is awarded to interfaces with an IPv4 address assigned.
	scoreHasIPv4 = 10
	// scoreFullDuplex is awarded to interfaces running in full duplex mode.
	scoreFullDuplex = 5
	// speedDivisor is used to convert Mbps to score points (10G = +100).
	speedDivisor = 100
)

// valueUnknown is the default value for unknown interface properties.
const valueUnknown = "unknown"

// InterfaceInfo contains detailed information about a network interface.
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
	Usable      bool   `json:"usable"`   // true if the interface is plausibly testable
}

// isVirtualInterfaceName returns true when name starts with a known
// virtual-interface prefix — virtual / platform-managed interfaces
// that should not be exposed by default when the UI filters for
// "usable" interfaces. Operators can still opt to show all from the UI.
// Prefix table is local to this function rather than a package global
// (linter: gochecknoglobals).
func isVirtualInterfaceName(name string) bool {
	prefixes := [...]string{
		// macOS virtual interfaces.
		"utun", "awdl", "llw", "anpi", "ap", "bridge", "stf", "gif", "p2p",
		// Linux container/tunnel/bridge surfaces.
		"docker", "br-", "veth", "virbr", "tun", "tap", "vlan", "wg",
		// Hypervisor/VPN virtual NICs.
		"vmnet", "vboxnet", "vnic",
	}
	lower := strings.ToLower(name)
	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

// computeUsable returns true when an interface is plausibly testable:
//   - state is "up" (carrier present)
//   - has either an IPv4 or a non-link-local IPv6 address
//   - is not a known virtual prefix (vmnet*, utun*, awdl*, docker*, ...)
//
// This is advisory only — the UI uses it to filter the default
// interface dropdown. Operators can still pick "show all" to reveal
// every detected interface.
func computeUsable(info InterfaceInfo) bool {
	if info.State != "up" {
		return false
	}
	if info.IPv4 == "" && info.IPv6 == "" {
		return false
	}
	if isVirtualInterfaceName(info.Name) {
		return false
	}
	return true
}

// extractIPAddresses extracts IPv4 and IPv6 addresses from interface addresses.
func extractIPAddresses(addrs []net.Addr) (string, string) {
	var ipv4Addr string
	var ipv6Addr string
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil || ip == nil {
			continue
		}
		if ip.To4() != nil && ipv4Addr == "" {
			ipv4Addr = ip.String()
		} else if ip.To16() != nil && ipv6Addr == "" && !ip.IsLinkLocalUnicast() {
			ipv6Addr = ip.String()
		}
	}
	return ipv4Addr, ipv6Addr
}

// DetectInterfaces returns a list of all network interfaces with detailed info.
func DetectInterfaces() ([]InterfaceInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	result := make([]InterfaceInfo, 0, len(interfaces))
	for _, iface := range interfaces {
		// Skip loopback.
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		info := InterfaceInfo{
			Name:        iface.Name,
			MAC:         iface.HardwareAddr.String(),
			Speed:       0,
			Duplex:      valueUnknown,
			State:       "",
			Driver:      "",
			Physical:    isPhysical(iface.Name),
			XDPSupport:  false,
			DPDKSupport: false,
			Score:       0,
			MTU:         iface.MTU,
			IPv4:        "",
			IPv6:        "",
			Usable:      false,
		}

		// Check interface state.
		if iface.Flags&net.FlagUp != 0 {
			info.State = "up"
		} else {
			info.State = "down"
		}

		// Get IP addresses (non-critical - interface may be up without addresses).
		addrs, _ := iface.Addrs() // Ignore error: addresses are optional.
		info.IPv4, info.IPv6 = extractIPAddresses(addrs)

		// Get Linux-specific info.
		info.Speed = getSpeed(iface.Name)
		info.Duplex = getDuplex(iface.Name)
		info.Driver = getDriver(iface.Name)
		info.XDPSupport = checkXDPSupport(iface.Name)
		info.DPDKSupport = checkDPDKSupport(iface.Name)

		// Calculate score.
		info.Score = calculateScore(info)

		// Compute the advisory "usable" flag (UI default filter).
		info.Usable = computeUsable(info)

		result = append(result, info)
	}

	return result, nil
}

// sysfsBasePath is the root path for sysfs network interface files.
const sysfsBasePath = "/sys/class/net"

// errNoSuitableInterfaces is returned when no suitable interfaces are found.
var errNoSuitableInterfaces = errors.New("no suitable interfaces found")

// errNoValidScoreInterfaces is returned when all interfaces have zero score.
var errNoValidScoreInterfaces = errors.New("no interfaces with valid score found (all down?)")

// sysfsNetPath returns a validated sysfs path for the given interface name.
// It prevents path traversal attacks by rejecting names with slashes or dots.
func sysfsNetPath(name string, subpath ...string) (string, error) {
	// Validate interface name doesn't contain path traversal characters.
	if strings.ContainsAny(name, "/\\..") || name == "" {
		return "", fmt.Errorf("invalid interface name: %s", name)
	}
	parts := append([]string{sysfsBasePath, name}, subpath...)
	fullPath := filepath.Clean(filepath.Join(parts...))
	// Verify the path is still within sysfs (defense in depth).
	if !strings.HasPrefix(fullPath, sysfsBasePath+"/") {
		return "", fmt.Errorf("path escapes sysfs: %s", fullPath)
	}
	return fullPath, nil
}

// readSysfsFile reads a file from the sysfs network interface directory.
// It validates the path to prevent path traversal attacks.
func readSysfsFile(name string, subpath ...string) ([]byte, error) {
	// Validate interface name doesn't contain path traversal characters.
	if strings.ContainsAny(name, "/\\..") || name == "" {
		return nil, fmt.Errorf("invalid interface name: %s", name)
	}
	for _, sp := range subpath {
		if strings.ContainsAny(sp, "/\\..") {
			return nil, fmt.Errorf("invalid subpath: %s", sp)
		}
	}
	// Build and validate the path.
	parts := append([]string{sysfsBasePath, name}, subpath...)
	fullPath := filepath.Clean(filepath.Join(parts...))
	if !strings.HasPrefix(fullPath, sysfsBasePath+"/") {
		return nil, fmt.Errorf("path escapes sysfs: %s", fullPath)
	}
	// Open the validated system file.
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("open sysfs file %s: %w", fullPath, err)
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read sysfs file %s: %w", fullPath, err)
	}
	return data, nil
}

// isPhysical checks if the interface is a physical device.
func isPhysical(name string) bool {
	// Check if device exists in /sys/class/net/<name>/device.
	devicePath, err := sysfsNetPath(name, "device")
	if err != nil {
		return false
	}
	_, err = os.Stat(devicePath)
	return err == nil
}

// getSpeed returns the interface speed in Mbps.
func getSpeed(name string) int {
	data, err := readSysfsFile(name, "speed")
	if err != nil {
		return 0
	}
	speed, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	if speed < 0 {
		return 0 // Speed is -1 if not available
	}
	return speed
}

// getDuplex returns the interface duplex mode.
func getDuplex(name string) string {
	data, err := readSysfsFile(name, "duplex")
	if err != nil {
		return valueUnknown
	}
	return strings.TrimSpace(string(data))
}

// getDriver returns the driver name for the interface.
func getDriver(name string) string {
	driverPath, err := sysfsNetPath(name, "device", "driver")
	if err != nil {
		return valueUnknown
	}
	link, err := os.Readlink(driverPath)
	if err != nil {
		return valueUnknown
	}
	return filepath.Base(link)
}

// checkXDPSupport checks if the interface supports AF_XDP.
func checkXDPSupport(name string) bool {
	// Check for known XDP-capable drivers.
	driver := getDriver(name)
	xdpDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb", "igc",
	}
	return slices.Contains(xdpDrivers, driver)
}

// checkDPDKSupport checks if the interface can be used with DPDK.
func checkDPDKSupport(name string) bool {
	// DPDK support is primarily determined by driver.
	driver := getDriver(name)
	dpdkDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb",
		"e1000", "e1000e", "fm10k",
	}
	return slices.Contains(dpdkDrivers, driver)
}

// calculateScore calculates the auto-selection score for an interface.
func calculateScore(info InterfaceInfo) int {
	score := 0

	// Must be up to be considered.
	if info.State != "up" {
		return 0
	}

	// Physical NICs strongly preferred.
	if info.Physical {
		score += scorePhysical
	}

	// Higher speed = higher score (10G = +100, 1G = +10).
	score += info.Speed / speedDivisor

	// AF_XDP support bonus.
	if info.XDPSupport {
		score += scoreXDPSupport
	}

	// DPDK support bonus.
	if info.DPDKSupport {
		score += scoreDPDKSupport
	}

	// Has IPv4 address (configured).
	if info.IPv4 != "" {
		score += scoreHasIPv4
	}

	// Full duplex preferred.
	if info.Duplex == "full" {
		score += scoreFullDuplex
	}

	return score
}

// GetBestInterface returns the interface with the highest score.
func GetBestInterface() (*InterfaceInfo, error) {
	interfaces, err := DetectInterfaces()
	if err != nil {
		return nil, err
	}

	if len(interfaces) == 0 {
		return nil, errNoSuitableInterfaces
	}

	best := &interfaces[0]
	for i := range interfaces {
		if interfaces[i].Score > best.Score {
			best = &interfaces[i]
		}
	}

	if best.Score == 0 {
		return nil, errNoValidScoreInterfaces
	}

	return best, nil
}
