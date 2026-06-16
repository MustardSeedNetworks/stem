/*
 * The Stem - Test Documentation
 *
 * TSN test help content and category registry.
 */

package help

// ============================================================================
// TSN Tests - Time-Sensitive Networking Tests.
// ============================================================================

func tsnGateTiming() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       "gate_timing",
			Name:     "TSN Gate Timing Test",
			Standard: "IEEE 802.1Qbv",
			Category: StandardTSN,
		},
		tsnGateTimingDescriptions(),
		tsnGateTimingUsage(),
		tsnGateTimingDetails(),
	)
}

func tsnGateTimingDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Verifies Time-Aware Shaper gate timing accuracy.",
		TechDesc: `The TSN Gate Timing Test validates that Time-Aware Shaper (TAS) gates
per IEEE 802.1Qbv open and close at the correct times. This is critical for
deterministic latency in industrial and automotive networks where precise timing
enables guaranteed delivery windows.

The test sends time-synchronized traffic and measures if it arrives within the
expected gate windows.`,
		LaymanDesc: `Time-Sensitive Networking (TSN) is for networks where timing
is everything - like factory automation where a robot arm must move at EXACTLY
the right moment.

TSN uses "time gates" like traffic lights for network packets:
• Green light: Your packet can pass NOW
• Red light: Wait for your scheduled slot

This test verifies:
• Do the gates open at the right time?
• Is timing precise enough for your application?
• Measured in microseconds (millionths of a second)

If gates aren't perfectly timed, industrial equipment might not work correctly.`,
	}
}

func tsnGateTimingUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• Industrial automation networks
• Automotive ethernet validation
• Any application requiring deterministic timing
• IEEE 802.1Qbv validation`,
		WhenNotToUse: `• Traditional IT networks
• Networks without TSN support`,
	}
}

func tsnGateTimingDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         tsnGateTimingParameters(),
		Metrics:            tsnGateTimingMetrics(),
		SuccessCriteria:    "All gates within timing tolerance",
		FailureExplanation: "Network cannot support deterministic timing requirements",
		Examples:           tsnGateTimingExamples(),
		Tips:               tsnGateTimingTips(),
		CommonIssues:       nil,
		RFCSection:         "",
		SeeAlso:            []string{"traffic_isolation", "scheduled_latency"},
	}
}

func tsnGateTimingParameters() []Parameter {
	return []Parameter{
		{
			Name:       "Schedule",
			Flag:       "--schedule",
			Type:       "JSON or file path",
			Default:    "",
			Required:   true,
			TechDesc:   "Gate Control List schedule to validate",
			LaymanDesc: "The timing schedule to test against",
			Example:    "--schedule gate_schedule.json",
		},
		{
			Name:       "Tolerance",
			Flag:       "--tolerance",
			Type:       "integer (nanoseconds)",
			Default:    "1000",
			Required:   false,
			TechDesc:   "Acceptable timing deviation",
			LaymanDesc: "How much timing error is acceptable",
			Example:    "--tolerance 500",
		},
	}
}

func tsnGateTimingMetrics() []Metric {
	return []Metric{
		{
			Name:       "Gate Accuracy",
			Unit:       "nanoseconds deviation",
			GoodRange:  "Within tolerance",
			BadMeaning: "Gates not opening/closing on schedule",
		},
		{
			Name:       "Jitter",
			Unit:       "nanoseconds",
			GoodRange:  "<tolerance",
			BadMeaning: "Timing too inconsistent for TSN",
		},
	}
}

func tsnGateTimingExamples() []Example {
	return []Example{
		{
			Desc:    "Gate timing validation",
			Command: "stem test -i eth0 -t gate_timing --schedule schedule.json",
			Output:  "Gate accuracy: 250ns avg, 850ns max - PASS",
		},
	}
}

func tsnGateTimingTips() []string {
	return []string{
		"Ensure all devices are time-synchronized via PTP (IEEE 1588)",
		"Test with realistic traffic patterns",
		"Measure during worst-case scenarios",
	}
}

func tsnTrafficIsolation() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       "traffic_isolation",
			Name:     "TSN Traffic Class Isolation Test",
			Standard: "IEEE 802.1Qbv/Qbu",
			Category: StandardTSN,
		},
		tsnTrafficIsolationDescriptions(),
		tsnTrafficIsolationUsage(),
		tsnTrafficIsolationDetails(),
	)
}

func tsnTrafficIsolationDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Verifies that critical traffic is protected from other traffic classes.",
		TechDesc: `The Traffic Class Isolation Test verifies that TSN traffic classes
