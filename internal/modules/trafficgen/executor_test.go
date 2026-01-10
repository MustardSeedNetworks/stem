// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

//nolint:testpackage // White-box testing requires access to unexported functions.
package trafficgen

import (
	"errors"
	"math"
	"testing"

	"github.com/krisarmstrong/stem/internal/modules/modtypes"
)

// newMockExecutor creates an executor with mock context for testing.
// This allows testing Execute logic without requiring actual dataplane.
func newMockExecutor() *Executor {
	return &Executor{
		Module: New(),
		ctx:    nil, // nil context for testing error paths
	}
}

// TestNewExecutor verifies executor creation behavior.
// On stub builds (non-CGO/non-Linux), this will fail with ErrNotSupported.
func TestNewExecutor(t *testing.T) {
	// NewExecutor requires a dataplane context which is stubbed on non-Linux.
	// We test that it returns an error as expected.
	executor, err := NewExecutor("eth0")

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
	executor, err := NewExecutor("")

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
			executor, err := NewExecutor(iface)
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

// TestSafeUint32FromInt tests the safeUint32FromInt helper function.
func TestSafeUint32FromInt(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		fallback uint32
		expected uint32
	}{
		{
			name:     "positive value within range",
			value:    100,
			fallback: 0,
			expected: 100,
		},
		{
			name:     "zero value",
			value:    0,
			fallback: 10,
			expected: 0,
		},
		{
			name:     "negative value returns fallback",
			value:    -1,
			fallback: 42,
			expected: 42,
		},
		{
			name:     "large negative value returns fallback",
			value:    -1000000,
			fallback: 99,
			expected: 99,
		},
		{
			name:     "max int32 value",
			value:    math.MaxInt32,
			fallback: 0,
			expected: math.MaxInt32,
		},
		{
			name:     "max uint32 as int (if int is 64-bit)",
			value:    math.MaxUint32,
			fallback: 0,
			expected: math.MaxUint32,
		},
		{
			name:     "value just above max uint32 returns fallback",
			value:    math.MaxUint32 + 1,
			fallback: 55,
			expected: 55,
		},
		{
			name:     "large positive value over uint32 max returns fallback",
			value:    int(math.MaxInt64 / 2),
			fallback: 123,
			expected: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeUint32FromInt(tt.value, tt.fallback)
			if result != tt.expected {
				t.Errorf("safeUint32FromInt(%d, %d) = %d, want %d",
					tt.value, tt.fallback, result, tt.expected)
			}
		})
	}
}

