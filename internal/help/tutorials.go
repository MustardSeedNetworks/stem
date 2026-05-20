/*
 * Seed Test Suite - Tutorials
 *
 * Step-by-step guides for common tasks.
 */

package help

// GetAllTutorials returns all available tutorials.
func GetAllTutorials() map[string]Tutorial {
	return map[string]Tutorial{
		"quickstart":   QuickstartTutorial(),
		"reflector":    ReflectorTutorial(),
		"rfc2544":      RFC2544Tutorial(),
		"y1564":        Y1564Tutorial(),
		"troubleshoot": TroubleshootTutorial(),
		"results":      ResultsTutorial(),
	}
}

// QuickstartTutorial - Getting started guide.
func QuickstartTutorial() Tutorial {
	return Tutorial{
		ID:          "quickstart",
		Title:       "Your First Test in 5 Minutes",
		Duration:    "5 min",
		Level:       "Beginner",
		Description: "Get up and running with your first network test",
		Steps: []TutorialStep{
			{
				Title: "Check Your Setup",
				Content: `Before running tests, make sure you have:
• Two network interfaces (one for testing, one for management)
• OR a remote reflector device
• Seed Test Suite installed and licensed

Check your interfaces:`,
				Command:  "ip link show",
				Expected: "Lists your network interfaces (eth0, enp3s0, etc.)",
				Tip:      "The interface for testing should be connected to the network segment you want to test",
			},
			{
				Title: "Start the Reflector",
				Content: `For testing, you need a reflector at the far end of your test path.
This can be on the same machine (loopback testing) or a remote device.

Start a local reflector:`,
				Command:  "stem reflect -i eth0",
				Expected: "Reflector started on eth0:3842",
				Tip:      "In production, run the reflector on a separate device at the remote end",
			},
			{
				Title: "Run Your First Test",
				Content: `Now let's run a simple throughput test. Open a new terminal
(keep the reflector running) and run:`,
				Command:  "stem test -i eth1 -t throughput --target localhost",
				Expected: "Testing throughput...\nMax throughput: 985 Mbps (98.5%)",
				Tip:      "Use --target with the IP of your remote reflector for real testing",
			},
			{
				Title: "View Results",
				Content: `Results are displayed immediately. For a permanent record,
save to a file:`,
				Command:  "stem test -i eth1 -t throughput --output results.json",
				Expected: "Results saved to results.json",
				Tip:      "Use .csv extension for spreadsheet-compatible output",
			},
			{
				Title: "Explore More Tests",
				Content: `You've completed your first test! Now explore other tests:

• stem help tests     - List all available tests
• stem help latency   - Learn about latency testing
• stem tutorial rfc2544 - Deep dive into RFC 2544`,
				Command:  "stem help tests",
				Expected: "",
				Tip:      "Start with RFC 2544 tests for basic network benchmarking",
			},
		},
	}
}

