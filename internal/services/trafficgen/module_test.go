// SPDX-License-Identifier: BUSL-1.1

package trafficgen_test

import (
	"slices"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/trafficgen"
)

// Test constants.
const (
	expectedModuleName    = "trafficgen"
	expectedDisplayName   = "TrafficGen"
	expectedColorHex      = "#ca8a04"
	expectedStandardRef   = "Custom Traffic"
	expectedTestTypeCount = 1
)

func TestNew(t *testing.T) {
	m := trafficgen.New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestModuleName(t *testing.T) {
	m := trafficgen.New()
	if got := m.Name(); got != expectedModuleName {
		t.Errorf("Name() = %q, want %q", got, expectedModuleName)
	}
	// Also verify constant matches.
	if trafficgen.ModuleName != expectedModuleName {
		t.Errorf("ModuleName constant = %q, want %q", trafficgen.ModuleName, expectedModuleName)
	}
}

func TestModuleDisplayName(t *testing.T) {
	m := trafficgen.New()
	if got := m.DisplayName(); got != expectedDisplayName {
		t.Errorf("DisplayName() = %q, want %q", got, expectedDisplayName)
	}
	// Also verify constant matches.
	if trafficgen.DisplayName != expectedDisplayName {
		t.Errorf("DisplayName constant = %q, want %q", trafficgen.DisplayName, expectedDisplayName)
	}
}

func TestModuleDescription(t *testing.T) {
	m := trafficgen.New()
	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	// Should mention custom traffic generation.
	if !containsSubstring(desc, "traffic") {
		t.Error("Description() should mention traffic")
	}
	if !containsSubstring(desc, "stream") || !containsSubstring(desc, "pattern") {
		t.Error("Description() should mention stream or pattern")
	}
}

func TestModuleColor(t *testing.T) {
	m := trafficgen.New()
	if got := m.Color(); got != expectedColorHex {
		t.Errorf("Color() = %q, want %q", got, expectedColorHex)
	}
	// Verify it's a valid hex color (yellow).
	if len(m.Color()) != 7 || m.Color()[0] != '#' {
		t.Errorf("Color() has invalid format: %s", m.Color())
	}
	// Also verify constant matches.
	if trafficgen.ColorHex != expectedColorHex {
		t.Errorf("ColorHex constant = %q, want %q", trafficgen.ColorHex, expectedColorHex)
	}
}

func TestModuleStandard(t *testing.T) {
	m := trafficgen.New()
	if got := m.Standard(); got != expectedStandardRef {
		t.Errorf("Standard() = %q, want %q", got, expectedStandardRef)
	}
	// Also verify constant matches.
	if trafficgen.StandardRef != expectedStandardRef {
		t.Errorf("StandardRef constant = %q, want %q", trafficgen.StandardRef, expectedStandardRef)
	}
}

func TestModuleTestTypes(t *testing.T) {
	m := trafficgen.New()
	tests := m.TestTypes()

	if len(tests) != expectedTestTypeCount {
		t.Errorf("TestTypes() returned %d tests, want %d", len(tests), expectedTestTypeCount)
	}

	// Verify the expected test type is present.
	expectedTypes := []string{"custom_stream"}

	for _, expected := range expectedTypes {
		if !slices.Contains(tests, expected) {
			t.Errorf("TestTypes() missing expected test type: %s", expected)
		}
	}
}

func TestModuleCanRun(t *testing.T) {
	m := trafficgen.New()

	// Test valid trafficgen test type.
	if !m.CanRun("custom_stream") {
		t.Error("CanRun(\"custom_stream\") = false, want true")
	}

	// Test invalid test types - trafficgen should NOT run other module tests.
	invalidTests := []string{
		"rfc2544_throughput", // Benchmark module.
		"rfc2544_latency",    // Benchmark module.
		"rfc2544_frame_loss", // Benchmark module.
		"y1564",              // ServiceTest module.
		"y1564_config",       // ServiceTest module.
		"y1731_delay",        // Measure module.
		"rfc2889_forwarding", // Certify module.
		"rfc6349_throughput", // Certify module.
		"reflect",            // Reflector module (moved from trafficgen).
		"invalid",            // Nonexistent.
		"",                   // Empty string.
		"trafficgen",         // Invalid (module name, not test type).
		"stream",             // Invalid (not the full test type).
		"custom",             // Invalid (not the full test type).
	}
	for _, test := range invalidTests {
		if m.CanRun(test) {
			t.Errorf("CanRun(%q) = true, want false", test)
		}
	}
}

func TestModuleTestDescription(t *testing.T) {
	m := trafficgen.New()

	testCases := []struct {
		testType    string
		shouldExist bool
		contains    string
	}{
		{"custom_stream", true, "traffic"},
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

func TestModuleCustomStreamDescription(t *testing.T) {
	m := trafficgen.New()
	desc := m.TestDescription("custom_stream")

	// Verify description mentions key aspects.
	expectations := []string{
		"Custom",  // Custom traffic.
		"pattern", // Configurable patterns.
	}

	for _, expected := range expectations {
		if !containsSubstring(desc, expected) {
			t.Errorf("TestDescription(\"custom_stream\") = %q, should contain %q", desc, expected)
		}
	}
}

func TestModuleMultipleInstances(t *testing.T) {
	// Verify multiple instances are independent.
	m1 := trafficgen.New()
	m2 := trafficgen.New()

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

func TestModuleReflectNotInTrafficGen(t *testing.T) {
	m := trafficgen.New()

	// Verify that "reflect" is NOT in trafficgen (it was moved to reflector module).
	tests := m.TestTypes()
	if slices.Contains(tests, "reflect") {
		t.Error("TrafficGen should not contain 'reflect' test type (moved to Reflector module)")
	}

	if m.CanRun("reflect") {
		t.Error("TrafficGen.CanRun(\"reflect\") should return false")
	}
}

func TestModuleIsCustomTrafficGenerator(t *testing.T) {
	m := trafficgen.New()

	// TrafficGen is for custom traffic, not standards-based testing.
	if m.Standard() == "RFC 2544" || m.Standard() == "ITU-T Y.1564" {
		t.Error("TrafficGen should not be a standards-based test module")
	}

	// Should have only one test type (custom_stream).
	if len(m.TestTypes()) != 1 {
		t.Errorf("TrafficGen should have exactly 1 test type, got %d", len(m.TestTypes()))
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
