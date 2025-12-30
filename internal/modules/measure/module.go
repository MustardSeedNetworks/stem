// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// Package measure implements the Measure module for Y.1731 OAM testing.
// This module owns frame delay, frame loss, and loopback measurements.
package measure

const (
	// ModuleName is the unique identifier for the Measure module.
	ModuleName = "measure"

	// DisplayName is the human-readable name.
	DisplayName = "Measure"

	// ColorHex is the module's UI color (Blue).
	ColorHex = "#2563eb"

	// StandardRef is the primary standard this module implements.
	StandardRef = "ITU-T Y.1731"
)

// testTypes lists all test types owned by the Measure module.
var testTypes = []string{
	"y1731_delay",
	"y1731_loss",
	"y1731_slm",
	"y1731_loopback",
}

// testDescriptions provides descriptions for each test type.
var testDescriptions = map[string]string{
	"y1731_delay":    "ITU-T Y.1731 Frame Delay (DMM/DMR)",
	"y1731_loss":     "ITU-T Y.1731 Frame Loss (LMM/LMR)",
	"y1731_slm":      "ITU-T Y.1731 Synthetic Loss Measurement",
	"y1731_loopback": "ITU-T Y.1731 Loopback (LBM/LBR)",
}

// Module implements the modules.Module interface for Y.1731 OAM testing.
type Module struct{}

// New creates a new Measure module instance.
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
	return "Y.1731 Ethernet OAM - frame delay, loss, and loopback measurements"
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
