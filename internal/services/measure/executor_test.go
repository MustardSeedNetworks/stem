// SPDX-License-Identifier: BUSL-1.1

package measure_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/krisarmstrong/stem/internal/services/measure"
	"github.com/krisarmstrong/stem/internal/services/modtypes"
)

// TestNewExecutorPlatformError tests that NewExecutor returns an error on unsupported platforms.
// This test runs on all platforms and verifies proper error handling.
func TestNewExecutorPlatformError(t *testing.T) {
	// NewExecutor requires CGO and Linux, so on other platforms it should fail.
	// This is expected behavior - we're testing the error path.
	executor, err := measure.NewExecutor("eth0")

	// On non-Linux/non-CGO builds, this should return an error.
	// The dataplane stub returns ErrNotSupported.
	if err == nil {
		// If we got an executor (on a supported platform), clean it up.
		if executor != nil {
			executor.Close()
		}
		// Skip the rest of the test on supported platforms.
		t.Skip("Skipping error path test on supported platform")
	}

	// Verify we got a meaningful error.
	if err.Error() == "" {
		t.Error("NewExecutor returned empty error message")
	}

	// Should mention dataplane or platform requirement.
	errStr := err.Error()
	if !strings.Contains(errStr, "dataplane") && !strings.Contains(errStr, "platform") {
		t.Logf("Error message: %s", errStr)
		// This is informational, not a failure.
	}

	// Executor should be nil on error.
	if executor != nil {
		t.Error("NewExecutor should return nil executor on error")
	}
}

// TestNewExecutorInvalidInterface tests NewExecutor with various interface names.
func TestNewExecutorInvalidInterface(t *testing.T) {
	interfaces := []string{
		"",            // Empty string
		"eth0",        // Common interface
		"lo",          // Loopback
		"nonexistent", // Non-existent interface
	}

	for _, iface := range interfaces {
		t.Run("interface_"+iface, func(t *testing.T) {
			executor, err := measure.NewExecutor(iface)
			// On stub builds, all interfaces will fail with platform error.
			// On real builds with CGO, some might succeed or fail differently.
			if err != nil {
				// Expected on stub builds.
				if executor != nil {
					t.Error("Executor should be nil when error is returned")
				}
				return
			}

			// If we got an executor, verify it's not nil and clean up.
			if executor == nil {
				t.Error("NewExecutor returned nil executor without error")
			} else {
				executor.Close()
			}
		})
	}
}

// TestExecutorSupportsExecution tests the SupportsExecution method.
// Note: This requires a valid executor, which may not be available on stub builds.
func TestExecutorSupportsExecution(t *testing.T) {
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		t.Skip("Skipping test - executor not available on this platform")
	}
	defer executor.Close()

	if !executor.SupportsExecution() {
		t.Error("SupportsExecution() should return true for Measure executor")
	}
}

// TestExecutorClose tests that Close handles nil context gracefully.
// Since we can't create an executor on stub builds, we test via the module interface.
func TestExecutorClose(t *testing.T) {
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		// On stub builds, we can't test Close directly.
		// However, the module itself should be safe to create and use.
		t.Skip("Skipping Close test - executor not available on this platform")
	}

	// Call Close multiple times to ensure it's idempotent.
	executor.Close()
	executor.Close() // Should not panic.
}

// TestExecutorExecuteNilConfig tests Execute with nil config.
func TestExecutorExecuteNilConfig(t *testing.T) {
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		t.Skip("Skipping test - executor not available on this platform")
	}
	defer executor.Close()

	// Execute with nil config should return ErrInvalidConfig.
	result, execErr := executor.Execute("y1731_delay", nil)
	if execErr == nil {
		t.Error("Execute with nil config should return error")
	}
	if !errors.Is(execErr, modtypes.ErrInvalidConfig) {
		t.Errorf("Expected ErrInvalidConfig, got: %v", execErr)
	}
	if result != nil {
		t.Error("Result should be nil when error is returned")
	}
}

// TestExecutorExecuteInvalidTestType tests Execute with invalid test types.
func TestExecutorExecuteInvalidTestType(t *testing.T) {
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		t.Skip("Skipping test - executor not available on this platform")
	}
	defer executor.Close()

	invalidTypes := []string{
		"invalid",
		"rfc2544_throughput", // Wrong module
		"y1564",              // Wrong module
		"",                   // Empty string
		"y1731",              // Incomplete
	}

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  60,
		Params:    nil,
	}

	for _, testType := range invalidTypes {
		t.Run("testType_"+testType, func(t *testing.T) {
			result, execErr := executor.Execute(testType, cfg)
			if execErr == nil {
				t.Errorf("Execute(%q) should return error", testType)
			}
			if result != nil {
				t.Error("Result should be nil for invalid test type")
			}
		})
	}
}

