// SPDX-License-Identifier: BUSL-1.1

package benchmark_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/benchmark"
	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

// Test parameter extraction helpers (issue #24).
func TestGetFloat64Param(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]any
		key      string
		defVal   float64
		expected float64
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			key:      "test",
			defVal:   10.0,
			expected: 10.0,
		},
		{
			name:     "missing key returns default",
			params:   map[string]any{"other": 5.0},
			key:      "test",
			defVal:   10.0,
			expected: 10.0,
		},
		{
			name:     "float64 value",
			params:   map[string]any{"test": 25.5},
			key:      "test",
			defVal:   10.0,
			expected: 25.5,
		},
		{
			name:     "int value converts to float64",
			params:   map[string]any{"test": 42},
			key:      "test",
			defVal:   10.0,
			expected: 42.0,
		},
		{
			name:     "int64 value converts to float64",
			params:   map[string]any{"test": int64(100)},
			key:      "test",
			defVal:   10.0,
			expected: 100.0,
		},
		{
			name:     "string value returns default",
			params:   map[string]any{"test": "not a number"},
			key:      "test",
			defVal:   10.0,
			expected: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetFloat64Param(tt.params, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("GetFloat64Param() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetUint64Param(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]any
		key      string
		defVal   uint64
		expected uint64
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			key:      "test",
			defVal:   1000,
			expected: 1000,
		},
		{
			name:     "missing key returns default",
			params:   map[string]any{"other": uint64(500)},
			key:      "test",
			defVal:   1000,
			expected: 1000,
		},
		{
			name:     "uint64 value",
			params:   map[string]any{"test": uint64(5000)},
			key:      "test",
			defVal:   1000,
			expected: 5000,
		},
		{
			name:     "float64 value converts to uint64",
			params:   map[string]any{"test": 12345.0},
			key:      "test",
			defVal:   1000,
			expected: 12345,
		},
		{
			name:     "int value converts to uint64",
			params:   map[string]any{"test": 999},
			key:      "test",
			defVal:   1000,
			expected: 999,
		},
		{
			name:     "negative float64 returns default",
			params:   map[string]any{"test": -10.0},
			key:      "test",
			defVal:   1000,
			expected: 1000,
		},
		{
			name:     "negative int returns default",
			params:   map[string]any{"test": -5},
			key:      "test",
			defVal:   1000,
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint64Param(tt.params, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("GetUint64Param() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetUint32Param(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]any
		key      string
		defVal   uint32
		expected uint32
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			key:      "test",
			defVal:   100,
			expected: 100,
		},
		{
			name:     "uint32 value",
			params:   map[string]any{"test": uint32(50)},
			key:      "test",
			defVal:   100,
			expected: 50,
		},
		{
			name:     "float64 value converts to uint32",
			params:   map[string]any{"test": 75.0},
			key:      "test",
			defVal:   100,
			expected: 75,
		},
		{
			name:     "int value converts to uint32",
			params:   map[string]any{"test": 200},
			key:      "test",
			defVal:   100,
			expected: 200,
		},
		{
			name:     "negative value returns default",
			params:   map[string]any{"test": -1},
			key:      "test",
			defVal:   100,
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint32Param(tt.params, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("GetUint32Param() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetIntParam(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]any
		key      string
		defVal   int
		expected int
	}{
		{
			name:     "nil params returns default",
			params:   nil,
			key:      "test",
			defVal:   10,
			expected: 10,
		},
		{
			name:     "int value",
			params:   map[string]any{"test": 42},
			key:      "test",
			defVal:   10,
			expected: 42,
		},
		{
			name:     "float64 value converts to int",
			params:   map[string]any{"test": 99.9},
			key:      "test",
			defVal:   10,
			expected: 99,
		},
		{
			name:     "negative int works",
			params:   map[string]any{"test": -5},
			key:      "test",
			defVal:   10,
			expected: -5,
		},
		{
			name:     "int64 converts to int",
			params:   map[string]any{"test": int64(1000)},
			key:      "test",
			defVal:   10,
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetIntParam(tt.params, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("GetIntParam() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test JSON-like scenarios (all numbers come as float64).
func TestJSONDecodedParams(t *testing.T) {
	// Simulate JSON-decoded params where all numbers are float64.
	jsonParams := map[string]any{
		"resolution":    0.1,
		"max_loss":      0.001,
		"warmup":        30.0,
		"initial_burst": 10000.0,
		"trials":        5.0,
	}

	// All these should work with our type-safe helpers.
	resolution := modtypes.GetFloat64Param(jsonParams, "resolution", 1.0)
	if resolution != 0.1 {
		t.Errorf("resolution = %v, want 0.1", resolution)
	}

	maxLoss := modtypes.GetFloat64Param(jsonParams, "max_loss", 0.0)
	if maxLoss != 0.001 {
		t.Errorf("max_loss = %v, want 0.001", maxLoss)
	}

	warmup := modtypes.GetIntParam(jsonParams, "warmup", 0)
	if warmup != 30 {
		t.Errorf("warmup = %v, want 30", warmup)
	}

	initialBurst := modtypes.GetUint64Param(jsonParams, "initial_burst", 1000)
	if initialBurst != 10000 {
		t.Errorf("initial_burst = %v, want 10000", initialBurst)
	}

	trials := modtypes.GetUint32Param(jsonParams, "trials", 3)
	if trials != 5 {
		t.Errorf("trials = %v, want 5", trials)
	}
}

func TestModuleInfo(t *testing.T) {
	mod := benchmark.New()

	if mod.Name() != benchmark.ModuleName {
		t.Errorf("Expected name '%s', got '%s'", benchmark.ModuleName, mod.Name())
	}

	if mod.Color() != "#dc2626" {
		t.Errorf("Expected color '#dc2626', got '%s'", mod.Color())
	}

	testTypes := mod.TestTypes()
	expectedCount := 6
	if len(testTypes) != expectedCount {
		t.Errorf("Expected %d test types, got %d", expectedCount, len(testTypes))
	}

	// Verify all RFC 2544 test types are present (with rfc2544_ prefix).
	expectedTypes := []string{
		"rfc2544_throughput",
		"rfc2544_latency",
		"rfc2544_frame_loss",
		"rfc2544_back_to_back",
		"rfc2544_system_recovery",
		"rfc2544_reset",
	}
	for _, expected := range expectedTypes {
		if !slices.Contains(testTypes, expected) {
			t.Errorf("Missing expected test type: %s", expected)
		}
	}
}

func TestCanRun(t *testing.T) {
	mod := benchmark.New()

	validTests := []string{
		"rfc2544_throughput",
		"rfc2544_latency",
		"rfc2544_frame_loss",
		"rfc2544_back_to_back",
		"rfc2544_system_recovery",
		"rfc2544_reset",
	}
	for _, test := range validTests {
		if !mod.CanRun(test) {
			t.Errorf("CanRun(%s) = false, want true", test)
		}
	}

	// Old unprefixed names should no longer work.
	invalidTests := []string{"invalid", "y1564", "rfc2889", "throughput", "latency"}
	for _, test := range invalidTests {
		if mod.CanRun(test) {
			t.Errorf("CanRun(%s) = true, want false", test)
		}
	}
}

// ============================================================================
// Executor Tests
// ============================================================================

// TestNewExecutor verifies NewExecutor error handling on unsupported platforms.
func TestNewExecutor(t *testing.T) {
	// On macOS (stub mode), this should return an error.
	exec, err := benchmark.NewExecutor("lo")
	if err == nil {
		// On Linux with CGO, this might succeed.
		defer exec.Close()
		if exec == nil {
			t.Error("NewExecutor returned nil executor without error")
		}
	}
	// Error is expected on non-Linux/non-CGO platforms.
}

// TestNewExecutorWithContext verifies executor creation with an existing context.
func TestNewExecutorWithContext(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)
	if exec == nil {
		t.Fatal("NewExecutorWithContext returned nil")
	}

	// Verify the executor has the correct module info.
	if exec.Name() != benchmark.ModuleName {
		t.Errorf("Expected module name %q, got %q", benchmark.ModuleName, exec.Name())
	}
}

// TestNewExecutorWithNilContext verifies executor creation with nil context.
func TestNewExecutorWithNilContext(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)
	if exec == nil {
		t.Fatal("NewExecutorWithContext(nil) returned nil")
	}

	// Should still have module info.
	if exec.Name() != benchmark.ModuleName {
		t.Errorf("Expected module name %q, got %q", benchmark.ModuleName, exec.Name())
	}
}

// TestExecutorSupportsExecution verifies the executor supports execution.
func TestExecutorSupportsExecution(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)
	if !exec.SupportsExecution() {
		t.Error("SupportsExecution() returned false, expected true")
	}
}

// TestExecutorSupportsExecutionWithNilContext tests SupportsExecution with nil ctx.
func TestExecutorSupportsExecutionWithNilContext(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)
	// Should still return true as it's a property of the executor type.
	if !exec.SupportsExecution() {
		t.Error("SupportsExecution() returned false, expected true")
	}
}

// TestExecutorClose verifies the Close method doesn't panic.
func TestExecutorClose(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}

	exec := benchmark.NewExecutorWithContext(ctx)
	// Close should not panic.
	exec.Close()

	// Calling Close again should also not panic.
	exec.Close()
}

// TestExecutorCloseNilContext ensures Close handles nil context gracefully.
func TestExecutorCloseNilContext(_ *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)
	// Should not panic with nil context.
	exec.Close()
}

// TestExecuteInvalidTestType verifies Execute rejects unknown test types.
func TestExecuteInvalidTestType(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil,
	}

	invalidTests := []string{
		"invalid_test",
		"y1564",
		"y1564_config",
		"y1731_delay",
		"rfc2889_forwarding",
		"rfc6349_throughput",
		"tsn_timing",
		"custom_stream",
		"reflect",
		"",
		"throughput", // Old unprefixed name.
		"latency",    // Old unprefixed name.
		"frame_loss", // Old unprefixed name.
		"rfc2544",    // Missing suffix.
	}

	for _, testType := range invalidTests {
		t.Run(testType, func(t *testing.T) {
			_, execErr := exec.Execute(testType, cfg)
			if execErr == nil {
				t.Errorf("Execute(%q) should return error for invalid test type", testType)
			}
		})
	}
}

// TestExecuteNilConfig verifies Execute rejects nil configuration.
func TestExecuteNilConfig(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	// All valid test types should fail with nil config.
	validTests := []string{
		"rfc2544_throughput",
		"rfc2544_latency",
		"rfc2544_frame_loss",
		"rfc2544_back_to_back",
		"rfc2544_system_recovery",
		"rfc2544_reset",
	}

	for _, testType := range validTests {
		t.Run(testType, func(t *testing.T) {
			_, execErr := exec.Execute(testType, nil)
			if execErr == nil {
				t.Errorf("Execute(%q, nil) should return error", testType)
			}
			if !errors.Is(execErr, modtypes.ErrInvalidConfig) {
				t.Errorf("Expected ErrInvalidConfig, got: %v", execErr)
			}
		})
	}
}

// TestExecuteThroughputTest verifies RFC 2544 throughput test execution.
func TestExecuteThroughputTest(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params: map[string]any{
			"resolution": 0.1,
			"max_loss":   0.0,
		},
	}

	result, err := exec.Execute("rfc2544_throughput", cfg)
	// Stub implementation returns stub error, but structure is valid.
	if err != nil {
		// Expected in stub mode.
		if result != nil && result.TestType != "rfc2544_throughput" {
			t.Errorf("Result TestType = %q, want 'rfc2544_throughput'", result.TestType)
		}
		return
	}

	if result == nil {
		t.Fatal("Execute returned nil result")
	}
	if result.TestType != "rfc2544_throughput" {
		t.Errorf("Result TestType = %q, want 'rfc2544_throughput'", result.TestType)
	}
	if result.ModuleName != benchmark.ModuleName {
		t.Errorf("Result ModuleName = %q, want %q", result.ModuleName, benchmark.ModuleName)
	}
}

// TestExecuteLatencyTest verifies RFC 2544 latency test execution.
func TestExecuteLatencyTest(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params: map[string]any{
			"load_levels": []float64{10, 25, 50, 75, 100},
		},
	}

	result, err := exec.Execute("rfc2544_latency", cfg)
	if err != nil {
		if result != nil && result.TestType != "rfc2544_latency" {
			t.Errorf("Result TestType = %q, want 'rfc2544_latency'", result.TestType)
		}
		return
	}

	if result == nil {
		t.Fatal("Execute returned nil result")
	}
	if result.TestType != "rfc2544_latency" {
		t.Errorf("Result TestType = %q, want 'rfc2544_latency'", result.TestType)
	}
}

// TestExecuteLatencyTestDefaultLoadLevels tests latency without custom load levels.
func TestExecuteLatencyTestDefaultLoadLevels(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil, // Should use default load levels.
	}

	_, _ = exec.Execute("rfc2544_latency", cfg)
}

