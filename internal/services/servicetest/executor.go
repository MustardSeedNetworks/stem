// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package servicetest

import (
	"errors"
	"fmt"
	"time"

	"github.com/krisarmstrong/stem/internal/services/modtypes"
	"github.com/krisarmstrong/stem/internal/services/orchestrator/dataplane"
)

// Default Y.1564 test parameters.
const (
	defaultServiceID            = 1
	defaultServiceName          = "Service-1"
	defaultFrameSize            = 1518
	defaultCIRMbps              = 100.0
	defaultEIRMbps              = 0.0
	defaultFDThresholdMs        = 10.0
	defaultFDVThresholdMs       = 5.0
	defaultFLRThresholdPct      = 0.01
	defaultPerfDurationSec      = 900 // 15 minutes
	defaultMEFConfigDurationSec = 60
	defaultMEFPerfDurationMin   = 15
	secondsPerMinute            = 60
	defaultAvailabilityPct      = 99.99
	microsecondsPerMillisecond  = 1000.0
	maxUint32                   = 4294967295
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

	if e.ctx == nil {
		return nil, errors.New("dataplane context is not configured")
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
	case "y1564_config", "y1564_perf", "y1564":
		data, runErr = e.runY1564(testType, cfg)
	case "mef_config", "mef_perf", "mef":
		data, runErr = e.runMEF(testType, cfg)
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

func (e *Executor) runY1564(testType string, cfg *modtypes.TestConfig) (any, error) {
	switch testType {
	case "y1564_config":
		service := e.buildY1564Service(cfg)
		data, err := e.ctx.RunY1564ConfigTest(service)
		if err != nil {
			return nil, fmt.Errorf("y1564 config test: %w", err)
		}
		return data, nil
	case "y1564_perf":
		service := e.buildY1564Service(cfg)
		duration := e.safeDuration(cfg.Duration, defaultPerfDurationSec)
		data, err := e.ctx.RunY1564PerfTest(service, duration)
		if err != nil {
			return nil, fmt.Errorf("y1564 perf test: %w", err)
		}
		return data, nil
	case "y1564":
		service := e.buildY1564Service(cfg)
		configResult, configErr := e.ctx.RunY1564ConfigTest(service)
		if configErr != nil {
			return nil, fmt.Errorf("y1564 config test: %w", configErr)
		}

		duration := e.safeDuration(cfg.Duration, defaultPerfDurationSec)
		perfResult, perfErr := e.ctx.RunY1564PerfTest(service, duration)
		if perfErr != nil {
			return nil, fmt.Errorf("y1564 perf test: %w", perfErr)
		}

		return map[string]any{
			"config":      configResult,
			"performance": perfResult,
		}, nil
	default:
		return nil, modtypes.ErrTestNotImplemented
	}
}

func (e *Executor) runMEF(testType string, cfg *modtypes.TestConfig) (any, error) {
	mefConfig := e.buildMEFConfig(cfg)

	switch testType {
	case "mef_config":
		data, err := e.ctx.RunMEFConfigTest(mefConfig)
		if err != nil {
			return nil, fmt.Errorf("mef config test: %w", err)
		}
		return data, nil
	case "mef_perf":
		data, err := e.ctx.RunMEFPerfTest(mefConfig)
		if err != nil {
			return nil, fmt.Errorf("mef performance test: %w", err)
		}
		return data, nil
	case "mef":
		configResult, perfResult, runErr := e.ctx.RunMEFFullTest(mefConfig)
		if runErr != nil {
			return nil, fmt.Errorf("mef full test: %w", runErr)
		}
		return map[string]any{
			"config":      configResult,
			"performance": perfResult,
		}, nil
	default:
		return nil, modtypes.ErrTestNotImplemented
	}
}

// configureContext sets up the dataplane context from test config.
func (e *Executor) configureContext(cfg *modtypes.TestConfig) error {
	if e.ctx == nil {
		return errors.New("dataplane context is not configured")
	}

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

func (e *Executor) buildMEFConfig(cfg *modtypes.TestConfig) *dataplane.MEFConfig {
	mefConfig := &dataplane.MEFConfig{
		ServiceID: "",
		CIRMbps:   modtypes.GetFloat64Param(cfg.Params, "cir", defaultCIRMbps),
		EIRMbps:   modtypes.GetFloat64Param(cfg.Params, "eir", defaultEIRMbps),
		CBSBytes:  modtypes.GetUint32Param(cfg.Params, "cbs", 0),
		EBSBytes:  modtypes.GetUint32Param(cfg.Params, "ebs", 0),
		FDThresholdUs: modtypes.GetFloat64Param(
			cfg.Params,
			"fd_threshold_us",
			defaultFDThresholdMs*microsecondsPerMillisecond,
		),
		FDVThresholdUs: modtypes.GetFloat64Param(
			cfg.Params,
			"fdv_threshold_us",
			defaultFDVThresholdMs*microsecondsPerMillisecond,
		),
		FLRThresholdPct: modtypes.GetFloat64Param(
			cfg.Params,
			"flr_threshold_pct",
			defaultFLRThresholdPct,
		),
		AvailabilityPct: modtypes.GetFloat64Param(
			cfg.Params,
			"availability_pct",
			defaultAvailabilityPct,
		),
		ConfigDurationSec: modtypes.GetUint32Param(
			cfg.Params,
			"config_duration_sec",
			defaultMEFConfigDurationSec,
		),
		PerfDurationMin: modtypes.GetUint32Param(
			cfg.Params,
			"perf_duration_min",
			defaultMEFPerfDurationMin,
		),
		CoS:        modtypes.GetUint32Param(cfg.Params, "cos", 0),
		FrameSizes: nil,
	}

	if cfg.Duration > 0 {
		converted := modtypes.SafeIntToUint32(cfg.Duration / secondsPerMinute)
		if converted == 0 {
			converted = modtypes.SafeIntToUint32(cfg.Duration)
		}
		if converted > 0 {
			mefConfig.PerfDurationMin = converted
		}
	}

	if cfg.FrameSize > 0 {
		mefConfig.FrameSizes = []uint32{cfg.FrameSize}
	}

	if serviceID, ok := cfg.Params["service_id"].(string); ok {
		mefConfig.ServiceID = serviceID
	}

	return mefConfig
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
	if _, ok := cfg.Params["cbs"]; ok {
		service.SLA.CBSBytes = modtypes.GetUint32Param(cfg.Params, "cbs", service.SLA.CBSBytes)
	}
	if _, ok := cfg.Params["ebs"]; ok {
		service.SLA.EBSBytes = modtypes.GetUint32Param(cfg.Params, "ebs", service.SLA.EBSBytes)
	}
	if _, ok := cfg.Params["fd_threshold_ms"]; ok {
		service.SLA.FDThresholdMs = modtypes.GetFloat64Param(
			cfg.Params,
			"fd_threshold_ms",
			service.SLA.FDThresholdMs,
		)
	}
	if _, ok := cfg.Params["fdv_threshold_ms"]; ok {
		service.SLA.FDVThresholdMs = modtypes.GetFloat64Param(
			cfg.Params,
			"fdv_threshold_ms",
			service.SLA.FDVThresholdMs,
		)
	}
	if _, ok := cfg.Params["flr_threshold_pct"]; ok {
		service.SLA.FLRThresholdPct = modtypes.GetFloat64Param(
			cfg.Params,
			"flr_threshold_pct",
			service.SLA.FLRThresholdPct,
		)
	}

	// Service-specific parameters
	if _, ok := cfg.Params["frame_size"]; ok {
		service.FrameSize = modtypes.GetUint32Param(cfg.Params, "frame_size", service.FrameSize)
	}
	if _, ok := cfg.Params["cos"]; ok {
		service.CoS = modtypes.GetUint8Param(cfg.Params, "cos", service.CoS)
	}
	if val, ok := cfg.Params["enabled"]; ok {
		if enabled, okBool := val.(bool); okBool {
			service.Enabled = enabled
		}
	}
}

// safeDuration returns duration as uint32 seconds, clamping to max if needed.
func (e *Executor) safeDuration(duration int, fallback uint32) uint32 {
	if duration <= 0 {
		return fallback
	}
	converted := modtypes.SafeIntToUint32(duration)
	if converted == 0 {
		return fallback
	}
	return converted
}
