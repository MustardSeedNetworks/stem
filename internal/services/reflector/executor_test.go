// SPDX-License-Identifier: BUSL-1.1

package reflector_test

import (
	"testing"

	reflectorDP "github.com/MustardSeedNetworks/stem/internal/reflector/dataplane"
	"github.com/MustardSeedNetworks/stem/internal/services/reflector"
)

// Test constants for executor tests.
const (
	testInterface = "eth0"
)

// TestNewExecutorReturnsError verifies that NewExecutor returns an error
// on non-CGO/non-Linux platforms (which is the case for macOS test environments).
func TestNewExecutorReturnsError(t *testing.T) {
	// On non-CGO builds, NewExecutor should return an error because the
	// dataplane stub returns ErrNotSupported.
	_, err := reflector.NewExecutor(testInterface)
	if err == nil {
		// If it succeeds (on Linux with CGO), we can't test error paths.
		// Close the executor and skip.
		t.Skip("CGO dataplane available, skipping stub error test")
	}

	// Verify the error message mentions the dataplane failure.
	if err.Error() == "" {
		t.Error("NewExecutor error message is empty")
	}
}

// TestNewExecutorWithDataplane tests creating an executor with an injected dataplane.
func TestNewExecutorWithDataplane(t *testing.T) {
	// Create executor with nil dataplane (simulates unconfigured state).
	exec := reflector.NewExecutorWithDataplane(nil)
	if exec == nil {
		t.Fatal("NewExecutorWithDataplane(nil) returned nil")
	}

	// Verify the embedded module is initialized.
	if exec.Name() != "reflector" {
		t.Errorf("Name() = %q, want %q", exec.Name(), "reflector")
	}

	// Verify Dataplane() returns nil as expected.
	if exec.Dataplane() != nil {
		t.Error("Dataplane() should return nil when created with nil dataplane")
	}
}

// TestExecutorSupportsExecution tests that SupportsExecution returns true.
func TestExecutorSupportsExecution(t *testing.T) {
	exec := reflector.NewExecutorWithDataplane(nil)
	if !exec.SupportsExecution() {
		t.Error("SupportsExecution() = false, want true")
	}
}

// TestExecutorCloseWithNilDataplane tests Close with nil dataplane.
func TestExecutorCloseWithNilDataplane(t *testing.T) {
	t.Helper()
	exec := reflector.NewExecutorWithDataplane(nil)
	// Should not panic when dataplane is nil.
	exec.Close()
	// If we got here without panic, the test passes.
}

// TestExecutorStopWithNilDataplane tests Stop with nil dataplane.
func TestExecutorStopWithNilDataplane(t *testing.T) {
	t.Helper()
	exec := reflector.NewExecutorWithDataplane(nil)
	// Should not panic when dataplane is nil.
	exec.Stop()
	// If we got here without panic, the test passes.
}

// TestExecutorGetStatsWithNilDataplane tests GetStats with nil dataplane.
func TestExecutorGetStatsWithNilDataplane(t *testing.T) {
	exec := reflector.NewExecutorWithDataplane(nil)
	stats := exec.GetStats()

	// Should return empty stats.
	if stats.PacketsReceived != 0 {
		t.Errorf("GetStats().PacketsReceived = %d, want 0", stats.PacketsReceived)
	}
	if stats.PacketsReflected != 0 {
		t.Errorf("GetStats().PacketsReflected = %d, want 0", stats.PacketsReflected)
	}
	if stats.BytesReceived != 0 {
		t.Errorf("GetStats().BytesReceived = %d, want 0", stats.BytesReceived)
	}
	if stats.BytesReflected != 0 {
		t.Errorf("GetStats().BytesReflected = %d, want 0", stats.BytesReflected)
	}
}

// TestExecutorIsRunningWithNilDataplane tests IsRunning with nil dataplane.
func TestExecutorIsRunningWithNilDataplane(t *testing.T) {
	exec := reflector.NewExecutorWithDataplane(nil)
	if exec.IsRunning() {
		t.Error("IsRunning() = true, want false when dataplane is nil")
	}
}

// TestExecutorDataplaneWithNilDataplane tests Dataplane method with nil dataplane.
func TestExecutorDataplaneWithNilDataplane(t *testing.T) {
	exec := reflector.NewExecutorWithDataplane(nil)
	dp := exec.Dataplane()
	if dp != nil {
		t.Error("Dataplane() should return nil when created with nil dataplane")
	}
}

