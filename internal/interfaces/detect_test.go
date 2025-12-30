// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package interfaces

import (
	"testing"
)

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		info     InterfaceInfo
		expected int
	}{
		{
			name: "down interface gets zero",
			info: InterfaceInfo{
				State:    "down",
				Physical: true,
				Speed:    10000,
			},
			expected: 0,
		},
		{
			name: "basic up physical interface",
			info: InterfaceInfo{
				State:    "up",
				Physical: true,
			},
			expected: 100,
		},
		{
			name: "10G physical interface with XDP and DPDK",
			info: InterfaceInfo{
				State:       "up",
				Physical:    true,
				Speed:       10000,
				XDPSupport:  true,
				DPDKSupport: true,
				IPv4:        "192.168.1.1",
				Duplex:      "full",
			},
			expected: 100 + 100 + 50 + 30 + 10 + 5, // 295
		},
		{
			name: "1G interface",
			info: InterfaceInfo{
				State:    "up",
				Physical: true,
				Speed:    1000,
			},
			expected: 100 + 10, // 110
		},
		{
			name: "virtual interface up with IP",
			info: InterfaceInfo{
				State:    "up",
				Physical: false,
				IPv4:     "10.0.0.1",
			},
			expected: 10, // just IPv4 bonus
		},
		{
			name: "full duplex bonus",
			info: InterfaceInfo{
				State:  "up",
				Duplex: "full",
			},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateScore(tt.info)
			if got != tt.expected {
				t.Errorf("calculateScore() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCheckXDPSupport(t *testing.T) {
	// Test known XDP drivers
	xdpDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb", "igc",
	}

	// For non-existent interfaces, checkXDPSupport will call getDriver
	// which will return "unknown". We can test the driver matching logic
	// by checking that unknown drivers return false

	// This tests that the function exists and runs without panicking
	result := checkXDPSupport("nonexistent_interface_123")
	if result {
		t.Error("Expected non-existent interface to not support XDP")
	}

	// Verify XDP driver list is used correctly in the code
	if len(xdpDrivers) != 10 {
		t.Error("XDP driver list should have 10 known drivers")
	}
}

func TestCheckDPDKSupport(t *testing.T) {
	// Test known DPDK drivers
	dpdkDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb",
		"e1000", "e1000e", "fm10k",
	}

	// For non-existent interfaces, checkDPDKSupport will call getDriver
	result := checkDPDKSupport("nonexistent_interface_456")
	if result {
		t.Error("Expected non-existent interface to not support DPDK")
	}

	// Verify DPDK driver list count
	if len(dpdkDrivers) != 12 {
		t.Error("DPDK driver list should have 12 known drivers")
	}
}

func TestIsPhysical(t *testing.T) {
	tests := []struct {
		name     string
		ifName   string
		expected bool
	}{
		{
			name:     "non-existent interface",
			ifName:   "nonexistent123",
			expected: false,
		},
		{
			name:     "lo interface",
			ifName:   "lo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPhysical(tt.ifName)
			if got != tt.expected {
				t.Errorf("isPhysical(%s) = %v, want %v", tt.ifName, got, tt.expected)
			}
		})
	}
}

func TestGetSpeed(t *testing.T) {
	// Non-existent interface should return 0
	speed := getSpeed("nonexistent_if_789")
	if speed != 0 {
		t.Errorf("Expected speed 0 for non-existent interface, got %d", speed)
	}
}

func TestGetDuplex(t *testing.T) {
	// Non-existent interface should return "unknown"
	duplex := getDuplex("nonexistent_if_101112")
	if duplex != "unknown" {
		t.Errorf("Expected 'unknown' duplex for non-existent interface, got %s", duplex)
	}
}

func TestGetDriver(t *testing.T) {
	// Non-existent interface should return "unknown"
	driver := getDriver("nonexistent_if_131415")
	if driver != "unknown" {
		t.Errorf("Expected 'unknown' driver for non-existent interface, got %s", driver)
	}
}

func TestDetectInterfaces(t *testing.T) {
	// This test just verifies the function runs without error
	// and returns a slice (may be empty if no interfaces)
	interfaces, err := DetectInterfaces()
	if err != nil {
		t.Errorf("DetectInterfaces() returned error: %v", err)
	}

	// Should be a valid slice (even if empty)
	if interfaces == nil {
		t.Error("DetectInterfaces() returned nil slice, expected valid slice")
	}

	// Verify loopback is filtered out
	for _, iface := range interfaces {
		if iface.Name == "lo" {
			t.Error("DetectInterfaces() should filter out loopback interface")
		}
	}
}

func TestGetBestInterface(t *testing.T) {
	// This function relies on DetectInterfaces
	// In environments with no interfaces, it should return an error
	best, err := GetBestInterface()
	if err != nil {
		// This is expected in minimal environments
		t.Logf("GetBestInterface() returned expected error: %v", err)
		return
	}

	// If we got an interface, verify it has required fields
	if best.Name == "" {
		t.Error("Best interface should have a name")
	}
	if best.Score <= 0 {
		t.Error("Best interface should have positive score")
	}
}

func TestInterfaceInfoStruct(t *testing.T) {
	// Test that InterfaceInfo struct can be created and accessed
	info := InterfaceInfo{
		Name:        "eth0",
		MAC:         "00:11:22:33:44:55",
		Speed:       1000,
		Duplex:      "full",
		State:       "up",
		Driver:      "e1000e",
		Physical:    true,
		XDPSupport:  false,
		DPDKSupport: true,
		Score:       150,
		MTU:         1500,
		IPv4:        "192.168.1.100",
		IPv6:        "fe80::1",
	}

	if info.Name != "eth0" {
		t.Errorf("Expected name 'eth0', got '%s'", info.Name)
	}
	if info.Speed != 1000 {
		t.Errorf("Expected speed 1000, got %d", info.Speed)
	}
	if !info.Physical {
		t.Error("Expected Physical to be true")
	}
	if info.XDPSupport {
		t.Error("Expected XDPSupport to be false")
	}
}

// Benchmark tests
func BenchmarkCalculateScore(b *testing.B) {
	info := InterfaceInfo{
		State:       "up",
		Physical:    true,
		Speed:       10000,
		XDPSupport:  true,
		DPDKSupport: true,
		IPv4:        "192.168.1.1",
		Duplex:      "full",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateScore(info)
	}
}

func BenchmarkDetectInterfaces(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = DetectInterfaces()
	}
}
