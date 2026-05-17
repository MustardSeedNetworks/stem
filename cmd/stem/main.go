// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
//
// # The Stem - Network Performance Testing
//
// The main entry point for the stem command-line application.
// Provides subcommands for reflector mode, test master mode, web interface,
// TUI interface, and help/documentation access.
//
// Usage:
//
//	stem reflect --interface eth0       # Reflector mode (Tier 1)
//	stem test --type throughput         # Test Master mode (Tier 2)
//	stem web --port 8080                # WebUI
//	stem tui                            # Terminal UI
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/krisarmstrong/stem/internal/api"
	"github.com/krisarmstrong/stem/internal/auth"
	"github.com/krisarmstrong/stem/internal/help"
	"github.com/krisarmstrong/stem/internal/license"
	"github.com/krisarmstrong/stem/internal/logging"
	reflectorConfig "github.com/krisarmstrong/stem/internal/reflector/config"
	reflectorDP "github.com/krisarmstrong/stem/internal/reflector/dataplane"
	reflectorTUI "github.com/krisarmstrong/stem/internal/reflector/tui"
	modules "github.com/krisarmstrong/stem/internal/services"
	testmasterDP "github.com/krisarmstrong/stem/internal/services/orchestrator/dataplane"
	testmasterTUI "github.com/krisarmstrong/stem/internal/services/orchestrator/tui"
	"github.com/krisarmstrong/stem/internal/version"
)

// CLI constants.
const (
	ProductName            = "The Stem"
	Company                = "Mustard Seed Networks"
	DefaultProfile         = "all"
	DefaultReflectionMode  = "all"
	DefaultSignatureFilter = "all"
)

// Test result display constants.
const (
	resultPass = "PASS"
	resultFail = "FAIL"
)

// Default values for test configuration.
const (
	defaultTestDuration     = 60
	defaultResolution       = 0.1
	defaultMaxLoss          = 0.0
	defaultWarmup           = 2
	defaultTrials           = 3
	defaultFDThreshold      = 10.0
	defaultFDVThreshold     = 5.0
	defaultFLRThreshold     = 0.01
	defaultFrameSize        = 1518
	defaultWebPort          = 8080
	defaultLoadLevelStep    = 10.0
	defaultLoadLevelMax     = 100.0
	defaultBackToBackBurst  = 10000
	defaultOverloadRate     = 100.0
	defaultOverloadDuration = 60
	nsToUsConversion        = 1000.0
	trialWarningDays        = 3
)

// CLI formatting constants.
const (
	minArgsCount         = 2
	bannerWidth          = 60
	licenseBannerWidth   = 50
	statsIntervalSeconds = 5
	shutdownDelayMs      = 100
)