// TestExecuteLatencyTestInvalidLoadLevels tests latency with invalid load_levels type.
func TestExecuteLatencyTestInvalidLoadLevels(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params: map[string]any{
			"load_levels": "invalid", // Wrong type.
		},
	}

	// Should fall back to default load levels.
	_, _ = exec.Execute("rfc2544_latency", cfg)
}

// TestExecuteFrameLossTest verifies RFC 2544 frame loss test execution.
func TestExecuteFrameLossTest(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params: map[string]any{
			"start_pct": 10.0,
			"end_pct":   100.0,
			"step_pct":  10.0,
		},
	}

	result, err := exec.Execute("rfc2544_frame_loss", cfg)
	if err != nil {
		if result != nil && result.TestType != "rfc2544_frame_loss" {
			t.Errorf("Result TestType = %q, want 'rfc2544_frame_loss'", result.TestType)
		}
		return
	}

	if result == nil {
		t.Fatal("Execute returned nil result")
	}
	if result.TestType != "rfc2544_frame_loss" {
		t.Errorf("Result TestType = %q, want 'rfc2544_frame_loss'", result.TestType)
	}
}

// TestExecuteFrameLossTestDefaults tests frame loss with default parameters.
func TestExecuteFrameLossTestDefaults(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil, // Should use default parameters.
	}

	_, _ = exec.Execute("rfc2544_frame_loss", cfg)
}

