// Package help provides the static, hand-maintained content behind `stem
// help` and `stem tutorial`: per-command flag/example reference (this file)
// and step-by-step tutorials (tutorials.go).
package help

// GetAllCommands returns help for all CLI commands.
func GetAllCommands() map[string]CommandHelp {
	return map[string]CommandHelp{
		"reflect":    ReflectCommand(),
		"test":       TestCommand(),
		"web":        WebCommand(),
		"license":    LicenseCommand(),
		"version":    VersionCommand(),
		"help":       helpCommand(),
		"tutorial":   TutorialCommand(),
		"glossary":   GlossaryCommand(),
		"list-tests": ListTestsCommand(),
		"install-ca": InstallCACommand(),
	}
}

// ReflectCommand documents the reflect subcommand.
func ReflectCommand() CommandHelp {
	return CommandHelp{
		Name:    "reflect",
		Summary: "Run packet reflection mode for remote testing",
		Description: `The reflect command starts a standalone packet reflector, which receives test
packets and sends them back to their source. This is used as the far-end device
when running tests from another location.

	Profiles select the packet signature and reflection behavior expected by the
	remote tester. Use netally (or its ito alias) for EtherScope and CyberScope.`,
		Usage: "stem reflect [flags]",
		Flags: []FlagHelp{
			{
				Short:      "-i",
				Long:       "--interface",
				Type:       TypeString,
				Default:    "",
				Required:   true,
				TechDesc:   "Network interface name for packet reflection",
				LaymanDesc: "Which network port to use (e.g., eth0, enp3s0)",
			},
			{
				Short:      "",
				Long:       "--port",
				Type:       TypeInteger,
				Default:    "0",
				Required:   false,
				TechDesc:   "Override the profile's UDP port (0 uses the profile default)",
				LaymanDesc: "Override the UDP port selected by the profile",
			},
			{
				Short:      "",
				Long:       "--profile",
				Type:       TypeString,
				Default:    "all",
				Required:   false,
				TechDesc:   "Reflection profile: all, netally, ito, msn, or custom",
				LaymanDesc: "Choose the packet format used by the remote tester",
			},
			{
				Short:      "",
				Long:       "--oui",
				Type:       TypeString,
				Default:    "",
				Required:   false,
				TechDesc:   "Only reflect packets from MAC addresses matching this OUI prefix",
				LaymanDesc: "Only respond to packets from a specific device manufacturer",
			},
			{
				Short:      "",
				Long:       "--at-boot",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				Required:   false,
				TechDesc:   "Ask the daemon to start this reflector again whenever it starts",
				LaymanDesc: "Bring the reflector back automatically after a reboot",
			},
		},
		Examples: []Example{
			{
				Desc:    "Start reflector on eth0",
				Command: "stem reflect -i eth0",
				Output:  "Interface: eth0\nProfile: all\nMode: all",
			},
			{Desc: "Reflect NetAlly test traffic", Command: "stem reflect -i eth0 --profile netally"},
			{Desc: "Keep reflecting across reboots", Command: "stem reflect -i eth0 --profile netally --at-boot"},
			{Desc: "Watch live counters in the web UI", Command: "stem web -p 8444"},
		},
		SeeAlso: []string{"test", "web"},
	}
}

// TestCommand documents the test subcommand.
func TestCommand() CommandHelp {
	return CommandHelp{
		Name:    "test",
		Summary: "Run network performance tests",
		Description: `The test command runs network performance tests against a reflector
	or remote endpoint. It supports all test types from RFC 2544, Y.1564, RFC 2889,
	RFC 6349, Y.1731, MEF, and TSN test suites.

	Tests can be run individually by specifying the test type, or as part of a
	complete test suite run.

	Results are displayed in real-time and can be saved to files for later analysis.`,
		Usage:    "stem test -i <interface> -t <test_type> [flags]",
		Flags:    testCommandFlags(),
		Examples: testCommandExamples(),
		SeeAlso:  []string{"reflect", "web", "help", "tutorial"},
	}
}

func testCommandFlags() []FlagHelp {
	return append(testCommandBasicFlags(), testCommandAdvancedFlags()...)
}

