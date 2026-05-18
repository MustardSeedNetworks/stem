// SPDX-License-Identifier: BUSL-1.1

package certify

import (
	"errors"
	"math"
	"testing"

	"github.com/krisarmstrong/stem/internal/services/modtypes"
	"github.com/krisarmstrong/stem/internal/services/orchestrator/dataplane"
)

// TestModtypesGetBoolParam tests the modtypes.GetBoolParam helper function.
func TestModtypesGetBoolParam(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]any
		key      string
		defVal   bool
		expected bool
	}{
		{
			name:     "nil params returns default true",
			params:   nil,
			key:      "test",
			defVal:   true,
			expected: true,
		},
		{
			name:     "nil params returns default false",
			params:   nil,
			key:      "test",
			defVal:   false,
			expected: false,
		},
		{
			name:     "missing key returns default",
			params:   map[string]any{"other": true},
			key:      "test",
			defVal:   false,
			expected: false,
		},
		{
			name:     "bool true value",
			params:   map[string]any{"test": true},
			key:      "test",
			defVal:   false,
			expected: true,
		},
		{
			name:     "bool false value",
			params:   map[string]any{"test": false},
			key:      "test",
			defVal:   true,
			expected: false,
		},
		{
			name:     "string value returns default",
			params:   map[string]any{"test": "true"},
			key:      "test",
			defVal:   false,
			expected: false,
		},
		{
			name:     "int value returns default",
			params:   map[string]any{"test": 1},
			key:      "test",
			defVal:   false,
			expected: false,
		},
		{
			name:     "float value returns default",
			params:   map[string]any{"test": 1.0},
			key:      "test",
			defVal:   true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modtypes.GetBoolParam(tt.params, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("modtypes.GetBoolParam() = %v, want %v", result, tt.expected)
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
			name:     "negative value returns 0",
			value:    -1,
			expected: 0,
		},
		{
			name:     "max int32",
			value:    math.MaxInt32,
			expected: math.MaxInt32,
		},
		{
			name:     "large positive value",
			value:    1000000,
			expected: 1000000,
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

// TestBuildRFC2889Config tests the buildRFC2889Config function.
func TestBuildRFC2889ConfigDefaults(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  0,
		Params:    nil,
	}

	config := buildRFC2889Config(cfg)
	if config == nil {
		t.Fatal("buildRFC2889Config returned nil")
	}
	if config.FrameSize != 64 {
		t.Errorf("FrameSize = %d, want 64", config.FrameSize)
	}
	if config.WarmupSec != defaultRFC2889WarmupSec {
		t.Errorf("WarmupSec = %d, want %d", config.WarmupSec, defaultRFC2889WarmupSec)
	}
	if config.AddressCount != defaultRFC2889AddressCt {
		t.Errorf("AddressCount = %d, want %d", config.AddressCount, defaultRFC2889AddressCt)
	}
	if config.DurationSec != defaultRFC2889Duration {
		t.Errorf("DurationSec = %d, want %d", config.DurationSec, defaultRFC2889Duration)
	}
}

func TestBuildRFC2889ConfigCustomDuration(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 128,
		Duration:  120,
		Params:    map[string]any{},
	}

	config := buildRFC2889Config(cfg)
	if config == nil {
		t.Fatal("buildRFC2889Config returned nil")
	}
	if config.DurationSec != 120 {
		t.Errorf("DurationSec = %d, want 120", config.DurationSec)
	}
}

func TestBuildRFC2889ConfigCustomParams(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 256,
		Duration:  0,
		Params: map[string]any{
			"warmup_sec":          uint32(10),
			"address_count":       uint32(4096),
			"acceptable_loss_pct": 0.5,
			"port_count":          uint32(4),
			"pattern":             uint32(1),
			"duration_sec":        uint32(90),
		},
	}

	config := buildRFC2889Config(cfg)
	if config == nil {
		t.Fatal("buildRFC2889Config returned nil")
	}
	if config.WarmupSec != 10 {
		t.Errorf("WarmupSec = %d, want 10", config.WarmupSec)
	}
	if config.AddressCount != 4096 {
		t.Errorf("AddressCount = %d, want 4096", config.AddressCount)
	}
	if config.AcceptableLossPct != 0.5 {
		t.Errorf("AcceptableLossPct = %f, want 0.5", config.AcceptableLossPct)
	}
	if config.PortCount != 4 {
		t.Errorf("PortCount = %d, want 4", config.PortCount)
	}
	if config.Pattern != 1 {
		t.Errorf("Pattern = %d, want 1", config.Pattern)
	}
	if config.DurationSec != 90 {
		t.Errorf("DurationSec = %d, want 90", config.DurationSec)
	}
}

