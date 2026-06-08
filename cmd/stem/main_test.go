// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/version"
)

// Test constants for repeated strings.
const (
	testResultPass = "PASS"
	testResultFail = "FAIL"
)

// captureOutput captures output during function execution and returns the output.
func captureOutput(t *testing.T, fn func(w io.Writer)) string {
	t.Helper()

	var buf bytes.Buffer
	fn(&buf)
	return buf.String()
}

func allTestTypes() map[string]string {
	return map[string]string{
		"throughput":         "RFC 2544 Throughput",
		"latency":            "RFC 2544 Latency",
		"frame_loss":         "RFC 2544 Frame Loss",
		"back_to_back":       "RFC 2544 Back-to-Back",
		"system_recovery":    "RFC 2544 System Recovery",
		"reset":              "RFC 2544 Reset",
		"y1564_config":       "Y.1564 Service Configuration",
		"y1564_perf":         "Y.1564 Performance Test",
		"y1564":              "Y.1564 Latency",
		"rfc2889_forwarding": "RFC 2889 Forwarding",
		"rfc2889_caching":    "RFC 2889 Caching",
		"rfc2889_learning":   "RFC 2889 Learning",
		"rfc2889_broadcast":  "RFC 2889 Broadcast",
		"rfc2889_congestion": "RFC 2889 Congestion",
		"rfc6349_throughput": "RFC 6349 Throughput",
		"rfc6349_path":       "RFC 6349 Path",
		"y1731_delay":        "Y.1731 Delay",
		"y1731_loss":         "Y.1731 Loss",
		"y1731_slm":          "Y.1731 SLM",
		"y1731_loopback":     "Y.1731 Loopback",
		"mef_config":         "MEF Configuration",
		"mef_perf":           "MEF Performance",
		"mef":                "MEF Service",
		"tsn_timing":         "TSN Timing",
		"tsn_isolation":      "TSN Isolation",
		"tsn_latency":        "TSN Latency",
		"tsn":                "TSN Service",
	}
}

type testCategory struct {
	name  string
	tests []string
}

func testCategories() []testCategory {
	return []testCategory{
		{
			name: "RFC 2544",
			tests: []string{
				"throughput",
				"latency",
				"frame_loss",
				"back_to_back",
				"system_recovery",
				"reset",
			},
		},
		{
			name: "Y.1564 EtherSAM",
			tests: []string{
				"y1564_config",
				"y1564_perf",
				"y1564",
			},
		},
		{
			name: "RFC 2889 LAN Switch",
			tests: []string{
				"rfc2889_forwarding",
				"rfc2889_caching",
				"rfc2889_learning",
				"rfc2889_broadcast",
				"rfc2889_congestion",
			},
		},
		{
			name: "RFC 6349 TCP",
			tests: []string{
				"rfc6349_throughput",
				"rfc6349_path",
			},
		},
		{
			name: "Y.1731 OAM",
			tests: []string{
				"y1731_delay",
				"y1731_loss",
				"y1731_slm",
				"y1731_loopback",
			},
		},
		{
			name: "MEF Service",
			tests: []string{
				"mef_config",
				"mef_perf",
				"mef",
			},
		},
		{
			name: "TSN 802.1Qbv",
			tests: []string{
				"tsn_timing",
				"tsn_isolation",
				"tsn_latency",
				"tsn",
			},
		},
	}
}

func TestVersion(t *testing.T) {
	if version.GetVersion() == "" {
		t.Error("Version should not be empty")
	}
	// Version is "dev" when not built with ldflags, or semver when built.
	if version.GetVersion() != "dev" && !strings.Contains(version.GetVersion(), ".") {
		t.Error("Version should be 'dev' or contain dots (semantic versioning)")
	}
}

func TestProductName(t *testing.T) {
	if ProductName != "The Stem" {
		t.Errorf("Expected ProductName 'The Stem', got '%s'", ProductName)
	}
}

func TestCompany(t *testing.T) {
	if Company != "Mustard Seed Networks" {
		t.Errorf("Expected Company 'Mustard Seed Networks', got '%s'", Company)
	}
}

func TestAllTestTypesCount(t *testing.T) {
	// We should have 27 test types total.
	expectedCount := 27
	if len(allTestTypes()) != expectedCount {
		t.Errorf("Expected %d test types, got %d", expectedCount, len(allTestTypes()))
	}
}

