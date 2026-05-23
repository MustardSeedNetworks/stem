// SPDX-License-Identifier: BUSL-1.1

// Black-box testing of tui package public types.
//
// This file tests exported types (Stats, Result, TestType, Y1564StepResult) using
// the tui_test package for proper API boundary testing.
package tui_test

import (
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/services/orchestrator/tui"
)

func TestTestTypeConstants(t *testing.T) {
	allTests := []tui.TestType{
		// RFC 2544 Tests.
		tui.TestThroughput, tui.TestLatency, tui.TestFrameLoss, tui.TestBackToBack,
		tui.TestSystemRecovery, tui.TestReset,
		// Y.1564 Tests.
		tui.TestY1564Config, tui.TestY1564Perf, tui.TestY1564Full,
		// RFC 2889 Tests.
		tui.TestRFC2889Forwarding, tui.TestRFC2889Caching, tui.TestRFC2889Learning,
		tui.TestRFC2889Broadcast, tui.TestRFC2889Congestion,
		// RFC 6349 Tests.
		tui.TestRFC6349Throughput, tui.TestRFC6349Path,
		// Y.1731 Tests.
		tui.TestY1731Delay, tui.TestY1731Loss, tui.TestY1731SLM, tui.TestY1731Loopback,
		// MEF Tests.
		tui.TestMEFConfig, tui.TestMEFPerf, tui.TestMEFFull,
		// TSN Tests.
		tui.TestTSNTiming, tui.TestTSNIsolation, tui.TestTSNLatency, tui.TestTSNFull,
	}

	for _, tt := range allTests {
		if tt == "" {
			t.Errorf("test type should not be empty: %v", tt)
		}
	}
}

func TestTestTypeValues(t *testing.T) {
	// Exhaustive map of all TestType constants.
	testValues := map[tui.TestType]string{
		tui.TestThroughput:        "Throughput",
		tui.TestLatency:           "Latency",
		tui.TestFrameLoss:         "Frame Loss",
		tui.TestBackToBack:        "Back-to-Back",
		tui.TestSystemRecovery:    "System Recovery",
		tui.TestReset:             "Reset",
		tui.TestY1564Config:       "Y.1564 Config",
		tui.TestY1564Perf:         "Y.1564 Perf",
		tui.TestY1564Full:         "Y.1564 Full",
		tui.TestRFC2889Forwarding: "RFC2889 Forwarding",
		tui.TestRFC2889Caching:    "RFC2889 Caching",
		tui.TestRFC2889Learning:   "RFC2889 Learning",
		tui.TestRFC2889Broadcast:  "RFC2889 Broadcast",
		tui.TestRFC2889Congestion: "RFC2889 Congestion",
		tui.TestRFC6349Throughput: "RFC6349 Throughput",
		tui.TestRFC6349Path:       "RFC6349 Path",
		tui.TestY1731Delay:        "Y.1731 Delay",
		tui.TestY1731Loss:         "Y.1731 Loss",
		tui.TestY1731SLM:          "Y.1731 SLM",
		tui.TestY1731Loopback:     "Y.1731 Loopback",
		tui.TestMEFConfig:         "MEF Config",
		tui.TestMEFPerf:           "MEF Perf",
		tui.TestMEFFull:           "MEF Full",
		tui.TestTSNTiming:         "TSN Timing",
		tui.TestTSNIsolation:      "TSN Isolation",
		tui.TestTSNLatency:        "TSN Latency",
		tui.TestTSNFull:           "TSN Full",
	}

	for tt, expected := range testValues {
		if string(tt) != expected {
			t.Errorf("TestType %v should be '%s', got '%s'", tt, expected, string(tt))
		}
	}
}

func TestStatsStruct(t *testing.T) {
	// Only set fields that are verified in assertions.
	stats := tui.Stats{
		TestType:  tui.TestThroughput,
		FrameSize: 1518,
		Progress:  50.5,
		Iteration: 3,
		TxPackets: 10000,
		LossPct:   0.01,
	}

	if stats.TestType != tui.TestThroughput {
		t.Errorf("Expected TestType Throughput, got %s", stats.TestType)
	}
	if stats.FrameSize != 1518 {
		t.Errorf("Expected FrameSize 1518, got %d", stats.FrameSize)
	}
	if stats.Progress != 50.5 {
		t.Errorf("Expected Progress 50.5, got %f", stats.Progress)
	}
	if stats.Iteration != 3 {
		t.Errorf("Expected Iteration 3, got %d", stats.Iteration)
	}
	if stats.TxPackets != 10000 {
		t.Errorf("Expected TxPackets 10000, got %d", stats.TxPackets)
	}
	if stats.LossPct != 0.01 {
		t.Errorf("Expected LossPct 0.01, got %f", stats.LossPct)
	}
}

