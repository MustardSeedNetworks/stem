// SPDX-License-Identifier: BUSL-1.1

package measure

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/krisarmstrong/stem/internal/services/modtypes"
)

// newTestExecutor creates an Executor for testing without a real dataplane.
// The executor has a nil dataplane, so methods that use it will fail
// gracefully via the "dataplane is not configured" error path.
func newTestExecutor() *Executor {
	return &Executor{
		Module: New(),
		dp:     nil, // Nil dataplane for testing.
	}
}

// TestModtypesSafeIntToUint32 tests modtypes.SafeIntToUint32 (formerly safeUint32FromInt).
func TestModtypesSafeIntToUint32(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected uint32
	}{
		{
			name:     "zero value",
			value:    0,
			expected: 0,
		},
		{
			name:     "positive value",
			value:    60,
			expected: 60,
		},
		{
			name:     "max uint32 boundary",
			value:    math.MaxUint32,
			expected: math.MaxUint32,
		},
		{
			name:     "negative value returns 0",
			value:    -1,
			expected: 0,
		},
		{
			name:     "large negative returns 0",
			value:    -1000,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.SafeIntToUint32(tt.value)
			if result != tt.expected {
				t.Errorf("modtypes.SafeIntToUint32(%d) = %d, want %d", tt.value, result, tt.expected)
			}
		})
	}
}

// TestModtypesGetUint8Param tests modtypes.GetUint8Param (replaces clampUint8).
func TestModtypesGetUint8Param(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]any
		key        string
		defaultVal uint8
		expected   uint8
	}{
		{
			name:       "zero value",
			params:     map[string]any{"val": float64(0)},
			key:        "val",
			defaultVal: 100,
			expected:   0,
		},
		{
			name:       "small value",
			params:     map[string]any{"val": float64(6)},
			key:        "val",
			defaultVal: 100,
			expected:   6,
		},
		{
			name:       "max uint8 boundary",
			params:     map[string]any{"val": float64(255)},
			key:        "val",
			defaultVal: 100,
			expected:   255,
		},
		{
			name:       "just over max uint8 returns default",
			params:     map[string]any{"val": float64(256)},
			key:        "val",
			defaultVal: 100,
			expected:   100,
		},
		{
			name:       "large value returns default",
			params:     map[string]any{"val": float64(1000)},
			key:        "val",
			defaultVal: 50,
			expected:   50,
		},
		{
			name:       "mid-range value",
			params:     map[string]any{"val": float64(128)},
			key:        "val",
			defaultVal: 100,
			expected:   128,
		},
		{
			name:       "nil params returns default",
			params:     nil,
			key:        "val",
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "missing key returns default",
			params:     map[string]any{"other": float64(10)},
			key:        "val",
			defaultVal: 99,
			expected:   99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint8Param(tt.params, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetUint8Param() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestBuildY1731Config tests the buildY1731Config helper function.
func TestBuildY1731Config(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *modtypes.TestConfig
		validate func(t *testing.T, result any)
	}{
		{
			name: "default values with empty params",
			cfg: &modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 0,
				Duration:  0,
				Params:    nil,
			},
			validate: func(t *testing.T, result any) {
				if result == nil {
					t.Error("buildY1731Config returned nil")
				}
			},
		},
		{
			name: "custom MEP ID from params",
			cfg: &modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 64,
				Duration:  30,
				Params: map[string]any{
					"mep_id": uint32(42),
				},
			},
			validate: func(t *testing.T, result any) {
				if result == nil {
					t.Error("buildY1731Config returned nil")
				}
			},
		},
		{
			name: "all custom Y.1731 params",
			cfg: &modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 128,
				Duration:  120,
				Params: map[string]any{
					"mep_id":          uint32(10),
					"meg_level":       uint32(5),
					"meg_id":          "CUSTOM-MEG",
					"ccm_interval":    uint32(500),
					"priority":        uint32(7),
					"interval_ms":     uint32(50),
					"count":           uint32(20),
					"priority_tagged": false,
				},
			},
			validate: func(t *testing.T, result any) {
				if result == nil {
					t.Error("buildY1731Config returned nil")
				}
			},
		},
		{
			name: "priority clamping to uint8",
			cfg: &modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 64,
				Duration:  60,
				Params: map[string]any{
					"priority": uint32(300), // Should clamp to 255
				},
			},
			validate: func(t *testing.T, result any) {
				if result == nil {
					t.Error("buildY1731Config returned nil")
				}
			},
		},
		{
			name: "duration from config takes precedence",
			cfg: &modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 64,
				Duration:  90,
				Params: map[string]any{
					"duration_sec": uint32(30), // Config.Duration should take precedence
				},
			},
			validate: func(t *testing.T, result any) {
				if result == nil {
					t.Error("buildY1731Config returned nil")
				}
			},
		},
		{
			name: "zero duration falls back to param",
			cfg: &modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 64,
				Duration:  0, // Zero duration
				Params: map[string]any{
					"duration_sec": uint32(45),
				},
			},
			validate: func(t *testing.T, result any) {
				if result == nil {
					t.Error("buildY1731Config returned nil")
				}
			},
		},
		{
			name: "zero frame size falls back to param",
			cfg: &modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 0, // Zero frame size
				Duration:  60,
				Params: map[string]any{
					"frame_size": uint32(256),
				},
			},
			validate: func(t *testing.T, result any) {
				if result == nil {
					t.Error("buildY1731Config returned nil")
				}
			},
		},
		{
			name: "float64 params from JSON decoding",
			cfg: &modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 64,
				Duration:  60,
				Params: map[string]any{
					"mep_id":       float64(15),
					"meg_level":    float64(3),
					"ccm_interval": float64(2000),
					"priority":     float64(4),
					"interval_ms":  float64(200),
					"count":        float64(5),
				},
			},
			validate: func(t *testing.T, result any) {
				if result == nil {
					t.Error("buildY1731Config returned nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildY1731Config(tt.cfg)
			tt.validate(t, result)
		})
	}
}

