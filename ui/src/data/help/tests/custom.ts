/**
 * @fileoverview The Stem - Custom Traffic Generation Test Definitions
 * @description Help content for custom traffic generation tests.
 */

import type { TestHelp } from '../types';

export const customTests: Record<string, TestHelp> = {
  custom_stream: {
    id: 'custom_stream',
    name: 'Custom Traffic Stream',
    standard: 'N/A',
    category: 'TrafficGen',
    summary: 'Generate custom traffic patterns for specialized testing.',
    techDesc: `Custom traffic generation allows creation of arbitrary traffic patterns
including burst mode, specific MAC addresses, VLAN tagging, and controlled rates.
Useful for stress testing, QoS validation, and network diagnostics.`,
    laymanDesc: `Create your own custom traffic patterns for specialized tests.
Useful when standard tests don't cover your specific scenario.`,
    whenToUse: `• Custom stress testing scenarios
• QoS and traffic shaping validation
• Network diagnostics and debugging
• Vendor-specific testing requirements`,
    whenNotToUse: `• Use standard tests (RFC 2544, Y.1564) when applicable
• For certification or compliance testing`,
    parameters: [
      {
        name: 'Frame Size',
        flag: '--frame-size',
        type: 'integer (bytes)',
        defaultValue: '1518',
        required: false,
        techDesc: 'Ethernet frame size in bytes (64-9216)',
        laymanDesc: 'Size of each packet to generate',
        example: '--frame-size 512',
      },
      {
        name: 'Rate Percent',
        flag: '--rate-pct',
        type: 'float (percentage)',
        defaultValue: '100.0',
        required: false,
        techDesc: 'Traffic rate as percentage of line rate',
        laymanDesc: 'How fast to send traffic (100% = maximum speed)',
        example: '--rate-pct 50.0',
      },
      {
        name: 'Duration',
        flag: '--duration',
        type: 'integer (seconds)',
        defaultValue: '60',
        required: false,
        techDesc: 'Duration of traffic generation',
        laymanDesc: 'How long to generate traffic',
        example: '--duration 300',
      },
      {
        name: 'Warmup',
        flag: '--warmup',
        type: 'integer (seconds)',
        defaultValue: '2',
        required: false,
        techDesc: 'Warmup period before measurements',
        laymanDesc: 'Time to stabilize before counting',
        example: '--warmup 5',
      },
      {
        name: 'Stream ID',
        flag: '--stream-id',
        type: 'integer',
        defaultValue: '1',
        required: false,
        techDesc: 'Unique identifier for this traffic stream',
        laymanDesc: 'Label for tracking this traffic',
        example: '--stream-id 42',
      },
      {
        name: 'Burst Mode',
        flag: '--burst-mode',
        type: 'boolean',
        defaultValue: 'false',
        required: false,
        techDesc: 'Enable burst traffic mode instead of continuous',
        laymanDesc: 'Send traffic in bursts instead of steady flow',
        example: '--burst-mode',
      },
      {
        name: 'Burst Size',
        flag: '--burst-size',
        type: 'integer (frames)',
        defaultValue: '100',
        required: false,
        techDesc: 'Number of frames per burst (when burst-mode enabled)',
        laymanDesc: 'How many packets per burst',
        example: '--burst-size 50',
      },
      {
        name: 'Inter-Burst Gap',
        flag: '--inter-burst-gap-us',
        type: 'integer (microseconds)',
        defaultValue: '1000',
        required: false,
        techDesc: 'Gap between bursts in microseconds',
        laymanDesc: 'Pause between bursts',
        example: '--inter-burst-gap-us 500',
      },
      {
        name: 'Source MAC',
        flag: '--src-mac',
        type: 'string (MAC address)',
        defaultValue: '(interface MAC)',
        required: false,
        techDesc: 'Source MAC address for generated frames',
        laymanDesc: 'Sender address on packets',
        example: '--src-mac 00:11:22:33:44:55',
      },
      {
        name: 'Destination MAC',
        flag: '--dst-mac',
        type: 'string (MAC address)',
        defaultValue: 'ff:ff:ff:ff:ff:ff',
        required: false,
        techDesc: 'Destination MAC address for generated frames',
        laymanDesc: 'Target address on packets',
        example: '--dst-mac 00:aa:bb:cc:dd:ee',
      },
      {
        name: 'VLAN ID',
        flag: '--vlan-id',
        type: 'integer (0-4095)',
        defaultValue: '0',
        required: false,
        techDesc: 'VLAN identifier for 802.1Q tagged frames',
        laymanDesc: 'Virtual network tag (0 = no VLAN)',
        example: '--vlan-id 100',
      },
      {
        name: 'VLAN Priority',
        flag: '--vlan-priority',
        type: 'integer (0-7)',
        defaultValue: '0',
        required: false,
        techDesc: 'Priority Code Point for 802.1p CoS',
        laymanDesc: 'Traffic priority (7 = highest)',
        example: '--vlan-priority 5',
      },
    ],
    metrics: [
      {
        name: 'Tx Rate',
        unit: 'Mbps',
        goodRange: 'Matches configured rate',
        badMeaning: 'Unable to achieve requested rate',
      },
      {
        name: 'Tx Packets',
        unit: 'count',
        goodRange: 'Stable count increase',
        badMeaning: 'Transmission issues',
      },
      {
        name: 'Rx Packets',
        unit: 'count',
        goodRange: 'Matches Tx when looped back',
        badMeaning: 'Packet loss detected',
      },
    ],
    passCriteria: 'Traffic generated at requested rate',
    failMeaning: 'Unable to generate or receive traffic',
    examples: [
      {
        desc: 'Generate 50% rate traffic with 512-byte frames',
        command: 'stem test -i eth0 -t custom_stream --rate-pct 50 --frame-size 512',
        output: 'Tx Rate: 500 Mbps, Tx: 1.2M pps',
      },
      {
        desc: 'Burst mode with VLAN tagging',
        command: 'stem test -i eth0 -t custom_stream --burst-mode --vlan-id 100',
        output: 'Burst Mode: 100 frames/burst, VLAN 100',
      },
    ],
    tips: [
      'Use burst mode to test switch buffer behavior',
      'VLAN tagging requires 802.1Q-capable equipment',
      'Monitor receiver to verify packet delivery',
    ],
    seeAlso: ['throughput', 'back_to_back'],
  },
};
