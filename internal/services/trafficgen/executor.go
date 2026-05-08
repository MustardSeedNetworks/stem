// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package trafficgen

import (
	"fmt"

	"github.com/krisarmstrong/stem/internal/services/modtypes"
	"github.com/krisarmstrong/stem/internal/services/orchestrator/dataplane"
)

const (
	defaultRatePct         = 100.0 // Match TUI/WebUI: 100% line rate.
	defaultWarmupSec       = 2     // Match TUI/WebUI: 2 seconds.
	defaultDurationSec     = 60    // Match TUI/WebUI: 60 seconds.
	defaultStreamID        = 1     // Match TUI/WebUI: stream 1.
	defaultBurstSize       = 100
	defaultInterBurstGapUs = 1000
)

// Executor wraps the TrafficGen module with test execution capability.
type Executor struct {
	*Module

	ctx *dataplane.Context
}

// NewExecutor creates a new TrafficGen executor with a dataplane context.
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

// SupportsExecution returns true as TrafficGen supports test execution.
func (e *Executor) SupportsExecution() bool {
	return true
}

// Close releases the dataplane context resources.
func (e *Executor) Close() {
	if e.ctx != nil {
		e.ctx.Close()
	}
}

// Execute runs a custom traffic generation test.
func (e *Executor) Execute(testType string, cfg *modtypes.TestConfig) (*modtypes.Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("trafficgen module cannot run test type: %s", testType)
	}

	if cfg == nil {
		return nil, modtypes.ErrInvalidConfig
	}

	result := &modtypes.Result{
		TestType:   testType,
		ModuleName: ModuleName,
		Success:    false,
		Error:      "",
		Data:       nil,
	}

	if testType != "custom_stream" {
		return nil, modtypes.ErrTestNotImplemented
	}

	if e.ctx == nil {
		result.Error = "dataplane context is not configured"
		return result, fmt.Errorf("trafficgen %s failed: %s", testType, result.Error)
	}

	config := &dataplane.TrafficGenConfig{
		FrameSize:       cfg.FrameSize,
		RatePct:         modtypes.GetFloat64Param(cfg.Params, "rate_pct", defaultRatePct),
		DurationSec:     modtypes.SafeIntToUint32(cfg.Duration),
		WarmupSec:       modtypes.GetUint32Param(cfg.Params, "warmup_sec", defaultWarmupSec),
		StreamID:        modtypes.GetUint32Param(cfg.Params, "stream_id", defaultStreamID),
		BurstMode:       modtypes.GetBoolParam(cfg.Params, "burst_mode", false),
		BurstSize:       modtypes.GetUint32Param(cfg.Params, "burst_size", defaultBurstSize),
		InterBurstGapUs: modtypes.GetUint32Param(cfg.Params, "inter_burst_gap_us", defaultInterBurstGapUs),
		SrcMac:          modtypes.GetStringParam(cfg.Params, "src_mac", ""),
		DstMac:          modtypes.GetStringParam(cfg.Params, "dst_mac", ""),
		VlanID:          modtypes.GetUint16Param(cfg.Params, "vlan_id", 0),
		VlanPriority:    modtypes.GetUint8Param(cfg.Params, "vlan_priority", 0),
	}

	if config.DurationSec == 0 {
		config.DurationSec = modtypes.GetUint32Param(cfg.Params, "duration_sec", defaultDurationSec)
	}

	data, runErr := e.ctx.RunCustomStreamTest(config)
	if runErr != nil {
		result.Error = runErr.Error()
		return result, fmt.Errorf("trafficgen %s failed: %w", testType, runErr)
	}

	result.Success = true
	result.Data = data
	return result, nil
}
