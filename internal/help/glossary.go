/*
 * Seed Test Suite - Network Glossary
 *
 * Definitions for network terminology, accessible to both technical users
 * and newcomers to network testing.
 *
 * Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
 */

package help

import "maps"

// GetGlossary returns all glossary entries.
func GetGlossary() map[string]GlossaryEntry {
	glossary := make(map[string]GlossaryEntry)
	addEntries(glossary, getBandwidthTerms())
	addEntries(glossary, getLatencyTerms())
	addEntries(glossary, getLossTerms())
	addEntries(glossary, getProtocolTerms())
	addEntries(glossary, getAddressingTerms())
	addEntries(glossary, getQoSTerms())
	addEntries(glossary, getEquipmentTerms())
	addEntries(glossary, getServiceTerms())
	addEntries(glossary, getMeasurementTerms())
	addEntries(glossary, getFramePacketTerms())
	addEntries(glossary, getOAMTerms())
	addEntries(glossary, getTSNTerms())
	addEntries(glossary, getLayerTerms())
	addEntries(glossary, getTestingTerms())
	addEntries(glossary, getCongestionTerms())
	addEntries(glossary, getStandardsTerms())
	return glossary
}

// addEntries adds all entries from src to dst.
func addEntries(dst, src map[string]GlossaryEntry) {
	maps.Copy(dst, src)
}

// getBandwidthTerms returns bandwidth and rate terms.
func getBandwidthTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"bandwidth": {
			Term:      "Bandwidth",
			FullName:  "Network Bandwidth",
			TechDef:   "The maximum data transfer capacity of a network link, measured in bits per second (bps)",
			LaymanDef: "How much data your network can carry - like the width of a highway. More bandwidth = more data can flow at once",
			Related:   []string{"throughput", "cir", "line_rate"},
		},
		"throughput": {
			Term:      "Throughput",
			FullName:  "Network Throughput",
			TechDef:   "The actual rate of successful data transfer, accounting for protocol overhead and losses",
			LaymanDef: "The real speed you get - always less than the maximum because of overhead. Like your actual driving speed vs the speed limit",
			Related:   []string{"bandwidth", "goodput"},
		},
		"cir": {
			Term:      "CIR",
			FullName:  "Committed Information Rate",
			TechDef:   "The guaranteed bandwidth specified in a service level agreement that the carrier commits to deliver",
			LaymanDef: "The speed your ISP promises you'll always get, no matter what",
			Related:   []string{"eir", "sla", "bandwidth"},
		},
		"eir": {
			Term:      "EIR",
			FullName:  "Excess Information Rate",
			TechDef:   "Bandwidth above the CIR that may be available when the network has spare capacity",
			LaymanDef: "Extra speed you might get when the network isn't busy - not guaranteed, but a nice bonus",
			Related:   []string{"cir", "cbs", "ebs"},
		},
		"cbs": {
			Term:      "CBS",
			FullName:  "Committed Burst Size",
			TechDef:   "Maximum amount of data that can be sent at CIR in a single burst",
			LaymanDef: "How much data you can send at once at your guaranteed speed",
			Related:   []string{"cir", "ebs"},
		},
		"ebs": {
			Term:      "EBS",
			FullName:  "Excess Burst Size",
			TechDef:   "Maximum burst size allowed at EIR when excess bandwidth is available",
			LaymanDef: "How much bonus data you can send when extra bandwidth is available",
			Related:   []string{"eir", "cbs"},
		},
		"line_rate": {
			Term:      "Line Rate",
			FullName:  "Line Rate / Wire Speed",
			TechDef:   "The maximum theoretical bit rate of a physical interface (e.g., 1 Gbps, 10 Gbps)",
			LaymanDef: "The maximum speed the cable/port can physically handle - like the top speed of a car",
			Related:   []string{"bandwidth", "throughput"},
		},
		"goodput": {
			Term:      "Goodput",
			FullName:  "Application Goodput",
			TechDef:   "The rate of useful data transfer excluding protocol overhead, retransmissions, and headers",
			LaymanDef: "The actual useful data speed your applications see - always less than throughput due to overhead",
			Related:   []string{"throughput"},
		},
	}
}

