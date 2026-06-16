/*
 * The Stem - Test Documentation
 *
 * RFC 2544 test help content.
 */

package help

// ============================================================================
// RFC 2544 Tests - Benchmarking Methodology for Network Interconnect Devices.
// ============================================================================

func rfc2544Throughput() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       TestTypeThroughput,
			Name:     "Throughput Test",
			Standard: "RFC 2544 Section 26.1",
			Category: StandardRFC2544,
		},
		rfc2544ThroughputDescriptions(),
		rfc2544ThroughputUsage(),
		rfc2544ThroughputDetails(),
	)
}

func rfc2544ThroughputDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Finds the maximum speed your network can handle without dropping packets.",
		TechDesc: `The throughput test uses binary search to determine the maximum rate at which
	the DUT (Device Under Test) can forward frames without any frame loss. Starting at the
	theoretical maximum rate, the test iteratively adjusts the offered load based on whether
	frames were lost, converging on the maximum lossless rate. The result represents the
	maximum forwarding rate at which zero packet loss occurs for a given frame size.

	The test is performed for each of the standard frame sizes (64, 128, 256, 512, 1024,
	1280, 1518 bytes) to characterize performance across the full range of packet sizes.
	Binary search precision can be configured but typically converges within 0.1% accuracy.`,
		LaymanDesc: `Think of your network like a highway. This test finds out how many cars
	(data packets) can travel on it before traffic jams (packet loss) start happening.

	Here's how it works:
	1. Start by sending as much traffic as the network should handle
	2. If any packets get lost, slow down and try again
	3. If all packets arrive, speed up and try again
	4. Keep adjusting until we find the exact maximum speed with zero loss

	The result tells you the real-world capacity of your network equipment. If you're
	paying for a 1 Gbps connection, this test tells you if you're actually getting
	1 Gbps or something less.`,
	}
}

func rfc2544ThroughputUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• Validating new network equipment before deployment
	• Troubleshooting slow network performance
	• Verifying ISP is delivering promised bandwidth
	• Baseline testing after configuration changes
	• Quality assurance for network upgrades
	• SLA verification for service providers`,
		WhenNotToUse: `• If you need latency measurements (use Latency test instead)
	• For TCP application performance (use RFC 6349 TCP Throughput)
	• For switch MAC table testing (use RFC 2889 tests)
	• For carrier ethernet service activation (use Y.1564)
	• If network is in production and can't tolerate test traffic`,
	}
}

func rfc2544ThroughputDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         rfc2544ThroughputParameters(),
		Metrics:            rfc2544ThroughputMetrics(),
		SuccessCriteria:    "Zero frame loss at the reported throughput rate",
		FailureExplanation: "Unable to achieve any rate without frame loss - check connectivity and configuration",
		Examples:           rfc2544ThroughputExamples(),
		Tips:               rfc2544ThroughputTips(),
		CommonIssues:       rfc2544ThroughputIssues(),
		RFCSection:         "Section 26.1",
		SeeAlso:            []string{TestTypeLatency, TestTypeFrameLoss, "y1564_config"},
	}
}

func rfc2544ThroughputParameters() []Parameter {
	return []Parameter{
		{
			Name:       LabelFrameSizes,
			Flag:       FlagFrameSizes,
			Type:       "comma-separated integers (bytes)",
			Default:    DefaultFrameSizes,
			Required:   false,
			TechDesc:   "Ethernet frame sizes to test, as specified in RFC 2544",
			LaymanDesc: "Different packet sizes to try - small packets stress the network differently than large ones",
			Example:    ExampleFrameSizes,
		},
		{
			Name:       "Trial Duration",
			Flag:       FlagDuration,
			Type:       "integer (seconds)",
			Default:    "60",
			Required:   false,
			TechDesc:   "Duration of each trial iteration in the binary search",
			LaymanDesc: "How long to run each speed test - longer is more accurate but takes more time",
			Example:    "--duration 30",
		},
		{
			Name:       "Resolution",
			Flag:       "--resolution",
			Type:       "float (percentage)",
			Default:    "0.1",
			Required:   false,
			TechDesc:   "Minimum step size for binary search convergence",
			LaymanDesc: "How precisely to find the maximum rate - smaller is more precise",
			Example:    "--resolution 0.5",
		},
		{
			Name:       "Loss Tolerance",
			Flag:       "--loss-tolerance",
			Type:       "float (percentage)",
			Default:    "0.0",
			Required:   false,
			TechDesc:   "Maximum acceptable frame loss rate (0.0 = zero loss required)",
			LaymanDesc: "How much packet loss is acceptable - usually zero for this test",
			Example:    "--loss-tolerance 0.001",
		},
	}
}

