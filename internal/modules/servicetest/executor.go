// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package servicetest

import (
	"fmt"
	"time"

	"github.com/krisarmstrong/stem/internal/modules/modtypes"
	"github.com/krisarmstrong/stem/internal/testmaster/dataplane"
)

// Default Y.1564 test parameters.
const (
	defaultServiceID       = 1
	defaultServiceName     = "Service-1"
	defaultFrameSize       = 1518
	defaultCIRMbps         = 100.0
	defaultEIRMbps         = 0.0
	defaultFDThresholdMs   = 10.0
	defaultFDVThresholdMs  = 5.0
	defaultFLRThresholdPct = 0.01
	defaultPerfDurationSec = 900 // 15 minutes
	maxUint32              = 4294967295
)

// Executor wraps the ServiceTest module with test execution capability.
type Executor struct {
	*Module

	ctx *dataplane.Context
}

// NewExecutor creates a new ServiceTest executor with a dataplane context.
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

// SupportsExecution returns true as ServiceTest supports test execution.
func (e *Executor) SupportsExecution() bool {
	return true
}

// Close releases the dataplane context resources.
func (e *Executor) Close() {
	if e.ctx != nil {
		e.ctx.Close()
	}
}

// Execute runs a Y.1564 or MEF test and returns the result.
func (e *Executor) Execute(testType string, cfg *modtypes.TestConfig) (*modtypes.Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("servicetest module cannot run test type: %s", testType)
	}

	if cfg == nil {
		return nil, modtypes.ErrInvalidConfig
	}

	// Configure the context.
	err := e.configureContext(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to configure context: %w", err)
	}

	// Execute the test.
	result := &modtypes.Result{
		TestType:   testType,
		ModuleName: ModuleName,
		Success:    false,
		Error:      "",
		Data:       nil,
	}

	var data any
	var runErr error

	switch testType {
	case "y1564_config":
		service := e.buildY1564Service(cfg)
		data, runErr = e.ctx.RunY1564ConfigTest(service)

	case "y1564_perf":
		service := e.buildY1564Service(cfg)
		duration := e.safeDuration(cfg.Duration, defaultPerfDurationSec)
		data, runErr = e.ctx.RunY1564PerfTest(service, duration)

	case "y1564":
		// Full Y.1564 test: config + performance.
		service := e.buildY1564Service(cfg)

		// Run config test first.
		configResult, configErr := e.ctx.RunY1564ConfigTest(service)
		if configErr != nil {
			result.Error = fmt.Sprintf("config test failed: %v", configErr)
			return result, fmt.Errorf("y1564 config test: %w", configErr)
		}

		// Run performance test.
		duration := e.safeDuration(cfg.Duration, defaultPerfDurationSec)
		perfResult, perfErr := e.ctx.RunY1564PerfTest(service, duration)
		if perfErr != nil {
			result.Error = fmt.Sprintf("perf test failed: %v", perfErr)
			return result, fmt.Errorf("y1564 perf test: %w", perfErr)
		}

		// Combine results.
		data = map[string]any{
			"config":      configResult,
			"performance": perfResult,
		}

	case "mef_config", "mef_perf", "mef":
		// MEF tests are similar to Y.1564, but not yet implemented in dataplane.
		return nil, modtypes.ErrTestNotImplemented

	default:
		return nil, modtypes.ErrTestNotImplemented
	}

	if runErr != nil {
		result.Error = runErr.Error()
		return result, fmt.Errorf("servicetest %s failed: %w", testType, runErr)
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

	err := e.ctx.Configure(dpCfg)
	if err != nil {
		return fmt.Errorf("configure dataplane: %w", err)
	}
	return nil
}

// buildY1564Service creates a Y1564Service from the test config.
func (e *Executor) buildY1564Service(cfg *modtypes.TestConfig) *dataplane.Y1564Service {
	service := &dataplane.Y1564Service{
		ServiceID:   defaultServiceID,
		ServiceName: defaultServiceName,
		SLA: dataplane.Y1564SLA{
			CIRMbps:         defaultCIRMbps,
			EIRMbps:         defaultEIRMbps,
			CBSBytes:        0,
			EBSBytes:        0,
			FDThresholdMs:   defaultFDThresholdMs,
			FDVThresholdMs:  defaultFDVThresholdMs,
			FLRThresholdPct: defaultFLRThresholdPct,
		},
		FrameSize: defaultFrameSize,
		CoS:       0,
		Enabled:   true,
	}

	if cfg.FrameSize > 0 {
		service.FrameSize = cfg.FrameSize
	}

	// Extract SLA and service parameters from config.
	e.extractY1564Params(cfg, service)

	return service
}

// extractY1564Params extracts SLA and service parameters from config using type-safe helpers.
func (e *Executor) extractY1564Params(cfg *modtypes.TestConfig, service *dataplane.Y1564Service) {
	if cfg.Params == nil {
		return
	}

	// Extract SLA parameters using type-safe helper.
	// Only update if parameter is explicitly set (check existence first).
	if _, ok := cfg.Params["cir"]; ok {
		service.SLA.CIRMbps = modtypes.GetFloat64Param(cfg.Params, "cir", service.SLA.CIRMbps)
	}
	if _, ok := cfg.Params["eir"]; ok {
		service.SLA.EIRMbps = modtypes.GetFloat64Param(cfg.Params, "eir", service.SLA.EIRMbps)
	}
	if _, ok := cfg.Params["fd_threshold"]; ok {
		service.SLA.FDThresholdMs = modtypes.GetFloat64Param(
			cfg.Params, "fd_threshold", service.SLA.FDThresholdMs,
		)
	}
	if _, ok := cfg.Params["fdv_threshold"]; ok {
		service.SLA.FDVThresholdMs = modtypes.GetFloat64Param(
			cfg.Params, "fdv_threshold", service.SLA.FDVThresholdMs,
		)
	}
	if _, ok := cfg.Params["flr_threshold"]; ok {
		service.SLA.FLRThresholdPct = modtypes.GetFloat64Param(
			cfg.Params, "flr_threshold", service.SLA.FLRThresholdPct,
		)
	}

	// Extract service identification.
	if name, ok := cfg.Params["service_name"].(string); ok {
		service.ServiceName = name
	}
	if v, ok := cfg.Params["service_id"]; ok {
		// Handle both uint32 and float64 (from JSON).
		switch id := v.(type) {
		case uint32:
			service.ServiceID = id
		case float64:
			if id >= 0 && id <= maxUint32 {
				service.ServiceID = uint32(id)
			}
		case int:
			if id >= 0 && id <= maxUint32 {
				service.ServiceID = uint32(id) // Safe: validated above.
			}
		}
	}
}

// safeDuration converts an int duration to uint32 safely.
// Returns defaultVal if duration is <= 0 or would overflow uint32.
func (e *Executor) safeDuration(duration int, defaultVal uint32) uint32 {
	if duration <= 0 || duration > maxUint32 {
		return defaultVal
	}
	return uint32(duration)
}
