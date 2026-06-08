// SPDX-License-Identifier: BUSL-1.1

// Package benchmark implements the Benchmark module for RFC 2544 device testing.
// This module owns throughput, latency, frame loss, back-to-back, and recovery tests.
package benchmark

import "slices"

const (
	// ModuleName is the unique identifier for the Benchmark module.
	ModuleName = "benchmark"

	// DisplayName is the human-readable name.
	DisplayName = "Benchmark"

	// ColorHex is the module's UI color (Red).
	ColorHex = "#dc2626"

	// StandardRef is the primary standard this module implements.
	StandardRef = "RFC 2544"
)

func testTypes() []string {
	return []string{
		"rfc2544_throughput",
		"rfc2544_latency",
		"rfc2544_frame_loss",
		"rfc2544_back_to_back",
		"rfc2544_system_recovery",
		"rfc2544_reset",
	}
}

func testDescriptions() map[string]string {
	return map[string]string{
		"rfc2544_throughput":      "RFC 2544 Section 26.1 - Maximum throughput with zero loss",
		"rfc2544_latency":         "RFC 2544 Section 26.2 - Round-trip latency at various loads",
		"rfc2544_frame_loss":      "RFC 2544 Section 26.3 - Frame loss rate vs offered load",
		"rfc2544_back_to_back":    "RFC 2544 Section 26.4 - Maximum burst capacity",
		"rfc2544_system_recovery": "RFC 2544 Section 26.5 - Recovery time after overload",
		"rfc2544_reset":           "RFC 2544 Section 26.6 - Device reset recovery time",
	}
}

// Module implements the services.Module interface for RFC 2544 benchmarking.
type Module struct{}

// New creates a new Benchmark module instance.
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
	return "RFC 2544 device benchmarking - throughput, latency, frame loss, and recovery tests"
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
