// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package measure

import (
	"fmt"
	"math"

	"github.com/krisarmstrong/stem/internal/modules/modtypes"
	"github.com/krisarmstrong/stem/internal/testmaster/dataplane"
)

// Executor wraps the Measure module with test execution capability.
// Y.1731 OAM tests execute via the dataplane on supported platforms.
const (
	// Y.1731 defaults - aligned with TUI/WebUI.
	defaultMEPID       = 1
	defaultMEGLevel    = 4            // Match TUI/WebUI: service level.
	defaultCCMInterval = 1000         // Match TUI/WebUI: 1000ms (1s).
	defaultPriority    = 6            // Match TUI/WebUI: priority 6.
	defaultDuration    = 60           // Match TUI/WebUI: 60 seconds.
	defaultIntervalMs  = 100          // Match TUI/WebUI: 100ms measurement cadence.
	defaultCount       = 10           // 10 frames per interval.
	defaultFrameSize   = 64           // Match TUI/WebUI: 64 bytes.
	defaultMEGID       = "MSN-MEG-01" // Match TUI/WebUI default MEG ID.
)

type Executor struct {
	*Module

	ctx *dataplane.Context
}

// NewExecutor creates a new Measure executor with a dataplane context.
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

// SupportsExecution returns true as Measure can accept execution requests.
func (e *Executor) SupportsExecution() bool {
	return true
}

// Close releases any resources.
func (e *Executor) Close() {
	if e.ctx != nil {
		e.ctx.Close()
	}
}

// Execute runs a Y.1731 OAM test.
func (e *Executor) Execute(testType string, cfg *modtypes.TestConfig) (*modtypes.Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("measure module cannot run test type: %s", testType)
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

	ycfg := buildY1731Config(cfg)

	var data any
	var runErr error

	switch testType {
	case "y1731_delay":
		data, runErr = e.ctx.RunY1731DelayTest(ycfg)
	case "y1731_loss":
		data, runErr = e.ctx.RunY1731LossTest(ycfg)
	case "y1731_slm":
		data, runErr = e.ctx.RunY1731SyntheticLossTest(ycfg)
	case "y1731_loopback":
		data, runErr = e.ctx.RunY1731LoopbackTest(ycfg)
	default:
		return nil, modtypes.ErrTestNotImplemented
	}

	if runErr != nil {
		result.Error = runErr.Error()
		return result, fmt.Errorf("measure %s failed: %w", testType, runErr)
	}

	result.Success = true
	result.Data = data
	return result, nil
}

func buildY1731Config(cfg *modtypes.TestConfig) *dataplane.Y1731Config {
	config := &dataplane.Y1731Config{
		MEPID:          modtypes.GetUint32Param(cfg.Params, "mep_id", defaultMEPID),
		MEGLevel:       modtypes.GetUint32Param(cfg.Params, "meg_level", defaultMEGLevel),
		MEGID:          modtypes.GetStringParam(cfg.Params, "meg_id", defaultMEGID),
		CCMInterval:    modtypes.GetUint32Param(cfg.Params, "ccm_interval", defaultCCMInterval),
		Priority:       clampUint8(modtypes.GetUint32Param(cfg.Params, "priority", defaultPriority)),
		DurationSec:    safeUint32FromInt(cfg.Duration, 0),
		IntervalMs:     modtypes.GetUint32Param(cfg.Params, "interval_ms", defaultIntervalMs),
		Count:          modtypes.GetUint32Param(cfg.Params, "count", defaultCount),
		FrameSize:      cfg.FrameSize,
		PriorityTagged: modtypes.GetBoolParam(cfg.Params, "priority_tagged", true), // Match TUI/WebUI.
	}

	if config.DurationSec == 0 {
		config.DurationSec = modtypes.GetUint32Param(cfg.Params, "duration_sec", defaultDuration)
	}

	if config.FrameSize == 0 {
		config.FrameSize = modtypes.GetUint32Param(cfg.Params, "frame_size", defaultFrameSize)
	}

	return config
}

func safeUint32FromInt(value int, fallback uint32) uint32 {
	if value < 0 || value > math.MaxUint32 {
		return fallback
	}
	return uint32(value)
}

func clampUint8(value uint32) uint8 {
	if value > math.MaxUint8 {
		return math.MaxUint8
	}
	return uint8(value)
}