// getLatencyTerms returns latency and timing terms.
func getLatencyTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"latency": {
			Term:      "Latency",
			FullName:  "Network Latency",
			TechDef:   "The time delay for a packet to travel from source to destination, usually measured in milliseconds or microseconds",
			LaymanDef: "The 'lag' - how long it takes for data to get from here to there. Lower is better for video calls and gaming",
			Related:   []string{"rtt", "jitter", "delay"},
		},
		"rtt": {
			Term:      "RTT",
			FullName:  "Round-Trip Time",
			TechDef:   "Total time for a packet to travel to a destination and back, including processing delays",
			LaymanDef: "Time for a message to go there and come back - what you see when you ping something",
			Related:   []string{"latency", "ping"},
		},
		"jitter": {
			Term:      "Jitter",
			FullName:  "Packet Delay Variation",
			TechDef:   "The variation in latency between packets in a flow. Low jitter means consistent timing",
			LaymanDef: "How consistent the delay is. High jitter = choppy video/audio because packets arrive at uneven times",
			Related:   []string{"latency", "fdv"},
		},
		"fdv": {
			Term:      "FDV",
			FullName:  "Frame Delay Variation",
			TechDef:   "ITU-T term for jitter in carrier ethernet contexts",
			LaymanDef: "Same as jitter - how much the timing varies between packets",
			Related:   []string{"jitter", "fd"},
		},
		"fd": {
			Term:      "FD",
			FullName:  "Frame Delay",
			TechDef:   "ITU-T term for the delay of individual frames in carrier ethernet",
			LaymanDef: "How long it takes for a single data packet to arrive",
			Related:   []string{"latency", "fdv"},
		},
	}
}

// getLossTerms returns loss-related terms.
func getLossTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"packet_loss": {
			Term:      "Packet Loss",
			FullName:  "Packet Loss Rate",
			TechDef:   "The percentage of packets that fail to reach their destination",
			LaymanDef: "Lost messages - like letters that never arrive. Even 0.1% loss can cause problems for video calls",
			Related:   []string{"flr", "frame_loss"},
		},
		"flr": {
			Term:      "FLR",
			FullName:  "Frame Loss Ratio",
			TechDef:   "ITU-T term for the ratio of lost frames to total frames sent",
			LaymanDef: "What percentage of data packets got lost",
			Related:   []string{"packet_loss"},
		},
	}
}

// getProtocolTerms returns protocol-related terms.
func getProtocolTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"ethernet": {
			Term:      "Ethernet",
			FullName:  "IEEE 802.3 Ethernet",
			TechDef:   "The most common networking technology for local area networks (LANs), operating at Layer 2",
			LaymanDef: "The standard way computers talk to each other on a local network - those cables with the clip",
			Related:   []string{"mac_address", "frame"},
		},
		"tcp": {
			Term:      "TCP",
			FullName:  "Transmission Control Protocol",
			TechDef:   "A connection-oriented transport protocol that guarantees reliable, ordered delivery of data",
			LaymanDef: "A careful delivery service - makes sure everything arrives in order and resends if something gets lost",
			Related:   []string{"udp", "ip"},
		},
		"udp": {
			Term:      "UDP",
			FullName:  "User Datagram Protocol",
			TechDef:   "A connectionless transport protocol that provides fast but unreliable delivery",
			LaymanDef: "A fast delivery service - doesn't check if things arrive, but faster than TCP. Used for video streaming",
			Related:   []string{"tcp", "ip"},
		},
		"ip": {
			Term:      "IP",
			FullName:  "Internet Protocol",
			TechDef:   "The principal protocol for routing packets across network boundaries (Layer 3)",
			LaymanDef: "The addressing system of the internet - how devices find each other",
			Related:   []string{"tcp", "udp", "routing"},
		},
		"vlan": {
			Term:      "VLAN",
			FullName:  "Virtual Local Area Network",
			TechDef:   "A logical network partition that separates traffic at Layer 2 using 802.1Q tags",
			LaymanDef: "A way to divide a physical network into separate virtual networks - like having multiple neighborhoods on one street",
			Related:   []string{"cos", "qos"},
		},
	}
}

