// SPDX-License-Identifier: BUSL-1.1

package benchmark_test

import (
	"errors"
	"testing"

	"github.com/krisarmstrong/stem/internal/services/benchmark"
	"github.com/krisarmstrong/stem/internal/services/modtypes"
	"github.com/krisarmstrong/stem/internal/services/orchestrator/dataplane"
)

// skipRealDataplaneInShort skips tests that drive the real CGO dataplane via
// dataplane.NewTestContext(). Under the stub build (CGO disabled) those calls
// return ErrNotSupported, which is the path these tests are written for. Under
// the CGO+linux build (the race-detector job) they exercise the real C
// dataplane, which requires a NIC/privileges the CI runner lacks — the C
// trials fail with -22 (EINVAL) and the uninitialized test context panics.
// The race job runs `go test -short`, so honoring -short keeps these tests in
// their intended stub mode in the normal (CGO-off) backend job while skipping
// the hardware path in CI. Production runs on real hardware and is unaffected.
func skipRealDataplaneInShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("exercises the real CGO dataplane; requires NIC/hardware — skipped under -short")
	}
}

// TestGetLoadLevels tests the getLoadLevels helper function.
func TestGetLoadLevels(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	tests := []struct {
		name     string
		cfg      *modtypes.TestConfig
		expected []float64
	}{
		{
			name:     "nil params returns defaults",
			cfg:      &modtypes.TestConfig{Interface: "", FrameSize: 0, Duration: 0, Params: nil},
			expected: benchmark.DefaultLoadLevelsForTest(),
		},
		{
			name:     "empty params returns defaults",
			cfg:      &modtypes.TestConfig{Interface: "", FrameSize: 0, Duration: 0, Params: map[string]any{}},
			expected: benchmark.DefaultLoadLevelsForTest(),
		},
		{
			name: "custom load levels",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"load_levels": []float64{10, 50, 100},
				},
			},
			expected: []float64{10, 50, 100},
		},
		{
			name: "wrong type returns defaults",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"load_levels": "invalid",
				},
			},
			expected: benchmark.DefaultLoadLevelsForTest(),
		},
		{
			name: "int slice returns defaults (wrong type)",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"load_levels": []int{10, 50, 100},
				},
			},
			expected: benchmark.DefaultLoadLevelsForTest(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := benchmark.LoadLevelsForTest(exec, tt.cfg)
			if len(result) != len(tt.expected) {
				t.Errorf("getLoadLevels() returned %d elements, want %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("getLoadLevels()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

// TestGetFrameLossParams tests the getFrameLossParams helper function.
func TestGetFrameLossParams(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	tests := []struct {
		name         string
		cfg          *modtypes.TestConfig
		expectedS    float64
		expectedE    float64
		expectedStep float64
	}{
		{
			name:         "nil params returns defaults",
			cfg:          &modtypes.TestConfig{Interface: "", FrameSize: 0, Duration: 0, Params: nil},
			expectedS:    benchmark.DefaultStartPctForTest(),
			expectedE:    benchmark.DefaultEndPctForTest(),
			expectedStep: benchmark.DefaultStepPctForTest(),
		},
		{
			name:         "empty params returns defaults",
			cfg:          &modtypes.TestConfig{Interface: "", FrameSize: 0, Duration: 0, Params: map[string]any{}},
			expectedS:    benchmark.DefaultStartPctForTest(),
			expectedE:    benchmark.DefaultEndPctForTest(),
			expectedStep: benchmark.DefaultStepPctForTest(),
		},
		{
			name: "custom parameters",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"start_pct": 20.0,
					"end_pct":   80.0,
					"step_pct":  5.0,
				},
			},
			expectedS:    20.0,
			expectedE:    80.0,
			expectedStep: 5.0,
		},
		{
			name: "partial parameters",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"start_pct": 15.0,
				},
			},
			expectedS:    15.0,
			expectedE:    benchmark.DefaultEndPctForTest(),
			expectedStep: benchmark.DefaultStepPctForTest(),
		},
		{
			name: "int values convert",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"start_pct": 25,
					"end_pct":   90,
					"step_pct":  5,
				},
			},
			expectedS:    25.0,
			expectedE:    90.0,
			expectedStep: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, e, step := benchmark.FrameLossParamsForTest(exec, tt.cfg)
			if s != tt.expectedS {
				t.Errorf("start_pct = %v, want %v", s, tt.expectedS)
			}
			if e != tt.expectedE {
				t.Errorf("end_pct = %v, want %v", e, tt.expectedE)
			}
			if step != tt.expectedStep {
				t.Errorf("step_pct = %v, want %v", step, tt.expectedStep)
			}
		})
	}
}