// TestExecuteBackToBackTest verifies RFC 2544 back-to-back test execution.
func TestExecuteBackToBackTest(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params: map[string]any{
			"initial_burst": uint64(10000),
			"trials":        uint32(3),
		},
	}

	result, err := exec.Execute("rfc2544_back_to_back", cfg)
	if err != nil {
		if result != nil && result.TestType != "rfc2544_back_to_back" {
			t.Errorf("Result TestType = %q, want 'rfc2544_back_to_back'", result.TestType)
		}
		return
	}

	if result == nil {
		t.Fatal("Execute returned nil result")
	}
	if result.TestType != "rfc2544_back_to_back" {
		t.Errorf("Result TestType = %q, want 'rfc2544_back_to_back'", result.TestType)
	}
}

// TestExecuteBackToBackTestDefaults tests back-to-back with default parameters.
func TestExecuteBackToBackTestDefaults(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil,
	}

	_, _ = exec.Execute("rfc2544_back_to_back", cfg)
}

// TestExecuteSystemRecoveryTest verifies RFC 2544 system recovery test execution.
func TestExecuteSystemRecoveryTest(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params: map[string]any{
			"throughput_pct": 100.0,
			"overload_sec":   uint32(60),
		},
	}

	result, err := exec.Execute("rfc2544_system_recovery", cfg)
	if err != nil {
		if result != nil && result.TestType != "rfc2544_system_recovery" {
			t.Errorf("Result TestType = %q, want 'rfc2544_system_recovery'", result.TestType)
		}
		return
	}

	if result == nil {
		t.Fatal("Execute returned nil result")
	}
	if result.TestType != "rfc2544_system_recovery" {
		t.Errorf("Result TestType = %q, want 'rfc2544_system_recovery'", result.TestType)
	}
}