// getAddressingTerms returns addressing-related terms.
func getAddressingTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"mac_address": {
			Term:      "MAC Address",
			FullName:  "Media Access Control Address",
			TechDef:   "A unique 48-bit hardware identifier assigned to network interfaces",
			LaymanDef: "A unique ID burned into every network device - like a serial number. Format: 00:1A:2B:3C:4D:5E",
			Related:   []string{"ip", "ethernet"},
		},
		"oui": {
			Term:      "OUI",
			FullName:  "Organizationally Unique Identifier",
			TechDef:   "The first three octets of a MAC address, identifying the manufacturer",
			LaymanDef: "The first half of a MAC address tells you who made the device - like a brand name",
			Related:   []string{"mac_address"},
		},
	}
}

// getQoSTerms returns Quality of Service terms.
func getQoSTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"qos": {
			Term:      "QoS",
			FullName:  "Quality of Service",
			TechDef:   "Network mechanisms that prioritize certain traffic to guarantee performance levels",
			LaymanDef: "Traffic priority - making sure important data (like voice calls) gets through before less important data",
			Related:   []string{"cos", "dscp", "sla"},
		},
		"cos": {
			Term:      "CoS",
			FullName:  "Class of Service",
			TechDef:   "A 3-bit field in 802.1Q VLAN tags indicating priority (0-7)",
			LaymanDef: "Priority levels (0-7) where 7 is highest. Voice might be 6, regular data might be 0",
			Related:   []string{"qos", "dscp", "vlan"},
		},
		"dscp": {
			Term:      "DSCP",
			FullName:  "Differentiated Services Code Point",
			TechDef:   "A 6-bit field in IP headers for packet classification and priority marking",
			LaymanDef: "Like CoS but for IP packets - a way to mark traffic priority that works across networks",
			Related:   []string{"qos", "cos"},
		},
	}
}

// getEquipmentTerms returns equipment-related terms.
func getEquipmentTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"dut": {
			Term:      "DUT",
			FullName:  "Device Under Test",
			TechDef:   "The network equipment being tested or benchmarked",
			LaymanDef: "The thing you're testing - could be a router, switch, firewall, etc.",
			Related:   []string{"sut"},
		},
		"sut": {
			Term:      "SUT",
			FullName:  "System Under Test",
			TechDef:   "The complete system or network being tested, may include multiple devices",
			LaymanDef: "The whole setup you're testing, not just one device",
			Related:   []string{"dut"},
		},
		"switch": {
			Term:      "Switch",
			FullName:  "Network Switch",
			TechDef:   "A Layer 2 device that forwards frames based on MAC addresses",
			LaymanDef: "A device that connects multiple computers on a network and directs traffic to the right destination",
			Related:   []string{"router", "bridge"},
		},
		"router": {
			Term:      "Router",
			FullName:  "Network Router",
			TechDef:   "A Layer 3 device that forwards packets between different networks based on IP addresses",
			LaymanDef: "A device that connects different networks and decides where to send data - like a traffic controller",
			Related:   []string{"switch", "gateway"},
		},
		"reflector": {
			Term:      "Reflector",
			FullName:  "Packet Reflector",
			TechDef:   "A device or function that returns received test packets to their source for round-trip measurements",
			LaymanDef: "A device that bounces test packets back - like a mirror for network traffic",
			Related:   []string{"loopback"},
		},
	}
}

// getServiceTerms returns service-related terms.
func getServiceTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"sla": {
			Term:      "SLA",
			FullName:  "Service Level Agreement",
			TechDef:   "A contract defining measurable service quality parameters like bandwidth, latency, and availability",
			LaymanDef: "A promise from your service provider about what quality you'll get - and what happens if they don't deliver",
			Related:   []string{"cir", "eir", "qos"},
		},
		"carrier_ethernet": {
			Term:      "Carrier Ethernet",
			FullName:  "Carrier-Grade Ethernet Service",
			TechDef:   "Ethernet services provided by telecommunications carriers with SLA guarantees",
			LaymanDef: "Professional ethernet service from a phone/cable company with guaranteed quality - not just best effort",
			Related:   []string{"mef", "y1564"},
		},
		"mef": {
			Term:      "MEF",
			FullName:  "Metro Ethernet Forum",
			TechDef:   "Industry organization defining standards for carrier ethernet services",
			LaymanDef: "The group that sets standards for professional ethernet services - their certification means it meets industry standards",
			Related:   []string{"carrier_ethernet", "y1564"},
		},
	}
}

