/*
 * The Stem - English Messages
 *
 * All user-facing strings in English.
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

package i18n

//nolint:gochecknoglobals // Static message catalog.
var englishMessages = map[string]string{
	// Application
	"app.name":        "The Stem",
	"app.description": "Network Performance Testing Tool",
	"app.copyright":   "Copyright (c) 2025 Mustard Seed Networks. All rights reserved.",

	// Commands
	"cmd.reflect.name":    "reflect",
	"cmd.reflect.summary": "Run packet reflection mode for remote testing",
	"cmd.reflect.desc":    "Starts the packet reflector to bounce test packets back to their source.",

	"cmd.test.name":    "test",
	"cmd.test.summary": "Run network performance tests",
	"cmd.test.desc":    "Execute RFC 2544, Y.1564, Y.1731, and other network tests.",

	"cmd.web.name":    "web",
	"cmd.web.summary": "Start the Test Master web interface",
	"cmd.web.desc":    "Launches the graphical web interface for test configuration and monitoring.",

	"cmd.license.name":    "license",
	"cmd.license.summary": "Manage license activation",
	"cmd.license.desc":    "Activate, deactivate, or check license status.",

	"cmd.version.name":    "version",
	"cmd.version.summary": "Display version information",
	"cmd.version.desc":    "Shows version, build info, and license status.",

	"cmd.help.name":    "help",
	"cmd.help.summary": "Get help on commands, tests, and concepts",
	"cmd.help.desc":    "Displays detailed documentation for any topic.",

	"cmd.tutorial.name":    "tutorial",
	"cmd.tutorial.summary": "Interactive tutorials for learning",
	"cmd.tutorial.desc":    "Step-by-step guides for common tasks.",

	"cmd.glossary.name":    "glossary",
	"cmd.glossary.summary": "Network terminology definitions",
	"cmd.glossary.desc":    "Look up network testing terms and concepts.",

	// Flags
	"flag.interface":   "Network interface",
	"flag.interface.d": "Network interface for packet transmission (e.g., eth0)",
	"flag.port":        "Port number",
	"flag.port.d":      "TCP/UDP port number",
	"flag.verbose":     "Verbose output",
	"flag.verbose.d":   "Enable detailed logging",
	"flag.duration":    "Test duration",
	"flag.duration.d":  "How long to run the test (in seconds)",
	"flag.output":      "Output file",
	"flag.output.d":    "File to save results to",
	"flag.config":      "Configuration file",
	"flag.config.d":    "Path to configuration file",

	// Test Categories
	"cat.rfc2544":      "RFC 2544",
	"cat.rfc2544.name": "Benchmarking Methodology for Network Interconnect Devices",
	"cat.rfc2544.desc": "Standard tests for measuring network device performance.",

	"cat.y1564":      "Y.1564",
	"cat.y1564.name": "Ethernet Service Activation Test",
	"cat.y1564.desc": "Service turn-up testing for carrier ethernet.",

	"cat.y1731":      "Y.1731",
	"cat.y1731.name": "Ethernet OAM",
	"cat.y1731.desc": "Operations, Administration, and Maintenance for ethernet services.",

	"cat.rfc2889":      "RFC 2889",
	"cat.rfc2889.name": "Benchmarking Methodology for LAN Switching Devices",
	"cat.rfc2889.desc": "Tests specific to network switches.",

	"cat.rfc6349":      "RFC 6349",
	"cat.rfc6349.name": "Framework for TCP Throughput Testing",
	"cat.rfc6349.desc": "TCP performance testing considering protocol behavior.",

	"cat.mef":      "MEF",
	"cat.mef.name": "Metro Ethernet Forum",
	"cat.mef.desc": "Carrier ethernet service certification tests.",

	"cat.tsn":      "TSN",
	"cat.tsn.name": "Time-Sensitive Networking",
	"cat.tsn.desc": "Deterministic networking for industrial applications.",

	// Test Names
	"test.throughput":        "Throughput Test",
	"test.throughput.desc":   "Finds the maximum speed without packet loss",
	"test.latency":           "Latency Test",
	"test.latency.desc":      "Measures packet delay through the network",
	"test.frame_loss":        "Frame Loss Test",
	"test.frame_loss.desc":   "Measures packet loss at various rates",
	"test.back_to_back":      "Back-to-Back Test",
	"test.back_to_back.desc": "Measures burst handling capacity",

	"test.y1564_config":      "Y.1564 Configuration Test",
	"test.y1564_config.desc": "Validates service at CIR percentages",
	"test.y1564_perf":        "Y.1564 Performance Test",
	"test.y1564_perf.desc":   "Extended performance verification",

	"test.frame_delay":         "Frame Delay Test",
	"test.frame_delay.desc":    "OAM-based delay measurement",
	"test.synthetic_loss":      "Synthetic Loss Measurement",
	"test.synthetic_loss.desc": "OAM-based loss measurement",

	// Status Messages
	"status.starting":   "Starting...",
	"status.running":    "Running",
	"status.completed":  "Completed",
	"status.failed":     "Failed",
	"status.cancelled":  "Cancelled",
	"status.waiting":    "Waiting",
	"status.connecting": "Connecting...",

	// Results
	"result.pass":    "PASS",
	"result.fail":    "FAIL",
	"result.warning": "WARNING",
	"result.info":    "INFO",

	// Units
	"unit.bps":     "bps",
	"unit.kbps":    "Kbps",
	"unit.mbps":    "Mbps",
	"unit.gbps":    "Gbps",
	"unit.pps":     "pps",
	"unit.kpps":    "Kpps",
	"unit.mpps":    "Mpps",
	"unit.ms":      "ms",
	"unit.us":      "us",
	"unit.ns":      "ns",
	"unit.percent": "%",
	"unit.bytes":   "bytes",

	// Errors
	"err.interface_required":  "Network interface is required",
	"err.interface_not_found": "Interface not found: %s",
	"err.test_type_required":  "Test type is required",
	"err.test_type_invalid":   "Invalid test type: %s",
	"err.license_required":    "Valid license required for this feature",
	"err.license_expired":     "License has expired",
	"err.connection_failed":   "Failed to connect to reflector",
	"err.permission_denied":   "Permission denied (try running as root)",
	"err.port_in_use":         "Port %d is already in use",
	"err.config_not_found":    "Configuration file not found: %s",
	"err.config_invalid":      "Invalid configuration: %s",

	// Prompts
	"prompt.continue":         "Press Enter to continue...",
	"prompt.confirm":          "Are you sure? (y/n): ",
	"prompt.select_test":      "Select a test type:",
	"prompt.select_interface": "Select a network interface:",

	// Help
	"help.usage":     "Usage",
	"help.examples":  "Examples",
	"help.flags":     "Flags",
	"help.see_also":  "See Also",
	"help.technical": "Technical Description",
	"help.simple":    "Simple Explanation",
	"help.when_use":  "When to Use",
	"help.when_not":  "When Not to Use",
	"help.tips":      "Tips",
	"help.issues":    "Common Issues",
	"help.params":    "Parameters",
	"help.metrics":   "Output Metrics",

	// UI Labels
	"ui.dashboard":  "Dashboard",
	"ui.tests":      "Tests",
	"ui.results":    "Results",
	"ui.reflector":  "Reflector",
	"ui.settings":   "Settings",
	"ui.help":       "Help",
	"ui.start":      "Start",
	"ui.stop":       "Stop",
	"ui.cancel":     "Cancel",
	"ui.save":       "Save",
	"ui.export":     "Export",
	"ui.refresh":    "Refresh",
	"ui.filter":     "Filter",
	"ui.search":     "Search",
	"ui.loading":    "Loading...",
	"ui.no_results": "No results found",
	"ui.error":      "Error",
	"ui.success":    "Success",
	"ui.warning":    "Warning",

	// Time
	"time.now":       "Now",
	"time.today":     "Today",
	"time.yesterday": "Yesterday",
	"time.last_week": "Last Week",
	"time.ago":       "%s ago",
	"time.seconds":   "seconds",
	"time.minutes":   "minutes",
	"time.hours":     "hours",
	"time.days":      "days",

	// Reflector
	"reflector.mode":    "Reflector Mode",
	"reflector.started": "Reflector started on %s",
	"reflector.stopped": "Reflector stopped",
	"reflector.stats":   "Reflector Statistics",
	"reflector.packets": "Packets Reflected",
	"reflector.bytes":   "Bytes Reflected",
	"reflector.rate":    "Current Rate",
	"reflector.uptime":  "Uptime",

	// License
	"license.status":      "License Status",
	"license.tier":        "License Tier",
	"license.valid_until": "Valid Until",
	"license.features":    "Available Features",
	"license.activate":    "Activate License",
	"license.deactivate":  "Deactivate License",
	"license.trial":       "Trial Mode",
	"license.expired":     "Expired",

	// Modules
	"module.reflector":   "Reflector",
	"module.benchmark":   "Benchmark",
	"module.servicetest": "Service Test",
	"module.trafficgen":  "Traffic Generator",
	"module.measure":     "Measure",
	"module.certify":     "Certify",

	// Test Parameters - RFC 2544
	"param.frameSizes":      "Frame Sizes",
	"param.frameSizes.d":    "Ethernet frame sizes in bytes to test",
	"param.duration":        "Duration",
	"param.duration.d":      "Test duration in seconds",
	"param.resolution":      "Resolution",
	"param.resolution.d":    "Binary search resolution percentage",
	"param.maxLoss":         "Max Loss",
	"param.maxLoss.d":       "Maximum acceptable frame loss percentage",
	"param.warmup":          "Warmup",
	"param.warmup.d":        "Warmup period in seconds before measurements",
	"param.trials":          "Trials",
	"param.trials.d":        "Number of trial iterations per test point",
	"param.stepSize":        "Step Size",
	"param.stepSize.d":      "Rate step size for frame loss testing",
	"param.bidirectional":   "Bidirectional",
	"param.bidirectional.d": "Run tests in both directions simultaneously",

	// Test Parameters - Y.1564
	"param.cir":            "CIR",
	"param.cir.d":          "Committed Information Rate in Mbps",
	"param.eir":            "EIR",
	"param.eir.d":          "Excess Information Rate in Mbps",
	"param.cbs":            "CBS",
	"param.cbs.d":          "Committed Burst Size in KB",
	"param.ebs":            "EBS",
	"param.ebs.d":          "Excess Burst Size in KB",
	"param.vlanId":         "VLAN ID",
	"param.vlanId.d":       "VLAN identifier for tagged traffic (0-4095)",
	"param.pcp":            "PCP",
	"param.pcp.d":          "Priority Code Point for 802.1p CoS (0-7)",
	"param.colorAware":     "Color Aware",
	"param.colorAware.d":   "Enable color-aware traffic conditioning",
	"param.flrThreshold":   "FLR Threshold",
	"param.flrThreshold.d": "Frame Loss Ratio acceptance threshold",
	"param.fdThreshold":    "FD Threshold",
	"param.fdThreshold.d":  "Frame Delay acceptance threshold in ms",
	"param.fdvThreshold":   "FDV Threshold",
	"param.fdvThreshold.d": "Frame Delay Variation threshold in ms",

	// Test Parameters - TSN
	"param.maxLatencyNs":     "Max Latency",
	"param.maxLatencyNs.d":   "Maximum acceptable latency in nanoseconds",
	"param.maxJitterNs":      "Max Jitter",
	"param.maxJitterNs.d":    "Maximum acceptable jitter in nanoseconds",
	"param.requirePTPSync":   "Require PTP Sync",
	"param.requirePTPSync.d": "Require PTP synchronization before testing",
	"param.baseTimeNs":       "Base Time",
	"param.baseTimeNs.d":     "Base time for gate control list in nanoseconds",
	"param.cycleTimeNs":      "Cycle Time",
	"param.cycleTimeNs.d":    "Gate cycle time in nanoseconds",
	"param.trafficClass":     "Traffic Class",
	"param.trafficClass.d":   "IEEE 802.1Q traffic class for scheduled traffic",

	// Test Parameters - TrafficGen
	"param.ratePct":           "Rate Percent",
	"param.ratePct.d":         "Traffic rate as percentage of line rate",
	"param.streamId":          "Stream ID",
	"param.streamId.d":        "Unique identifier for traffic stream",
	"param.burstMode":         "Burst Mode",
	"param.burstMode.d":       "Enable burst traffic mode",
	"param.burstSize":         "Burst Size",
	"param.burstSize.d":       "Number of frames per burst",
	"param.interBurstGapUs":   "Inter-Burst Gap",
	"param.interBurstGapUs.d": "Gap between bursts in microseconds",
	"param.srcMac":            "Source MAC",
	"param.srcMac.d":          "Source MAC address for generated frames",
	"param.dstMac":            "Destination MAC",
	"param.dstMac.d":          "Destination MAC address for generated frames",

	// TrafficGen Category
	"cat.trafficgen":      "TrafficGen",
	"cat.trafficgen.name": "Custom Traffic Generation",
	"cat.trafficgen.desc": "Generate custom traffic patterns for specialized testing",

	// Custom Stream Test
	"test.custom_stream":      "Custom Traffic Stream",
	"test.custom_stream.desc": "Generate custom traffic patterns for specialized testing",
}
