// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/license"
	"github.com/MustardSeedNetworks/stem/internal/logging"
	"github.com/MustardSeedNetworks/stem/internal/services"
	testmasterDP "github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

// testCmdParams holds parameters for running test suite.
type testCmdParams struct {
	cir          float64
	eir          float64
	fdThreshold  float64
	fdvThreshold float64
	flrThreshold float64
	duration     int
	jsonOutput   bool
	csvOutput    bool
}

// testCmdFlags holds parsed command line flags for test command.
type testCmdFlags struct {
	iface        string
	testTypes    string
	duration     int
	frameSizes   string
	resolution   float64
	maxLoss      float64
	warmup       int
	cir          float64
	eir          float64
	fdThreshold  float64
	fdvThreshold float64
	flrThreshold float64
	jsonOutput   bool
	csvOutput    bool
}

// validateTestTypesList validates a list of test types.
func validateTestTypesList(tests []string) bool {
	for _, t := range tests {
		if mod := services.GetModuleForTest(t); mod == nil {
			_, _ = fmt.Fprintf(os.Stdout, "Error: Unknown test type '%s'\n", t)
			_, _ = fmt.Fprintln(os.Stdout, "Run 'stem list-tests' to see available tests")
			return false
		}
	}
	return true
}

// checkTestLicense checks that the license is valid for running tests.
func checkTestLicense() bool {
	mgr, err := license.NewManager()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Warning: License check failed: %v\n", err)
		return true // Allow to continue with warning.
	}

	state := mgr.GetState()
	switch {
	case state == nil:
		_, _ = fmt.Fprintln(os.Stdout, "No active license. Starting 14-day trial...")
		result := mgr.StartTrial()
		if !result.Success {
			_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
			return false
		}
		_, _ = fmt.Fprintf(os.Stdout, "%s\n\n", result.Message)
	case !mgr.IsActivated():
		_, _ = fmt.Fprintln(os.Stdout, "Error: License expired. Please activate a valid license.")
		_, _ = fmt.Fprintln(os.Stdout, "Run 'stem license --status' for details")
		return false
	case license.Tier(state.Tier) < license.TierProfessional && !state.IsTrialMode:
		_, _ = fmt.Fprintln(os.Stdout, "Error: Professional features require a Tier 2 (Professional) license")
		_, _ = fmt.Fprintln(os.Stdout, "Your license: Tier 1 (Reflector only)")
		return false
	}
	return true
}

// printTestConfiguration prints the test configuration.
func printTestConfiguration(
	iface, testTypes, frameSizes string,
	duration int,
	resolution, maxLoss float64,
	warmup int,
) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s - Network Testing\n", ProductName, version.GetVersion())
	_, _ = fmt.Fprintln(os.Stdout, strings.Repeat("=", bannerWidth))
	_, _ = fmt.Fprintf(os.Stdout, "Interface:    %s\n", iface)
	_, _ = fmt.Fprintf(os.Stdout, "Tests:        %s\n", testTypes)
	_, _ = fmt.Fprintf(os.Stdout, "Duration:     %d seconds\n", duration)
	_, _ = fmt.Fprintf(os.Stdout, "Frame sizes:  %s\n", frameSizes)
	_, _ = fmt.Fprintf(os.Stdout, "Resolution:   %.2f%%\n", resolution)
	_, _ = fmt.Fprintf(os.Stdout, "Max loss:     %.2f%%\n", maxLoss)
	_, _ = fmt.Fprintf(os.Stdout, "Warmup:       %d seconds\n", warmup)
	_, _ = fmt.Fprintln(os.Stdout, strings.Repeat("=", bannerWidth))
}

// parseTestFlags parses test command flags and returns the parsed values.
func parseTestFlags(args []string) (*testCmdFlags, error) {
	fs := flag.NewFlagSet(subTest, flag.ContinueOnError)

	// Basic options.
	iface := fs.String("interface", "", "Network interface")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")
	testTypes := fs.String("type", testTypeThroughput, "Test type(s), comma-separated")
	fs.StringVar(testTypes, "t", testTypeThroughput, "Test type (shorthand)")
	duration := fs.Int("duration", defaultTestDuration, "Test duration in seconds")
	fs.IntVar(duration, "d", defaultTestDuration, "Test duration (shorthand)")
	frameSizes := fs.String("frame-sizes", "64,128,256,512,1024,1280,1518", "Frame sizes")

	// Advanced options.
	resolution := fs.Float64("resolution", defaultResolution, "Binary search resolution %")
	maxLoss := fs.Float64("max-loss", defaultMaxLoss, "Maximum acceptable loss %")
	warmup := fs.Int("warmup", defaultWarmup, "Warmup period in seconds")
	_ = fs.Int("trials", defaultTrials, "Number of trials") // Used in config.

	// Y.1564 options.
	cir := fs.Float64("cir", 0, "Committed Information Rate (Mbps)")
	eir := fs.Float64("eir", 0, "Excess Information Rate (Mbps)")
	fdThreshold := fs.Float64("fd-threshold", defaultFDThreshold, "Frame Delay threshold (ms)")
	fdvThreshold := fs.Float64(
		"fdv-threshold",
		defaultFDVThreshold,
		"Frame Delay Variation threshold (ms)",
	)
	flrThreshold := fs.Float64(
		"flr-threshold",
		defaultFLRThreshold,
		"Frame Loss Rate threshold (%)",
	)

	// Output format.
	jsonOutput := fs.Bool("json", false, "Output results in JSON")
	csvOutput := fs.Bool("csv", false, "Output results in CSV")

	parseErr := fs.Parse(args)
	if parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			return nil, parseErr
		}
		return nil, fmt.Errorf("failed to parse test flags: %w", parseErr)
	}

	return &testCmdFlags{
		iface:        *iface,
		testTypes:    *testTypes,
		duration:     *duration,
		frameSizes:   *frameSizes,
		resolution:   *resolution,
		maxLoss:      *maxLoss,
		warmup:       *warmup,
		cir:          *cir,
		eir:          *eir,
		fdThreshold:  *fdThreshold,
		fdvThreshold: *fdvThreshold,
		flrThreshold: *flrThreshold,
		jsonOutput:   *jsonOutput,
		csvOutput:    *csvOutput,
	}, nil
}

