// SPDX-License-Identifier: BUSL-1.1

package benchmark

import "github.com/krisarmstrong/stem/internal/services/modtypes"

// DefaultLoadLevelsForTest exposes default load levels for tests.
func DefaultLoadLevelsForTest() []float64 {
	return defaultLoadLevels()
}

// LoadLevelsForTest exposes load level parsing for tests.
func LoadLevelsForTest(exec *Executor, cfg *modtypes.TestConfig) []float64 {
	return exec.getLoadLevels(cfg)
}

// FrameLossParamsForTest exposes frame loss parameters for tests.
func FrameLossParamsForTest(exec *Executor, cfg *modtypes.TestConfig) (float64, float64, float64) {
	return exec.getFrameLossParams(cfg)
}

// BackToBackParamsForTest exposes back-to-back parameters for tests.
func BackToBackParamsForTest(exec *Executor, cfg *modtypes.TestConfig) (uint64, uint32) {
	return exec.getBackToBackParams(cfg)
}

// RecoveryParamsForTest exposes recovery parameters for tests.
func RecoveryParamsForTest(exec *Executor, cfg *modtypes.TestConfig) (float64, uint32) {
	return exec.getRecoveryParams(cfg)
}

// ConfigureContextForTest exposes context configuration for tests.
func ConfigureContextForTest(exec *Executor, cfg *modtypes.TestConfig) error {
	return exec.configureContext(cfg)
}

// DefaultResolutionForTest exposes default resolution.
func DefaultResolutionForTest() float64 {
	return defaultResolution
}

// DefaultAcceptableLossForTest exposes default acceptable loss.
func DefaultAcceptableLossForTest() float64 {
	return defaultAcceptableLoss
}

// DefaultStartPctForTest exposes default start percentage.
func DefaultStartPctForTest() float64 {
	return defaultStartPct
}

// DefaultEndPctForTest exposes default end percentage.
func DefaultEndPctForTest() float64 {
	return defaultEndPct
}

// DefaultStepPctForTest exposes default step percentage.
func DefaultStepPctForTest() float64 {
	return defaultStepPct
}

// DefaultInitialBurstForTest exposes default initial burst.
func DefaultInitialBurstForTest() uint64 {
	return defaultInitialBurst
}

// DefaultTrialsForTest exposes default trial count.
func DefaultTrialsForTest() uint32 {
	return defaultTrials
}

// DefaultThroughputPctForTest exposes default throughput percentage.
func DefaultThroughputPctForTest() float64 {
	return defaultThroughputPct
}

// DefaultOverloadSecForTest exposes default overload seconds.
func DefaultOverloadSecForTest() uint32 {
	return defaultOverloadSec
}
