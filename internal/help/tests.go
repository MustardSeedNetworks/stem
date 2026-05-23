/*
 * Seed Test Suite - Test Documentation
 *
 * Comprehensive documentation for all 27 test types across 7 categories.
 */

package help

// GetAllTests returns help content for all tests.
func GetAllTests() map[string]TestHelp {
	return map[string]TestHelp{
		// RFC 2544 Tests
		"throughput":      rfc2544Throughput(),
		"latency":         rfc2544Latency(),
		"frame_loss":      rfc2544FrameLoss(),
		"back_to_back":    rfc2544BackToBack(),
		"system_recovery": rfc2544SystemRecovery(),
		"reset":           rfc2544Reset(),

		// Y.1564 Tests
		"y1564_config":      y1564Config(),
		"y1564_performance": y1564Performance(),
		"y1564_full":        y1564Full(),

		// RFC 2889 Tests
		"forwarding":    rfc2889Forwarding(),
		"address_cache": rfc2889AddressCache(),
		"learning_rate": rfc2889LearningRate(),
		"broadcast":     rfc2889Broadcast(),
		"congestion":    rfc2889Congestion(),

		// RFC 6349 Tests
		"tcp_throughput": rfc6349TCPThroughput(),
		"path_analysis":  rfc6349PathAnalysis(),

		// Y.1731 Tests
		"frame_delay":      y1731FrameDelay(),
		"y1731_frame_loss": y1731FrameLoss(),
		"synthetic_loss":   y1731SyntheticLoss(),
		"loopback":         y1731Loopback(),

		// MEF Tests
		"mef_config":      mefConfig(),
		"mef_performance": mefPerformance(),
		"mef_full":        mefFull(),

		// TSN Tests
		"gate_timing":       tsnGateTiming(),
		"traffic_isolation": tsnTrafficIsolation(),
		"scheduled_latency": tsnScheduledLatency(),
		"tsn_full":          tsnFull(),
	}
}

// ============================================================================
// RFC 2544 Tests - Benchmarking Methodology for Network Interconnect Devices.
// ============================================================================

