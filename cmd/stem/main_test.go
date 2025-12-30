// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/krisarmstrong/stem/internal/version"
)

func TestVersion(t *testing.T) {
	if version.Version == "" {
		t.Error("Version should not be empty")
	}
	// Version is "dev" when not built with ldflags, or semver when built
	if version.Version != "dev" && !strings.Contains(version.Version, ".") {
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
	// We should have 27 test types total
	expectedCount := 27
	if len(allTestTypes) != expectedCount {
		t.Errorf("Expected %d test types, got %d", expectedCount, len(allTestTypes))
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

	for _, test := range rfc2544Tests {
		if _, ok := allTestTypes[test]; !ok {
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

	for _, test := range y1564Tests {
		if _, ok := allTestTypes[test]; !ok {
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

	for _, test := range rfc2889Tests {
		if _, ok := allTestTypes[test]; !ok {
			t.Errorf("Missing RFC 2889 test type: %s", test)
		}
	}
}

func TestAllTestTypesRFC6349(t *testing.T) {
	rfc6349Tests := []string{
		"rfc6349_throughput",
		"rfc6349_path",
	}

	for _, test := range rfc6349Tests {
		if _, ok := allTestTypes[test]; !ok {
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

	for _, test := range y1731Tests {
		if _, ok := allTestTypes[test]; !ok {
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

	for _, test := range mefTests {
		if _, ok := allTestTypes[test]; !ok {
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

	for _, test := range tsnTests {
		if _, ok := allTestTypes[test]; !ok {
			t.Errorf("Missing TSN test type: %s", test)
		}
	}
}

func TestTestCategoriesCount(t *testing.T) {
	expectedCategories := 7
	if len(testCategories) != expectedCategories {
		t.Errorf("Expected %d test categories, got %d", expectedCategories, len(testCategories))
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

	for i, expected := range expectedNames {
		if testCategories[i].name != expected {
			t.Errorf("Expected category %d name '%s', got '%s'", i, expected, testCategories[i].name)
		}
	}
}

func TestTestCategoriesTestsExist(t *testing.T) {
	// Verify all tests in categories exist in allTestTypes
	for _, cat := range testCategories {
		for _, test := range cat.tests {
			if _, ok := allTestTypes[test]; !ok {
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
	// Invalid frame sizes should be skipped
	result := parseFrameSizes("64,invalid,128")
	expected := []int{64, 128}
	if len(result) != len(expected) {
		t.Errorf("Expected %d valid sizes, got %d", len(expected), len(result))
	}
}

func TestBoolToPassFail(t *testing.T) {
	if boolToPassFail(true) != "PASS" {
		t.Error("Expected 'PASS' for true")
	}
	if boolToPassFail(false) != "FAIL" {
		t.Error("Expected 'FAIL' for false")
	}
}

func TestPrintVersion(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printVersion()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, ProductName) {
		t.Error("printVersion should contain ProductName")
	}
	if !strings.Contains(output, version.Version) {
		t.Error("printVersion should contain Version")
	}
	if !strings.Contains(output, Company) {
		t.Error("printVersion should contain Company")
	}
}

func TestPrintUsage(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printUsage()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should contain key sections
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

func TestListTestsCmd(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listTestsCmd([]string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should list test categories
	for _, cat := range testCategories {
		if !strings.Contains(output, cat.name) {
			t.Errorf("listTestsCmd should list category '%s'", cat.name)
		}
	}

	// Should list key test types
	keyTests := []string{"throughput", "latency", "y1564"}
	for _, test := range keyTests {
		if !strings.Contains(output, test) {
			t.Errorf("listTestsCmd should list test '%s'", test)
		}
	}
}

func TestListTestsCmdByModule(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listTestsCmd([]string{"--by-module"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should list module names (including Reflector)
	modules := []string{"Reflector", "Benchmark", "ServiceTest", "TrafficGen", "Measure", "Certify"}
	for _, mod := range modules {
		if !strings.Contains(output, mod) {
			t.Errorf("listTestsCmd --by-module should list module '%s'", mod)
		}
	}

	// Should list module colors (including Reflector cyan)
	colors := []string{"#0891b2", "#dc2626", "#ea580c", "#ca8a04", "#2563eb", "#16a34a"}
	for _, color := range colors {
		if !strings.Contains(output, color) {
			t.Errorf("listTestsCmd --by-module should show color '%s'", color)
		}
	}

	// Should show "across 6 modules" not "categories"
	if !strings.Contains(output, "6 modules") {
		t.Error("listTestsCmd --by-module should show 'across X modules'")
	}
}

func TestDefaultFrameSizes(t *testing.T) {
	// Default RFC 2544 frame sizes
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
	// Should support jumbo frames (9216 is > 9216 so it's excluded)
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
	// Verify each test type has a non-empty description
	for testType, desc := range allTestTypes {
		if desc == "" {
			t.Errorf("Test type '%s' has empty description", testType)
		}
		if len(desc) < 10 {
			t.Errorf("Test type '%s' has too short description: '%s'", testType, desc)
		}
	}
}

func TestTestTypeDescriptionsContainStandard(t *testing.T) {
	// RFC 2544 tests should reference RFC 2544
	rfc2544Tests := []string{"throughput", "latency", "frame_loss", "back_to_back", "system_recovery", "reset"}
	for _, test := range rfc2544Tests {
		desc := allTestTypes[test]
		if !strings.Contains(desc, "RFC 2544") {
			t.Errorf("RFC 2544 test '%s' description should reference RFC 2544: '%s'", test, desc)
		}
	}

	// Y.1564 tests should reference Y.1564 or ITU-T
	y1564Tests := []string{"y1564_config", "y1564_perf", "y1564"}
	for _, test := range y1564Tests {
		desc := allTestTypes[test]
		if !strings.Contains(desc, "Y.1564") && !strings.Contains(desc, "ITU-T") {
			t.Errorf("Y.1564 test '%s' description should reference Y.1564 or ITU-T: '%s'", test, desc)
		}
	}
}

// Table-driven test for frame size edge cases
func TestParseFrameSizesEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected length
	}{
		{"empty string", "", 0},
		{"spaces", "  ", 0},
		{"single value", "1518", 1},
		{"trailing comma", "64,128,", 2},
		{"leading comma", ",64,128", 2},
		{"duplicate values", "64,64,128", 3},
		{"large value over limit", "16384", 0}, // 16384 > 9216 limit
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

// Benchmark tests
func BenchmarkParseFrameSizes(b *testing.B) {
	input := "64,128,256,512,1024,1280,1518"
	for i := 0; i < b.N; i++ {
		parseFrameSizes(input)
	}
}

func BenchmarkBoolToPassFail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		boolToPassFail(i%2 == 0)
	}
}