// getMeasurementTerms returns measurement-related terms.
func getMeasurementTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"pps": {
			Term:      "pps",
			FullName:  "Packets Per Second",
			TechDef:   "Rate measurement of packet forwarding, typically expressed in thousands (Kpps) or millions (Mpps)",
			LaymanDef: "How many individual data packets per second - important for small packet performance",
			Related:   []string{"bps", "fps"},
		},
		"bps": {
			Term:      "bps",
			FullName:  "Bits Per Second",
			TechDef:   "Rate of data transfer. Common units: Kbps, Mbps, Gbps (thousands, millions, billions)",
			LaymanDef: "Speed measurement - Mbps (megabits per second) is what ISPs advertise. 1 Gbps = 1000 Mbps",
			Related:   []string{"pps", "bandwidth"},
		},
		"fps": {
			Term:      "fps",
			FullName:  "Frames Per Second",
			TechDef:   "Ethernet frame forwarding rate, equivalent to pps at Layer 2",
			LaymanDef: "Same as packets per second, but specifically for Ethernet frames",
			Related:   []string{"pps"},
		},
	}
}

// getFramePacketTerms returns frame and packet terms.
func getFramePacketTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"frame": {
			Term:      "Frame",
			FullName:  "Ethernet Frame",
			TechDef:   "A Layer 2 data unit containing MAC addresses, type/length, payload, and FCS",
			LaymanDef: "A package of data with addresses on it - the basic unit that travels on an Ethernet network",
			Related:   []string{"packet", "mtu"},
		},
		"packet": {
			Term:      "Packet",
			FullName:  "Network Packet",
			TechDef:   "A Layer 3 data unit containing IP addresses and payload",
			LaymanDef: "A package of data with internet addresses - what travels across the internet",
			Related:   []string{"frame", "segment"},
		},
		"mtu": {
			Term:      "MTU",
			FullName:  "Maximum Transmission Unit",
			TechDef:   "The largest packet or frame size that can be transmitted on a network segment",
			LaymanDef: "The biggest data package allowed - usually 1500 bytes on regular networks",
			Related:   []string{"frame", "jumbo_frame"},
		},
		"jumbo_frame": {
			Term:      "Jumbo Frame",
			FullName:  "Jumbo Ethernet Frame",
			TechDef:   "Ethernet frames larger than 1518 bytes, typically up to 9000 bytes",
			LaymanDef: "Extra-large data packages for efficient bulk transfers - like using a truck instead of a car",
			Related:   []string{"mtu", "frame"},
		},
		"ifg": {
			Term:      "IFG",
			FullName:  "Inter-Frame Gap",
			TechDef:   "Minimum idle time between Ethernet frames (96 bit times = 9.6us at 10 Mbps)",
			LaymanDef: "Required pause between sending packets - like the space between cars on a highway",
			Related:   []string{"frame"},
		},
	}
}

// getOAMTerms returns OAM-related terms.
func getOAMTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"oam": {
			Term:      "OAM",
			FullName:  "Operations, Administration, and Maintenance",
			TechDef:   "Network management functions for monitoring, troubleshooting, and maintaining services",
			LaymanDef: "Built-in tools for checking if the network is healthy and finding problems",
			Related:   []string{"y1731", "cfm"},
		},
		"cfm": {
			Term:      "CFM",
			FullName:  "Connectivity Fault Management",
			TechDef:   "IEEE 802.1ag standard for detecting, isolating, and reporting connectivity faults",
			LaymanDef: "Automatic system for finding where network connections are broken",
			Related:   []string{"oam", "y1731"},
		},
		"mep": {
			Term:      "MEP",
			FullName:  "Maintenance Entity Group End Point",
			TechDef:   "An endpoint that generates and terminates OAM frames",
			LaymanDef: "A monitoring point at the edge of a network - where OAM measurements start and end",
			Related:   []string{"mip", "oam"},
		},
		"mip": {
			Term:      "MIP",
			FullName:  "Maintenance Entity Group Intermediate Point",
			TechDef:   "An intermediate point that can respond to OAM frames for fault isolation",
			LaymanDef: "A monitoring point in the middle of a network - helps narrow down where problems are",
			Related:   []string{"mep", "oam"},
		},
	}
}

