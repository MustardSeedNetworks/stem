// SPDX-License-Identifier: BUSL-1.1

package trafficgen

// This file exports internal symbols for testing purposes.
// It is only compiled with test builds (due to _test.go suffix).

// NewMockExecutor creates an executor with nil context for testing.
// This allows testing Execute logic without requiring actual dataplane.
func NewMockExecutor() *Executor {
	return &Executor{
		Module: New(),
		ctx:    nil,
	}
}

// NewMockExecutorWithNilModule creates an executor with nil module and context.
// Used for testing edge cases.
func NewMockExecutorWithNilModule() *Executor {
	return &Executor{
		Module: nil,
		ctx:    nil,
	}
}

// Test constants exports for black-box testing.
const (
	TestDefaultRatePct         = defaultRatePct
	TestDefaultWarmupSec       = defaultWarmupSec
	TestDefaultDurationSec     = defaultDurationSec
	TestDefaultStreamID        = defaultStreamID
	TestDefaultBurstSize       = defaultBurstSize
	TestDefaultInterBurstGapUs = defaultInterBurstGapUs
)