// ReflectorTutorial - Setting up packet reflection.
func ReflectorTutorial() Tutorial {
	return Tutorial{
		ID:          "reflector",
		Title:       "Setting Up Packet Reflection",
		Duration:    "10 min",
		Level:       "Beginner",
		Description: "Learn how to configure and optimize the packet reflector",
		Steps: []TutorialStep{
			{
				Title: "Understanding Reflection",
				Content: `The reflector is a key component for network testing. It receives
test packets and sends them back to the source, enabling round-trip measurements.

Why reflection matters:
• Measures real network path performance
• No need for special hardware at the remote end
• Supports all test types

You can run the reflector on:
• A dedicated Linux device
• A server at the remote site
• The same machine (for loopback testing)`,
				Command:  "",
				Expected: "",
				Tip:      "",
			},
			{
				Title:    "Basic Reflector Setup",
				Content:  `Start a basic reflector on your interface:`,
				Command:  "stem reflect -i eth0",
				Expected: "Reflector running on eth0:3842 (AF_PACKET mode)\nWeb UI: https://localhost:8444",
				Tip:      "The reflector auto-detects the best performance mode for your system",
			},
			{
				Title: "High-Performance Mode",
				Content: `For maximum performance, use AF_XDP or DPDK mode:

AF_XDP mode (requires kernel 5.x+):`,
				Command:  "stem reflect -i eth0 --mode af_xdp",
				Expected: "Reflector running in AF_XDP mode (~5-10 Mpps)",
				Tip:      "AF_XDP requires the interface to support XDP",
			},
			{
				Title: "Filtering Traffic",
				Content: `In production environments, you may want to filter which traffic
gets reflected. Use OUI filtering to only respond to specific devices:`,
				Command:  "stem reflect -i eth0 --filter-oui 00:c0:17",
				Expected: "Reflecting only packets from MAC addresses starting with 00:c0:17",
				Tip:      "This prevents the reflector from responding to unrelated traffic",
			},
			{
				Title: "Monitoring with Web UI",
				Content: `The reflector includes a web interface for monitoring:

• Real-time packet statistics
• Performance graphs
• Configuration status

Access it at:`,
				Command:  "# Open in browser: https://<reflector-ip>:8444",
				Expected: "Web interface showing live statistics",
				Tip:      "Use -w 0 to disable the web interface if not needed",
			},
			{
				Title: "Running as a Service",
				Content: `For production, run the reflector as a system service:

Create /etc/systemd/system/stem-reflector.service:

[Unit]
Description=Seed Test Suite Reflector
After=network.target

[Service]
ExecStart=/usr/local/bin/stem reflect -i eth0
Restart=always

[Install]
WantedBy=multi-user.target

Then enable and start:`,
				Command:  "sudo systemctl enable stem-reflector && sudo systemctl start stem-reflector",
				Expected: "Reflector running as a system service",
				Tip:      "Check status with: systemctl status stem-reflector",
			},
		},
	}
}

// RFC2544Tutorial - RFC 2544 deep dive.
func RFC2544Tutorial() Tutorial {
	return Tutorial{
		ID:          "rfc2544",
		Title:       "Understanding RFC 2544 Tests",
		Duration:    "15 min",
		Level:       "Intermediate",
		Description: "Master the fundamental network benchmarking tests",
		Steps:       rfc2544TutorialSteps(),
	}
}

func rfc2544TutorialSteps() []TutorialStep {
	steps := rfc2544TutorialIntroSteps()
	steps = append(steps, rfc2544TutorialTestSteps()...)
	steps = append(steps, rfc2544TutorialSummarySteps()...)
	return steps
}

func rfc2544TutorialIntroSteps() []TutorialStep {
	return []TutorialStep{
		{
			Title: "What is RFC 2544?",
			Content: `RFC 2544 "Benchmarking Methodology for Network Interconnect Devices"
	is the foundation of network performance testing. Published in 1999, it defines
	standardized tests that allow fair comparison between network equipment.

	The standard defines 6 tests:
	1. Throughput - Maximum lossless forwarding rate
	2. Latency - Round-trip delay
	3. Frame Loss Rate - Loss at various rates
	4. Back-to-Back - Burst capacity
	5. System Recovery - Recovery from overload
	6. Reset - Device restart time

	Let's explore each one.`,
			Command:  "",
			Expected: "",
			Tip:      "",
		},
	}
}