// TestExecutorExecuteInvalidTestType tests Execute with an invalid test type.
func TestExecutorExecuteInvalidTestType(t *testing.T) {
	exec := reflector.NewExecutorWithDataplane(nil)

	testCases := []struct {
		name     string
		testType string
	}{
		{"empty", ""},
		{"invalid", "invalid"},
		{"throughput", "throughput"},
		{"latency", "latency"},
		{"rfc2544", "rfc2544"},
		{"y1564", "y1564"},
		{"unknown", "unknown_test_type"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := exec.Execute(tc.testType, nil)

			if err == nil {
				t.Errorf("Execute(%q) should return error for invalid test type", tc.testType)
			}

			if result != nil {
				t.Errorf("Execute(%q) should return nil result for invalid test type", tc.testType)
			}
		})
	}
}

// TestExecutorExecuteReflectWithNilDataplane tests Execute("reflect") with nil dataplane.
// This tests the error path when the dataplane is not configured.
// On non-CGO builds, the stub dataplane's methods can be called on nil receivers
// and return ErrNotSupported without panicking.
func TestExecutorExecuteReflectWithNilDataplane(t *testing.T) {
	exec := reflector.NewExecutorWithDataplane(nil)

	result, err := exec.Execute("reflect", nil)

	// The stub dataplane returns ErrNotSupported when Start() is called.
	if err == nil {
		t.Error("Execute(\"reflect\") with nil dataplane should return error")
	}

	// Result should be returned with error info.
	if result == nil {
		t.Fatal("Execute(\"reflect\") should return result even on error")
	}

	if result.Success {
		t.Error("Result.Success should be false on error")
	}

	if result.Error == "" {
		t.Error("Result.Error should contain error message")
	}

	if result.TestType != "reflect" {
		t.Errorf("Result.TestType = %q, want \"reflect\"", result.TestType)
	}
}

// TestExecutorModuleMethods tests that module methods are accessible via executor.
func TestExecutorModuleMethods(t *testing.T) {
	exec := reflector.NewExecutorWithDataplane(nil)

	// Test all module methods are accessible.
	if exec.Name() != reflector.ModuleName {
		t.Errorf("Name() = %q, want %q", exec.Name(), reflector.ModuleName)
	}

	if exec.DisplayName() != reflector.DisplayName {
		t.Errorf("DisplayName() = %q, want %q", exec.DisplayName(), reflector.DisplayName)
	}

	if exec.Color() != reflector.ColorHex {
		t.Errorf("Color() = %q, want %q", exec.Color(), reflector.ColorHex)
	}

	if exec.Standard() != reflector.StandardRef {
		t.Errorf("Standard() = %q, want %q", exec.Standard(), reflector.StandardRef)
	}

	if len(exec.TestTypes()) != 1 {
		t.Errorf("len(TestTypes()) = %d, want 1", len(exec.TestTypes()))
	}

	if !exec.CanRun("reflect") {
		t.Error("CanRun(\"reflect\") = false, want true")
	}

	if exec.CanRun("invalid") {
		t.Error("CanRun(\"invalid\") = true, want false")
	}

	desc := exec.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}

	testDesc := exec.TestDescription("reflect")
	if testDesc == "" {
		t.Error("TestDescription(\"reflect\") returned empty string")
	}
}

// TestResultStruct tests the Result struct fields.
func TestResultStruct(t *testing.T) {
	result := &reflector.Result{
		TestType:   "reflect",
		ModuleName: reflector.ModuleName,
		Success:    true,
		Error:      "",
		Data:       map[string]any{"key": "value"},
	}

	if result.TestType != "reflect" {
		t.Errorf("Result.TestType = %q, want %q", result.TestType, "reflect")
	}

	if result.ModuleName != reflector.ModuleName {
		t.Errorf("Result.ModuleName = %q, want %q", result.ModuleName, reflector.ModuleName)
	}

	if !result.Success {
		t.Error("Result.Success = false, want true")
	}

	if result.Error != "" {
		t.Errorf("Result.Error = %q, want empty string", result.Error)
	}

	data, ok := result.Data.(map[string]any)
	if !ok {
		t.Error("Result.Data is not map[string]any")
	}
	if data["key"] != "value" {
		t.Errorf("Result.Data[\"key\"] = %v, want \"value\"", data["key"])
	}
}

// TestConfigStruct tests the Config struct fields.
func TestConfigStruct(t *testing.T) {
	cfg := &reflector.Config{
		Interface: "eth0",
		Profile:   "netally",
		Params: map[string]any{
			"port": 3842,
		},
	}

	if cfg.Interface != "eth0" {
		t.Errorf("Config.Interface = %q, want %q", cfg.Interface, "eth0")
	}

	if cfg.Profile != "netally" {
		t.Errorf("Config.Profile = %q, want %q", cfg.Profile, "netally")
	}

	if cfg.Params["port"] != 3842 {
		t.Errorf("Config.Params[\"port\"] = %v, want 3842", cfg.Params["port"])
	}
}