func testCommandBasicFlags() []FlagHelp {
	return []FlagHelp{
		{
			Short:      "-i",
			Long:       "--interface",
			Type:       TypeString,
			Default:    "",
			Required:   true,
			TechDesc:   "Network interface for test traffic",
			LaymanDesc: "Which network port to use for testing",
		},
		{
			Short:      "-t",
			Long:       "--type",
			Type:       TypeString,
			Default:    "rfc2544_throughput",
			Required:   false,
			TechDesc:   "Test type to run (use 'stem list-tests' for the catalog)",
			LaymanDesc: "Which test to run (for example, rfc2544_throughput or y1564_config)",
		},
		{
			Short:      "",
			Long:       "--peer",
			Type:       TypeString,
			Default:    "",
			Required:   false,
			TechDesc:   "Reflector host name or IPv4 address",
			LaymanDesc: "Remote reflector to test through",
		},
		{
			Long:       "--peer-port",
			Type:       TypeInteger,
			Default:    "3842",
			TechDesc:   "Reflector UDP port",
			LaymanDesc: "UDP port used by the reflector",
		},
		{
			Short:      "",
			Long:       FlagFrameSizes,
			Type:       TypeString,
			Default:    DefaultFrameSizes,
			Required:   false,
			TechDesc:   "Frame sizes to test (comma-separated)",
			LaymanDesc: "Packet sizes to use for testing",
		},
		{
			Short:      "-d",
			Long:       FlagDuration,
			Type:       TypeInteger,
			Default:    "60",
			Required:   false,
			TechDesc:   "Test duration per step (seconds)",
			LaymanDesc: "How long to test at each speed",
		},
	}
}

func testCommandAdvancedFlags() []FlagHelp {
	return []FlagHelp{
		{
			Long:       "--resolution",
			Type:       "float",
			Default:    "0.1",
			TechDesc:   "Binary-search resolution percentage",
			LaymanDesc: "How precisely throughput is narrowed",
		},
		{
			Long:       "--max-loss",
			Type:       "float",
			Default:    "0",
			TechDesc:   "Maximum acceptable loss percentage",
			LaymanDesc: "Packet loss allowed before a step fails",
		},
		{
			Long:       "--warmup",
			Type:       TypeInteger,
			Default:    "2",
			TechDesc:   "Warmup period in seconds",
			LaymanDesc: "Time allowed before measurements begin",
		},
		{
			Long:       "--trials",
			Type:       TypeInteger,
			Default:    "3",
			TechDesc:   "Number of trials",
			LaymanDesc: "How many times to repeat each measurement",
		},
		{
			Long:       "--cir",
			Type:       "float",
			Default:    "0",
			TechDesc:   "Committed Information Rate in Mbps",
			LaymanDesc: "Guaranteed service rate for Y.1564",
		},
		{
			Long:       "--eir",
			Type:       "float",
			Default:    "0",
			TechDesc:   "Excess Information Rate in Mbps",
			LaymanDesc: "Additional service rate for Y.1564",
		},
		{
			Long:       "--fd-threshold",
			Type:       "float",
			Default:    "10",
			TechDesc:   "Frame Delay threshold in milliseconds",
			LaymanDesc: "Maximum allowed frame delay",
		},
		{
			Long:       "--fdv-threshold",
			Type:       "float",
			Default:    "5",
			TechDesc:   "Frame Delay Variation threshold in milliseconds",
			LaymanDesc: "Maximum allowed jitter",
		},
		{
			Long:       "--flr-threshold",
			Type:       "float",
			Default:    "0.01",
			TechDesc:   "Frame Loss Rate threshold percentage",
			LaymanDesc: "Maximum allowed packet loss",
		},
		{
			Long:       "--json",
			Type:       TypeBoolean,
			Default:    ValueFalse,
			TechDesc:   "Write results as JSON",
			LaymanDesc: "Show machine-readable JSON results",
		},
		{
			Long:       "--csv",
			Type:       TypeBoolean,
			Default:    ValueFalse,
			TechDesc:   "Write results as CSV",
			LaymanDesc: "Show spreadsheet-compatible results",
		},
	}
}