// createTestConfig creates a dataplane config from test flags.
func createTestConfig(flags *testCmdFlags) *testmasterDP.Config {
	return &testmasterDP.Config{
		Interface:      flags.iface,
		LineRate:       0,
		AutoDetect:     true,
		TestType:       testmasterDP.TestThroughput,
		FrameSize:      0,
		IncludeJumbo:   false,
		TrialDuration:  time.Duration(flags.duration) * time.Second,
		WarmupPeriod:   time.Duration(flags.warmup) * time.Second,
		InitialRatePct: 0,
		ResolutionPct:  flags.resolution,
		MaxIterations:  0,
		AcceptableLoss: flags.maxLoss,
		HWTimestamp:    false,
		MeasureLatency: true,
		UsePacing:      false,
		BatchSize:      0,
		UseDPDK:        false,
		DPDKArgs:       "",
	}
}

// parseFrameSizes parses comma-separated frame sizes with validation warnings.
func parseFrameSizes(s string) []int {
	parts := strings.Split(s, ",")
	sizes := make([]int, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		size, err := strconv.Atoi(trimmed)
		if err != nil {
			logging.Warn("invalid frame size ignored", "value", trimmed, "error", err)
			continue
		}
		if size < 64 || size > 9216 {
			logging.Warn("frame size out of range (64-9216), ignored", "value", size)
			continue
		}
		sizes = append(sizes, size)
	}
	return sizes
}

func testCmd(args []string) error {
	flags, err := parseTestFlags(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	if flags.iface == "" {
		_, _ = fmt.Fprintln(os.Stdout, "Error: --interface is required")
		return errors.New("missing interface")
	}

	// Validate test types.
	tests := strings.Split(flags.testTypes, ",")
	for i, t := range tests {
		tests[i] = strings.TrimSpace(t)
	}
	if !validateTestTypesList(tests) {
		return errors.New("invalid test types")
	}

	// Check license.
	if !checkTestLicense() {
		return errors.New("license check failed")
	}

	// Parse frame sizes.
	frameSizeList := parseFrameSizes(flags.frameSizes)
	if len(frameSizeList) == 0 {
		_, _ = fmt.Fprintln(os.Stdout, "Error: No valid frame sizes specified")
		return errors.New("no valid frame sizes")
	}

	printTestConfiguration(
		flags.iface, flags.testTypes, flags.frameSizes,
		flags.duration, flags.resolution, flags.maxLoss, flags.warmup,
	)

	// Create and configure dataplane context.
	ctx, ctxErr := testmasterDP.NewContext(flags.iface)
	if ctxErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to initialize dataplane: %v\n", ctxErr)
		return ctxErr
	}
	defer ctx.Close()

	cfg := createTestConfig(flags)
	cfgErr := ctx.Configure(cfg)
	if cfgErr != nil {
		ctx.Close()
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to configure: %v\n", cfgErr)
		return cfgErr
	}

	// Setup signal handler.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		_, _ = fmt.Fprintln(os.Stdout, "\nCancelling test...")
		ctx.Cancel()
	}()

	// Run tests.
	params := testCmdParams{
		cir:          flags.cir,
		eir:          flags.eir,
		fdThreshold:  flags.fdThreshold,
		fdvThreshold: flags.fdvThreshold,
		flrThreshold: flags.flrThreshold,
		duration:     flags.duration,
		jsonOutput:   flags.jsonOutput,
		csvOutput:    flags.csvOutput,
	}
	allResults := runTestSuite(ctx, tests, frameSizeList, params)

	// Final output.
	if flags.jsonOutput && len(allResults) > 0 {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		_, _ = fmt.Fprintf(os.Stdout, "\n%s\n", string(data))
	} else if flags.csvOutput && len(allResults) > 0 {
		printCSVResults(allResults)
	}

	_, _ = fmt.Fprintln(os.Stdout, "\nTest suite complete.")

	return nil
}