func rfc2544ThroughputMetrics() []Metric {
	return []Metric{
		{
			Name:       "Max Rate",
			Unit:       "% of line rate",
			GoodRange:  ">95% is excellent, >80% is acceptable",
			BadMeaning: "Below 80% indicates a bottleneck or configuration issue",
		},
		{
			Name:       "Throughput",
			Unit:       "Mbps or Gbps",
			GoodRange:  "Close to rated interface speed",
			BadMeaning: "Significantly below rated speed indicates problem",
		},
		{
			Name:       "Frame Loss",
			Unit:       "percentage",
			GoodRange:  "0.000%",
			BadMeaning: "Any loss at the reported rate indicates test instability",
		},
	}
}

func rfc2544ThroughputExamples() []Example {
	return []Example{
		{
			Desc:    "Basic throughput test on eth0",
			Command: "stem test -i eth0 -t throughput",
			Output: `Frame Size  Max Rate    Throughput
64 bytes    98.5%       985 Mbps
1518 bytes  99.2%       992 Mbps`,
		},
		{
			Desc:    "Quick test with fewer frame sizes",
			Command: "stem test -i eth0 -t throughput --frame-sizes 64,1518 --duration 30",
			Output:  "Completed in 2 minutes",
		},
		{
			Desc:    "High-precision test",
			Command: "stem test -i eth0 -t throughput --resolution 0.01 --duration 120",
			Output:  "Results accurate to 0.01%",
		},
	}
}

func rfc2544ThroughputTips() []string {
	return []string{
		"Run multiple iterations and average results for production validation",
		"Test during low-traffic periods for most accurate baseline",
		"Compare results across different times of day to detect congestion patterns",
		"Use the same frame sizes for before/after comparisons",
		"Small frames (64 bytes) stress packet processing; large frames test raw bandwidth",
	}
}

func rfc2544ThroughputIssues() []Issue {
	return []Issue{
		{
			Problem:  "Test shows 0% throughput",
			Cause:    "Interface not connected or wrong interface specified",
			Solution: "Verify cable connection and interface name with 'ip link show'",
		},
		{
			Problem:  "Results vary significantly between runs",
			Cause:    "Other traffic on the network or DUT instability",
			Solution: "Test during maintenance window or isolate test path",
		},
		{
			Problem:  "Low throughput on small packets (64 bytes)",
			Cause:    "Normal - small packets require more CPU per bit transferred",
			Solution: "This is expected; focus on packet rate (pps) for small frames",
		},
	}
}

func rfc2544Latency() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       TestTypeLatency,
			Name:     "Latency Test",
			Standard: "RFC 2544 Section 26.2",
			Category: StandardRFC2544,
		},
		rfc2544LatencyDescriptions(),
		rfc2544LatencyUsage(),
		rfc2544LatencyDetails(),
	)
}

