// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
//
// Package servicetest implements the ServiceTest module for Y.1564 and MEF testing.
// This module owns service activation and performance validation tests.
package servicetest

const (
	// ModuleName is the unique identifier for the ServiceTest module.
	ModuleName = "servicetest"

	// DisplayName is the human-readable name.
	DisplayName = "ServiceTest"

	// ColorHex is the module's UI color (Orange).
	ColorHex = "#ea580c"

	// StandardRef is the primary standard this module implements.
	StandardRef = "ITU-T Y.1564"
)

// testTypes lists all test types owned by the ServiceTest module.
var testTypes = []string{
	// Y.1564 EtherSAM
	"y1564_config",
	"y1564_perf",
	"y1564",
	// MEF Service
	"mef_config",
	"mef_perf",
	"mef",
}

// testDescriptions provides descriptions for each test type.
var testDescriptions = map[string]string{
	"y1564_config": "ITU-T Y.1564 Service Configuration Test",
	"y1564_perf":   "ITU-T Y.1564 Service Performance Test (15+ min)",
	"y1564":        "ITU-T Y.1564 Full Test (config + performance)",
	"mef_config":   "MEF 48/49 Service Configuration Test",
	"mef_perf":     "MEF 48/49 Service Performance Test",
	"mef":          "MEF 48/49 Full Test Suite",
}

// Module implements the modules.Module interface for service activation testing.
type Module struct{}

// New creates a new ServiceTest module instance.
func New() *Module {
	return &Module{}
}

// Name returns the module's unique identifier.
func (m *Module) Name() string {
	return ModuleName
}

// DisplayName returns the human-readable name.
func (m *Module) DisplayName() string {
	return DisplayName
}

// Description returns a brief description of the module's purpose.
func (m *Module) Description() string {
	return "Y.1564 and MEF service activation - configuration and performance validation"
}

// Color returns the module's UI color in hex format.
func (m *Module) Color() string {
	return ColorHex
}

// Standard returns the primary standard this module implements.
func (m *Module) Standard() string {
	return StandardRef
}

// TestTypes returns the list of test types this module can execute.
func (m *Module) TestTypes() []string {
	return testTypes
}

// CanRun returns true if this module can execute the given test type.
func (m *Module) CanRun(testType string) bool {
	for _, t := range testTypes {
		if t == testType {
			return true
		}
	}
	return false
}

// TestDescription returns the description for a given test type.
func (m *Module) TestDescription(testType string) string {
	return testDescriptions[testType]
}
