// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package trafficgen

import (
	"fmt"

	"github.com/krisarmstrong/stem/internal/modules/modtypes"
)

// Executor wraps the TrafficGen module with execution capability.
// Custom traffic generation is not yet implemented in the dataplane.
type Executor struct {
	*Module

	iface string
}

// NewExecutor creates a new TrafficGen executor.
func NewExecutor(iface string) (*Executor, error) {
	return &Executor{
		Module: New(),
		iface:  iface,
	}, nil
}

// SupportsExecution returns true as TrafficGen can accept execution requests.
func (e *Executor) SupportsExecution() bool {
	return true
}

// Close releases any resources.
func (e *Executor) Close() {
	// No resources to release yet.
}

// Execute runs a traffic generation operation.
// Currently returns ErrTestNotImplemented as custom streams are not yet in the dataplane.
func (e *Executor) Execute(testType string, _ *modtypes.TestConfig) (*modtypes.Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("trafficgen module cannot run test type: %s", testType)
	}

	result := &modtypes.Result{
		TestType:   testType,
		ModuleName: ModuleName,
		Success:    false,
		Error:      "",
		Data:       nil,
	}

	switch testType {
	case "custom_stream":
		result.Error = "Custom traffic stream generation requires additional dataplane implementation"
		return result, modtypes.ErrTestNotImplemented

	default:
		return nil, modtypes.ErrTestNotImplemented
	}
}