func main() {
	// Initialize structured logging.
	logLevel := os.Getenv("STEM_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logFormat := os.Getenv("STEM_LOG_FORMAT")
	if logFormat == "" {
		logFormat = "text" // Use text for CLI, json for production.
	}
	err := logging.Init(&logging.Config{
		Level:      logLevel,
		Format:     logFormat,
		AddSource:  false,
		File:       "",
		MaxSize:    0,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
		Component:  "stem",
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: logging initialization failed: %v\n", err)
		// Continue with default logging.
	}

	if len(os.Args) < minArgsCount {
		printUsage(os.Stdout)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "auth":
		authCmd(os.Args[2:])
	case "reflect":
		if cmdErr := reflectCmd(os.Args[2:]); cmdErr != nil {
			os.Exit(1)
		}
	case "test":
		if cmdErr := testCmd(os.Args[2:]); cmdErr != nil {
			os.Exit(1)
		}
	case "web":
		webCmd(os.Args[2:])
	case "tui":
		if cmdErr := tuiCmd(os.Args[2:]); cmdErr != nil {
			os.Exit(1)
		}
	case "license":
		licenseCmd(os.Args[2:])
	case "list-tests":
		listTestsCmd(os.Args[2:])
	case "help", "--help", "-h":
		helpCmd(os.Args[2:])
	case "tutorial":
		tutorialCmd(os.Args[2:])
	case "glossary":
		glossaryCmd(os.Args[2:])
	case "version", "--version", "-v":
		printVersion(os.Stdout)
	default:
		_, _ = fmt.Fprintf(os.Stdout, "Unknown command: %s\n", os.Args[1])
		printUsage(os.Stdout)
		os.Exit(1)
	}
}

func printVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "%s %s\n", ProductName, version.GetVersion())
	_, _ = fmt.Fprintf(w, "Commit: %s\n", version.GetCommit())
	_, _ = fmt.Fprintf(w, "Built:  %s\n", version.GetBuildTime())
	_, _ = fmt.Fprintf(w, "Copyright (c) 2025 %s\n", Company)
	_, _ = fmt.Fprintln(w, "Network Performance Testing")
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintf(w, `%s %s
%s - Network Performance Testing`, ProductName, version.GetVersion(), Company)
	_, _ = fmt.Fprint(w, `

USAGE:
    stem <command> [options]

COMMANDS:
    auth         Authentication management (reset-password)
    reflect      Start packet reflector (Tier 1 license)
    test         Run network tests (Tier 2 license required)
    web          Start WebUI server
    tui          Start terminal UI dashboard
    license      Manage license activation
    help         Get help on commands, tests, and concepts
    tutorial     Step-by-step learning guides
    glossary     Network terminology definitions
    list-tests   Show all available test types (--by-module for module view)
    version      Show version information

REFLECT OPTIONS:
    -i, --interface    Network interface to use (required)
    --profile          Preset profile: netally, msn, all, custom (default: all)
    --port             UDP port filter (default: any)
    --oui              OUI filter for MAC addresses

TEST OPTIONS:
    -i, --interface    Network interface to use (required)
    -t, --type         Test type (see 'stem list-tests' for all options)
    -d, --duration     Test duration in seconds (default: 60)
    --frame-sizes      Comma-separated frame sizes (default: 64,128,256,512,1024,1280,1518)
    --resolution       Binary search resolution %% (default: 0.1)
    --max-loss         Maximum acceptable loss %% (default: 0.0)
    --warmup           Warmup period in seconds (default: 2)
    --trials           Number of trials per test (default: 3)
    --json             Output results in JSON format
    --csv              Output results in CSV format

Y.1564 OPTIONS:
    --cir              Committed Information Rate in Mbps
    --eir              Excess Information Rate in Mbps
    --fd-threshold     Frame Delay threshold in ms (default: 10)
    --fdv-threshold    Frame Delay Variation threshold in ms (default: 5)
    --flr-threshold    Frame Loss Rate threshold %% (default: 0.01)

WEB OPTIONS:
    -p, --port         HTTP port (default: 8080)
    --host             Bind address (default: 0.0.0.0)

LICENSE OPTIONS:
    --activate <key>   Activate with license key
    --trial            Start 14-day trial
    --status           Show license status
    --deactivate       Remove license

EXAMPLES:
    # Reflector mode
    stem reflect -i eth0 --profile netally

    # RFC 2544 throughput test
    stem test -i eth0 -t throughput -d 60

    # RFC 2544 full suite
    stem test -i eth0 -t throughput,latency,frame_loss,back_to_back

    # Y.1564 service test
    stem test -i eth0 -t y1564 --cir 100 --eir 50

    # Start WebUI
    stem web -p 8080

    # License management
    stem license --status
    stem license --trial
    stem license --activate XXXX-XXXX-XXXX-XXXX

For more information: https://mustardseednetworks.com
`)
}

func listTestsCmd(args []string) {
	fs := flag.NewFlagSet("list-tests", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output in JSON format")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		// JSON output using modules.
		moduleInfos := modules.GetAllModuleInfos()
		data, _ := json.MarshalIndent(map[string]any{
			"modules": moduleInfos,
			"count":   len(moduleInfos),
		}, "", "  ")
		_, _ = fmt.Fprintln(os.Stdout, string(data))
		return
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s - Available Test Types by Module\n", ProductName)
	_, _ = fmt.Fprintln(os.Stdout, strings.Repeat("=", bannerWidth))
	hs := help.NewSystem()
	allMods := modules.GetAllModules()
	totalTests := 0

	for _, mod := range allMods {
		_, _ = fmt.Fprintf(
			os.Stdout,
			"\n%s [%s] (%s):\n",
			mod.DisplayName(),
			mod.Color(),
			mod.Standard(),
		)
		_, _ = fmt.Fprintf(os.Stdout, "  %s\n", mod.Description())
		_, _ = fmt.Fprintln(os.Stdout)
		for _, t := range mod.TestTypes() {
			desc := ""
			if test, ok := hs.Tests[t]; ok {
				desc = test.Summary
			}
			if desc == "" {
				_, _ = fmt.Fprintf(os.Stdout, "    %-20s\n", t)
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "    %-20s %s\n", t, desc)
			}
			totalTests++
		}
	}
	_, _ = fmt.Fprintf(
		os.Stdout,
		"\nTotal: %d test types across %d modules\n",
		totalTests,
		len(allMods),
	)
}