are properly isolated. It ensures that high-priority scheduled traffic is not
impacted by lower-priority or best-effort traffic, and that frame preemption
(802.1Qbu) is functioning correctly if configured.`,
		LaymanDesc: `In a TSN network, critical traffic (like robot control commands)
must be protected from interference by regular traffic (like file downloads).

This test verifies:
• Critical packets get through on time even when network is busy
• Lower priority traffic doesn't interfere with important traffic
• Protection mechanisms are working

Think of it like an ambulance lane - emergency vehicles get through regardless
of regular traffic.`,
	}
}

func tsnTrafficIsolationUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• Mixed traffic TSN networks
• Industrial networks with critical + regular traffic
• Validating traffic priority enforcement`,
		WhenNotToUse: `• Networks without traffic class differentiation`,
	}
}

func tsnTrafficIsolationDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         tsnTrafficIsolationParameters(),
		Metrics:            tsnTrafficIsolationMetrics(),
		SuccessCriteria:    "",
		FailureExplanation: "",
		Examples:           tsnTrafficIsolationExamples(),
		Tips:               nil,
		CommonIssues:       nil,
		RFCSection:         "",
		SeeAlso:            []string{"gate_timing", "scheduled_latency"},
	}
}

func tsnTrafficIsolationParameters() []Parameter {
	return []Parameter{
		{
			Name:       "Critical Class",
			Flag:       "--critical-class",
			Type:       "integer (priority)",
			Default:    "7",
			Required:   false,
			TechDesc:   "Priority of critical traffic class",
			LaymanDesc: "Priority level of the important traffic",
			Example:    "--critical-class 7",
		},
		{
			Name:       "Background Load",
			Flag:       "--background-load",
			Type:       "integer (Mbps)",
			Default:    "line rate",
			Required:   false,
			TechDesc:   "Background traffic rate",
			LaymanDesc: "How much competing traffic to generate",
			Example:    "--background-load 800",
		},
	}
}

func tsnTrafficIsolationMetrics() []Metric {
	return []Metric{
		{
			Name:       "Isolation Effectiveness",
			Unit:       "pass/fail",
			GoodRange:  "Pass",
			BadMeaning: "Critical traffic affected by other classes",
		},
		{
			Name:       "Critical Latency Under Load",
			Unit:       "microseconds",
			GoodRange:  "Within requirements",
			BadMeaning: "Critical traffic delayed by background traffic",
		},
	}
}

func tsnTrafficIsolationExamples() []Example {
	return []Example{
		{
			Desc:    "Traffic isolation test",
			Command: "stem test -i eth0 -t traffic_isolation --critical-class 7",
			Output:  "Isolation: PASS, Critical latency: 15µs under full load",
		},
	}
}

func tsnScheduledLatency() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       "scheduled_latency",
			Name:     "TSN Scheduled Latency Test",
			Standard: "IEEE 802.1Qbv",
			Category: StandardTSN,
		},
		tsnScheduledLatencyDescriptions(),
		tsnScheduledLatencyUsage(),
		tsnScheduledLatencyDetails(),
	)
}

func tsnScheduledLatencyDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Measures if packets arrive exactly when scheduled.",
		TechDesc: `The Scheduled Latency Test measures the end-to-end latency for
scheduled traffic and verifies it meets deterministic timing requirements.
Unlike best-effort latency which can vary, TSN scheduled traffic must arrive
within a precise time window.`,
		LaymanDesc: `In TSN networks, packets don't just need to arrive - they need
to arrive at EXACTLY the right time.

Regular networks: "Here's your packet... eventually"
TSN networks: "Here's your packet at exactly 10:00:00.000125"

This test measures:
• Does traffic arrive in its scheduled window?
• How consistent is the timing?
• Is the network deterministic enough for your application?

Critical for factory automation, robotics, and automotive applications where
"close enough" timing isn't good enough.`,
	}
}

func tsnScheduledLatencyUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• Deterministic latency validation
• Industrial control network certification
• Automotive ethernet validation`,
		WhenNotToUse: `• Networks without timing requirements`,
	}
}

func tsnScheduledLatencyDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         tsnScheduledLatencyParameters(),
		Metrics:            tsnScheduledLatencyMetrics(),
		SuccessCriteria:    "",
		FailureExplanation: "",
		Examples:           tsnScheduledLatencyExamples(),
		Tips:               nil,
		CommonIssues:       nil,
		RFCSection:         "",
		SeeAlso:            []string{"gate_timing", "traffic_isolation"},
	}
}

func tsnScheduledLatencyParameters() []Parameter {
	return []Parameter{
		{
			Name:       "Target Latency",
			Flag:       "--target-latency",
			Type:       "integer (microseconds)",
			Default:    "1000",
			Required:   false,
			TechDesc:   "Expected latency for scheduled traffic",
			LaymanDesc: "The timing target in microseconds",
			Example:    "--target-latency 500",
		},
		{
			Name:       "Window",
			Flag:       "--window",
			Type:       "integer (microseconds)",
			Default:    "100",
			Required:   false,
			TechDesc:   "Acceptable deviation from target",
			LaymanDesc: "How much variation is acceptable",
			Example:    "--window 50",
		},
	}
}

func tsnScheduledLatencyMetrics() []Metric {
	return []Metric{
		{
			Name:       "Scheduled Latency",
			Unit:       "microseconds",
			GoodRange:  "Within target ± window",
			BadMeaning: "Traffic not meeting timing requirements",
		},
		{
			Name:       "Timing Variance",
			Unit:       "microseconds",
			GoodRange:  "<window",
			BadMeaning: "Too much variation for deterministic operation",
		},
	}
}

func tsnScheduledLatencyExamples() []Example {
	return []Example{
		{
			Desc:    "Scheduled latency test",
			Command: "stem test -i eth0 -t scheduled_latency --target-latency 500 --window 50",
			Output:  "Latency: 485µs avg, 510µs max - PASS",
		},
	}
}

func tsnFull() TestHelp {
	return TestHelp{
		ID:       "tsn_full",
		Name:     "TSN Full Validation Suite",
		Standard: "IEEE 802.1Qbv/Qbu",
		Category: StandardTSN,

		Summary: "Complete TSN network validation - all timing and isolation tests.",

		TechDesc: `Runs the complete TSN validation suite including Gate Timing,
Traffic Isolation, and Scheduled Latency tests. This comprehensive test validates
that a TSN network is properly configured for deterministic operation.`,

		LaymanDesc: `The complete test for Time-Sensitive Networks - runs everything.

Includes:
1. Gate Timing - are time slots working correctly?
2. Traffic Isolation - is critical traffic protected?
3. Scheduled Latency - is timing deterministic?

Use this for final validation of TSN industrial networks before putting them
into production.`,

		WhenToUse: `• Complete TSN network validation
• Pre-production certification
• Industrial network acceptance`,

		WhenNotToUse: `• Troubleshooting specific issues (use individual tests)`,

		Parameters: nil,

		Metrics: nil,

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "Full TSN validation",
				Command: "stem test -i eth0 -t tsn_full --schedule schedule.json",
				Output:  "Gate Timing: PASS, Isolation: PASS, Latency: PASS",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "",
		SeeAlso:    []string{"gate_timing", "traffic_isolation", "scheduled_latency"},
	}
}

// GetAllCategories returns all test categories.
func GetAllCategories() map[string]Category {
	return map[string]Category{
		CatRFC2544: {
			ID:       CatRFC2544,
			Name:     StandardRFC2544,
			FullName: "Benchmarking Methodology for Network Interconnect Devices",
			Summary:  "The standard tests for measuring raw network equipment performance.",
			Description: `RFC 2544 defines benchmarking methodology for network devices.
These tests measure fundamental performance characteristics: throughput, latency,
frame loss, and burst handling. Use these tests for equipment validation and comparison.`,
			Tests: []string{
				TestTypeThroughput,
				TestTypeLatency,
				TestTypeFrameLoss,
				"back_to_back",
				"system_recovery",
				"reset",
			},
			WhenToUse: "Equipment benchmarking, performance validation, comparing vendors",
			Standard:  StandardRFC2544,
			SeeAlso:   []string{CatY1564, CatRFC2889},
		},
		CatY1564: {
			ID:       CatY1564,
			Name:     StandardY1564,
			FullName: "Ethernet Service Activation Test Methodology",
			Summary:  "The carrier standard for turning up ethernet services.",
			Description: `ITU-T Y.1564 defines the methodology for activating and validating