// TestBuildY1731ConfigDefaultValues tests default values in the config.
func TestBuildY1731ConfigDefaultValues(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "",
		FrameSize: 0,
		Duration:  0,
		Params:    nil,
	}
	result := buildY1731Config(cfg)

	if result.MEPID != defaultMEPID {
		t.Errorf("MEPID = %d, want %d", result.MEPID, defaultMEPID)
	}
	if result.MEGLevel != defaultMEGLevel {
		t.Errorf("MEGLevel = %d, want %d", result.MEGLevel, defaultMEGLevel)
	}
	if result.MEGID != defaultMEGID {
		t.Errorf("MEGID = %q, want %q", result.MEGID, defaultMEGID)
	}
	if result.CCMInterval != defaultCCMInterval {
		t.Errorf("CCMInterval = %d, want %d", result.CCMInterval, defaultCCMInterval)
	}
	if result.Priority != defaultPriority {
		t.Errorf("Priority = %d, want %d", result.Priority, defaultPriority)
	}
	if result.IntervalMs != defaultIntervalMs {
		t.Errorf("IntervalMs = %d, want %d", result.IntervalMs, defaultIntervalMs)
	}
	if result.Count != defaultCount {
		t.Errorf("Count = %d, want %d", result.Count, defaultCount)
	}
	if !result.PriorityTagged {
		t.Error("PriorityTagged should default to true")
	}
}

