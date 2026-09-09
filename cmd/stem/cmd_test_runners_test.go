// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"errors"
	"strings"
	"testing"

	testmasterDP "github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

func TestRunTestRejectsUnknownTypes(t *testing.T) {
	result, err := runTest(nil, "not_a_test", 0, 0, 0, 0, 0, 0)
	if err == nil {
		t.Fatal("runTest unknown type = nil error")
	}
	if result != nil {
		t.Errorf("runTest unknown type result = %#v, want nil", result)
	}
	if !errors.Is(err, errUnknownTestType) {
		t.Errorf("runTest unknown type error = %v, want errUnknownTestType", err)
	}
}

func TestRunTestSuiteReturnsExecutionErrors(t *testing.T) {
	results, err := runTestSuite(nil, []string{testTypeThroughput}, []int{-1}, testCmdParams{})
	if err == nil {
		t.Fatal("runTestSuite invalid frame size = nil error")
	}
	if len(results) != 0 {
		t.Errorf("runTestSuite invalid frame size results = %#v, want none", results)
	}
}

// TestPrintTestResultRendersEachResultType is the operator's only view of a
// run on the CLI. Nanoseconds are converted to microseconds on the way out,
// which is where a rendering mistake silently changes a reported latency by
// three orders of magnitude, so the converted values are asserted, not just
// the labels.
func TestPrintTestResultRendersEachResultType(t *testing.T) {
	tests := []struct {
		name   string
		result any
		want   []string
	}{
		{
			name: "throughput",
			result: &testmasterDP.ThroughputResultCLI{
				FrameSize: 1518, MaxRatePct: 98.5, MaxRateMbps: 985.25, MaxRatePPS: 81000,
				Iterations: 7,
				Latency:    testmasterDP.LatencyStats{MinNs: 1000, AvgNs: 2500, MaxNs: 90000},
			},
			want: []string{
				"Max Rate:    98.50% (985.25 Mbps, 81000 pps)",
				"Iterations:  7",
				"Latency:     min=1.00us avg=2.50us max=90.00us",
			},
		},
		{
			name: "latency across load levels",
			result: []testmasterDP.LatencyResultCLI{
				{
					LoadPct: 10,
					Latency: testmasterDP.LatencyStats{
						MinNs: 1000,
						AvgNs: 2000,
						MaxNs: 3000,
						P99Ns: 4000,
					},
				},
				{
					LoadPct: 100,
					Latency: testmasterDP.LatencyStats{
						MinNs: 5000,
						AvgNs: 6000,
						MaxNs: 7000,
						P99Ns: 8000,
					},
				},
			},
			want: []string{
				"Load 10%: min=1.00us avg=2.00us max=3.00us p99=4.00us",
				"Load 100%: min=5.00us avg=6.00us max=7.00us p99=8.00us",
			},
		},
		{
			name: "frame loss",
			result: []testmasterDP.FrameLossResultCLI{
				{OfferedPct: 100, FramesTx: 1000, FramesRx: 990, LossPct: 1.0},
			},
			want: []string{"Load 100%: TX=1000 RX=990 Loss=1.0000%"},
		},
		{
			name: "back to back",
			result: &testmasterDP.BackToBackResultCLI{
				MaxBurstFrames:  12345,
				BurstDurationUs: 678,
				Trials:          3,
			},
			want: []string{"Max Burst:   12345 frames", "Duration:    678 us", "Trials:      3"},
		},
		{
			name:   "system recovery",
			result: &testmasterDP.RecoveryResultCLI{RecoveryTimeMs: 12.5, FramesLost: 42},
			want:   []string{"Recovery Time: 12.50 ms", "Frames Lost:   42"},
		},
		{
			name:   "reset",
			result: &testmasterDP.ResetResultCLI{ResetTimeMs: 250.75, FramesLost: 9},
			want:   []string{"Reset Time:  250.75 ms", "Frames Lost: 9"},
		},
		{
			name: "y1564 configuration, failing step",
			result: &testmasterDP.Y1564ConfigResult{
				ServiceID: 1,
				Steps: [4]testmasterDP.Y1564StepResult{
					{OfferedRatePct: 25, FLRPct: 0, FDAvgMs: 1.5, FDVMs: 0.5, StepPass: true},
					{OfferedRatePct: 100, FLRPct: 2.5, FDAvgMs: 9, FDVMs: 4, StepPass: false},
				},
				ServicePass: false,
			},
			want: []string{
				"Service 1: " + resultFail,
				"Step 1: 25% rate, FLR=0.0000% FD=1.50ms FDV=0.50ms [" + resultPass + "]",
				"Step 2: 100% rate, FLR=2.5000% FD=9.00ms FDV=4.00ms [" + resultFail + "]",
			},
		},
		{
			name: "y1564 performance, mixed verdicts",
			result: &testmasterDP.Y1564PerfResult{
				ServiceID: 2, DurationSec: 60, FramesTx: 1000, FramesRx: 999,
				FLRPct: 0.1, FDAvgMs: 3.5, FDVMs: 1.25,
				FLRPass: false, FDPass: true, FDVPass: true, ServicePass: false,
			},
			want: []string{
				"Service 2 Performance: " + resultFail,
				"Duration:  60 sec",
				"Frames:    TX=1000 RX=999",
				"FLR:       0.1000% [" + resultFail + "]",
				"FD:        3.50 ms [" + resultPass + "]",
				"FDV:       1.25 ms [" + resultPass + "]",
			},
		},
		{
			name: "an unimplemented test type reports its status and note",
			result: map[string]string{
				"test": "tsn_timing", "status": "not_implemented", "note": "needs dataplane support",
			},
			want: []string{"Status: not_implemented", "Note: needs dataplane support"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() { printTestResult(tt.name, tt.result, false) })
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("printTestResult output is missing %q:\n%s", want, out)
				}
			}
		})
	}
}

