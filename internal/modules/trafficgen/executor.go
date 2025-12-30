// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package trafficgen

import (
	"errors"
	"fmt"
)

// Result is a generic test result.
type Result struct {
	TestType   string      `json:"testType"`
	ModuleName string      `json:"module"`
	Success    bool        `json:"success"`
	Error      string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

// TestConfig holds configuration for test execution.
type TestConfig struct {
	Interface string
	FrameSize uint32
	Duration  int
	Params    map[string]interface{}
}

// ErrTestNotImplemented is returned for unimplemented tests.
var ErrTestNotImplemented = errors.New("test type not implemented")

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
	// No resources to release yet
}

// Execute runs a traffic generation operation.
// Currently returns ErrTestNotImplemented as custom streams are not yet in the dataplane.
func (e *Executor) Execute(testType string, _ *TestConfig) (*Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("trafficgen module cannot run test type: %s", testType)
	}

	result := &Result{
		TestType:   testType,
		ModuleName: ModuleName,
		Success:    false,
	}

	switch testType {
	case "custom_stream":
		result.Error = "Custom traffic stream generation requires additional dataplane implementation"
		return result, ErrTestNotImplemented

	default:
		return nil, ErrTestNotImplemented
	}
}