// TestGetBackToBackParams tests the getBackToBackParams helper function.
func TestGetBackToBackParams(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	tests := []struct {
		name           string
		cfg            *modtypes.TestConfig
		expectedBurst  uint64
		expectedTrials uint32
	}{
		{
			name:           "nil params returns defaults",
			cfg:            &modtypes.TestConfig{Interface: "", FrameSize: 0, Duration: 0, Params: nil},
			expectedBurst:  benchmark.DefaultInitialBurstForTest(),
			expectedTrials: benchmark.DefaultTrialsForTest(),
		},
		{
			name:           "empty params returns defaults",
			cfg:            &modtypes.TestConfig{Interface: "", FrameSize: 0, Duration: 0, Params: map[string]any{}},
			expectedBurst:  benchmark.DefaultInitialBurstForTest(),
			expectedTrials: benchmark.DefaultTrialsForTest(),
		},
		{
			name: "custom parameters",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"initial_burst": uint64(50000),
					"trials":        uint32(5),
				},
			},
			expectedBurst:  50000,
			expectedTrials: 5,
		},
		{
			name: "float64 values convert",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"initial_burst": 25000.0,
					"trials":        10.0,
				},
			},
			expectedBurst:  25000,
			expectedTrials: 10,
		},
		{
			name: "int values convert",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"initial_burst": 15000,
					"trials":        7,
				},
			},
			expectedBurst:  15000,
			expectedTrials: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			burst, trials := benchmark.BackToBackParamsForTest(exec, tt.cfg)
			if burst != tt.expectedBurst {
				t.Errorf("initial_burst = %v, want %v", burst, tt.expectedBurst)
			}
			if trials != tt.expectedTrials {
				t.Errorf("trials = %v, want %v", trials, tt.expectedTrials)
			}
		})
	}
}

// TestGetRecoveryParams tests the getRecoveryParams helper function.
func TestGetRecoveryParams(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	tests := []struct {
		name             string
		cfg              *modtypes.TestConfig
		expectedThrough  float64
		expectedOverload uint32
	}{
		{
			name:             "nil params returns defaults",
			cfg:              &modtypes.TestConfig{Interface: "", FrameSize: 0, Duration: 0, Params: nil},
			expectedThrough:  benchmark.DefaultThroughputPctForTest(),
			expectedOverload: benchmark.DefaultOverloadSecForTest(),
		},
		{
			name:             "empty params returns defaults",
			cfg:              &modtypes.TestConfig{Interface: "", FrameSize: 0, Duration: 0, Params: map[string]any{}},
			expectedThrough:  benchmark.DefaultThroughputPctForTest(),
			expectedOverload: benchmark.DefaultOverloadSecForTest(),
		},
		{
			name: "custom parameters",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"throughput_pct": 75.0,
					"overload_sec":   uint32(120),
				},
			},
			expectedThrough:  75.0,
			expectedOverload: 120,
		},
		{
			name: "float64 values convert",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"throughput_pct": 50.5,
					"overload_sec":   90.0,
				},
			},
			expectedThrough:  50.5,
			expectedOverload: 90,
		},
		{
			name: "int values convert",
			cfg: &modtypes.TestConfig{
				Interface: "",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"throughput_pct": 80,
					"overload_sec":   45,
				},
			},
			expectedThrough:  80.0,
			expectedOverload: 45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			through, overload := benchmark.RecoveryParamsForTest(exec, tt.cfg)
			if through != tt.expectedThrough {
				t.Errorf("throughput_pct = %v, want %v", through, tt.expectedThrough)
			}
			if overload != tt.expectedOverload {
				t.Errorf("overload_sec = %v, want %v", overload, tt.expectedOverload)
			}
		})
	}
}

