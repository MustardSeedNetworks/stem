// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package certify

import (
	"fmt"

	"github.com/krisarmstrong/stem/internal/modules/common"
)

// Executor wraps the Certify module with test execution capability.
// RFC 2889, RFC 6349, and TSN tests are not yet implemented in the dataplane.
type Executor struct {
	*Module
	iface string
}

// NewExecutor creates a new Certify executor.
func NewExecutor(iface string) (*Executor, error) {
	return &Executor{
		Module: New(),
		iface:  iface,
	}, nil
}

// SupportsExecution returns true as Certify can accept execution requests.
func (e *Executor) SupportsExecution() bool {
	return true
}

// Close releases any resources.
func (e *Executor) Close() {
	// No resources to release yet
}

// Execute runs an RFC 2889, RFC 6349, or TSN test.
// Currently returns ErrTestNotImplemented for unimplemented tests.
func (e *Executor) Execute(testType string, _ *common.TestConfig) (*common.Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("certify module cannot run test type: %s", testType)
	}

	result := &common.Result{
		TestType:   testType,
		ModuleName: ModuleName,
		Success:    false,
	}

	// These tests are defined but not yet implemented in the C dataplane
	switch testType {
	case "rfc2889_forwarding", "rfc2889_caching", "rfc2889_learning",
		"rfc2889_broadcast", "rfc2889_congestion":
		result.Error = "RFC 2889 switch tests require additional dataplane implementation"
		return result, common.ErrTestNotImplemented

	case "rfc6349_throughput", "rfc6349_path":
		result.Error = "RFC 6349 TCP tests require additional dataplane implementation"
		return result, common.ErrTestNotImplemented

	case "tsn_timing", "tsn_isolation", "tsn_latency", "tsn":
		result.Error = "IEEE 802.1Qbv TSN tests require additional dataplane implementation"
		return result, common.ErrTestNotImplemented

	default:
		return nil, common.ErrTestNotImplemented
	}
}