// TestBuildRFC6349Config tests the buildRFC6349Config function.
func TestBuildRFC6349ConfigDefaults(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 1500,
		Duration:  0,
		Params:    nil,
	}

	config := buildRFC6349Config(cfg)
	if config == nil {
		t.Fatal("buildRFC6349Config returned nil")
	}
	if config.TargetRateMbps != defaultRFC6349TargetRate {
		t.Errorf("TargetRateMbps = %f, want %f", config.TargetRateMbps, defaultRFC6349TargetRate)
	}
	if config.MinRTTMs != defaultRFC6349MinRTTMs {
		t.Errorf("MinRTTMs = %f, want %f", config.MinRTTMs, defaultRFC6349MinRTTMs)
	}
	if config.MaxRTTMs != defaultRFC6349MaxRTTMs {
		t.Errorf("MaxRTTMs = %f, want %f", config.MaxRTTMs, defaultRFC6349MaxRTTMs)
	}
	if config.RWNDSize != defaultRFC6349RWND {
		t.Errorf("RWNDSize = %d, want %d", config.RWNDSize, defaultRFC6349RWND)
	}
	if config.MSS != defaultRFC6349MSS {
		t.Errorf("MSS = %d, want %d", config.MSS, defaultRFC6349MSS)
	}
	if config.DurationSec != defaultRFC6349Duration {
		t.Errorf("DurationSec = %d, want %d", config.DurationSec, defaultRFC6349Duration)
	}
}

func TestBuildRFC6349ConfigCustomDuration(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 1500,
		Duration:  60,
		Params:    map[string]any{},
	}

	config := buildRFC6349Config(cfg)
	if config == nil {
		t.Fatal("buildRFC6349Config returned nil")
	}
	if config.DurationSec != 60 {
		t.Errorf("DurationSec = %d, want 60", config.DurationSec)
	}
}

func TestBuildRFC6349ConfigCustomParams(t *testing.T) {
	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 1500,
		Duration:  0,
		Params: map[string]any{
			"target_rate_mbps": 1000.0,
			"min_rtt_ms":       0.5,
			"max_rtt_ms":       200.0,
			"rwnd_size":        uint32(131072),
			"parallel_streams": uint32(4),
			"mss":              uint32(8960),
			"mode":             uint32(1),
			"duration_sec":     uint32(45),
		},
	}

	config := buildRFC6349Config(cfg)
	if config == nil {
		t.Fatal("buildRFC6349Config returned nil")
	}
	if config.TargetRateMbps != 1000.0 {
		t.Errorf("TargetRateMbps = %f, want 1000.0", config.TargetRateMbps)
	}
	if config.MinRTTMs != 0.5 {
		t.Errorf("MinRTTMs = %f, want 0.5", config.MinRTTMs)
	}
	if config.MaxRTTMs != 200.0 {
		t.Errorf("MaxRTTMs = %f, want 200.0", config.MaxRTTMs)
	}
	if config.RWNDSize != 131072 {
		t.Errorf("RWNDSize = %d, want 131072", config.RWNDSize)
	}
	if config.ParallelStreams != 4 {
		t.Errorf("ParallelStreams = %d, want 4", config.ParallelStreams)
	}
	if config.MSS != 8960 {
		t.Errorf("MSS = %d, want 8960", config.MSS)
	}
	if config.Mode != 1 {
		t.Errorf("Mode = %d, want 1", config.Mode)
	}
	if config.DurationSec != 45 {
		t.Errorf("DurationSec = %d, want 45", config.DurationSec)
	}
}

func TestBuildTSNConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *modtypes.TestConfig
		validate func(t *testing.T, config *dataplane.TSNConfig)
	}{
		{
			name: "nil params uses defaults",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 0,
				Duration:  0,
				Params:    nil,
			},
			validate: assertTSNDefaults,
		},
		{
			name: "custom duration and frame size from fields",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 128,
				Duration:  120,
				Params:    map[string]any{},
			},
			validate: func(t *testing.T, config *dataplane.TSNConfig) {
				assertTSNDurationAndFrameSize(t, config, 120, 128)
			},
		},
		{
			name: "custom params",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 0,
				Duration:  0,
				Params: map[string]any{
					"warmup_sec":          uint32(10),
					"max_latency_ns":      uint32(500000),
					"max_jitter_ns":       uint32(50000),
					"require_ptp_sync":    false,
					"max_sync_offset_ns":  uint32(500),
					"ptp_enabled":         false,
					"preemption_enabled":  true,
					"num_traffic_classes": uint32(4),
					"base_time_ns":        uint64(1000000),
					"cycle_time_ns":       uint32(500000),
					"traffic_class":       uint32(3),
					"frame_size":          uint32(256),
					"duration_sec":        uint32(90),
				},
			},
			validate: assertTSNCustomParams,
		},
		{
			name: "frame_size fallback when FrameSize is 0",
			cfg: &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 0,
				Duration:  60,
				Params: map[string]any{
					"frame_size": uint32(512),
				},
			},
			validate: func(t *testing.T, config *dataplane.TSNConfig) {
				assertTSNFrameSize(t, config, 512)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := buildTSNConfig(tt.cfg)
			if config == nil {
				t.Fatal("buildTSNConfig returned nil")
			}
			tt.validate(t, config)
		})
	}
}

func assertTSNDefaults(t *testing.T, config *dataplane.TSNConfig) {
	t.Helper()

	if config.WarmupSec != defaultTSNWarmupSec {
		t.Errorf("WarmupSec = %d, want %d", config.WarmupSec, defaultTSNWarmupSec)
	}
	if config.MaxLatencyNs != defaultTSNMaxLatencyNs {
		t.Errorf("MaxLatencyNs = %d, want %d", config.MaxLatencyNs, defaultTSNMaxLatencyNs)
	}
	if config.MaxJitterNs != defaultTSNMaxJitterNs {
		t.Errorf("MaxJitterNs = %d, want %d", config.MaxJitterNs, defaultTSNMaxJitterNs)
	}
	if !config.RequirePTPSync {
		t.Error("RequirePTPSync should be true by default")
	}
	if !config.PTPEnabled {
		t.Error("PTPEnabled should be true by default")
	}
	if config.PreemptionEnabled {
		t.Error("PreemptionEnabled should be false by default")
	}
	if config.NumTrafficClasses != defaultTSNClassCount {
		t.Errorf("NumTrafficClasses = %d, want %d", config.NumTrafficClasses, defaultTSNClassCount)
	}
	if config.DurationSec != defaultTSNDuration {
		t.Errorf("DurationSec = %d, want %d", config.DurationSec, defaultTSNDuration)
	}
	if config.FrameSize != defaultTSNFrameSize {
		t.Errorf("FrameSize = %d, want %d", config.FrameSize, defaultTSNFrameSize)
	}
	if config.TrafficClass != defaultTSNTrafficClass {
		t.Errorf("TrafficClass = %d, want %d", config.TrafficClass, defaultTSNTrafficClass)
	}
}

func assertTSNDurationAndFrameSize(t *testing.T, config *dataplane.TSNConfig, duration, frameSize uint32) {
	t.Helper()

	if config.DurationSec != duration {
		t.Errorf("DurationSec = %d, want %d", config.DurationSec, duration)
	}
	if config.FrameSize != frameSize {
		t.Errorf("FrameSize = %d, want %d", config.FrameSize, frameSize)
	}
}

