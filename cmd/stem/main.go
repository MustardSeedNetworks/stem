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
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/krisarmstrong/stem/internal/help"
	"github.com/krisarmstrong/stem/internal/license"
	"github.com/krisarmstrong/stem/internal/version"
	reflectorConfig "github.com/krisarmstrong/stem/internal/reflector/config"
	reflectorDP "github.com/krisarmstrong/stem/internal/reflector/dataplane"
	reflectorTUI "github.com/krisarmstrong/stem/internal/reflector/tui"
	testmasterDP "github.com/krisarmstrong/stem/internal/testmaster/dataplane"
	testmasterTUI "github.com/krisarmstrong/stem/internal/testmaster/tui"
	"github.com/krisarmstrong/stem/internal/web"
)

const (
	ProductName = "The Stem"
	Company     = "Mustard Seed Networks"
)

// All supported test types
var allTestTypes = map[string]string{
	// RFC 2544 (6 tests)
	"throughput":      "RFC 2544 Section 26.1 - Maximum throughput with zero loss",
	"latency":         "RFC 2544 Section 26.2 - Round-trip latency at various loads",
	"frame_loss":      "RFC 2544 Section 26.3 - Frame loss rate vs offered load",
	"back_to_back":    "RFC 2544 Section 26.4 - Maximum burst capacity",
	"system_recovery": "RFC 2544 Section 26.5 - Recovery time after overload",
	"reset":           "RFC 2544 Section 26.6 - Device reset recovery time",

	// Y.1564 / EtherSAM (3 tests)
	"y1564_config": "ITU-T Y.1564 Service Configuration Test",
	"y1564_perf":   "ITU-T Y.1564 Service Performance Test (15+ min)",
	"y1564":        "ITU-T Y.1564 Full Test (config + performance)",

	// RFC 2889 LAN Switch (5 tests)
	"rfc2889_forwarding": "RFC 2889 Forwarding rate test",
	"rfc2889_caching":    "RFC 2889 Address caching capacity",
	"rfc2889_learning":   "RFC 2889 Address learning rate",
	"rfc2889_broadcast":  "RFC 2889 Broadcast forwarding",
	"rfc2889_congestion": "RFC 2889 Congestion control",

	// RFC 6349 TCP (2 tests)
	"rfc6349_throughput": "RFC 6349 TCP throughput (BDP analysis)",
	"rfc6349_path":       "RFC 6349 Path analysis (RTT/bandwidth)",

	// Y.1731 OAM (4 tests)
	"y1731_delay":    "ITU-T Y.1731 Frame Delay (DMM/DMR)",
	"y1731_loss":     "ITU-T Y.1731 Frame Loss (LMM/LMR)",
	"y1731_slm":      "ITU-T Y.1731 Synthetic Loss Measurement",
	"y1731_loopback": "ITU-T Y.1731 Loopback (LBM/LBR)",

	// MEF Service (3 tests)
	"mef_config": "MEF 48/49 Service Configuration Test",
	"mef_perf":   "MEF 48/49 Service Performance Test",
	"mef":        "MEF 48/49 Full Test Suite",

	// TSN 802.1Qbv (4 tests)
	"tsn_timing":    "IEEE 802.1Qbv Gate timing accuracy",
	"tsn_isolation": "IEEE 802.1Qbv Traffic class isolation",
	"tsn_latency":   "IEEE 802.1Qbv Scheduled latency",
	"tsn":           "IEEE 802.1Qbv Full TSN test suite",
}

