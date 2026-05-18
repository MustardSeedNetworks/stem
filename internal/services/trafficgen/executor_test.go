// SPDX-License-Identifier: BUSL-1.1

// Black-box testing of trafficgen executor.
//
// Tests verify public API behavior. The export_test.go file in package trafficgen
// provides NewMockExecutor, NewMockExecutorWithNilModule, and test constants.
package trafficgen_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/krisarmstrong/stem/internal/services/modtypes"
	"github.com/krisarmstrong/stem/internal/services/trafficgen"
)

// TestNewExecutor verifies executor creation behavior.
// On stub builds (non-CGO/non-Linux), this will fail with ErrNotSupported.
func TestNewExecutor(t *testing.T) {
	// NewExecutor requires a dataplane context which is stubbed on non-Linux.
	// We test that it returns an error as expected.
	executor, err := trafficgen.NewExecutor("eth0")

	// On stub builds, we expect an error.
	if err == nil {
		// If we got a valid executor (Linux/CGO build), clean up.
		if executor != nil {
			executor.Close()
		}
		t.Skip("Dataplane available; skipping stub error test")
	}

	// Verify error is related to platform/dataplane unavailability.
	if executor != nil {
		t.Error("NewExecutor() should return nil executor on error")
	}

	// Error should mention dataplane or platform.
	errStr := err.Error()
	if errStr == "" {
		t.Error("NewExecutor() error should have a message")
	}
}

// TestNewExecutorEmptyInterface tests with empty interface name.
func TestNewExecutorEmptyInterface(t *testing.T) {
	executor, err := trafficgen.NewExecutor("")

	if err == nil {
		if executor != nil {
			executor.Close()
		}
		t.Skip("Dataplane available; skipping stub error test")
	}

	if executor != nil {
		t.Error("NewExecutor(\"\") should return nil executor on error")
	}
}

// TestNewExecutorVariousInterfaces tests with various interface names.
func TestNewExecutorVariousInterfaces(t *testing.T) {
	interfaces := []string{"lo", "en0", "eth1", "bond0"}

	for _, iface := range interfaces {
		t.Run(iface, func(t *testing.T) {
			executor, err := trafficgen.NewExecutor(iface)
			if err == nil {
				executor.Close()
				return // Dataplane available
			}
			// On stub builds, error is expected
			if executor != nil {
				t.Errorf("NewExecutor(%q) should return nil executor on error", iface)
			}
		})
	}
}

// TestModtypesSafeIntToUint32 tests the modtypes.SafeIntToUint32 helper function.
func TestModtypesSafeIntToUint32(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected uint32
	}{
		{
			name:     "positive value within range",
			value:    100,
			expected: 100,
		},
		{
			name:     "zero value",
			value:    0,
			expected: 0,
		},
		{
			name:     "negative value returns 0",
			value:    -1,
			expected: 0,
		},
		{
			name:     "large negative value returns 0",
			value:    -1000000,
			expected: 0,
		},
		{
			name:     "max int32 value",
			value:    math.MaxInt32,
			expected: math.MaxInt32,
		},
		{
			name:     "max uint32 as int (if int is 64-bit)",
			value:    math.MaxUint32,
			expected: math.MaxUint32,
		},
		{
			name:     "value just above max uint32 returns max",
			value:    math.MaxUint32 + 1,
			expected: math.MaxUint32,
		},
		{
			name:     "large positive value over uint32 max returns max",
			value:    int(math.MaxInt64 / 2),
			expected: math.MaxUint32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.SafeIntToUint32(tt.value)
			if result != tt.expected {
				t.Errorf("modtypes.SafeIntToUint32(%d) = %d, want %d",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestModtypesGetUint16Param tests the modtypes.GetUint16Param helper function.
func TestModtypesGetUint16Param(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]any
		key        string
		defaultVal uint16
		expected   uint16
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
			params:     map[string]any{"val": float64(100)},
			key:        "val",
			defaultVal: 0,
			expected:   100,
		},
		{
			name:       "mid range value",
			params:     map[string]any{"val": float64(32000)},
			key:        "val",
			defaultVal: 0,
			expected:   32000,
		},
		{
			name:       "max uint16 value",
			params:     map[string]any{"val": float64(math.MaxUint16)},
			key:        "val",
			defaultVal: 0,
			expected:   math.MaxUint16,
		},
		{
			name:       "value just above max uint16 returns default",
			params:     map[string]any{"val": float64(math.MaxUint16 + 1)},
			key:        "val",
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "large value returns default",
			params:     map[string]any{"val": float64(math.MaxUint32)},
			key:        "val",
			defaultVal: 99,
			expected:   99,
		},
		{
			name:       "typical VLAN ID",
			params:     map[string]any{"val": float64(4094)},
			key:        "val",
			defaultVal: 0,
			expected:   4094,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint16Param(tt.params, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetUint16Param() = %d, want %d",
					result, tt.expected)
			}
		})
	}
}

