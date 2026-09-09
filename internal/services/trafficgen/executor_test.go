// SPDX-License-Identifier: BUSL-1.1

// Black-box testing of trafficgen executor.
//
// Tests verify public API behavior. The export_test.go file in package trafficgen
// provides NewMockExecutor, NewMockExecutorWithNilModule, and test constants.
package trafficgen_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/services/modtypes"
	"github.com/MustardSeedNetworks/stem/internal/services/trafficgen"
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