func testCommandExamples() []Example {
	return []Example{
		{
			Desc:    "Run throughput test",
			Command: "stem test -i eth0 -t rfc2544_throughput",
			Output:  "Test running... Results: Max Rate 98.5%",
		},
		{
			Desc:    "Run Y.1564 service test",
			Command: "stem test -i eth0 -t y1564_config --cir 100",
			Output:  "Step 1/4 PASS, Step 2/4 PASS, ...",
		},
		{
			Desc:    "Emit JSON results",
			Command: "stem test -i eth0 -t rfc2544_latency --json",
			Output:  "{ ... }",
		},
	}
}

// WebCommand documents the web subcommand.
func WebCommand() CommandHelp {
	return CommandHelp{
		Name:    "web",
		Summary: "Start the Test Master web interface",
		Description: `The web command starts the Test Master graphical web interface.
This provides a full-featured GUI for configuring and running tests, viewing
results, and monitoring reflector status.

The web interface includes:
• Test configuration with all parameters
• Real-time test progress and results
• Historical results browser
• Reflector status monitoring
• Help and documentation`,
		Usage: "stem web [flags]",
		Flags: []FlagHelp{
			{
				Short:      "-p",
				Long:       "--port",
				Type:       TypeInteger,
				Default:    "8444",
				Required:   false,
				TechDesc:   "HTTPS port for web interface",
				LaymanDesc: "Port number for the web interface (HTTPS by default)",
			},
			{
				Short:      "",
				Long:       "--host",
				Type:       TypeString,
				Default:    "0.0.0.0",
				Required:   false,
				TechDesc:   "IP address to bind to (0.0.0.0 for all interfaces)",
				LaymanDesc: "Which IP address to listen on",
			},
		},
		Examples: []Example{
			{
				Desc:    "Start web interface on default port",
				Command: "stem web",
				Output:  "Test Master UI available at https://localhost:8444",
			},
			{
				Desc:    "Start on custom port",
				Command: "stem web -p 9000",
				Output:  "Test Master UI available at https://localhost:9000",
			},
		},
		SeeAlso: []string{"reflect", "test"},
	}
}

// LicenseCommand documents the license subcommand.
func LicenseCommand() CommandHelp {
	return CommandHelp{
		Name:    "license",
		Summary: "Manage license activation",
		Description: `The license command handles license activation and status.
Stem Professional requires a valid license key for operation. The license
determines which features are available:

• Reflector tier: Packet reflection only (Free)
• Professional tier: Full test suite (RFC 2544 / Y.1564 / Y.1731 / RFC 2889 / RFC 6349 / MEF / TSN) plus API access`,
		Usage: "stem license [subcommand] [flags]",
		Flags: []FlagHelp{
			{
				Short:      "",
				Long:       "--activate",
				Type:       TypeString,
				Default:    "",
				Required:   false,
				TechDesc:   "Signed license token to activate (format: MSN1.<payload>.<signature>)",
				LaymanDesc: "Your license key from Mustard Seed Networks",
			},
			{
				Short:      "",
				Long:       "--trial",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				Required:   false,
				TechDesc:   "Start the 14-day Professional trial",
				LaymanDesc: "Try all Professional tests for 14 days",
			},
			{
				Short:      "",
				Long:       "--status",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				Required:   false,
				TechDesc:   "Show current license status",
				LaymanDesc: "Check what license is currently active",
			},
			{
				Short:      "",
				Long:       "--deactivate",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				Required:   false,
				TechDesc:   "Deactivate current license",
				LaymanDesc: "Remove the current license",
			},
		},
		Examples: []Example{
			{
				Desc:    "Activate a license",
				Command: "stem license --activate MSN1.<payload>.<signature>",
				Output:  "License activated: Professional tier",
			},
			{
				Desc:    "Check license status",
				Command: "stem license --status",
				Output:  "License: Professional tier\nFeatures: reflector, rfc2544, y1564, ...",
			},
		},
		SeeAlso: []string{"version"},
	}
}

