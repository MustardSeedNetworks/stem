/*
 * The Stem - Test Documentation
 *
 * MEF test help content.
 */

package help

// ============================================================================
// MEF Tests - Carrier Ethernet Service Tests.
// ============================================================================

func mefConfig() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       "mef_config",
			Name:     "MEF Service Configuration Test",
			Standard: "MEF 14/48",
			Category: StandardMEF,
		},
		mefConfigDescriptions(),
		mefConfigUsage(),
		mefConfigDetails(),
	)
}

func mefConfigDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Validates carrier ethernet service configuration per MEF standards.",
		TechDesc: `The MEF Service Configuration Test validates that a carrier ethernet
service meets MEF specifications for bandwidth profiles, Class of Service (CoS)
identification, and frame handling. Tests include CIR/EIR validation, frame delay,
delay variation, and loss for each configured traffic class.`,
		LaymanDesc: `The official carrier ethernet validation, as defined by the MEF
(Metro Ethernet Forum) industry group.

MEF is the organization that sets standards for business ethernet services. This
test verifies:
• Bandwidth matches what you're paying for
• Traffic priority (CoS) is working correctly
• All service parameters match the contract

Using MEF tests means you're testing against industry-standard criteria, making
results comparable and contractually meaningful.`,
	}
}

func mefConfigUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• MEF-certified service validation
• Multi-CoS service testing
• Carrier service acceptance`,
		WhenNotToUse: `• Simple single-class services (Y.1564 may suffice)`,
	}
}

func mefConfigDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         mefConfigParameters(),
		Metrics:            mefConfigMetrics(),
		SuccessCriteria:    "",
		FailureExplanation: "",
		Examples:           mefConfigExamples(),
		Tips:               nil,
		CommonIssues:       nil,
		RFCSection:         "",
		SeeAlso:            []string{"y1564_config", "mef_performance"},
	}
}

func mefConfigParameters() []Parameter {
	return []Parameter{
		{
			Name:       TermCIR,
			Flag:       FlagCIR,
			Type:       "integer (Mbps)",
			Default:    "1000",
			Required:   true,
			TechDesc:   TermCIRFull,
			LaymanDesc: "Guaranteed bandwidth",
			Example:    ExampleCIR100,
		},
		{
			Name:       "CoS",
			Flag:       "--cos",
			Type:       "integer (0-7)",
			Default:    "0",
			Required:   false,
			TechDesc:   "Class of Service to test",
			LaymanDesc: "Priority level (0=lowest, 7=highest)",
			Example:    "--cos 5",
		},
	}
}

func mefConfigMetrics() []Metric {
	return []Metric{
		{
			Name:       "Bandwidth Compliance",
			Unit:       "pass/fail",
			GoodRange:  "Pass",
			BadMeaning: "Service not meeting bandwidth commitment",
		},
		{
			Name:       "CoS Handling",
			Unit:       "pass/fail",
			GoodRange:  "Pass",
			BadMeaning: "Priority not being honored",
		},
	}
}

func mefConfigExamples() []Example {
	return []Example{
		{
			Desc:    "MEF configuration test",
			Command: "stem test -i eth0 -t mef_config --cir 100 --cos 5",
			Output:  "Bandwidth: PASS, CoS: PASS",
		},
	}
}

func mefPerformance() TestHelp {
	return TestHelp{
		ID:       "mef_performance",
		Name:     "MEF Performance Test",
		Standard: "MEF 14/48",
		Category: StandardMEF,

		Summary: "Extended MEF service quality validation.",

		TechDesc: `The MEF Performance Test runs extended duration tests per MEF
specifications to validate sustained service quality. This includes performance
monitoring across all configured traffic classes over the specified duration.`,

		LaymanDesc: `Long-running test to make sure your carrier service stays good
over time, not just during a quick test.

Runs the MEF configuration tests for an extended period (15+ minutes) to verify:
• Consistent performance over time
• No degradation under sustained load
• All service classes maintained

This catches problems that only show up after running for a while.`,

		WhenToUse: `• After MEF Config test passes
• Extended service validation
• Pre-production sign-off`,

		WhenNotToUse: `• Quick spot checks`,

		Parameters: []Parameter{
			{
				Name:       LabelDuration,
				Flag:       FlagDuration,
				Type:       "integer (minutes)",
				Default:    "15",
				Required:   false,
				TechDesc:   "Test duration",
				LaymanDesc: "How long to run",
				Example:    ExampleDuration60,
			},
		},

		Metrics: nil,

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "MEF performance test",
				Command: "stem test -i eth0 -t mef_performance --cir 100 --duration 15",
				Output:  "15-minute performance: PASS",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "",
		SeeAlso:    []string{"mef_config", "y1564_performance"},
	}
}

func mefFull() TestHelp {
	return TestHelp{
		ID:       "mef_full",
		Name:     "MEF Full Test Suite",
		Standard: "MEF 14/48",
		Category: StandardMEF,

		Summary: "Complete MEF service validation - configuration plus performance.",

		TechDesc: `Runs the complete MEF test suite including Service Configuration
Test followed by Service Performance Test. This is the full validation sequence
for MEF-certified ethernet services.`,

		LaymanDesc: `The complete MEF certification test - everything in one run.

Runs both:
1. Configuration Test - verify service is set up right
2. Performance Test - verify it stays good over time

Use this for official service acceptance when MEF compliance is required.`,

		WhenToUse: `• Official MEF service acceptance
• Complete service validation`,

		WhenNotToUse: `• Troubleshooting (use individual tests)`,

		Parameters: nil,

		Metrics: nil,

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "Full MEF test",
				Command: "stem test -i eth0 -t mef_full --cir 100",
				Output:  "MEF Config: PASS, MEF Performance: PASS - Service Accepted",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "",
		SeeAlso:    []string{"mef_config", "mef_performance", "y1564_full"},
	}
}
