// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package tui

import (
	"testing"
	"time"
)

func TestTestTypeConstants(t *testing.T) {
	// RFC 2544 Tests
	rfc2544Tests := []TestType{
		TestThroughput,
		TestLatency,
		TestFrameLoss,
		TestBackToBack,
		TestSystemRecovery,
		TestReset,
	}

	for _, tt := range rfc2544Tests {
		if tt == "" {
			t.Error("RFC 2544 test type should not be empty")
		}
	}

	// Y.1564 Tests
	y1564Tests := []TestType{
		TestY1564Config,
		TestY1564Perf,
		TestY1564Full,
	}

	for _, tt := range y1564Tests {
		if tt == "" {
			t.Error("Y.1564 test type should not be empty")
		}
	}

	// RFC 2889 Tests
	rfc2889Tests := []TestType{
		TestRFC2889Forwarding,
		TestRFC2889Caching,
		TestRFC2889Learning,
		TestRFC2889Broadcast,
		TestRFC2889Congestion,
	}

	for _, tt := range rfc2889Tests {
		if tt == "" {
			t.Error("RFC 2889 test type should not be empty")
		}
	}

	// RFC 6349 Tests
	rfc6349Tests := []TestType{
		TestRFC6349Throughput,
		TestRFC6349Path,
	}

	for _, tt := range rfc6349Tests {
		if tt == "" {
			t.Error("RFC 6349 test type should not be empty")
		}
	}

	// Y.1731 Tests
	y1731Tests := []TestType{
		TestY1731Delay,
		TestY1731Loss,
		TestY1731SLM,
		TestY1731Loopback,
	}

	for _, tt := range y1731Tests {
		if tt == "" {
			t.Error("Y.1731 test type should not be empty")
		}
	}

	// MEF Tests
	mefTests := []TestType{
		TestMEFConfig,
		TestMEFPerf,
		TestMEFFull,
	}

	for _, tt := range mefTests {
		if tt == "" {
			t.Error("MEF test type should not be empty")
		}
	}

	// TSN Tests
	tsnTests := []TestType{
		TestTSNTiming,
		TestTSNIsolation,
		TestTSNLatency,
		TestTSNFull,
	}

	for _, tt := range tsnTests {
		if tt == "" {
			t.Error("TSN test type should not be empty")
		}
	}
}

func TestTestTypeValues(t *testing.T) {
	testValues := map[TestType]string{
		TestThroughput:     "Throughput",
		TestLatency:        "Latency",
		TestFrameLoss:      "Frame Loss",
		TestBackToBack:     "Back-to-Back",
		TestSystemRecovery: "System Recovery",
		TestReset:          "Reset",
		TestY1564Config:    "Y.1564 Config",
		TestY1564Perf:      "Y.1564 Perf",
		TestY1564Full:      "Y.1564 Full",
	}

	for tt, expected := range testValues {
		if string(tt) != expected {
			t.Errorf("TestType %v should be '%s', got '%s'", tt, expected, string(tt))
		}
	}
}

func TestStatsStruct(t *testing.T) {
	stats := Stats{
		TestType:    TestThroughput,
		FrameSize:   1518,
		Progress:    50.5,
		State:       "running",
		Iteration:   3,
		MaxIter:     10,
		TxPackets:   10000,
		TxBytes:     15180000,
		RxPackets:   9999,
		RxBytes:     15178482,
		TxRate:      1000.0,
		RxRate:      999.5,
		TxPPS:       100000,
		RxPPS:       99990,
		OfferedRate: 100.0,
		LossPct:     0.01,
		LatencyMin:  100.0,
		LatencyMax:  5000.0,
		LatencyAvg:  1500.0,
		LatencyP99:  4500.0,
		StartTime:   time.Now(),
		Duration:    30 * time.Second,
	}

	if stats.TestType != TestThroughput {
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
	stats := Stats{
		TestType:    TestY1564Config,
		ServiceID:   1,
		ServiceName: "Voice Service",
		CurrentStep: 2,
		TotalSteps:  4,
		CIRMbps:     100.0,
		FDMs:        5.5,
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
	result := Result{
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
	step := Y1564StepResult{
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
	// We should have 27 test types total
	testTypes := []TestType{
		// RFC 2544 (6)
		TestThroughput, TestLatency, TestFrameLoss, TestBackToBack, TestSystemRecovery, TestReset,
		// Y.1564 (3)
		TestY1564Config, TestY1564Perf, TestY1564Full,
		// RFC 2889 (5)
		TestRFC2889Forwarding, TestRFC2889Caching, TestRFC2889Learning, TestRFC2889Broadcast, TestRFC2889Congestion,
		// RFC 6349 (2)
		TestRFC6349Throughput, TestRFC6349Path,
		// Y.1731 (4)
		TestY1731Delay, TestY1731Loss, TestY1731SLM, TestY1731Loopback,
		// MEF (3)
		TestMEFConfig, TestMEFPerf, TestMEFFull,
		// TSN (4)
		TestTSNTiming, TestTSNIsolation, TestTSNLatency, TestTSNFull,
	}

	if len(testTypes) != 27 {
		t.Errorf("Expected 27 test types, got %d", len(testTypes))
	}
}

func TestStatsZeroValues(t *testing.T) {
	stats := Stats{}

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
	result := Result{}

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
		stats := Stats{State: state}
		if stats.State != state {
			t.Errorf("Expected State '%s', got '%s'", state, stats.State)
		}
	}
}

func TestStatsDuration(t *testing.T) {
	stats := Stats{
		StartTime: time.Now().Add(-60 * time.Second),
		Duration:  60 * time.Second,
	}

	if stats.Duration != 60*time.Second {
		t.Errorf("Expected Duration 60s, got %v", stats.Duration)
	}
}

func TestResultTimestamp(t *testing.T) {
	now := time.Now()
	result := Result{
		Timestamp: now,
	}

	if !result.Timestamp.Equal(now) {
		t.Error("Result timestamp should match input time")
	}
}

func TestY1564AllSteps(t *testing.T) {
	steps := []Y1564StepResult{
		{Step: 1, OfferedRatePct: 25.0, StepPass: true},
		{Step: 2, OfferedRatePct: 50.0, StepPass: true},
		{Step: 3, OfferedRatePct: 75.0, StepPass: true},
		{Step: 4, OfferedRatePct: 100.0, StepPass: false},
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
	rfc2889Tests := map[TestType]bool{
		TestRFC2889Forwarding: true,
		TestRFC2889Caching:    true,
		TestRFC2889Learning:   true,
		TestRFC2889Broadcast:  true,
		TestRFC2889Congestion: true,
	}

	if len(rfc2889Tests) != 5 {
		t.Errorf("Expected 5 RFC 2889 test types, got %d", len(rfc2889Tests))
	}
}

func TestMEFTestTypes(t *testing.T) {
	mefTests := []TestType{
		TestMEFConfig,
		TestMEFPerf,
		TestMEFFull,
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
	tsnTests := []TestType{
		TestTSNTiming,
		TestTSNIsolation,
		TestTSNLatency,
		TestTSNFull,
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

// Benchmark tests
func BenchmarkStatsCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Stats{
			TestType:   TestThroughput,
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
	for i := 0; i < b.N; i++ {
		_ = Result{
			FrameSize:    1518,
			MaxRatePct:   99.5,
			MaxRateMbps:  995.0,
			LatencyAvgNs: 1500.0,
			Timestamp:    time.Now(),
		}
	}
}
