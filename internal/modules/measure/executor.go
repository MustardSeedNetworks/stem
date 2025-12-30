// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package measure

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

// Executor wraps the Measure module with test execution capability.
// Y.1731 OAM tests are not yet implemented in the dataplane.
type Executor struct {
	*Module
	iface string
}

// NewExecutor creates a new Measure executor.
func NewExecutor(iface string) (*Executor, error) {
	return &Executor{
		Module: New(),
		iface:  iface,
	}, nil
}

// SupportsExecution returns true as Measure can accept execution requests.
func (e *Executor) SupportsExecution() bool {
	return true
}

// Close releases any resources.
func (e *Executor) Close() {
	// No resources to release yet
}

// Execute runs a Y.1731 OAM test.
// Currently returns ErrTestNotImplemented as Y.1731 is not yet in the dataplane.
func (e *Executor) Execute(testType string, cfg *TestConfig) (*Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("measure module cannot run test type: %s", testType)
	}

	result := &Result{
		TestType:   testType,
		ModuleName: ModuleName,
		Success:    false,
	}

	// Y.1731 tests are defined but not yet implemented in the C dataplane
	switch testType {
	case "y1731_delay", "y1731_loss", "y1731_slm", "y1731_loopback":
		result.Error = "Y.1731 OAM tests require additional dataplane implementation"
		return result, ErrTestNotImplemented
	default:
		return nil, ErrTestNotImplemented
	}
}