func rfc2544LatencyDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Measures round-trip delay time for packets at various throughput levels.",
		TechDesc: `The latency test measures the time required for a frame to travel from the
	originating device through the DUT and back. This is performed at the throughput rate
	determined by the throughput test (or a specified rate) to measure latency under realistic
	load conditions.

	Latency is measured by inserting timestamp information in test frames and calculating
	the difference between transmission and reception times. The test reports minimum,
	maximum, and average latency values, plus standard deviation for jitter analysis.

	Per RFC 2544, latency is defined as the time interval starting when the last bit of
	the input frame reaches the input port and ending when the first bit of the output
	frame is seen on the output port.`,
		LaymanDesc: `This test measures "lag" - how long it takes for a message to get from
	point A to point B and back.

	Think of it like measuring how long it takes to:
	1. Send a letter
	2. Have someone receive it
	3. Have them send it back
	4. Receive the reply

	Lower numbers are better:
	• Under 1ms: Excellent (good for video calls, gaming, trading)
	• 1-10ms: Good for most applications
	• 10-50ms: Acceptable for general use
	• Over 50ms: May cause noticeable delays

	This test also measures "jitter" - how consistent the delay is. High jitter means
	the delay keeps changing, which can cause choppy video or audio.`,
	}
}

func rfc2544LatencyUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• Validating low-latency network requirements
	• VoIP and video conferencing quality assurance
	• Financial trading infrastructure validation
	• Gaming network performance testing
	• Real-time control system networks
	• Comparing network paths for latency-sensitive applications`,
		WhenNotToUse: `• If you only need bandwidth measurements (use Throughput test)
	• For packet loss analysis at various rates (use Frame Loss test)
	• For precise one-way delay (use Y.1731 Frame Delay)`,
	}
}

func rfc2544LatencyDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         rfc2544LatencyParameters(),
		Metrics:            rfc2544LatencyMetrics(),
		SuccessCriteria:    "Latency within acceptable range for intended application",
		FailureExplanation: "Network may not be suitable for latency-sensitive applications",
		Examples:           rfc2544LatencyExamples(),
		Tips:               rfc2544LatencyTips(),
		CommonIssues:       rfc2544LatencyIssues(),
		RFCSection:         "Section 26.2",
		SeeAlso:            []string{TestTypeThroughput, "frame_delay", "y1564_performance"},
	}
}

func rfc2544LatencyParameters() []Parameter {
	return []Parameter{
		{
			Name:       LabelFrameSizes,
			Flag:       FlagFrameSizes,
			Type:       "comma-separated integers (bytes)",
			Default:    DefaultFrameSizes,
			Required:   false,
			TechDesc:   "Frame sizes at which to measure latency",
			LaymanDesc: "Packet sizes to test - larger packets may have slightly higher latency",
			Example:    ExampleFrameSizes,
		},
		{
			Name:       "Rate",
			Flag:       "--rate",
			Type:       "float (percentage) or 'auto'",
			Default:    ValueAuto,
			Required:   false,
			TechDesc:   "Rate at which to measure latency (auto uses throughput test result)",
			LaymanDesc: "Network speed during test - 'auto' uses maximum lossless rate",
			Example:    "--rate 80",
		},
		{
			Name:       LabelDuration,
			Flag:       FlagDuration,
			Type:       "integer (seconds)",
			Default:    "120",
			Required:   false,
			TechDesc:   "Test duration for statistical accuracy",
			LaymanDesc: "How long to collect measurements - longer is more accurate",
			Example:    ExampleDuration60,
		},
		{
			Name:       "Sample Count",
			Flag:       "--samples",
			Type:       "integer",
			Default:    "20",
			Required:   false,
			TechDesc:   "Number of latency samples to collect",
			LaymanDesc: "How many individual measurements to take",
			Example:    "--samples 100",
		},
	}
}

func rfc2544LatencyMetrics() []Metric {
	return []Metric{
		{
			Name:       "Average Latency",
			Unit:       "microseconds (µs)",
			GoodRange:  "<1000µs (1ms) for most applications",
			BadMeaning: "High latency indicates network congestion or distance",
		},
		{
			Name:       "Minimum Latency",
			Unit:       "microseconds (µs)",
			GoodRange:  "Close to average indicates stable network",
			BadMeaning: "Much lower than average suggests variable queuing",
		},
		{
			Name:       "Maximum Latency",
			Unit:       "microseconds (µs)",
			GoodRange:  "Within 2x of average",
			BadMeaning: "Spikes indicate intermittent congestion",
		},
		{
			Name:       "Jitter (Std Dev)",
			Unit:       "microseconds (µs)",
			GoodRange:  "<100µs for voice/video",
			BadMeaning: "High jitter causes quality issues in real-time apps",
		},
	}
}

func rfc2544LatencyExamples() []Example {
	return []Example{
		{
			Desc:    "Basic latency test",
			Command: "stem test -i eth0 -t latency",
			Output:  "Avg: 125µs, Min: 98µs, Max: 245µs, Jitter: 23µs",
		},
		{
			Desc:    "Latency at specific rate",
			Command: "stem test -i eth0 -t latency --rate 50",
			Output:  "Latency measured at 50% line rate",
		},
	}
}

func rfc2544LatencyTips() []string {
	return []string{
		"Test at multiple rates to understand how latency changes with load",
		"Store-and-forward switches add more latency than cut-through switches",
		"Each network hop typically adds 10-100µs of latency",
		"Compare against baseline measurements to detect degradation",
	}
}

func rfc2544LatencyIssues() []Issue {
	return []Issue{
		{
			Problem:  "High latency spikes (max >> average)",
			Cause:    "Buffer bloat or periodic congestion",
			Solution: "Check for competing traffic or enable QoS",
		},
		{
			Problem:  "Latency increases with frame size",
			Cause:    "Normal serialization delay",
			Solution: "This is expected; larger frames take longer to transmit",
		},
	}
}

func rfc2544FrameLoss() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       TestTypeFrameLoss,
			Name:     "Frame Loss Rate Test",
			Standard: "RFC 2544 Section 26.3",
			Category: StandardRFC2544,
		},
		rfc2544FrameLossDescriptions(),
		rfc2544FrameLossUsage(),
		rfc2544FrameLossDetails(),
	)
}

func rfc2544FrameLossDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Measures what percentage of packets are lost at different network speeds.",
		TechDesc: `The frame loss rate test determines the percentage of frames that are not
	forwarded by the DUT at various offered loads. The test starts at 100% of the theoretical
	maximum rate and decreases in steps (typically 10% or configurable) until zero frame
	loss is achieved.

	This characterizes the DUT's behavior under overload conditions and identifies the
	"knee" of the performance curve where frame loss begins to occur. Results are typically
	presented as a graph showing frame loss percentage vs. offered load.

	Unlike the throughput test which finds one point (max lossless rate), this test maps
	the entire performance curve.`,
		LaymanDesc: `This test answers the question: "How many packets get lost as I push
	more traffic through the network?"

	Imagine a highway:
	• At low traffic, all cars get through (0% loss)
	• As traffic increases, some exits get backed up
	• Eventually, cars start missing their exits (packet loss)

	This test creates a "stress curve" showing:
	• At what speed does loss start happening?
	• How much loss at each speed level?
	• How does the network behave when overloaded?

	This helps you understand network behavior during peak usage and plan capacity.`,
	}
}

func rfc2544FrameLossUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• Understanding network behavior under overload
	• Capacity planning and upgrade justification
	• Comparing equipment performance characteristics
	• Identifying congestion points
	• Quality validation for bulk data transfers`,
		WhenNotToUse: `• For finding maximum lossless rate (use Throughput test)
	• For latency analysis (use Latency test)
	• For service activation (use Y.1564)`,
	}
}

