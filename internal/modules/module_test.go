// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package modules

import (
	"testing"

	"github.com/krisarmstrong/stem/internal/modules/benchmark"
	"github.com/krisarmstrong/stem/internal/modules/certify"
	"github.com/krisarmstrong/stem/internal/modules/measure"
	"github.com/krisarmstrong/stem/internal/modules/reflector"
	"github.com/krisarmstrong/stem/internal/modules/servicetest"
	"github.com/krisarmstrong/stem/internal/modules/trafficgen"
)

func TestRegistry(t *testing.T) {
	reg := NewRegistry()

	// Register a module
	bm := benchmark.New()
	reg.Register(bm)

	// Test Get
	if got := reg.Get("benchmark"); got == nil {
		t.Error("Get('benchmark') returned nil")
	}
	if got := reg.Get("nonexistent"); got != nil {
		t.Error("Get('nonexistent') should return nil")
	}

	// Test ModuleForTest
	if got := reg.ModuleForTest("throughput"); got == nil {
		t.Error("ModuleForTest('throughput') returned nil")
	}
	if got := reg.ModuleForTest("nonexistent"); got != nil {
		t.Error("ModuleForTest('nonexistent') should return nil")
	}

	// Test counts
	if reg.ModuleCount() != 1 {
		t.Errorf("ModuleCount() = %d, want 1", reg.ModuleCount())
	}
	if reg.TestCount() != 6 {
		t.Errorf("TestCount() = %d, want 6 (RFC 2544 tests)", reg.TestCount())
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Test that default registry has all 6 modules
	modules := DefaultRegistry.AllModules()
	if len(modules) != 6 {
		t.Errorf("DefaultRegistry has %d modules, want 6", len(modules))
	}

	// Test module lookup
	names := []string{"reflector", "benchmark", "servicetest", "trafficgen", "measure", "certify"}
	for _, name := range names {
		if m := DefaultRegistry.Get(name); m == nil {
			t.Errorf("DefaultRegistry.Get(%q) returned nil", name)
		}
	}
}

func TestBenchmarkModule(t *testing.T) {
	m := benchmark.New()

	if m.Name() != "benchmark" {
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

	// Test test types
	tests := m.TestTypes()
	if len(tests) != 6 {
		t.Errorf("TestTypes() returned %d tests, want 6", len(tests))
	}

	// Test CanRun
	if !m.CanRun("throughput") {
		t.Error("CanRun('throughput') should return true")
	}
	if m.CanRun("y1564_config") {
		t.Error("CanRun('y1564_config') should return false")
	}
}

func TestServiceTestModule(t *testing.T) {
	m := servicetest.New()

	if m.Name() != "servicetest" {
		t.Errorf("Name() = %q, want 'servicetest'", m.Name())
	}
	if m.Color() != "#ea580c" {
		t.Errorf("Color() = %q, want '#ea580c'", m.Color())
	}

	// Test Y.1564 tests
	if !m.CanRun("y1564_config") {
		t.Error("CanRun('y1564_config') should return true")
	}
	if !m.CanRun("mef") {
		t.Error("CanRun('mef') should return true")
	}
}

func TestTrafficGenModule(t *testing.T) {
	m := trafficgen.New()

	if m.Name() != "trafficgen" {
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

	if m.Name() != "reflector" {
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

	if m.Name() != "measure" {
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

	if m.Name() != "certify" {
		t.Errorf("Name() = %q, want 'certify'", m.Name())
	}
	if m.Color() != "#16a34a" {
		t.Errorf("Color() = %q, want '#16a34a'", m.Color())
	}

	// Test various standards
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
	info := ToInfo(m)

	if info.Name != "benchmark" {
		t.Errorf("info.Name = %q, want 'benchmark'", info.Name)
	}
	if info.DisplayName != "Benchmark" {
		t.Errorf("info.DisplayName = %q, want 'Benchmark'", info.DisplayName)
	}
	if info.Color != "#dc2626" {
		t.Errorf("info.Color = %q, want '#dc2626'", info.Color)
	}
	if len(info.Tests) != 6 {
		t.Errorf("len(info.Tests) = %d, want 6", len(info.Tests))
	}
}

func TestGetModuleForTest(t *testing.T) {
	// RFC 2544 tests -> benchmark
	if m := GetModuleForTest("throughput"); m == nil || m.Name() != "benchmark" {
		t.Error("throughput should map to benchmark module")
	}

	// Y.1564 tests -> servicetest
	if m := GetModuleForTest("y1564_config"); m == nil || m.Name() != "servicetest" {
		t.Error("y1564_config should map to servicetest module")
	}

	// Y.1731 tests -> measure
	if m := GetModuleForTest("y1731_delay"); m == nil || m.Name() != "measure" {
		t.Error("y1731_delay should map to measure module")
	}

	// RFC 2889/6349/TSN tests -> certify
	if m := GetModuleForTest("rfc2889_forwarding"); m == nil || m.Name() != "certify" {
		t.Error("rfc2889_forwarding should map to certify module")
	}
	if m := GetModuleForTest("rfc6349_throughput"); m == nil || m.Name() != "certify" {
		t.Error("rfc6349_throughput should map to certify module")
	}
}

func TestAllModuleInfos(t *testing.T) {
	infos := GetAllModuleInfos()
	if len(infos) != 6 {
		t.Errorf("GetAllModuleInfos() returned %d infos, want 6", len(infos))
	}
}

func TestGetModuleForReflect(t *testing.T) {
	// reflect -> reflector module (not trafficgen)
	if m := GetModuleForTest("reflect"); m == nil || m.Name() != "reflector" {
		t.Error("reflect should map to reflector module")
	}

	// custom_stream -> trafficgen module
	if m := GetModuleForTest("custom_stream"); m == nil || m.Name() != "trafficgen" {
		t.Error("custom_stream should map to trafficgen module")
	}
}