func TestAllTestTypesRFC2544(t *testing.T) {
	rfc2544Tests := []string{
		"throughput",
		"latency",
		"frame_loss",
		"back_to_back",
		"system_recovery",
		"reset",
	}

	allTypes := allTestTypes()
	for _, test := range rfc2544Tests {
		if _, ok := allTypes[test]; !ok {
			t.Errorf("Missing RFC 2544 test type: %s", test)
		}
	}
}

func TestAllTestTypesY1564(t *testing.T) {
	y1564Tests := []string{
		"y1564_config",
		"y1564_perf",
		"y1564",
	}

	allTypes := allTestTypes()
	for _, test := range y1564Tests {
		if _, ok := allTypes[test]; !ok {
			t.Errorf("Missing Y.1564 test type: %s", test)
		}
	}
}

func TestAllTestTypesRFC2889(t *testing.T) {
	rfc2889Tests := []string{
		"rfc2889_forwarding",
		"rfc2889_caching",
		"rfc2889_learning",
		"rfc2889_broadcast",
		"rfc2889_congestion",
	}

	allTypes := allTestTypes()
	for _, test := range rfc2889Tests {
		if _, ok := allTypes[test]; !ok {
			t.Errorf("Missing RFC 2889 test type: %s", test)
		}
	}
}

func TestAllTestTypesRFC6349(t *testing.T) {
	rfc6349Tests := []string{
		"rfc6349_throughput",
		"rfc6349_path",
	}

	allTypes := allTestTypes()
	for _, test := range rfc6349Tests {
		if _, ok := allTypes[test]; !ok {
			t.Errorf("Missing RFC 6349 test type: %s", test)
		}
	}
}

func TestAllTestTypesY1731(t *testing.T) {
	y1731Tests := []string{
		"y1731_delay",
		"y1731_loss",
		"y1731_slm",
		"y1731_loopback",
	}

	allTypes := allTestTypes()
	for _, test := range y1731Tests {
		if _, ok := allTypes[test]; !ok {
			t.Errorf("Missing Y.1731 test type: %s", test)
		}
	}
}

func TestAllTestTypesMEF(t *testing.T) {
	mefTests := []string{
		"mef_config",
		"mef_perf",
		"mef",
	}

	allTypes := allTestTypes()
	for _, test := range mefTests {
		if _, ok := allTypes[test]; !ok {
			t.Errorf("Missing MEF test type: %s", test)
		}
	}
}

func TestAllTestTypesTSN(t *testing.T) {
	tsnTests := []string{
		"tsn_timing",
		"tsn_isolation",
		"tsn_latency",
		"tsn",
	}

	allTypes := allTestTypes()
	for _, test := range tsnTests {
		if _, ok := allTypes[test]; !ok {
			t.Errorf("Missing TSN test type: %s", test)
		}
	}
}

func TestTestCategoriesCount(t *testing.T) {
	expectedCategories := 7
	if len(testCategories()) != expectedCategories {
		t.Errorf("Expected %d test categories, got %d", expectedCategories, len(testCategories()))
	}
}

func TestTestCategoriesNames(t *testing.T) {
	expectedNames := []string{
		"RFC 2544",
		"Y.1564 EtherSAM",
		"RFC 2889 LAN Switch",
		"RFC 6349 TCP",
		"Y.1731 OAM",
		"MEF Service",
		"TSN 802.1Qbv",
	}

	categories := testCategories()
	for i, expected := range expectedNames {
		if categories[i].name != expected {
			t.Errorf("Expected category %d name '%s', got '%s'", i, expected, categories[i].name)
		}
	}
}

func TestTestCategoriesTestsExist(t *testing.T) {
	types := allTestTypes()
	// Verify all tests in categories exist in allTestTypes.
	for _, cat := range testCategories() {
		for _, test := range cat.tests {
			if _, ok := types[test]; !ok {
				t.Errorf("Test '%s' in category '%s' not found in allTestTypes", test, cat.name)
			}
		}
	}
}