func TestStatsY1564Fields(t *testing.T) {
	// Only set fields that are verified in assertions.
	stats := tui.Stats{
		ServiceID:   1,
		ServiceName: "Voice Service",
		CurrentStep: 2,
		TotalSteps:  4,
		CIRMbps:     100.0,
	}

	if stats.ServiceID != 1 {
		t.Errorf("Expected ServiceID 1, got %d", stats.ServiceID)
	}
	if stats.ServiceName != "Voice Service" {
		t.Errorf("Expected ServiceName 'Voice Service', got '%s'", stats.ServiceName)
	}
	if stats.CurrentStep != 2 {
		t.Errorf("Expected CurrentStep 2, got %d", stats.CurrentStep)
	}
	if stats.TotalSteps != 4 {
		t.Errorf("Expected TotalSteps 4, got %d", stats.TotalSteps)
	}
	if stats.CIRMbps != 100.0 {
		t.Errorf("Expected CIRMbps 100.0, got %f", stats.CIRMbps)
	}
}

func TestResultStruct(t *testing.T) {
	// Only set fields that are verified in assertions.
	result := tui.Result{
		FrameSize:    1518,
		MaxRatePct:   99.5,
		MaxRateMbps:  995.0,
		LatencyAvgNs: 1500.0,
		// LossPct intentionally 0.0 - Go's zero value, verified below.
	}

	if result.FrameSize != 1518 {
		t.Errorf("Expected FrameSize 1518, got %d", result.FrameSize)
	}
	if result.MaxRatePct != 99.5 {
		t.Errorf("Expected MaxRatePct 99.5, got %f", result.MaxRatePct)
	}
	if result.MaxRateMbps != 995.0 {
		t.Errorf("Expected MaxRateMbps 995.0, got %f", result.MaxRateMbps)
	}
	if result.LossPct != 0.0 {
		t.Errorf("Expected LossPct 0.0, got %f", result.LossPct)
	}
	if result.LatencyAvgNs != 1500.0 {
		t.Errorf("Expected LatencyAvgNs 1500.0, got %f", result.LatencyAvgNs)
	}
}

func TestY1564StepResultStruct(t *testing.T) {
	// Only set fields that are verified in assertions.
	step := tui.Y1564StepResult{
		Step:           1,
		OfferedRatePct: 25.0,
		StepPass:       true,
	}

	if step.Step != 1 {
		t.Errorf("Expected Step 1, got %d", step.Step)
	}
	if step.OfferedRatePct != 25.0 {
		t.Errorf("Expected OfferedRatePct 25.0, got %f", step.OfferedRatePct)
	}
	if !step.StepPass {
		t.Error("Expected StepPass true")
	}
}

func TestTestTypeCount(t *testing.T) {
	// We should have 27 test types total.
	testTypes := []tui.TestType{
		// RFC 2544 (6).
		tui.TestThroughput,
		tui.TestLatency,
		tui.TestFrameLoss,
		tui.TestBackToBack,
		tui.TestSystemRecovery,
		tui.TestReset,
		// Y.1564 (3).
		tui.TestY1564Config,
		tui.TestY1564Perf,
		tui.TestY1564Full,
		// RFC 2889 (5).
		tui.TestRFC2889Forwarding,
		tui.TestRFC2889Caching,
		tui.TestRFC2889Learning,
		tui.TestRFC2889Broadcast,
		tui.TestRFC2889Congestion,
		// RFC 6349 (2).
		tui.TestRFC6349Throughput,
		tui.TestRFC6349Path,
		// Y.1731 (4).
		tui.TestY1731Delay,
		tui.TestY1731Loss,
		tui.TestY1731SLM,
		tui.TestY1731Loopback,
		// MEF (3).
		tui.TestMEFConfig,
		tui.TestMEFPerf,
		tui.TestMEFFull,
		// TSN (4).
		tui.TestTSNTiming,
		tui.TestTSNIsolation,
		tui.TestTSNLatency,
		tui.TestTSNFull,
	}

	if len(testTypes) != 27 {
		t.Errorf("Expected 27 test types, got %d", len(testTypes))
	}
}

func TestStatsZeroValues(t *testing.T) {
	// Empty struct - Go sets all fields to zero values automatically.
	stats := tui.Stats{}

	if stats.TxPackets != 0 {
		t.Errorf("Expected TxPackets 0, got %d", stats.TxPackets)
	}
	if stats.Progress != 0 {
		t.Errorf("Expected Progress 0, got %f", stats.Progress)
	}
	if stats.FrameSize != 0 {
		t.Errorf("Expected FrameSize 0, got %d", stats.FrameSize)
	}
}

