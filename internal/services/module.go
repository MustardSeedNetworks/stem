// SPDX-License-Identifier: BUSL-1.1

// Package modules provides the module-oriented architecture layer for The Stem.
// Modules (Benchmark, ServiceTest, TrafficGen, Measure, Certify) own workflows
// and delegate to underlying subsystems (testmaster, reflector, dataplane).
package modules

// Module represents a testing module in The Stem.
// Each module owns specific workflows and maps to underlying subsystems.
type Module interface {
	// Name returns the module's unique identifier (e.g., "benchmark").
	Name() string

	// DisplayName returns the human-readable name (e.g., "Benchmark").
	DisplayName() string

	// Description returns a brief description of the module's purpose.
	Description() string

	// Color returns the module's UI color in hex format (e.g., "#dc2626").
	Color() string

	// TestTypes returns the list of test types this module can execute.
	TestTypes() []string

	// CanRun returns true if this module can execute the given test type.
	CanRun(testType string) bool

	// Standard returns the primary standard this module implements (e.g., "RFC 2544").
	Standard() string
}

// TestType represents a specific test that can be executed.
type TestType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Standard    string `json:"standard"`
	ModuleName  string `json:"module"`
}

// ModuleInfo contains metadata about a module for API responses.
type ModuleInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description"`
	Color       string   `json:"color"`
	Standard    string   `json:"standard"`
	Tests       []string `json:"tests"`
}

// ToInfo converts a Module to its API-friendly representation.
func ToInfo(m Module) ModuleInfo {
	return ModuleInfo{
		Name:        m.Name(),
		DisplayName: m.DisplayName(),
		Description: m.Description(),
		Color:       m.Color(),
		Standard:    m.Standard(),
		Tests:       m.TestTypes(),
	}
}
