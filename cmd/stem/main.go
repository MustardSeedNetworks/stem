// SPDX-License-Identifier: BUSL-1.1
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
//	stem web --port 8444                # WebUI (HTTPS by default)
//	stem tui                            # Terminal UI
package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/logging"
	"github.com/MustardSeedNetworks/stem/internal/version"
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

// Test-type wire identifiers. Match the executor's dispatch table; kept
// here as untyped string consts so the CLI dispatch switch is readable
// and grep-friendly.
const (
	testTypeThroughput     = "throughput"
	testTypeLatency        = "latency"
	testTypeFrameLoss      = "frame_loss"
	testTypeBackToBack     = "back_to_back"
	testTypeSystemRecovery = "system_recovery"
	testTypeReset          = "reset"
	testTypeY1564Config    = "y1564_config"
	testTypeY1564          = "y1564"
	testTypeY1564Perf      = "y1564_perf"
	testTypeTSN            = "tsn"
	testTypeMEF            = "mef"
)

// Default values for test configuration.
// Subcommand verbs. Hoisted out of the dispatch switch and per-subcommand
// [flag.NewFlagSet] calls so the canonical spelling lives in one place.
const (
	subReflect = "reflect"
	subTest    = "test"
)

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
	defaultWebPort          = 8444
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

	if !dispatchSubcommand(os.Args[1], os.Args[2:]) {
		_, _ = fmt.Fprintf(os.Stdout, "Unknown command: %s\n", os.Args[1])
		printUsage(os.Stdout)
		os.Exit(1)
	}
}

// dispatchSubcommand routes the CLI verb to its handler. Returns false
// if the command is unknown so main() can print usage and exit. Pulled
// out of main() to keep cyclomatic complexity below the project lint
// threshold as new verbs (install-ca, install-ca --uninstall) are added.
func dispatchSubcommand(cmd string, args []string) bool {
	switch cmd {
	case subReflect:
		if cmdErr := reflectCmd(args); cmdErr != nil {
			os.Exit(1)
		}
	case subTest:
		if cmdErr := testCmd(args); cmdErr != nil {
			os.Exit(1)
		}
	case "web":
		webCmd(args)
	case "tui":
		if cmdErr := tuiCmd(args); cmdErr != nil {
			os.Exit(1)
		}
	case "license":
		licenseCmd(args)
	case "list-tests":
		listTestsCmd(args)
	case "help", "--help", "-h":
		helpCmd(args)
	case "tutorial":
		tutorialCmd(args)
	case "glossary":
		glossaryCmd(args)
	case "version", "--version", "-v":
		printVersion(os.Stdout)
	case "install-ca":
		if cmdErr := installCACmd(args); cmdErr != nil {
			os.Exit(1)
		}
	default:
		return false
	}
	return true
}

func printVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "%s %s\n", ProductName, version.GetVersion())
	_, _ = fmt.Fprintf(w, "Commit: %s\n", version.GetCommit())
	_, _ = fmt.Fprintf(w, "Built:  %s\n", version.GetBuildTime())
	_, _ = fmt.Fprintf(w, "Copyright © %d %s\n", time.Now().Year(), Company)
	_, _ = fmt.Fprintln(w, "Network Performance Testing")
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintf(w, `%s %s
%s - Network Performance Testing`, ProductName, version.GetVersion(), Company)
	_, _ = fmt.Fprint(w, `

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
    list-tests   Show all available test types (grouped by module)
    install-ca   Install stem's self-signed root certificate into the OS trust store
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
    -p, --port         HTTPS port (default: 8444)
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

    # Start WebUI (HTTPS by default on :8444)
    stem web -p 8444

    # License management
    stem license --status
    stem license --trial
    stem license --activate XXXX-XXXX-XXXX-XXXX

For more information: https://mustardseednetworks.com
`)
}
