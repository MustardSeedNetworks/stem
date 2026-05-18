// SPDX-License-Identifier: BUSL-1.1

package servicetest_test

import (
	"slices"
	"testing"

	"github.com/krisarmstrong/stem/internal/services/servicetest"
)

// Test constants.
const (
	expectedModuleName    = "servicetest"
	expectedDisplayName   = "ServiceTest"
	expectedColorHex      = "#ea580c"
	expectedStandardRef   = "ITU-T Y.1564"
	expectedTestTypeCount = 6
)

func TestNew(t *testing.T) {
	m := servicetest.New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestModuleName(t *testing.T) {
	m := servicetest.New()
	if got := m.Name(); got != expectedModuleName {
		t.Errorf("Name() = %q, want %q", got, expectedModuleName)
	}
	// Also verify constant matches.
	if servicetest.ModuleName != expectedModuleName {
		t.Errorf("ModuleName constant = %q, want %q", servicetest.ModuleName, expectedModuleName)
	}
}

func TestModuleDisplayName(t *testing.T) {
	m := servicetest.New()
	if got := m.DisplayName(); got != expectedDisplayName {
		t.Errorf("DisplayName() = %q, want %q", got, expectedDisplayName)
	}
	// Also verify constant matches.
	if servicetest.DisplayName != expectedDisplayName {
		t.Errorf("DisplayName constant = %q, want %q", servicetest.DisplayName, expectedDisplayName)
	}
}

func TestModuleDescription(t *testing.T) {
	m := servicetest.New()
	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	// Should mention Y.1564 and MEF.
	if !containsSubstring(desc, "Y.1564") {
		t.Error("Description() should mention Y.1564")
	}
	if !containsSubstring(desc, "MEF") {
		t.Error("Description() should mention MEF")
	}
}

func TestModuleColor(t *testing.T) {
	m := servicetest.New()
	if got := m.Color(); got != expectedColorHex {
		t.Errorf("Color() = %q, want %q", got, expectedColorHex)
	}
	// Verify it's a valid hex color (orange).
	if len(m.Color()) != 7 || m.Color()[0] != '#' {
		t.Errorf("Color() has invalid format: %s", m.Color())
	}
	// Also verify constant matches.
	if servicetest.ColorHex != expectedColorHex {
		t.Errorf("ColorHex constant = %q, want %q", servicetest.ColorHex, expectedColorHex)
	}
}

func TestModuleStandard(t *testing.T) {
	m := servicetest.New()
	if got := m.Standard(); got != expectedStandardRef {
		t.Errorf("Standard() = %q, want %q", got, expectedStandardRef)
	}
	// Also verify constant matches.
	if servicetest.StandardRef != expectedStandardRef {
		t.Errorf("StandardRef constant = %q, want %q", servicetest.StandardRef, expectedStandardRef)
	}
}

func TestModuleTestTypes(t *testing.T) {
	m := servicetest.New()
	tests := m.TestTypes()

	if len(tests) != expectedTestTypeCount {
		t.Errorf("TestTypes() returned %d tests, want %d", len(tests), expectedTestTypeCount)
	}

	// Verify all expected test types are present.
	expectedTypes := []string{
		// Y.1564 EtherSAM.
		"y1564_config",
		"y1564_perf",
		"y1564",
		// MEF Service.
		"mef_config",
		"mef_perf",
		"mef",
	}

	for _, expected := range expectedTypes {
		if !slices.Contains(tests, expected) {
			t.Errorf("TestTypes() missing expected test type: %s", expected)
		}
	}
}

func TestModuleCanRun(t *testing.T) {
	m := servicetest.New()

	// Test valid Y.1564 test types.
	y1564Tests := []string{"y1564_config", "y1564_perf", "y1564"}
	for _, test := range y1564Tests {
		if !m.CanRun(test) {
			t.Errorf("CanRun(%q) = false, want true (Y.1564)", test)
		}
	}

	// Test valid MEF test types.
	mefTests := []string{"mef_config", "mef_perf", "mef"}
	for _, test := range mefTests {
		if !m.CanRun(test) {
			t.Errorf("CanRun(%q) = false, want true (MEF)", test)
		}
	}

	// Test invalid test types.
	invalidTests := []string{
		"rfc2544_throughput", // Benchmark module.
		"rfc2544_latency",    // Benchmark module.
		"y1731_delay",        // Measure module.
		"rfc2889_forwarding", // Certify module.
		"custom_stream",      // TrafficGen module.
		"reflect",            // Reflector module.
		"invalid",            // Nonexistent.
		"",                   // Empty string.
		"y1564_unknown",      // Invalid Y.1564 test.
		"mef_unknown",        // Invalid MEF test.
	}
	for _, test := range invalidTests {
		if m.CanRun(test) {
			t.Errorf("CanRun(%q) = true, want false", test)
		}
	}
}

func TestModuleTestDescription(t *testing.T) {
	m := servicetest.New()

	testCases := []struct {
		testType    string
		shouldExist bool
		contains    string
	}{
		{"y1564_config", true, "Configuration"},
		{"y1564_perf", true, "Performance"},
		{"y1564", true, "Y.1564"},
		{"mef_config", true, "MEF"},
		{"mef_perf", true, "Performance"},
		{"mef", true, "MEF"},
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

func TestModuleY1564PerfDuration(t *testing.T) {
	m := servicetest.New()
	desc := m.TestDescription("y1564_perf")

	// Y.1564 performance tests are typically 15+ minutes.
	if !containsSubstring(desc, "15") {
		t.Errorf("TestDescription(\"y1564_perf\") = %q, should mention 15 minute duration", desc)
	}
}

func TestModuleMultipleInstances(t *testing.T) {
	// Verify multiple instances are independent.
	m1 := servicetest.New()
	m2 := servicetest.New()

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

func TestModuleTestTypeCategories(t *testing.T) {
	m := servicetest.New()

	// Verify Y.1564 and MEF categories are present.
	tests := m.TestTypes()

	y1564Count := 0
	mefCount := 0

	for _, test := range tests {
		if len(test) >= 5 && test[:5] == "y1564" {
			y1564Count++
		}
		if len(test) >= 3 && test[:3] == "mef" {
			mefCount++
		}
	}

	expectedY1564 := 3
	expectedMEF := 3

	if y1564Count != expectedY1564 {
		t.Errorf("Expected %d Y.1564 tests, got %d", expectedY1564, y1564Count)
	}
	if mefCount != expectedMEF {
		t.Errorf("Expected %d MEF tests, got %d", expectedMEF, mefCount)
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