// TestExecuteInvalidTestTypeInternal tests invalid test type handling.
func TestExecuteInvalidTestTypeInternal(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil,
	}

	invalidTests := []string{
		"invalid",
		"y1564",
		"rfc2889_forwarding",
		"",
		"throughput",
	}

	for _, testType := range invalidTests {
		t.Run(testType, func(t *testing.T) {
			_, err := exec.Execute(testType, cfg)
			if err == nil {
				t.Errorf("Execute(%q) should return error", testType)
			}
		})
	}
}

// TestExecuteNilConfigInternal tests nil config handling.
func TestExecuteNilConfigInternal(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

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
			_, err := exec.Execute(testType, nil)
			if err == nil {
				t.Errorf("Execute(%q, nil) should return error", testType)
			}
			if !errors.Is(err, modtypes.ErrInvalidConfig) {
				t.Errorf("Expected ErrInvalidConfig, got: %v", err)
			}
		})
	}
}

// TestConfigureContextNilContext tests configureContext with nil context.
func TestConfigureContextNilContext(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  60,
		Params: map[string]any{
			"resolution": 0.1,
			"max_loss":   0.0,
			"warmup":     10,
		},
	}

	// With nil context, this will either panic or return an error.
	// Using recover to handle either case.
	defer func() {
		_ = recover() // Just capture any panic.
	}()

	err := benchmark.ConfigureContextForTest(exec, cfg)
	// If we get here without panic, verify it returned an error.
	if err == nil {
		t.Error("configureContext with nil ctx should return error or panic")
	}
}

// TestDefaultConstants verifies the default constant values.
func TestDefaultConstants(t *testing.T) {
	// Verify default constants are reasonable values.
	defaultResolution := benchmark.DefaultResolutionForTest()
	if defaultResolution <= 0 || defaultResolution >= 100 {
		t.Errorf("DefaultResolutionForTest() = %v, should be between 0 and 100", defaultResolution)
	}

	defaultAcceptableLoss := benchmark.DefaultAcceptableLossForTest()
	if defaultAcceptableLoss < 0 || defaultAcceptableLoss >= 100 {
		t.Errorf("DefaultAcceptableLossForTest() = %v, should be between 0 and 100", defaultAcceptableLoss)
	}

	defaultStartPct := benchmark.DefaultStartPctForTest()
	if defaultStartPct <= 0 || defaultStartPct >= 100 {
		t.Errorf("DefaultStartPctForTest() = %v, should be between 0 and 100", defaultStartPct)
	}

	defaultEndPct := benchmark.DefaultEndPctForTest()
	if defaultEndPct <= 0 || defaultEndPct > 100 {
		t.Errorf("DefaultEndPctForTest() = %v, should be between 0 and 100", defaultEndPct)
	}

	defaultStepPct := benchmark.DefaultStepPctForTest()
	if defaultStepPct <= 0 || defaultStepPct > 100 {
		t.Errorf("DefaultStepPctForTest() = %v, should be between 0 and 100", defaultStepPct)
	}

	defaultInitialBurst := benchmark.DefaultInitialBurstForTest()
	if defaultInitialBurst == 0 {
		t.Error("DefaultInitialBurstForTest() should not be 0")
	}

	defaultTrials := benchmark.DefaultTrialsForTest()
	if defaultTrials == 0 {
		t.Error("DefaultTrialsForTest() should not be 0")
	}

	defaultThroughputPct := benchmark.DefaultThroughputPctForTest()
	if defaultThroughputPct <= 0 || defaultThroughputPct > 100 {
		t.Errorf("DefaultThroughputPctForTest() = %v, should be between 0 and 100", defaultThroughputPct)
	}

	defaultOverloadSec := benchmark.DefaultOverloadSecForTest()
	if defaultOverloadSec == 0 {
		t.Error("DefaultOverloadSecForTest() should not be 0")
	}

	// Verify default load levels.
	defaultLevels := benchmark.DefaultLoadLevelsForTest()
	if len(defaultLevels) == 0 {
		t.Error("defaultLoadLevels should not be empty")
	}
	for i, level := range defaultLevels {
		if level <= 0 || level > 100 {
			t.Errorf("defaultLoadLevels[%d] = %v, should be between 0 and 100", i, level)
		}
	}
}

