// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package benchmark

import (
	"errors"
	"fmt"
	"time"

	"github.com/krisarmstrong/stem/internal/testmaster/dataplane"
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

// ErrInvalidConfig is returned for invalid configuration.
var ErrInvalidConfig = errors.New("invalid test configuration")

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
func (e *Executor) Execute(testType string, cfg *TestConfig) (*Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("benchmark module cannot run test type: %s", testType)
	}

	if cfg == nil {
		return nil, ErrInvalidConfig
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
	result := &Result{
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
		return nil, ErrTestNotImplemented
	}

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	result.Data = data
	return result, nil
}

// configureContext sets up the dataplane context from test config.
func (e *Executor) configureContext(cfg *TestConfig) error {
	dpCfg := &dataplane.Config{
		Interface:  cfg.Interface,
		AutoDetect: true,
	}

	if cfg.Duration > 0 {
		dpCfg.TrialDuration = time.Duration(cfg.Duration) * time.Second
	}

	// Extract additional parameters
	if cfg.Params != nil {
		if resolution, ok := cfg.Params["resolution"].(float64); ok {
			dpCfg.ResolutionPct = resolution
		}
		if maxLoss, ok := cfg.Params["max_loss"].(float64); ok {
			dpCfg.AcceptableLoss = maxLoss
		}
		if warmup, ok := cfg.Params["warmup"].(int); ok {
			dpCfg.WarmupPeriod = time.Duration(warmup) * time.Second
		}
	}

	return e.ctx.Configure(dpCfg)
}

// getLoadLevels extracts load levels from config or returns defaults.
func (e *Executor) getLoadLevels(cfg *TestConfig) []float64 {
	if cfg.Params != nil {
		if levels, ok := cfg.Params["load_levels"].([]float64); ok {
			return levels
		}
	}
	return []float64{10, 25, 50, 75, 90, 100}
}

// getFrameLossParams extracts frame loss parameters from config.
func (e *Executor) getFrameLossParams(cfg *TestConfig) (float64, float64, float64) {
	startPct := 10.0
	endPct := 100.0
	stepPct := 10.0

	if cfg.Params != nil {
		if v, ok := cfg.Params["start_pct"].(float64); ok {
			startPct = v
		}
		if v, ok := cfg.Params["end_pct"].(float64); ok {
			endPct = v
		}
		if v, ok := cfg.Params["step_pct"].(float64); ok {
			stepPct = v
		}
	}

	return startPct, endPct, stepPct
}

// getBackToBackParams extracts back-to-back test parameters.
func (e *Executor) getBackToBackParams(cfg *TestConfig) (uint64, uint32) {
	initialBurst := uint64(10000)
	trials := uint32(3)

	if cfg.Params != nil {
		if v, ok := cfg.Params["initial_burst"].(uint64); ok {
			initialBurst = v
		}
		if v, ok := cfg.Params["trials"].(uint32); ok {
			trials = v
		}
	}

	return initialBurst, trials
}

// getRecoveryParams extracts system recovery test parameters.
func (e *Executor) getRecoveryParams(cfg *TestConfig) (float64, uint32) {
	throughputPct := 100.0
	overloadSec := uint32(60)

	if cfg.Params != nil {
		if v, ok := cfg.Params["throughput_pct"].(float64); ok {
			throughputPct = v
		}
		if v, ok := cfg.Params["overload_sec"].(uint32); ok {
			overloadSec = v
		}
	}

	return throughputPct, overloadSec
}