// TestPrintTestResultStaysSilentUnderJSON: with --json the results are
// marshalled in one batch at the end, so per-result printing would corrupt
// the document an operator pipes into jq.
func TestPrintTestResultStaysSilentUnderJSON(t *testing.T) {
	out := captureStdout(t, func() {
		printTestResult("throughput", &testmasterDP.ThroughputResultCLI{MaxRatePct: 100}, true)
	})

	if out != "" {
		t.Errorf("printTestResult under --json printed %q, want nothing", out)
	}
}

// TestPrintTestResultFallsBackForUnknownShapes: a result type the switch does
// not know must still reach the operator rather than vanish.
func TestPrintTestResultFallsBackForUnknownShapes(t *testing.T) {
	out := captureStdout(t, func() { printTestResult("odd", struct{ Answer int }{42}, false) })

	if !strings.Contains(out, "Result:") || !strings.Contains(out, "42") {
		t.Errorf("printTestResult dropped an unrecognised result: %q", out)
	}
}

// TestPrintCSVResultsEmitsHeaderAndThroughputRows: the CSV is what gets pasted
// into a report, so the header and the column order are the contract.
func TestPrintCSVResultsEmitsHeaderAndThroughputRows(t *testing.T) {
	out := captureStdout(t, func() {
		printCSVResults([]any{
			&testmasterDP.ThroughputResultCLI{
				FrameSize: 64, MaxRatePct: 50.5, MaxRateMbps: 505.25,
				Latency: testmasterDP.LatencyStats{AvgNs: 3000},
			},
			&testmasterDP.ThroughputResultCLI{
				FrameSize: 1518, MaxRatePct: 99.9, MaxRateMbps: 999,
				Latency: testmasterDP.LatencyStats{AvgNs: 1500},
			},
		})
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf(
			"printCSVResults emitted %d lines, want a header and two rows:\n%s",
			len(lines),
			out,
		)
	}
	if lines[0] != "test_type,frame_size,max_rate_pct,max_rate_mbps,loss_pct,latency_avg_us" {
		t.Errorf("header = %q", lines[0])
	}
	if lines[1] != "throughput,64,50.50,505.25,0,3.00" {
		t.Errorf("first row = %q, want throughput,64,50.50,505.25,0,3.00", lines[1])
	}
	if lines[2] != "throughput,1518,99.90,999.00,0,1.50" {
		t.Errorf("second row = %q, want throughput,1518,99.90,999.00,0,1.50", lines[2])
	}
}

// TestPrintCSVResultsSkipsNonThroughputResults: only throughput fits these
// columns, so a mixed suite must not emit half-filled rows.
func TestPrintCSVResultsSkipsNonThroughputResults(t *testing.T) {
	out := captureStdout(t, func() {
		printCSVResults([]any{
			&testmasterDP.BackToBackResultCLI{MaxBurstFrames: 10},
			map[string]string{"status": "not_implemented"},
		})
	})

	if lines := strings.Split(strings.TrimSpace(out), "\n"); len(lines) != 1 {
		t.Errorf(
			"printCSVResults emitted %d lines for non-throughput results, want the header alone:\n%s",
			len(lines),
			out,
		)
	}
}