// TestExecutorWithNilContext tests executor behavior with nil context.
func TestExecutorWithNilContext(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	// Module methods should work.
	if exec.Name() != benchmark.ModuleName {
		t.Errorf("Name() = %q, want %q", exec.Name(), benchmark.ModuleName)
	}
	if exec.DisplayName() != benchmark.DisplayName {
		t.Errorf("benchmark.DisplayName() = %q, want %q", exec.DisplayName(), benchmark.DisplayName)
	}
	if !exec.SupportsExecution() {
		t.Error("SupportsExecution() should return true")
	}
	if !exec.CanRun("rfc2544_throughput") {
		t.Error("CanRun('rfc2544_throughput') should return true")
	}

	// Close should not panic with nil context.
	exec.Close()
}

// verifyErrorResult checks result fields when an error is expected.
func verifyErrorResult(t *testing.T, result *modtypes.Result, testType string) {
	t.Helper()
	if result == nil {
		return
	}
	if result.TestType != testType {
		t.Errorf("Result.TestType = %q, want %q", result.TestType, testType)
	}
	if result.ModuleName != benchmark.ModuleName {
		t.Errorf("Result.benchmark.ModuleName = %q, want %q", result.ModuleName, benchmark.ModuleName)
	}
	if result.Success {
		t.Error("Result.Success should be false when configureContext fails")
	}
	if result.Error == "" {
		t.Error("Result.Error should be populated when configureContext fails")
	}
}

// TestExecuteConfigureContextError tests Execute when configureContext fails.
func TestExecuteConfigureContextError(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  60,
		Params:    nil,
	}

	// Execute should fail because configureContext will fail with nil ctx.
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
			// Use recover in case configureContext panics.
			defer func() {
				_ = recover()
			}()

			result, err := exec.Execute(testType, cfg)
			if err == nil {
				t.Error("Execute should return error when configureContext fails")
			}
			verifyErrorResult(t, result, testType)
		})
	}
}

// TestConfigureContextWithDuration tests configureContext with various durations.
func TestConfigureContextWithDuration(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	tests := []struct {
		name     string
		duration int
	}{
		{"zero duration", 0},
		{"positive duration", 60},
		{"large duration", 3600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  tt.duration,
				Params:    nil,
			}

			// Will panic or return error due to nil ctx.
			defer func() {
				_ = recover()
			}()

			_ = benchmark.ConfigureContextForTest(exec, cfg)
		})
	}
}

// TestConfigureContextWithWarmup tests configureContext with various warmup values.
func TestConfigureContextWithWarmupHelper(t *testing.T) {
	exec := benchmark.NewExecutorWithContext(nil)

	tests := []struct {
		name   string
		warmup int
	}{
		{"zero warmup", 0},
		{"positive warmup", 30},
		{"negative warmup", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  60,
				Params: map[string]any{
					"warmup": tt.warmup,
				},
			}

			// Will panic or return error due to nil ctx.
			defer func() {
				_ = recover()
			}()

			_ = benchmark.ConfigureContextForTest(exec, cfg)
		})
	}
}