func rfc2544TutorialTestSteps() []TutorialStep {
	return []TutorialStep{
		{
			Title: "Throughput Test",
			Content: `The throughput test finds the maximum rate at which the DUT can
	forward packets without any loss.

	How it works:
	1. Start at 100% rate
	2. If any packets are lost, reduce rate
	3. If no loss, increase rate
	4. Use binary search to converge on exact maximum

	Run a throughput test:`,
			Command: "stem test -i eth0 -t throughput",
			Expected: `Frame Size | Max Rate | Throughput
64 bytes   | 98.5%    | 985 Mbps`,
			Tip: "Small frames (64 bytes) test packet processing; large frames test raw bandwidth",
		},
		{
			Title: "Latency Test",
			Content: `The latency test measures round-trip time at the throughput rate.
	This shows how long packets take to traverse the network.

	Key metrics:
	• Average latency - typical delay
	• Minimum latency - best case
	• Maximum latency - worst case
	• Jitter (std dev) - consistency

	Run a latency test:`,
			Command:  "stem test -i eth0 -t latency",
			Expected: "Avg: 125µs, Min: 98µs, Max: 245µs, Jitter: 23µs",
			Tip:      "Latency is measured at the throughput rate for realistic results",
		},
		{
			Title: "Frame Loss Rate Test",
			Content: `While throughput finds one point (max lossless rate), frame loss
	maps the entire performance curve.

	Test at decreasing rates to see:
	• At what rate does loss start?
	• How does loss increase with rate?
	• Where is the "knee" of the curve?

	Run frame loss test:`,
			Command: "stem test -i eth0 -t frame_loss",
			Expected: `Rate   | Loss
100%   | 2.3%
90%    | 0.1%
80%    | 0.0%`,
			Tip: "Use results to set operating thresholds below the loss point",
		},
		{
			Title: "Back-to-Back Test",
			Content: `Real traffic comes in bursts. The back-to-back test measures how
	many frames can be sent in a burst without loss.

	This reveals:
	• Device buffer capacity
	• Handling of traffic bursts
	• Risk of drops during peak traffic

	Run back-to-back test:`,
			Command:  "stem test -i eth0 -t back_to_back",
			Expected: "Max burst: 2048 frames at 64 bytes",
			Tip:      "Results indicate effective buffer size - important for video/streaming",
		},
	}
}

func rfc2544TutorialSummarySteps() []TutorialStep {
	return []TutorialStep{
		{
			Title:   "Combining Tests",
			Content: `For comprehensive validation, run all RFC 2544 tests together:`,
			Command: "stem test -i eth0 -t throughput,latency,frame_loss,back_to_back",
			Expected: `Running RFC 2544 test suite...
[Complete results for all tests]`,
			Tip: "Save results with --output report.json for documentation",
		},
		{
			Title: "Interpreting Results",
			Content: `What do the numbers mean?

	Throughput:
	• >95%: Excellent - equipment performing near spec
	• 80-95%: Good - acceptable for most uses
	• <80%: Investigate - potential bottleneck or misconfiguration

	Latency:
	• <1ms: Excellent for voice/video
	• 1-10ms: Good for most applications
	• 10-50ms: Acceptable for general use
	• >50ms: May cause noticeable delays

	Frame Loss:
	• 0%: Required for guaranteed delivery apps
	• <0.1%: OK for best-effort traffic
	• >0.1%: Indicates congestion or problems`,
			Command:  "",
			Expected: "",
			Tip:      "",
		},
	}
}

func Y1564Tutorial() Tutorial {
	return Tutorial{
		ID:          "y1564",
		Title:       "Carrier Ethernet Testing with Y.1564",
		Duration:    "15 min",
		Level:       "Intermediate",
		Description: "Learn the industry standard for service activation testing",
		Steps:       y1564TutorialSteps(),
	}
}

func y1564TutorialSteps() []TutorialStep {
	steps := y1564TutorialIntroSteps()
	steps = append(steps, y1564TutorialTestSteps()...)
	steps = append(steps, y1564TutorialFailureSteps()...)
	return steps
}

func y1564TutorialIntroSteps() []TutorialStep {
	return []TutorialStep{
		{
			Title: "What is Y.1564?",
			Content: `ITU-T Y.1564 is the industry standard for Ethernet Service Activation
	Testing (SAT). It's used by carriers worldwide to validate that services meet
	their Service Level Agreements (SLAs).

	When you buy a carrier ethernet service, Y.1564 tests verify:
	• You're getting the bandwidth you're paying for
	• Latency meets contractual limits
	• Packet loss is within acceptable bounds
	• Service works at all load levels

	Y.1564 includes two main tests:
	1. Service Configuration Test - validates at 25%, 50%, 75%, 100%
	2. Service Performance Test - extended duration validation`,
			Command:  "",
			Expected: "",
			Tip:      "",
		},
		{
			Title: "Understanding Your Service",
			Content: `Before testing, know your service parameters:

	CIR (Committed Information Rate):
	  The guaranteed bandwidth - what you're paying for
	  Example: 100 Mbps

	EIR (Excess Information Rate):
	  Bonus bandwidth when available (optional)
	  Example: 50 Mbps (total burst to 150 Mbps)

	SLA Thresholds:
	  • Frame Delay: max latency (e.g., <10ms)
	  • Frame Delay Variation: max jitter (e.g., <3ms)
	  • Frame Loss Ratio: max loss (e.g., <0.001%)

	Get these from your service contract.`,
			Command:  "",
			Expected: "",
			Tip:      "",
		},
	}
}

