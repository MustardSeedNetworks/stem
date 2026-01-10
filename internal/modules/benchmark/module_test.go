// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package benchmark_test

import (
	"slices"
	"testing"

	"github.com/krisarmstrong/stem/internal/modules/benchmark"
)

// Test constants.
const (
	expectedModuleName    = "benchmark"
	expectedDisplayName   = "Benchmark"
	expectedColorHex      = "#dc2626"
	expectedStandardRef   = "RFC 2544"
	expectedTestTypeCount = 6
)

func TestNew(t *testing.T) {
	m := benchmark.New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestModuleName(t *testing.T) {
	m := benchmark.New()
	if got := m.Name(); got != expectedModuleName {
		t.Errorf("Name() = %q, want %q", got, expectedModuleName)
	}
	// Also verify constant matches.
	if benchmark.ModuleName != expectedModuleName {
		t.Errorf("ModuleName constant = %q, want %q", benchmark.ModuleName, expectedModuleName)
	}
}

func TestModuleDisplayName(t *testing.T) {
	m := benchmark.New()
	if got := m.DisplayName(); got != expectedDisplayName {
		t.Errorf("DisplayName() = %q, want %q", got, expectedDisplayName)
	}
	// Also verify constant matches.
	if benchmark.DisplayName != expectedDisplayName {
		t.Errorf("DisplayName constant = %q, want %q", benchmark.DisplayName, expectedDisplayName)
	}
}

func TestModuleDescription(t *testing.T) {
	m := benchmark.New()
	desc := m.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	// Should mention RFC 2544.
	if !containsSubstring(desc, "RFC 2544") {
		t.Error("Description() should mention RFC 2544")
	}
	// Should mention benchmarking.
	if !containsSubstring(desc, "benchmark") {
		t.Error("Description() should mention benchmarking")
	}
}

func TestModuleColor(t *testing.T) {
	m := benchmark.New()
	if got := m.Color(); got != expectedColorHex {
		t.Errorf("Color() = %q, want %q", got, expectedColorHex)
	}
	// Verify it's a valid hex color (red).
	if len(m.Color()) != 7 || m.Color()[0] != '#' {
		t.Errorf("Color() has invalid format: %s", m.Color())
	}
	// Also verify constant matches.
	if benchmark.ColorHex != expectedColorHex {
		t.Errorf("ColorHex constant = %q, want %q", benchmark.ColorHex, expectedColorHex)
	}
}

func TestModuleStandard(t *testing.T) {
	m := benchmark.New()
	if got := m.Standard(); got != expectedStandardRef {
		t.Errorf("Standard() = %q, want %q", got, expectedStandardRef)
	}
	// Also verify constant matches.
	if benchmark.StandardRef != expectedStandardRef {
		t.Errorf("StandardRef constant = %q, want %q", benchmark.StandardRef, expectedStandardRef)
	}
}

func TestModuleTestTypes(t *testing.T) {
	m := benchmark.New()
	tests := m.TestTypes()

	if len(tests) != expectedTestTypeCount {
		t.Errorf("TestTypes() returned %d tests, want %d", len(tests), expectedTestTypeCount)
	}

	// Verify all RFC 2544 test types are present.
	expectedTypes := []string{
		"rfc2544_throughput",
		"rfc2544_latency",
		"rfc2544_frame_loss",
		"rfc2544_back_to_back",
		"rfc2544_system_recovery",
		"rfc2544_reset",
	}

	for _, expected := range expectedTypes {
		if !slices.Contains(tests, expected) {
			t.Errorf("TestTypes() missing expected test type: %s", expected)
		}
	}
}

func TestModuleCanRunValid(t *testing.T) {
	m := benchmark.New()

	// Test all valid RFC 2544 test types.
	validTests := []string{
		"rfc2544_throughput",
		"rfc2544_latency",
		"rfc2544_frame_loss",
		"rfc2544_back_to_back",
		"rfc2544_system_recovery",
		"rfc2544_reset",
	}
	for _, test := range validTests {
		if !m.CanRun(test) {
			t.Errorf("CanRun(%q) = false, want true", test)
		}
	}
}

func TestModuleCanRunInvalid(t *testing.T) {
	m := benchmark.New()

	// Test invalid test types - benchmark should NOT run other module tests.
	invalidTests := []string{
		"y1564",              // ServiceTest module.
		"y1564_config",       // ServiceTest module.
		"y1564_perf",         // ServiceTest module.
		"y1731_delay",        // Measure module.
		"y1731_loss",         // Measure module.
		"rfc2889_forwarding", // Certify module.
		"rfc6349_throughput", // Certify module.
		"tsn_timing",         // Certify module.
		"custom_stream",      // TrafficGen module.
		"reflect",            // Reflector module.
		"invalid",            // Nonexistent.
		"",                   // Empty string.
		"throughput",         // Old unprefixed name.
		"latency",            // Old unprefixed name.
		"frame_loss",         // Old unprefixed name.
	}
	for _, test := range invalidTests {
		if m.CanRun(test) {
			t.Errorf("CanRun(%q) = true, want false", test)
		}
	}
}

func TestModuleTestDescription(t *testing.T) {
	m := benchmark.New()

	testCases := []struct {
		testType    string
		shouldExist bool
		contains    string
	}{
		{"rfc2544_throughput", true, "RFC 2544 Section 26.1"},
		{"rfc2544_latency", true, "RFC 2544 Section 26.2"},
		{"rfc2544_frame_loss", true, "RFC 2544 Section 26.3"},
		{"rfc2544_back_to_back", true, "RFC 2544 Section 26.4"},
		{"rfc2544_system_recovery", true, "RFC 2544 Section 26.5"},
		{"rfc2544_reset", true, "RFC 2544 Section 26.6"},
		{"invalid", false, ""},
		{"", false, ""},
		{"throughput", false, ""}, // Old unprefixed name.
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

func TestModuleTestDescriptionContents(t *testing.T) {
	m := benchmark.New()

	// Verify specific descriptions have meaningful content.
	testCases := []struct {
		testType     string
		expectations []string
	}{
		{"rfc2544_throughput", []string{"throughput", "zero loss"}},
		{"rfc2544_latency", []string{"latency", "load"}},
		{"rfc2544_frame_loss", []string{"loss", "rate"}},
		{"rfc2544_back_to_back", []string{"burst", "capacity"}},
		{"rfc2544_system_recovery", []string{"Recovery", "overload"}},
		{"rfc2544_reset", []string{"reset", "recovery"}},
	}

	for _, tc := range testCases {
		t.Run(tc.testType, func(t *testing.T) {
			desc := m.TestDescription(tc.testType)
			for _, expected := range tc.expectations {
				if !containsSubstring(desc, expected) {
					t.Errorf("TestDescription(%q) = %q, should contain %q", tc.testType, desc, expected)
				}
			}
		})
	}
}

func TestModuleMultipleInstances(t *testing.T) {
	// Verify multiple instances are independent.
	m1 := benchmark.New()
	m2 := benchmark.New()

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
	if m1.DisplayName() != m2.DisplayName() {
		t.Error("Different instances should have same DisplayName()")
	}
	if m1.Description() != m2.Description() {
		t.Error("Different instances should have same Description()")
	}
}

func TestModuleIsRFC2544Based(t *testing.T) {
	m := benchmark.New()

	// Benchmark module should be RFC 2544 based.
	if m.Standard() != "RFC 2544" {
		t.Errorf("Standard() = %q, want 'RFC 2544'", m.Standard())
	}

	// Should have exactly 6 RFC 2544 test types.
	if len(m.TestTypes()) != 6 {
		t.Errorf("Benchmark should have exactly 6 test types, got %d", len(m.TestTypes()))
	}

	// All test types should start with rfc2544_.
	for _, tt := range m.TestTypes() {
		if len(tt) < 8 || tt[:8] != "rfc2544_" {
			t.Errorf("Test type %q should start with 'rfc2544_'", tt)
		}
	}
}

func TestModuleConstants(t *testing.T) {
	// Verify all exported constants are correct.
	if benchmark.ModuleName != "benchmark" {
		t.Errorf("ModuleName = %q, want 'benchmark'", benchmark.ModuleName)
	}
	if benchmark.DisplayName != "Benchmark" {
		t.Errorf("DisplayName = %q, want 'Benchmark'", benchmark.DisplayName)
	}
	if benchmark.ColorHex != "#dc2626" {
		t.Errorf("ColorHex = %q, want '#dc2626'", benchmark.ColorHex)
	}
	if benchmark.StandardRef != "RFC 2544" {
		t.Errorf("StandardRef = %q, want 'RFC 2544'", benchmark.StandardRef)
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