carrier ethernet services. These tests verify that a service meets its SLA parameters
at progressive load levels and over extended duration.`,
			Tests:     []string{"y1564_config", "y1564_performance", "y1564_full"},
			WhenToUse: "Carrier service activation, SLA validation, service acceptance",
			Standard:  StandardITUY1564,
			SeeAlso:   []string{CatMEF, CatRFC2544},
		},
		CatRFC2889: {
			ID:       CatRFC2889,
			Name:     StandardRFC2889,
			FullName: "Benchmarking Methodology for LAN Switching Devices",
			Summary:  "Tests specifically for switch/bridge performance characteristics.",
			Description: `RFC 2889 extends RFC 2544 for testing LAN switches. These tests
measure switch-specific characteristics like forwarding rate across multiple ports,
MAC address table capacity, learning rate, and congestion handling.`,
			Tests:     []string{"forwarding", "address_cache", "learning_rate", "broadcast", "congestion"},
			WhenToUse: "Switch validation, data center planning, MAC table capacity verification",
			Standard:  StandardRFC2889,
			SeeAlso:   []string{CatRFC2544},
		},
		CatRFC6349: {
			ID:       CatRFC6349,
			Name:     StandardRFC6349,
			FullName: "Framework for TCP Throughput Testing",
			Summary:  "Tests that measure real TCP application performance.",
			Description: `RFC 6349 provides methodology for testing TCP throughput, which
represents actual application performance. These tests measure achievable TCP throughput
and help identify network factors affecting TCP performance.`,
			Tests:     []string{"tcp_throughput", "path_analysis"},
			WhenToUse: "Application performance testing, WAN optimization, TCP troubleshooting",
			Standard:  StandardRFC6349,
			SeeAlso:   []string{CatRFC2544},
		},
		CatY1731: {
			ID:       CatY1731,
			Name:     "Y.1731",
			FullName: "OAM Functions and Mechanisms for Ethernet Networks",
			Summary:  "Operations, Administration, and Maintenance for carrier ethernet.",
			Description: `ITU-T Y.1731 defines OAM functions for monitoring and maintaining
ethernet services. These tools provide in-service monitoring capabilities including
delay measurement, loss measurement, and connectivity verification.`,
			Tests:     []string{"frame_delay", "y1731_frame_loss", "synthetic_loss", "loopback"},
			WhenToUse: "Production monitoring, SLA verification, fault isolation",
			Standard:  StandardITUY1731,
			SeeAlso:   []string{CatY1564, CatMEF},
		},
		CatMEF: {
			ID:       CatMEF,
			Name:     StandardMEF,
			FullName: "Metro Ethernet Forum Service Tests",
			Summary:  "Industry standard tests for carrier ethernet services.",
			Description: `MEF (Metro Ethernet Forum) defines service specifications and
testing methodologies for carrier ethernet. These tests validate services against
MEF specifications including bandwidth profiles and Class of Service.`,
			Tests:     []string{"mef_config", "mef_performance", "mef_full"},
			WhenToUse: "MEF-certified service validation, multi-CoS testing, carrier acceptance",
			Standard:  "MEF 14/48",
			SeeAlso:   []string{CatY1564, CatY1731},
		},
		CatTSN: {
			ID:       CatTSN,
			Name:     StandardTSN,
			FullName: "Time-Sensitive Networking",
			Summary:  "Tests for deterministic, time-critical industrial networks.",
			Description: `IEEE 802.1 Time-Sensitive Networking tests validate networks
requiring deterministic timing. These tests verify that time-aware shaping, traffic
isolation, and scheduled latency meet industrial automation requirements.`,
			Tests:     []string{"gate_timing", "traffic_isolation", "scheduled_latency", "tsn_full"},
			WhenToUse: "Industrial automation, automotive ethernet, deterministic networking",
			Standard:  "IEEE 802.1Qbv/Qbu",
			SeeAlso:   []string{CatRFC2544},
		},
	}
}