// TestModtypesGetUint8Param tests the modtypes.GetUint8Param helper function.
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
			params:     map[string]any{"val": float64(7)},
			key:        "val",
			defaultVal: 0,
			expected:   7,
		},
		{
			name:       "max uint8 value",
			params:     map[string]any{"val": float64(math.MaxUint8)},
			key:        "val",
			defaultVal: 0,
			expected:   math.MaxUint8,
		},
		{
			name:       "value just above max uint8 returns default",
			params:     map[string]any{"val": float64(math.MaxUint8 + 1)},
			key:        "val",
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "large value returns default",
			params:     map[string]any{"val": float64(math.MaxUint32)},
			key:        "val",
			defaultVal: 99,
			expected:   99,
		},
		{
			name:       "typical VLAN priority",
			params:     map[string]any{"val": float64(7)},
			key:        "val",
			defaultVal: 0,
			expected:   7,
		},
		{
			name:       "CoS value",
			params:     map[string]any{"val": float64(6)},
			key:        "val",
			defaultVal: 0,
			expected:   6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint8Param(tt.params, "val", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetUint8Param() = %d, want %d",
					result, tt.expected)
			}
		})
	}
}

// TestExecutorSupportsExecution tests the SupportsExecution method.
func TestExecutorSupportsExecution(t *testing.T) {
	// Use mock executor since SupportsExecution doesn't need dataplane.
	executor := trafficgen.NewMockExecutor()

	if !executor.SupportsExecution() {
		t.Error("SupportsExecution() should return true")
	}
}

// TestExecutorSupportsExecutionAlwaysTrue verifies it always returns true.
func TestExecutorSupportsExecutionAlwaysTrue(t *testing.T) {
	// Test with various executor states.
	testCases := []struct {
		name     string
		executor *trafficgen.Executor
	}{
		{
			name:     "mock executor",
			executor: trafficgen.NewMockExecutor(),
		},
		{
			name:     "executor with nil module",
			executor: trafficgen.NewMockExecutorWithNilModule(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// SupportsExecution should always return true regardless of state.
			if !tc.executor.SupportsExecution() {
				t.Error("SupportsExecution() should always return true")
			}
		})
	}
}

// TestExecutorClose tests the Close method handles nil context gracefully.
func TestExecutorClose(t *testing.T) {
	// Test that Close on executor with nil context doesn't panic.
	executor := trafficgen.NewMockExecutor()

	// This should not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close() panicked on executor with nil context: %v", r)
		}
	}()

	// Close is safe on nil context.
	executor.Close()
}

// TestExecutorCloseWithNilContext tests Close with nil context.
func TestExecutorCloseWithNilContext(t *testing.T) {
	// Create an executor with nil context via the exported helper.
	executor := trafficgen.NewMockExecutor()

	// This should not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close() panicked with nil context: %v", r)
		}
	}()

	executor.Close()
}

// TestExecuteInvalidTestType tests Execute with invalid test type.
func TestExecuteInvalidTestType(t *testing.T) {
	// Use mock executor to test the validation logic.
	executor := trafficgen.NewMockExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 1518,
		Duration:  60,
		Params:    make(map[string]any),
	}

	// Test with invalid test type.
	result, err := executor.Execute("invalid_test", cfg)
	if err == nil {
		t.Error("Execute() with invalid test type should return error")
	}
	if result != nil {
		t.Error("Execute() with invalid test type should return nil result")
	}

	// Error should mention the invalid test type.
	if err != nil && !strings.Contains(err.Error(), "cannot run") {
		t.Errorf("Error should mention 'cannot run', got: %v", err)
	}
}

// TestExecuteNilConfig tests Execute with nil config.
func TestExecuteNilConfig(t *testing.T) {
	// Use mock executor to test the validation logic.
	executor := trafficgen.NewMockExecutor()

	result, err := executor.Execute("custom_stream", nil)
	if err == nil {
		t.Error("Execute() with nil config should return error")
	}
	if !errors.Is(err, modtypes.ErrInvalidConfig) {
		t.Errorf("Expected ErrInvalidConfig, got: %v", err)
	}
	if result != nil {
		t.Error("Execute() with nil config should return nil result")
	}
}