// Test categories for help display
var testCategories = []struct {
	name  string
	tests []string
}{
	{"RFC 2544", []string{"throughput", "latency", "frame_loss", "back_to_back", "system_recovery", "reset"}},
	{"Y.1564 EtherSAM", []string{"y1564_config", "y1564_perf", "y1564"}},
	{"RFC 2889 LAN Switch", []string{"rfc2889_forwarding", "rfc2889_caching", "rfc2889_learning", "rfc2889_broadcast", "rfc2889_congestion"}},
	{"RFC 6349 TCP", []string{"rfc6349_throughput", "rfc6349_path"}},
	{"Y.1731 OAM", []string{"y1731_delay", "y1731_loss", "y1731_slm", "y1731_loopback"}},
	{"MEF Service", []string{"mef_config", "mef_perf", "mef"}},
	{"TSN 802.1Qbv", []string{"tsn_timing", "tsn_isolation", "tsn_latency", "tsn"}},
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "reflect":
		reflectCmd(os.Args[2:])
	case "test":
		testCmd(os.Args[2:])
	case "web":
		webCmd(os.Args[2:])
	case "tui":
		tuiCmd(os.Args[2:])
	case "license":
		licenseCmd(os.Args[2:])
	case "list-tests":
		listTestsCmd()
	case "help", "--help", "-h":
		helpCmd(os.Args[2:])
	case "tutorial":
		tutorialCmd(os.Args[2:])
	case "glossary":
		glossaryCmd(os.Args[2:])
	case "version", "--version", "-v":
		printVersion()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("%s %s\n", ProductName, version.Version)
	fmt.Printf("Commit: %s\n", version.Commit)
	fmt.Printf("Built:  %s\n", version.BuildTime)
	fmt.Printf("Copyright (c) 2025 %s\n", Company)
	fmt.Println("Network Performance Testing")
}