// getSignatureFilter maps profile name to signature filter.
func getSignatureFilter(profile string) string {
	switch profile {
	case "netally", "ito":
		return "ito"
	case "msn":
		return "msn"
	case "custom":
		return "custom"
	default:
		return DefaultSignatureFilter
	}
}

// reflectorStatsLoop displays reflector stats periodically until interrupted.
func reflectorStatsLoop(dp *reflectorDP.Dataplane) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(statsIntervalSeconds * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			_, _ = fmt.Fprintln(os.Stdout, "\nShutting down reflector...")
			go dp.Stop()
			time.Sleep(shutdownDelayMs * time.Millisecond)
			stats := dp.GetStats()
			_, _ = fmt.Fprintf(os.Stdout, "\nFinal Statistics:\n")
			_, _ = fmt.Fprintf(os.Stdout, "  Packets Received:  %d\n", stats.PacketsReceived)
			_, _ = fmt.Fprintf(os.Stdout, "  Packets Reflected: %d\n", stats.PacketsReflected)
			_, _ = fmt.Fprintf(os.Stdout, "  Bytes Received:    %d\n", stats.BytesReceived)
			_, _ = fmt.Fprintf(os.Stdout, "  Bytes Reflected:   %d\n", stats.BytesReflected)
			return
		case <-ticker.C:
			stats := dp.GetStats()
			_, _ = fmt.Fprintf(
				os.Stdout,
				"\r[Stats] RX: %d pkts | TX: %d pkts | Signatures: ITO=%d RFC2544=%d Y.1564=%d MSN=%d",
				stats.PacketsReceived,
				stats.PacketsReflected,
				stats.SigProbeOT+stats.SigDataOT+stats.SigLatency,
				stats.SigRFC2544,
				stats.SigY1564,
				stats.SigMSN,
			)
		}
	}
}

func reflectCmd(args []string) error {
	parsed, fs, parseErr := parseReflectFlags(args)
	if parseErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", parseErr)
		return parseErr
	}

	if err := requireReflectInterface(parsed.iface, fs); err != nil {
		return err
	}

	if err := checkReflectorLicense(); err != nil {
		return err
	}

	sigFilter := getSignatureFilter(parsed.profile)

	cfg := buildReflectorConfig(parsed, sigFilter)

	// Create reflector dataplane.
	dp, err := reflectorDP.New(cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to create reflector: %v\n", err)
		return err
	}
	defer dp.Close()

	// Start reflector.
	startErr := dp.Start()
	if startErr != nil {
		dp.Close() // Cleanup before exit.
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to start reflector: %v\n", startErr)
		return startErr
	}

	printReflectorStartup(parsed)
	_, _ = fmt.Fprintln(os.Stdout, "\nReflector started. Press Ctrl+C to stop.")

	if parsed.useTUI {
		tuiApp := reflectorTUI.New(dp)
		tuiErr := tuiApp.Run()
		if tuiErr != nil {
			_, _ = fmt.Fprintf(os.Stdout, "TUI error: %v\n", tuiErr)
		}
	} else {
		reflectorStatsLoop(dp)
	}

	return nil
}

type reflectCmdArgs struct {
	iface   string
	profile string
	oui     string
	port    uint16
	useTUI  bool
}

func parseReflectFlags(args []string) (*reflectCmdArgs, *flag.FlagSet, error) {
	fs := flag.NewFlagSet("reflect", flag.ExitOnError)
	iface := fs.String("interface", "", "Network interface")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")
	profile := fs.String("profile", DefaultProfile, "Preset profile")
	port := fs.Uint("port", 0, "UDP port filter")
	oui := fs.String("oui", "", "OUI filter")
	useTUI := fs.Bool("tui", false, "Launch TUI dashboard")

	if err := fs.Parse(args); err != nil {
		return nil, fs, err
	}

	if *port > math.MaxUint16 {
		return nil, fs, fmt.Errorf("port %d out of valid range (0-%d)", *port, math.MaxUint16)
	}

	return &reflectCmdArgs{
		iface:   *iface,
		profile: *profile,
		port:    uint16(*port),
		oui:     *oui,
		useTUI:  *useTUI,
	}, fs, nil
}