func rfc2544Throughput() TestHelp {
	return buildTestHelp(
		testHelpMeta{
			ID:       "throughput",
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
		SeeAlso:            []string{"latency", "frame_loss", "y1564_config"},
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
			ID:       "latency",
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
		SeeAlso:            []string{"throughput", "frame_delay", "y1564_performance"},
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
			Default:    "auto",
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
			ID:       "frame_loss",
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
		SeeAlso:            []string{"throughput", "y1731_frame_loss"},
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
		SeeAlso:            []string{"throughput", "congestion"},
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
		SeeAlso:    []string{"throughput", "reset"},
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
		SeeAlso:            []string{"y1564_performance", "throughput", "mef_config"},
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

// ============================================================================
// RFC 2889 Tests - Benchmarking Methodology for LAN Switching Devices.
// ============================================================================

func rfc2889Forwarding() TestHelp {
	return TestHelp{
		ID:       "forwarding",
		Name:     "Forwarding Rate Test",
		Standard: "RFC 2889 Section 5.2",
		Category: StandardRFC2889,

		Summary: "Measures how fast a switch can move packets between ports.",

		TechDesc: `The forwarding rate test measures the maximum rate at which a switch
can forward frames from multiple input ports to multiple output ports. Unlike RFC 2544
throughput which tests a single flow, this test exercises the switch fabric with
multiple concurrent flows to characterize aggregate forwarding capacity.

The test typically uses a full mesh pattern where each port sends to every other port,
stressing the switch's backplane and forwarding ASIC capabilities.`,

		LaymanDesc: `How fast can your switch shuffle packets between all its ports at once?

A switch is like a mail sorting facility:
• RFC 2544 tests: One mail truck arriving, one leaving
• This test: All trucks arriving and leaving simultaneously

This reveals:
• Can the switch handle traffic on all ports at once?
• Is there enough internal bandwidth (backplane)?
• Will performance drop when many ports are busy?

Important for environments with many active connections simultaneously.`,

		WhenToUse: `• Evaluating switch aggregate capacity
• Data center switch validation
• Core switch performance testing
• Comparing switch architectures`,

		WhenNotToUse: `• Single port-pair testing (use RFC 2544)
• Service provider testing (use Y.1564)`,

		Parameters: []Parameter{
			{
				Name:       "Ports",
				Flag:       "--ports",
				Type:       "comma-separated interfaces",
				Default:    "all available",
				Required:   false,
				TechDesc:   "Interfaces to include in the test",
				LaymanDesc: "Which switch ports to test",
				Example:    "--ports eth0,eth1,eth2,eth3",
			},
			{
				Name:       "Pattern",
				Flag:       "--pattern",
				Type:       "string",
				Default:    "mesh",
				Required:   false,
				TechDesc:   "Traffic pattern: mesh, pair, or custom",
				LaymanDesc: "How traffic flows between ports",
				Example:    "--pattern mesh",
			},
		},

		Metrics: []Metric{
			{
				Name:       "Aggregate Rate",
				Unit:       "Mpps (million packets per second)",
				GoodRange:  "Near theoretical switch capacity",
				BadMeaning: "Switch fabric bottleneck",
			},
		},

		SuccessCriteria:    "Aggregate throughput meets switch specifications",
		FailureExplanation: "Switch may not handle full load scenarios",

		Examples: []Example{
			{
				Desc:    "4-port switch test",
				Command: "stem test -t forwarding --ports eth0,eth1,eth2,eth3",
				Output:  "Aggregate: 5.95 Mpps across 4 ports",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "Section 5.2",
		SeeAlso:    []string{"address_cache", "throughput"},
	}
}

func rfc2889AddressCache() TestHelp {
	return TestHelp{
		ID:       "address_cache",
		Name:     "Address Caching Capacity Test",
		Standard: "RFC 2889 Section 5.5",
		Category: StandardRFC2889,

		Summary: "Determines how many MAC addresses a switch can remember.",

		TechDesc: `The address caching capacity test determines the maximum number of MAC
addresses a switch can store in its forwarding table while maintaining forwarding
performance. The test progressively increases the number of source MAC addresses
until the switch can no longer learn new addresses or forwarding performance degrades.

This is critical for large networks where switches must track many connected devices.`,

		LaymanDesc: `Every device on a network has a unique address (MAC address). A switch
needs to remember these addresses to send traffic to the right place.

This test answers:
• How many devices can this switch keep track of?
• What happens when the limit is reached?
• Does performance drop when the table is full?

Think of it like a phonebook:
• Small switch: Remembers 8,000 addresses
• Large switch: Remembers 128,000+ addresses
• When full: May flood traffic to all ports or drop packets`,

		WhenToUse: `• Large campus network planning
• Data center switch validation
• Network segmentation planning
• Virtualization environments (many VMs = many MACs)`,

		WhenNotToUse: `• Small networks with few devices
• Router testing (routers use IP, not MAC tables)`,

		Parameters: []Parameter{
			{
				Name:       "Start Count",
				Flag:       "--start-count",
				Type:       "integer",
				Default:    "1000",
				Required:   false,
				TechDesc:   "Initial number of MAC addresses",
				LaymanDesc: "Starting number of fake devices to simulate",
				Example:    "--start-count 5000",
			},
			{
				Name:       "Step",
				Flag:       "--step",
				Type:       "integer",
				Default:    "1000",
				Required:   false,
				TechDesc:   "Increment between iterations",
				LaymanDesc: "How many more to add each round",
				Example:    "--step 2000",
			},
		},

		Metrics: []Metric{
			{
				Name:       "Max Addresses",
				Unit:       "count",
				GoodRange:  "Matches switch specifications",
				BadMeaning: "Below spec indicates software/hardware limitation",
			},
		},

		SuccessCriteria:    "Address capacity meets deployment requirements",
		FailureExplanation: "May need a switch with larger MAC table",

		Examples: []Example{
			{
				Desc:    "Test MAC table capacity",
				Command: "stem test -i eth0 -t address_cache",
				Output:  "MAC table capacity: 16,384 addresses",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "Section 5.5",
		SeeAlso:    []string{"learning_rate", "forwarding"},
	}
}

func rfc2889LearningRate() TestHelp {
	return TestHelp{
		ID:       "learning_rate",
		Name:     "Address Learning Rate Test",
		Standard: "RFC 2889 Section 5.6",
		Category: StandardRFC2889,

		Summary: "Measures how fast a switch can learn new device addresses.",

		TechDesc: `The address learning rate test measures how quickly a switch can
populate its MAC address table with new entries. This is tested by sending frames
with new source MAC addresses at increasing rates and determining the maximum
rate at which the switch can learn addresses without missing any.

Important for environments where devices frequently join or leave the network.`,

		LaymanDesc: `When a new device connects to your network, how fast can the switch
"register" it?

In dynamic environments like:
• WiFi networks where people come and go
• Virtual machines spinning up and down
• Conference rooms where laptops connect/disconnect

A slow learning rate means:
• Brief connectivity issues for new devices
• Traffic flooding while the switch figures things out
• Potential security implications

Faster learning = smoother experience for users joining the network.`,

		WhenToUse: `• Highly dynamic environments
• VM/container infrastructure
• Wireless network deployments
• Guest network validation`,

		WhenNotToUse: `• Static networks with fixed devices
• Small offices with few devices`,

		Parameters: nil,

		Metrics: []Metric{
			{
				Name:       "Learning Rate",
				Unit:       "addresses per second",
				GoodRange:  "1000+ for enterprise switches",
				BadMeaning: "May cause delays for new device connectivity",
			},
		},

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "Test learning rate",
				Command: "stem test -i eth0 -t learning_rate",
				Output:  "Learning rate: 5,000 addresses/second",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "Section 5.6",
		SeeAlso:    []string{"address_cache"},
	}
}

func rfc2889Broadcast() TestHelp {
	return TestHelp{
		ID:       "broadcast",
		Name:     "Broadcast Frame Handling Test",
		Standard: "RFC 2889 Section 5.7",
		Category: StandardRFC2889,

		Summary: "Tests how the switch handles broadcast traffic.",

		TechDesc: `The broadcast frame handling test measures how a switch processes
broadcast frames that must be forwarded to all ports. This tests both the forwarding
capacity for broadcast traffic and the impact on unicast forwarding performance
when broadcast load increases.

Excessive broadcast traffic can overwhelm switches and end devices.`,

		LaymanDesc: `Broadcast messages go to ALL devices on the network. How does your
switch handle this?

Examples of broadcasts:
• "Who has IP address 192.168.1.1?" (ARP request)
• "I'm a printer, anyone want to print?" (discovery)
• "Time sync to all devices" (NTP broadcast)

Too much broadcast traffic can:
• Slow down the entire network
• Overwhelm computers with unwanted messages
• Indicate a network problem ("broadcast storm")

This test checks if your switch can handle normal broadcast levels without problems.`,

		WhenToUse: `• Networks with many broadcast-heavy protocols
• VLAN sizing decisions
• Broadcast storm recovery testing`,

		WhenNotToUse: `• Isolated point-to-point testing`,

		Parameters: nil,

		Metrics: []Metric{
			{
				Name:       "Broadcast Handling Rate",
				Unit:       "frames per second",
				GoodRange:  "Equal to unicast forwarding rate",
				BadMeaning: "Broadcast processing is limited",
			},
			{
				Name:       "Unicast Impact",
				Unit:       "percentage degradation",
				GoodRange:  "<5% impact on unicast",
				BadMeaning: "Broadcasts affecting normal traffic",
			},
		},

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "Broadcast handling test",
				Command: "stem test -i eth0 -t broadcast",
				Output:  "Broadcast rate: 148,810 fps, Unicast impact: 2%",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "Section 5.7",
		SeeAlso:    []string{"forwarding", "congestion"},
	}
}

func rfc2889Congestion() TestHelp {
	return TestHelp{
		ID:       "congestion",
		Name:     "Congestion Control Test",
		Standard: "RFC 2889 Section 5.8",
		Category: StandardRFC2889,

		Summary: "Tests switch behavior when ports are oversubscribed.",

		TechDesc: `The congestion control test measures how a switch behaves when the
aggregate input rate exceeds the output port capacity. This characterizes the
switch's queuing and dropping behavior, buffer management, and fairness across
input ports during congestion.

This reveals how the switch allocates resources when demand exceeds capacity.`,

		LaymanDesc: `What happens when too much traffic tries to go to the same place?

Imagine a funnel:
• 4 liters pouring into a 1-liter opening
• Some water spills (packet loss)
• How the switch handles this "spill" matters

Good congestion handling:
• Fair distribution of bandwidth
• Predictable behavior
• Minimal impact on other traffic

Bad congestion handling:
• One source hogs all bandwidth
• Unpredictable performance
• Affects unrelated traffic

This test reveals your switch's personality under stress.`,

		WhenToUse: `• Server farm switch validation
• Quality of Service tuning
• Understanding oversubscription effects`,

		WhenNotToUse: `• Non-blocking switch architectures
• When congestion shouldn't occur by design`,

		Parameters: nil,

		Metrics: []Metric{
			{
				Name:       "Head-of-Line Blocking",
				Unit:       "percentage",
				GoodRange:  "0% (no blocking)",
				BadMeaning: "Congestion affecting unrelated flows",
			},
			{
				Name:       "Fairness Index",
				Unit:       "0.0 - 1.0",
				GoodRange:  ">0.9 (fair distribution)",
				BadMeaning: "Uneven bandwidth allocation",
			},
		},

		SuccessCriteria:    "",
		FailureExplanation: "",

		Examples: []Example{
			{
				Desc:    "Congestion control test",
				Command: "stem test -t congestion --ports eth0,eth1,eth2,eth3 --output eth3",
				Output:  "HOL Blocking: 0%, Fairness: 0.98",
			},
		},

		Tips:         nil,
		CommonIssues: nil,

		RFCSection: "Section 5.8",
		SeeAlso:    []string{"forwarding", "back_to_back"},
	}
}

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
		SeeAlso:            []string{"path_analysis", "throughput"},
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
			Default:    "auto",
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
		SeeAlso:    []string{"tcp_throughput", "latency"},
	}
}

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
		SeeAlso:            []string{"latency", "y1731_frame_loss"},
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
		SeeAlso:    []string{"frame_loss", "frame_delay"},
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
		"rfc2544": {
			ID:       "rfc2544",
			Name:     StandardRFC2544,
			FullName: "Benchmarking Methodology for Network Interconnect Devices",
			Summary:  "The standard tests for measuring raw network equipment performance.",
			Description: `RFC 2544 defines benchmarking methodology for network devices.
These tests measure fundamental performance characteristics: throughput, latency,
frame loss, and burst handling. Use these tests for equipment validation and comparison.`,
			Tests:     []string{"throughput", "latency", "frame_loss", "back_to_back", "system_recovery", "reset"},
			WhenToUse: "Equipment benchmarking, performance validation, comparing vendors",
			Standard:  StandardRFC2544,
			SeeAlso:   []string{"y1564", "rfc2889"},
		},
		"y1564": {
			ID:       "y1564",
			Name:     StandardY1564,
			FullName: "Ethernet Service Activation Test Methodology",
			Summary:  "The carrier standard for turning up ethernet services.",
			Description: `ITU-T Y.1564 defines the methodology for activating and validating
carrier ethernet services. These tests verify that a service meets its SLA parameters
at progressive load levels and over extended duration.`,
			Tests:     []string{"y1564_config", "y1564_performance", "y1564_full"},
			WhenToUse: "Carrier service activation, SLA validation, service acceptance",
			Standard:  StandardITUY1564,
			SeeAlso:   []string{"mef", "rfc2544"},
		},
		"rfc2889": {
			ID:       "rfc2889",
			Name:     StandardRFC2889,
			FullName: "Benchmarking Methodology for LAN Switching Devices",
			Summary:  "Tests specifically for switch/bridge performance characteristics.",
			Description: `RFC 2889 extends RFC 2544 for testing LAN switches. These tests
measure switch-specific characteristics like forwarding rate across multiple ports,
MAC address table capacity, learning rate, and congestion handling.`,
			Tests:     []string{"forwarding", "address_cache", "learning_rate", "broadcast", "congestion"},
			WhenToUse: "Switch validation, data center planning, MAC table capacity verification",
			Standard:  StandardRFC2889,
			SeeAlso:   []string{"rfc2544"},
		},
		"rfc6349": {
			ID:       "rfc6349",
			Name:     StandardRFC6349,
			FullName: "Framework for TCP Throughput Testing",
			Summary:  "Tests that measure real TCP application performance.",
			Description: `RFC 6349 provides methodology for testing TCP throughput, which
represents actual application performance. These tests measure achievable TCP throughput
and help identify network factors affecting TCP performance.`,
			Tests:     []string{"tcp_throughput", "path_analysis"},
			WhenToUse: "Application performance testing, WAN optimization, TCP troubleshooting",
			Standard:  StandardRFC6349,
			SeeAlso:   []string{"rfc2544"},
		},
		"y1731": {
			ID:       "y1731",
			Name:     "Y.1731",
			FullName: "OAM Functions and Mechanisms for Ethernet Networks",
			Summary:  "Operations, Administration, and Maintenance for carrier ethernet.",
			Description: `ITU-T Y.1731 defines OAM functions for monitoring and maintaining
ethernet services. These tools provide in-service monitoring capabilities including
delay measurement, loss measurement, and connectivity verification.`,
			Tests:     []string{"frame_delay", "y1731_frame_loss", "synthetic_loss", "loopback"},
			WhenToUse: "Production monitoring, SLA verification, fault isolation",
			Standard:  StandardITUY1731,
			SeeAlso:   []string{"y1564", "mef"},
		},
		"mef": {
			ID:       "mef",
			Name:     StandardMEF,
			FullName: "Metro Ethernet Forum Service Tests",
			Summary:  "Industry standard tests for carrier ethernet services.",
			Description: `MEF (Metro Ethernet Forum) defines service specifications and
testing methodologies for carrier ethernet. These tests validate services against
MEF specifications including bandwidth profiles and Class of Service.`,
			Tests:     []string{"mef_config", "mef_performance", "mef_full"},
			WhenToUse: "MEF-certified service validation, multi-CoS testing, carrier acceptance",
			Standard:  "MEF 14/48",
			SeeAlso:   []string{"y1564", "y1731"},
		},
		"tsn": {
			ID:       "tsn",
			Name:     StandardTSN,
			FullName: "Time-Sensitive Networking",
			Summary:  "Tests for deterministic, time-critical industrial networks.",
			Description: `IEEE 802.1 Time-Sensitive Networking tests validate networks
requiring deterministic timing. These tests verify that time-aware shaping, traffic
isolation, and scheduled latency meet industrial automation requirements.`,
			Tests:     []string{"gate_timing", "traffic_isolation", "scheduled_latency", "tsn_full"},
			WhenToUse: "Industrial automation, automotive ethernet, deterministic networking",
			Standard:  "IEEE 802.1Qbv/Qbu",
			SeeAlso:   []string{"rfc2544"},
		},
	}
}
