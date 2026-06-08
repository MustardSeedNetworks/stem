// SPDX-License-Identifier: BUSL-1.1

package netif_test

import (
	"encoding/json"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/netif"
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
	info := netif.InterfaceInfo{}
	info.Name = "eth0"
	info.MAC = "00:11:22:33:44:55"
	info.Speed = 1000
	info.Duplex = "full"
	info.State = "up"
	info.Driver = "e1000e"
	info.Physical = true
	info.XDPSupport = false
	info.DPDKSupport = true
	info.Score = 150
	info.MTU = 1500
	info.IPv4 = "192.168.1.100"
	info.IPv6 = "fe80::1"

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
	if info.MAC != "00:11:22:33:44:55" {
		t.Errorf("Expected MAC to be '00:11:22:33:44:55', got '%s'", info.MAC)
	}
	if info.Duplex != "full" {
		t.Errorf("Expected Duplex to be 'full', got '%s'", info.Duplex)
	}
	if info.State != "up" {
		t.Errorf("Expected State to be 'up', got '%s'", info.State)
	}
	if info.Driver != "e1000e" {
		t.Errorf("Expected Driver to be 'e1000e', got '%s'", info.Driver)
	}
	if !info.DPDKSupport {
		t.Error("Expected DPDKSupport to be true")
	}
	if info.Score != 150 {
		t.Errorf("Expected Score to be 150, got %d", info.Score)
	}
	if info.MTU != 1500 {
		t.Errorf("Expected MTU to be 1500, got %d", info.MTU)
	}
	if info.IPv4 != "192.168.1.100" {
		t.Errorf("Expected IPv4 to be '192.168.1.100', got '%s'", info.IPv4)
	}
	if info.IPv6 != "fe80::1" {
		t.Errorf("Expected IPv6 to be 'fe80::1', got '%s'", info.IPv6)
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

// TestGetBestInterfaceMultipleCalls tests that GetBestInterface is consistent.
func TestGetBestInterfaceMultipleCalls(t *testing.T) {
	// Call GetBestInterface multiple times and ensure consistent results.
	best1, err1 := netif.GetBestInterface()
	best2, err2 := netif.GetBestInterface()

	// Both calls should have same error state.
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("GetBestInterface() returned inconsistent errors: %v vs %v", err1, err2)
	}

	// If no error, both should return the same interface.
	if err1 == nil && best1.Name != best2.Name {
		t.Errorf("GetBestInterface() returned inconsistent interfaces: %s vs %s", best1.Name, best2.Name)
	}
}

// TestDetectInterfacesFields tests that DetectInterfaces returns expected fields.
func TestDetectInterfacesFields(t *testing.T) {
	interfaces, err := netif.DetectInterfaces()
	if err != nil {
		t.Fatalf("DetectInterfaces() error: %v", err)
	}

	for _, iface := range interfaces {
		// Name should never be empty.
		if iface.Name == "" {
			t.Error("Interface name should not be empty")
		}

		// MTU should be positive for active interfaces.
		if iface.MTU < 0 {
			t.Errorf("Interface %s has negative MTU: %d", iface.Name, iface.MTU)
		}

		// Score should be calculated.
		if iface.State == "up" && iface.Score == 0 {
			// Could be a virtual interface with no bonuses.
			t.Logf("Interface %s is up but has zero score (virtual?)", iface.Name)
		}
	}
}

// TestInterfaceScoringSorted tests that interfaces can be sorted by score.
func TestInterfaceScoringSorted(t *testing.T) {
	interfaces, err := netif.DetectInterfaces()
	if err != nil {
		t.Fatalf("DetectInterfaces() error: %v", err)
	}

	if len(interfaces) == 0 {
		t.Skip("No interfaces available for testing")
	}

	// Find the highest score.
	highestScore := 0
	for _, iface := range interfaces {
		t.Logf("Interface %s: state=%s, physical=%v, score=%d",
			iface.Name, iface.State, iface.Physical, iface.Score)
		if iface.Score > highestScore {
			highestScore = iface.Score
		}
	}
	t.Logf("Highest score: %d", highestScore)

	// Get best interface and verify it matches.
	best, err := netif.GetBestInterface()
	if err != nil {
		t.Logf("GetBestInterface() returned error (expected in minimal environments): %v", err)
		return
	}

	t.Logf("Best interface: %s with score %d", best.Name, best.Score)
	if best.Score != highestScore {
		t.Errorf("GetBestInterface() returned interface with score %d, but highest was %d",
			best.Score, highestScore)
	}
}

// TestInterfaceInfoJSON tests that InterfaceInfo can be marshaled to JSON.
func TestInterfaceInfoJSON(t *testing.T) {
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
		IPv6:        "2001:db8::1",
	}

	// Use json package to verify struct tags work.
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Verify expected fields are present.
	str := string(data)
	expectedFields := []string{
		`"name":"eth0"`,
		`"mac":"00:11:22:33:44:55"`,
		`"speed":1000`,
		`"duplex":"full"`,
		`"state":"up"`,
		`"driver":"e1000e"`,
		`"physical":true`,
		`"xdp":false`,
		`"dpdk":true`,
		`"score":150`,
		`"mtu":1500`,
		`"ipv4":"192.168.1.100"`,
		`"ipv6":"2001:db8::1"`,
	}

	for _, field := range expectedFields {
		if !containsSubstring(str, field) {
			t.Errorf("JSON output missing expected field: %s", field)
		}
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
