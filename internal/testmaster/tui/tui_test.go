// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

//nolint:govet // unusedwrite: Test structs set fields for completeness even if not all fields are verified.
package tui_test

import (
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/testmaster/tui"
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
	//nolint:govet // Test data setup - intentionally initializing all fields for struct validation
	stats := tui.Stats{
		TestType:     tui.TestThroughput,
		FrameSize:    1518,
		Progress:     50.5,
		State:        "running",
		Iteration:    3,
		MaxIter:      10,
		TxPackets:    10000,
		TxBytes:      15180000,
		RxPackets:    9999,
		RxBytes:      15178482,
		TxRate:       1000.0,
		RxRate:       999.5,
		TxPPS:        100000,
		RxPPS:        99990,
		OfferedRate:  100.0,
		LossPct:      0.01,
		LatencyMin:   100.0,
		LatencyMax:   5000.0,
		LatencyAvg:   1500.0,
		LatencyP99:   4500.0,
		StartTime:    time.Now(),
		Duration:     30 * time.Second,
		ServiceID:    0,
		ServiceName:  "",
		CurrentStep:  0,
		TotalSteps:   0,
		CIRMbps:      0,
		FDMs:         0,
		FDVMs:        0,
		FLRPct:       0,
		FDThreshold:  0,
		FDVThreshold: 0,
		FLRThreshold: 0,
		FDPass:       false,
		FDVPass:      false,
		FLRPass:      false,
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
	//nolint:govet // Test data setup - intentionally initializing all fields for struct validation
	stats := tui.Stats{
		TestType:     tui.TestY1564Config,
		FrameSize:    0,
		Progress:     0,
		State:        "",
		Iteration:    0,
		MaxIter:      0,
		TxPackets:    0,
		TxBytes:      0,
		RxPackets:    0,
		RxBytes:      0,
		TxRate:       0,
		RxRate:       0,
		TxPPS:        0,
		RxPPS:        0,
		OfferedRate:  0,
		LossPct:      0,
		LatencyMin:   0,
		LatencyMax:   0,
		LatencyAvg:   0,
		LatencyP99:   0,
		StartTime:    time.Time{},
		Duration:     0,
		ServiceID:    1,
		ServiceName:  "Voice Service",
		CurrentStep:  2,
		TotalSteps:   4,
		CIRMbps:      100.0,
		FDMs:         5.5,
		FDVMs:        0,
		FLRPct:       0,
		FDThreshold:  0,
		FDVThreshold: 0,
		FLRThreshold: 0,
		FDPass:       false,
		FDVPass:      false,
		FLRPass:      false,
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
	result := tui.Result{
		FrameSize:    1518,
		MaxRatePct:   99.5,
		MaxRateMbps:  995.0,
		LossPct:      0.0,
		LatencyAvgNs: 1500.0,
		Timestamp:    time.Now(),
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
	step := tui.Y1564StepResult{
		Step:           1,
		OfferedRatePct: 25.0,
		FLRPct:         0.001,
		FDMs:           5.5,
		FDVMs:          1.2,
		FLRPass:        true,
		FDPass:         true,
		FDVPass:        true,
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
		tui.TestThroughput, tui.TestLatency, tui.TestFrameLoss, tui.TestBackToBack, tui.TestSystemRecovery, tui.TestReset,
		// Y.1564 (3).
		tui.TestY1564Config, tui.TestY1564Perf, tui.TestY1564Full,
		// RFC 2889 (5).
		tui.TestRFC2889Forwarding, tui.TestRFC2889Caching, tui.TestRFC2889Learning, tui.TestRFC2889Broadcast, tui.TestRFC2889Congestion,
		// RFC 6349 (2).
		tui.TestRFC6349Throughput, tui.TestRFC6349Path,
		// Y.1731 (4).
		tui.TestY1731Delay, tui.TestY1731Loss, tui.TestY1731SLM, tui.TestY1731Loopback,
		// MEF (3).
		tui.TestMEFConfig, tui.TestMEFPerf, tui.TestMEFFull,
		// TSN (4).
		tui.TestTSNTiming, tui.TestTSNIsolation, tui.TestTSNLatency, tui.TestTSNFull,
	}

	if len(testTypes) != 27 {
		t.Errorf("Expected 27 test types, got %d", len(testTypes))
	}
}

func TestStatsZeroValues(t *testing.T) {
	//nolint:govet // Test data setup - intentionally initializing all fields for struct validation
	stats := tui.Stats{
		TestType:     "",
		FrameSize:    0,
		Progress:     0,
		State:        "",
		Iteration:    0,
		MaxIter:      0,
		TxPackets:    0,
		TxBytes:      0,
		RxPackets:    0,
		RxBytes:      0,
		TxRate:       0,
		RxRate:       0,
		TxPPS:        0,
		RxPPS:        0,
		OfferedRate:  0,
		LossPct:      0,
		LatencyMin:   0,
		LatencyMax:   0,
		LatencyAvg:   0,
		LatencyP99:   0,
		StartTime:    time.Time{},
		Duration:     0,
		ServiceID:    0,
		ServiceName:  "",
		CurrentStep:  0,
		TotalSteps:   0,
		CIRMbps:      0,
		FDMs:         0,
		FDVMs:        0,
		FLRPct:       0,
		FDThreshold:  0,
		FDVThreshold: 0,
		FLRThreshold: 0,
		FDPass:       false,
		FDVPass:      false,
		FLRPass:      false,
	}

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
	result := tui.Result{
		FrameSize:    0,
		MaxRatePct:   0,
		MaxRateMbps:  0,
		LossPct:      0,
		LatencyAvgNs: 0,
		Timestamp:    time.Time{},
	}

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
		//nolint:govet // Test data setup - intentionally initializing all fields for struct validation
		stats := tui.Stats{
			TestType:     "",
			FrameSize:    0,
			Progress:     0,
			State:        state,
			Iteration:    0,
			MaxIter:      0,
			TxPackets:    0,
			TxBytes:      0,
			RxPackets:    0,
			RxBytes:      0,
			TxRate:       0,
			RxRate:       0,
			TxPPS:        0,
			RxPPS:        0,
			OfferedRate:  0,
			LossPct:      0,
			LatencyMin:   0,
			LatencyMax:   0,
			LatencyAvg:   0,
			LatencyP99:   0,
			StartTime:    time.Time{},
			Duration:     0,
			ServiceID:    0,
			ServiceName:  "",
			CurrentStep:  0,
			TotalSteps:   0,
			CIRMbps:      0,
			FDMs:         0,
			FDVMs:        0,
			FLRPct:       0,
			FDThreshold:  0,
			FDVThreshold: 0,
			FLRThreshold: 0,
			FDPass:       false,
			FDVPass:      false,
			FLRPass:      false,
		}
		if stats.State != state {
			t.Errorf("Expected State '%s', got '%s'", state, stats.State)
		}
	}
}

func TestStatsDuration(t *testing.T) {
	//nolint:govet // Test data setup - intentionally initializing all fields for struct validation
	stats := tui.Stats{
		TestType:     "",
		FrameSize:    0,
		Progress:     0,
		State:        "",
		Iteration:    0,
		MaxIter:      0,
		TxPackets:    0,
		TxBytes:      0,
		RxPackets:    0,
		RxBytes:      0,
		TxRate:       0,
		RxRate:       0,
		TxPPS:        0,
		RxPPS:        0,
		OfferedRate:  0,
		LossPct:      0,
		LatencyMin:   0,
		LatencyMax:   0,
		LatencyAvg:   0,
		LatencyP99:   0,
		StartTime:    time.Now().Add(-60 * time.Second),
		Duration:     60 * time.Second,
		ServiceID:    0,
		ServiceName:  "",
		CurrentStep:  0,
		TotalSteps:   0,
		CIRMbps:      0,
		FDMs:         0,
		FDVMs:        0,
		FLRPct:       0,
		FDThreshold:  0,
		FDVThreshold: 0,
		FLRThreshold: 0,
		FDPass:       false,
		FDVPass:      false,
		FLRPass:      false,
	}

	if stats.Duration != 60*time.Second {
		t.Errorf("Expected Duration 60s, got %v", stats.Duration)
	}
}

func TestResultTimestamp(t *testing.T) {
	now := time.Now()
	result := tui.Result{
		FrameSize:    0,
		MaxRatePct:   0,
		MaxRateMbps:  0,
		LossPct:      0,
		LatencyAvgNs: 0,
		Timestamp:    now,
	}

	if !result.Timestamp.Equal(now) {
		t.Error("Result timestamp should match input time")
	}
}

func TestY1564AllSteps(t *testing.T) {
	steps := []tui.Y1564StepResult{
		{
			Step: 1, OfferedRatePct: 25.0, FLRPct: 0, FDMs: 0, FDVMs: 0,
			FLRPass: false, FDPass: false, FDVPass: false, StepPass: true,
		},
		{
			Step: 2, OfferedRatePct: 50.0, FLRPct: 0, FDMs: 0, FDVMs: 0,
			FLRPass: false, FDPass: false, FDVPass: false, StepPass: true,
		},
		{
			Step: 3, OfferedRatePct: 75.0, FLRPct: 0, FDMs: 0, FDVMs: 0,
			FLRPass: false, FDPass: false, FDVPass: false, StepPass: true,
		},
		{
			Step: 4, OfferedRatePct: 100.0, FLRPct: 0, FDMs: 0, FDVMs: 0,
			FLRPass: false, FDPass: false, FDVPass: false, StepPass: false,
		},
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
	// Intentionally testing only RFC2889 subset.
	//nolint:exhaustive // This test specifically validates RFC2889 test types only.
	rfc2889Tests := map[tui.TestType]bool{
		tui.TestRFC2889Forwarding: true,
		tui.TestRFC2889Caching:    true,
		tui.TestRFC2889Learning:   true,
		tui.TestRFC2889Broadcast:  true,
		tui.TestRFC2889Congestion: true,
	}

	if len(rfc2889Tests) != 5 {
		t.Errorf("Expected 5 RFC 2889 test types, got %d", len(rfc2889Tests))
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
		_ = tui.Stats{
			TestType:     tui.TestThroughput,
			FrameSize:    1518,
			Progress:     50.5,
			State:        "",
			Iteration:    0,
			MaxIter:      0,
			TxPackets:    10000,
			TxBytes:      0,
			RxPackets:    9999,
			RxBytes:      0,
			TxRate:       1000.0,
			RxRate:       999.5,
			TxPPS:        0,
			RxPPS:        0,
			OfferedRate:  0,
			LossPct:      0,
			LatencyMin:   0,
			LatencyMax:   0,
			LatencyAvg:   1500.0,
			LatencyP99:   0,
			StartTime:    time.Time{},
			Duration:     0,
			ServiceID:    0,
			ServiceName:  "",
			CurrentStep:  0,
			TotalSteps:   0,
			CIRMbps:      0,
			FDMs:         0,
			FDVMs:        0,
			FLRPct:       0,
			FDThreshold:  0,
			FDVThreshold: 0,
			FLRThreshold: 0,
			FDPass:       false,
			FDVPass:      false,
			FLRPass:      false,
		}
	}
}

func BenchmarkResultCreation(b *testing.B) {
	for b.Loop() {
		_ = tui.Result{
			FrameSize:    1518,
			MaxRatePct:   99.5,
			MaxRateMbps:  995.0,
			LossPct:      0,
			LatencyAvgNs: 1500.0,
			Timestamp:    time.Now(),
		}
	}
}