func requireReflectInterface(iface string, fs *flag.FlagSet) error {
	if iface == "" {
		_, _ = fmt.Fprintln(os.Stdout, "Error: --interface is required")
		fs.Usage()
		return errors.New("missing interface")
	}
	return nil
}

func checkReflectorLicense() error {
	// Check license (Tier 1 minimum).
	mgr, err := license.NewManager()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Warning: License check failed: %v\n", err)
		return nil
	}
	if mgr.IsActivated() {
		return nil
	}

	_, _ = fmt.Fprintln(os.Stdout, "No active license. Starting 14-day trial...")
	result := mgr.StartTrial()
	if !result.Success {
		_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
		return fmt.Errorf("license trial failed: %s", result.Message)
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s\n", result.Message)
	return nil
}

func buildReflectorConfig(parsed *reflectCmdArgs, sigFilter string) *reflectorConfig.Config {
	cfg := &reflectorConfig.Config{
		Interface:       parsed.iface,
		Verbose:         false,
		SignatureFilter: sigFilter,
		WebUI:           reflectorConfig.WebUIConfig{Enabled: false, Port: 0},
		TUI:             reflectorConfig.TUIConfig{Enabled: parsed.useTUI},
		Filtering: reflectorConfig.FilterConfig{
			Port:      parsed.port,
			FilterOUI: false,
			OUI:       "00:c0:17", // Default NetAlly OUI.
			FilterMAC: false,
		},
		Reflection: reflectorConfig.ReflectConfig{
			Mode: DefaultReflectionMode,
		},
		Platform: reflectorConfig.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
		Stats:    reflectorConfig.StatsConfig{Format: "text", Interval: 0},
	}

	if parsed.oui != "" {
		cfg.Filtering.FilterOUI = true
		cfg.Filtering.OUI = parsed.oui
	}

	return cfg
}

func printReflectorStartup(parsed *reflectCmdArgs) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s - Reflector\n", ProductName, version.GetVersion())
	_, _ = fmt.Fprintf(os.Stdout, "Interface:  %s\n", parsed.iface)
	_, _ = fmt.Fprintf(os.Stdout, "Profile:    %s\n", parsed.profile)
	if parsed.port > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "Port:       %d\n", parsed.port)
	}
	if parsed.oui != "" {
		_, _ = fmt.Fprintf(os.Stdout, "OUI:        %s\n", parsed.oui)
	}
}

