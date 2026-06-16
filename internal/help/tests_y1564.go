/*
 * The Stem - Test Documentation
 *
 * Y.1564 test help content.
 */

package help

// ============================================================================
// Y.1564 Tests - Ethernet Service Activation Test Methodology.
// ============================================================================

func y1564Config() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       "y1564_config",
			Name:     "Y.1564 Service Configuration Test",
			Standard: StandardITUY1564,
			Category: StandardY1564,
		},
		y1564ConfigDescriptions(),
		y1564ConfigUsage(),
		y1564ConfigDetails(),
	)
}

func y1564ConfigDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Validates carrier ethernet service at 25%, 50%, 75%, and 100% of committed rate.",
		TechDesc: `The Y.1564 Service Configuration Test validates that an Ethernet service
	meets its Service Level Agreement (SLA) parameters at progressive load steps. The test
	verifies CIR (Committed Information Rate), EIR (Excess Information Rate), frame delay,
	frame delay variation (jitter), and frame loss at 25%, 50%, 75%, and 100% of the
	configured rates.

	This methodical approach ensures the service can meet commitments at all traffic
	levels, not just under specific conditions. Each step must pass defined thresholds
	before the service is considered properly configured.`,
		LaymanDesc: `When you buy an ethernet service from a carrier (like a 100 Mbps
	business internet connection), this test verifies you're getting what you paid for.

	The test checks your connection at different levels:
	• 25% speed: Light usage - is basic connectivity working?
	• 50% speed: Medium usage - is half the promised speed reliable?
	• 75% speed: Heavy usage - still performing well?
	• 100% speed: Maximum usage - getting full promised bandwidth?

	At each level, it checks:
	• Speed: Are you getting the bandwidth you're paying for?
	• Delay: Is the connection responsive?
	• Reliability: Are packets being delivered without loss?

	This is the industry standard for "turning up" new carrier ethernet services.`,
	}
}