func rfc2544FrameLossDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         rfc2544FrameLossParameters(),
		Metrics:            rfc2544FrameLossMetrics(),
		SuccessCriteria:    "Zero loss at planned operating rate",
		FailureExplanation: "Network cannot sustain planned traffic levels",
		Examples:           rfc2544FrameLossExamples(),
		Tips:               rfc2544FrameLossTips(),
		CommonIssues:       rfc2544FrameLossIssues(),
		RFCSection:         "Section 26.3",
		SeeAlso:            []string{TestTypeThroughput, "y1731_frame_loss"},
	}
}

func rfc2544FrameLossParameters() []Parameter {
	return []Parameter{
		{
			Name:       LabelFrameSizes,
			Flag:       FlagFrameSizes,
			Type:       "comma-separated integers (bytes)",
			Default:    DefaultFrameSizes,
			Required:   false,
			TechDesc:   "Frame sizes at which to measure loss rate",
			LaymanDesc: "Packet sizes to test",
			Example:    ExampleFrameSizes,
		},
		{
			Name:       "Start Rate",
			Flag:       "--start-rate",
			Type:       "float (percentage)",
			Default:    "100",
			Required:   false,
			TechDesc:   "Initial offered load as percentage of line rate",
			LaymanDesc: "Starting speed - usually 100% (maximum)",
			Example:    "--start-rate 100",
		},
		{
			Name:       "Step Size",
			Flag:       "--step",
			Type:       "float (percentage)",
			Default:    "10",
			Required:   false,
			TechDesc:   "Decrease step between iterations",
			LaymanDesc: "How much to slow down between tests",
			Example:    "--step 5",
		},
		{
			Name:       LabelDuration,
			Flag:       FlagDuration,
			Type:       "integer (seconds)",
			Default:    "60",
			Required:   false,
			TechDesc:   "Duration of each rate iteration",
			LaymanDesc: "How long to test at each speed",
			Example:    "--duration 30",
		},
	}
}