// getTSNTerms returns TSN-related terms.
func getTSNTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"tsn": {
			Term:      "TSN",
			FullName:  "Time-Sensitive Networking",
			TechDef:   "IEEE 802.1 standards for deterministic, low-latency networking in industrial applications",
			LaymanDef: "Special networking for factories and cars where timing must be EXACT - packets arrive at precise times",
			Related:   []string{"tas", "ptp"},
		},
		"tas": {
			Term:      "TAS",
			FullName:  "Time-Aware Shaper",
			TechDef:   "IEEE 802.1Qbv mechanism that uses time-based gates to schedule traffic transmission",
			LaymanDef: "Traffic lights for network packets - only lets certain traffic through at specific times",
			Related:   []string{"tsn", "gcl"},
		},
		"gcl": {
			Term:      "GCL",
			FullName:  "Gate Control List",
			TechDef:   "The schedule defining when each traffic class gate opens and closes in TAS",
			LaymanDef: "The timetable for network traffic - says when each type of traffic is allowed to go",
			Related:   []string{"tas", "tsn"},
		},
		"ptp": {
			Term:      "PTP",
			FullName:  "Precision Time Protocol",
			TechDef:   "IEEE 1588 protocol for precise clock synchronization, achieving sub-microsecond accuracy",
			LaymanDef: "A way to synchronize clocks across a network to nanosecond precision - essential for TSN",
			Related:   []string{"tsn", "ntp"},
		},
	}
}

// getLayerTerms returns network layer terms.
func getLayerTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"layer2": {
			Term:      "Layer 2",
			FullName:  "Data Link Layer",
			TechDef:   "OSI model layer handling MAC addressing and frame transmission",
			LaymanDef: "The network level that deals with physical connections and MAC addresses",
			Related:   []string{"layer3", "ethernet", "mac_address"},
		},
		"layer3": {
			Term:      "Layer 3",
			FullName:  "Network Layer",
			TechDef:   "OSI model layer handling IP addressing and routing",
			LaymanDef: "The network level that deals with IP addresses and routing between networks",
			Related:   []string{"layer2", "ip", "routing"},
		},
	}
}

// getTestingTerms returns testing-related terms.
func getTestingTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"binary_search": {
			Term:      "Binary Search",
			FullName:  "Binary Search Algorithm",
			TechDef:   "Algorithm used in RFC 2544 to efficiently find maximum throughput by halving the search range",
			LaymanDef: "A smart way to find the right speed - instead of trying every speed, it narrows down quickly by halves",
			Related:   []string{"throughput"},
		},
		"baseline": {
			Term:      "Baseline",
			FullName:  "Performance Baseline",
			TechDef:   "Reference measurements taken under known good conditions for comparison",
			LaymanDef: "Your 'normal' measurements - so you can tell if something changes later",
			Related:   []string{"benchmark"},
		},
		"benchmark": {
			Term:      "Benchmark",
			FullName:  "Performance Benchmark",
			TechDef:   "Standardized tests for comparing equipment or network performance",
			LaymanDef: "Official tests for comparing how fast different equipment is",
			Related:   []string{"baseline", "rfc2544"},
		},
	}
}

// getCongestionTerms returns congestion-related terms.
func getCongestionTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"congestion": {
			Term:      "Congestion",
			FullName:  "Network Congestion",
			TechDef:   "Condition where network demand exceeds capacity, causing queuing or packet loss",
			LaymanDef: "Network traffic jam - too much data trying to go through at once",
			Related:   []string{"buffer", "queue"},
		},
		"buffer": {
			Term:      "Buffer",
			FullName:  "Packet Buffer",
			TechDef:   "Memory used to temporarily store packets during congestion or processing",
			LaymanDef: "Storage space in network equipment for packets waiting their turn",
			Related:   []string{"queue", "buffer_bloat"},
		},
		"buffer_bloat": {
			Term:      "Buffer Bloat",
			FullName:  "Bufferbloat",
			TechDef:   "Excessive buffering that causes high latency without improving throughput",
			LaymanDef: "When buffers are so big that packets wait too long - causes lag even though nothing is lost",
			Related:   []string{"buffer", "latency"},
		},
		"queue": {
			Term:      "Queue",
			FullName:  "Packet Queue",
			TechDef:   "A line of packets waiting to be processed or transmitted",
			LaymanDef: "A waiting line for packets - like a line at the store",
			Related:   []string{"buffer", "qos"},
		},
		"hol_blocking": {
			Term:      "HOL Blocking",
			FullName:  "Head-of-Line Blocking",
			TechDef:   "When a packet at the front of a queue blocks packets behind it destined for available outputs",
			LaymanDef: "When one slow lane backs up traffic for other lanes - like one slow checkout line blocking people who just want to pay",
			Related:   []string{"queue", "congestion"},
		},
	}
}