func y1564ConfigUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• New service activation and turn-up
	• Service verification after carrier maintenance
	• SLA dispute resolution
	• Regular service quality audits
	• Contract compliance validation`,
		WhenNotToUse: `• For raw equipment benchmarking (use RFC 2544)
	• For extended performance validation (use Y.1564 Performance test)
	• For TCP application testing (use RFC 6349)`,
	}
}

func y1564ConfigDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         y1564ConfigParameters(),
		Metrics:            y1564ConfigMetrics(),
		SuccessCriteria:    "All metrics within thresholds at all four CIR steps (25%, 50%, 75%, 100%)",
		FailureExplanation: "Service does not meet SLA - contact carrier for resolution",
		Examples:           y1564ConfigExamples(),
		Tips:               y1564ConfigTips(),
		CommonIssues:       y1564ConfigIssues(),
		RFCSection:         "",
		SeeAlso:            []string{"y1564_performance", TestTypeThroughput, "mef_config"},
	}
}

func y1564ConfigParameters() []Parameter {
	return []Parameter{
		{
			Name:       TermCIR,
			Flag:       FlagCIR,
			Type:       "integer (Mbps)",
			Default:    "1000",
			Required:   true,
			TechDesc:   "Committed Information Rate - guaranteed bandwidth",
			LaymanDesc: "The speed your contract guarantees (e.g., 100 for 100 Mbps)",
			Example:    ExampleCIR100,
		},
		{
			Name:       "EIR",
			Flag:       "--eir",
			Type:       "integer (Mbps)",
			Default:    "0",
			Required:   false,
			TechDesc:   "Excess Information Rate - burst bandwidth above CIR",
			LaymanDesc: "Extra bandwidth you might get when the network isn't busy",
			Example:    "--eir 50",
		},
		{
			Name:       "Frame Delay Threshold",
			Flag:       "--delay-threshold",
			Type:       "float (milliseconds)",
			Default:    "10.0",
			Required:   false,
			TechDesc:   "Maximum acceptable frame delay",
			LaymanDesc: "Maximum allowed delay in milliseconds",
			Example:    "--delay-threshold 5.0",
		},
		{
			Name:       "Jitter Threshold",
			Flag:       "--jitter-threshold",
			Type:       "float (milliseconds)",
			Default:    "3.0",
			Required:   false,
			TechDesc:   "Maximum acceptable frame delay variation",
			LaymanDesc: "Maximum allowed variation in delay",
			Example:    "--jitter-threshold 2.0",
		},
		{
			Name:       "Loss Threshold",
			Flag:       "--loss-threshold",
			Type:       "float (percentage)",
			Default:    "0.001",
			Required:   false,
			TechDesc:   "Maximum acceptable frame loss rate",
			LaymanDesc: "Maximum allowed packet loss (0.001% = 1 in 100,000)",
			Example:    "--loss-threshold 0.0001",
		},
		{
			Name:       "Step Duration",
			Flag:       "--step-duration",
			Type:       "integer (seconds)",
			Default:    "60",
			Required:   false,
			TechDesc:   "Duration of each CIR step test",
			LaymanDesc: "How long to test at each speed level",
			Example:    "--step-duration 120",
		},
	}
}

func y1564ConfigMetrics() []Metric {
	return []Metric{
		{
			Name:       "IR (Information Rate)",
			Unit:       UnitMbps,
			GoodRange:  "Within 1% of configured CIR",
			BadMeaning: "Service not delivering promised bandwidth",
		},
		{
			Name:       "FD (Frame Delay)",
			Unit:       "milliseconds",
			GoodRange:  "Below threshold at all steps",
			BadMeaning: "Exceeds SLA delay commitment",
		},
		{
			Name:       "FDV (Frame Delay Variation)",
			Unit:       "milliseconds",
			GoodRange:  "Below threshold at all steps",
			BadMeaning: "Jitter too high for voice/video",
		},
		{
			Name:       "FLR (Frame Loss Ratio)",
			Unit:       "percentage",
			GoodRange:  "Below threshold (typically <0.001%)",
			BadMeaning: "Unacceptable packet loss",
		},
	}
}

func y1564ConfigExamples() []Example {
	return []Example{
		{
			Desc:    "Test 100 Mbps service",
			Command: "stem test -i eth0 -t y1564_config --cir 100",
			Output:  "Step 25%: PASS, Step 50%: PASS, Step 75%: PASS, Step 100%: PASS",
		},
		{
			Desc:    "Test with strict thresholds",
			Command: "stem test -i eth0 -t y1564_config --cir 1000 --delay-threshold 2.0 --loss-threshold 0.0001",
			Output:  "Testing 1 Gbps service with strict SLA",
		},
	}
}

func y1564ConfigTips() []string {
	return []string{
		"Run this test when first activating a new carrier service",
		"Keep test results as baseline for future comparison",
		"If any step fails, the service needs adjustment before acceptance",
		"CIR should match your contract exactly",
	}
}

func y1564ConfigIssues() []Issue {
	return []Issue{
		{
			Problem:  "Fails at 100% CIR but passes at 75%",
			Cause:    "Service not provisioned to full contracted rate",
			Solution: "Contact carrier to verify provisioning",
		},
		{
			Problem:  "High frame delay at all steps",
			Cause:    "Distance/routing issue or congested path",
			Solution: "Request path optimization from carrier",
		},
	}
}

func y1564Performance() TestHelp {
	return TestHelp{
		ID:       "y1564_performance",
		Name:     "Y.1564 Service Performance Test",
		Standard: StandardITUY1564,
		Category: StandardY1564,

		Summary: "Extended duration test to validate service quality over time.",

		TechDesc: `The Y.1564 Service Performance Test validates that an Ethernet service
maintains its SLA parameters over an extended period. Unlike the Configuration Test
which validates at progressive rates, the Performance Test runs at the full CIR for
an extended duration (typically 15 minutes to 24 hours) to verify sustained performance.

This test detects issues that only appear under sustained load, such as thermal
throttling, memory leaks, or intermittent congestion patterns.`,

		LaymanDesc: `After passing the initial speed test, can your network connection
maintain that performance for hours?

Think of it like a car test:
• Configuration test: "Can it reach 100 mph?" (short sprint)
• Performance test: "Can it maintain 100 mph for hours?" (endurance)

This test runs your connection at full speed for an extended time to catch:
• Equipment that overheats and slows down
• Intermittent problems that come and go
• Issues that only appear under sustained load
• Time-of-day performance variations

A 15-minute test is typical for activation; longer tests (hours) are used for
thorough validation.`,

		WhenToUse: `• After passing Service Configuration Test
• Validating SLA compliance over time
• Detecting intermittent issues
• Extended burn-in testing`,

		WhenNotToUse: `• Initial service turn-up (do Configuration Test first)