func assertTSNCustomParams(t *testing.T, config *dataplane.TSNConfig) {
	t.Helper()

	if config.WarmupSec != 10 {
		t.Errorf("WarmupSec = %d, want 10", config.WarmupSec)
	}
	if config.MaxLatencyNs != 500000 {
		t.Errorf("MaxLatencyNs = %d, want 500000", config.MaxLatencyNs)
	}
	if config.MaxJitterNs != 50000 {
		t.Errorf("MaxJitterNs = %d, want 50000", config.MaxJitterNs)
	}
	if config.RequirePTPSync {
		t.Error("RequirePTPSync should be false")
	}
	if config.MaxSyncOffsetNs != 500 {
		t.Errorf("MaxSyncOffsetNs = %d, want 500", config.MaxSyncOffsetNs)
	}
	if config.PTPEnabled {
		t.Error("PTPEnabled should be false")
	}
	if !config.PreemptionEnabled {
		t.Error("PreemptionEnabled should be true")
	}
	if config.NumTrafficClasses != 4 {
		t.Errorf("NumTrafficClasses = %d, want 4", config.NumTrafficClasses)
	}
	if config.BaseTimeNs != 1000000 {
		t.Errorf("BaseTimeNs = %d, want 1000000", config.BaseTimeNs)
	}
	if config.CycleTimeNs != 500000 {
		t.Errorf("CycleTimeNs = %d, want 500000", config.CycleTimeNs)
	}
	if config.TrafficClass != 3 {
		t.Errorf("TrafficClass = %d, want 3", config.TrafficClass)
	}
	if config.FrameSize != 256 {
		t.Errorf("FrameSize = %d, want 256", config.FrameSize)
	}
	if config.DurationSec != 90 {
		t.Errorf("DurationSec = %d, want 90", config.DurationSec)
	}
}

func assertTSNFrameSize(t *testing.T, config *dataplane.TSNConfig, frameSize uint32) {
	t.Helper()

	if config.FrameSize != frameSize {
		t.Errorf("FrameSize = %d, want %d", config.FrameSize, frameSize)
	}
}

// TestExecutorSupportsExecution tests the SupportsExecution method.
func TestExecutorSupportsExecution(t *testing.T) {
	exec := NewExecutorWithContext(nil)
	if !exec.SupportsExecution() {
		t.Error("SupportsExecution() returned false, expected true")
	}
}

// TestExecutorClose tests the Close method.
func TestExecutorClose(_ *testing.T) {
	// Test with nil context.
	exec := NewExecutorWithContext(nil)
	exec.Close() // Should not panic.

	// Call again to ensure idempotent.
	exec.Close()

	// Test with valid test context.
	testCtx := dataplane.NewTestContext()
	execWithCtx := NewExecutorWithContext(testCtx)
	execWithCtx.Close() // Should not panic and should call ctx.Close().
}

// TestExecuteCanRunValidation tests that Execute validates test type via CanRun.
func TestExecuteCanRunValidation(t *testing.T) {
	exec := NewExecutorWithContext(nil)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil,
	}

	invalidTests := []string{
		"invalid_test",
		"rfc2544_throughput",
		"y1564",
		"",
		"rfc2889", // Missing suffix.
	}

	for _, testType := range invalidTests {
		t.Run(testType, func(t *testing.T) {
			_, err := exec.Execute(testType, cfg)
			if err == nil {
				t.Errorf("Execute(%q) should return error for invalid test type", testType)
			}
		})
	}
}

// TestExecuteNilConfig tests that Execute returns ErrInvalidConfig for nil config.
func TestExecuteNilConfig(t *testing.T) {
	exec := NewExecutorWithContext(nil)

	validTests := []string{
		"rfc2889_forwarding",
		"rfc6349_throughput",
		"tsn_timing",
	}

	for _, testType := range validTests {
		t.Run(testType, func(t *testing.T) {
			_, err := exec.Execute(testType, nil)
			if !errors.Is(err, modtypes.ErrInvalidConfig) {
				t.Errorf("Expected ErrInvalidConfig, got: %v", err)
			}
		})
	}
}

// TestNewExecutorWithContext tests the NewExecutorWithContext function.
func TestNewExecutorWithContext(t *testing.T) {
	exec := NewExecutorWithContext(nil)
	if exec == nil {
		t.Fatal("NewExecutorWithContext returned nil")
	}

	if exec.Name() != ModuleName {
		t.Errorf("Expected module name %q, got %q", ModuleName, exec.Name())
	}

	if exec.Module == nil {
		t.Error("Executor Module field is nil")
	}
}

