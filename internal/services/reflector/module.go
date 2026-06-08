// SPDX-License-Identifier: BUSL-1.1

// Package reflector implements the Reflector module for packet reflection/loopback.
// This module owns the reflector operational mode (Tier 1) for remote testing support.
package reflector

import "slices"

const (
	// ModuleName is the unique identifier for the Reflector module.
	ModuleName = "reflector"

	// DisplayName is the human-readable name.
	DisplayName = "Reflector"

	// ColorHex is the module's UI color (Cyan - distinct from other modules).
	ColorHex = "#0891b2"

	// StandardRef indicates this is a standalone operational mode.
	StandardRef = "Loopback/Echo"
)

func testTypes() []string {
	return []string{
		"reflect",
	}
}

func testDescriptions() map[string]string {
	return map[string]string{
		"reflect": "Packet reflector mode - echoes received packets for remote device testing",
	}
}

// Module implements the services.Module interface for packet reflection.
type Module struct{}

// New creates a new Reflector module instance.
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
	return "Packet reflection/loopback for remote device testing (Tier 1 mode)"
}

// Color returns the module's UI color in hex format.
func (m *Module) Color() string {
	return ColorHex
}

// Standard returns the operational mode this module implements.
func (m *Module) Standard() string {
	return StandardRef
}

// TestTypes returns the list of operation types this module can execute.
func (m *Module) TestTypes() []string {
	return testTypes()
}

// CanRun returns true if this module can execute the given operation type.
func (m *Module) CanRun(testType string) bool {
	return slices.Contains(testTypes(), testType)
}

// TestDescription returns the description for a given operation type.
func (m *Module) TestDescription(testType string) string {
	return testDescriptions()[testType]
}