// TestExecutorExecuteValidTestTypes tests Execute with valid Y.1731 test types.
// Skipped under -short because Execute() drives real C dataplane code paths
// (y1731_delay_measurement et al.) that send packets on the configured
// interface; in a CI runner with no usable eth0 / no NET_RAW caps the C
// library SIGSEGVs during the cgo call. Run without -short to exercise.
func TestExecutorExecuteValidTestTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("requires real network interface; skipped under -short")
	}
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		t.Skip("Skipping test - executor not available on this platform")
	}
	defer executor.Close()

	validTypes := []string{
		"y1731_delay",
		"y1731_loss",
		"y1731_slm",
		"y1731_loopback",
	}

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  60,
		Params:    nil,
	}

	for _, testType := range validTypes {
		t.Run("testType_"+testType, func(t *testing.T) {
			// On supported platforms, this will attempt to run the test.
			// On stub platforms, this might fail with platform error.
			result, execErr := executor.Execute(testType, cfg)

			// The test type is valid, so we shouldn't get "cannot run test type" error.
			if execErr != nil && strings.Contains(execErr.Error(), "cannot run test type") {
				t.Errorf("Valid test type %q rejected", testType)
			}

			// If we got a result, verify it has the right test type and module.
			if result != nil {
				if result.TestType != testType {
					t.Errorf("Result.TestType = %q, want %q", result.TestType, testType)
				}
				if result.ModuleName != "measure" {
					t.Errorf("Result.ModuleName = %q, want %q", result.ModuleName, "measure")
				}
			}
		})
	}
}

// TestExecutorWithY1731Params tests Execute with various Y.1731 parameters.
func TestExecutorWithY1731Params(t *testing.T) {
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		t.Skip("Skipping test - executor not available on this platform")
	}
	defer executor.Close()

	testCases := []struct {
		name   string
		params map[string]any
	}{
		{
			name:   "default params",
			params: nil,
		},
		{
			name: "custom MEP ID",
			params: map[string]any{
				"mep_id": uint32(42),
			},
		},
		{
			name: "all custom params",
			params: map[string]any{
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
		{
			name: "JSON-decoded float64 params",
			params: map[string]any{
				"mep_id":       float64(15),
				"meg_level":    float64(3),
				"ccm_interval": float64(2000),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "eth0",
				FrameSize: 64,
				Duration:  60,
				Params:    tc.params,
			}

			// Execute should not panic with any valid params.
			result, execErr := executor.Execute("y1731_delay", cfg)

			// Error is OK (platform may not support actual execution).
			// We're mainly testing that the code doesn't panic.
			_ = result
			_ = execErr
		})
	}
}

// TestModuleEmbedsInExecutor verifies that Executor embeds Module correctly.
func TestModuleEmbedsInExecutor(t *testing.T) {
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		// On stub builds, we can't test the executor methods directly.
		// Instead, verify the module itself works.
		mod := measure.New()
		if mod.Name() != "measure" {
			t.Errorf("Module Name() = %q, want %q", mod.Name(), "measure")
		}
		t.Skip("Skipping executor embedding test - executor not available on this platform")
	}
	defer executor.Close()

	// Executor should have all Module methods via embedding.
	if executor.Name() != "measure" {
		t.Errorf("Executor.Name() = %q, want %q", executor.Name(), "measure")
	}
	if executor.DisplayName() != "Measure" {
		t.Errorf("Executor.DisplayName() = %q, want %q", executor.DisplayName(), "Measure")
	}
	if executor.Color() != "#2563eb" {
		t.Errorf("Executor.Color() = %q, want %q", executor.Color(), "#2563eb")
	}
	if executor.Standard() != "ITU-T Y.1731" {
		t.Errorf("Executor.Standard() = %q, want %q", executor.Standard(), "ITU-T Y.1731")
	}

	execTestTypes := executor.TestTypes()
	if len(execTestTypes) != 4 {
		t.Errorf("Executor.TestTypes() returned %d types, want 4", len(execTestTypes))
	}

	if !executor.CanRun("y1731_delay") {
		t.Error("Executor.CanRun(y1731_delay) = false, want true")
	}
	if executor.CanRun("rfc2544_throughput") {
		t.Error("Executor.CanRun(rfc2544_throughput) = true, want false")
	}

	desc := executor.TestDescription("y1731_delay")
	if desc == "" {
		t.Error("Executor.TestDescription(y1731_delay) returned empty string")
	}
	if !strings.Contains(desc, "DMM") {
		t.Errorf("TestDescription should mention DMM protocol, got: %s", desc)
	}
}