// TestSafeUint16 tests the safeUint16 helper function.
func TestSafeUint16(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected uint16
	}{
		{
			name:     "zero value",
			value:    0,
			expected: 0,
		},
		{
			name:     "small value",
			value:    100,
			expected: 100,
		},
		{
			name:     "mid range value",
			value:    32000,
			expected: 32000,
		},
		{
			name:     "max uint16 value",
			value:    math.MaxUint16,
			expected: math.MaxUint16,
		},
		{
			name:     "value just above max uint16 returns max",
			value:    math.MaxUint16 + 1,
			expected: math.MaxUint16,
		},
		{
			name:     "large value returns max uint16",
			value:    math.MaxUint32,
			expected: math.MaxUint16,
		},
		{
			name:     "typical VLAN ID",
			value:    4094,
			expected: 4094,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeUint16(tt.value)
			if result != tt.expected {
				t.Errorf("safeUint16(%d) = %d, want %d",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestSafeUint8 tests the safeUint8 helper function.
func TestSafeUint8(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected uint8
	}{
		{
			name:     "zero value",
			value:    0,
			expected: 0,
		},
		{
			name:     "small value",
			value:    7,
			expected: 7,
		},
		{
			name:     "max uint8 value",
			value:    math.MaxUint8,
			expected: math.MaxUint8,
		},
		{
			name:     "value just above max uint8 returns max",
			value:    math.MaxUint8 + 1,
			expected: math.MaxUint8,
		},
		{
			name:     "large value returns max uint8",
			value:    math.MaxUint32,
			expected: math.MaxUint8,
		},
		{
			name:     "typical VLAN priority",
			value:    7,
			expected: 7,
		},
		{
			name:     "CoS value",
			value:    6,
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeUint8(tt.value)
			if result != tt.expected {
				t.Errorf("safeUint8(%d) = %d, want %d",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestExecutorSupportsExecution tests the SupportsExecution method.
func TestExecutorSupportsExecution(t *testing.T) {
	// Use mock executor since SupportsExecution doesn't need dataplane.
	executor := newMockExecutor()

	if !executor.SupportsExecution() {
		t.Error("SupportsExecution() should return true")
	}
}

// TestExecutorSupportsExecutionAlwaysTrue verifies it always returns true.
func TestExecutorSupportsExecutionAlwaysTrue(t *testing.T) {
	// Test with various executor states.
	testCases := []struct {
		name     string
		executor *Executor
	}{
		{
			name:     "mock executor",
			executor: newMockExecutor(),
		},
		{
			name: "executor with nil module",
			executor: &Executor{
				Module: nil,
				ctx:    nil,
			},
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
	executor := newMockExecutor()

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
	// Create an executor manually with nil context.
	executor := &Executor{
		Module: New(),
		ctx:    nil,
	}

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
	executor := newMockExecutor()

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
	if err != nil && !containsSubstring(err.Error(), "cannot run") {
		t.Errorf("Error should mention 'cannot run', got: %v", err)
	}
}

// TestExecuteNilConfig tests Execute with nil config.
func TestExecuteNilConfig(t *testing.T) {
	// Use mock executor to test the validation logic.
	executor := newMockExecutor()

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
	executor := newMockExecutor()

	// Test invalid test type checked before nil config.
	result, err := executor.Execute("invalid_test", nil)
	if err == nil {
		t.Error("Execute() should return error")
	}
	// Should fail on test type check first.
	if containsSubstring(err.Error(), "invalid config") {
		t.Error("Should fail on test type check before config check")
	}
	if result != nil {
		t.Error("Execute() should return nil result")
	}
}

// TestExecuteCustomStream tests Execute with valid custom_stream config.
// Uses mock executor - dataplane call will fail but we can test the setup logic.
func TestExecuteCustomStream(t *testing.T) {
	executor := newMockExecutor()

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
	if result.ModuleName != ModuleName {
		t.Errorf("result.ModuleName = %s, want %s", result.ModuleName, ModuleName)
	}
	if result.Error == "" {
		t.Error("result.Error should be set on failure")
	}
}

// TestExecuteWithDefaultParams tests Execute uses defaults for missing params.
func TestExecuteWithDefaultParams(t *testing.T) {
	executor := newMockExecutor()

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
	executor := newMockExecutor()

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
	executor := newMockExecutor()

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
	executor := newMockExecutor()

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
	executor := newMockExecutor()

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
	executor := newMockExecutor()

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
	executor := newMockExecutor()

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
	if !containsSubstring(err.Error(), "cannot run") {
		t.Errorf("Error should mention 'cannot run', got: %v", err)
	}
}

// TestExecuteNonImplementedTestType tests that non-custom_stream but "valid" types
// return ErrTestNotImplemented - but since trafficgen only has custom_stream,
// any other test type should fail the CanRun check first.
func TestExecuteNonImplementedTestType(t *testing.T) {
	executor := newMockExecutor()

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
	executor, err := NewExecutor("eth0")
	if err != nil {
		// Create a mock executor to test embedding.
		executor = &Executor{
			Module: New(),
			ctx:    nil,
		}
	} else {
		defer executor.Close()
	}

	// Test that embedded Module methods work.
	if executor.Name() != ModuleName {
		t.Errorf("executor.Name() = %s, want %s", executor.Name(), ModuleName)
	}

	if executor.DisplayName() != DisplayName {
		t.Errorf("executor.DisplayName() = %s, want %s", executor.DisplayName(), DisplayName)
	}

	if executor.Color() != ColorHex {
		t.Errorf("executor.Color() = %s, want %s", executor.Color(), ColorHex)
	}

	if executor.Standard() != StandardRef {
		t.Errorf("executor.Standard() = %s, want %s", executor.Standard(), StandardRef)
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

	if defaultRatePct != expectedDefaultRatePct {
		t.Errorf("defaultRatePct = %v, want %v", defaultRatePct, expectedDefaultRatePct)
	}
	if defaultWarmupSec != expectedDefaultWarmupSec {
		t.Errorf("defaultWarmupSec = %v, want %v", defaultWarmupSec, expectedDefaultWarmupSec)
	}
	if defaultDurationSec != expectedDefaultDurationSec {
		t.Errorf("defaultDurationSec = %v, want %v", defaultDurationSec, expectedDefaultDurationSec)
	}
	if defaultStreamID != expectedDefaultStreamID {
		t.Errorf("defaultStreamID = %v, want %v", defaultStreamID, expectedDefaultStreamID)
	}
	if defaultBurstSize != expectedDefaultBurstSize {
		t.Errorf("defaultBurstSize = %v, want %v", defaultBurstSize, expectedDefaultBurstSize)
	}
	if defaultInterBurstGapUs != expectedDefaultInterBurstGapUs {
		t.Errorf("defaultInterBurstGapUs = %v, want %v", defaultInterBurstGapUs, expectedDefaultInterBurstGapUs)
	}
}

// TestSafeUint32FromIntBoundary tests boundary conditions for safeUint32FromInt.
func TestSafeUint32FromIntBoundary(t *testing.T) {
	// Test the exact boundaries.
	tests := []struct {
		name     string
		value    int
		fallback uint32
		expected uint32
	}{
		{
			name:     "value at -1",
			value:    -1,
			fallback: 100,
			expected: 100,
		},
		{
			name:     "value at 0",
			value:    0,
			fallback: 100,
			expected: 0,
		},
		{
			name:     "value at 1",
			value:    1,
			fallback: 100,
			expected: 1,
		},
		{
			name:     "value at MaxUint32",
			value:    math.MaxUint32,
			fallback: 100,
			expected: math.MaxUint32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeUint32FromInt(tt.value, tt.fallback)
			if result != tt.expected {
				t.Errorf("safeUint32FromInt(%d, %d) = %d, want %d",
					tt.value, tt.fallback, result, tt.expected)
			}
		})
	}
}

// TestSafeUint16Boundary tests boundary conditions for safeUint16.
func TestSafeUint16Boundary(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected uint16
	}{
		{
			name:     "value at MaxUint16 - 1",
			value:    math.MaxUint16 - 1,
			expected: math.MaxUint16 - 1,
		},
		{
			name:     "value at MaxUint16",
			value:    math.MaxUint16,
			expected: math.MaxUint16,
		},
		{
			name:     "value at MaxUint16 + 1",
			value:    math.MaxUint16 + 1,
			expected: math.MaxUint16,
		},
		{
			name:     "value at MaxUint16 + 2",
			value:    math.MaxUint16 + 2,
			expected: math.MaxUint16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeUint16(tt.value)
			if result != tt.expected {
				t.Errorf("safeUint16(%d) = %d, want %d",
					tt.value, result, tt.expected)
			}
		})
	}
}

// TestSafeUint8Boundary tests boundary conditions for safeUint8.
func TestSafeUint8Boundary(t *testing.T) {
	tests := []struct {
		name     string
		value    uint32
		expected uint8
	}{
		{
			name:     "value at MaxUint8 - 1",
			value:    math.MaxUint8 - 1,
			expected: math.MaxUint8 - 1,
		},
		{
			name:     "value at MaxUint8",
			value:    math.MaxUint8,
			expected: math.MaxUint8,
		},
		{
			name:     "value at MaxUint8 + 1",
			value:    math.MaxUint8 + 1,
			expected: math.MaxUint8,
		},
		{
			name:     "value at MaxUint8 + 2",
			value:    math.MaxUint8 + 2,
			expected: math.MaxUint8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeUint8(tt.value)
			if result != tt.expected {
				t.Errorf("safeUint8(%d) = %d, want %d",
					tt.value, result, tt.expected)
			}
		})
	}
}

// containsSubstring checks if str contains substr.
func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
