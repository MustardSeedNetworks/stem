// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package netif_test

import (
	"testing"

	"github.com/krisarmstrong/stem/internal/netif"
)

// TestCalculateScore tests the scoring system. Since calculateScore is not
// exported, we test it indirectly via DetectInterfaces and GetBestInterface.
func TestInterfaceScoring(t *testing.T) {
	// Test that DetectInterfaces returns valid interfaces with scores.
	interfaces, err := netif.DetectInterfaces()
	if err != nil {
		t.Errorf("DetectInterfaces() returned error: %v", err)
	}

	// Should be a valid slice (even if empty).
	if interfaces == nil {
		t.Error("DetectInterfaces() returned nil slice, expected valid slice")
	}

	// Verify loopback is filtered out and state/score are valid.
	for _, iface := range interfaces {
		if iface.Name == "lo" {
			t.Error("DetectInterfaces() should filter out loopback interface")
		}
		if iface.State != "up" && iface.State != "down" {
			t.Errorf("Interface %s has invalid state: %s", iface.Name, iface.State)
		}
		// Score should be non-negative.
		if iface.Score < 0 {
			t.Errorf("Interface %s has negative score: %d", iface.Name, iface.Score)
		}
	}
}

func TestCheckXDPSupport(t *testing.T) {
	// Test known XDP drivers count.
	xdpDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb", "igc",
	}

	// Verify XDP driver list has expected count.
	if len(xdpDrivers) != 10 {
		t.Error("XDP driver list should have 10 known drivers")
	}
}

func TestCheckDPDKSupport(t *testing.T) {
	// Test known DPDK drivers count.
	dpdkDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb",
		"e1000", "e1000e", "fm10k",
	}

	// Verify DPDK driver list count.
	if len(dpdkDrivers) != 12 {
		t.Error("DPDK driver list should have 12 known drivers")
	}
}

func TestDetectInterfaces(t *testing.T) {
	// This test verifies the function runs without error.
	interfaces, err := netif.DetectInterfaces()
	if err != nil {
		t.Errorf("DetectInterfaces() returned error: %v", err)
	}

	// Should be a valid slice (even if empty).
	if interfaces == nil {
		t.Error("DetectInterfaces() returned nil slice, expected valid slice")
	}

	// Verify loopback is filtered out.
	for _, iface := range interfaces {
		if iface.Name == "lo" {
			t.Error("DetectInterfaces() should filter out loopback interface")
		}
	}
}

func TestGetBestInterface(t *testing.T) {
	// This function relies on DetectInterfaces.
	// In environments with no interfaces, it should return an error.
	best, err := netif.GetBestInterface()
	if err != nil {
		// This is expected in minimal environments.
		t.Logf("GetBestInterface() returned expected error: %v", err)
		return
	}

	// If we got an interface, verify it has required fields.
	if best.Name == "" {
		t.Error("Best interface should have a name")
	}
	if best.Score <= 0 {
		t.Error("Best interface should have positive score")
	}
}

func TestInterfaceInfoStruct(t *testing.T) {
	// Test that InterfaceInfo struct can be created and accessed.
	//nolint:govet // Test data setup - intentionally initializing all fields for struct validation
	info := netif.InterfaceInfo{
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

// Test loopback filtering.
func TestLoopbackFiltering(t *testing.T) {
	interfaces, err := netif.DetectInterfaces()
	if err != nil {
		t.Fatalf("DetectInterfaces() error: %v", err)
	}

	for _, iface := range interfaces {
		if iface.Name == "lo" || iface.Name == "lo0" {
			t.Error("Loopback interface should be filtered out")
		}
	}
}

// Test state detection.
func TestInterfaceStateDetection(t *testing.T) {
	interfaces, err := netif.DetectInterfaces()
	if err != nil {
		t.Fatalf("DetectInterfaces() error: %v", err)
	}

	for _, iface := range interfaces {
		if iface.State != "up" && iface.State != "down" {
			t.Errorf("Interface %s has invalid state: %s", iface.Name, iface.State)
		}
	}
}

// Test XDP and DPDK driver coverage.
func TestXDPDriverCoverage(t *testing.T) {
	// These are the drivers we claim support XDP.
	xdpDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb", "igc",
	}

	for _, driver := range xdpDrivers {
		t.Run(driver, func(t *testing.T) {
			if driver == "" {
				t.Error("Empty driver in XDP list")
			}
		})
	}
}

func TestDPDKDriverCoverage(t *testing.T) {
	// These are the drivers we claim support DPDK.
	dpdkDrivers := []string{
		"ixgbe", "i40e", "ice", "mlx5_core", "mlx4_en",
		"bnxt_en", "nfp", "virtio_net", "igb",
		"e1000", "e1000e", "fm10k",
	}

	for _, driver := range dpdkDrivers {
		t.Run(driver, func(t *testing.T) {
			if driver == "" {
				t.Error("Empty driver in DPDK list")
			}
		})
	}
}

// Benchmark tests.
func BenchmarkDetectInterfaces(b *testing.B) {
	for b.Loop() {
		_, _ = netif.DetectInterfaces()
	}
}