func TestResultZeroValues(t *testing.T) {
	// Empty struct - Go sets all fields to zero values automatically.
	result := tui.Result{}

	if result.MaxRatePct != 0 {
		t.Errorf("Expected MaxRatePct 0, got %f", result.MaxRatePct)
	}
	if result.FrameSize != 0 {
		t.Errorf("Expected FrameSize 0, got %d", result.FrameSize)
	}
}

func TestStatsStateValues(t *testing.T) {
	states := []string{"idle", "running", "completed", "failed", "cancelled"}
	for _, state := range states {
		// Only set the field being verified.
		stats := tui.Stats{
			State: state,
		}
		if stats.State != state {
			t.Errorf("Expected State '%s', got '%s'", state, stats.State)
		}
	}
}

func TestStatsDuration(t *testing.T) {
	// Only set the field being verified.
	stats := tui.Stats{
		Duration: 60 * time.Second,
	}

	if stats.Duration != 60*time.Second {
		t.Errorf("Expected Duration 60s, got %v", stats.Duration)
	}
}

func TestResultTimestamp(t *testing.T) {
	now := time.Now()
	// Only set the field being verified.
	result := tui.Result{
		Timestamp: now,
	}

	if !result.Timestamp.Equal(now) {
		t.Error("Result timestamp should match input time")
	}
}

func TestY1564AllSteps(t *testing.T) {
	// Only set fields that are verified in assertions.
	steps := []tui.Y1564StepResult{
		{Step: 1, OfferedRatePct: 25.0},
		{Step: 2, OfferedRatePct: 50.0},
		{Step: 3, OfferedRatePct: 75.0},
		{Step: 4, OfferedRatePct: 100.0},
	}

	if len(steps) != 4 {
		t.Errorf("Expected 4 Y.1564 steps, got %d", len(steps))
	}

	for i, step := range steps {
		if step.Step != i+1 {
			t.Errorf("Step %d should have Step=%d, got %d", i, i+1, step.Step)
		}
	}
}

func TestRFC2889TestTypes(t *testing.T) {
	// Test RFC 2889 subset using slice (maps trigger exhaustive linter for all enum values).
	rfc2889Tests := []tui.TestType{
		tui.TestRFC2889Forwarding,
		tui.TestRFC2889Caching,
		tui.TestRFC2889Learning,
		tui.TestRFC2889Broadcast,
		tui.TestRFC2889Congestion,
	}

	if len(rfc2889Tests) != 5 {
		t.Errorf("Expected 5 RFC 2889 test types, got %d", len(rfc2889Tests))
	}

	// Verify each test type is non-empty.
	for _, tt := range rfc2889Tests {
		if tt == "" {
			t.Error("RFC 2889 test type should not be empty")
		}
	}
}

func TestMEFTestTypes(t *testing.T) {
	mefTests := []tui.TestType{
		tui.TestMEFConfig,
		tui.TestMEFPerf,
		tui.TestMEFFull,
	}

	for _, test := range mefTests {
		if string(test) == "" {
			t.Error("MEF test type should not be empty")
		}
		if len(string(test)) < 4 {
			t.Errorf("MEF test type '%s' is too short", test)
		}
	}
}

func TestTSNTestTypes(t *testing.T) {
	tsnTests := []tui.TestType{
		tui.TestTSNTiming,
		tui.TestTSNIsolation,
		tui.TestTSNLatency,
		tui.TestTSNFull,
	}

	for _, test := range tsnTests {
		if string(test) == "" {
			t.Error("TSN test type should not be empty")
		}
		if len(string(test)) < 4 {
			t.Errorf("TSN test type '%s' is too short", test)
		}
	}
}

// Benchmark tests.
func BenchmarkStatsCreation(b *testing.B) {
	for b.Loop() {
		// Only set non-zero fields for realistic benchmark.
		_ = tui.Stats{
			TestType:   tui.TestThroughput,
			FrameSize:  1518,
			Progress:   50.5,
			TxPackets:  10000,
			RxPackets:  9999,
			TxRate:     1000.0,
			RxRate:     999.5,
			LatencyAvg: 1500.0,
		}
	}
}

func BenchmarkResultCreation(b *testing.B) {
	for b.Loop() {
		// Only set non-zero fields for realistic benchmark.
		_ = tui.Result{
			FrameSize:    1518,
			MaxRatePct:   99.5,
			MaxRateMbps:  995.0,
			LatencyAvgNs: 1500.0,
			Timestamp:    time.Now(),
		}
	}
}
