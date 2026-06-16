/*
 * The Stem - Test Documentation
 *
 * Y.1731 test help content.
 */

package help

// ============================================================================
// Y.1731 Tests - OAM Functions for Ethernet Networks.
// ============================================================================

func y1731FrameDelay() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       "frame_delay",
			Name:     "Frame Delay Measurement",
			Standard: StandardITUY1731,
			Category: "Y.1731",
		},
		y1731FrameDelayDescriptions(),
		y1731FrameDelayUsage(),
		y1731FrameDelayDetails(),
	)
}

func y1731FrameDelayDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Precise one-way and two-way delay measurements using OAM.",
		TechDesc: `Y.1731 Frame Delay Measurement (DMM/DMR) provides precise delay
measurement using Operations, Administration, and Maintenance (OAM) frames.
Unlike RFC 2544 latency which uses test traffic, Y.1731 can measure delay
on production networks using lightweight OAM frames.

Supports both one-way delay (requires synchronized clocks) and two-way delay
(no synchronization required) measurements.`,
		LaymanDesc: `Super-precise timing measurements for carrier networks.

Think of it as a stopwatch for network packets:
• Two-way: Round trip time (like a ping, but more precise)
• One-way: Time in just one direction (needs synchronized clocks)

Used by:
• Carriers to verify SLA compliance
• Financial networks needing precise timing
• Real-time applications requiring guaranteed delay

More accurate than regular ping because it measures at the network level,
not the application level.`,
	}
}

func y1731FrameDelayUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• SLA monitoring in production
• Carrier ethernet service monitoring
• When you need precise timing measurements`,
		WhenNotToUse: `• Initial service turn-up (use Y.1564)
• If carrier network doesn't support Y.1731`,
	}
}

func y1731FrameDelayDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         y1731FrameDelayParameters(),
		Metrics:            y1731FrameDelayMetrics(),
		SuccessCriteria:    "",
		FailureExplanation: "",
		Examples:           y1731FrameDelayExamples(),
		Tips:               nil,
		CommonIssues:       nil,
		RFCSection:         "",
		SeeAlso:            []string{TestTypeLatency, "y1731_frame_loss"},
	}
}

func y1731FrameDelayParameters() []Parameter {
	return []Parameter{
		{
			Name:       "Mode",
			Flag:       "--mode",
			Type:       "string",
			Default:    "two-way",
			Required:   false,
			TechDesc:   "Measurement mode: one-way or two-way",
			LaymanDesc: "One-way needs synchronized clocks; two-way doesn't",
			Example:    "--mode one-way",
		},
		{
			Name:       "Interval",
			Flag:       "--interval",
			Type:       "integer (milliseconds)",
			Default:    "100",
			Required:   false,
			TechDesc:   "Interval between measurements",
			LaymanDesc: "How often to take measurements",
			Example:    "--interval 1000",
		},
	}
}

func y1731FrameDelayMetrics() []Metric {
	return []Metric{
		{
			Name:       "Frame Delay",
			Unit:       "microseconds",
			GoodRange:  "Per SLA definition",
			BadMeaning: "Exceeds SLA threshold",
		},
		{
			Name:       "Frame Delay Variation",
			Unit:       "microseconds",
			GoodRange:  "Per SLA definition",
			BadMeaning: "High jitter may affect quality",
		},
	}
}

func y1731FrameDelayExamples() []Example {
	return []Example{
		{
			Desc:    "Two-way delay measurement",
			Command: "stem test -t frame_delay --mode two-way",
			Output:  "Delay: 1.234ms, Jitter: 0.089ms",
		},
	}
}

