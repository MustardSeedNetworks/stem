/*
 * Seed Test Suite - CLI Command Help
 *
 * Documentation for all CLI commands and their options.
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

package help

// GetAllCommands returns help for all CLI commands.
func GetAllCommands() map[string]CommandHelp {
	return map[string]CommandHelp{
		"reflect":  ReflectCommand,
		"test":     TestCommand,
		"web":      WebCommand,
		"license":  LicenseCommand,
		"version":  VersionCommand,
		"help":     HelpCommand,
		"tutorial": TutorialCommand,
		"glossary": GlossaryCommand,
	}
}

// ReflectCommand documents the reflect subcommand.
//
//nolint:gochecknoglobals // Static help content.
var ReflectCommand = CommandHelp{
	Name:    "reflect",
	Summary: "Run packet reflection mode for remote testing",
	Description: `The reflect command starts the packet reflector, which receives test
packets and sends them back to their source. This is used as the far-end device
when running tests from another location.

The reflector supports multiple performance modes:
• AF_PACKET: Standard Linux socket mode, good for up to ~1-2 Mpps
• AF_XDP: High-performance eBPF mode, good for ~5-10 Mpps
• DPDK: Maximum performance mode, capable of line-rate on 10G+ links

The appropriate mode is selected automatically based on available system capabilities,
or can be specified manually.`,
	Usage: "stem reflect [flags]",
	Flags: []FlagHelp{
		{
			Short:      "-i",
			Long:       "--interface",
			Type:       "string",
			Default:    "",
			Required:   true,
			TechDesc:   "Network interface name for packet reflection",
			LaymanDesc: "Which network port to use (e.g., eth0, enp3s0)",
		},
		{
			Short:      "-p",
			Long:       "--port",
			Type:       "integer",
			Default:    "3842",
			Required:   false,
			TechDesc:   "UDP port for test traffic",
			LaymanDesc: "Port number for test packets (default is fine for most cases)",
		},
		{
			Short:      "",
			Long:       "--filter-oui",
			Type:       "string",
			Default:    "",
			Required:   false,
			TechDesc:   "Only reflect packets from MACs matching this OUI prefix",
			LaymanDesc: "Only respond to packets from specific device manufacturers",
		},
		{
			Short:      "",
			Long:       "--mode",
			Type:       "string",
			Default:    "auto",
			Required:   false,
			TechDesc:   "Dataplane mode: auto, af_packet, af_xdp, dpdk",
			LaymanDesc: "Performance mode - 'auto' picks the best available",
		},
		{
			Short:      "-w",
			Long:       "--web",
			Type:       "integer",
			Default:    "8080",
			Required:   false,
			TechDesc:   "Web UI port, 0 to disable",
			LaymanDesc: "Port for the web interface (set to 0 to disable)",
		},
		{
			Short:      "-v",
			Long:       "--verbose",
			Type:       "boolean",
			Default:    "false",
			Required:   false,
			TechDesc:   "Enable verbose logging",
			LaymanDesc: "Show detailed information about what's happening",
		},
	},
	Examples: []Example{
		{
			Desc:    "Start reflector on eth0",
			Command: "stem reflect -i eth0",
			Output:  "Reflector started on eth0:3842 (AF_XDP mode)",
		},
		{
			Desc:    "Start reflector with web UI on custom port",
			Command: "stem reflect -i eth0 -w 9000",
			Output:  "Reflector started, Web UI at http://localhost:9000",
		},
		{
			Desc:    "Start reflector in DPDK mode",
			Command: "stem reflect -i eth0 --mode dpdk",
			Output:  "Reflector started in DPDK mode",
		},
		{
			Desc:    "Start reflector without web UI",
			Command: "stem reflect -i eth0 -w 0",
			Output:  "Reflector started (web UI disabled)",
		},
	},
	SeeAlso: []string{"test", "web"},
}

// TestCommand documents the test subcommand.
//
//nolint:gochecknoglobals // Static help content.
var TestCommand = CommandHelp{
	Name:    "test",
	Summary: "Run network performance tests",
	Description: `The test command runs network performance tests against a reflector
or remote endpoint. It supports all test types from RFC 2544, Y.1564, RFC 2889,
RFC 6349, Y.1731, MEF, and TSN test suites.

Tests can be run individually by specifying the test type, or as part of a
complete test suite run.

Results are displayed in real-time and can be saved to files for later analysis.`,
	Usage: "stem test -i <interface> -t <test_type> [flags]",
	Flags: []FlagHelp{
		{
			Short:      "-i",
			Long:       "--interface",
			Type:       "string",
			Default:    "",
			Required:   true,
			TechDesc:   "Network interface for test traffic",
			LaymanDesc: "Which network port to use for testing",
		},
		{
			Short:      "-t",
			Long:       "--test",
			Type:       "string",
			Default:    "",
			Required:   true,
			TechDesc:   "Test type to run (use 'stem help tests' for list)",
			LaymanDesc: "Which test to run (e.g., throughput, latency, y1564_config)",
		},
		{
			Short:      "",
			Long:       "--target",
			Type:       "string",
			Default:    "",
			Required:   false,
			TechDesc:   "Remote reflector IP address (for remote testing)",
			LaymanDesc: "IP address of the remote reflector device",
		},
		{
			Short:      "-d",
			Long:       "--duration",
			Type:       "integer",
			Default:    "60",
			Required:   false,
			TechDesc:   "Test duration in seconds",
			LaymanDesc: "How long to run the test",
		},
		{
			Short:      "",
			Long:       "--frame-sizes",
			Type:       "string",
			Default:    "64,128,256,512,1024,1280,1518",
			Required:   false,
			TechDesc:   "Comma-separated frame sizes in bytes",
			LaymanDesc: "Packet sizes to test (comma-separated)",
		},
		{
			Short:      "",
			Long:       "--cir",
			Type:       "integer",
			Default:    "1000",
			Required:   false,
			TechDesc:   "Committed Information Rate in Mbps (for Y.1564/MEF tests)",
			LaymanDesc: "Guaranteed bandwidth to test (in Mbps)",
		},
		{
			Short:      "",
			Long:       "--output",
			Type:       "string",
			Default:    "",
			Required:   false,
			TechDesc:   "Output file for results (JSON or CSV based on extension)",
			LaymanDesc: "File to save results to",
		},
		{
			Short:      "-v",
			Long:       "--verbose",
			Type:       "boolean",
			Default:    "false",
			Required:   false,
			TechDesc:   "Enable verbose output",
			LaymanDesc: "Show detailed progress information",
		},
	},
	Examples: []Example{
		{
			Desc:    "Run throughput test",
			Command: "stem test -i eth0 -t throughput",
			Output:  "Testing throughput at frame sizes 64, 128, 256, 512, 1024, 1280, 1518...",
		},
		{
			Desc:    "Run latency test with custom duration",
			Command: "stem test -i eth0 -t latency -d 120",
			Output:  "Running 2-minute latency test...",
		},
		{
			Desc:    "Run Y.1564 config test",
			Command: "stem test -i eth0 -t y1564_config --cir 100",
			Output:  "Testing service at 25%, 50%, 75%, 100% of 100 Mbps CIR...",
		},
		{
			Desc:    "Run test against remote reflector",
			Command: "stem test -i eth0 -t throughput --target 192.168.1.100",
			Output:  "Connecting to reflector at 192.168.1.100...",
		},
		{
			Desc:    "Save results to file",
			Command: "stem test -i eth0 -t throughput --output results.json",
			Output:  "Results saved to results.json",
		},
	},
	SeeAlso: []string{"reflect", "help"},
}

// WebCommand documents the web subcommand.
//
//nolint:gochecknoglobals // Static help content.
var WebCommand = CommandHelp{
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
			Type:       "integer",
			Default:    "8080",
			Required:   false,
			TechDesc:   "HTTP port for web interface",
			LaymanDesc: "Port number for the web interface",
		},
		{
			Short:      "",
			Long:       "--host",
			Type:       "string",
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
			Output:  "Test Master UI available at http://localhost:8080",
		},
		{
			Desc:    "Start on custom port",
			Command: "stem web -p 9000",
			Output:  "Test Master UI available at http://localhost:9000",
		},
	},
	SeeAlso: []string{"reflect", "test"},
}

// LicenseCommand documents the license subcommand.
//
//nolint:gochecknoglobals // Static help content.
var LicenseCommand = CommandHelp{
	Name:    "license",
	Summary: "Manage license activation",
	Description: `The license command handles license activation and status.
Seed Test Suite requires a valid license key for operation. The license
determines which features are available:

• Reflector tier: Packet reflection only
• TestSuite tier: Full test suite (RFC 2544, Y.1564, etc.)
• Enterprise tier: All features plus API and multi-user support`,
	Usage: "stem license [subcommand] [flags]",
	Flags: []FlagHelp{
		{
			Short:      "-k",
			Long:       "--key",
			Type:       "string",
			Default:    "",
			Required:   false,
			TechDesc:   "License key to activate (format: XXXX-XXXX-XXXX-XXXX)",
			LaymanDesc: "Your license key from Mustard Seed Networks",
		},
		{
			Short:      "",
			Long:       "--status",
			Type:       "boolean",
			Default:    "false",
			Required:   false,
			TechDesc:   "Show current license status",
			LaymanDesc: "Check what license is currently active",
		},
		{
			Short:      "",
			Long:       "--deactivate",
			Type:       "boolean",
			Default:    "false",
			Required:   false,
			TechDesc:   "Deactivate current license",
			LaymanDesc: "Remove the current license",
		},
	},
	Examples: []Example{
		{
			Desc:    "Activate a license",
			Command: "stem license -k ABCD-1234-EFGH-5678",
			Output:  "License activated: TestSuite tier",
		},
		{
			Desc:    "Check license status",
			Command: "stem license --status",
			Output:  "License: TestSuite tier\nFeatures: reflector, rfc2544, y1564, ...",
		},
	},
	SeeAlso: []string{"version"},
}

// VersionCommand documents the version subcommand.
//
//nolint:gochecknoglobals // Static help content.
var VersionCommand = CommandHelp{
	Name:    "version",
	Summary: "Display version information",
	Description: `Shows the current version of Seed Test Suite along with
build information, license status, and available features.`,
	Usage: "stem version",
	Flags: []FlagHelp{},
	Examples: []Example{
		{
			Desc:    "Show version",
			Command: "stem version",
			Output:  "Seed Test Suite v1.0.0\nBuild: 2025-01-15\nLicense: TestSuite tier",
		},
	},
	SeeAlso: []string{"license"},
}

// HelpCommand documents the help subcommand.
//
//nolint:gochecknoglobals // Static help content.
var HelpCommand = CommandHelp{
	Name:    "help",
	Summary: "Get help on commands, tests, and concepts",
	Description: `The help command provides detailed information about commands,
tests, and network testing concepts. You can get help on:

• Commands: stem help reflect
• Tests: stem help throughput
• Categories: stem help rfc2544
• Concepts: Use the glossary command for definitions`,
	Usage: "stem help [topic]",
	Flags: []FlagHelp{},
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

// TutorialCommand documents the tutorial subcommand.
//
//nolint:gochecknoglobals // Static help content.
var TutorialCommand = CommandHelp{
	Name:    "tutorial",
	Summary: "Interactive tutorials for learning Seed Test Suite",
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

// GlossaryCommand documents the glossary subcommand.
//
//nolint:gochecknoglobals // Static help content.
var GlossaryCommand = CommandHelp{
	Name:    "glossary",
	Summary: "Network terminology definitions",
	Description: `The glossary command provides definitions for network testing
terminology. Each term includes both a technical definition for engineers
and a plain-English explanation for newcomers.

Run without arguments to see categories, or specify a term to get its
definition.`,
	Usage: "stem glossary [term]",
	Flags: []FlagHelp{},
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