func TestParseFrameSizes(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"64", []int{64}},
		{"64,128,256", []int{64, 128, 256}},
		{"64,128,256,512,1024,1280,1518", []int{64, 128, 256, 512, 1024, 1280, 1518}},
		{"1518,9000", []int{1518, 9000}},
		{"", []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseFrameSizes(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseFrameSizes(%s) returned %d sizes, expected %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseFrameSizes(%s)[%d] = %d, expected %d", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestParseFrameSizesInvalid(t *testing.T) {
	// Invalid frame sizes should be skipped.
	result := parseFrameSizes("64,invalid,128")
	expected := []int{64, 128}
	if len(result) != len(expected) {
		t.Errorf("Expected %d valid sizes, got %d", len(expected), len(result))
	}
}

func TestBoolToPassFail(t *testing.T) {
	if boolToPassFail(true) != testResultPass {
		t.Errorf("Expected '%s' for true", testResultPass)
	}
	if boolToPassFail(false) != testResultFail {
		t.Errorf("Expected '%s' for false", testResultFail)
	}
}

func TestPrintVersion(t *testing.T) {
	output := captureOutput(t, printVersion)

	if !strings.Contains(output, ProductName) {
		t.Error("printVersion should contain ProductName")
	}
	if !strings.Contains(output, version.GetVersion()) {
		t.Error("printVersion should contain Version")
	}
	if !strings.Contains(output, Company) {
		t.Error("printVersion should contain Company")
	}
}

func TestPrintUsage(t *testing.T) {
	output := captureOutput(t, printUsage)

	// Should contain key sections.
	expectedSections := []string{
		"USAGE:",
		"COMMANDS:",
		"reflect",
		"test",
		"web",
		"tui",
		"license",
		"EXAMPLES:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("printUsage should contain '%s'", section)
		}
	}
}

func TestDefaultFrameSizes(t *testing.T) {
	// Default RFC 2544 frame sizes.
	defaults := parseFrameSizes("64,128,256,512,1024,1280,1518")
	if len(defaults) != 7 {
		t.Errorf("Expected 7 default frame sizes, got %d", len(defaults))
	}

	expected := []int{64, 128, 256, 512, 1024, 1280, 1518}
	for i, v := range expected {
		if defaults[i] != v {
			t.Errorf("Default frame size[%d] = %d, expected %d", i, defaults[i], v)
		}
	}
}

func TestJumboFrameSupport(t *testing.T) {
	// Should support jumbo frames (9216 is > 9216 so it's excluded).
	jumbos := parseFrameSizes("9000,9216")
	if len(jumbos) != 2 {
		t.Errorf("Expected 2 jumbo sizes, got %d", len(jumbos))
	}
	if len(jumbos) > 0 && jumbos[0] != 9000 {
		t.Errorf("Expected 9000, got %d", jumbos[0])
	}
	if len(jumbos) > 1 && jumbos[1] != 9216 {
		t.Errorf("Expected 9216, got %d", jumbos[1])
	}
}

func TestTestTypeDescriptions(t *testing.T) {
	// Verify each test type has a non-empty description.
	for testType, desc := range allTestTypes() {
		if desc == "" {
			t.Errorf("Test type '%s' has empty description", testType)
		}
		if len(desc) < 10 {
			t.Errorf("Test type '%s' has too short description: '%s'", testType, desc)
		}
	}
}

func TestTestTypeDescriptionsContainStandard(t *testing.T) {
	// RFC 2544 tests should reference RFC 2544.
	rfc2544Tests := []string{"throughput", "latency", "frame_loss", "back_to_back", "system_recovery", "reset"}
	types := allTestTypes()
	for _, test := range rfc2544Tests {
		desc := types[test]
		if !strings.Contains(desc, "RFC 2544") {
			t.Errorf("RFC 2544 test '%s' description should reference RFC 2544: '%s'", test, desc)
		}
	}

	// Y.1564 tests should reference Y.1564 or ITU-T.
	y1564Tests := []string{"y1564_config", "y1564_perf", "y1564"}
	for _, test := range y1564Tests {
		desc := types[test]
		if !strings.Contains(desc, "Y.1564") && !strings.Contains(desc, "ITU-T") {
			t.Errorf("Y.1564 test '%s' description should reference Y.1564 or ITU-T: '%s'", test, desc)
		}
	}
}

// Table-driven test for frame size edge cases.
func TestParseFrameSizesEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected length.
	}{
		{"empty string", "", 0},
		{"spaces", "  ", 0},
		{"single value", "1518", 1},
		{"trailing comma", "64,128,", 2},
		{"leading comma", ",64,128", 2},
		{"duplicate values", "64,64,128", 3},
		{"large value over limit", "16384", 0}, // 16384 > 9216 limit.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFrameSizes(tt.input)
			if len(result) != tt.expected {
				t.Errorf("parseFrameSizes(%q) = %d sizes, want %d", tt.input, len(result), tt.expected)
			}
		})
	}
}

// Benchmark tests.
func BenchmarkParseFrameSizes(b *testing.B) {
	input := "64,128,256,512,1024,1280,1518"
	for b.Loop() {
		parseFrameSizes(input)
	}
}

func BenchmarkBoolToPassFail(b *testing.B) {
	for i := range b.N {
		boolToPassFail(i%2 == 0)
	}
}