// TestExecuteWithTestContext tests Execute with a valid test context.
// This allows testing all the code paths in Execute without needing Linux/CGO.
func TestExecuteWithTestContext(t *testing.T) {
	skipRealDataplaneInShort(t)
	// Use NewTestContext which returns a valid but non-functional context.
	ctx := dataplane.NewTestContext()
	exec := benchmark.NewExecutorWithContext(ctx)
	defer exec.Close()

	testCases := []struct {
		testType string
		params   map[string]any
	}{
		{
			testType: "rfc2544_throughput",
			params:   nil,
		},
		{
			testType: "rfc2544_latency",
			params: map[string]any{
				"load_levels": []float64{10, 50, 100},
			},
		},
		{
			testType: "rfc2544_latency",
			params:   nil, // Test default load levels.
		},
		{
			testType: "rfc2544_frame_loss",
			params: map[string]any{
				"start_pct": 10.0,
				"end_pct":   100.0,
				"step_pct":  10.0,
			},
		},
		{
			testType: "rfc2544_back_to_back",
			params: map[string]any{
				"initial_burst": uint64(10000),
				"trials":        uint32(3),
			},
		},
		{
			testType: "rfc2544_system_recovery",
			params: map[string]any{
				"throughput_pct": 100.0,
				"overload_sec":   uint32(60),
			},
		},
		{
			testType: "rfc2544_reset",
			params:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testType, func(t *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  10,
				Params:    tc.params,
			}

			result, err := exec.Execute(tc.testType, cfg)
			verifyExecuteOutcome(t, result, err, tc.testType)
		})
	}
}

// verifyExecuteOutcome checks the result based on error presence.
func verifyExecuteOutcome(t *testing.T, result *modtypes.Result, err error, testType string) {
	t.Helper()
	if err == nil {
		verifySuccessResult(t, result)
		return
	}
	verifyFailureResult(t, result, testType)
}

// verifySuccessResult checks result when no error occurred.
func verifySuccessResult(t *testing.T, result *modtypes.Result) {
	t.Helper()
	if result == nil {
		t.Error("Execute returned nil result without error")
		return
	}
	if !result.Success {
		t.Error("Result.Success should be true when no error")
	}
}

// verifyFailureResult checks result when an error occurred.
func verifyFailureResult(t *testing.T, result *modtypes.Result, testType string) {
	t.Helper()
	if result == nil {
		return
	}
	if result.TestType != testType {
		t.Errorf("Result.TestType = %q, want %q", result.TestType, testType)
	}
	if result.ModuleName != benchmark.ModuleName {
		t.Errorf("Result.benchmark.ModuleName = %q, want %q", result.ModuleName, benchmark.ModuleName)
	}
	if result.Success {
		t.Error("Result.Success should be false when error")
	}
	if result.Error == "" {
		t.Error("Result.Error should not be empty when error")
	}
}