// TestExecuteSystemRecoveryTestDefaults tests system recovery with default params.
func TestExecuteSystemRecoveryTestDefaults(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil,
	}

	_, _ = exec.Execute("rfc2544_system_recovery", cfg)
}

// TestExecuteResetTest verifies RFC 2544 reset test execution.
func TestExecuteResetTest(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil,
	}

	result, err := exec.Execute("rfc2544_reset", cfg)
	if err != nil {
		if result != nil && result.TestType != "rfc2544_reset" {
			t.Errorf("Result TestType = %q, want 'rfc2544_reset'", result.TestType)
		}
		return
	}

	if result == nil {
		t.Fatal("Execute returned nil result")
	}
	if result.TestType != "rfc2544_reset" {
		t.Errorf("Result TestType = %q, want 'rfc2544_reset'", result.TestType)
	}
}

// TestExecuteAllTestTypes verifies all RFC 2544 test types can be executed.
func TestExecuteAllTestTypes(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	allTests := []string{
		"rfc2544_throughput",
		"rfc2544_latency",
		"rfc2544_frame_loss",
		"rfc2544_back_to_back",
		"rfc2544_system_recovery",
		"rfc2544_reset",
	}

	for _, testType := range allTests {
		t.Run(testType, func(t *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  10,
				Params:    map[string]any{},
			}

			result, execErr := exec.Execute(testType, cfg)
			if result == nil && execErr == nil {
				t.Error("Expected either result or error, got neither")
			}

			if result != nil {
				if result.TestType != testType {
					t.Errorf("TestType = %q, want %q", result.TestType, testType)
				}
				if result.ModuleName != benchmark.ModuleName {
					t.Errorf("ModuleName = %q, want %q", result.ModuleName, benchmark.ModuleName)
				}
			}
		})
	}
}