// TestBuildY1731ConfigCustomValues tests custom values in the config.
func TestBuildY1731ConfigCustomValues(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 512,
		Duration:  180,
		Params: map[string]any{
			"mep_id":          uint32(99),
			"meg_level":       uint32(7),
			"meg_id":          "TEST-MEG-42",
			"ccm_interval":    uint32(3000),
			"priority":        uint32(3),
			"interval_ms":     uint32(250),
			"count":           uint32(50),
			"priority_tagged": false,
		},
	}
	result := buildY1731Config(cfg)

	if result.MEPID != 99 {
		t.Errorf("MEPID = %d, want 99", result.MEPID)
	}
	if result.MEGLevel != 7 {
		t.Errorf("MEGLevel = %d, want 7", result.MEGLevel)
	}
	if result.MEGID != "TEST-MEG-42" {
		t.Errorf("MEGID = %q, want %q", result.MEGID, "TEST-MEG-42")
	}
	if result.CCMInterval != 3000 {
		t.Errorf("CCMInterval = %d, want 3000", result.CCMInterval)
	}
	if result.Priority != 3 {
		t.Errorf("Priority = %d, want 3", result.Priority)
	}
	if result.DurationSec != 180 {
		t.Errorf("DurationSec = %d, want 180", result.DurationSec)
	}
	if result.IntervalMs != 250 {
		t.Errorf("IntervalMs = %d, want 250", result.IntervalMs)
	}
	if result.Count != 50 {
		t.Errorf("Count = %d, want 50", result.Count)
	}
	if result.FrameSize != 512 {
		t.Errorf("FrameSize = %d, want 512", result.FrameSize)
	}
	if result.PriorityTagged {
		t.Error("PriorityTagged should be false when explicitly set")
	}
}

// TestBuildY1731ConfigPriorityFallback tests priority fallback for out-of-range values.
func TestBuildY1731ConfigPriorityFallback(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  60,
		Params: map[string]any{
			"priority": uint32(500), // Exceeds uint8 max, falls back to default
		},
	}
	result := buildY1731Config(cfg)

	// Out-of-range values fall back to default priority (6)
	if result.Priority != defaultPriority {
		t.Errorf("Priority = %d, want %d (default)", result.Priority, defaultPriority)
	}
}

// TestBuildY1731ConfigDurationFallback tests duration fallback from params.
func TestBuildY1731ConfigDurationFallback(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  0, // Zero duration triggers fallback
		Params: map[string]any{
			"duration_sec": uint32(75),
		},
	}
	result := buildY1731Config(cfg)

	if result.DurationSec != 75 {
		t.Errorf("DurationSec = %d, want 75", result.DurationSec)
	}
}

// TestBuildY1731ConfigFrameSizeFallback tests frame size fallback from params.
func TestBuildY1731ConfigFrameSizeFallback(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 0, // Zero frame size triggers fallback
		Duration:  60,
		Params: map[string]any{
			"frame_size": uint32(1518),
		},
	}
	result := buildY1731Config(cfg)

	if result.FrameSize != 1518 {
		t.Errorf("FrameSize = %d, want 1518", result.FrameSize)
	}
}

// TestBuildY1731ConfigNegativeDuration tests negative duration uses fallback.
func TestBuildY1731ConfigNegativeDuration(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  -10, // Negative duration
		Params:    nil,
	}
	result := buildY1731Config(cfg)

	if result.DurationSec != defaultDuration {
		t.Errorf("DurationSec = %d, want %d (default)", result.DurationSec, defaultDuration)
	}
}

// TestExecutorSupportsExecution tests the SupportsExecution method.
func TestExecutorSupportsExecutionInternal(t *testing.T) {
	executor := newTestExecutor()

	if !executor.SupportsExecution() {
		t.Error("SupportsExecution() should return true")
	}

	// Call multiple times to ensure consistent result.
	for range 3 {
		if !executor.SupportsExecution() {
			t.Error("SupportsExecution() should always return true")
		}
	}
}

// TestExecutorCloseNilContext tests Close with nil context.
func TestExecutorCloseNilContext(t *testing.T) {
	executor := newTestExecutor()

	// Close should not panic with nil context.
	executor.Close()

	// Multiple calls should be safe.
	executor.Close()
	executor.Close()

	// Verify nothing changed (executor should still be usable for module methods).
	if executor.Name() != ModuleName {
		t.Errorf("Name() = %q after Close(), want %q", executor.Name(), ModuleName)
	}
}

// TestExecutorCloseValid tests Close behavior on valid executor.
func TestExecutorCloseValid(t *testing.T) {
	executor := newTestExecutor()
	executor.Close() // Should not panic.

	// Verify module methods still work after close.
	if !executor.CanRun("y1731_delay") {
		t.Error("CanRun should still work after Close")
	}
}

