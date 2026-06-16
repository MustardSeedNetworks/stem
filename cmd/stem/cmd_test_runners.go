// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"fmt"
	"math"
	"os"

	testmasterDP "github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

// runTestSuite runs all tests for given frame sizes.
func runTestSuite(
	ctx *testmasterDP.Context,
	tests []string,
	frameSizes []int,
	params testCmdParams,
) []any {
	var allResults []any

	for _, testType := range tests {
		for _, frameSize := range frameSizes {
			if frameSize < 0 || frameSize > math.MaxUint32 {
				_, _ = fmt.Fprintf(
					os.Stdout,
					"Error: frame size %d out of valid range\n",
					frameSize,
				)
				continue
			}
			ctx.SetFrameSize(uint32(frameSize))

			_, _ = fmt.Fprintf(
				os.Stdout,
				"\n[Running %s test with frame size %d bytes]\n",
				testType,
				frameSize,
			)

			result, err := runTest(
				ctx, testType,
				params.cir, params.eir,
				params.fdThreshold, params.fdvThreshold, params.flrThreshold,
				params.duration,
			)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stdout, "Error: %s test failed: %v\n", testType, err)
				continue
			}

			allResults = append(allResults, result)
			printTestResult(testType, result, params.jsonOutput)
		}
	}

	return allResults
}

// runTest executes a single test and returns the result.
func runTest(
	ctx *testmasterDP.Context,
	testType string,
	cir, eir, fdThreshold, fdvThreshold, flrThreshold float64,
	duration int,
) (any, error) {
	switch testType {
	case testTypeThroughput:
		return runThroughputTest(ctx)
	case testTypeLatency:
		return runLatencyTest(ctx)
	case testTypeFrameLoss:
		return runFrameLossTest(ctx)
	case testTypeBackToBack:
		return runBackToBackTest(ctx)
	case testTypeSystemRecovery:
		return runSystemRecoveryTest(ctx)
	case testTypeReset:
		return runResetTest(ctx)
	case "y1564_config", "y1564":
		return runY1564ConfigTest(ctx, cir, eir, fdThreshold, fdvThreshold, flrThreshold)
	case testTypeY1564Perf:
		return runY1564PerfTest(ctx, cir, eir, fdThreshold, fdvThreshold, flrThreshold, duration)
	default:
		return map[string]string{
			"test":   testType,
			"status": "not_implemented",
			"note":   "This test type requires additional dataplane support",
		}, nil
	}
}

// runThroughputTest executes RFC 2544 throughput test.
func runThroughputTest(ctx *testmasterDP.Context) (any, error) {
	result, err := ctx.RunThroughputTest()
	if err != nil {
		return nil, fmt.Errorf("throughput test failed: %w", err)
	}
	return result, nil
}

// runLatencyTest executes RFC 2544 latency test.
func runLatencyTest(ctx *testmasterDP.Context) (any, error) {
	loadLevels := []float64{defaultLoadLevelStep, 25, 50, 75, 90, defaultLoadLevelMax}
	result, err := ctx.RunLatencyTest(loadLevels)
	if err != nil {
		return nil, fmt.Errorf("latency test failed: %w", err)
	}
	return result, nil
}

// runFrameLossTest executes RFC 2544 frame loss test.
func runFrameLossTest(ctx *testmasterDP.Context) (any, error) {
	result, err := ctx.RunFrameLossTest(
		defaultLoadLevelStep,
		defaultLoadLevelMax,
		defaultLoadLevelStep,
	)
	if err != nil {
		return nil, fmt.Errorf("frame loss test failed: %w", err)
	}
	return result, nil
}

// runBackToBackTest executes RFC 2544 back-to-back test.
func runBackToBackTest(ctx *testmasterDP.Context) (any, error) {
	result, err := ctx.RunBackToBackTest(defaultBackToBackBurst, defaultTrials)
	if err != nil {
		return nil, fmt.Errorf("back-to-back test failed: %w", err)
	}
	return result, nil
}