// validateTestTypesList validates a list of test types.
func validateTestTypesList(tests []string) bool {
	for _, t := range tests {
		if mod := modules.GetModuleForTest(t); mod == nil {
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
	case state.Tier < license.TierTestSuite && !state.IsTrialMode:
		_, _ = fmt.Fprintln(os.Stdout, "Error: Test Suite requires Tier 2 license")
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

// parseTestFlags parses test command flags and returns the parsed values.
func parseTestFlags(args []string) (*testCmdFlags, error) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	// Basic options.
	iface := fs.String("interface", "", "Network interface")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")
	testTypes := fs.String("type", "throughput", "Test type(s), comma-separated")
	fs.StringVar(testTypes, "t", "throughput", "Test type (shorthand)")
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

// runTest executes a single test and returns the result.
func runTest(
	ctx *testmasterDP.Context,
	testType string,
	cir, eir, fdThreshold, fdvThreshold, flrThreshold float64,
	duration int,
) (any, error) {
	switch testType {
	case "throughput":
		return runThroughputTest(ctx)
	case "latency":
		return runLatencyTest(ctx)
	case "frame_loss":
		return runFrameLossTest(ctx)
	case "back_to_back":
		return runBackToBackTest(ctx)
	case "system_recovery":
		return runSystemRecoveryTest(ctx)
	case "reset":
		return runResetTest(ctx)
	case "y1564_config", "y1564":
		return runY1564ConfigTest(ctx, cir, eir, fdThreshold, fdvThreshold, flrThreshold)
	case "y1564_perf":
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

func webCmd(args []string) {
	fs := flag.NewFlagSet("web", flag.ExitOnError)
	port := fs.Int("port", defaultWebPort, "HTTP port (1-65535)")
	fs.IntVar(port, "p", defaultWebPort, "HTTP port (shorthand)")
	host := fs.String("host", "0.0.0.0", "Bind address")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate port range.
	if *port < 1 || *port > 65535 {
		_, _ = fmt.Fprintf(os.Stderr, "Error: port must be between 1 and 65535, got %d\n", *port)
		os.Exit(1)
	}

<<<<<<< Updated upstream
	_, _ = fmt.Fprintf(os.Stdout, "%s %s - WebUI Server\n", ProductName, version.GetVersion())
||||||| Stash base
	_, _ = fmt.Fprintf(os.Stdout, "%s %s - WebUI Server\n", ProductName, version.Version())
=======
	// First boot: auto-generate credentials if not configured.
	ensureStemCredentials(*port)

	_, _ = fmt.Fprintf(os.Stdout, "%s %s - WebUI Server\n", ProductName, version.Version())
>>>>>>> Stashed changes
	_, _ = fmt.Fprintf(os.Stdout, "Starting on http://%s:%d\n", *host, *port)

	srv, err := api.NewServer(*port)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Hint: Set STEM_AUTH_USERNAME and STEM_AUTH_PASSWORD environment variables\n",
		)
		os.Exit(1)
	}
	err = srv.Run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

const (
	// stemEnvironmentFile is where auto-generated credentials are persisted.
	stemEnvironmentFile = "/etc/stem/environment"
	// stemPasswordLength is the length for auto-generated passwords.
	stemPasswordLength = 16
)

// ensureStemCredentials generates credentials on first boot if not configured.
func ensureStemCredentials(port int) {
	if os.Getenv("STEM_AUTH_PASSWORD") != "" {
		return
	}

	password, err := auth.GenerateSecurePassword(stemPasswordLength)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Fatal: Failed to generate password: %v\n", err)
		os.Exit(1)
	}

	username := os.Getenv("STEM_AUTH_USERNAME")
	if username == "" {
		username = "admin"
		os.Setenv("STEM_AUTH_USERNAME", username)
	}
	os.Setenv("STEM_AUTH_PASSWORD", password)

	writeEnvironmentFile(username, password)
	printFirstBootCredentials(username, password, port)
}

// writeEnvironmentFile persists credentials to the systemd environment file.
func writeEnvironmentFile(username, password string) {
	dir := "/etc/stem"
	if mkdirErr := os.MkdirAll(dir, 0o750); mkdirErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to create %s: %v\n", dir, mkdirErr)
		return
	}

	content := fmt.Sprintf("STEM_AUTH_USERNAME=%s\nSTEM_AUTH_PASSWORD=%s\n", username, password)
	if writeErr := os.WriteFile(stemEnvironmentFile, []byte(content), 0o600); writeErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to write environment file: %v\n", writeErr)
	}
}

// printFirstBootCredentials prints the auto-generated credentials banner.
func printFirstBootCredentials(username, password string, port int) {
	_, _ = fmt.Fprintf(os.Stderr, `
╔══════════════════════════════════════════════════════════════════╗
║              THE STEM - FIRST BOOT CREDENTIALS                   ║
║              Mustard Seed Networks                               ║
╠══════════════════════════════════════════════════════════════════╣
║  Username: %-53s ║
║  Password: %-53s ║
║  Web UI:   http://localhost:%-38d ║
║                                                                  ║
║  IMPORTANT: Save this password now!                              ║
║  Use 'stem auth reset-password' to generate a new one.           ║
╚══════════════════════════════════════════════════════════════════╝
`, username, password, port)
}

// authCmd handles the auth subcommand.
func authCmd(args []string) {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "Usage: stem auth <subcommand>")
		_, _ = fmt.Fprintln(os.Stderr, "")
		_, _ = fmt.Fprintln(os.Stderr, "Subcommands:")
		_, _ = fmt.Fprintln(os.Stderr, "  reset-password   Generate a new admin password")
		os.Exit(1)
	}

	switch args[0] {
	case "reset-password":
		authResetPassword()
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Unknown auth subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

// authResetPassword generates a new password and writes it to the environment file.
func authResetPassword() {
	password, err := auth.GenerateSecurePassword(stemPasswordLength)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Failed to generate password: %v\n", err)
		os.Exit(1)
	}

	username := os.Getenv("STEM_AUTH_USERNAME")
	if username == "" {
		username = "admin"
	}

	writeEnvironmentFile(username, password)

	_, _ = fmt.Fprintln(os.Stdout, "╔══════════════════════════════════════════════════════════════════╗")
	_, _ = fmt.Fprintln(os.Stdout, "║              THE STEM - PASSWORD RESET                           ║")
	_, _ = fmt.Fprintln(os.Stdout, "╠══════════════════════════════════════════════════════════════════╣")
	_, _ = fmt.Fprintf(os.Stdout, "║  Username: %-53s ║\n", username)
	_, _ = fmt.Fprintf(os.Stdout, "║  Password: %-53s ║\n", password)
	_, _ = fmt.Fprintln(os.Stdout, "║                                                                  ║")
	_, _ = fmt.Fprintln(os.Stdout, "║  Restart the service for changes to take effect:                 ║")
	_, _ = fmt.Fprintln(os.Stdout, "║    sudo systemctl restart stem                                   ║")
	_, _ = fmt.Fprintln(os.Stdout, "╚══════════════════════════════════════════════════════════════════╝")
}

