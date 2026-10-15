/**
 * @fileoverview Stem - Custom Traffic Generation Test Definitions
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
    whenNotToUse: `• Use standard tests (RFC 2544, Y.1564) when applicable
• For certification or compliance testing`,
    parameters: [
      {
        name: 'Frame Size',
        setting: 'trafficGen.frameSize',
        type: 'integer (bytes)',
        defaultValue: '1518',
        required: false,
        techDesc: 'Ethernet frame size in bytes (64-9216)',
        laymanDesc: 'Size of each packet to generate',
      },
      {
        name: 'Rate Percent',
        setting: 'trafficGen.ratePct',
        type: 'float (percentage)',
        defaultValue: '100.0',
        required: false,
        techDesc: 'Traffic rate as percentage of line rate',
        laymanDesc: 'How fast to send traffic (100% = maximum speed)',
      },
      {
        name: 'Duration',
        setting: 'trafficGen.duration',
        type: 'integer (seconds)',
        defaultValue: '60',
        required: false,
        techDesc: 'Duration of traffic generation',
        laymanDesc: 'How long to generate traffic',
      },
      {
        name: 'Warmup',
        setting: 'trafficGen.warmup',
        type: 'integer (seconds)',
        defaultValue: '2',
        required: false,
        techDesc: 'Warmup period before measurements',
        laymanDesc: 'Time to stabilize before counting',
      },
      {
        name: 'Stream ID',
        setting: 'trafficGen.streamId',
        type: 'integer',
        defaultValue: '1',
        required: false,
        techDesc: 'Unique identifier for this traffic stream',
        laymanDesc: 'Label for tracking this traffic',
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
        desc: 'Custom stream with default settings (set rate, frame size, burst mode and VLAN in the test configuration form)',
        command: 'stem test -i eth0 --peer 192.0.2.10 -t custom_stream',
        output: 'Tx Rate: 500 Mbps, Tx: 1.2M pps',
      },
    ],
    tips: [
      'Use burst mode to test switch buffer behavior',
      'VLAN tagging requires 802.1Q-capable equipment',
      'Monitor receiver to verify packet delivery',
    ],
    seeAlso: [
      'https://www.rfc-editor.org/rfc/rfc2544.html',
      'rfc2544_throughput',
      'rfc2544_back_to_back',
    ],
  },
};
