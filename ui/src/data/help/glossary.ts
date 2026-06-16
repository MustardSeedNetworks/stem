/**
 * @fileoverview The Stem - Help Content Glossary
 * @description Glossary entries and search function for the WebUI help system.
 */

import type { GlossaryEntry } from './types';

export const glossary: Record<string, GlossaryEntry> = {
  cir: {
    term: 'CIR',
    fullName: 'Committed Information Rate',
    category: 'Bandwidth',
    techDef:
      'The guaranteed bandwidth as specified in a service level agreement, below which the network commits to deliver traffic with minimal loss or delay.',
    laymanDef: 'The speed your internet/network contract guarantees you will always get.',
    related: ['eir', 'bandwidth', 'sla'],
  },
  eir: {
    term: 'EIR',
    fullName: 'Excess Information Rate',
    category: 'Bandwidth',
    techDef:
      'The rate above CIR up to which the network will attempt to deliver frames, but without guarantees.',
    laymanDef: 'Bonus bandwidth you might get when the network is not busy, but no promises.',
    related: ['cir', 'bandwidth'],
  },
  bandwidth: {
    term: 'Bandwidth',
    fullName: 'Network Bandwidth',
    category: 'Bandwidth',
    techDef:
      'Maximum rate of data transfer across a network path, typically measured in bits per second.',
    laymanDef:
      'How much data can flow through your connection per second - like water through a pipe.',
    related: ['throughput', 'cir'],
  },
  throughput: {
    term: 'Throughput',
    fullName: 'Network Throughput',
    category: 'Bandwidth',
    techDef:
      'The actual rate of successful data transfer, accounting for protocol overhead and network conditions.',
    laymanDef:
      'How much useful data actually gets through - slightly less than bandwidth due to overhead.',
    related: ['bandwidth', 'goodput'],
  },
  latency: {
    term: 'Latency',
    fullName: 'Network Latency',
    category: 'Latency',
    techDef: 'The time delay for a packet to travel from source to destination.',
    laymanDef: 'The "lag" - how long it takes for your data to reach the other end.',
    related: ['rtt', 'delay', 'jitter'],
  },
  rtt: {
    term: 'RTT',
    fullName: 'Round-Trip Time',
    category: 'Latency',
    techDef:
      'The time for a signal to travel to a destination and back, including processing delays.',
    laymanDef: 'How long a round trip takes - send a message, get a reply.',
    related: ['latency', 'ping'],
  },
  jitter: {
    term: 'Jitter',
    fullName: 'Packet Delay Variation',
    category: 'Latency',
    techDef: 'The variation in latency between successive packets in a flow.',
    laymanDef: 'How much the delay wobbles around - matters for video calls and gaming.',
    related: ['latency', 'pdv'],
  },
  pdv: {
    term: 'PDV',
    fullName: 'Packet Delay Variation',
    category: 'Latency',
    techDef: 'ITU-T standard term for jitter, measuring the difference in delay between packets.',
    laymanDef: 'The technical name for jitter - timing inconsistency.',
    related: ['jitter', 'latency'],
  },
  pps: {
    term: 'pps',
    fullName: 'Packets Per Second',
    category: 'Performance',
    techDef: 'The rate at which network packets are processed or transmitted.',
    laymanDef: 'How many packets your network can handle each second.',
    related: ['throughput', 'mpps'],
  },
  mpps: {
    term: 'Mpps',
    fullName: 'Million Packets Per Second',
    category: 'Performance',
    techDef: 'Standard unit for measuring high-speed packet processing rates.',
    laymanDef: 'Millions of packets per second - for measuring fast switches and routers.',
    related: ['pps'],
  },
  mtu: {
    term: 'MTU',
    fullName: 'Maximum Transmission Unit',
    category: 'Protocol',
    techDef: 'The largest packet size (in bytes) that can be transmitted without fragmentation.',
    laymanDef: 'The biggest chunk of data you can send at once - usually 1500 bytes.',
    related: ['frame_size', 'jumbo_frames'],
  },
  jumbo_frames: {
    term: 'Jumbo Frames',
    fullName: 'Jumbo Frames',
    category: 'Protocol',
    techDef: 'Ethernet frames larger than 1500 bytes, typically up to 9000 bytes.',
    laymanDef: 'Extra-large packets for data centers - more efficient but need special support.',
    related: ['mtu', 'frame_size'],
  },
  mac: {
    term: 'MAC',
    fullName: 'Media Access Control',
    category: 'Protocol',
    techDef: 'Layer 2 hardware address that uniquely identifies network interfaces.',
    laymanDef: 'The unique hardware address burned into every network card.',
    related: ['ethernet', 'layer2'],
  },
  dut: {
    term: 'DUT',
    fullName: 'Device Under Test',
    category: 'Testing',
    techDef: 'The network device being tested and evaluated.',
    laymanDef: 'The thing you are testing.',
    related: ['sut'],
  },
  sut: {
    term: 'SUT',
    fullName: 'System Under Test',
    category: 'Testing',
    techDef: 'The complete system including all devices in the test path.',
    laymanDef: 'Everything you are testing together as a system.',
    related: ['dut'],
  },
  sla: {
    term: 'SLA',
    fullName: 'Service Level Agreement',
    category: 'Service',
    techDef: 'Contractual agreement specifying performance guarantees and penalties.',
    laymanDef: 'The contract that says what your provider promises to deliver.',
    related: ['cir', 'oam'],
  },
  oam: {
    term: 'OAM',
    fullName: 'Operations, Administration, and Maintenance',
    category: 'Service',
    techDef: 'Tools and protocols for monitoring and managing networks, including Y.1731.',
    laymanDef: 'Built-in network monitoring tools that carriers use.',
    related: ['y1731', 'cfm'],
  },
  cfm: {
    term: 'CFM',
    fullName: 'Connectivity Fault Management',
    category: 'Service',
    techDef: 'IEEE 802.1ag protocol for detecting and isolating connectivity faults.',
    laymanDef: 'System for automatically finding network problems.',
    related: ['oam', 'y1731'],
  },
  xdp: {
    term: 'XDP',
    fullName: 'eXpress Data Path',
    category: 'Technology',
    techDef:
      'Linux kernel technology for high-performance packet processing before the network stack.',
    laymanDef: 'Fast-path technology that processes packets super quickly.',
    related: ['afxdp', 'dpdk'],
  },
  afxdp: {
    term: 'AF_XDP',
    fullName: 'Address Family XDP',
    category: 'Technology',
    techDef: 'Linux socket type for zero-copy packet processing using XDP.',
    laymanDef: 'Way to handle network packets extremely fast in Linux.',
    related: ['xdp', 'dpdk'],
  },
  dpdk: {
    term: 'DPDK',
    fullName: 'Data Plane Development Kit',
    category: 'Technology',
    techDef: 'Set of libraries for fast packet processing by bypassing the kernel.',
    laymanDef: 'Technology for super-fast networking, used in carriers and data centers.',
    related: ['xdp', 'afxdp'],
  },
  tos: {
    term: 'ToS',
    fullName: 'Type of Service',
    category: 'QoS',
    techDef: 'IPv4 header field used for quality of service prioritization.',
    laymanDef: 'Tag in IP packets to mark priority level.',
    related: ['dscp', 'cos'],
  },
  dscp: {
    term: 'DSCP',
    fullName: 'Differentiated Services Code Point',
    category: 'QoS',
    techDef: '6-bit field in IP header for traffic classification and QoS.',
    laymanDef: 'Modern way to mark packet priority - tells the network how important a packet is.',
    related: ['tos', 'cos'],
  },
  cos: {
    term: 'CoS',
    fullName: 'Class of Service',
    category: 'QoS',
    techDef: '3-bit field in VLAN tag for Layer 2 traffic classification (IEEE 802.1p).',
    laymanDef: 'Priority marking at the Ethernet level (layer 2).',
    related: ['dscp', 'vlan'],
  },
  vlan: {
    term: 'VLAN',
    fullName: 'Virtual LAN',
    category: 'Protocol',
    techDef: 'Logical network partition at Layer 2, typically using IEEE 802.1Q tagging.',
    laymanDef: 'Virtual networks that keep traffic separated on the same physical equipment.',
    related: ['cos', 'qinq'],
  },
  qinq: {
    term: 'Q-in-Q',
    fullName: 'IEEE 802.1ad (Double Tagging)',
    category: 'Protocol',
    techDef: 'Stacking of VLAN tags for provider/customer separation.',
    laymanDef: 'Putting a VLAN tag inside another VLAN tag - used by carriers.',
    related: ['vlan', 'svlan'],
  },
  bdp: {
    term: 'BDP',
    fullName: 'Bandwidth-Delay Product',
    category: 'Performance',
    techDef: 'Product of bandwidth and RTT, representing data in flight for optimal TCP.',
    laymanDef: 'How much data should be "in the air" at once for maximum speed.',
    related: ['rtt', 'tcp'],
  },
  tcp: {
    term: 'TCP',
    fullName: 'Transmission Control Protocol',
    category: 'Protocol',
    techDef: 'Connection-oriented transport protocol with reliability and flow control.',
    laymanDef: 'The protocol that makes sure data arrives correctly and in order.',
    related: ['bdp', 'udp'],
  },
  udp: {
    term: 'UDP',
    fullName: 'User Datagram Protocol',
    category: 'Protocol',
    techDef: 'Connectionless transport protocol without reliability guarantees.',
    laymanDef: 'Fast but unreliable - used for video streaming and gaming.',
    related: ['tcp'],
  },
  tsn: {
    term: 'TSN',
    fullName: 'Time-Sensitive Networking',
    category: 'Technology',
    techDef: 'IEEE 802.1 standards for deterministic, low-latency Ethernet communication.',
    laymanDef: 'Network technology where packets arrive at EXACTLY the right time.',
    related: ['tas', 'gcl'],
  },
  tas: {
    term: 'TAS',
    fullName: 'Time-Aware Shaper',
    category: 'Technology',
    techDef: 'IEEE 802.1Qbv mechanism for scheduled traffic transmission.',
    laymanDef: 'Network feature that opens and closes time gates for traffic.',
    related: ['tsn', 'gcl'],
  },
  gcl: {
    term: 'GCL',
    fullName: 'Gate Control List',
    category: 'Technology',
    techDef: 'Schedule defining when traffic classes can transmit in TSN.',
    laymanDef: 'The schedule that says when each type of traffic can go.',
    related: ['tsn', 'tas'],
  },
  reflector: {
    term: 'Reflector',
    fullName: 'Packet Reflector',
    category: 'Testing',
    techDef: 'Endpoint that returns received packets for round-trip measurement.',
    laymanDef: 'Device at the far end that bounces packets back for testing.',
    related: ['loopback'],
  },
  loopback: {
    term: 'Loopback',
    fullName: 'Network Loopback',
    category: 'Testing',
    techDef: 'Interface or mechanism that returns traffic to its source.',
    laymanDef: 'Sending traffic back to where it came from for testing.',
    related: ['reflector'],
  },
  binary_search: {
    term: 'Binary Search',
    fullName: 'Binary Search Algorithm',
    category: 'Testing',
    techDef: 'Algorithm that halves the search space each iteration to find optimal rate.',
    laymanDef: 'Quickly finding the right speed by trying half as much each time.',
    related: ['throughput'],
  },
  frame_loss: {
    term: 'Frame Loss',
    fullName: 'Packet/Frame Loss',
    category: 'Performance',
    techDef: 'Percentage of transmitted frames that fail to arrive at destination.',
    laymanDef: 'Packets that got lost along the way.',
    related: ['packet_loss'],
  },
};

// Helper to search glossary
export function searchGlossary(keyword: string): GlossaryEntry[] {
  const lower = keyword.toLowerCase();
  return Object.values(glossary).filter(
    (e) =>
      e.term.toLowerCase().includes(lower) ||
      e.fullName.toLowerCase().includes(lower) ||
      e.techDef.toLowerCase().includes(lower) ||
      e.laymanDef.toLowerCase().includes(lower),
  );
}