// tuiReflectMode runs the reflector TUI mode.
func tuiReflectMode(iface string) error {
	if iface == "" {
		_, _ = fmt.Fprintln(os.Stdout, "Error: --interface is required for reflect mode")
		_, _ = fmt.Fprintln(os.Stdout, "Usage: stem tui --mode reflect -i eth0")
		return errors.New("missing interface")
	}

	// Check license (Tier 1 minimum).
	mgr, err := license.NewManager()
	if err != nil {
		logging.Warn("license manager initialization failed", "error", err)
	}
	if mgr != nil && !mgr.IsActivated() {
		result := mgr.StartTrial()
		if !result.Success {
			_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
			return fmt.Errorf("license trial failed: %s", result.Message)
		}
	}

	// Build reflector config.
	cfg := &reflectorConfig.Config{
		Interface:       iface,
		Verbose:         false,
		SignatureFilter: DefaultSignatureFilter,
		WebUI:           reflectorConfig.WebUIConfig{Enabled: false, Port: 0},
		TUI:             reflectorConfig.TUIConfig{Enabled: true},
		Filtering: reflectorConfig.FilterConfig{
			Port:      0,
			FilterOUI: false,
			OUI:       "",
			FilterMAC: false,
		},
		Reflection: reflectorConfig.ReflectConfig{
			Mode: DefaultReflectionMode,
		},
		Platform: reflectorConfig.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
		Stats:    reflectorConfig.StatsConfig{Format: "text", Interval: 0},
	}

	// Create and start reflector dataplane.
	dp, dpErr := reflectorDP.New(cfg)
	if dpErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to create reflector: %v\n", dpErr)
		return dpErr
	}
	defer dp.Close()

	tuiStartErr := dp.Start()
	if tuiStartErr != nil {
		dp.Close() // Cleanup before exit.
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to start reflector: %v\n", tuiStartErr)
		return tuiStartErr
	}

	// Launch reflector TUI.
	tuiApp := reflectorTUI.New(dp)
	tuiRunErr := tuiApp.Run()
	if tuiRunErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "TUI error: %v\n", tuiRunErr)
		return tuiRunErr
	}

	return nil
}

// tuiTestMode runs the testmaster TUI mode.
func tuiTestMode() error {
	// Check license (Tier 2 required).
	mgr, mgrErr := license.NewManager()
	if mgrErr != nil {
		logging.Warn("license manager initialization failed", "error", mgrErr)
	}
	if mgr != nil {
		state := mgr.GetState()
		if state == nil {
			result := mgr.StartTrial()
			if !result.Success {
				_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
				return fmt.Errorf("license trial failed: %s", result.Message)
			}
		} else if state.Tier < license.TierTestSuite && !state.IsTrialMode {
			_, _ = fmt.Fprintln(os.Stdout, "Error: Test Suite TUI requires Tier 2 license")
			return errors.New("license tier too low")
		}
	}

	// Launch testmaster TUI.
	tuiApp := testmasterTUI.New()

	// Set up callbacks.
	tuiApp.OnQuit = func() {
		tuiApp.Stop()
	}

	tuiApp.Logf("The Stem TUI started")
	tuiApp.Logf("Press F1 to start test, F2 to stop, F10 to quit")

	tuiRunErr := tuiApp.Run()
	if tuiRunErr != nil {
		_, _ = fmt.Fprintf(os.Stdout, "TUI error: %v\n", tuiRunErr)
		return tuiRunErr
	}

	return nil
}