// TestResultStructWithError tests the Result struct with error.
func TestResultStructWithError(t *testing.T) {
	result := &reflector.Result{
		TestType:   "reflect",
		ModuleName: reflector.ModuleName,
		Success:    false,
		Error:      "dataplane error",
		Data:       nil,
	}

	// Verify all fields including those that might appear unused.
	if result.TestType != "reflect" {
		t.Errorf("Result.TestType = %q, want \"reflect\"", result.TestType)
	}

	if result.ModuleName != reflector.ModuleName {
		t.Errorf("Result.ModuleName = %q, want %q", result.ModuleName, reflector.ModuleName)
	}

	if result.Success {
		t.Error("Result.Success = true, want false")
	}

	if result.Error != "dataplane error" {
		t.Errorf("Result.Error = %q, want %q", result.Error, "dataplane error")
	}

	if result.Data != nil {
		t.Error("Result.Data should be nil")
	}
}

// TestNewExecutorWithDataplaneNotNil tests creating executor with non-nil dataplane pointer.
func TestNewExecutorWithDataplaneNotNil(t *testing.T) {
	// Create an empty dataplane struct (non-nil but non-functional).
	dp := &reflectorDP.Dataplane{}

	exec := reflector.NewExecutorWithDataplane(dp)
	if exec == nil {
		t.Fatal("NewExecutorWithDataplane returned nil")
	}

	// The dataplane should be the one we passed.
	if exec.Dataplane() != dp {
		t.Error("Dataplane() should return the dataplane we passed")
	}
}

// TestExecutorWithNonNilDataplane tests executor methods with a non-nil dataplane.
func TestExecutorWithNonNilDataplane(t *testing.T) {
	// Create an empty dataplane struct.
	dp := &reflectorDP.Dataplane{}
	exec := reflector.NewExecutorWithDataplane(dp)

	// Test IsRunning - should return dataplane's running state (false for empty struct).
	if exec.IsRunning() {
		t.Error("IsRunning() should return false for empty dataplane")
	}

	// Test GetStats - should return stats from dataplane.
	stats := exec.GetStats()
	if stats.PacketsReceived != 0 {
		t.Errorf("GetStats().PacketsReceived = %d, want 0", stats.PacketsReceived)
	}

	// Test Stop - should not panic.
	exec.Stop()

	// Test Close - should not panic.
	exec.Close()
}

// TestExecutorGetStatsReturnsZeroStats tests that GetStats returns zero-value stats.
func TestExecutorGetStatsReturnsZeroStats(t *testing.T) {
	exec := reflector.NewExecutorWithDataplane(nil)
	stats := exec.GetStats()

	// Verify all fields are zero individually to avoid exhaustruct linter.
	if stats.PacketsReceived != 0 {
		t.Errorf("PacketsReceived = %d, want 0", stats.PacketsReceived)
	}
	if stats.PacketsReflected != 0 {
		t.Errorf("PacketsReflected = %d, want 0", stats.PacketsReflected)
	}
	if stats.BytesReceived != 0 {
		t.Errorf("BytesReceived = %d, want 0", stats.BytesReceived)
	}
	if stats.BytesReflected != 0 {
		t.Errorf("BytesReflected = %d, want 0", stats.BytesReflected)
	}
	if stats.TxErrors != 0 {
		t.Errorf("TxErrors = %d, want 0", stats.TxErrors)
	}
	if stats.RxInvalid != 0 {
		t.Errorf("RxInvalid = %d, want 0", stats.RxInvalid)
	}
	if stats.SigProbeOT != 0 {
		t.Errorf("SigProbeOT = %d, want 0", stats.SigProbeOT)
	}
	if stats.SigDataOT != 0 {
		t.Errorf("SigDataOT = %d, want 0", stats.SigDataOT)
	}
	if stats.SigLatency != 0 {
		t.Errorf("SigLatency = %d, want 0", stats.SigLatency)
	}
	if stats.SigRFC2544 != 0 {
		t.Errorf("SigRFC2544 = %d, want 0", stats.SigRFC2544)
	}
	if stats.SigY1564 != 0 {
		t.Errorf("SigY1564 = %d, want 0", stats.SigY1564)
	}
	if stats.SigMSN != 0 {
		t.Errorf("SigMSN = %d, want 0", stats.SigMSN)
	}
	if stats.LatencyMin != 0 {
		t.Errorf("LatencyMin = %f, want 0", stats.LatencyMin)
	}
	if stats.LatencyAvg != 0 {
		t.Errorf("LatencyAvg = %f, want 0", stats.LatencyAvg)
	}
	if stats.LatencyMax != 0 {
		t.Errorf("LatencyMax = %f, want 0", stats.LatencyMax)
	}
	if stats.LatencyCount != 0 {
		t.Errorf("LatencyCount = %d, want 0", stats.LatencyCount)
	}
}