func y1564TutorialTestSteps() []TutorialStep {
	return []TutorialStep{
		{
			Title: "Service Configuration Test",
			Content: `The Configuration Test validates service at progressive load levels:

	Step 1: 25% of CIR - Basic connectivity
	Step 2: 50% of CIR - Medium load
	Step 3: 75% of CIR - Heavy load
	Step 4: 100% of CIR - Full committed rate

	All steps must pass (meet SLA thresholds).

	Run Configuration Test:`,
			Command: "stem test -i eth0 -t y1564_config --cir 100 --delay-threshold 10 --loss-threshold 0.001",
			Expected: `Step 25%: PASS (IR: 25.0 Mbps, FD: 2.1ms, FLR: 0.000%)
Step 50%: PASS
Step 75%: PASS
Step 100%: PASS`,
			Tip: "If any step fails, the service needs adjustment before acceptance",
		},
		{
			Title:   "Service Performance Test",
			Content: `After Configuration passes, run Performance Test to verify sustained quality:`,
			Command: "stem test -i eth0 -t y1564_performance --cir 100 --duration 15",
			Expected: `15-minute performance test at 100 Mbps
All metrics within SLA for entire duration`,
			Tip: "Standard duration is 15 minutes; use longer for thorough validation",
		},
		{
			Title:   "Full SAC Test",
			Content: `For official service activation, run the complete SAC test:`,
			Command: "stem test -i eth0 -t y1564_full --cir 100 --perf-duration 15",
			Expected: `Configuration Test: PASS
Performance Test: PASS

SERVICE ACCEPTED`,
			Tip: "Save results as documentation for service acceptance",
		},
	}
}

func y1564TutorialFailureSteps() []TutorialStep {
	return []TutorialStep{
		{
			Title: "When Tests Fail",
			Content: `Common Y.1564 failures and causes:

	Fails at 100% but passes at lower rates:
	  Cause: Service not provisioned to full CIR
	  Action: Contact carrier to verify provisioning

	High frame delay at all steps:
	  Cause: Long path, congestion, or routing issue
	  Action: Request path optimization from carrier

	Frame loss at specific rates:
	  Cause: Bottleneck somewhere in path
	  Action: Work with carrier to identify and resolve

	Intermittent failures during Performance Test:
	  Cause: Time-of-day congestion or instability
	  Action: Test at different times; report pattern to carrier`,
			Command:  "",
			Expected: "",
			Tip:      "",
		},
	}
}

