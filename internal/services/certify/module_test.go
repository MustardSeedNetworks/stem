// SPDX-License-Identifier: BUSL-1.1

package certify_test

import (
	"slices"
	"testing"

	"github.com/krisarmstrong/stem/internal/services/certify"
)

// Test constants.
const (
	expectedModuleName    = "certify"
	expectedDisplayName   = "Certify"
	expectedColorHex      = "#16a34a"
	expectedStandardRef   = "RFC 2889/6349/TSN"
	expectedTestTypeCount = 11
)

func TestNew(t *testing.T) {
	m := certify.New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestModuleName(t *testing.T) {
	m := certify.New()
	if got := m.Name(); got != expectedModuleName {
		t.Errorf("Name() = %q, want %q", got, expectedModuleName)
	}
	// Also verify constant matches.
	if certify.ModuleName != expectedModuleName {
		t.Errorf("ModuleName constant = %q, want %q", certify.ModuleName, expectedModuleName)
	}
}

func TestModuleDisplayName(t *testing.T) {
	m := certify.New()
	if got := m.DisplayName(); got != expectedDisplayName {
		t.Errorf("DisplayName() = %q, want %q", got, expectedDisplayName)
	}
	// Also verify constant matches.
	if certify.DisplayName != expectedDisplayName {
		t.Errorf("DisplayName constant = %q, want %q", certify.DisplayName, expectedDisplayName)
	}
}

func TestModuleDescription(t *testing.T) {
	m := certify.New()
	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	// Should mention the standards.
	if !containsSubstring(desc, "RFC 2889") {
		t.Error("Description() should mention RFC 2889")
	}
	if !containsSubstring(desc, "RFC 6349") {
		t.Error("Description() should mention RFC 6349")
	}
	if !containsSubstring(desc, "TSN") {
		t.Error("Description() should mention TSN")
	}
}

func TestModuleColor(t *testing.T) {
	m := certify.New()
	if got := m.Color(); got != expectedColorHex {
		t.Errorf("Color() = %q, want %q", got, expectedColorHex)
	}
	// Verify it's a valid hex color.
	if len(m.Color()) != 7 || m.Color()[0] != '#' {
		t.Errorf("Color() has invalid format: %s", m.Color())
	}
	// Also verify constant matches.
	if certify.ColorHex != expectedColorHex {
		t.Errorf("ColorHex constant = %q, want %q", certify.ColorHex, expectedColorHex)
	}
}

func TestModuleStandard(t *testing.T) {
	m := certify.New()
	if got := m.Standard(); got != expectedStandardRef {
		t.Errorf("Standard() = %q, want %q", got, expectedStandardRef)
	}
	// Also verify constant matches.
	if certify.StandardRef != expectedStandardRef {
		t.Errorf("StandardRef constant = %q, want %q", certify.StandardRef, expectedStandardRef)
	}
}

func TestModuleTestTypes(t *testing.T) {
	m := certify.New()
	tests := m.TestTypes()

	if len(tests) != expectedTestTypeCount {
		t.Errorf("TestTypes() returned %d tests, want %d", len(tests), expectedTestTypeCount)
	}

	// Verify all expected test types are present.
	expectedTypes := []string{
		// RFC 2889 LAN Switch.
		"rfc2889_forwarding",
		"rfc2889_caching",
		"rfc2889_learning",
		"rfc2889_broadcast",
		"rfc2889_congestion",
		// RFC 6349 TCP.
		"rfc6349_throughput",
		"rfc6349_path",
		// TSN 802.1Qbv.
		"tsn_timing",
		"tsn_isolation",
		"tsn_latency",
		"tsn",
	}

	for _, expected := range expectedTypes {
		if !slices.Contains(tests, expected) {
			t.Errorf("TestTypes() missing expected test type: %s", expected)
		}
	}
}

func TestModuleCanRun(t *testing.T) {
	m := certify.New()

	// Test valid RFC 2889 test types.
	rfc2889Tests := []string{
		"rfc2889_forwarding",
		"rfc2889_caching",
		"rfc2889_learning",
		"rfc2889_broadcast",
		"rfc2889_congestion",
	}
	for _, test := range rfc2889Tests {
		if !m.CanRun(test) {
			t.Errorf("CanRun(%q) = false, want true (RFC 2889)", test)
		}
	}

	// Test valid RFC 6349 test types.
	rfc6349Tests := []string{"rfc6349_throughput", "rfc6349_path"}
	for _, test := range rfc6349Tests {
		if !m.CanRun(test) {
			t.Errorf("CanRun(%q) = false, want true (RFC 6349)", test)
		}
	}

	// Test valid TSN test types.
	tsnTests := []string{"tsn_timing", "tsn_isolation", "tsn_latency", "tsn"}
	for _, test := range tsnTests {
		if !m.CanRun(test) {
			t.Errorf("CanRun(%q) = false, want true (TSN)", test)
		}
	}

	// Test invalid test types.
	invalidTests := []string{
		"rfc2544_throughput", // Benchmark module.
		"rfc2544_latency",    // Benchmark module.
		"y1564",              // ServiceTest module.
		"y1731_delay",        // Measure module.
		"custom_stream",      // TrafficGen module.
		"reflect",            // Reflector module.
		"invalid",            // Nonexistent.
		"",                   // Empty string.
		"rfc2889",            // Invalid (missing suffix).
		"rfc6349",            // Invalid (missing suffix).
		"tsn_unknown",        // Invalid TSN test.
	}
	for _, test := range invalidTests {
		if m.CanRun(test) {
			t.Errorf("CanRun(%q) = true, want false", test)
		}
	}
}

func TestModuleTestDescription(t *testing.T) {
	m := certify.New()

	testCases := []struct {
		testType    string
		shouldExist bool
		contains    string
	}{
		{"rfc2889_forwarding", true, "RFC 2889"},
		{"rfc2889_caching", true, "caching"},
		{"rfc2889_learning", true, "learning"},
		{"rfc2889_broadcast", true, "Broadcast"},
		{"rfc2889_congestion", true, "Congestion"},
		{"rfc6349_throughput", true, "TCP"},
		{"rfc6349_path", true, "Path"},
		{"tsn_timing", true, "802.1Qbv"},
		{"tsn_isolation", true, "isolation"},
		{"tsn_latency", true, "latency"},
		{"tsn", true, "TSN"},
		{"invalid", false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.testType, func(t *testing.T) {
			desc := m.TestDescription(tc.testType)
			if tc.shouldExist {
				if desc == "" {
					t.Errorf("TestDescription(%q) returned empty string", tc.testType)
				}
				if tc.contains != "" && !containsSubstring(desc, tc.contains) {
					t.Errorf("TestDescription(%q) = %q, should contain %q", tc.testType, desc, tc.contains)
				}
			} else if desc != "" {
				t.Errorf("TestDescription(%q) = %q, want empty string for invalid type", tc.testType, desc)
			}
		})
	}
}

func TestModuleMultipleInstances(t *testing.T) {
	// Verify multiple instances are independent.
	m1 := certify.New()
	m2 := certify.New()

	if m1 == m2 {
		t.Error("New() should return distinct instances")
	}

	// But they should return the same values.
	if m1.Name() != m2.Name() {
		t.Error("Different instances should have same Name()")
	}
	if m1.Color() != m2.Color() {
		t.Error("Different instances should have same Color()")
	}
}

// containsSubstring checks if str contains substr (case-insensitive would require strings pkg).
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || containsExact(str, substr))
}

func containsExact(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