// runSystemRecoveryTest executes RFC 2544 system recovery test.
func runSystemRecoveryTest(ctx *testmasterDP.Context) (any, error) {
	result, err := ctx.RunSystemRecoveryTest(defaultOverloadRate, defaultOverloadDuration)
	if err != nil {
		return nil, fmt.Errorf("system recovery test failed: %w", err)
	}
	return result, nil
}

// runResetTest executes RFC 2544 reset test.
func runResetTest(ctx *testmasterDP.Context) (any, error) {
	result, err := ctx.RunResetTest()
	if err != nil {
		return nil, fmt.Errorf("reset test failed: %w", err)
	}
	return result, nil
}

// newY1564Service creates a Y.1564 service with given SLA parameters.
func newY1564Service(
	cir, eir, fdThreshold, fdvThreshold, flrThreshold float64,
) *testmasterDP.Y1564Service {
	return &testmasterDP.Y1564Service{
		ServiceID:   1,
		ServiceName: "Service-1",
		SLA: testmasterDP.Y1564SLA{
			CIRMbps:         cir,
			EIRMbps:         eir,
			CBSBytes:        0,
			EBSBytes:        0,
			FDThresholdMs:   fdThreshold,
			FDVThresholdMs:  fdvThreshold,
			FLRThresholdPct: flrThreshold,
		},
		FrameSize: defaultFrameSize,
		CoS:       0,
		Enabled:   true,
	}
}

// runY1564ConfigTest executes Y.1564 configuration test.
func runY1564ConfigTest(
	ctx *testmasterDP.Context,
	cir, eir, fdThreshold, fdvThreshold, flrThreshold float64,
) (any, error) {
	service := newY1564Service(cir, eir, fdThreshold, fdvThreshold, flrThreshold)
	result, err := ctx.RunY1564ConfigTest(service)
	if err != nil {
		return nil, fmt.Errorf("Y.1564 config test failed: %w", err)
	}
	return result, nil
}

// runY1564PerfTest executes Y.1564 performance test.
func runY1564PerfTest(
	ctx *testmasterDP.Context,
	cir, eir, fdThreshold, fdvThreshold, flrThreshold float64,
	duration int,
) (any, error) {
	if duration < 0 || duration > math.MaxUint32 {
		return nil, fmt.Errorf("duration %d out of valid range (0-%d)", duration, math.MaxUint32)
	}
	service := newY1564Service(cir, eir, fdThreshold, fdvThreshold, flrThreshold)
	result, err := ctx.RunY1564PerfTest(service, uint32(duration))
	if err != nil {
		return nil, fmt.Errorf("Y.1564 perf test failed: %w", err)
	}
	return result, nil
}

