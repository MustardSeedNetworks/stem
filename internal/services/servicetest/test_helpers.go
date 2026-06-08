// SPDX-License-Identifier: BUSL-1.1

package servicetest

import (
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

// BuildY1564ServiceForTest exposes service building for tests.
func BuildY1564ServiceForTest(exec *Executor, cfg *modtypes.TestConfig) *dataplane.Y1564Service {
	return exec.buildY1564Service(cfg)
}

// BuildMEFConfigForTest exposes MEF config building for tests.
func BuildMEFConfigForTest(exec *Executor, cfg *modtypes.TestConfig) *dataplane.MEFConfig {
	return exec.buildMEFConfig(cfg)
}

// ExtractY1564ParamsForTest exposes parameter extraction for tests.
func ExtractY1564ParamsForTest(exec *Executor, cfg *modtypes.TestConfig, service *dataplane.Y1564Service) {
	exec.extractY1564Params(cfg, service)
}

// ConfigureContextForTest exposes context configuration for tests.
func ConfigureContextForTest(exec *Executor, cfg *modtypes.TestConfig) error {
	return exec.configureContext(cfg)
}

// ContextForTest exposes executor context for tests.
func ContextForTest(exec *Executor) *dataplane.Context {
	return exec.ctx
}

// RunY1564ForTest exposes Y.1564 execution branches for tests.
func RunY1564ForTest(exec *Executor, testType string, cfg *modtypes.TestConfig) (any, error) {
	return exec.runY1564(testType, cfg)
}

// RunMEFForTest exposes MEF execution branches for tests.
func RunMEFForTest(exec *Executor, testType string, cfg *modtypes.TestConfig) (any, error) {
	return exec.runMEF(testType, cfg)
}

// DefaultServiceIDForTest exposes default service ID.
func DefaultServiceIDForTest() uint32 {
	return defaultServiceID
}

// DefaultServiceNameForTest exposes default service name.
func DefaultServiceNameForTest() string {
	return defaultServiceName
}

// DefaultFrameSizeForTest exposes default frame size.
func DefaultFrameSizeForTest() uint32 {
	return defaultFrameSize
}

// DefaultCIRMbpsForTest exposes default CIR Mbps.
func DefaultCIRMbpsForTest() float64 {
	return defaultCIRMbps
}

// DefaultEIRMbpsForTest exposes default EIR Mbps.
func DefaultEIRMbpsForTest() float64 {
	return defaultEIRMbps
}

// DefaultMEFConfigDurationSecForTest exposes default MEF config duration.
func DefaultMEFConfigDurationSecForTest() uint32 {
	return defaultMEFConfigDurationSec
}

// DefaultMEFPerfDurationMinForTest exposes default MEF perf duration.
func DefaultMEFPerfDurationMinForTest() uint32 {
	return defaultMEFPerfDurationMin
}

// DefaultAvailabilityPctForTest exposes default availability percentage.
func DefaultAvailabilityPctForTest() float64 {
	return defaultAvailabilityPct
}

// DefaultPerfDurationSecForTest exposes default performance duration in seconds.
func DefaultPerfDurationSecForTest() uint32 {
	return defaultPerfDurationSec
}

// MaxUint32ForTest exposes max uint32 value.
func MaxUint32ForTest() uint32 {
	return maxUint32
}

// SafeDurationForTest exposes safeDuration for tests.
func SafeDurationForTest(exec *Executor, duration int, fallback uint32) uint32 {
	return exec.safeDuration(duration, fallback)
}
