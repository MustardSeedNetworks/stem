// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package servicetest

import (
	"fmt"
	"time"

	"github.com/krisarmstrong/stem/internal/modules/common"
	"github.com/krisarmstrong/stem/internal/testmaster/dataplane"
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
func (e *Executor) Execute(testType string, cfg *common.TestConfig) (*common.Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("servicetest module cannot run test type: %s", testType)
	}

	if cfg == nil {
		return nil, common.ErrInvalidConfig
	}

	// Configure the context
	if err := e.configureContext(cfg); err != nil {
		return nil, fmt.Errorf("failed to configure context: %w", err)
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
	case "y1564_config":
		service := e.buildY1564Service(cfg)
		data, err = e.ctx.RunY1564ConfigTest(service)

	case "y1564_perf":
		service := e.buildY1564Service(cfg)
		duration := e.safeDuration(cfg.Duration, 900) // 15 minutes default
		data, err = e.ctx.RunY1564PerfTest(service, duration)

	case "y1564":
		// Full Y.1564 test: config + performance
		service := e.buildY1564Service(cfg)

		// Run config test first
		configResult, configErr := e.ctx.RunY1564ConfigTest(service)
		if configErr != nil {
			result.Error = fmt.Sprintf("config test failed: %v", configErr)
			return result, fmt.Errorf("y1564 config test: %w", configErr)
		}

		// Run performance test
		duration := e.safeDuration(cfg.Duration, 900) // 15 minutes default
		perfResult, perfErr := e.ctx.RunY1564PerfTest(service, duration)
		if perfErr != nil {
			result.Error = fmt.Sprintf("perf test failed: %v", perfErr)
			return result, fmt.Errorf("y1564 perf test: %w", perfErr)
		}

		// Combine results
		data = map[string]interface{}{
			"config":      configResult,
			"performance": perfResult,
		}

	case "mef_config", "mef_perf", "mef":
		// MEF tests are similar to Y.1564, but not yet implemented in dataplane
		return nil, common.ErrTestNotImplemented

	default:
		return nil, common.ErrTestNotImplemented
	}

	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("servicetest %s failed: %w", testType, err)
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

	if err := e.ctx.Configure(dpCfg); err != nil {
		return fmt.Errorf("configure dataplane: %w", err)
	}
	return nil
}

// buildY1564Service creates a Y1564Service from the test config.
func (e *Executor) buildY1564Service(cfg *common.TestConfig) *dataplane.Y1564Service {
	service := &dataplane.Y1564Service{
		ServiceID:   1,
		ServiceName: "Service-1",
		FrameSize:   1518,
		Enabled:     true,
		SLA: dataplane.Y1564SLA{
			CIRMbps:         100,
			EIRMbps:         0,
			FDThresholdMs:   10,
			FDVThresholdMs:  5,
			FLRThresholdPct: 0.01,
		},
	}

	if cfg.FrameSize > 0 {
		service.FrameSize = cfg.FrameSize
	}

	// Extract SLA and service parameters from config
	e.extractY1564Params(cfg, service)

	return service
}

// extractY1564Params extracts SLA and service parameters from config using type-safe helpers.
func (e *Executor) extractY1564Params(cfg *common.TestConfig, service *dataplane.Y1564Service) {
	if cfg.Params == nil {
		return
	}

	// Extract SLA parameters using type-safe helper
	// Only update if parameter is explicitly set (check existence first)
	if _, ok := cfg.Params["cir"]; ok {
		service.SLA.CIRMbps = common.GetFloat64Param(cfg.Params, "cir", service.SLA.CIRMbps)
	}
	if _, ok := cfg.Params["eir"]; ok {
		service.SLA.EIRMbps = common.GetFloat64Param(cfg.Params, "eir", service.SLA.EIRMbps)
	}
	if _, ok := cfg.Params["fd_threshold"]; ok {
		service.SLA.FDThresholdMs = common.GetFloat64Param(cfg.Params, "fd_threshold", service.SLA.FDThresholdMs)
	}
	if _, ok := cfg.Params["fdv_threshold"]; ok {
		service.SLA.FDVThresholdMs = common.GetFloat64Param(cfg.Params, "fdv_threshold", service.SLA.FDVThresholdMs)
	}
	if _, ok := cfg.Params["flr_threshold"]; ok {
		service.SLA.FLRThresholdPct = common.GetFloat64Param(cfg.Params, "flr_threshold", service.SLA.FLRThresholdPct)
	}

	// Extract service identification
	if name, ok := cfg.Params["service_name"].(string); ok {
		service.ServiceName = name
	}
	if v, ok := cfg.Params["service_id"]; ok {
		// Handle both uint32 and float64 (from JSON)
		switch id := v.(type) {
		case uint32:
			service.ServiceID = id
		case float64:
			if id >= 0 && id <= float64(^uint32(0)) {
				service.ServiceID = uint32(id)
			}
		case int:
			const maxUint32 = 4294967295 // math.MaxUint32
			if id >= 0 && id <= maxUint32 {
				service.ServiceID = uint32(id) // Safe: validated above
			}
		}
	}
}

// safeDuration converts an int duration to uint32 safely.
// Returns defaultVal if duration is <= 0 or would overflow uint32.
func (e *Executor) safeDuration(duration int, defaultVal uint32) uint32 {
	const maxUint32 = 4294967295 // math.MaxUint32
	if duration <= 0 || duration > maxUint32 {
		return defaultVal
	}
	return uint32(duration)
}