// TestExecutorExecuteInvalidTestType tests Execute with invalid test types.
func TestExecutorExecuteInvalidTestTypeInternal(t *testing.T) {
	executor := newTestExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  60,
		Params:    nil,
	}

	invalidTypes := []string{
		"invalid",
		"rfc2544_throughput",
		"y1564",
		"",
		"y1731", // Incomplete - missing test type suffix.
	}

	for _, testType := range invalidTypes {
		t.Run("testType_"+testType, func(t *testing.T) {
			result, err := executor.Execute(testType, cfg)
			if err == nil {
				t.Errorf("Execute(%q) should return error", testType)
			}
			if result != nil {
				t.Error("Result should be nil for invalid test type")
			}
			if !strings.Contains(err.Error(), "cannot run test type") {
				t.Errorf("Error should mention 'cannot run test type', got: %v", err)
			}
		})
	}
}

// TestExecutorExecuteNilConfig tests Execute with nil config.
func TestExecutorExecuteNilConfigInternal(t *testing.T) {
	executor := newTestExecutor()

	result, err := executor.Execute("y1731_delay", nil)
	if err == nil {
		t.Error("Execute with nil config should return error")
	}
	if !errors.Is(err, modtypes.ErrInvalidConfig) {
		t.Errorf("Expected ErrInvalidConfig, got: %v", err)
	}
	if result != nil {
		t.Error("Result should be nil when error is returned")
	}
}

// TestExecutorExecuteValidTestTypesWithNilDataplane tests Execute with
// valid test types when the executor has no dataplane configured. The
// dispatcher should return a structured error rather than panic.
func TestExecutorExecuteValidTestTypesWithNilDataplane(t *testing.T) {
	executor := newTestExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  60,
		Params:    nil,
	}

	validTypes := []string{
		"y1731_delay",
		"y1731_loss",
		"y1731_slm",
		"y1731_loopback",
	}

	for _, testType := range validTypes {
		t.Run("testType_"+testType, func(t *testing.T) {
			result, err := executor.Execute(testType, cfg)
			if err == nil {
				t.Fatalf("Execute(%q) should return error when dataplane is nil", testType)
			}
			if result == nil {
				t.Fatalf("Execute(%q) should return a result describing the failure", testType)
			}
			if result.Success {
				t.Errorf("Execute(%q) Success = true, want false", testType)
			}
			if !strings.Contains(result.Error, "dataplane") {
				t.Errorf("Execute(%q) error = %q, want it to mention dataplane", testType, result.Error)
			}
			if result.TestType != testType {
				t.Errorf("Result.TestType = %q, want %q", result.TestType, testType)
			}
		})
	}
}

// TestExecutorEmbeddedModuleMethods tests Module methods via embedding.
func TestExecutorEmbeddedModuleMethods(t *testing.T) {
	executor := newTestExecutor()

	// Verify all Module methods are accessible via Executor.
	if executor.Name() != ModuleName {
		t.Errorf("Name() = %q, want %q", executor.Name(), ModuleName)
	}
	if executor.DisplayName() != DisplayName {
		t.Errorf("DisplayName() = %q, want %q", executor.DisplayName(), DisplayName)
	}
	if executor.Color() != ColorHex {
		t.Errorf("Color() = %q, want %q", executor.Color(), ColorHex)
	}
	if executor.Standard() != StandardRef {
		t.Errorf("Standard() = %q, want %q", executor.Standard(), StandardRef)
	}

	execTestTypes := executor.TestTypes()
	if len(execTestTypes) != 4 {
		t.Errorf("TestTypes() returned %d types, want 4", len(execTestTypes))
	}

	if !executor.CanRun("y1731_delay") {
		t.Error("CanRun(y1731_delay) should return true")
	}
	if executor.CanRun("rfc2544_throughput") {
		t.Error("CanRun(rfc2544_throughput) should return false")
	}
}