func tuiCmd(args []string) error {
	fs := flag.NewFlagSet("tui", flag.ExitOnError)
	mode := fs.String("mode", "test", "TUI mode: test or reflect")
	iface := fs.String("interface", "", "Network interface (required for reflect mode)")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s %s - Terminal UI\n", ProductName, version.GetVersion())

	switch *mode {
	case "reflect", "reflector":
		if modeErr := tuiReflectMode(*iface); modeErr != nil {
			return modeErr
		}
	case "test", "testmaster", "":
		if modeErr := tuiTestMode(); modeErr != nil {
			return modeErr
		}
	default:
		_, _ = fmt.Fprintf(os.Stdout, "Error: Unknown TUI mode '%s'\n", *mode)
		_, _ = fmt.Fprintln(os.Stdout, "Valid modes: test, reflect")
		return fmt.Errorf("invalid TUI mode: %s", *mode)
	}

	return nil
}

func helpCmd(args []string) {
	fs := flag.NewFlagSet("help", flag.ExitOnError)
	simple := fs.Bool("simple", false, "Show simplified explanations for non-technical users")
	fs.BoolVar(simple, "s", false, "Show simplified explanations (shorthand)")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If no topic specified, show general help.
	if fs.NArg() == 0 {
		printUsage(os.Stdout)
		return
	}

	topic := strings.ToLower(fs.Arg(0))

	// Try to find the topic in our help system.
	if help.ShowHelp(topic, *simple) {
		return
	}

	// Not found in help system.
	_, _ = fmt.Fprintf(os.Stdout, "No help found for '%s'\n\n", topic)
	_, _ = fmt.Fprintln(os.Stdout, "Available help topics:")
	_, _ = fmt.Fprintln(os.Stdout, "  Commands:   reflect, test, web, license")
	_, _ = fmt.Fprintln(
		os.Stdout,
		"  Tests:      throughput, latency, frame_loss, y1564_config, ...",
	)
	_, _ = fmt.Fprintln(
		os.Stdout,
		"  Categories: rfc2544, y1564, rfc2889, rfc6349, y1731, mef, tsn",
	)
	_, _ = fmt.Fprintln(os.Stdout, "\nUse 'stem help tests' for a complete list of tests.")
	_, _ = fmt.Fprintln(os.Stdout, "Use 'stem glossary' for network terminology definitions.")
	_, _ = fmt.Fprintln(os.Stdout, "Use 'stem tutorial' for step-by-step guides.")
}

func tutorialCmd(args []string) {
	// If no tutorial specified, list available tutorials.
	if len(args) == 0 {
		help.ShowTutorial("")
		return
	}

	tutorialID := strings.ToLower(strings.TrimSpace(args[0]))

	if !help.ShowTutorial(tutorialID) {
		_, _ = fmt.Fprintf(os.Stdout, "Tutorial '%s' not found.\n\n", tutorialID)
		help.ShowTutorial("") // Show available tutorials.
	}
}

func glossaryCmd(args []string) {
	fs := flag.NewFlagSet("glossary", flag.ExitOnError)
	simple := fs.Bool("simple", false, "Show only simple definitions")
	fs.BoolVar(simple, "s", false, "Show only simple definitions (shorthand)")
	search := fs.String("search", "", "Search for terms containing keyword")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Search mode.
	if *search != "" {
		hs := help.NewSystem()
		results := hs.SearchGlossary(*search)
		if len(results) == 0 {
			_, _ = fmt.Fprintf(os.Stdout, "No terms found matching '%s'\n", *search)
			return
		}
		_, _ = fmt.Fprintf(os.Stdout, "Terms matching '%s':\n\n", *search)
		for _, entry := range results {
			_, _ = fmt.Fprintf(os.Stdout, "  %s - %s\n", entry.Term, entry.FullName)
		}
		_, _ = fmt.Fprintln(os.Stdout, "\nUse 'stem glossary <term>' for full definition.")
		return
	}

	// If no term specified, list all terms.
	if fs.NArg() == 0 {
		help.ShowGlossary("", *simple)
		return
	}

	term := strings.ToLower(fs.Arg(0))

	if !help.ShowGlossary(term, *simple) {
		_, _ = fmt.Fprintf(os.Stdout, "Term '%s' not found in glossary.\n\n", term)
		_, _ = fmt.Fprintln(os.Stdout, "Use 'stem glossary' to see all available terms.")
		_, _ = fmt.Fprintln(os.Stdout, "Use 'stem glossary --search <keyword>' to search.")
	}
}