func TroubleshootTutorial() Tutorial {
	return Tutorial{
		ID:          "troubleshoot",
		Title:       "Troubleshooting Test Failures",
		Duration:    "10 min",
		Level:       "Intermediate",
		Description: "Diagnose and resolve common testing issues",
		Steps: []TutorialStep{
			{
				Title: "Test Shows 0% Throughput",
				Content: `If your test shows 0% throughput or times out:

Check connectivity:`,
				Command:  "ping <reflector-ip>",
				Expected: "Should get responses",
				Tip:      "No response = network connectivity issue, not a test problem",
			},
			{
				Title:    "Verify Interface",
				Content:  `Make sure you're using the correct interface:`,
				Command:  "ip link show",
				Expected: "Interface should be UP with carrier",
				Tip:      "Look for 'state UP' and check for errors",
			},
			{
				Title:    "Check Reflector Status",
				Content:  `Verify the reflector is running and receiving packets:`,
				Command:  "curl -k https://<reflector-ip>:8444/api/v1/stats",
				Expected: "Shows packet counts and status",
				Tip:      "If packets_received is 0, traffic isn't reaching reflector",
			},
			{
				Title: "Results Vary Significantly",
				Content: `If results change dramatically between test runs:

Possible causes:
• Other traffic on the network
• DUT under load from other sources
• Environmental issues (overheating)
• Intermittent cable/connector problems

Troubleshoot:`,
				Command:  "stem test -i eth0 -t throughput --duration 120",
				Expected: "Longer test reveals patterns in variability",
				Tip:      "Test during maintenance window for most consistent results",
			},
			{
				Title: "Lower Than Expected Performance",
				Content: `Getting 800 Mbps on a 1 Gbps link?

Common causes:
• Duplex mismatch (check both ends)
• Speed negotiation issues
• Frame size overhead
• Ethernet IFG and preamble

Check interface settings:`,
				Command:  "ethtool eth0",
				Expected: "Speed: 1000Mb/s, Duplex: Full",
				Tip:      "Remember: theoretical max is always higher than achievable throughput",
			},
			{
				Title:    "Permission Denied Errors",
				Content:  `Network testing often requires root privileges:`,
				Command:  "sudo stem reflect -i eth0",
				Expected: "Reflector starts successfully",
				Tip:      "AF_XDP and DPDK modes especially require elevated privileges",
			},
		},
	}
}

// ResultsTutorial - Understanding results.
func ResultsTutorial() Tutorial {
	return Tutorial{
		ID:          "results",
		Title:       "Understanding Test Results",
		Duration:    "10 min",
		Level:       "Beginner",
		Description: "Learn to interpret and use test results effectively",
		Steps: []TutorialStep{
			{
				Title: "Reading Throughput Results",
				Content: `Throughput results show maximum lossless rate per frame size:

Frame Size | Max Rate | Throughput
64 bytes   | 98.5%    | 985 Mbps
1518 bytes | 99.2%    | 992 Mbps

What the columns mean:
• Frame Size: Packet size tested
• Max Rate: % of theoretical maximum achieved
• Throughput: Actual bits per second

Why different frame sizes?
• Small frames test packet processing (forwarding speed)
• Large frames test raw bandwidth capability
• Real traffic is a mix of sizes`,
				Command:  "",
				Expected: "",
				Tip:      "",
			},
			{
				Title: "Reading Latency Results",
				Content: `Latency results show timing statistics:

Frame Size | Avg    | Min   | Max    | Jitter
64 bytes   | 125µs  | 98µs  | 245µs  | 23µs

What each metric means:
• Avg: Typical delay most packets experience
• Min: Best-case delay (empty queues)
• Max: Worst-case delay (highest spike)
• Jitter: How much delay varies

For real-time apps (voice/video):
• Average matters for overall quality
• Max matters for worst-case experience
• Jitter affects smoothness`,
				Command:  "",
				Expected: "",
				Tip:      "",
			},
			{
				Title: "Pass/Fail Criteria",
				Content: `How to determine if results are good:

For RFC 2544 (equipment benchmarking):
• Compare to vendor specifications
• Compare to previous baseline
• Compare different vendors/models

For Y.1564 (service validation):
• Compare to SLA thresholds
• All metrics must be within limits
• All CIR steps must pass

General guidelines:
• Throughput >90% = Good
• Latency <1ms = Excellent for LAN
• Jitter <100µs = Good for voice/video
• Loss 0% = Required for guaranteed delivery`,
				Command:  "",
				Expected: "",
				Tip:      "",
			},
			{
				Title:    "Saving and Comparing Results",
				Content:  `Save results for documentation and comparison:`,
				Command:  "stem test -i eth0 -t throughput --output baseline_2025-01.json",
				Expected: "Results saved with timestamp",
				Tip:      "Include date in filename for easy comparison over time",
			},
			{
				Title:    "Creating Reports",
				Content:  `Generate formatted reports:`,
				Command:  "stem test -i eth0 -t throughput --output report.csv",
				Expected: "CSV file importable to Excel/Google Sheets",
				Tip:      "CSV format works well for creating charts and analysis",
			},
		},
	}
}