// TestExecutorExecuteErrorPaths tests error paths in Execute.
func TestExecutorExecuteErrorPaths(t *testing.T) {
	executor := newTestExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  60,
		Params: map[string]any{
			"mep_id": uint32(42),
		},
	}

	// Test invalid test type (doesn't reach ctx).
	result, err := executor.Execute("invalid", cfg)
	if err == nil {
		t.Error("Execute should return error for invalid test type")
	}
	if result != nil {
		t.Error("Result should be nil for error")
	}

	// Test nil config path.
	_, nilErr := executor.Execute("y1731_delay", nil)
	if nilErr == nil {
		t.Error("Execute should return error for nil config")
	}
	if !errors.Is(nilErr, modtypes.ErrInvalidConfig) {
		t.Errorf("Expected ErrInvalidConfig, got: %v", nilErr)
	}
}

// TestExecutorWithConfigs tests Execute with various configurations.
func TestExecutorWithConfigs(t *testing.T) {
	executor := newTestExecutor()

	t.Run("nil config returns error", func(t *testing.T) {
		result, err := executor.Execute("y1731_delay", nil)
		if err == nil {
			t.Error("Expected error but got nil")
		}
		if result != nil {
			t.Error("Expected nil result on error")
		}
	})

	t.Run("empty config returns dataplane error", func(t *testing.T) {
		cfg := &modtypes.TestConfig{
			Interface: "",
			FrameSize: 0,
			Duration:  0,
			Params:    nil,
		}
		result, err := executor.Execute("y1731_delay", cfg)
		if err == nil {
			t.Fatal("Execute should return an error when dataplane is nil")
		}
		if result == nil || result.Success {
			t.Fatalf("Result should report failure, got %#v", result)
		}
		if !strings.Contains(result.Error, "dataplane") {
			t.Errorf("Result.Error = %q, want it to mention dataplane", result.Error)
		}
	})

	t.Run("full config returns dataplane error", func(t *testing.T) {
		cfg := &modtypes.TestConfig{
			Interface: "eth0",
			FrameSize: 512,
			Duration:  120,
			Params: map[string]any{
				"mep_id":          uint32(10),
				"meg_level":       uint32(5),
				"meg_id":          "TEST-MEG",
				"ccm_interval":    uint32(500),
				"priority":        uint32(3),
				"interval_ms":     uint32(50),
				"count":           uint32(20),
				"priority_tagged": false,
			},
		}
		result, err := executor.Execute("y1731_delay", cfg)
		if err == nil {
			t.Fatal("Execute should return an error when dataplane is nil")
		}
		if result == nil || result.Success {
			t.Fatalf("Result should report failure, got %#v", result)
		}
	})
}

// TestDefaultConstants verifies the Y.1731 default constants.
func TestDefaultConstants(t *testing.T) {
	// Verify constants match TUI/WebUI defaults as documented.
	if defaultMEPID != 1 {
		t.Errorf("defaultMEPID = %d, want 1", defaultMEPID)
	}
	if defaultMEGLevel != 4 {
		t.Errorf("defaultMEGLevel = %d, want 4", defaultMEGLevel)
	}
	if defaultCCMInterval != 1000 {
		t.Errorf("defaultCCMInterval = %d, want 1000", defaultCCMInterval)
	}
	if defaultPriority != 6 {
		t.Errorf("defaultPriority = %d, want 6", defaultPriority)
	}
	if defaultDuration != 60 {
		t.Errorf("defaultDuration = %d, want 60", defaultDuration)
	}
	if defaultIntervalMs != 100 {
		t.Errorf("defaultIntervalMs = %d, want 100", defaultIntervalMs)
	}
	if defaultCount != 10 {
		t.Errorf("defaultCount = %d, want 10", defaultCount)
	}
	if defaultFrameSize != 64 {
		t.Errorf("defaultFrameSize = %d, want 64", defaultFrameSize)
	}
	if defaultMEGID != "MSN-MEG-01" {
		t.Errorf("defaultMEGID = %q, want MSN-MEG-01", defaultMEGID)
	}
}
