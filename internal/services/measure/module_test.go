// SPDX-License-Identifier: BUSL-1.1

package measure_test

import (
	"slices"
	"testing"

	"github.com/krisarmstrong/stem/internal/services/measure"
)

// Test constants.
const (
	expectedModuleName    = "measure"
	expectedDisplayName   = "Measure"
	expectedColorHex      = "#2563eb"
	expectedStandardRef   = "ITU-T Y.1731"
	expectedTestTypeCount = 4
)

func TestNew(t *testing.T) {
	m := measure.New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestModuleName(t *testing.T) {
	m := measure.New()
	if got := m.Name(); got != expectedModuleName {
		t.Errorf("Name() = %q, want %q", got, expectedModuleName)
	}
	// Also verify constant matches.
	if measure.ModuleName != expectedModuleName {
		t.Errorf("ModuleName constant = %q, want %q", measure.ModuleName, expectedModuleName)
	}
}

func TestModuleDisplayName(t *testing.T) {
	m := measure.New()
	if got := m.DisplayName(); got != expectedDisplayName {
		t.Errorf("DisplayName() = %q, want %q", got, expectedDisplayName)
	}
	// Also verify constant matches.
	if measure.DisplayName != expectedDisplayName {
		t.Errorf("DisplayName constant = %q, want %q", measure.DisplayName, expectedDisplayName)
	}
}

func TestModuleDescription(t *testing.T) {
	m := measure.New()
	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	// Should mention Y.1731 and OAM capabilities.
	if !containsSubstring(desc, "Y.1731") {
		t.Error("Description() should mention Y.1731")
	}
}

func TestModuleColor(t *testing.T) {
	m := measure.New()
	if got := m.Color(); got != expectedColorHex {
		t.Errorf("Color() = %q, want %q", got, expectedColorHex)
	}
	// Verify it's a valid hex color.
	if len(m.Color()) != 7 || m.Color()[0] != '#' {
		t.Errorf("Color() has invalid format: %s", m.Color())
	}
	// Also verify constant matches.
	if measure.ColorHex != expectedColorHex {
		t.Errorf("ColorHex constant = %q, want %q", measure.ColorHex, expectedColorHex)
	}
}

func TestModuleStandard(t *testing.T) {
	m := measure.New()
	if got := m.Standard(); got != expectedStandardRef {
		t.Errorf("Standard() = %q, want %q", got, expectedStandardRef)
	}
	// Also verify constant matches.
	if measure.StandardRef != expectedStandardRef {
		t.Errorf("StandardRef constant = %q, want %q", measure.StandardRef, expectedStandardRef)
	}
}

func TestModuleTestTypes(t *testing.T) {
	m := measure.New()
	tests := m.TestTypes()

	if len(tests) != expectedTestTypeCount {
		t.Errorf("TestTypes() returned %d tests, want %d", len(tests), expectedTestTypeCount)
	}

	// Verify all expected test types are present.
	expectedTypes := []string{
		"y1731_delay",
		"y1731_loss",
		"y1731_slm",
		"y1731_loopback",
	}

	for _, expected := range expectedTypes {
		if !slices.Contains(tests, expected) {
			t.Errorf("TestTypes() missing expected test type: %s", expected)
		}
	}
}

func TestModuleCanRun(t *testing.T) {
	m := measure.New()

	// Test valid Y.1731 test types.
	validTests := []string{
		"y1731_delay",
		"y1731_loss",
		"y1731_slm",
		"y1731_loopback",
	}
	for _, test := range validTests {
		if !m.CanRun(test) {
			t.Errorf("CanRun(%q) = false, want true", test)
		}
	}

	// Test invalid test types.
	invalidTests := []string{
		"rfc2544_throughput", // Benchmark module.
		"rfc2544_latency",    // Benchmark module.
		"y1564",              // ServiceTest module.
		"rfc2889_forwarding", // Certify module.
		"custom_stream",      // TrafficGen module.
		"reflect",            // Reflector module.
		"invalid",            // Nonexistent.
		"",                   // Empty string.
		"y1731",              // Invalid (missing suffix).
		"y1731_unknown",      // Invalid Y.1731 test.
	}
	for _, test := range invalidTests {
		if m.CanRun(test) {
			t.Errorf("CanRun(%q) = true, want false", test)
		}
	}
}

func TestModuleTestDescription(t *testing.T) {
	m := measure.New()

	testCases := []struct {
		testType    string
		shouldExist bool
		contains    string
	}{
		{"y1731_delay", true, "Delay"},
		{"y1731_loss", true, "Loss"},
		{"y1731_slm", true, "Synthetic"},
		{"y1731_loopback", true, "Loopback"},
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

func TestModuleTestDescriptionY1731Protocol(t *testing.T) {
	m := measure.New()

	// Verify descriptions mention Y.1731 protocol message types.
	protocolMappings := map[string]string{
		"y1731_delay":    "DMM", // Delay Measurement Message.
		"y1731_loss":     "LMM", // Loss Measurement Message.
		"y1731_loopback": "LBM", // Loopback Message.
	}

	for testType, expectedProtocol := range protocolMappings {
		desc := m.TestDescription(testType)
		if !containsSubstring(desc, expectedProtocol) {
			t.Errorf("TestDescription(%q) = %q, should mention %s protocol", testType, desc, expectedProtocol)
		}
	}
}

func TestModuleMultipleInstances(t *testing.T) {
	// Verify multiple instances are independent.
	m1 := measure.New()
	m2 := measure.New()

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

// containsSubstring checks if str contains substr.
func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
