// SPDX-License-Identifier: BUSL-1.1

package trafficgen

import (
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

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

// NewExecutorWithTestContext creates an executor holding a dataplane context
// that owns no C resources. Close and Cancel are safe against it on every
// build; Execute is not, because RunCustomStreamTest hands the absent C context
// straight to the dataplane (issue #1096).
func NewExecutorWithTestContext() *Executor {
	return &Executor{
		Module: New(),
		ctx:    dataplane.NewTestContext(),
	}
}

// BuildTrafficGenConfig exposes the parameter-defaulting logic that Execute
// applies before handing a config to the dataplane.
func BuildTrafficGenConfig(cfg *modtypes.TestConfig) *dataplane.TrafficGenConfig {
	return buildTrafficGenConfig(cfg)
}