// VersionCommand documents the version subcommand.
func VersionCommand() CommandHelp {
	return CommandHelp{
		Name:    "version",
		Summary: "Display version information",
		Description: `Shows the current version of Stem along with build
information, license status, and available features.`,
		Usage: "stem version",
		Flags: []FlagHelp{},
		Examples: []Example{
			{
				Desc:    "Show version",
				Command: "stem version",
				Output:  "Stem v0.23.0\nBuild: 2026-05-29\nLicense: Professional tier",
			},
		},
		SeeAlso: []string{"license"},
	}
}

// helpCommand documents the help subcommand.
func helpCommand() CommandHelp {
	return CommandHelp{
		Name:    "help",
		Summary: "Get help on commands, tests, and concepts",
		Description: `The help command provides detailed information about commands,
tests, and network testing concepts. You can get help on:

• Commands: stem help reflect
• Tests: stem help throughput
• Categories: stem help rfc2544
• Concepts: Use the glossary command for definitions`,
		Usage: "stem help [flags] [topic]",
		Flags: []FlagHelp{
			{
				Short:      "-s",
				Long:       "--simple",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				TechDesc:   "Show simplified explanations",
				LaymanDesc: "Use less technical wording",
			},
		},
		Examples: []Example{
			{
				Desc:    "Get help on a command",
				Command: "stem help reflect",
				Output:  "[Detailed reflect command documentation]",
			},
			{
				Desc:    "Get help on a test",
				Command: "stem help throughput",
				Output:  "[Detailed throughput test documentation]",
			},
			{
				Desc:    "Get help on a test category",
				Command: "stem help rfc2544",
				Output:  "[RFC 2544 category overview and test list]",
			},
			{
				Desc:    "List all available tests",
				Command: "stem help tests",
				Output:  "[List of all 27 tests by category]",
			},
		},
		SeeAlso: []string{"tutorial", "glossary"},
	}
}

// TutorialCommand documents the tutorial subcommand.
func TutorialCommand() CommandHelp {
	return CommandHelp{
		Name:    "tutorial",
		Summary: "Interactive tutorials for learning The Stem",
		Description: `The tutorial command provides step-by-step guides for common
tasks. Tutorials are designed for both beginners and experienced users who
want to learn specific features.

Run without arguments to list available tutorials, or specify a tutorial
name to start it.`,
		Usage: "stem tutorial [name]",
		Flags: []FlagHelp{},
		Examples: []Example{
			{
				Desc:    "List available tutorials",
				Command: "stem tutorial",
				Output:  "Available tutorials:\n  quickstart    - Your First Test in 5 Minutes\n  reflector     - Setting Up Packet Reflection\n  ...",
			},
			{
				Desc:    "Start quickstart tutorial",
				Command: "stem tutorial quickstart",
				Output:  "[Interactive tutorial begins]",
			},
		},
		SeeAlso: []string{"help", "glossary"},
	}
}

// GlossaryCommand documents the glossary subcommand.
func GlossaryCommand() CommandHelp {
	return CommandHelp{
		Name:    "glossary",
		Summary: "Network terminology definitions",
		Description: `The glossary command provides definitions for network testing
terminology. Each term includes both a technical definition for engineers
and a plain-English explanation for newcomers.

Run without arguments to see categories, or specify a term to get its
definition.`,
		Usage: "stem glossary [flags] [term]",
		Flags: []FlagHelp{
			{
				Short:      "-s",
				Long:       "--simple",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				TechDesc:   "Show only simple definitions",
				LaymanDesc: "Use less technical definitions",
			},
			{
				Long:       "--search",
				Type:       TypeString,
				Default:    "",
				TechDesc:   "Search for terms containing a keyword",
				LaymanDesc: "Find glossary entries by keyword",
			},
		},
		Examples: []Example{
			{
				Desc:    "List glossary categories",
				Command: "stem glossary",
				Output:  "Glossary Categories:\n  Bandwidth & Rate\n  Latency & Timing\n  ...",
			},
			{
				Desc:    "Look up a term",
				Command: "stem glossary cir",
				Output:  "CIR - Committed Information Rate\n\nTechnical: The guaranteed bandwidth...\n\nSimple: The speed your ISP promises...",
			},
			{
				Desc:    "Search for terms",
				Command: "stem glossary --search latency",
				Output:  "Terms matching 'latency':\n  latency, rtt, jitter, fdv, ...",
			},
		},
		SeeAlso: []string{"help", "tutorial"},
	}
}

