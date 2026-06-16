/*
 * The Stem - Test Documentation
 *
 * RFC 6349 test help content.
 */

package help

// ============================================================================
// RFC 6349 Tests - Framework for TCP Throughput Testing.
// ============================================================================

func rfc6349TCPThroughput() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       "tcp_throughput",
			Name:     "TCP Throughput Test",
			Standard: StandardRFC6349,
			Category: StandardRFC6349,
		},
		rfc6349TCPThroughputDescriptions(),
		rfc6349TCPThroughputUsage(),
		rfc6349TCPThroughputDetails(),
	)
}

func rfc6349TCPThroughputDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Measures real application throughput using TCP (like actual file downloads).",
		TechDesc: `The RFC 6349 TCP Throughput test measures the achieved TCP throughput
	between endpoints, accounting for the effects of TCP flow control, congestion
	avoidance, and operating system TCP stack behavior. Unlike Layer 2/3 tests that
	use raw frames, this test uses actual TCP connections.

	The test measures TCP Efficiency (ratio of actual to ideal throughput) and Buffer
	Delay (extra latency introduced by network buffers). These metrics reveal how
	real applications will perform over the network path.`,
		LaymanDesc: `This test measures REAL download/upload speeds - the kind you
	actually experience.

	Other tests (RFC 2544) measure raw network speed. But real applications:
	• Use TCP protocol (adds overhead for reliability)
	• Are affected by network latency
	• Slow down when packets are lost
	• Behave differently than raw speed tests

	This test shows:
	• Actual file transfer speeds you'll get
	• Why your "1 Gbps" connection downloads at 800 Mbps
	• If your network is optimized for real applications

	This is the difference between "theoretical maximum" and "what you'll actually get."`,
	}
}

func rfc6349TCPThroughputUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• Application performance troubleshooting
	• WAN optimization validation
	• Cloud connectivity assessment
	• Real-world performance baselining`,
		WhenNotToUse: `• Layer 2 equipment testing (use RFC 2544)
	• Service activation (use Y.1564)
	• When UDP performance is what matters`,
	}
}

func rfc6349TCPThroughputDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         rfc6349TCPThroughputParameters(),
		Metrics:            rfc6349TCPThroughputMetrics(),
		SuccessCriteria:    "TCP throughput meets application requirements",
		FailureExplanation: "Network may need optimization for TCP applications",
		Examples:           rfc6349TCPThroughputExamples(),
		Tips:               nil,
		CommonIssues:       nil,
		RFCSection:         "",
		SeeAlso:            []string{"path_analysis", TestTypeThroughput},
	}
}

func rfc6349TCPThroughputParameters() []Parameter {
	return []Parameter{
		{
			Name:       LabelDuration,
			Flag:       FlagDuration,
			Type:       "integer (seconds)",
			Default:    "30",
			Required:   false,
			TechDesc:   "Test duration",
			LaymanDesc: "How long to run the transfer test",
			Example:    ExampleDuration60,
		},
		{
			Name:       "Window Size",
			Flag:       "--window",
			Type:       "integer (KB) or 'auto'",
			Default:    ValueAuto,
			Required:   false,
			TechDesc:   "TCP window size (or auto-calculate from BDP)",
			LaymanDesc: "TCP buffer size - 'auto' calculates optimal value",
			Example:    "--window auto",
		},
	}
}

func rfc6349TCPThroughputMetrics() []Metric {
	return []Metric{
		{
			Name:       "TCP Throughput",
			Unit:       UnitMbps,
			GoodRange:  ">80% of link capacity",
			BadMeaning: "Application performance will suffer",
		},
		{
			Name:       "TCP Efficiency",
			Unit:       "percentage",
			GoodRange:  ">95%",
			BadMeaning: "Retransmissions reducing efficiency",
		},
		{
			Name:       "Buffer Delay",
			Unit:       "percentage of base RTT",
			GoodRange:  "<100%",
			BadMeaning: "Buffer bloat affecting latency",
		},
	}
}

func rfc6349TCPThroughputExamples() []Example {
	return []Example{
		{
			Desc:    "TCP throughput to remote server",
			Command: "stem test -t tcp_throughput --target 10.0.0.100",
			Output:  "TCP Throughput: 890 Mbps, Efficiency: 97%, Buffer Delay: 45%",
		},
	}
}

func rfc6349PathAnalysis() TestHelp {
	return TestHelp{
		ID:       "path_analysis",
		Name:     "Path Analysis Test",
		Standard: StandardRFC6349,
		Category: StandardRFC6349,

		Summary: "Analyzes what's limiting your network speed.",

		TechDesc: `Path Analysis characterizes the network path to determine the optimal
TCP parameters and identify performance bottlenecks. It measures Round-Trip Time
(RTT), path capacity (bottleneck bandwidth), and calculates the Bandwidth-Delay
Product (BDP) which determines optimal TCP window sizing.

This test helps diagnose why TCP throughput may not reach expected levels and
provides recommendations for TCP tuning.`,

		LaymanDesc: `Before speeding down a road, you'd want to know:
• How long is the trip? (latency)
• Are there bottlenecks? (narrow lanes)
• How much traffic can it handle?

This test answers those questions for your network:
• What's the slowest link in the path?
• How much delay is there?
• What TCP settings will work best?

Use this when downloads are slow and you want to know WHY, not just HOW slow.`,

		WhenToUse: `• TCP performance troubleshooting
• Before optimizing TCP settings
• Understanding WAN path characteristics`,

		WhenNotToUse: `• If you just need throughput numbers (use TCP Throughput test)`,

		Parameters: []Parameter{
			{
				Name:       "Target",
				Flag:       "--target",
				Type:       "IP address",
				Default:    "",
				Required:   true,
				TechDesc:   "Remote endpoint for path analysis",
				LaymanDesc: "The server you want to analyze the path to",
				Example:    "--target 192.168.1.100",
			},
		},

		Metrics: []Metric{
			{
				Name:       "RTT",
				Unit:       "milliseconds",
				GoodRange:  "<100ms for most applications",
				BadMeaning: "High latency will limit TCP throughput",
			},
			{
				Name:       "Bottleneck Bandwidth",
				Unit:       UnitMbps,
				GoodRange:  "Equal to expected path capacity",
				BadMeaning: "A link is slower than expected",
			},
			{
				Name:       "BDP",
				Unit:       "KB",
				GoodRange:  "Informational",
				BadMeaning: "N/A - used for TCP tuning recommendations",
			},
		},

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "Analyze path to server",
				Command: "stem test -t path_analysis --target 10.0.0.100",
				Output:  "RTT: 25ms, Bottleneck: 1000 Mbps, Optimal Window: 3.1 MB",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "",
		SeeAlso:    []string{"tcp_throughput", TestTypeLatency},
	}
}