// TestExecuteWithTestContextFrameSize tests Execute with various frame sizes.
func TestExecuteWithTestContextFrameSize(t *testing.T) {
	skipRealDataplaneInShort(t)
	ctx := dataplane.NewTestContext()
	exec := benchmark.NewExecutorWithContext(ctx)
	defer exec.Close()

	frameSizes := []uint32{0, 64, 128, 256, 512, 1024, 1280, 1518, 9000}

	for _, size := range frameSizes {
		t.Run("frameSize", func(_ *testing.T) {
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

// TestExecuteWithTestContextDefaultBranch tests the default case in Execute switch.
func TestExecuteWithTestContextDefaultBranch(t *testing.T) {
	skipRealDataplaneInShort(t)
	ctx := dataplane.NewTestContext()
	exec := benchmark.NewExecutorWithContext(ctx)
	defer exec.Close()

	// This should never happen because CanRun check comes first,
	// but we need to test the default branch.
	// We can't directly test this without bypassing CanRun,
	// but we can verify it's unreachable with valid inputs.
	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil,
	}

	// Invalid test types are caught by CanRun first.
	_, err := exec.Execute("nonexistent_test", cfg)
	if err == nil {
		t.Error("Execute with invalid test type should return error")
	}
}

// TestCloseWithTestContext tests Close with a valid context.
func TestCloseWithTestContext(_ *testing.T) {
	ctx := dataplane.NewTestContext()
	exec := benchmark.NewExecutorWithContext(ctx)

	// Close should not panic.
	exec.Close()

	// Verify ctx is cleared after close (if implemented).
	// Calling close again should still not panic.
	exec.Close()
}

// TestNewExecutorWithTestContext tests NewExecutorWithContext with test context.
func TestNewExecutorWithTestContext(t *testing.T) {
	ctx := dataplane.NewTestContext()
	exec := benchmark.NewExecutorWithContext(ctx)

	if exec == nil {
		t.Fatal("NewExecutorWithContext returned nil")
	}
	if exec.Name() != benchmark.ModuleName {
		t.Errorf("Name() = %q, want %q", exec.Name(), benchmark.ModuleName)
	}

	defer exec.Close()
}

// TestExecutorSupportsExecutionWithTestContext tests SupportsExecution.
func TestExecutorSupportsExecutionWithTestContext(t *testing.T) {
	ctx := dataplane.NewTestContext()
	exec := benchmark.NewExecutorWithContext(ctx)
	defer exec.Close()

	if !exec.SupportsExecution() {
		t.Error("SupportsExecution() returned false")
	}
}

// TestConfigureContextWithTestContext tests configureContext with valid context.
func TestConfigureContextWithTestContext(t *testing.T) {
	skipRealDataplaneInShort(t)
	ctx := dataplane.NewTestContext()
	exec := benchmark.NewExecutorWithContext(ctx)
	defer exec.Close()

	tests := []struct {
		name string
		cfg  *modtypes.TestConfig
	}{
		{
			name: "basic config",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  60,
				Params:    nil,
			},
		},
		{
			name: "with duration",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  120,
				Params:    nil,
			},
		},
		{
			name: "zero duration",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  0,
				Params:    nil,
			},
		},
		{
			name: "with params",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  60,
				Params: map[string]any{
					"resolution": 0.1,
					"max_loss":   0.0,
				},
			},
		},
		{
			name: "with warmup",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  60,
				Params: map[string]any{
					"warmup": 30,
				},
			},
		},
		{
			name: "zero warmup",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  60,
				Params: map[string]any{
					"warmup": 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := benchmark.ConfigureContextForTest(exec, tt.cfg)
			// In stub mode, Configure returns ErrNotSupported.
			if err == nil {
				t.Log("configureContext succeeded (not expected in stub mode)")
			} else if err.Error() == "" {
				// Verify the error is wrapped properly.
				t.Error("configureContext error should have message")
			}
		})
	}
}

// verifyExecutorSuccess checks a successfully created executor.
func verifyExecutorSuccess(t *testing.T, exec *benchmark.Executor) {
	t.Helper()
	if exec == nil {
		t.Error("NewExecutor returned nil without error")
		return
	}
	defer exec.Close()
	if exec.Name() != benchmark.ModuleName {
		t.Errorf("Name() = %q, want %q", exec.Name(), benchmark.ModuleName)
	}
}

// verifyExecutorFailure checks executor state when creation failed.
func verifyExecutorFailure(t *testing.T, exec *benchmark.Executor, err error) {
	t.Helper()
	if exec != nil {
		t.Error("NewExecutor returned non-nil executor with error")
	}
	if err.Error() == "" {
		t.Error("NewExecutor error should have message")
	}
}

// TestNewExecutorError tests NewExecutor error handling.
func TestNewExecutorError(t *testing.T) {
	// On non-Linux/non-CGO platforms, NewExecutor always fails.
	exec, err := benchmark.NewExecutor("lo")
	if err == nil {
		verifyExecutorSuccess(t, exec)
	} else {
		verifyExecutorFailure(t, exec, err)
	}
}

// TestNewExecutorDifferentInterfaces tests NewExecutor with different interface names.
func TestNewExecutorDifferentInterfaces(t *testing.T) {
	interfaces := []string{"lo", "eth0", "en0", "wlan0", "nonexistent"}

	for _, iface := range interfaces {
		t.Run(iface, func(_ *testing.T) {
			exec, err := benchmark.NewExecutor(iface)
			if err == nil {
				defer exec.Close()
			}
			// Either error or success is valid depending on platform.
		})
	}
}