// printTestResult prints a test result.
func printTestResult(_ string, result any, jsonOutput bool) {
	if jsonOutput {
		return // Will be printed in batch at the end.
	}

	switch r := result.(type) {
	case *testmasterDP.ThroughputResultCLI:
		_, _ = fmt.Fprintf(
			os.Stdout,
			"  Max Rate:    %.2f%% (%.2f Mbps, %.0f pps)\n",
			r.MaxRatePct, r.MaxRateMbps, r.MaxRatePPS,
		)
		_, _ = fmt.Fprintf(os.Stdout, "  Iterations:  %d\n", r.Iterations)
		_, _ = fmt.Fprintf(
			os.Stdout,
			"  Latency:     min=%.2fus avg=%.2fus max=%.2fus\n",
			r.Latency.MinNs/nsToUsConversion,
			r.Latency.AvgNs/nsToUsConversion,
			r.Latency.MaxNs/nsToUsConversion,
		)

	case []testmasterDP.LatencyResultCLI:
		for _, lr := range r {
			_, _ = fmt.Fprintf(
				os.Stdout,
				"  Load %.0f%%: min=%.2fus avg=%.2fus max=%.2fus p99=%.2fus\n",
				lr.LoadPct,
				lr.Latency.MinNs/nsToUsConversion,
				lr.Latency.AvgNs/nsToUsConversion,
				lr.Latency.MaxNs/nsToUsConversion,
				lr.Latency.P99Ns/nsToUsConversion,
			)
		}

	case []testmasterDP.FrameLossResultCLI:
		for _, fl := range r {
			_, _ = fmt.Fprintf(
				os.Stdout,
				"  Load %.0f%%: TX=%d RX=%d Loss=%.4f%%\n",
				fl.OfferedPct, fl.FramesTx, fl.FramesRx, fl.LossPct,
			)
		}

	case *testmasterDP.BackToBackResultCLI:
		_, _ = fmt.Fprintf(os.Stdout, "  Max Burst:   %d frames\n", r.MaxBurstFrames)
		_, _ = fmt.Fprintf(os.Stdout, "  Duration:    %d us\n", r.BurstDurationUs)
		_, _ = fmt.Fprintf(os.Stdout, "  Trials:      %d\n", r.Trials)

	case *testmasterDP.RecoveryResultCLI:
		_, _ = fmt.Fprintf(os.Stdout, "  Recovery Time: %.2f ms\n", r.RecoveryTimeMs)
		_, _ = fmt.Fprintf(os.Stdout, "  Frames Lost:   %d\n", r.FramesLost)

	case *testmasterDP.ResetResultCLI:
		_, _ = fmt.Fprintf(os.Stdout, "  Reset Time:  %.2f ms\n", r.ResetTimeMs)
		_, _ = fmt.Fprintf(os.Stdout, "  Frames Lost: %d\n", r.FramesLost)

	case *testmasterDP.Y1564ConfigResult:
		passStr := resultPass
		if !r.ServicePass {
			passStr = resultFail
		}
		_, _ = fmt.Fprintf(os.Stdout, "  Service %d: %s\n", r.ServiceID, passStr)
		for i, step := range r.Steps {
			stepPass := resultPass
			if !step.StepPass {
				stepPass = resultFail
			}
			_, _ = fmt.Fprintf(
				os.Stdout,
				"    Step %d: %.0f%% rate, FLR=%.4f%% FD=%.2fms FDV=%.2fms [%s]\n",
				i+1, step.OfferedRatePct, step.FLRPct, step.FDAvgMs, step.FDVMs, stepPass,
			)
		}

	case *testmasterDP.Y1564PerfResult:
		passStr := resultPass
		if !r.ServicePass {
			passStr = resultFail
		}
		_, _ = fmt.Fprintf(os.Stdout, "  Service %d Performance: %s\n", r.ServiceID, passStr)
		_, _ = fmt.Fprintf(os.Stdout, "    Duration:  %d sec\n", r.DurationSec)
		_, _ = fmt.Fprintf(os.Stdout, "    Frames:    TX=%d RX=%d\n", r.FramesTx, r.FramesRx)
		_, _ = fmt.Fprintf(os.Stdout, "    FLR:       %.4f%% [%s]\n", r.FLRPct, boolToPassFail(r.FLRPass))
		_, _ = fmt.Fprintf(os.Stdout, "    FD:        %.2f ms [%s]\n", r.FDAvgMs, boolToPassFail(r.FDPass))
		_, _ = fmt.Fprintf(os.Stdout, "    FDV:       %.2f ms [%s]\n", r.FDVMs, boolToPassFail(r.FDVPass))

	case map[string]string:
		_, _ = fmt.Fprintf(os.Stdout, "  Status: %s\n", r["status"])
		if note, ok := r["note"]; ok {
			_, _ = fmt.Fprintf(os.Stdout, "  Note: %s\n", note)
		}

	default:
		_, _ = fmt.Fprintf(os.Stdout, "  Result: %+v\n", result)
	}
}

func boolToPassFail(b bool) string {
	if b {
		return resultPass
	}
	return resultFail
}

func printCSVResults(results []any) {
	// Print CSV header.
	_, _ = fmt.Fprintln(
		os.Stdout,
		"\ntest_type,frame_size,max_rate_pct,max_rate_mbps,loss_pct,latency_avg_us",
	)

	for _, r := range results {
		if result, ok := r.(*testmasterDP.ThroughputResultCLI); ok {
			_, _ = fmt.Fprintf(
				os.Stdout,
				"throughput,%d,%.2f,%.2f,0,%.2f\n",
				result.FrameSize,
				result.MaxRatePct,
				result.MaxRateMbps,
				result.Latency.AvgNs/nsToUsConversion,
			)
		}
	}
}