// TestExecuteValidationOrder tests that validation happens in correct order.
func TestExecuteValidationOrder(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	// Test invalid test type checked before nil config.
	result, err := executor.Execute("invalid_test", nil)
	if err == nil {
		t.Error("Execute() should return error")
	}
	// Should fail on test type check first.
	if strings.Contains(err.Error(), "invalid config") {
		t.Error("Should fail on test type check before config check")
	}
	if result != nil {
		t.Error("Execute() should return nil result")
	}
}

// TestExecuteCustomStream tests Execute with valid custom_stream config.
// Uses mock executor - dataplane call will fail but we can test the setup logic.
func TestExecuteCustomStream(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 1518,
		Duration:  10,
		Params: map[string]any{
			"rate_pct":           50.0,
			"warmup_sec":         float64(2),
			"stream_id":          float64(1),
			"burst_mode":         false,
			"burst_size":         float64(100),
			"inter_burst_gap_us": float64(1000),
			"src_mac":            "00:11:22:33:44:55",
			"dst_mac":            "66:77:88:99:aa:bb",
			"vlan_id":            float64(100),
			"vlan_priority":      float64(5),
		},
	}

	// With nil context, Execute will fail at dataplane call.
	result, err := executor.Execute("custom_stream", cfg)

	// Should return result with error (dataplane failed).
	if err == nil {
		t.Error("Execute() with nil context should return error")
	}
	if result == nil {
		t.Fatal("Execute() should return result even on failure")
	}
	if result.Success {
		t.Error("result.Success should be false on error")
	}
	if result.TestType != "custom_stream" {
		t.Errorf("result.TestType = %s, want custom_stream", result.TestType)
	}
	if result.ModuleName != trafficgen.ModuleName {
		t.Errorf("result.ModuleName = %s, want %s", result.ModuleName, trafficgen.ModuleName)
	}
	if result.Error == "" {
		t.Error("result.Error should be set on failure")
	}
}

// TestExecuteWithDefaultParams tests Execute uses defaults for missing params.
func TestExecuteWithDefaultParams(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	// Minimal config with no params - should use defaults.
	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  0, // Test that 0 duration uses param fallback.
		Params:    nil,
	}

	result, err := executor.Execute("custom_stream", cfg)

	// Should fail at dataplane but config should be valid.
	if err == nil {
		t.Skip("Unexpected success - dataplane may be available")
	}
	if result == nil {
		t.Error("Execute() should return result")
	}
}

// TestExecuteWithEmptyParams tests Execute with empty params map.
func TestExecuteWithEmptyParams(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 1518,
		Duration:  60,
		Params:    map[string]any{},
	}

	result, err := executor.Execute("custom_stream", cfg)

	// Should fail at dataplane but config should be valid.
	if err == nil {
		t.Skip("Unexpected success - dataplane may be available")
	}
	if result == nil {
		t.Error("Execute() should return result")
	}
}

// TestExecuteWithBurstMode tests Execute with burst mode enabled.
func TestExecuteWithBurstMode(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 512,
		Duration:  30,
		Params: map[string]any{
			"burst_mode":         true,
			"burst_size":         float64(500),
			"inter_burst_gap_us": float64(2000),
		},
	}

	result, err := executor.Execute("custom_stream", cfg)

	// Should fail at dataplane but config should be valid.
	if err == nil {
		t.Skip("Unexpected success - dataplane may be available")
	}
	if result == nil {
		t.Error("Execute() should return result")
	}
}

// TestExecuteWithVLAN tests Execute with VLAN configuration.
func TestExecuteWithVLAN(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 1518,
		Duration:  60,
		Params: map[string]any{
			"vlan_id":       float64(4094), // Max VLAN ID.
			"vlan_priority": float64(7),    // Max priority.
		},
	}

	result, err := executor.Execute("custom_stream", cfg)

	// Should fail at dataplane but config should be valid.
	if err == nil {
		t.Skip("Unexpected success - dataplane may be available")
	}
	if result == nil {
		t.Error("Execute() should return result")
	}
}