// displayLicenseStatus displays the license status.
func displayLicenseStatus(mgr *license.Manager) {
	state := mgr.GetState()
	fp := mgr.GetFingerprint()

	_, _ = fmt.Fprintf(os.Stdout, "%s - License Status\n", ProductName)
	_, _ = fmt.Fprintln(os.Stdout, strings.Repeat("=", licenseBannerWidth))

	switch {
	case state == nil:
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Not Activated")
		_, _ = fmt.Fprintln(os.Stdout, "\nTo start a 14-day trial:")
		_, _ = fmt.Fprintln(os.Stdout, "  stem license --trial")
		_, _ = fmt.Fprintln(os.Stdout, "\nTo activate with a license key:")
		_, _ = fmt.Fprintln(os.Stdout, "  stem license --activate XXXX-XXXX-XXXX-XXXX")
	case state.IsTrialMode:
		remaining := mgr.TrialDaysRemaining()
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Trial Mode")
		_, _ = fmt.Fprintf(os.Stdout, "Days Left: %d\n", remaining)
		_, _ = fmt.Fprintf(os.Stdout, "Tier:      %s (full access during trial)\n", state.Tier)
		if remaining <= trialWarningDays {
			_, _ = fmt.Fprintln(os.Stdout, "\nWarning: Trial ending soon!")
			_, _ = fmt.Fprintln(os.Stdout, "Activate a license to continue using The Stem")
		}
	default:
		_, _ = fmt.Fprintln(os.Stdout, "Status:    Licensed")
		_, _ = fmt.Fprintf(os.Stdout, "Tier:      %s\n", state.Tier)
		_, _ = fmt.Fprintf(os.Stdout, "Key:       %s\n", license.FormatKey(state.LicenseKey))
		_, _ = fmt.Fprintf(os.Stdout, "Expires:   %s\n", state.ExpiresAt.Format("2006-01-02"))
	}

	_, _ = fmt.Fprintf(os.Stdout, "\nDevice ID: %s\n", fp.Hash())
	_, _ = fmt.Fprintf(os.Stdout, "Platform:  %s\n", fp.Platform)

	if state != nil && len(state.Features) > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "\nEnabled Features:\n")
		for _, f := range state.Features {
			_, _ = fmt.Fprintf(os.Stdout, "  - %s\n", f)
		}
	}
}

func licenseCmd(args []string) {
	fs := flag.NewFlagSet("license", flag.ExitOnError)
	activate := fs.String("activate", "", "Activate with license key")
	trial := fs.Bool("trial", false, "Start 14-day trial")
	status := fs.Bool("status", false, "Show license status")
	deactivate := fs.Bool("deactivate", false, "Remove license")

	err := fs.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	mgr, err := license.NewManager()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to initialize license manager: %v\n", err)
		os.Exit(1)
	}

	switch {
	case *activate != "":
		result := mgr.Activate(*activate)
		if result.Success {
			_, _ = fmt.Fprintf(os.Stdout, "Success: %s\n", result.Message)
			_, _ = fmt.Fprintf(os.Stdout, "Tier: %s\n", result.Tier)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
			os.Exit(1)
		}

	case *trial:
		result := mgr.StartTrial()
		if result.Success {
			_, _ = fmt.Fprintf(os.Stdout, "Success: %s\n", result.Message)
			if result.DaysRemaining > 0 {
				_, _ = fmt.Fprintf(os.Stdout, "Days remaining: %d\n", result.DaysRemaining)
			}
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Message)
			os.Exit(1)
		}

	case *deactivate:
		deactErr := mgr.Deactivate()
		if deactErr != nil {
			_, _ = fmt.Fprintf(os.Stdout, "Error: Failed to deactivate: %v\n", deactErr)
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stdout, "License deactivated successfully")

	case *status:
		fallthrough
	default:
		displayLicenseStatus(mgr)
	}
}
