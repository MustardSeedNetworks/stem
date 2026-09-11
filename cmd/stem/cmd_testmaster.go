// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/MustardSeedNetworks/stem/internal/api"
	"github.com/MustardSeedNetworks/stem/internal/daemonclient"
	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
	"github.com/MustardSeedNetworks/stem/internal/logging"
	"github.com/MustardSeedNetworks/stem/internal/services"
	"github.com/MustardSeedNetworks/stem/internal/version"
)

// testCmdFlags holds parsed command line flags for test command.
type testCmdFlags struct {
	iface        string
	peer         string
	peerPort     uint16
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
	peer := fs.String("peer", "", "Reflector host or IPv4 address")
	peerPort := uint16(defaultPeerPort)
	fs.Func("peer-port", "Reflector UDP port", func(value string) error {
		parsed, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid reflector UDP port: %w", err)
		}
		peerPort = uint16(parsed)
		return nil
	})
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
		peer:         *peer,
		peerPort:     peerPort,
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

// parseFrameSizes parses comma-separated frame sizes with validation warnings.
func parseFrameSizes(s string) []uint32 {
	parts := strings.Split(s, ",")
	sizes := make([]uint32, 0, len(parts))
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
		sizes = append(sizes, uint32(size))
	}
	return sizes
}

// buildStartRequest turns parsed flags into the run plan the daemon
// executes. The daemon owns defaults it can see (line rate, interface
// validation); the CLI only forwards what the operator asked for.
func buildStartRequest(
	flags *testCmdFlags,
	tests []string,
	frameSizes []uint32,
	seconds uint32,
) api.TestStartRequest {
	steps := make([]api.TestStepRequest, 0, len(tests))
	for _, testType := range tests {
		steps = append(steps, api.TestStepRequest{
			TestType: testType,
			Config: &api.TestConfig{
				RFC2544: &api.RFC2544TestConfig{
					Duration:   flags.duration,
					FrameSizes: frameSizes,
					Resolution: flags.resolution,
					MaxLoss:    flags.maxLoss,
					Warmup:     flags.warmup,
					Trials:     defaultTrials,
				},
				Y1564: &api.Y1564TestConfig{
					CIR:                flags.cir,
					EIR:                flags.eir,
					FrameSizes:         frameSizes,
					FDThreshold:        flags.fdThreshold,
					FDVThreshold:       flags.fdvThreshold,
					FLRThreshold:       flags.flrThreshold,
					ConfigStepDuration: seconds,
					PerfTestDuration:   seconds,
				},
			},
		})
	}
	return api.TestStartRequest{
		Interface: flags.iface,
		Peer:      flags.peer,
		PeerPort:  flags.peerPort,
		Tests:     steps,
	}
}

// maxDurationSeconds bounds a single test step at one day. The flag is an
// int, so without a bound the conversion to the wire's uint32 would wrap a
// negative or absurd duration into a plausible-looking one.
const maxDurationSeconds = 86400

// validateDuration refuses a duration the wire format cannot carry
// faithfully and returns the checked value, rather than silently measuring
// something else. The bound and the conversion live together so the
// conversion is provably in range.
func validateDuration(duration int) (uint32, error) {
	if duration <= 0 || duration > maxDurationSeconds {
		return 0, fmt.Errorf("duration must be between 1 and %d seconds, got %d", maxDurationSeconds, duration)
	}
	return uint32(duration), nil
}

// reportNoDaemon explains the one thing that stops the CLI working. There
// is deliberately no standalone fallback: a CLI that quietly opened its own
// dataplane would be the second orchestration path #1166 removes, and it
// would fight the daemon for the interface.
func reportNoDaemon(err error) error {
	_, _ = fmt.Fprintf(os.Stdout, "Error: %v\n", err)
	switch {
	case errors.Is(err, daemonconn.ErrUnreadable):
		_, _ = fmt.Fprintln(os.Stdout,
			"The daemon's credential is readable only by the account it runs as. "+
				"Run this as root (sudo stem test ...) or as the stem user.")
	case errors.Is(err, daemonconn.ErrPermissions):
		_, _ = fmt.Fprintln(os.Stdout,
			"Another account can read the daemon's credential, so it will not be used. "+
				"Restore it with 'chmod 600' and restart the daemon to reissue the token.")
	default:
		_, _ = fmt.Fprintln(os.Stdout,
			"Tests run in the Stem daemon. Start it with 'systemctl start stem' "+
				"(or 'stem web') and try again.")
		_, _ = fmt.Fprintf(os.Stdout, "Looked in: %s\n", strings.Join(daemonconn.SearchDirs(), ", "))
	}
	return err
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

	seconds, durErr := validateDuration(flags.duration)
	if durErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: %v\n", durErr)
		return durErr
	}

	tests := strings.Split(flags.testTypes, ",")
	for i, t := range tests {
		tests[i] = strings.TrimSpace(t)
	}
	if !validateTestTypesList(tests) {
		return errors.New("invalid test types")
	}

	frameSizeList := parseFrameSizes(flags.frameSizes)
	if len(frameSizeList) == 0 {
		_, _ = fmt.Fprintln(os.Stdout, "Error: No valid frame sizes specified")
		return errors.New("no valid frame sizes")
	}

	// Entitlement is the daemon's answer, not the CLI's. Checking a licence
	// here as well would be a second decision to keep in step — and the old
	// local check started a trial as a side effect of running a test.
	client, err := daemonclient.Discover()
	if err != nil {
		return reportNoDaemon(err)
	}

	printTestConfiguration(
		flags.iface, flags.testTypes, flags.frameSizes,
		flags.duration, flags.resolution, flags.maxLoss, flags.warmup,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runID, startErr := client.Start(ctx, buildStartRequest(flags, tests, frameSizeList, seconds))
	if startErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: %v\n", startErr)
		return startErr
	}
	_, _ = fmt.Fprintf(os.Stdout, "Run %s started\n", runID)

	quiet := flags.jsonOutput || flags.csvOutput
	status, watchErr := watchRun(ctx, client, runID, quiet)
	if watchErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "\nError: %v\n", watchErr)
		return watchErr
	}

	result, resultErr := client.Result(ctx)
	if resultErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "\nError: %v\n", resultErr)
		return resultErr
	}
	return reportRunOutcome(status, result, flags)
}

// reportRunOutcome prints the daemon's result in the requested format and
// turns a failed or cancelled run into a non-zero exit.
func reportRunOutcome(status api.Stats, result api.TestResultResponse, flags *testCmdFlags) error {
	switch {
	case flags.jsonOutput:
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("encode result: %w", err)
		}
		_, _ = fmt.Fprintf(os.Stdout, "\n%s\n", string(data))
	case flags.csvOutput:
		renderCSV(result.Data)
	default:
		_, _ = fmt.Fprintf(os.Stdout, "\n\nRun %s %s\n", status.SuiteID, status.TestStatus)
		if result.Message != "" {
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", result.Message)
		}
		renderResult(result.Data)
	}

	if status.TestStatus != "completed" || !result.Success {
		if result.Error != "" {
			return fmt.Errorf("run %s: %s", status.SuiteID, result.Error)
		}
		return fmt.Errorf("run %s ended %s", status.SuiteID, status.TestStatus)
	}
	return nil
}