// TestNewExecutor tests the NewExecutor function.
// This test verifies error handling since NewContext returns error on non-Linux.
func TestNewExecutor(t *testing.T) {
	// On non-Linux/non-CGO builds, NewExecutor should return an error.
	exec, err := NewExecutor("lo")
	// In stub mode, we expect an error.
	if err != nil {
		// This is expected on non-Linux.
		if exec != nil {
			t.Error("NewExecutor returned non-nil executor with error")
		}
		return
	}

	// If we get here (Linux with CGO), verify the executor is valid.
	if exec == nil {
		t.Fatal("NewExecutor returned nil without error")
	}
	defer exec.Close()

	if exec.Name() != ModuleName {
		t.Errorf("Expected module name %q, got %q", ModuleName, exec.Name())
	}
}

// TestExecutorModuleEmbedding tests that the Executor properly embeds Module.
func TestExecutorModuleEmbedding(t *testing.T) {
	exec := NewExecutorWithContext(nil)

	// All Module methods should work.
	if exec.Name() != ModuleName {
		t.Errorf("Name() = %q, want %q", exec.Name(), ModuleName)
	}
	if exec.DisplayName() != DisplayName {
		t.Errorf("DisplayName() = %q, want %q", exec.DisplayName(), DisplayName)
	}
	if exec.Color() != ColorHex {
		t.Errorf("Color() = %q, want %q", exec.Color(), ColorHex)
	}
	if exec.Standard() != StandardRef {
		t.Errorf("Standard() = %q, want %q", exec.Standard(), StandardRef)
	}

	types := exec.TestTypes()
	if len(types) != 11 {
		t.Errorf("TestTypes() returned %d types, want 11", len(types))
	}

	if !exec.CanRun("rfc2889_forwarding") {
		t.Error("CanRun(rfc2889_forwarding) = false, want true")
	}
	if exec.CanRun("invalid") {
		t.Error("CanRun(invalid) = true, want false")
	}
}

// TestExecuteAllTestTypes tests Execute for all valid test types with a test context.
func TestExecuteAllTestTypes(t *testing.T) {
	testCtx := dataplane.NewTestContext()
	exec := NewExecutorWithContext(testCtx)

	allTests := []string{
		"rfc2889_forwarding",
		"rfc2889_caching",
		"rfc2889_learning",
		"rfc2889_broadcast",
		"rfc2889_congestion",
		"rfc6349_throughput",
		"rfc6349_path",
		"tsn_timing",
		"tsn_isolation",
		"tsn_latency",
		"tsn",
	}

	for _, testType := range allTests {
		t.Run(testType, func(t *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  10,
				Params:    map[string]any{},
			}

			result, err := exec.Execute(testType, cfg)

			// In stub mode, all tests return ErrNotSupported.
			// But the result struct should still be populated correctly.
			if result == nil {
				t.Fatal("Execute returned nil result")
			}

			if result.TestType != testType {
				t.Errorf("TestType = %q, want %q", result.TestType, testType)
			}

			if result.ModuleName != ModuleName {
				t.Errorf("ModuleName = %q, want %q", result.ModuleName, ModuleName)
			}

			// In stub mode, success should be false.
			if result.Success {
				t.Error("Success = true, want false (stub mode)")
			}

			// Error should be set in stub mode.
			if err == nil {
				t.Error("Expected error from stub mode")
			}

			if result.Error == "" {
				t.Error("Result.Error should be set")
			}
		})
	}
}

// TestExecuteRFC2889WithParams tests RFC 2889 execution with various parameters.
func TestExecuteRFC2889WithParams(t *testing.T) {
	testCtx := dataplane.NewTestContext()
	exec := NewExecutorWithContext(testCtx)

	testCases := []struct {
		name   string
		params map[string]any
	}{
		{
			name:   "default params",
			params: nil,
		},
		{
			name: "custom warmup",
			params: map[string]any{
				"warmup_sec": uint32(10),
			},
		},
		{
			name: "custom address count",
			params: map[string]any{
				"address_count": uint32(16384),
			},
		},
		{
			name: "acceptable loss pct",
			params: map[string]any{
				"acceptable_loss_pct": 0.5,
			},
		},
		{
			name: "acceptable loss fallback",
			params: map[string]any{
				"acceptable_loss_pct": 0.0,
				"acceptable_loss":     0.25,
			},
		},
		{
			name: "float64 params (JSON style)",
			params: map[string]any{
				"warmup_sec":    5.0,
				"address_count": 8192.0,
				"port_count":    4.0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 64,
				Duration:  60,
				Params:    tc.params,
			}

			result, _ := exec.Execute("rfc2889_forwarding", cfg)
			if result == nil {
				t.Fatal("Execute returned nil result")
			}

			if result.TestType != "rfc2889_forwarding" {
				t.Errorf("TestType = %q, want rfc2889_forwarding", result.TestType)
			}
		})
	}
}