func rfc2544FrameLossMetrics() []Metric {
	return []Metric{
		{
			Name:       "Loss Rate",
			Unit:       "percentage",
			GoodRange:  "0% at operating rate",
			BadMeaning: "Any loss at normal operating rates is problematic",
		},
		{
			Name:       "Loss Start Point",
			Unit:       "% of line rate",
			GoodRange:  ">90% is good",
			BadMeaning: "Loss starting below 80% indicates serious issues",
		},
	}
}

func rfc2544FrameLossExamples() []Example {
	return []Example{
		{
			Desc:    "Frame loss rate sweep",
			Command: "stem test -i eth0 -t frame_loss",
			Output:  "100%: 2.3% loss, 90%: 0.1% loss, 80%: 0% loss",
		},
		{
			Desc:    "Fine-grained analysis",
			Command: "stem test -i eth0 -t frame_loss --step 5 --start-rate 100",
			Output:  "Detailed curve with 5% increments",
		},
	}
}

func rfc2544FrameLossTips() []string {
	return []string{
		"Use results to set traffic engineering thresholds",
		"Compare with throughput test to validate consistency",
		"High loss at low rates indicates configuration problems",
	}
}

func rfc2544FrameLossIssues() []Issue {
	return []Issue{
		{
			Problem:  "Loss at all rates including 10%",
			Cause:    "Major connectivity or configuration issue",
			Solution: "Check physical layer, duplex settings, VLAN config",
		},
	}
}

func rfc2544BackToBack() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       "back_to_back",
			Name:     "Back-to-Back Frames Test",
			Standard: "RFC 2544 Section 26.4",
			Category: StandardRFC2544,
		},
		rfc2544BackToBackDescriptions(),
		rfc2544BackToBackUsage(),
		rfc2544BackToBackDetails(),
	)
}

func rfc2544BackToBackDescriptions() testHelpDescriptions {
	return testHelpDescriptions{
		Summary: "Measures how many packets can be sent in a burst without any loss.",
		TechDesc: `The back-to-back frames test measures the maximum number of frames that
can be transmitted at the minimum legal inter-frame gap (IFG) before a frame is lost.
This characterizes the DUT's buffer capacity and its ability to handle traffic bursts.

The test sends frames back-to-back at minimum IFG (96 bit times for Ethernet) and
measures how many consecutive frames can be successfully forwarded. The test starts
with a small burst and increases until frame loss occurs.

This is critical for understanding behavior with bursty traffic patterns typical of
many real-world applications.`,
		LaymanDesc: `This test measures "burst capacity" - how much data can be sent all at
once without overwhelming the network.

Real network traffic comes in bursts:
• You click a link → burst of web page data
• Video starts → burst of buffered video
• File transfer → continuous burst of data

This test finds out:
• How big a burst can the network handle?
• At what point does it start dropping packets?
• Is there enough buffer space for your applications?

Higher numbers are better - they mean the network can handle bigger "waves" of data.`,
	}
}