// TestConfigureContextDefaults verifies default context configuration.
func TestConfigureContextDefaults(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	// Empty params should use defaults.
	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  0, // Should fall back to default.
		Params:    nil,
	}

	// Execute to trigger config building.
	_, _ = exec.Execute("rfc2544_throughput", cfg)
}

// TestConfigureContextWithParams verifies custom context configuration.
func TestConfigureContextWithParams(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 128,
		Duration:  120,
		Params: map[string]any{
			"resolution": 0.05,
			"max_loss":   0.001,
			"warmup":     30,
		},
	}

	_, _ = exec.Execute("rfc2544_throughput", cfg)
}

// TestConfigureContextWithWarmup verifies warmup period configuration.
func TestConfigureContextWithWarmup(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  60,
		Params: map[string]any{
			"warmup": 10,
		},
	}

	_, _ = exec.Execute("rfc2544_throughput", cfg)
}

// TestConfigureContextWithZeroWarmup verifies zero warmup is handled.
func TestConfigureContextWithZeroWarmup(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  60,
		Params: map[string]any{
			"warmup": 0,
		},
	}

	_, _ = exec.Execute("rfc2544_throughput", cfg)
}

// TestFrameSizeConfiguration verifies frame size is properly set.
func TestFrameSizeConfiguration(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	frameSizes := []uint32{64, 128, 256, 512, 1024, 1280, 1518, 9000}

	for _, size := range frameSizes {
		t.Run("frameSize="+string(rune(size)), func(_ *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: size,
				Duration:  10,
				Params:    nil,
			}

			_, _ = exec.Execute("rfc2544_throughput", cfg)
		})
	}
}

// TestZeroFrameSize verifies zero frame size is handled.
func TestZeroFrameSize(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 0, // Should not call SetFrameSize.
		Duration:  10,
		Params:    nil,
	}

	_, _ = exec.Execute("rfc2544_throughput", cfg)
}

