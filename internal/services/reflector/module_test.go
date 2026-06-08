// SPDX-License-Identifier: BUSL-1.1

package reflector_test

import (
	"slices"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/reflector"
)

// Test constants.
const (
	expectedModuleName    = "reflector"
	expectedDisplayName   = "Reflector"
	expectedColorHex      = "#0891b2"
	expectedStandardRef   = "Loopback/Echo"
	expectedTestTypeCount = 1
)

func TestNew(t *testing.T) {
	m := reflector.New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestModuleName(t *testing.T) {
	m := reflector.New()
	if got := m.Name(); got != expectedModuleName {
		t.Errorf("Name() = %q, want %q", got, expectedModuleName)
	}
	// Also verify constant matches.
	if reflector.ModuleName != expectedModuleName {
		t.Errorf("ModuleName constant = %q, want %q", reflector.ModuleName, expectedModuleName)
	}
}

func TestModuleDisplayName(t *testing.T) {
	m := reflector.New()
	if got := m.DisplayName(); got != expectedDisplayName {
		t.Errorf("DisplayName() = %q, want %q", got, expectedDisplayName)
	}
	// Also verify constant matches.
	if reflector.DisplayName != expectedDisplayName {
		t.Errorf("DisplayName constant = %q, want %q", reflector.DisplayName, expectedDisplayName)
	}
}

func TestModuleDescription(t *testing.T) {
	m := reflector.New()
	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	// Should mention packet reflection/loopback.
	if !containsSubstring(desc, "reflection") && !containsSubstring(desc, "loopback") {
		t.Error("Description() should mention reflection or loopback")
	}
	// Should mention Tier 1.
	if !containsSubstring(desc, "Tier 1") {
		t.Error("Description() should mention Tier 1 mode")
	}
}

func TestModuleColor(t *testing.T) {
	m := reflector.New()
	if got := m.Color(); got != expectedColorHex {
		t.Errorf("Color() = %q, want %q", got, expectedColorHex)
	}
	// Verify it's a valid hex color (cyan).
	if len(m.Color()) != 7 || m.Color()[0] != '#' {
		t.Errorf("Color() has invalid format: %s", m.Color())
	}
	// Also verify constant matches.
	if reflector.ColorHex != expectedColorHex {
		t.Errorf("ColorHex constant = %q, want %q", reflector.ColorHex, expectedColorHex)
	}
}

func TestModuleStandard(t *testing.T) {
	m := reflector.New()
	if got := m.Standard(); got != expectedStandardRef {
		t.Errorf("Standard() = %q, want %q", got, expectedStandardRef)
	}
	// Also verify constant matches.
	if reflector.StandardRef != expectedStandardRef {
		t.Errorf("StandardRef constant = %q, want %q", reflector.StandardRef, expectedStandardRef)
	}
}

func TestModuleTestTypes(t *testing.T) {
	m := reflector.New()
	tests := m.TestTypes()

	if len(tests) != expectedTestTypeCount {
		t.Errorf("TestTypes() returned %d tests, want %d", len(tests), expectedTestTypeCount)
	}

	// Verify the expected test type is present.
	expectedTypes := []string{"reflect"}

	for _, expected := range expectedTypes {
		if !slices.Contains(tests, expected) {
			t.Errorf("TestTypes() missing expected test type: %s", expected)
		}
	}
}

func TestModuleCanRun(t *testing.T) {
	m := reflector.New()

	// Test valid reflector operation type.
	if !m.CanRun("reflect") {
		t.Error("CanRun(\"reflect\") = false, want true")
	}

	// Test invalid test types - reflector should NOT run other module tests.
	invalidTests := []string{
		"rfc2544_throughput", // Benchmark module.
		"rfc2544_latency",    // Benchmark module.
		"rfc2544_frame_loss", // Benchmark module.
		"y1564",              // ServiceTest module.
		"y1564_config",       // ServiceTest module.
		"y1731_delay",        // Measure module.
		"rfc2889_forwarding", // Certify module.
		"rfc6349_throughput", // Certify module.
		"custom_stream",      // TrafficGen module.
		"invalid",            // Nonexistent.
		"",                   // Empty string.
		"reflector",          // Invalid (module name, not operation).
		"echo",               // Invalid (not the operation name).
		"loopback",           // Invalid (not the operation name).
	}
	for _, test := range invalidTests {
		if m.CanRun(test) {
			t.Errorf("CanRun(%q) = true, want false", test)
		}
	}
}

func TestModuleTestDescription(t *testing.T) {
	m := reflector.New()

	testCases := []struct {
		testType    string
		shouldExist bool
		contains    string
	}{
		{"reflect", true, "reflector"},
		{"invalid", false, ""},
		{"", false, ""},
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

func TestModuleReflectDescription(t *testing.T) {
	m := reflector.New()
	desc := m.TestDescription("reflect")

	// Verify description mentions key aspects.
	expectations := []string{
		"echo",   // Echoes packets.
		"remote", // Remote testing.
	}

	for _, expected := range expectations {
		if !containsSubstring(desc, expected) {
			t.Errorf("TestDescription(\"reflect\") = %q, should contain %q", desc, expected)
		}
	}
}

func TestModuleMultipleInstances(t *testing.T) {
	// Verify multiple instances are independent.
	m1 := reflector.New()
	m2 := reflector.New()

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
	if m1.Standard() != m2.Standard() {
		t.Error("Different instances should have same Standard()")
	}
}

func TestModuleIsOperationalMode(t *testing.T) {
	m := reflector.New()

	// Reflector is an operational mode, not a test module.
	// Standard should reflect this.
	if m.Standard() == "RFC 2544" || m.Standard() == "ITU-T Y.1564" {
		t.Error("Reflector should not be a standards-based test module")
	}

	// Should have only one operation type.
	if len(m.TestTypes()) != 1 {
		t.Errorf("Reflector should have exactly 1 operation type, got %d", len(m.TestTypes()))
	}
}

// containsSubstring checks if str contains substr.
func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