func printUsage() {
	fmt.Printf(`%s %s
%s - Network Performance Testing`, ProductName, version.Version, Company)
	fmt.Print(`

USAGE:
    stem <command> [options]

COMMANDS:
    reflect      Start packet reflector (Tier 1 license)
    test         Run network tests (Tier 2 license required)
    web          Start WebUI server
    tui          Start terminal UI dashboard
    license      Manage license activation
    help         Get help on commands, tests, and concepts
    tutorial     Step-by-step learning guides
    glossary     Network terminology definitions
    list-tests   Show all available test types
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

func listTestsCmd() {
	fmt.Printf("%s - Available Test Types\n", ProductName)
	fmt.Println(strings.Repeat("=", 60))

	for _, cat := range testCategories {
		fmt.Printf("\n%s:\n", cat.name)
		for _, t := range cat.tests {
			desc := allTestTypes[t]
			fmt.Printf("  %-20s %s\n", t, desc)
		}
	}

	fmt.Printf("\nTotal: %d test types across %d categories\n", len(allTestTypes), len(testCategories))
}

func reflectCmd(args []string) {
	fs := flag.NewFlagSet("reflect", flag.ExitOnError)
	iface := fs.String("interface", "", "Network interface")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")
	profile := fs.String("profile", "all", "Preset profile")
	port := fs.Int("port", 0, "UDP port filter")
	oui := fs.String("oui", "", "OUI filter")
	useTUI := fs.Bool("tui", false, "Launch TUI dashboard")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *iface == "" {
		fmt.Println("Error: --interface is required")
		fs.Usage()
		os.Exit(1)
	}

	// Check license (Tier 1 minimum)
	mgr, err := license.NewManager()
	if err != nil {
		fmt.Printf("Warning: License check failed: %v\n", err)
	} else if !mgr.IsActivated() {
		fmt.Println("No active license. Starting 14-day trial...")
		result := mgr.StartTrial()
		if !result.Success {
			fmt.Printf("Error: %s\n", result.Message)
			os.Exit(1)
		}
		fmt.Printf("%s\n", result.Message)
	}

	// Map profile to signature filter
	sigFilter := "all"
	switch *profile {
	case "netally", "ito":
		sigFilter = "ito"
	case "msn":
		sigFilter = "msn"
	case "custom":
		sigFilter = "custom"
	case "all":
		sigFilter = "all"
	}

	// Build reflector config with defaults
	cfg := &reflectorConfig.Config{
		Interface:       *iface,
		SignatureFilter: sigFilter,
		Filtering: reflectorConfig.FilterConfig{
			Port: uint16(*port),
			OUI:  "00:c0:17", // Default NetAlly OUI
		},
		Reflection: reflectorConfig.ReflectConfig{
			Mode: "all",
		},
	}

	if *oui != "" {
		cfg.Filtering.FilterOUI = true
		cfg.Filtering.OUI = *oui
	}

	// Create reflector dataplane
	dp, err := reflectorDP.New(cfg)
	if err != nil {
		fmt.Printf("Error: Failed to create reflector: %v\n", err)
		os.Exit(1)
	}
	defer dp.Close()

	// Start reflector
	if err := dp.Start(); err != nil {
		fmt.Printf("Error: Failed to start reflector: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s - Reflector\n", ProductName, version.Version)
	fmt.Printf("Interface:  %s\n", *iface)
	fmt.Printf("Profile:    %s\n", *profile)
	if *port > 0 {
		fmt.Printf("Port:       %d\n", *port)
	}
	if *oui != "" {
		fmt.Printf("OUI:        %s\n", *oui)
	}
	fmt.Println("\nReflector started. Press Ctrl+C to stop.")

	if *useTUI {
		// Launch TUI
		tuiApp := reflectorTUI.New(dp)
		if err := tuiApp.Run(); err != nil {
			fmt.Printf("TUI error: %v\n", err)
		}
	} else {
		// Wait for interrupt
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Print stats periodically
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-sigChan:
				fmt.Println("\nShutting down reflector...")
				dp.Stop()
				stats := dp.GetStats()
				fmt.Printf("\nFinal Statistics:\n")
				fmt.Printf("  Packets Received:  %d\n", stats.PacketsReceived)
				fmt.Printf("  Packets Reflected: %d\n", stats.PacketsReflected)
				fmt.Printf("  Bytes Received:    %d\n", stats.BytesReceived)
				fmt.Printf("  Bytes Reflected:   %d\n", stats.BytesReflected)
				return
			case <-ticker.C:
				stats := dp.GetStats()
				fmt.Printf("\r[Stats] RX: %d pkts | TX: %d pkts | Signatures: ITO=%d RFC2544=%d Y.1564=%d MSN=%d",
					stats.PacketsReceived, stats.PacketsReflected,
					stats.SigProbeOT+stats.SigDataOT+stats.SigLatency,
					stats.SigRFC2544, stats.SigY1564, stats.SigMSN)
			}
		}
	}
}

func testCmd(args []string) {
	fs := flag.NewFlagSet("test", flag.ExitOnError)

	// Basic options
	iface := fs.String("interface", "", "Network interface")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")
	testTypes := fs.String("type", "throughput", "Test type(s), comma-separated")
	fs.StringVar(testTypes, "t", "throughput", "Test type (shorthand)")
	duration := fs.Int("duration", 60, "Test duration in seconds")
	fs.IntVar(duration, "d", 60, "Test duration (shorthand)")
	frameSizes := fs.String("frame-sizes", "64,128,256,512,1024,1280,1518", "Frame sizes")

	// Advanced options
	resolution := fs.Float64("resolution", 0.1, "Binary search resolution %")
	maxLoss := fs.Float64("max-loss", 0.0, "Maximum acceptable loss %")
	warmup := fs.Int("warmup", 2, "Warmup period in seconds")
	_ = fs.Int("trials", 3, "Number of trials") // Used in config

	// Y.1564 options
	cir := fs.Float64("cir", 0, "Committed Information Rate (Mbps)")
	eir := fs.Float64("eir", 0, "Excess Information Rate (Mbps)")
	fdThreshold := fs.Float64("fd-threshold", 10, "Frame Delay threshold (ms)")
	fdvThreshold := fs.Float64("fdv-threshold", 5, "Frame Delay Variation threshold (ms)")
	flrThreshold := fs.Float64("flr-threshold", 0.01, "Frame Loss Rate threshold (%)")

	// Output format
	jsonOutput := fs.Bool("json", false, "Output results in JSON")
	csvOutput := fs.Bool("csv", false, "Output results in CSV")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *iface == "" {
		fmt.Println("Error: --interface is required")
		fs.Usage()
		os.Exit(1)
	}

	// Validate test types
	tests := strings.Split(*testTypes, ",")
	for i, t := range tests {
		tests[i] = strings.TrimSpace(t)
		if _, ok := allTestTypes[tests[i]]; !ok {
			fmt.Printf("Error: Unknown test type '%s'\n", tests[i])
			fmt.Println("Run 'stem list-tests' to see available tests")
			os.Exit(1)
		}
	}

	// Check license (Tier 2 required for tests)
	mgr, err := license.NewManager()
	if err != nil {
		fmt.Printf("Warning: License check failed: %v\n", err)
	} else {
		state := mgr.GetState()
		if state == nil {
			fmt.Println("No active license. Starting 14-day trial...")
			result := mgr.StartTrial()
			if !result.Success {
				fmt.Printf("Error: %s\n", result.Message)
				os.Exit(1)
			}
			fmt.Printf("%s\n\n", result.Message)
		} else if !mgr.IsActivated() {
			fmt.Println("Error: License expired. Please activate a valid license.")
			fmt.Println("Run 'stem license --status' for details")
			os.Exit(1)
		} else if state.Tier < license.TierTestSuite && !state.IsTrialMode {
			fmt.Println("Error: Test Suite requires Tier 2 license")
			fmt.Println("Your license: Tier 1 (Reflector only)")
			os.Exit(1)
		}
	}

	// Parse frame sizes
	frameSizeList := parseFrameSizes(*frameSizes)
	if len(frameSizeList) == 0 {
		fmt.Println("Error: No valid frame sizes specified")
		os.Exit(1)
	}

	// Print test configuration
	fmt.Printf("%s %s - Network Testing\n", ProductName, version.Version)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Interface:    %s\n", *iface)
	fmt.Printf("Tests:        %s\n", *testTypes)
	fmt.Printf("Duration:     %d seconds\n", *duration)
	fmt.Printf("Frame sizes:  %s\n", *frameSizes)
	fmt.Printf("Resolution:   %.2f%%\n", *resolution)
	fmt.Printf("Max loss:     %.2f%%\n", *maxLoss)
	fmt.Printf("Warmup:       %d seconds\n", *warmup)
	fmt.Println(strings.Repeat("=", 60))

	// Create dataplane context
	ctx, err := testmasterDP.NewContext(*iface)
	if err != nil {
		fmt.Printf("Error: Failed to initialize dataplane: %v\n", err)
		os.Exit(1)
	}
	defer ctx.Close()

	// Configure the context
	cfg := &testmasterDP.Config{
		Interface:      *iface,
		AutoDetect:     true,
		TrialDuration:  time.Duration(*duration) * time.Second,
		WarmupPeriod:   time.Duration(*warmup) * time.Second,
		ResolutionPct:  *resolution,
		AcceptableLoss: *maxLoss,
		MeasureLatency: true,
	}

	if err := ctx.Configure(cfg); err != nil {
		fmt.Printf("Error: Failed to configure: %v\n", err)
		os.Exit(1)
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nCancelling test...")
		ctx.Cancel()
	}()

	// Run all requested tests
	var allResults []interface{}

	for _, testType := range tests {
		for _, frameSize := range frameSizeList {
			ctx.SetFrameSize(uint32(frameSize))

			fmt.Printf("\n[Running %s test with frame size %d bytes]\n", testType, frameSize)

			result, err := runTest(ctx, testType, *cir, *eir, *fdThreshold, *fdvThreshold, *flrThreshold, *duration)
			if err != nil {
				fmt.Printf("Error: %s test failed: %v\n", testType, err)
				continue
			}

			allResults = append(allResults, result)
			printTestResult(testType, result, *jsonOutput)
		}
	}

	// Final output
	if *jsonOutput && len(allResults) > 0 {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Printf("\n%s\n", string(data))
	} else if *csvOutput && len(allResults) > 0 {
		printCSVResults(allResults)
	}

	fmt.Println("\nTest suite complete.")
}

// parseFrameSizes parses comma-separated frame sizes
func parseFrameSizes(s string) []int {
	var sizes []int
	for _, part := range strings.Split(s, ",") {
		size, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && size >= 64 && size <= 9216 {
			sizes = append(sizes, size)
		}
	}
	return sizes
}

// runTest executes a single test and returns the result
func runTest(ctx *testmasterDP.Context, testType string, cir, eir, fdThreshold, fdvThreshold, flrThreshold float64, duration int) (interface{}, error) {
	switch testType {
	case "throughput":
		return ctx.RunThroughputTest()

	case "latency":
		loadLevels := []float64{10, 25, 50, 75, 90, 100}
		return ctx.RunLatencyTest(loadLevels)

	case "frame_loss":
		return ctx.RunFrameLossTest(10, 100, 10)

	case "back_to_back":
		return ctx.RunBackToBackTest(10000, 3)

	case "system_recovery":
		return ctx.RunSystemRecoveryTest(100.0, 60)

	case "reset":
		return ctx.RunResetTest()

	case "y1564_config", "y1564":
		service := &testmasterDP.Y1564Service{
			ServiceID:   1,
			ServiceName: "Service-1",
			SLA: testmasterDP.Y1564SLA{
				CIRMbps:         cir,
				EIRMbps:         eir,
				FDThresholdMs:   fdThreshold,
				FDVThresholdMs:  fdvThreshold,
				FLRThresholdPct: flrThreshold,
			},
			FrameSize: 1518,
			Enabled:   true,
		}
		return ctx.RunY1564ConfigTest(service)

	case "y1564_perf":
		service := &testmasterDP.Y1564Service{
			ServiceID:   1,
			ServiceName: "Service-1",
			SLA: testmasterDP.Y1564SLA{
				CIRMbps:         cir,
				EIRMbps:         eir,
				FDThresholdMs:   fdThreshold,
				FDVThresholdMs:  fdvThreshold,
				FLRThresholdPct: flrThreshold,
			},
			FrameSize: 1518,
			Enabled:   true,
		}
		return ctx.RunY1564PerfTest(service, uint32(duration))

	default:
		// For tests not yet implemented in dataplane, return a placeholder
		return map[string]string{
			"test":   testType,
			"status": "not_implemented",
			"note":   "This test type requires additional dataplane support",
		}, nil
	}
}

// printTestResult prints a test result
func printTestResult(testType string, result interface{}, jsonOutput bool) {
	if jsonOutput {
		return // Will be printed in batch at the end
	}

	switch r := result.(type) {
	case *testmasterDP.ThroughputResultCLI:
		fmt.Printf("  Max Rate:    %.2f%% (%.2f Mbps, %.0f pps)\n", r.MaxRatePct, r.MaxRateMbps, r.MaxRatePPS)
		fmt.Printf("  Iterations:  %d\n", r.Iterations)
		fmt.Printf("  Latency:     min=%.2fus avg=%.2fus max=%.2fus\n",
			r.Latency.MinNs/1000, r.Latency.AvgNs/1000, r.Latency.MaxNs/1000)

	case []testmasterDP.LatencyResultCLI:
		for _, lr := range r {
			fmt.Printf("  Load %.0f%%: min=%.2fus avg=%.2fus max=%.2fus p99=%.2fus\n",
				lr.LoadPct, lr.Latency.MinNs/1000, lr.Latency.AvgNs/1000,
				lr.Latency.MaxNs/1000, lr.Latency.P99Ns/1000)
		}

	case []testmasterDP.FrameLossResultCLI:
		for _, fl := range r {
			fmt.Printf("  Load %.0f%%: TX=%d RX=%d Loss=%.4f%%\n",
				fl.OfferedPct, fl.FramesTx, fl.FramesRx, fl.LossPct)
		}

	case *testmasterDP.BackToBackResultCLI:
		fmt.Printf("  Max Burst:   %d frames\n", r.MaxBurstFrames)
		fmt.Printf("  Duration:    %d us\n", r.BurstDurationUs)
		fmt.Printf("  Trials:      %d\n", r.Trials)

	case *testmasterDP.RecoveryResultCLI:
		fmt.Printf("  Recovery Time: %.2f ms\n", r.RecoveryTimeMs)
		fmt.Printf("  Frames Lost:   %d\n", r.FramesLost)

	case *testmasterDP.ResetResultCLI:
		fmt.Printf("  Reset Time:  %.2f ms\n", r.ResetTimeMs)
		fmt.Printf("  Frames Lost: %d\n", r.FramesLost)

	case *testmasterDP.Y1564ConfigResult:
		passStr := "PASS"
		if !r.ServicePass {
			passStr = "FAIL"
		}
		fmt.Printf("  Service %d: %s\n", r.ServiceID, passStr)
		for i, step := range r.Steps {
			stepPass := "PASS"
			if !step.StepPass {
				stepPass = "FAIL"
			}
			fmt.Printf("    Step %d: %.0f%% rate, FLR=%.4f%% FD=%.2fms FDV=%.2fms [%s]\n",
				i+1, step.OfferedRatePct, step.FLRPct, step.FDAvgMs, step.FDVMs, stepPass)
		}

	case *testmasterDP.Y1564PerfResult:
		passStr := "PASS"
		if !r.ServicePass {
			passStr = "FAIL"
		}
		fmt.Printf("  Service %d Performance: %s\n", r.ServiceID, passStr)
		fmt.Printf("    Duration:  %d sec\n", r.DurationSec)
		fmt.Printf("    Frames:    TX=%d RX=%d\n", r.FramesTx, r.FramesRx)
		fmt.Printf("    FLR:       %.4f%% [%s]\n", r.FLRPct, boolToPassFail(r.FLRPass))
		fmt.Printf("    FD:        %.2f ms [%s]\n", r.FDAvgMs, boolToPassFail(r.FDPass))
		fmt.Printf("    FDV:       %.2f ms [%s]\n", r.FDVMs, boolToPassFail(r.FDVPass))

	case map[string]string:
		fmt.Printf("  Status: %s\n", r["status"])
		if note, ok := r["note"]; ok {
			fmt.Printf("  Note: %s\n", note)
		}

	default:
		fmt.Printf("  Result: %+v\n", result)
	}
}

func boolToPassFail(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}

func printCSVResults(results []interface{}) {
	// Print CSV header
	fmt.Println("\ntest_type,frame_size,max_rate_pct,max_rate_mbps,loss_pct,latency_avg_us")

	for _, r := range results {
		switch result := r.(type) {
		case *testmasterDP.ThroughputResultCLI:
			fmt.Printf("throughput,%d,%.2f,%.2f,0,%.2f\n",
				result.FrameSize, result.MaxRatePct, result.MaxRateMbps, result.Latency.AvgNs/1000)
		}
	}
}

func webCmd(args []string) {
	fs := flag.NewFlagSet("web", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP port")
	fs.IntVar(port, "p", 8080, "HTTP port (shorthand)")
	host := fs.String("host", "0.0.0.0", "Bind address")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s - WebUI Server\n", ProductName, version.Version)
	fmt.Printf("Starting on http://%s:%d\n", *host, *port)

	server := web.NewServer(*port)
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func tuiCmd(args []string) {
	fs := flag.NewFlagSet("tui", flag.ExitOnError)
	mode := fs.String("mode", "test", "TUI mode: test or reflect")
	iface := fs.String("interface", "", "Network interface (required for reflect mode)")
	fs.StringVar(iface, "i", "", "Network interface (shorthand)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s - Terminal UI\n", ProductName, version.Version)

	switch *mode {
	case "reflect", "reflector":
		if *iface == "" {
			fmt.Println("Error: --interface is required for reflect mode")
			fmt.Println("Usage: stem tui --mode reflect -i eth0")
			os.Exit(1)
		}

		// Check license (Tier 1 minimum)
		mgr, _ := license.NewManager()
		if mgr != nil && !mgr.IsActivated() {
			result := mgr.StartTrial()
			if !result.Success {
				fmt.Printf("Error: %s\n", result.Message)
				os.Exit(1)
			}
		}

		// Build reflector config
		cfg := &reflectorConfig.Config{
			Interface:       *iface,
			SignatureFilter: "all",
			Reflection: reflectorConfig.ReflectConfig{
				Mode: "all",
			},
		}

		// Create and start reflector dataplane
		dp, err := reflectorDP.New(cfg)
		if err != nil {
			fmt.Printf("Error: Failed to create reflector: %v\n", err)
			os.Exit(1)
		}
		defer dp.Close()

		if err := dp.Start(); err != nil {
			fmt.Printf("Error: Failed to start reflector: %v\n", err)
			os.Exit(1)
		}

		// Launch reflector TUI
		tuiApp := reflectorTUI.New(dp)
		if err := tuiApp.Run(); err != nil {
			fmt.Printf("TUI error: %v\n", err)
			os.Exit(1)
		}

	case "test", "testmaster", "":
		// Check license (Tier 2 required)
		mgr, _ := license.NewManager()
		if mgr != nil {
			state := mgr.GetState()
			if state == nil {
				result := mgr.StartTrial()
				if !result.Success {
					fmt.Printf("Error: %s\n", result.Message)
					os.Exit(1)
				}
			} else if state.Tier < license.TierTestSuite && !state.IsTrialMode {
				fmt.Println("Error: Test Suite TUI requires Tier 2 license")
				os.Exit(1)
			}
		}

		// Launch testmaster TUI
		tuiApp := testmasterTUI.New()

		// Set up callbacks
		tuiApp.OnQuit = func() {
			tuiApp.Stop()
		}

		tuiApp.Log("The Stem TUI started")
		tuiApp.Log("Press F1 to start test, F2 to stop, F10 to quit")

		if err := tuiApp.Run(); err != nil {
			fmt.Printf("TUI error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Error: Unknown TUI mode '%s'\n", *mode)
		fmt.Println("Valid modes: test, reflect")
		os.Exit(1)
	}
}

func helpCmd(args []string) {
	fs := flag.NewFlagSet("help", flag.ExitOnError)
	simple := fs.Bool("simple", false, "Show simplified explanations for non-technical users")
	fs.BoolVar(simple, "s", false, "Show simplified explanations (shorthand)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If no topic specified, show general help
	if fs.NArg() == 0 {
		printUsage()
		return
	}

	topic := strings.ToLower(fs.Arg(0))

	// Try to find the topic in our help system
	if help.ShowHelp(topic, *simple) {
		return
	}

	// Not found in help system
	fmt.Printf("No help found for '%s'\n\n", topic)
	fmt.Println("Available help topics:")
	fmt.Println("  Commands:   reflect, test, web, license")
	fmt.Println("  Tests:      throughput, latency, frame_loss, y1564_config, ...")
	fmt.Println("  Categories: rfc2544, y1564, rfc2889, rfc6349, y1731, mef, tsn")
	fmt.Println("\nUse 'stem help tests' for a complete list of tests.")
	fmt.Println("Use 'stem glossary' for network terminology definitions.")
	fmt.Println("Use 'stem tutorial' for step-by-step guides.")
}

func tutorialCmd(args []string) {
	// If no tutorial specified, list available tutorials
	if len(args) == 0 {
		help.ShowTutorial("")
		return
	}

	tutorialID := strings.ToLower(strings.TrimSpace(args[0]))

	if !help.ShowTutorial(tutorialID) {
		fmt.Printf("Tutorial '%s' not found.\n\n", tutorialID)
		help.ShowTutorial("") // Show available tutorials
	}
}

func glossaryCmd(args []string) {
	fs := flag.NewFlagSet("glossary", flag.ExitOnError)
	simple := fs.Bool("simple", false, "Show only simple definitions")
	fs.BoolVar(simple, "s", false, "Show only simple definitions (shorthand)")
	search := fs.String("search", "", "Search for terms containing keyword")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Search mode
	if *search != "" {
		hs := help.NewHelpSystem()
		results := hs.SearchGlossary(*search)
		if len(results) == 0 {
			fmt.Printf("No terms found matching '%s'\n", *search)
			return
		}
		fmt.Printf("Terms matching '%s':\n\n", *search)
		for _, entry := range results {
			fmt.Printf("  %s - %s\n", entry.Term, entry.FullName)
		}
		fmt.Println("\nUse 'stem glossary <term>' for full definition.")
		return
	}

	// If no term specified, list all terms
	if fs.NArg() == 0 {
		help.ShowGlossary("", *simple)
		return
	}

	term := strings.ToLower(fs.Arg(0))

	if !help.ShowGlossary(term, *simple) {
		fmt.Printf("Term '%s' not found in glossary.\n\n", term)
		fmt.Println("Use 'stem glossary' to see all available terms.")
		fmt.Println("Use 'stem glossary --search <keyword>' to search.")
	}
}

func licenseCmd(args []string) {
	fs := flag.NewFlagSet("license", flag.ExitOnError)
	activate := fs.String("activate", "", "Activate with license key")
	trial := fs.Bool("trial", false, "Start 14-day trial")
	status := fs.Bool("status", false, "Show license status")
	deactivate := fs.Bool("deactivate", false, "Remove license")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	mgr, err := license.NewManager()
	if err != nil {
		fmt.Printf("Error: Failed to initialize license manager: %v\n", err)
		os.Exit(1)
	}

	switch {
	case *activate != "":
		result := mgr.Activate(*activate)
		if result.Success {
			fmt.Printf("Success: %s\n", result.Message)
			fmt.Printf("Tier: %s\n", result.Tier)
		} else {
			fmt.Printf("Error: %s\n", result.Message)
			os.Exit(1)
		}

	case *trial:
		result := mgr.StartTrial()
		if result.Success {
			fmt.Printf("Success: %s\n", result.Message)
			if result.DaysRemaining > 0 {
				fmt.Printf("Days remaining: %d\n", result.DaysRemaining)
			}
		} else {
			fmt.Printf("Error: %s\n", result.Message)
			os.Exit(1)
		}

	case *deactivate:
		if err := mgr.Deactivate(); err != nil {
			fmt.Printf("Error: Failed to deactivate: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("License deactivated successfully")

	case *status:
		fallthrough
	default:
		state := mgr.GetState()
		fp := mgr.GetFingerprint()

		fmt.Printf("%s - License Status\n", ProductName)
		fmt.Println(strings.Repeat("=", 50))

		if state == nil {
			fmt.Println("Status:    Not Activated")
			fmt.Println("\nTo start a 14-day trial:")
			fmt.Println("  stem license --trial")
			fmt.Println("\nTo activate with a license key:")
			fmt.Println("  stem license --activate XXXX-XXXX-XXXX-XXXX")
		} else if state.IsTrialMode {
			remaining := mgr.TrialDaysRemaining()
			fmt.Println("Status:    Trial Mode")
			fmt.Printf("Days Left: %d\n", remaining)
			fmt.Printf("Tier:      %s (full access during trial)\n", state.Tier)
			if remaining <= 3 {
				fmt.Println("\nWarning: Trial ending soon!")
				fmt.Println("Activate a license to continue using The Stem")
			}
		} else {
			fmt.Println("Status:    Licensed")
			fmt.Printf("Tier:      %s\n", state.Tier)
			fmt.Printf("Key:       %s\n", license.FormatKey(state.LicenseKey))
			fmt.Printf("Expires:   %s\n", state.ExpiresAt.Format("2006-01-02"))
		}

		fmt.Printf("\nDevice ID: %s\n", fp.Hash())
		fmt.Printf("Platform:  %s\n", fp.Platform)

		// Show features
		if state != nil && len(state.Features) > 0 {
			fmt.Printf("\nEnabled Features:\n")
			for _, f := range state.Features {
				fmt.Printf("  - %s\n", f)
			}
		}
	}
}