// getStandardsTerms returns standards-related terms.
func getStandardsTerms() map[string]GlossaryEntry {
	return map[string]GlossaryEntry{
		"rfc2544": {
			Term:      "RFC 2544",
			FullName:  "Benchmarking Methodology for Network Interconnect Devices",
			TechDef:   "IETF standard defining procedures for measuring network device performance",
			LaymanDef: "The official rulebook for how to test network equipment performance",
			Related:   []string{"rfc2889", "rfc6349"},
		},
		"rfc2889": {
			Term:      "RFC 2889",
			FullName:  "Benchmarking Methodology for LAN Switching Devices",
			TechDef:   "IETF standard extending RFC 2544 for switch-specific testing",
			LaymanDef: "Testing rules specifically for network switches",
			Related:   []string{"rfc2544"},
		},
		"rfc6349": {
			Term:      "RFC 6349",
			FullName:  "Framework for TCP Throughput Testing",
			TechDef:   "IETF standard for testing TCP performance considering protocol behavior",
			LaymanDef: "Testing rules for measuring real application speeds (not just raw network speed)",
			Related:   []string{"tcp", "rfc2544"},
		},
		"y1564": {
			Term:      "Y.1564",
			FullName:  "ITU-T Y.1564",
			TechDef:   "ITU standard for Ethernet Service Activation Test methodology",
			LaymanDef: "The official test for verifying carrier ethernet service quality",
			Related:   []string{"y1731", "mef"},
		},
		"y1731": {
			Term:      "Y.1731",
			FullName:  "ITU-T Y.1731",
			TechDef:   "ITU standard for OAM functions in Ethernet-based networks",
			LaymanDef: "Standard for monitoring and maintaining carrier ethernet networks",
			Related:   []string{"y1564", "oam"},
		},
	}
}

// GetGlossaryTermsByCategory returns glossary terms grouped by category.
func GetGlossaryTermsByCategory() map[string][]string {
	return map[string][]string{
		"Bandwidth & Rate": {
			"bandwidth", "throughput", "cir", "eir", "cbs", "ebs", "line_rate", "goodput",
		},
		"Latency & Timing": {
			"latency", "rtt", "jitter", "fdv", "fd",
		},
		"Loss": {
			"packet_loss", "flr",
		},
		"Protocols": {
			"ethernet", "tcp", "udp", "ip", "vlan",
		},
		"Addressing": {
			"mac_address", "oui",
		},
		"Quality of Service": {
			"qos", "cos", "dscp",
		},
		"Equipment": {
			"dut", "sut", "switch", "router", "reflector",
		},
		"Service Terms": {
			"sla", "carrier_ethernet", "mef",
		},
		"Measurements": {
			"pps", "bps", "fps",
		},
		"Frame & Packet": {
			"frame", "packet", "mtu", "jumbo_frame", "ifg",
		},
		"OAM": {
			"oam", "cfm", "mep", "mip",
		},
		"TSN": {
			"tsn", "tas", "gcl", "ptp",
		},
		"Network Layers": {
			"layer2", "layer3",
		},
		"Testing": {
			"binary_search", "baseline", "benchmark",
		},
		"Congestion": {
			"congestion", "buffer", "buffer_bloat", "queue", "hol_blocking",
		},
		"Standards": {
			"rfc2544", "rfc2889", "rfc6349", "y1564", "y1731",
		},
	}
}
