// SPDX-License-Identifier: BUSL-1.1

package reflector

import (
	"fmt"

	reflectorConfig "github.com/krisarmstrong/stem/internal/reflector/config"
	reflectorDP "github.com/krisarmstrong/stem/internal/reflector/dataplane"
)

// Result is a generic operation result.
type Result struct {
	TestType   string `json:"testType"`
	ModuleName string `json:"module"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Data       any    `json:"data,omitempty"`
}

// Config holds configuration for reflector operation.
type Config struct {
	Interface string
	Profile   string // netally, msn, all, custom
	Params    map[string]any
}

// Executor wraps the Reflector module with execution capability.
type Executor struct {
	*Module

	dp  *reflectorDP.Dataplane
	cfg *reflectorConfig.Config
}

// NewExecutor creates a new Reflector executor.
func NewExecutor(iface string) (*Executor, error) {
	cfg := &reflectorConfig.Config{
		Interface:       iface,
		Verbose:         false,
		SignatureFilter: "all",
		WebUI:           reflectorConfig.WebUIConfig{Enabled: false, Port: 0},
		TUI:             reflectorConfig.TUIConfig{Enabled: false},
		Filtering:       reflectorConfig.FilterConfig{Port: 0, FilterOUI: false, OUI: "", FilterMAC: false},
		Reflection: reflectorConfig.ReflectConfig{
			Mode: "all",
		},
		Platform: reflectorConfig.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
		Stats:    reflectorConfig.StatsConfig{Format: "", Interval: 0},
	}

	dp, err := reflectorDP.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create reflector dataplane: %w", err)
	}

	return &Executor{
		Module: New(),
		dp:     dp,
		cfg:    cfg,
	}, nil
}

// NewExecutorWithDataplane creates an executor with an existing dataplane.
func NewExecutorWithDataplane(dp *reflectorDP.Dataplane) *Executor {
	return &Executor{
		Module: New(),
		dp:     dp,
		cfg:    nil,
	}
}

// SupportsExecution returns true as Reflector supports execution.
func (e *Executor) SupportsExecution() bool {
	return true
}

// Close releases the dataplane resources.
func (e *Executor) Close() {
	if e.dp != nil {
		e.dp.Close()
	}
}

// Execute runs the reflector operation.
func (e *Executor) Execute(testType string, _ *Config) (*Result, error) {
	if !e.CanRun(testType) {
		return nil, fmt.Errorf("reflector module cannot run operation type: %s", testType)
	}

	result := &Result{
		TestType:   testType,
		ModuleName: ModuleName,
		Success:    false,
		Error:      "",
		Data:       nil,
	}

	switch testType {
	case "reflect":
		if e.dp == nil {
			result.Error = "reflector dataplane is not configured"
			return result, fmt.Errorf("start reflector: %s", result.Error)
		}

		// Start reflector mode.
		err := e.dp.Start()
		if err != nil {
			result.Error = err.Error()
			return result, fmt.Errorf("start reflector: %w", err)
		}

		// Get initial stats.
		stats := e.dp.GetStats()

		result.Success = true
		result.Data = map[string]any{
			"status":  "running",
			"message": "Reflector started successfully",
			"stats":   stats,
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unknown reflector operation: %s", testType)
	}
}

// Stop stops the reflector if running.
func (e *Executor) Stop() {
	if e.dp != nil {
		e.dp.Stop()
	}
}

// GetStats returns the current reflector statistics.
func (e *Executor) GetStats() reflectorDP.Stats {
	if e.dp != nil {
		return e.dp.GetStats()
	}
	return reflectorDP.Stats{}
}

// IsRunning returns true if the reflector is currently running.
func (e *Executor) IsRunning() bool {
	if e.dp != nil {
		return e.dp.IsRunning()
	}
	return false
}

// Dataplane returns the underlying dataplane for direct access.
// Returns nil if no dataplane is configured.
func (e *Executor) Dataplane() *reflectorDP.Dataplane {
	return e.dp
}