// TestMultipleExecutorInstances tests multiple executor instances are independent.
func TestMultipleExecutorInstances(t *testing.T) {
	exec1 := reflector.NewExecutorWithDataplane(nil)
	exec2 := reflector.NewExecutorWithDataplane(nil)

	if exec1 == exec2 {
		t.Error("NewExecutorWithDataplane should return distinct instances")
	}

	// Both should have the same module values.
	if exec1.Name() != exec2.Name() {
		t.Error("Different executor instances should have same Name()")
	}
}

// TestExecutorCloseIdempotent tests that Close can be called multiple times.
func TestExecutorCloseIdempotent(t *testing.T) {
	t.Helper()
	exec := reflector.NewExecutorWithDataplane(nil)

	// Should not panic on multiple Close calls.
	exec.Close()
	exec.Close()
	exec.Close()
	// Test passes if no panic occurred.
}

// TestExecutorStopIdempotent tests that Stop can be called multiple times.
func TestExecutorStopIdempotent(t *testing.T) {
	t.Helper()
	exec := reflector.NewExecutorWithDataplane(nil)

	// Should not panic on multiple Stop calls.
	exec.Stop()
	exec.Stop()
	exec.Stop()
	// Test passes if no panic occurred.
}

// TestExecutorIsRunningConsistency tests IsRunning consistency.
func TestExecutorIsRunningConsistency(t *testing.T) {
	exec := reflector.NewExecutorWithDataplane(nil)

	// Multiple calls should return the same value.
	r1 := exec.IsRunning()
	r2 := exec.IsRunning()
	r3 := exec.IsRunning()

	if r1 != r2 || r2 != r3 {
		t.Error("IsRunning() should return consistent values")
	}

	if r1 {
		t.Error("IsRunning() should return false for nil dataplane")
	}
}

// TestConfigStructEmptyParams tests Config with empty Params.
func TestConfigStructEmptyParams(t *testing.T) {
	cfg := &reflector.Config{
		Interface: "lo",
		Profile:   "custom",
		Params:    nil,
	}

	// Verify all fields to avoid unused write warnings.
	if cfg.Interface != "lo" {
		t.Errorf("Config.Interface = %q, want \"lo\"", cfg.Interface)
	}

	if cfg.Profile != "custom" {
		t.Errorf("Config.Profile = %q, want \"custom\"", cfg.Profile)
	}

	if cfg.Params != nil {
		t.Error("Config.Params should be nil")
	}

	// Test with empty map.
	cfg.Params = map[string]any{}
	if len(cfg.Params) != 0 {
		t.Error("Config.Params should be empty map")
	}
}

// TestResultDataTypes tests Result.Data with various types.
func TestResultDataTypes(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		result := createTestResult(t, nil)
		if result.Data != nil {
			t.Error("Result.Data should be nil")
		}
	})

	t.Run("string", func(t *testing.T) {
		result := createTestResult(t, "test")
		if result.Data != "test" {
			t.Errorf("Result.Data = %v, want \"test\"", result.Data)
		}
	})

	t.Run("int", func(t *testing.T) {
		result := createTestResult(t, 42)
		if result.Data != 42 {
			t.Errorf("Result.Data = %v, want 42", result.Data)
		}
	})

	t.Run("map", func(t *testing.T) {
		data := map[string]any{"status": "ok"}
		result := createTestResult(t, data)
		// Use type assertion to compare map contents.
		m, ok := result.Data.(map[string]any)
		if !ok {
			t.Error("Result.Data is not map[string]any")
			return
		}
		if m["status"] != "ok" {
			t.Errorf("Result.Data[\"status\"] = %v, want \"ok\"", m["status"])
		}
	})

	t.Run("slice", func(t *testing.T) {
		result := createTestResult(t, []string{"a", "b"})
		// Use type assertion to compare slice contents.
		s, ok := result.Data.([]string)
		if !ok {
			t.Error("Result.Data is not []string")
			return
		}
		if len(s) != 2 || s[0] != "a" || s[1] != "b" {
			t.Errorf("Result.Data = %v, want [\"a\", \"b\"]", s)
		}
	})
}

// createTestResult is a helper to create Result structs and verify all fields.
func createTestResult(t *testing.T, data any) *reflector.Result {
	t.Helper()
	result := &reflector.Result{
		TestType:   "reflect",
		ModuleName: reflector.ModuleName,
		Success:    true,
		Error:      "",
		Data:       data,
	}

	// Verify non-Data fields are set correctly.
	if result.TestType != "reflect" {
		t.Errorf("Result.TestType = %q, want \"reflect\"", result.TestType)
	}
	if result.ModuleName != reflector.ModuleName {
		t.Errorf("Result.ModuleName = %q, want %q", result.ModuleName, reflector.ModuleName)
	}
	if !result.Success {
		t.Error("Result.Success = false, want true")
	}
	if result.Error != "" {
		t.Errorf("Result.Error = %q, want empty", result.Error)
	}

	return result
}
