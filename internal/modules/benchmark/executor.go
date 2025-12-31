// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package benchmark

import (
	"fmt"
	"time"

	"github.com/krisarmstrong/stem/internal/modules/common"
	"github.com/krisarmstrong/stem/internal/testmaster/dataplane"
)

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
func (e *Executor) Execute(testType string, cfg *common.TestConfig) (*common.Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("benchmark module cannot run test type: %s", testType)
	}

	if cfg == nil {
		return nil, common.ErrInvalidConfig
	}

	// Configure the context
	if err := e.configureContext(cfg); err != nil {
		return nil, fmt.Errorf("failed to configure context: %w", err)
	}

	// Set frame size if provided
	if cfg.FrameSize > 0 {
		e.ctx.SetFrameSize(cfg.FrameSize)
	}

	// Execute the test
	result := &common.Result{
		TestType:   testType,
		ModuleName: ModuleName,
		Success:    false,
	}

	var data interface{}
	var err error

	switch testType {
	case "throughput":
		data, err = e.ctx.RunThroughputTest()
	case "latency":
		loadLevels := e.getLoadLevels(cfg)
		data, err = e.ctx.RunLatencyTest(loadLevels)
	case "frame_loss":
		startPct, endPct, stepPct := e.getFrameLossParams(cfg)
		data, err = e.ctx.RunFrameLossTest(startPct, endPct, stepPct)
	case "back_to_back":
		initialBurst, trials := e.getBackToBackParams(cfg)
		data, err = e.ctx.RunBackToBackTest(initialBurst, trials)
	case "system_recovery":
		throughputPct, overloadSec := e.getRecoveryParams(cfg)
		data, err = e.ctx.RunSystemRecoveryTest(throughputPct, overloadSec)
	case "reset":
		data, err = e.ctx.RunResetTest()
	default:
		return nil, common.ErrTestNotImplemented
	}

	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("benchmark test %s failed: %w", testType, err)
	}

	result.Success = true
	result.Data = data
	return result, nil
}

// configureContext sets up the dataplane context from test config.
func (e *Executor) configureContext(cfg *common.TestConfig) error {
	dpCfg := &dataplane.Config{
		Interface:  cfg.Interface,
		AutoDetect: true,
	}

	if cfg.Duration > 0 {
		dpCfg.TrialDuration = time.Duration(cfg.Duration) * time.Second
	}

	// Extract additional parameters using type-safe helpers
	dpCfg.ResolutionPct = common.GetFloat64Param(cfg.Params, "resolution", 0.1)
	dpCfg.AcceptableLoss = common.GetFloat64Param(cfg.Params, "max_loss", 0.0)
	warmup := common.GetIntParam(cfg.Params, "warmup", 0)
	if warmup > 0 {
		dpCfg.WarmupPeriod = time.Duration(warmup) * time.Second
	}

	if err := e.ctx.Configure(dpCfg); err != nil {
		return fmt.Errorf("configure dataplane: %w", err)
	}
	return nil
}

// getLoadLevels extracts load levels from config or returns defaults.
func (e *Executor) getLoadLevels(cfg *common.TestConfig) []float64 {
	if cfg.Params != nil {
		if levels, ok := cfg.Params["load_levels"].([]float64); ok {
			return levels
		}
	}
	return []float64{10, 25, 50, 75, 90, 100}
}

// getFrameLossParams extracts frame loss parameters from config using type-safe helpers.
func (e *Executor) getFrameLossParams(cfg *common.TestConfig) (float64, float64, float64) {
	startPct := common.GetFloat64Param(cfg.Params, "start_pct", 10.0)
	endPct := common.GetFloat64Param(cfg.Params, "end_pct", 100.0)
	stepPct := common.GetFloat64Param(cfg.Params, "step_pct", 10.0)
	return startPct, endPct, stepPct
}

// getBackToBackParams extracts back-to-back test parameters using type-safe helpers.
func (e *Executor) getBackToBackParams(cfg *common.TestConfig) (uint64, uint32) {
	initialBurst := common.GetUint64Param(cfg.Params, "initial_burst", 10000)
	trials := common.GetUint32Param(cfg.Params, "trials", 3)
	return initialBurst, trials
}

// getRecoveryParams extracts system recovery test parameters using type-safe helpers.
func (e *Executor) getRecoveryParams(cfg *common.TestConfig) (float64, uint32) {
	throughputPct := common.GetFloat64Param(cfg.Params, "throughput_pct", 100.0)
	overloadSec := common.GetUint32Param(cfg.Params, "overload_sec", 60)
	return throughputPct, overloadSec
}