// ListTestsCommand documents the list-tests subcommand.
func ListTestsCommand() CommandHelp {
	return CommandHelp{
		Name:    "list-tests",
		Summary: "List every available test type, grouped by module",
		Description: `The list-tests command prints every test type stem can run, grouped by
module (Benchmark / ServiceTest / TrafficGen / Measure / Certify), with each
test's standard and a one-line description.

The default output is human-readable and module-grouped. Use --json for a
machine-parseable representation suitable for piping into other tools.`,
		Usage: "stem list-tests [flags]",
		Flags: []FlagHelp{
			{
				Short:      "",
				Long:       "--json",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				Required:   false,
				TechDesc:   "Emit module + test metadata as JSON",
				LaymanDesc: "Machine-readable output for scripts",
			},
		},
		Examples: []Example{
			{
				Desc:    "List all tests grouped by module",
				Command: "stem list-tests",
				Output:  "stem - Available Test Types by Module\n=====================================\nBenchmark (RFC 2544): throughput, latency, frame_loss, ...",
			},
			{
				Desc:    "Machine-readable JSON output",
				Command: "stem list-tests --json",
				Output:  "{ \"modules\": [...], \"count\": 5 }",
			},
		},
		SeeAlso: []string{"test", "help"},
	}
}

// InstallCACommand documents the install-ca subcommand.
func InstallCACommand() CommandHelp {
	return CommandHelp{
		Name:    "install-ca",
		Summary: "Install stem's self-signed root certificate into the OS trust store",
		Description: `The install-ca command installs stem's self-signed root certificate into
the operating system's trust store so browsers stop showing the "not secure"
warning when visiting the stem WebUI over HTTPS.

stem generates its self-signed root on first launch (at certs/server.crt).
Run stem at least once before install-ca so the certificate file exists.

Supported platforms:
  macOS    System Keychain (requires sudo)
  Linux    System CA bundle via update-ca-certificates / update-ca-trust
  Windows  LocalMachine\Root (requires elevated shell)

Verification:
  stem install-ca --print-fingerprint
  curl -k https://localhost:8444/__version | jq -r .tlsFingerprint
The two values must match.`,
		Usage: "stem install-ca [flags]",
		Flags: []FlagHelp{
			{
				Short:      "",
				Long:       "--cert",
				Type:       TypeString,
				Default:    "certs/server.crt",
				Required:   false,
				TechDesc:   "Path to the PEM-encoded certificate to install",
				LaymanDesc: "Which certificate file to install (default is stem's)",
			},
			{
				Short:      "",
				Long:       "--uninstall",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				Required:   false,
				TechDesc:   "Remove stem's certificate from the OS trust store",
				LaymanDesc: "Undo a previous install-ca",
			},
			{
				Short:      "",
				Long:       "--print-fingerprint",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				Required:   false,
				TechDesc:   "Print the SHA-256 fingerprint of the cert and exit without modifying the trust store",
				LaymanDesc: "Just show the certificate's fingerprint for verification",
			},
		},
		Examples: []Example{
			{
				Desc:    "Install stem's root into the OS trust store",
				Command: "sudo stem install-ca",
				Output:  "[ok] Installed.\nCertificate SHA-256 fingerprint:\n  AA:BB:CC:...",
			},
			{
				Desc:    "Print fingerprint only (no trust-store change)",
				Command: "stem install-ca --print-fingerprint",
				Output:  "AA:BB:CC:DD:...",
			},
			{
				Desc:    "Remove the previously installed root",
				Command: "sudo stem install-ca --uninstall",
				Output:  "[ok] Removed.",
			},
		},
		SeeAlso: []string{"web"},
	}
}