func y1731FrameLoss() TestHelp {
	return TestHelp{
		ID:       "y1731_frame_loss",
		Name:     "Frame Loss Measurement",
		Standard: StandardITUY1731,
		Category: "Y.1731",

		Summary: "Monitors packet loss on production carrier networks.",

		TechDesc: `Y.1731 Frame Loss Measurement (LMM/LMR) provides continuous loss
monitoring using OAM frame counters. Unlike RFC 2544 frame loss which requires
dedicated test traffic, Y.1731 loss measurement can operate alongside production
traffic by comparing frame counts between endpoints.`,

		LaymanDesc: `Continuously monitors if packets are being lost, without disrupting
normal traffic.

Works like an accountant for packets:
• Counts packets sent
• Counts packets received
• Reports any difference

Benefits:
• Works on live networks (no test traffic needed)
• Catches intermittent problems
• Continuous monitoring vs. point-in-time tests`,

		WhenToUse: `• Continuous service monitoring
• SLA compliance verification
• Proactive fault detection`,

		WhenNotToUse: `• Initial service testing (use Y.1564)
• If carrier doesn't support Y.1731`,

		Parameters: nil,

		Metrics: []Metric{
			{
				Name:       "Frame Loss Ratio",
				Unit:       "percentage",
				GoodRange:  "<0.001%",
				BadMeaning: "SLA may be violated",
			},
		},

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "Frame loss monitoring",
				Command: "stem test -t y1731_frame_loss --duration 3600",
				Output:  "Loss ratio: 0.0001% over 1 hour",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "",
		SeeAlso:    []string{TestTypeFrameLoss, "frame_delay"},
	}
}

func y1731SyntheticLoss() TestHelp {
	return TestHelp{
		ID:       "synthetic_loss",
		Name:     "Synthetic Loss Measurement",
		Standard: StandardITUY1731,
		Category: "Y.1731",

		Summary: "Continuous reliability monitoring using test signals.",

		TechDesc: `Synthetic Loss Measurement (SLM/SLR) uses dedicated OAM test frames
to measure loss independent of user traffic. This is useful when there's no or
variable user traffic, or when you want loss measurements isolated from user
traffic patterns.`,

		LaymanDesc: `Sends special test signals to continuously check if the network is
working, like a heartbeat monitor for your network connection.

Unlike frame loss measurement that counts real traffic, this sends its own test
messages to verify connectivity regardless of whether users are sending data.

Think of it like a "network heartbeat" - always checking, always monitoring.`,

		WhenToUse: `• Links with variable or no user traffic
• Backup path monitoring
• Standby connection verification`,

		WhenNotToUse: `• High-traffic links where frame loss measurement works`,

		Parameters: nil,

		Metrics: []Metric{
			{
				Name:       "Synthetic Loss Ratio",
				Unit:       "percentage",
				GoodRange:  "0%",
				BadMeaning: "Network path has problems",
			},
		},

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "Synthetic loss monitoring",
				Command: "stem test -t synthetic_loss --interval 1000",
				Output:  "Synthetic loss: 0%, Path status: OK",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "",
		SeeAlso:    []string{"y1731_frame_loss", "loopback"},
	}
}

func y1731Loopback() TestHelp {
	return TestHelp{
		ID:       "loopback",
		Name:     "Loopback Test",
		Standard: StandardITUY1731,
		Category: "Y.1731",

		Summary: "Quick connectivity check using OAM loopback.",

		TechDesc: `The Y.1731 Loopback (LB) function verifies connectivity between
Maintenance Entity Group End Points (MEPs). It's similar to ICMP ping but operates
at Layer 2 and can target specific MEP IDs in the carrier network.`,

		LaymanDesc: `A "ping" for carrier ethernet networks.

Like shouting into a canyon and waiting for an echo:
• Send a test message
• Wait for it to come back
• If it comes back, the path is working

Useful for:
• Quick "is it working?" checks
• Isolating where a problem is
• Verifying specific network segments`,

		WhenToUse: `• Quick connectivity verification
• Troubleshooting connectivity issues
• Verifying OAM path functionality`,

		WhenNotToUse: `• Performance testing (use other Y.1731 tests)`,

		Parameters: nil,

		Metrics: []Metric{
			{
				Name:       "Response",
				Unit:       "pass/fail",
				GoodRange:  "Pass",
				BadMeaning: "Connectivity problem",
			},
			{
				Name:       "Response Time",
				Unit:       "milliseconds",
				GoodRange:  "Consistent with path length",
				BadMeaning: "Unusually high indicates problem",
			},
		},

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "Loopback test",
				Command: "stem test -t loopback --target-mep 100",
				Output:  "Loopback response received in 1.2ms",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "",
		SeeAlso:    []string{"frame_delay", "synthetic_loss"},
	}
}