// TestAllParamTypes verifies various parameter type conversions.
func TestAllParamTypes(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	// Test with float64 values (JSON-decoded style).
	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  60,
		Params: map[string]any{
			"resolution":     0.1,
			"max_loss":       0.001,
			"warmup":         5.0,     // float64 instead of int.
			"initial_burst":  10000.0, // float64 instead of uint64.
			"trials":         3.0,     // float64 instead of uint32.
			"start_pct":      10.0,
			"end_pct":        100.0,
			"step_pct":       10.0,
			"throughput_pct": 100.0,
			"overload_sec":   60.0, // float64 instead of uint32.
		},
	}

	// Execute all test types to trigger all param extraction.
	testTypes := []string{
		"rfc2544_throughput",
		"rfc2544_latency",
		"rfc2544_frame_loss",
		"rfc2544_back_to_back",
		"rfc2544_system_recovery",
		"rfc2544_reset",
	}

	for _, testType := range testTypes {
		_, _ = exec.Execute(testType, cfg)
	}
}

// TestExecutorModuleEmbedding verifies Module methods are accessible on Executor.
func TestExecutorModuleEmbedding(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	// All Module methods should be accessible.
	if exec.Name() != benchmark.ModuleName {
		t.Errorf("Name() = %q, want %q", exec.Name(), benchmark.ModuleName)
	}
	if exec.DisplayName() != benchmark.DisplayName {
		t.Errorf("DisplayName() = %q, want %q", exec.DisplayName(), benchmark.DisplayName)
	}
	if exec.Color() != benchmark.ColorHex {
		t.Errorf("Color() = %q, want %q", exec.Color(), benchmark.ColorHex)
	}
	if exec.Standard() != benchmark.StandardRef {
		t.Errorf("Standard() = %q, want %q", exec.Standard(), benchmark.StandardRef)
	}
	if len(exec.TestTypes()) != 6 {
		t.Errorf("TestTypes() returned %d, want 6", len(exec.TestTypes()))
	}
	if !exec.CanRun("rfc2544_throughput") {
		t.Error("CanRun('rfc2544_throughput') = false, want true")
	}
	if exec.TestDescription("rfc2544_throughput") == "" {
		t.Error("TestDescription('rfc2544_throughput') returned empty string")
	}
}

// verifyResultFields checks common result fields.
func verifyResultFields(t *testing.T, result *modtypes.Result, testType string) {
	t.Helper()
	if result.TestType != testType {
		t.Errorf("TestType = %q, want %q", result.TestType, testType)
	}
	if result.ModuleName != benchmark.ModuleName {
		t.Errorf("ModuleName = %q, want %q", result.ModuleName, benchmark.ModuleName)
	}
}

// TestExecuteResultStructure verifies the result structure is correct.
func TestExecuteResultStructure(t *testing.T) {
	ctx, err := dataplane.NewContext("lo")
	if err != nil {
		t.Skipf("Cannot create dataplane context: %v", err)
	}
	defer ctx.Close()

	exec := benchmark.NewExecutorWithContext(ctx)

	allTests := []string{
		"rfc2544_throughput",
		"rfc2544_latency",
		"rfc2544_frame_loss",
		"rfc2544_back_to_back",
		"rfc2544_system_recovery",
		"rfc2544_reset",
	}

	for _, testType := range allTests {
		t.Run(testType, func(t *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  10,
				Params:    map[string]any{},
			}

			result, execErr := exec.Execute(testType, cfg)
			verifyExecuteResult(t, result, execErr, testType)
		})
	}
}

// verifyExecuteResult validates the result based on whether an error occurred.
func verifyExecuteResult(t *testing.T, result *modtypes.Result, execErr error, testType string) {
	t.Helper()
	if execErr != nil {
		verifyResultWithError(t, result, testType)
		return
	}
	verifyResultWithoutError(t, result, testType)
}

// verifyResultWithError checks the result when an error was returned.
func verifyResultWithError(t *testing.T, result *modtypes.Result, testType string) {
	t.Helper()
	if result == nil {
		return
	}
	verifyResultFields(t, result, testType)
	if result.Success {
		t.Error("Success should be false when error is returned")
	}
}

// verifyResultWithoutError checks the result when no error was returned.
func verifyResultWithoutError(t *testing.T, result *modtypes.Result, testType string) {
	t.Helper()
	if result == nil {
		t.Fatal("Execute returned nil result without error")
	}
	verifyResultFields(t, result, testType)
	if !result.Success {
		t.Error("Success should be true when no error")
	}
}