• Quick spot-checks (use Configuration Test)`,

		Parameters: []Parameter{
			{
				Name:       TermCIR,
				Flag:       FlagCIR,
				Type:       "integer (Mbps)",
				Default:    "1000",
				Required:   true,
				TechDesc:   TermCIRFull,
				LaymanDesc: "Your contracted bandwidth",
				Example:    ExampleCIR100,
			},
			{
				Name:       LabelDuration,
				Flag:       FlagDuration,
				Type:       "integer (minutes)",
				Default:    "15",
				Required:   false,
				TechDesc:   "Test duration in minutes",
				LaymanDesc: "How long to run the test",
				Example:    ExampleDuration60,
			},
		},

		Metrics: []Metric{
			{
				Name:       "Sustained Rate",
				Unit:       UnitMbps,
				GoodRange:  "Within 1% of CIR for entire duration",
				BadMeaning: "Performance degrades over time",
			},
			{
				Name:       "FLR Over Time",
				Unit:       "percentage",
				GoodRange:  "Consistently below threshold",
				BadMeaning: "Intermittent loss indicates instability",
			},
		},

		SuccessCriteria:    "All metrics within thresholds for entire test duration",
		FailureExplanation: "Service shows instability over time",

		Examples: []Example{
			{
				Desc:    "15-minute performance test",
				Command: "stem test -i eth0 -t y1564_performance --cir 100 --duration 15",
				Output:  "Performance stable over 15 minutes",
			},
			{
				Desc:    "Extended overnight test",
				Command: "stem test -i eth0 -t y1564_performance --cir 1000 --duration 480",
				Output:  "Running 8-hour endurance test",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "",
		SeeAlso:    []string{"y1564_config", "y1564_full"},
	}
}

func y1564Full() TestHelp {
	return TestHelp{
		ID:       "y1564_full",
		Name:     "Y.1564 Full SAC Test",
		Standard: StandardITUY1564,
		Category: StandardY1564,

		Summary: "Complete Service Activation Test - Configuration followed by Performance.",

		TechDesc: `The Full SAC (Service Activation) Test combines both the Service
Configuration Test and Service Performance Test into a complete validation sequence.
First, the Configuration Test validates SLA parameters at 25%, 50%, 75%, and 100%
of CIR. If all steps pass, the Performance Test runs for the specified duration
at full CIR.

This is the complete Y.1564 service activation methodology as defined by MEF and ITU-T.`,

		LaymanDesc: `The complete, official test for verifying a carrier ethernet service.

This runs both tests in sequence:
1. Configuration Test: Check speeds at 25%, 50%, 75%, 100%
2. If step 1 passes: Performance Test: Run at full speed for extended time

This is what carriers use to officially "turn up" a new service. When you sign
off on a Y.1564 SAC test, you're accepting the service as meeting specifications.

Total test time is typically 30 minutes for a standard activation.`,

		WhenToUse: `• Official service activation and acceptance
• Contract sign-off requiring full SAC test
• Comprehensive service validation`,

		WhenNotToUse: `• Quick troubleshooting (use individual tests)
• Time-constrained situations`,

		Parameters: []Parameter{
			{
				Name:       TermCIR,
				Flag:       FlagCIR,
				Type:       "integer (Mbps)",
				Default:    "1000",
				Required:   true,
				TechDesc:   TermCIRFull,
				LaymanDesc: "Your contracted bandwidth",
				Example:    ExampleCIR100,
			},
			{
				Name:       "Performance Duration",
				Flag:       "--perf-duration",
				Type:       "integer (minutes)",
				Default:    "15",
				Required:   false,
				TechDesc:   "Duration of performance test phase",
				LaymanDesc: "How long to run the endurance portion",
				Example:    "--perf-duration 30",
			},
		},

		Metrics: nil,

		SuccessCriteria:    "Both Configuration and Performance tests pass",
		FailureExplanation: "Service does not meet acceptance criteria",

		Examples: []Example{
			{
				Desc:    "Full service activation test",
				Command: "stem test -i eth0 -t y1564_full --cir 100",
				Output:  "Config Test: PASS, Performance Test: PASS - Service Accepted",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "",
		SeeAlso:    []string{"y1564_config", "y1564_performance", "mef_full"},
	}
}
