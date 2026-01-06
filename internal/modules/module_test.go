// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package modules_test

import (
	"testing"

	"github.com/krisarmstrong/stem/internal/modules"
	"github.com/krisarmstrong/stem/internal/modules/benchmark"
	"github.com/krisarmstrong/stem/internal/modules/certify"
	"github.com/krisarmstrong/stem/internal/modules/measure"
	"github.com/krisarmstrong/stem/internal/modules/reflector"
	"github.com/krisarmstrong/stem/internal/modules/servicetest"
	"github.com/krisarmstrong/stem/internal/modules/trafficgen"
)

// Module name constants for tests.
const (
	testModuleBenchmark   = "benchmark"
	testModuleCertify     = "certify"
	testModuleMeasure     = "measure"
	testModuleReflector   = "reflector"
	testModuleServiceTest = "servicetest"
	testModuleTrafficGen  = "trafficgen"
	expectedModuleCount   = 6
	expectedTestCount     = 29
	expectedColorLength   = 7
)

func TestRegistry(t *testing.T) {
	reg := modules.NewRegistry()

	// Register a module.
	bm := benchmark.New()
	reg.Register(bm)

	// Test Get.
	if got := reg.Get(testModuleBenchmark); got == nil {
		t.Error("Get('benchmark') returned nil")
	}
	if got := reg.Get("nonexistent"); got != nil {
		t.Error("Get('nonexistent') should return nil")
	}

	// Test ModuleForTest.
	if got := reg.ModuleForTest("throughput"); got == nil {
		t.Error("ModuleForTest('throughput') returned nil")
	}
	if got := reg.ModuleForTest("nonexistent"); got != nil {
		t.Error("ModuleForTest('nonexistent') should return nil")
	}

	// Test counts.
	if reg.ModuleCount() != 1 {
		t.Errorf("ModuleCount() = %d, want 1", reg.ModuleCount())
	}
	expectedTests := 6
	if reg.TestCount() != expectedTests {
		t.Errorf("TestCount() = %d, want %d (RFC 2544 tests)", reg.TestCount(), expectedTests)
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Test that default registry has all 6 modules.
	mods := modules.DefaultRegistry.AllModules()
	if len(mods) != expectedModuleCount {
		t.Errorf("DefaultRegistry has %d modules, want %d", len(mods), expectedModuleCount)
	}

	// Test module lookup.
	names := []string{
		testModuleReflector, testModuleBenchmark, testModuleServiceTest,
		testModuleTrafficGen, testModuleMeasure, testModuleCertify,
	}
	for _, name := range names {
		if m := modules.DefaultRegistry.Get(name); m == nil {
			t.Errorf("DefaultRegistry.Get(%q) returned nil", name)
		}
	}
}

func TestBenchmarkModule(t *testing.T) {
	m := benchmark.New()

	if m.Name() != testModuleBenchmark {
		t.Errorf("Name() = %q, want 'benchmark'", m.Name())
	}
	if m.DisplayName() != "Benchmark" {
		t.Errorf("DisplayName() = %q, want 'Benchmark'", m.DisplayName())
	}
	if m.Color() != "#dc2626" {
		t.Errorf("Color() = %q, want '#dc2626'", m.Color())
	}
	if m.Standard() != "RFC 2544" {
		t.Errorf("Standard() = %q, want 'RFC 2544'", m.Standard())
	}

	// Test test types.
	tests := m.TestTypes()
	expectedTests := 6
	if len(tests) != expectedTests {
		t.Errorf("TestTypes() returned %d tests, want %d", len(tests), expectedTests)
	}

	// Test CanRun.
	if !m.CanRun("throughput") {
		t.Error("CanRun('throughput') should return true")
	}
	if m.CanRun("y1564_config") {
		t.Error("CanRun('y1564_config') should return false")
	}
}

func TestServiceTestModule(t *testing.T) {
	m := servicetest.New()

	if m.Name() != testModuleServiceTest {
		t.Errorf("Name() = %q, want 'servicetest'", m.Name())
	}
	if m.Color() != "#ea580c" {
		t.Errorf("Color() = %q, want '#ea580c'", m.Color())
	}

	// Test Y.1564 tests.
	if !m.CanRun("y1564_config") {
		t.Error("CanRun('y1564_config') should return true")
	}
	if !m.CanRun("mef") {
		t.Error("CanRun('mef') should return true")
	}
}

func TestTrafficGenModule(t *testing.T) {
	m := trafficgen.New()

	if m.Name() != testModuleTrafficGen {
		t.Errorf("Name() = %q, want 'trafficgen'", m.Name())
	}
	if m.Color() != "#ca8a04" {
		t.Errorf("Color() = %q, want '#ca8a04'", m.Color())
	}
	if !m.CanRun("custom_stream") {
		t.Error("CanRun('custom_stream') should return true")
	}
	if m.CanRun("reflect") {
		t.Error("CanRun('reflect') should return false (now in reflector module)")
	}
}

func TestReflectorModule(t *testing.T) {
	m := reflector.New()

	if m.Name() != testModuleReflector {
		t.Errorf("Name() = %q, want 'reflector'", m.Name())
	}
	if m.DisplayName() != "Reflector" {
		t.Errorf("DisplayName() = %q, want 'Reflector'", m.DisplayName())
	}
	if m.Color() != "#0891b2" {
		t.Errorf("Color() = %q, want '#0891b2'", m.Color())
	}
	if m.Standard() != "Loopback/Echo" {
		t.Errorf("Standard() = %q, want 'Loopback/Echo'", m.Standard())
	}
	if !m.CanRun("reflect") {
		t.Error("CanRun('reflect') should return true")
	}
}

func TestMeasureModule(t *testing.T) {
	m := measure.New()

	if m.Name() != testModuleMeasure {
		t.Errorf("Name() = %q, want 'measure'", m.Name())
	}
	if m.Color() != "#2563eb" {
		t.Errorf("Color() = %q, want '#2563eb'", m.Color())
	}
	if !m.CanRun("y1731_delay") {
		t.Error("CanRun('y1731_delay') should return true")
	}
}

func TestCertifyModule(t *testing.T) {
	m := certify.New()

	if m.Name() != testModuleCertify {
		t.Errorf("Name() = %q, want 'certify'", m.Name())
	}
	if m.Color() != "#16a34a" {
		t.Errorf("Color() = %q, want '#16a34a'", m.Color())
	}

	// Test various standards.
	if !m.CanRun("rfc2889_forwarding") {
		t.Error("CanRun('rfc2889_forwarding') should return true")
	}
	if !m.CanRun("rfc6349_throughput") {
		t.Error("CanRun('rfc6349_throughput') should return true")
	}
	if !m.CanRun("tsn_timing") {
		t.Error("CanRun('tsn_timing') should return true")
	}
}

func TestToInfo(t *testing.T) {
	m := benchmark.New()
	info := modules.ToInfo(m)

	if info.Name != testModuleBenchmark {
		t.Errorf("info.Name = %q, want 'benchmark'", info.Name)
	}
	if info.DisplayName != "Benchmark" {
		t.Errorf("info.DisplayName = %q, want 'Benchmark'", info.DisplayName)
	}
	if info.Color != "#dc2626" {
		t.Errorf("info.Color = %q, want '#dc2626'", info.Color)
	}
	expectedTests := 6
	if len(info.Tests) != expectedTests {
		t.Errorf("len(info.Tests) = %d, want %d", len(info.Tests), expectedTests)
	}
}

func TestGetModuleForTest(t *testing.T) {
	// RFC 2544 tests -> benchmark.
	if m := modules.GetModuleForTest("throughput"); m == nil || m.Name() != testModuleBenchmark {
		t.Error("throughput should map to benchmark module")
	}

	// Y.1564 tests -> servicetest.
	if m := modules.GetModuleForTest("y1564_config"); m == nil || m.Name() != testModuleServiceTest {
		t.Error("y1564_config should map to servicetest module")
	}

	// Y.1731 tests -> measure.
	if m := modules.GetModuleForTest("y1731_delay"); m == nil || m.Name() != testModuleMeasure {
		t.Error("y1731_delay should map to measure module")
	}

	// RFC 2889/6349/TSN tests -> certify.
	if m := modules.GetModuleForTest("rfc2889_forwarding"); m == nil || m.Name() != testModuleCertify {
		t.Error("rfc2889_forwarding should map to certify module")
	}
	if m := modules.GetModuleForTest("rfc6349_throughput"); m == nil || m.Name() != testModuleCertify {
		t.Error("rfc6349_throughput should map to certify module")
	}
}

func TestAllModuleInfos(t *testing.T) {
	infos := modules.GetAllModuleInfos()
	if len(infos) != expectedModuleCount {
		t.Errorf("GetAllModuleInfos() returned %d infos, want %d", len(infos), expectedModuleCount)
	}
}

func TestGetModuleForReflect(t *testing.T) {
	// reflect -> reflector module (not trafficgen).
	if m := modules.GetModuleForTest("reflect"); m == nil || m.Name() != testModuleReflector {
		t.Error("reflect should map to reflector module")
	}

	// custom_stream -> trafficgen module.
	if m := modules.GetModuleForTest("custom_stream"); m == nil || m.Name() != testModuleTrafficGen {
		t.Error("custom_stream should map to trafficgen module")
	}
}

func TestRegistryEdgeCases(t *testing.T) {
	reg := modules.NewRegistry()

	// Registering nil should not panic (defensive).
	// Attempting to get from empty registry.
	if got := reg.Get("anything"); got != nil {
		t.Error("Get on empty registry should return nil")
	}
	if got := reg.ModuleForTest("anything"); got != nil {
		t.Error("ModuleForTest on empty registry should return nil")
	}

	// Register same module twice should overwrite.
	bm1 := benchmark.New()
	bm2 := benchmark.New()
	reg.Register(bm1)
	reg.Register(bm2)
	if reg.ModuleCount() != 1 {
		t.Errorf("Duplicate registration should overwrite, got count %d", reg.ModuleCount())
	}
}

func TestAllModulesOrdering(t *testing.T) {
	mods := modules.DefaultRegistry.AllModules()
	if len(mods) != expectedModuleCount {
		t.Fatalf("Expected %d modules, got %d", expectedModuleCount, len(mods))
	}

	// Verify we have all expected modules.
	found := make(map[string]bool)
	for _, m := range mods {
		found[m.Name()] = true
	}

	expected := []string{
		testModuleReflector, testModuleBenchmark, testModuleServiceTest,
		testModuleTrafficGen, testModuleMeasure, testModuleCertify,
	}
	for _, name := range expected {
		if !found[name] {
			t.Errorf("Missing module: %s", name)
		}
	}
}

func TestModuleColors(t *testing.T) {
	// Verify each module has a unique, valid hex color.
	mods := modules.DefaultRegistry.AllModules()
	colors := make(map[string]string)

	for _, m := range mods {
		color := m.Color()
		if len(color) != expectedColorLength || color[0] != '#' {
			t.Errorf("Module %s has invalid color format: %s", m.Name(), color)
		}
		if existing, ok := colors[color]; ok {
			t.Errorf("Duplicate color %s used by %s and %s", color, existing, m.Name())
		}
		colors[color] = m.Name()
	}

	// Verify expected colors.
	expectedColors := map[string]string{
		testModuleReflector:   "#0891b2",
		testModuleBenchmark:   "#dc2626",
		testModuleServiceTest: "#ea580c",
		testModuleTrafficGen:  "#ca8a04",
		testModuleMeasure:     "#2563eb",
		testModuleCertify:     "#16a34a",
	}

	for name, expectedColor := range expectedColors {
		m := modules.DefaultRegistry.Get(name)
		if m == nil {
			t.Errorf("Module %s not found", name)
			continue
		}
		if m.Color() != expectedColor {
			t.Errorf("Module %s has color %s, expected %s", name, m.Color(), expectedColor)
		}
	}
}

func TestModuleCanRunNegativeCases(t *testing.T) {
	testCases := []struct {
		moduleName string
		testType   string
	}{
		{testModuleBenchmark, "y1564_config"},
		{testModuleBenchmark, "reflect"},
		{testModuleServiceTest, "throughput"},
		{testModuleServiceTest, "rfc2889_forwarding"},
		{testModuleTrafficGen, "throughput"},
		{testModuleTrafficGen, "reflect"}, // reflect moved to reflector module
		{testModuleMeasure, "throughput"},
		{testModuleCertify, "throughput"},
		{testModuleReflector, "throughput"},
		{testModuleReflector, "custom_stream"},
	}

	for _, tc := range testCases {
		m := modules.DefaultRegistry.Get(tc.moduleName)
		if m == nil {
			t.Errorf("Module %s not found", tc.moduleName)
			continue
		}
		if m.CanRun(tc.testType) {
			t.Errorf("Module %s.CanRun(%s) should return false", tc.moduleName, tc.testType)
		}
	}
}

func TestAllTestTypesHaveModules(t *testing.T) {
	// All known test types should map to a module.
	testTypes := []string{
		// Reflector
		"reflect",
		// Benchmark
		"throughput", "latency", "frame_loss", "back_to_back", "system_recovery", "reset",
		// ServiceTest
		"y1564_config", "y1564_perf", "y1564", "mef_config", "mef_perf", "mef",
		// TrafficGen
		"custom_stream",
		// Measure
		"y1731_delay", "y1731_loss", "y1731_slm", "y1731_loopback",
		// Certify
		"rfc2889_forwarding", "rfc2889_caching", "rfc2889_learning",
		"rfc2889_broadcast", "rfc2889_congestion",
		"rfc6349_throughput", "rfc6349_path",
		"tsn_timing", "tsn_isolation", "tsn_latency", "tsn",
	}

	for _, tt := range testTypes {
		m := modules.GetModuleForTest(tt)
		if m == nil {
			t.Errorf("Test type %q has no module mapping", tt)
		}
	}
}

func TestToInfoComplete(t *testing.T) {
	// Test ToInfo for all modules.
	mods := modules.DefaultRegistry.AllModules()
	for _, m := range mods {
		info := modules.ToInfo(m)

		if info.Name != m.Name() {
			t.Errorf("ToInfo(%s).Name mismatch", m.Name())
		}
		if info.DisplayName != m.DisplayName() {
			t.Errorf("ToInfo(%s).DisplayName mismatch", m.Name())
		}
		if info.Description != m.Description() {
			t.Errorf("ToInfo(%s).Description mismatch", m.Name())
		}
		if info.Color != m.Color() {
			t.Errorf("ToInfo(%s).Color mismatch", m.Name())
		}
		if info.Standard != m.Standard() {
			t.Errorf("ToInfo(%s).Standard mismatch", m.Name())
		}
		if len(info.Tests) != len(m.TestTypes()) {
			t.Errorf("ToInfo(%s).Tests length mismatch: got %d, want %d",
				m.Name(), len(info.Tests), len(m.TestTypes()))
		}
	}
}

func TestTotalTestCount(t *testing.T) {
	// Count all tests across all modules.
	total := modules.DefaultRegistry.TestCount()

	// Expected: 1 (reflector) + 6 (benchmark) + 6 (servicetest) + 1 (trafficgen) + 4 (measure) + 11 (certify) = 29.
	if total != expectedTestCount {
		t.Errorf("Total test count = %d, want %d", total, expectedTestCount)
	}
}

// TestGetModule tests the GetModule convenience function.
func TestGetModule(t *testing.T) {
	// Test existing modules.
	moduleNames := []string{
		testModuleReflector, testModuleBenchmark, testModuleServiceTest,
		testModuleTrafficGen, testModuleMeasure, testModuleCertify,
	}

	for _, name := range moduleNames {
		m := modules.GetModule(name)
		if m == nil {
			t.Errorf("GetModule(%q) returned nil", name)
			continue
		}
		if m.Name() != name {
			t.Errorf("GetModule(%q).Name() = %q, want %q", name, m.Name(), name)
		}
	}

	// Test non-existent module.
	if m := modules.GetModule("nonexistent"); m != nil {
		t.Error("GetModule('nonexistent') should return nil")
	}

	// Test empty string.
	if m := modules.GetModule(""); m != nil {
		t.Error("GetModule('') should return nil")
	}
}

// TestGetAllModules tests the GetAllModules convenience function.
func TestGetAllModules(t *testing.T) {
	mods := modules.GetAllModules()

	if len(mods) != expectedModuleCount {
		t.Errorf("GetAllModules() returned %d modules, want %d", len(mods), expectedModuleCount)
	}

	// Verify all modules are present.
	found := make(map[string]bool)
	for _, m := range mods {
		found[m.Name()] = true
	}

	expectedNames := []string{
		testModuleReflector, testModuleBenchmark, testModuleServiceTest,
		testModuleTrafficGen, testModuleMeasure, testModuleCertify,
	}
	for _, name := range expectedNames {
		if !found[name] {
			t.Errorf("GetAllModules() missing module: %s", name)
		}
	}
}

// TestAllTestTypes tests the AllTestTypes method.
func TestAllTestTypes(t *testing.T) {
	types := modules.DefaultRegistry.AllTestTypes()

	if len(types) != expectedTestCount {
		t.Errorf("AllTestTypes() returned %d types, want %d", len(types), expectedTestCount)
	}

	// Verify each test type has required fields.
	for _, tt := range types {
		if tt.Name == "" {
			t.Error("AllTestTypes() returned a TestType with empty Name")
		}
		if tt.ModuleName == "" {
			t.Error("AllTestTypes() returned a TestType with empty ModuleName")
		}
		// Standard should be set from the module.
		if tt.Standard == "" {
			t.Errorf("AllTestTypes() returned TestType %q with empty Standard", tt.Name)
		}
	}

	// Verify specific test types exist.
	typeNames := make(map[string]bool)
	for _, tt := range types {
		typeNames[tt.Name] = true
	}

	expectedTypes := []string{
		"reflect", "throughput", "latency", "y1564_config", "custom_stream",
		"y1731_delay", "rfc2889_forwarding", "tsn_timing",
	}
	for _, name := range expectedTypes {
		if !typeNames[name] {
			t.Errorf("AllTestTypes() missing test type: %s", name)
		}
	}
}

// TestAllTestTypesModuleMapping tests that test types map to correct modules.
func TestAllTestTypesModuleMapping(t *testing.T) {
	types := modules.DefaultRegistry.AllTestTypes()

	// Create a map for easier lookup.
	typeMap := make(map[string]modules.TestType)
	for _, tt := range types {
		typeMap[tt.Name] = tt
	}

	// Verify module mappings.
	expectedMappings := map[string]string{
		"reflect":            testModuleReflector,
		"throughput":         testModuleBenchmark,
		"latency":            testModuleBenchmark,
		"y1564_config":       testModuleServiceTest,
		"custom_stream":      testModuleTrafficGen,
		"y1731_delay":        testModuleMeasure,
		"rfc2889_forwarding": testModuleCertify,
		"rfc6349_throughput": testModuleCertify,
		"tsn_timing":         testModuleCertify,
	}

	for testType, expectedModule := range expectedMappings {
		tt, ok := typeMap[testType]
		if !ok {
			t.Errorf("Test type %q not found in AllTestTypes()", testType)
			continue
		}
		if tt.ModuleName != expectedModule {
			t.Errorf("Test type %q has ModuleName %q, want %q", testType, tt.ModuleName, expectedModule)
		}
	}
}

// TestTestTypeStruct tests TestType struct fields.
func TestTestTypeStruct(t *testing.T) {
	tt := modules.TestType{
		Name:        "test_name",
		Description: "test description",
		Standard:    "RFC 2544",
		ModuleName:  "benchmark",
	}

	if tt.Name != "test_name" {
		t.Errorf("TestType.Name = %q, want 'test_name'", tt.Name)
	}
	if tt.Description != "test description" {
		t.Errorf("TestType.Description = %q, want 'test description'", tt.Description)
	}
	if tt.Standard != "RFC 2544" {
		t.Errorf("TestType.Standard = %q, want 'RFC 2544'", tt.Standard)
	}
	if tt.ModuleName != "benchmark" {
		t.Errorf("TestType.ModuleName = %q, want 'benchmark'", tt.ModuleName)
	}
}

// TestModuleInfoStruct tests ModuleInfo struct fields.
func TestModuleInfoStruct(t *testing.T) {
	info := modules.ModuleInfo{
		Name:        "test_module",
		DisplayName: "Test Module",
		Description: "A test module",
		Color:       "#ff0000",
		Standard:    "Test Standard",
		Tests:       []string{"test1", "test2"},
	}

	if info.Name != "test_module" {
		t.Errorf("ModuleInfo.Name = %q, want 'test_module'", info.Name)
	}
	if info.DisplayName != "Test Module" {
		t.Errorf("ModuleInfo.DisplayName = %q, want 'Test Module'", info.DisplayName)
	}
	if info.Description != "A test module" {
		t.Errorf("ModuleInfo.Description = %q, want 'A test module'", info.Description)
	}
	if info.Color != "#ff0000" {
		t.Errorf("ModuleInfo.Color = %q, want '#ff0000'", info.Color)
	}
	if info.Standard != "Test Standard" {
		t.Errorf("ModuleInfo.Standard = %q, want 'Test Standard'", info.Standard)
	}
	if len(info.Tests) != 2 {
		t.Errorf("len(ModuleInfo.Tests) = %d, want 2", len(info.Tests))
	}
}

// TestRegistryConcurrency tests thread safety of registry operations.
func TestRegistryConcurrency(t *testing.T) {
	t.Parallel()
	// This test verifies that concurrent access doesn't cause data races.
	// Run with -race flag to detect issues.
	const goroutines = 10
	done := make(chan bool, goroutines)

	for range goroutines {
		go func() {
			// Multiple concurrent reads.
			_ = modules.GetModule(testModuleBenchmark)
			_ = modules.GetAllModules()
			_ = modules.DefaultRegistry.AllTestTypes()
			_ = modules.DefaultRegistry.TestCount()
			_ = modules.DefaultRegistry.ModuleCount()
			_ = modules.GetModuleForTest("throughput")
			_ = modules.GetAllModuleInfos()
			done <- true
		}()
	}

	// Wait for all goroutines.
	for range goroutines {
		<-done
	}
}
