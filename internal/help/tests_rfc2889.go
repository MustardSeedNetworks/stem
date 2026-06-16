/*
 * The Stem - Test Documentation
 *
 * RFC 2889 test help content.
 */

package help

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
		SeeAlso:    []string{"address_cache", TestTypeThroughput},
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
