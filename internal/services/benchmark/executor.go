// SPDX-License-Identifier: BUSL-1.1

package benchmark

import (
	"fmt"
	"time"

	"github.com/krisarmstrong/stem/internal/services/modtypes"
	"github.com/krisarmstrong/stem/internal/services/orchestrator/dataplane"
)

// Default test parameters.
const (
	defaultResolution     = 0.1
	defaultAcceptableLoss = 0.0
	defaultStartPct       = 10.0
	defaultEndPct         = 100.0
	defaultStepPct        = 10.0
	defaultInitialBurst   = 10000
	defaultTrials         = 3
	defaultThroughputPct  = 100.0
	defaultOverloadSec    = 60
)

// defaultLoadLevels returns default load levels for latency tests.
func defaultLoadLevels() []float64 {
	return []float64{10, 25, 50, 75, 90, 100}
}

// Executor wraps the Benchmark module with test execution capability.
type Executor struct {
	*Module

	ctx *dataplane.Context
}

// NewExecutor creates a new Benchmark executor with a dataplane context.
func NewExecutor(iface string) (*Executor, error) {
	ctx, err := dataplane.NewContext(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to create dataplane context: %w", err)
	}

	return &Executor{
		Module: New(),
		ctx:    ctx,
	}, nil
}

// NewExecutorWithContext creates an executor with an existing dataplane context.
func NewExecutorWithContext(ctx *dataplane.Context) *Executor {
	return &Executor{
		Module: New(),
		ctx:    ctx,
	}
}

// SupportsExecution returns true as Benchmark supports test execution.
func (e *Executor) SupportsExecution() bool {
	return true
}

// Close releases the dataplane context resources.
func (e *Executor) Close() {
	if e.ctx != nil {
		e.ctx.Close()
	}
}

// Execute runs an RFC 2544 test and returns the result.
func (e *Executor) Execute(testType string, cfg *modtypes.TestConfig) (*modtypes.Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("benchmark module cannot run test type: %s", testType)
	}

	if cfg == nil {
		return nil, modtypes.ErrInvalidConfig
	}

	// Create result struct early to ensure it's populated even on error.
	result := &modtypes.Result{
		TestType:   testType,
		ModuleName: ModuleName,
		Success:    false,
		Error:      "",
		Data:       nil,
	}

	// Configure the context if available.
	if e.ctx != nil {
		err := e.configureContext(cfg)
		if err != nil {
			result.Error = err.Error()
			return result, fmt.Errorf("failed to configure context: %w", err)
		}

		// Set frame size if provided.
		if cfg.FrameSize > 0 {
			e.ctx.SetFrameSize(cfg.FrameSize)
		}
	}

	var data any
	var runErr error

	switch testType {
	case "rfc2544_throughput":
		data, runErr = e.ctx.RunThroughputTest()
	case "rfc2544_latency":
		loadLevels := e.getLoadLevels(cfg)
		data, runErr = e.ctx.RunLatencyTest(loadLevels)
	case "rfc2544_frame_loss":
		startPct, endPct, stepPct := e.getFrameLossParams(cfg)
		data, runErr = e.ctx.RunFrameLossTest(startPct, endPct, stepPct)
	case "rfc2544_back_to_back":
		initialBurst, trials := e.getBackToBackParams(cfg)
		data, runErr = e.ctx.RunBackToBackTest(initialBurst, trials)
	case "rfc2544_system_recovery":
		throughputPct, overloadSec := e.getRecoveryParams(cfg)
		data, runErr = e.ctx.RunSystemRecoveryTest(throughputPct, overloadSec)
	case "rfc2544_reset":
		data, runErr = e.ctx.RunResetTest()
	default:
		return nil, modtypes.ErrTestNotImplemented
	}

	if runErr != nil {
		result.Error = runErr.Error()
		return result, fmt.Errorf("benchmark test %s failed: %w", testType, runErr)
	}

	result.Success = true
	result.Data = data
	return result, nil
}

// configureContext sets up the dataplane context from test config.
func (e *Executor) configureContext(cfg *modtypes.TestConfig) error {
	dpCfg := &dataplane.Config{
		Interface:      cfg.Interface,
		LineRate:       0,
		AutoDetect:     true,
		TestType:       0,
		FrameSize:      0,
		IncludeJumbo:   false,
		TrialDuration:  0,
		WarmupPeriod:   0,
		InitialRatePct: 0,
		ResolutionPct:  0,
		MaxIterations:  0,
		AcceptableLoss: 0,
		HWTimestamp:    false,
		MeasureLatency: false,
		UsePacing:      false,
		BatchSize:      0,
		UseDPDK:        false,
		DPDKArgs:       "",
	}

	if cfg.Duration > 0 {
		dpCfg.TrialDuration = time.Duration(cfg.Duration) * time.Second
	}

	// Extract additional parameters using type-safe helpers.
	dpCfg.ResolutionPct = modtypes.GetFloat64Param(cfg.Params, "resolution", defaultResolution)
	dpCfg.AcceptableLoss = modtypes.GetFloat64Param(cfg.Params, "max_loss", defaultAcceptableLoss)
	warmup := modtypes.GetIntParam(cfg.Params, "warmup", 0)
	if warmup > 0 {
		dpCfg.WarmupPeriod = time.Duration(warmup) * time.Second
	}

	err := e.ctx.Configure(dpCfg)
	if err != nil {
		return fmt.Errorf("configure dataplane: %w", err)
	}
	return nil
}

// getLoadLevels extracts load levels from config or returns defaults.
func (e *Executor) getLoadLevels(cfg *modtypes.TestConfig) []float64 {
	if cfg.Params != nil {
		if levels, ok := cfg.Params["load_levels"].([]float64); ok {
			return levels
		}
	}
	return defaultLoadLevels()
}

// getFrameLossParams extracts frame loss parameters from config using type-safe helpers.
func (e *Executor) getFrameLossParams(cfg *modtypes.TestConfig) (float64, float64, float64) {
	startPct := modtypes.GetFloat64Param(cfg.Params, "start_pct", defaultStartPct)
	endPct := modtypes.GetFloat64Param(cfg.Params, "end_pct", defaultEndPct)
	stepPct := modtypes.GetFloat64Param(cfg.Params, "step_pct", defaultStepPct)
	return startPct, endPct, stepPct
}

// getBackToBackParams extracts back-to-back test parameters using type-safe helpers.
func (e *Executor) getBackToBackParams(cfg *modtypes.TestConfig) (uint64, uint32) {
	initialBurst := modtypes.GetUint64Param(cfg.Params, "initial_burst", defaultInitialBurst)
	trials := modtypes.GetUint32Param(cfg.Params, "trials", defaultTrials)
	return initialBurst, trials
}

// getRecoveryParams extracts system recovery test parameters using type-safe helpers.
func (e *Executor) getRecoveryParams(cfg *modtypes.TestConfig) (float64, uint32) {
	throughputPct := modtypes.GetFloat64Param(cfg.Params, "throughput_pct", defaultThroughputPct)
	overloadSec := modtypes.GetUint32Param(cfg.Params, "overload_sec", defaultOverloadSec)
	return throughputPct, overloadSec
}