// TestExecuteWithOverflowVLAN tests VLAN ID overflow handling.
func TestExecuteWithOverflowVLAN(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 1518,
		Duration:  60,
		Params: map[string]any{
			"vlan_id":       float64(100000), // Exceeds uint16 max.
			"vlan_priority": float64(300),    // Exceeds uint8 max.
		},
	}

	result, err := executor.Execute("custom_stream", cfg)

	// Should fail at dataplane but safeUint16/safeUint8 should clamp values.
	if err == nil {
		t.Skip("Unexpected success - dataplane may be available")
	}
	if result == nil {
		t.Error("Execute() should return result")
	}
}

// TestExecuteNonCustomStreamTestType tests that non-custom_stream types fail.
func TestExecuteNonCustomStreamTestType(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 1518,
		Duration:  60,
		Params:    map[string]any{},
	}

	// These test types don't belong to trafficgen module.
	invalidTypes := []string{
		"rfc2544_throughput",
		"y1564",
		"reflect",
		"rfc2889_forwarding",
		"y1731_delay",
	}

	for _, testType := range invalidTypes {
		t.Run(testType, func(t *testing.T) {
			result, err := executor.Execute(testType, cfg)
			if err == nil {
				t.Errorf("Execute(%q) should return error", testType)
			}
			if result != nil {
				t.Errorf("Execute(%q) should return nil result", testType)
			}
		})
	}
}

// TestExecuteUnsupportedTestType tests that CanRun check happens before config parsing.
func TestExecuteUnsupportedTestType(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	// Should fail on CanRun check even with valid config.
	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 1518,
		Duration:  60,
		Params:    map[string]any{},
	}

	result, err := executor.Execute("not_a_valid_test", cfg)
	if err == nil {
		t.Error("Execute() should return error for unsupported test type")
	}
	if result != nil {
		t.Error("Execute() should return nil result for unsupported test type")
	}
	// Error message should indicate the test type can't run.
	if !strings.Contains(err.Error(), "cannot run") {
		t.Errorf("Error should mention 'cannot run', got: %v", err)
	}
}

// TestExecuteNonImplementedTestType tests that non-custom_stream but "valid" types
// return ErrTestNotImplemented - but since trafficgen only has custom_stream,
// any other test type should fail the CanRun check first.
func TestExecuteNonImplementedTestType(t *testing.T) {
	executor := trafficgen.NewMockExecutor()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 1518,
		Duration:  60,
		Params:    map[string]any{},
	}

	// "custom" is not a valid test type for trafficgen.
	result, err := executor.Execute("custom", cfg)
	if err == nil {
		t.Error("Execute() should return error")
	}
	if result != nil {
		t.Error("Execute() should return nil result")
	}
}

// TestModuleEmbeddingInExecutor verifies that Executor embeds Module correctly.
func TestModuleEmbeddingInExecutor(t *testing.T) {
	executor, err := trafficgen.NewExecutor("eth0")
	if err != nil {
		// Create a mock executor to test embedding.
		executor = trafficgen.NewMockExecutor()
	} else {
		defer executor.Close()
	}

	// Test that embedded Module methods work.
	if executor.Name() != trafficgen.ModuleName {
		t.Errorf("executor.Name() = %s, want %s", executor.Name(), trafficgen.ModuleName)
	}

	if executor.DisplayName() != trafficgen.DisplayName {
		t.Errorf("executor.DisplayName() = %s, want %s", executor.DisplayName(), trafficgen.DisplayName)
	}

	if executor.Color() != trafficgen.ColorHex {
		t.Errorf("executor.Color() = %s, want %s", executor.Color(), trafficgen.ColorHex)
	}

	if executor.Standard() != trafficgen.StandardRef {
		t.Errorf("executor.Standard() = %s, want %s", executor.Standard(), trafficgen.StandardRef)
	}

	if !executor.CanRun("custom_stream") {
		t.Error("executor.CanRun(\"custom_stream\") should be true")
	}

	if executor.CanRun("invalid") {
		t.Error("executor.CanRun(\"invalid\") should be false")
	}

	execTestTypes := executor.TestTypes()
	if len(execTestTypes) != 1 {
		t.Errorf("executor.TestTypes() length = %d, want 1", len(execTestTypes))
	}
}