func rfc2544BackToBackUsage() testHelpUsage {
	return testHelpUsage{
		WhenToUse: `• Buffer sizing validation
• Burst traffic application assessment
• Quality of Service (QoS) tuning
• Comparing switch buffer architectures
• Video streaming infrastructure validation`,
		WhenNotToUse: `• For sustained throughput (use Throughput test)
• For latency requirements (use Latency test)`,
	}
}

func rfc2544BackToBackDetails() testHelpDetails {
	return testHelpDetails{
		Parameters:         rfc2544BackToBackParameters(),
		Metrics:            rfc2544BackToBackMetrics(),
		SuccessCriteria:    "Burst capacity meets application requirements",
		FailureExplanation: "May experience drops during traffic bursts",
		Examples:           rfc2544BackToBackExamples(),
		Tips:               rfc2544BackToBackTips(),
		CommonIssues:       nil,
		RFCSection:         "Section 26.4",
		SeeAlso:            []string{TestTypeThroughput, "congestion"},
	}
}

func rfc2544BackToBackParameters() []Parameter {
	return []Parameter{
		{
			Name:       LabelFrameSizes,
			Flag:       FlagFrameSizes,
			Type:       "comma-separated integers (bytes)",
			Default:    DefaultFrameSizes,
			Required:   false,
			TechDesc:   "Frame sizes at which to measure burst capacity",
			LaymanDesc: "Packet sizes to test",
			Example:    "--frame-sizes 64,1518",
		},
		{
			Name:       "Trials",
			Flag:       "--trials",
			Type:       "integer",
			Default:    "50",
			Required:   false,
			TechDesc:   "Number of trials to average",
			LaymanDesc: "How many times to repeat for accuracy",
			Example:    "--trials 100",
		},
	}
}

func rfc2544BackToBackMetrics() []Metric {
	return []Metric{
		{
			Name:       "Burst Size",
			Unit:       "frames",
			GoodRange:  "Depends on application requirements",
			BadMeaning: "Small burst size may cause issues with bursty traffic",
		},
		{
			Name:       "Buffer Equivalent",
			Unit:       "bytes or KB",
			GoodRange:  ">1MB for typical switches",
			BadMeaning: "Low buffer may cause congestion drops",
		},
	}
}

func rfc2544BackToBackExamples() []Example {
	return []Example{
		{
			Desc:    "Back-to-back test",
			Command: "stem test -i eth0 -t back_to_back",
			Output:  "Max burst: 2048 frames (3.1 MB buffer equivalent)",
		},
	}
}

func rfc2544BackToBackTips() []string {
	return []string{
		"Results indicate effective buffer size of the DUT",
		"Compare across different frame sizes",
		"Important for networks carrying video or bulk transfers",
	}
}