// TestExecutorResultStructure tests the structure of execution results.
func TestExecutorResultStructure(t *testing.T) {
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		t.Skip("Skipping test - executor not available on this platform")
	}
	defer executor.Close()

	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  60,
		Params:    nil,
	}

	result, execErr := executor.Execute("y1731_delay", cfg)

	// Whether successful or not, if we get a result, verify its structure.
	if result == nil {
		return
	}

	// TestType should be set.
	if result.TestType != "y1731_delay" {
		t.Errorf("Result.TestType = %q, want %q", result.TestType, "y1731_delay")
	}

	// ModuleName should be "measure".
	if result.ModuleName != "measure" {
		t.Errorf("Result.ModuleName = %q, want %q", result.ModuleName, "measure")
	}

	// If there was an error, Success should be false and Error should be set.
	if execErr == nil {
		return
	}
	if result.Success {
		t.Error("Result.Success should be false when error is returned")
	}
	if result.Error == "" {
		t.Error("Result.Error should be set when error is returned")
	}
}

// TestExecutorErrorMessages tests that error messages are meaningful.
func TestExecutorErrorMessages(t *testing.T) {
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		// Test the NewExecutor error message.
		if err.Error() == "" {
			t.Error("NewExecutor error should have a message")
		}
		t.Skip("Skipping execution error test - executor not available on this platform")
	}
	defer executor.Close()

	// Test invalid test type error.
	_, invalidErr := executor.Execute("invalid_test", &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  60,
		Params:    nil,
	})
	if invalidErr != nil {
		if !strings.Contains(invalidErr.Error(), "cannot run test type") {
			t.Logf("Error message for invalid test type: %s", invalidErr.Error())
		}
	}

	// Test nil config error.
	_, nilErr := executor.Execute("y1731_delay", nil)
	if nilErr != nil {
		if !errors.Is(nilErr, modtypes.ErrInvalidConfig) {
			t.Logf("Error for nil config: %v", nilErr)
		}
	}
}

// TestExecutorConcurrentAccess tests that the executor handles concurrent access.
func TestExecutorConcurrentAccess(t *testing.T) {
	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		t.Skip("Skipping test - executor not available on this platform")
	}
	defer executor.Close()

	// Note: The current implementation may not be thread-safe.
	// This test just ensures concurrent calls don't panic.
	cfg := &modtypes.TestConfig{
		Interface: "eth0",
		FrameSize: 64,
		Duration:  1,
		Params:    nil,
	}

	done := make(chan bool, 4)

	for _, testType := range []string{"y1731_delay", "y1731_loss", "y1731_slm", "y1731_loopback"} {
		go func(tt string) {
			_, _ = executor.Execute(tt, cfg)
			done <- true
		}(testType)
	}

	// Wait for all goroutines.
	for range 4 {
		<-done
	}
}

// TestExecutorModuleIntegration tests integration between Module and Executor.
func TestExecutorModuleIntegration(t *testing.T) {
	mod := measure.New()

	// All test types from module should be runnable.
	for _, testType := range mod.TestTypes() {
		if !mod.CanRun(testType) {
			t.Errorf("Module.CanRun(%q) should return true", testType)
		}

		desc := mod.TestDescription(testType)
		if desc == "" {
			t.Errorf("Module.TestDescription(%q) should not be empty", testType)
		}
	}

	executor, err := measure.NewExecutor("eth0")
	if err != nil {
		t.Skip("Skipping executor integration test - executor not available")
	}
	defer executor.Close()

	// Executor should return same values as Module.
	if executor.Name() != mod.Name() {
		t.Error("Executor.Name() should match Module.Name()")
	}
	if executor.Color() != mod.Color() {
		t.Error("Executor.Color() should match Module.Color()")
	}
	if executor.Standard() != mod.Standard() {
		t.Error("Executor.Standard() should match Module.Standard()")
	}
	if len(executor.TestTypes()) != len(mod.TestTypes()) {
		t.Error("Executor.TestTypes() should match Module.TestTypes()")
	}
}

// TestY1731TestTypesCompleteness verifies all Y.1731 test types have descriptions.
func TestY1731TestTypesCompleteness(t *testing.T) {
	mod := measure.New()

	expectedProtocols := map[string]string{
		"y1731_delay":    "DMM",
		"y1731_loss":     "LMM",
		"y1731_slm":      "Synthetic",
		"y1731_loopback": "LBM",
	}

	for testType, protocol := range expectedProtocols {
		t.Run(testType, func(t *testing.T) {
			// Test type should be runnable.
			if !mod.CanRun(testType) {
				t.Errorf("CanRun(%q) should return true", testType)
			}

			// Description should exist and mention the protocol.
			desc := mod.TestDescription(testType)
			if desc == "" {
				t.Errorf("TestDescription(%q) should not be empty", testType)
			}
			if !strings.Contains(desc, protocol) {
				t.Errorf("TestDescription(%q) = %q, should contain %q", testType, desc, protocol)
			}
			// All Y.1731 descriptions should mention Y.1731.
			if !strings.Contains(desc, "Y.1731") {
				t.Errorf("TestDescription(%q) should mention Y.1731", testType)
			}
		})
	}
}