// TestExecuteRFC6349WithParams tests RFC 6349 execution with various parameters.
func TestExecuteRFC6349WithParams(t *testing.T) {
	testCtx := dataplane.NewTestContext()
	exec := NewExecutorWithContext(testCtx)

	testCases := []struct {
		name   string
		params map[string]any
	}{
		{
			name:   "default params",
			params: nil,
		},
		{
			name: "custom target rate",
			params: map[string]any{
				"target_rate_mbps": 1000.0,
			},
		},
		{
			name: "custom RTT",
			params: map[string]any{
				"min_rtt_ms": 0.5,
				"max_rtt_ms": 50.0,
			},
		},
		{
			name: "custom RWND",
			params: map[string]any{
				"rwnd_size": uint32(131072),
			},
		},
		{
			name: "parallel streams",
			params: map[string]any{
				"parallel_streams": uint32(4),
			},
		},
		{
			name: "duration fallback",
			params: map[string]any{
				"duration_sec": uint32(45),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 1500,
				Duration:  0, // Use params or default.
				Params:    tc.params,
			}

			result, _ := exec.Execute("rfc6349_throughput", cfg)
			if result == nil {
				t.Fatal("Execute returned nil result")
			}

			if result.TestType != "rfc6349_throughput" {
				t.Errorf("TestType = %q, want rfc6349_throughput", result.TestType)
			}
		})
	}
}

// TestExecuteTSNWithParams tests TSN execution with various parameters.
func TestExecuteTSNWithParams(t *testing.T) {
	testCtx := dataplane.NewTestContext()
	exec := NewExecutorWithContext(testCtx)

	testCases := []struct {
		name   string
		params map[string]any
	}{
		{
			name:   "default params",
			params: nil,
		},
		{
			name: "ptp disabled",
			params: map[string]any{
				"require_ptp_sync": false,
				"ptp_enabled":      false,
			},
		},
		{
			name: "preemption enabled",
			params: map[string]any{
				"preemption_enabled": true,
			},
		},
		{
			name: "custom latency/jitter",
			params: map[string]any{
				"max_latency_ns": uint32(500000),
				"max_jitter_ns":  uint32(50000),
			},
		},
		{
			name: "custom traffic class",
			params: map[string]any{
				"num_traffic_classes": uint32(4),
				"traffic_class":       uint32(3),
			},
		},
		{
			name: "frame size fallback",
			params: map[string]any{
				"frame_size": uint32(256),
			},
		},
		{
			name: "base time ns",
			params: map[string]any{
				"base_time_ns":  uint64(1000000),
				"cycle_time_ns": uint32(500000),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &modtypes.TestConfig{
				Interface: "lo",
				FrameSize: 0, // Use params or default.
				Duration:  0,
				Params:    tc.params,
			}

			result, _ := exec.Execute("tsn_timing", cfg)
			if result == nil {
				t.Fatal("Execute returned nil result")
			}

			if result.TestType != "tsn_timing" {
				t.Errorf("TestType = %q, want tsn_timing", result.TestType)
			}
		})
	}
}

// TestExecuteDefaultSwitchCase tests that unimplemented test types return error.
func TestExecuteDefaultSwitchCase(t *testing.T) {
	testCtx := dataplane.NewTestContext()
	exec := NewExecutorWithContext(testCtx)

	cfg := &modtypes.TestConfig{
		Interface: "lo",
		FrameSize: 64,
		Duration:  10,
		Params:    nil,
	}

	// Test types that pass CanRun but aren't in the switch should not exist,
	// but we test the default case by checking that invalid types fail earlier.
	_, err := exec.Execute("unknown_test", cfg)
	if err == nil {
		t.Error("Expected error for unknown test type")
	}
}
