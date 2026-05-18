// SPDX-License-Identifier: BUSL-1.1

// Package trafficgen implements the TrafficGen module for custom traffic generation.
// This module owns custom traffic stream generation capabilities.
package trafficgen

import "slices"

const (
	// ModuleName is the unique identifier for the TrafficGen module.
	ModuleName = "trafficgen"

	// DisplayName is the human-readable name.
	DisplayName = "TrafficGen"

	// ColorHex is the module's UI color (Yellow).
	ColorHex = "#ca8a04"

	// StandardRef is the primary standard/mode this module implements.
	StandardRef = "Custom Traffic"
)

func testTypes() []string {
	return []string{
		"custom_stream",
	}
}

func testDescriptions() map[string]string {
	return map[string]string{
		"custom_stream": "Custom traffic stream generation with configurable patterns",
	}
}

// Module implements the modules.Module interface for traffic generation.
type Module struct{}

// New creates a new TrafficGen module instance.
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
	return "Custom traffic stream generation with configurable patterns"
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
	return testTypes()
}

// CanRun returns true if this module can execute the given test type.
func (m *Module) CanRun(testType string) bool {
	return slices.Contains(testTypes(), testType)
}

// TestDescription returns the description for a given test type.
func (m *Module) TestDescription(testType string) string {
	return testDescriptions()[testType]
}