func rfc2544SystemRecovery() TestHelp {
	return TestHelp{
		ID:       "system_recovery",
		Name:     "System Recovery Test",
		Standard: "RFC 2544 Section 26.5",
		Category: StandardRFC2544,

		Summary: "Measures how quickly the network recovers after being overloaded.",

		TechDesc: `The system recovery test measures how long a DUT takes to recover from
an overload condition. The test first overloads the DUT by transmitting at a rate
110% of the maximum throughput rate, then immediately reduces to 50% of the maximum
rate and measures the time until the DUT resumes normal forwarding.

This characterizes the DUT's ability to recover from congestion events and return
to normal operation. Long recovery times can impact application performance even
after the overload condition has passed.`,

		LaymanDesc: `After your network gets overwhelmed with too much traffic, how long
does it take to get back to normal?

Think of it like a busy restaurant:
• During a rush, orders get backed up
• When the rush ends, how quickly do things return to normal?
• Does the kitchen clear the backlog quickly, or stay chaotic?

This test:
1. Deliberately overwhelms the network
2. Then reduces traffic to normal levels
3. Measures how long until everything works smoothly again

Fast recovery (under 1 second) is good. Slow recovery means problems linger after
traffic spikes.`,

		WhenToUse: `• Mission-critical network validation
• Understanding DUT behavior after congestion
• Comparing equipment resilience
• Planning for traffic spike scenarios`,

		WhenNotToUse: `• For normal operating conditions (use Throughput test)
• For sustained overload behavior (use Frame Loss test)`,

		Parameters: []Parameter{
			{
				Name:       "Overload Rate",
				Flag:       "--overload-rate",
				Type:       "float (percentage)",
				Default:    "110",
				Required:   false,
				TechDesc:   "Rate used to overload the DUT (% of max throughput)",
				LaymanDesc: "How much to overload the network - 110% is standard",
				Example:    "--overload-rate 120",
			},
			{
				Name:       "Recovery Rate",
				Flag:       "--recovery-rate",
				Type:       "float (percentage)",
				Default:    "50",
				Required:   false,
				TechDesc:   "Rate at which to measure recovery (% of max throughput)",
				LaymanDesc: "Normal traffic level for recovery measurement",
				Example:    "--recovery-rate 60",
			},
		},

		Metrics: []Metric{
			{
				Name:       "Recovery Time",
				Unit:       "milliseconds",
				GoodRange:  "<1000ms",
				BadMeaning: "Long recovery impacts user experience",
			},
		},

		SuccessCriteria:    "Recovery time within acceptable limits for application",
		FailureExplanation: "DUT may cause extended service impact after congestion events",

		Examples: []Example{
			{
				Desc:    "System recovery test",
				Command: "stem test -i eth0 -t system_recovery",
				Output:  "Recovery time: 245ms",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "Section 26.5",
		SeeAlso:    []string{TestTypeThroughput, "reset"},
	}
}

func rfc2544Reset() TestHelp {
	return TestHelp{
		ID:       "reset",
		Name:     "Reset Test",
		Standard: "RFC 2544 Section 26.6",
		Category: StandardRFC2544,

		Summary: "Measures how long the device takes to recover from a hardware or software reset.",

		TechDesc: `The reset test measures the time required for a DUT to recover from
hardware or software reset events. The test establishes a forwarding baseline,
triggers a reset (power cycle, software reboot, or failover), and measures the
time until the DUT resumes forwarding at the baseline rate.

This is important for understanding service availability during planned or
unplanned restart events.`,

		LaymanDesc: `When network equipment restarts, how long is the network down?

Like rebooting your computer:
• How long until it's usable again?
• Does it come back at full speed immediately?
• Or does it take time to "warm up"?

This matters because:
• Planned maintenance: How long will users be affected?
• Unexpected crashes: How quickly does service restore?
• Software updates: What's the real downtime?

Lower reset times mean less disruption during maintenance or failures.`,

		WhenToUse: `• Maintenance window planning
• High-availability architecture design
• Equipment comparison for resilience
• SLA validation for uptime requirements`,

		WhenNotToUse: `• For normal performance testing (use Throughput test)
• For failover testing in HA pairs (use specific HA test)`,

		Parameters: []Parameter{
			{
				Name:       "Reset Type",
				Flag:       "--reset-type",
				Type:       "string",
				Default:    "software",
				Required:   false,
				TechDesc:   "Type of reset: 'software', 'power', or 'failover'",
				LaymanDesc: "How to restart: software reboot, power cycle, or switch to backup",
				Example:    "--reset-type power",
			},
		},

		Metrics: []Metric{
			{
				Name:       "Reset Time",
				Unit:       "seconds",
				GoodRange:  "<60s for most equipment",
				BadMeaning: "Long reset times impact availability SLAs",
			},
		},

		SuccessCriteria:    "Reset time meets availability requirements",
		FailureExplanation: "Equipment restart takes too long for required uptime SLA",

		Examples: []Example{
			{
				Desc:    "Software reset test",
				Command: "stem test -i eth0 -t reset --reset-type software",
				Output:  "Reset time: 45 seconds",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "Section 26.6",
		SeeAlso:    []string{"system_recovery"},
	}
}