// TestDefaultConstants verifies the default constant values.
func TestDefaultConstants(t *testing.T) {
	// Verify defaults match TUI/WebUI expectations per comments.
	const (
		expectedDefaultRatePct         = 100.0
		expectedDefaultWarmupSec       = 2
		expectedDefaultDurationSec     = 60
		expectedDefaultStreamID        = 1
		expectedDefaultBurstSize       = 100
		expectedDefaultInterBurstGapUs = 1000
	)

	if trafficgen.TestDefaultRatePct != expectedDefaultRatePct {
		t.Errorf("TestDefaultRatePct = %v, want %v", trafficgen.TestDefaultRatePct, expectedDefaultRatePct)
	}
	if trafficgen.TestDefaultWarmupSec != expectedDefaultWarmupSec {
		t.Errorf("TestDefaultWarmupSec = %v, want %v", trafficgen.TestDefaultWarmupSec, expectedDefaultWarmupSec)
	}
	if trafficgen.TestDefaultDurationSec != expectedDefaultDurationSec {
		t.Errorf("TestDefaultDurationSec = %v, want %v", trafficgen.TestDefaultDurationSec, expectedDefaultDurationSec)
	}
	if trafficgen.TestDefaultStreamID != expectedDefaultStreamID {
		t.Errorf("TestDefaultStreamID = %v, want %v", trafficgen.TestDefaultStreamID, expectedDefaultStreamID)
	}
	if trafficgen.TestDefaultBurstSize != expectedDefaultBurstSize {
		t.Errorf("TestDefaultBurstSize = %v, want %v", trafficgen.TestDefaultBurstSize, expectedDefaultBurstSize)
	}
	if trafficgen.TestDefaultInterBurstGapUs != expectedDefaultInterBurstGapUs {
		t.Errorf("TestDefaultInterBurstGapUs = %v, want %v",
			trafficgen.TestDefaultInterBurstGapUs, expectedDefaultInterBurstGapUs)
	}
}

// TestModtypesSafeIntToUint32Boundary tests boundary conditions for modtypes.SafeIntToUint32.
func TestModtypesSafeIntToUint32Boundary(t *testing.T) {
	// Test the exact boundaries.
	tests := []struct {
		name     string
		value    int
		expected uint32
	}{
		{
			name:     "value at -1",
			value:    -1,
			expected: 0,
		},
		{
			name:     "value at 0",
			value:    0,
			expected: 0,
		},
		{
			name:     "value at 1",
			value:    1,
			expected: 1,
		},
		{
			name:     "value at MaxUint32",
			value:    math.MaxUint32,
			expected: math.MaxUint32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.SafeIntToUint32(tt.value)
			if result != tt.expected {
				t.Errorf("modtypes.SafeIntToUint32(%d) = %d, want %d",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestModtypesGetUint16ParamBoundary tests boundary conditions for modtypes.GetUint16Param.
func TestModtypesGetUint16ParamBoundary(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]any
		defaultVal uint16
		expected   uint16
	}{
		{
			name:       "value at MaxUint16 - 1",
			params:     map[string]any{"val": float64(math.MaxUint16 - 1)},
			defaultVal: 0,
			expected:   math.MaxUint16 - 1,
		},
		{
			name:       "value at MaxUint16",
			params:     map[string]any{"val": float64(math.MaxUint16)},
			defaultVal: 0,
			expected:   math.MaxUint16,
		},
		{
			name:       "value at MaxUint16 + 1 returns default",
			params:     map[string]any{"val": float64(math.MaxUint16 + 1)},
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "value at MaxUint16 + 2 returns default",
			params:     map[string]any{"val": float64(math.MaxUint16 + 2)},
			defaultVal: 99,
			expected:   99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint16Param(tt.params, "val", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetUint16Param() = %d, want %d",
					result, tt.expected)
			}
		})
	}
}

// TestModtypesGetUint8ParamBoundary tests boundary conditions for modtypes.GetUint8Param.
func TestModtypesGetUint8ParamBoundary(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]any
		defaultVal uint8
		expected   uint8
	}{
		{
			name:       "value at MaxUint8 - 1",
			params:     map[string]any{"val": float64(math.MaxUint8 - 1)},
			defaultVal: 0,
			expected:   math.MaxUint8 - 1,
		},
		{
			name:       "value at MaxUint8",
			params:     map[string]any{"val": float64(math.MaxUint8)},
			defaultVal: 0,
			expected:   math.MaxUint8,
		},
		{
			name:       "value at MaxUint8 + 1 returns default",
			params:     map[string]any{"val": float64(math.MaxUint8 + 1)},
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "value at MaxUint8 + 2 returns default",
			params:     map[string]any{"val": float64(math.MaxUint8 + 2)},
			defaultVal: 99,
			expected:   99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetUint8Param(tt.params, "val", tt.defaultVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetUint8Param() = %d, want %d",
					result, tt.expected)
			}
		})
	}
}
